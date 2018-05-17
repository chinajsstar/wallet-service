package eth

import (
	"blockchain_server/chains/eth/token"
	"blockchain_server/conf"
	"blockchain_server/types"
	"blockchain_server/utils"
	"context"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"github.com/ethereum/go-ethereum"
	//"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	ctx         context.Context
	ctx_canncel context.CancelFunc

	c           *ethclient.Client
	addresslist *utils.FoldedStrings // 监控地址列表
	blockHeight uint64               // 当前区块高度
	scanblock   uint64               // 起始扫描高度

	//rctChannel 发送的充值 被Client.Manager.loopRechargeTxMessage 接收到
	rctChannel     types.RechargeTxChannel
	address_locker *sync.RWMutex
	confirm_count  uint16 // 交易确认数

	// TODO:后面支持动态增加ERC20代币的时候, 这个map应该改成协程同步的!
	// TODO:这个map中保存的token, 实际上是config中的指针, 并通过abi更新了相关字段信息
	// TODO:这个token应该要重新写会到config文件中去
	tokens    map[string]*types.Token // address->types.EthToken
	contracts map[string]*token.Token // symbol ->token.EthToken
	//abi       abi.ABI
}

func (self *Client) lock() {
	self.address_locker.Lock()
}

func (self *Client) unlock() {
	self.address_locker.Unlock()
}

var (
	instance *Client = nil
)

func ClientInstance() (*Client, error) {
	if instance != nil {
		return instance, nil
	}

	client, err := ethclient.Dial(config.MainConfiger().Clientconfig[types.Chain_eth].RPC_url)
	if nil != err {
		l4g.Error("create eth client error! message:%s", err.Error())
		return nil, err
	}
	c := &Client{c: client}
	if err := c.init(); err != nil {
		return nil, err
	}
	instance = c
	return instance, nil
}

func (self *Client) init() error {
	// self.ctx must
	self.ctx, self.ctx_canncel = context.WithCancel(context.Background())

	self.addresslist = new(utils.FoldedStrings)

	// 起始扫描高度为 当前真实高度 - 确认数
	if _, err := self.refreshBlockHeight(); err != nil {
		//l4g.Trace("eth refresh height faild, message:%s", err)
		return err
	}

	configer := config.MainConfiger().Clientconfig[types.Chain_eth]

	self.confirm_count = uint16(configer.TxConfirmNumber)
	if self.confirm_count <= 0 {
		return fmt.Errorf("eth confirm number must not be zero.")
	}

	self.scanblock = configer.Start_scan_Blocknumber
	// if start block number==0, then, begin with current block height
	if self.scanblock == 0 {
		self.scanblock = self.virtualBlockHeight()
	}

	self.address_locker = new(sync.RWMutex)

	// 如果为0, 则默认从最新的块开始扫描
	if self.scanblock == 0 {
		self.scanblock = self.virtualBlockHeight()
	}

	self.tokens = make(map[string]*types.Token, 128)
	l4g.Trace("----------------Init %s EthToken-----------------")
	// create configed tokens
	self.contracts = make(map[string]*token.Token)

	for symbol, tk := range configer.Tokens {
		tkaddress := common.HexToAddress(tk.Address)
		tmp_tk, err := token.NewToken(tkaddress, self.c)

		if err != nil {
			l4g.Trace("Create EthToken instance(%s) error:%s", symbol, err.Error())
			continue
		}

		tk.Name, err = tmp_tk.Name(nil)
		if err != nil {
			continue
		}
		if d, err := tmp_tk.Decimals(nil); err == nil {
			tk.Decimals = uint(d)
		}
		tk.Symbol, _ = tmp_tk.Symbol(nil)

		self.contracts[tk.Symbol] = tmp_tk
		l4g.Trace("EthToken information: {%s}", tk.String())

		self.tokens[tkaddress.String()] = tk
	}

	//var err error
	//if self.abi, err = abi.JSON(strings.NewReader(token.TokenABI)); err != nil {
	//	// TODO:here may should print error message and exit process
	//	return err
	//	//l4g.Error("Create erc20 token abi error:%s", err.Error())
	//}
	return nil
}

func (self *Client) SetNotifyChannel(ch chan interface{}) {
	return
}

func (self *Client) Start() error {
	if err := self.beginSubscribeBlockHaders(); err != nil {
		self.Stop()
		return err
	}

	if err := self.beginScanBlock(); err != nil {
		self.Stop()
		return err
	}
	return nil
}

func (self *Client) Name() string {
	return types.Chain_eth
}

func (self *Client) NewAccount(c uint32) ([]*types.Account, error) {
	if c > 256 {
		c = 256
	}
	accounts := make([]*types.Account, c)
	for i := 0; uint32(i) < c; i++ {
		if tmp, err := NewAccount(); err != nil {
			l4g.Trace("%s Create New account error message:%s", err.Error())
			return nil, err
		} else {
			accounts[i] = tmp
		}
	}
	return accounts, nil
}

func (self *Client) SubscribeRechargeTx(txRechChannel types.RechargeTxChannel) {
	self.rctChannel = txRechChannel
}

func (self *Client) txToken(tx *types.Transfer) *token.Token {
	if tx == nil || tx.TokenTx == nil {
		return nil
	}
	return self.contracts[tx.TokenTx.Contract.Symbol]
}

func (self *Client) buildRawTx(from, to common.Address, value uint64, input []byte) (*etypes.Transaction, error) {
	var (
		err             error
		nonce, gaslimit uint64
		gasprice        *big.Int
		pendingcode     []byte
	)

	if nonce, err = self.c.PendingNonceAt(context.TODO(), from); err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	if gasprice, err = self.c.SuggestGasPrice(context.TODO()); err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	if pendingcode, err = self.c.PendingCodeAt(context.TODO(), to); err != nil {
	} else {
		if len(pendingcode) == 0 { // 这是一个普通账号!!!
		} else { // 这是一个合约地址!!!
		}
	}

	amount := big.NewInt(int64(value))
	msg := ethereum.CallMsg{From: from, To: &to, Value: amount, Data: input}

	gaslimit, err = self.c.EstimateGas(context.TODO(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}
	return etypes.NewTransaction(nonce, to, amount, gaslimit, gasprice, input), nil
}

func (self *Client) blockTraceTx(tx *etypes.Transaction) (bool, error) {
	state := types.Tx_state_pending
	var tmptx *etypes.Transaction
	var err error

	for i := 0; i < 60; i++ {
		switch state {
		case types.Tx_state_pending, types.Tx_state_mined:
			{
				tmptx, err = self.c.TransactionByHash(context.TODO(), tx.Hash())

				if err != nil {
					if err == ethereum.NotFound {
						return false, err
					} else {
						return false, err
					}
				} else {
					if tmptx.Inblock != 0 {
						if state != types.Tx_state_mined {
							state = types.Tx_state_mined
							l4g.Trace("Transaction(%s) state changed to 'mind' onblock=%d",
								tmptx.Hash().String(), tmptx.Inblock)
						} else if int(int64(self.realBlockHeight())-int64(tmptx.Inblock)) > int(self.confirm_count) {
							l4g.Trace("Transaction(%s) state changed to 'confirmed'", tmptx.Hash().String())
							state = types.Tx_state_confirmed
						}
					}
				}
			}
		case types.Tx_state_confirmed:
			{
				receipt, err := self.c.TransactionReceipt(context.TODO(), tx.Hash())
				if err != nil {
					return false, err
				}

				// if codeat is not empty, means address is a contact address
				// in practice testing, receipt transaction of cantract,
				// if success it's 'Logs' field will not be empty, or will faild ,
				// this conclusion need more proving
				codeat, _ := self.c.PendingCodeAt(context.TODO(), *tx.To())
				//self.c.CodeAt()

				if (receipt.Status == etypes.ReceiptStatusFailed || (len(codeat) != 0 && len(receipt.Logs) == 0)) && false {
					state = types.Tx_state_unconfirmed
				} else {
					if receipt.GasUsed > tmptx.Gas() {
						state = types.Tx_state_unconfirmed
					} else {
						l4g.Trace("transaction(%s), receipt success!")
						state = types.Tx_state_confirmed
					}
				}
				return state == types.Tx_state_confirmed, nil
			}
		}
		time.Sleep(time.Second * 3)
		// TODO: to check quit channel and exit
	}
	return false, fmt.Errorf("trace tx(%s), time out", tx.Hash().String())
}

func (self *Client) approveTokenTx(ownerKey, spenderKey, contract_string string, value uint64) error {
	privOwnerKey, owner_string, err := ParseKey(ownerKey)
	if err != nil {
		return err
	}
	privSpenderKey, spender_string, err := ParseKey(spenderKey)
	if err != nil {
		return err
	}

	owner := common.HexToAddress(owner_string)
	spender := common.HexToAddress(spender_string)
	contract := common.HexToAddress(contract_string)

	input := common.FromHex("0x095ea7b3")
	input = append(input, common.LeftPadBytes(spender.Bytes(), 32)[:]...)
	input = append(input, common.LeftPadBytes(big.NewInt(int64(value)).Bytes(), 32)[:]...)

	// 这只是一个合约授权给地址转账的Transaction, 所以, value(ether.Wei)值为0
	etx, err := self.buildRawTx(owner, contract, 0, input)

	if err != nil {
		return err
	}

	txfee := etx.GasPrice()
	txfee = txfee.Mul(txfee, big.NewInt(int64(etx.Gas())))

	signer := etypes.HomesteadSigner{}

	{
		// 由Spender 发送 txfee 给owner, 作为执行授权的交易费
		stx, err := self.buildRawTx(spender, owner, txfee.Uint64(), nil)
		if err != nil {
			return err
		}

		signedTx, err := etypes.SignTx(stx, signer, privSpenderKey)
		if err != nil {
			return err
		}
		if err = self.c.SendTransaction(context.TODO(), signedTx); err != nil {
			return err
		}

		isok, err := self.blockTraceTx(signedTx)
		if err != nil {
			return err
		}

		if !isok {
			return fmt.Errorf("trace txfee send error!")
		}
	}

	signedTx, err := etypes.SignTx(etx, signer, privOwnerKey)
	if err != nil {
		return err
	}

	if err = self.c.SendTransaction(context.TODO(), signedTx); err != nil {
		l4g.Error("SendTransaction error:message:%s", err.Error())
		return err
	}

	var isok bool
	isok, err = self.blockTraceTx(signedTx)

	if err != nil {
		return err
	}
	if !isok {
		return fmt.Errorf("approve Token Tx(%s) faild", signedTx.Hash().String())
	}
	return nil
}

// from is a crypted private key
func (self *Client) SendTx(fromkey string, tx *types.Transfer) error {

	key, from, err := ParseKey(fromkey)
	if err != nil {
		return err
	}
	tx.From = from
	// 如果tx.From不等于tx.TokenTx.From, 应该是从用户地址转出token
	// 用户地址上应该是没有ether的. 则需要授权token
	// 如果tx.From==tx.TokenTx.From, 说明是用户从热钱包地址提币,
	// 热钱包地址应该保留了一定数量的ether
	if true {
		if tx.IsTokenTx() && tx.From != tx.TokenTx.From {
			if err := self.approveTokenTx(tx.TokenFromKey, fromkey,
				// tx.To == tx.TokenTx.Contract.Address, tx.TokenTx.Contract.Address,
				tx.TokenTx.ContractAddress(),
				tx.TokenTx.TokenDecimalValue()); err != nil {
				return err
			}
		}
	}

	etx, err := self.innerBuildTx(tx)
	if err != nil {
		return err
	}

	tx.State = types.Tx_state_BuildOk

	signer := etypes.HomesteadSigner{}
	signedTx, err := etypes.SignTx(etx, signer, key)

	tx.Tx_hash = signedTx.Hash().String()

	if err != nil {
		l4g.Error("sign Transaction error:%s", err.Error())
		return err
	}

	tx.State = types.Tx_state_Signed
	if err := self.c.SendTransaction(context.TODO(), signedTx); err != nil {
		l4g.Trace("Transaction gas * price + value = %d", signedTx.Cost().Uint64())
		l4g.Error("SendTransaction error: %s", err.Error())
		return err
	}
	tx.State = types.Tx_state_pending
	return nil
}

func (self *Client) GetBalance(addstr string, tokenname *string) (uint64, error) {
	address := common.HexToAddress(addstr)
	var bl uint64 = 0
	if tokenname == nil || strings.Trim(*tokenname, " ") == "" {
		if balance, err := self.c.BalanceAt(self.ctx, address, nil); err != nil {
			return 0, err
		} else {
			bl = balance.Uint64()
		}
	} else {
		tmpToken := self.contracts[*tokenname]
		if nil == tmpToken {
			return 0, fmt.Errorf("GetBalance, Not Supported assert type!")
		}
		//tmpToken := self.contracts[tmpToken.ContractAddress]
		if nil == tmpToken {
			return 0, fmt.Errorf("GetBalance, Not Supported assert type!")
		}
		if balance, err := tmpToken.BalanceOf(nil, address); err != nil {
			return 0, fmt.Errorf("Get EthToken[%s] Balance of %s error, message:%s",
				*tokenname, addstr, err.Error())
		} else {
			bl = balance.Uint64()
		}
	}
	return bl, nil
}

// 自动以标准的精度转换为client相关的精度数量
// 1,000,000,000
func (self *Client) toClientDecimal(v uint64) *big.Int {
	i := 18 - types.StandardDecimal
	ibig := big.NewInt(int64(v))
	if i > 0 {
		return ibig.Mul(ibig, big.NewInt(int64(math.Pow10(i))))
	} else {
		return ibig.Div(ibig, big.NewInt(int64(math.Pow10(-i))))
	}
}

func (c *Client) toStandardDecimalWithBig(ibig *big.Int) uint64 {
	i := types.StandardDecimal - 18
	if i > 0 {
		return ibig.Mul(ibig, big.NewInt(int64(math.Pow10(i)))).Uint64()
	} else {
		return ibig.Div(ibig, big.NewInt(int64(math.Pow10(-i)))).Uint64()
	}
}

// 从client相关的精度数量转为自定义标准的精度数量
func (self *Client) toStandardDecimalWithInt(v uint64) uint64 {
	i := types.StandardDecimal - 18
	ibig := big.NewInt(int64(v))
	if i > 0 {
		return ibig.Mul(ibig, big.NewInt(int64(math.Pow10(i)))).Uint64()
	} else {
		return ibig.Div(ibig, big.NewInt(int64(math.Pow10(-i)))).Uint64()
	}
}

// TODO: 创建出来的tx,如果是token的话, 交易总是失败, 目前发送Token不用这个方法, 后面再检查
//func (self *Client) buildStandardTx(tx *types.Transfer) (*etypes.Transaction, error) {
//	if tx.IsTokenTx() {
//		return nil, fmt.Errorf("Use 'buildTokenTrasferTx form token's transaction!")
//	}
//
//	from := common.HexToAddress(tx.From)
//	to := common.HexToAddress(tx.To)
//
//	value := self.toClientDecimal(tx.Value)
//
//	l4g.Trace("value is: %v", value)
//
//	var input []byte = nil
//
//	// gas limit
//	msg := ethereum.CallMsg{From: from, To: &to, Value: value, Data: input}
//	gaslimit, err := self.c.EstimateGas(context.TODO(), msg)
//	if err != nil {
//		return nil, err
//	}
//
//	// gas price
//	gasprice, err := self.c.SuggestGasPrice(context.TODO())
//	if err != nil {
//		return nil, err
//	}
//
//	// nonce
//	nonce, err := self.c.PendingNonceAt(context.TODO(), common.HexToAddress(tx.From))
//	if nil != err {
//		return nil, err
//	}
//
//	return etypes.NewTransaction(nonce, to, value, gaslimit, gasprice, input), nil
//}

func (self *Client) refreshBlockHeight() (uint64, error) {
	block, err := self.c.BlockByNumber(self.ctx, nil)
	if err != nil {
		l4g.Error("ETH get block height faild, message:%s", err.Error())
		return 0, err
	}
	h := block.NumberU64()
	atomic.StoreUint64(&self.blockHeight, h)
	return h, nil
}

func (self *Client) realBlockHeight() uint64 {
	return atomic.LoadUint64(&self.blockHeight)
}

func (self *Client) BlockHeight() uint64 {
	return self.virtualBlockHeight()
}

func (self *Client) virtualBlockHeight() uint64 {
	rh := self.realBlockHeight()
	if rh > uint64(self.confirm_count) {
		return rh - uint64(self.confirm_count)
	}
	return 0
}

//func (self *Client)updateTxWithTx(transfer *types.Transfer, transaction *etypes.Transaction) bool {
//	var state types.TxState
//
//	if inblock ==0 {
//		state = types.Tx_state_pending
//	} else if lastblock-inblock >= config.MainConfiger().Clientconfig[types.Chain_eth].TxConfirmNumber {
//		state = types.Tx_state_confirmed
//	} else {
//		state = types.Tx_state_mined
//	}
//
//	// TODO: 这里检查tx.To()是否为nil, 如果为空,则tx为一个合约的交易,
//	// TODO: 就需要调用TransactionReceipt(), 来获取合约的详细信息
//
//	tmp_tx := &types.Transfer{
//		Tx_hash:           tx.Hash().String(),
//		From:              tx.From(),
//		To :               tx.To().String(),
//		Value:             tx.Value().Uint64(),
//		Fee :             tx.Fee(),
//		Gaseprice:         tx.GasPrice().Uint64(),
//		Total :            tx.Cost().Uint64(),
//		blockNumber:       inblock,
//		ConfirmatedHeight: lastblock,
//		State:             state,
//	}
//}

// 如果是合约的交易, 需要使用TransactionReceipt来获取合约地址名称等!
func (self *Client) Tx(tx_hash string) (*types.Transfer, error) {
	hash := common.HexToHash(tx_hash)
	tx, err := self.c.TransactionByHash(self.ctx, hash)

	if err != nil {
		if err == ethereum.NotFound {
			return nil, types.NewTxNotFoundErr(tx_hash)
		}
		return nil, err
	}
	return self.toTx(tx), nil
}

func (self *Client) beginSubscribeBlockHaders() error {
	header_chan := make(chan *etypes.Header)
	subscription, err := self.c.SubscribeNewHead(self.ctx, true, header_chan)

	if err != nil || nil == subscription {
		l4g.Error("subscribefiler error:%s\n", err.Error())
		return err
	}
	l4g.Trace("eth Subscribe new block header, begin!")

	go func(header_chan chan *etypes.Header) {
		defer subscription.Unsubscribe()
		for {
			select {
			case <-self.ctx.Done():
				{
					l4g.Error("subscribe block header exit loop for:%s", self.ctx.Err().Error())
					close(header_chan)
					goto endfor
				}
			case header := <-header_chan:
				{
					h := header.Number.Uint64()
					// 设置当前块高为 真实高度 - 确认数的高度
					atomic.StoreUint64(&self.blockHeight, h)

					//l4g.Trace("get new block(%s) height:%d",
					//	header.Hash().String(), h)
				}
			}
		}
	endfor:
		l4g.Trace("Will stop subscribe block header!")
	}(header_chan)

	return nil
}

func (self *Client) beginScanBlock() error {
	go func() {
		var scanblock, top uint64
		for {
			// top block height
			top = self.virtualBlockHeight()
			scanblock = self.scanblock

			// the following express to make sure blockchain are not forked
			// height <= (self.scanblock-self.confirm_count)
			if nil == self.addresslist || len(*self.addresslist) == 0 ||
				//height <= self.scanblock - uint64(self.confirm_count) ||
				top < self.scanblock ||
				self.rctChannel == nil {
				//l4g.Trace("Recharge channel is nil, or block height(%d)==scanblock(%d), or addresslit is empty!", top, self.scanblock)
				time.Sleep(time.Second * 5)
				continue
			}

			block, err := self.c.BlockByNumber(self.ctx, big.NewInt(int64(scanblock)))

			if err != nil {
				l4g.Error("get block error, stop scanning block, message:%s", err.Error())
				goto endfor
			}

			l4g.Trace("scaning block :%d", self.scanblock)

			txs := block.Transactions()


			for _, tx := range txs {
				l4g.Trace("tx on block(%d), tx information:%s", block.NumberU64(), tx.String())
				to := tx.To()
				if to == nil { continue }

				addresses := []string{tx.From()}

				// check if 'to' is a cantract address
				if code, err := self.c.PendingCodeAt(context.TODO(), *to);
					err == nil && len(code) != 0 {
					tk := self.tokens[to.String()]
					if tk==nil { continue }
					tkowner, tkreciver, _, err :=  token.ParseTokenTxInput(tx.Data())
					if err==nil {
						if tkowner!="" 	 {addresses = append(addresses, tkowner) }
						if tkreciver!="" {addresses = append(addresses, tkreciver)}
					}
				} else {
					addresses = append(addresses, to.String())
				}

				for _, tmp := range addresses {
					if self.hasAddress(tmp) {
						// TODO: 测试发现tx中的inblock为0, 应该是库的bug, 先在这里手动设置
						// TODO: 以后需要看看ethereum库中相关部分
						tx.Inblock = scanblock
						tmp_tx := self.toTx(tx)
						// rctChannel 触发以后, 被ClientManager.loopRechargeTxMessage函数处理!
						self.rctChannel <- &types.RechargeTx{types.Chain_eth, tmp_tx, nil}
						break
					}
				}
			}
			self.scanblock++

			// scan 20 block, once save
			if self.scanblock%20 == 0 {
				self.saveConfigurations()
			}

			select {
			case <-self.ctx.Done():
				{
					l4g.Trace("stop scaning blocks! for message:%s", self.ctx.Err().Error())
					goto endfor
				}
			default:
				{
					time.Sleep(time.Millisecond * 200)
				}
			}
		}
	endfor:
	}()
	return nil
}

func (self *Client) hasAddress(address string) bool {
	self.lock()
	defer self.unlock()
	return self.addresslist.Contains(address)
}

func (self *Client) blockTime(bn uint64) uint64 {
	if block, err := self.c.BlockByNumber(self.ctx, big.NewInt(int64(bn))); err != nil {
		return 0
	} else {
		return block.Time().Uint64()
	}
}

func (self *Client) updateTxWithReceipt(tx *types.Transfer) error {
	height := self.virtualBlockHeight()

	receipt, err := self.c.TransactionReceipt(self.ctx, common.HexToHash(tx.Tx_hash))
	if err != nil {
		l4g.Error("Update Transaction(%s) with Receipt error : %s", tx.Tx_hash, err.Error())
		return err
	}

	// The jenises must set 'ByzantiumBlock' field, or
	// recipt.Status always be 'ReceiptStatusFailed'

	// if IsTokenTx()==true, len(receipt.Logs) will not be 0,
	// in practice testing, receipt transaction of cantract,
	// if success it's 'Logs' field will not be empty, or will faild ,
	// this conclusion need more proving, no offical document said that!!!!
	if (receipt.Status == etypes.ReceiptStatusFailed ||
		(tx.IsTokenTx() && len(receipt.Logs) == 0)) && false {
		tx.State = types.Tx_state_unconfirmed
		l4g.Trace("Transaction(%s) receipt statue: faild", tx.Tx_hash)
	} else {
		if receipt.GasUsed > tx.Gas { // not enough tx fee
			tx.State = types.Tx_state_unconfirmed
		} else {
			if tx.InBlock <= height {
				tx.State = types.Tx_state_confirmed
				tx.Gas = receipt.GasUsed
				tx.ConfirmatedHeight = height + uint64(self.confirm_count)
			} else {
				tx.State = types.Tx_state_mined
			}
		}
	}

	return nil
}


// 使用srcTx更新destTx, 如果为空, 则使用TransactionByHash获取最新的tx
// 如果使用了已经存在的Transaction来更新Transfer, 程序不会从网络去获取Transaction, 不会检查txhash是否存在!
// 则无法判断Transaction是否被分叉的问题
func (self *Client) updateTxWithTx(destTx *types.Transfer, srcTx *etypes.Transaction) error {
	if destTx.Confirmationsnumber == 0 {
		destTx.Confirmationsnumber = uint64(self.confirm_count)
	}

	height := self.virtualBlockHeight()

	// if input srcTx is nil, try to get transaction from network node
	if srcTx == nil {
		var err error
		srcTx, err = self.c.TransactionByHash(self.ctx, common.HexToHash(destTx.Tx_hash))
		if err != nil {
			if err == ethereum.NotFound {
				// 如果之前的状态为已经上块, 现在又找不到了, 说明可能被分叉了
				// 把状态设置为unconfirmed
				if destTx.State == types.Tx_state_mined || destTx.State == types.Tx_state_confirmed {
					destTx.State = types.Tx_state_unconfirmed
				} else {
					// 可能为交易才被提交, 这里不做任何处理
					return types.NewTxNotFoundErr(destTx.Tx_hash)
				}
			}
		}
	}
	// 用一个临时变量保存state, 等到更新完成后, 才设置destTx的状态
	// 如果设置之后再更新, 如果中间的步骤出错, 状态已经设置, 程序会认为交易已经成功
	var state types.TxState

	// 由于这里使用的是virtualBlockHeight,作为当前块高, 实际上已经减去了
	// confirm_count, 所以只要virtualblockheight = inblock 就可以视为已经确认!
	if srcTx.Inblock == 0 {
		state = types.Tx_state_pending
	} else if srcTx.Inblock <= height {
		//else if  (height > srcTx.Inblock) && uint16(height-srcTx.Inblock) > self.confirm_count {
		state = types.Tx_state_confirmed
	} else {
		state = types.Tx_state_mined
	}

	// time=0, 说明destTx 之前还没有更新过
	if destTx.Time == 0 {
		// check if 'to' is a cantract address
		to := srcTx.To()
		if to!=nil {
			if code, err := self.c.PendingCodeAt(context.TODO(), *to);
				err == nil && len(code) != 0 {
				tkfrom, tkto, value, err := token.ParseTokenTxInput(srcTx.Data())
				if err!=nil { return err }

				if destTx.TokenTx==nil {
					destTx.TokenTx = &types.TokenTx{}
				}

				if tkfrom!="" {
					destTx.TokenTx.From = tkfrom
				} else {
					destTx.TokenTx.From = srcTx.From()
				}

				destTx.TokenTx.To = tkto
				destTx.TokenTx.Value = value.Uint64()

				destTx.TokenTx.Contract = self.tokens[to.String()]
			}
			destTx.To = to.String()
		} else {
			return fmt.Errorf("Transaction(%s) To address is nil?????", srcTx.Hash().String())
		}
		destTx.From = srcTx.From()
		destTx.Value = self.toStandardDecimalWithBig(srcTx.Value())
		destTx.Total = self.toStandardDecimalWithBig(srcTx.Cost())
		destTx.Fee = destTx.Total - destTx.Value
		destTx.InBlock = srcTx.Inblock
		destTx.Gas = srcTx.Gas()
		destTx.Time = self.blockTime(destTx.InBlock)
	}
	destTx.State = state

	if state == types.Tx_state_confirmed {
		if err := self.updateTxWithReceipt(destTx); err == nil {
			if destTx.State == types.Tx_state_confirmed {
				destTx.ConfirmatedHeight = height + uint64(self.confirm_count)
			}
		}
	}

	return nil
}

func (self *Client) UpdateTx(tx *types.Transfer) error {
	if nil == tx || len(tx.Tx_hash) == 0 {
		return fmt.Errorf("Invalid paramater!")
	}

	if false {
		// this way of checking tx status was deprecated
		if tx.State == types.Tx_state_confirmed {
			if err := self.updateTxWithReceipt(tx); err != nil {
				return err
			}
			return nil
		} else { // tx.State==types.Tx_state_pending, Tx_state_pending, tx_state_unkown, tx_state_notfound
			if err := self.updateTxWithTx(tx, nil); err != nil {
				return err
			}
			return nil
		}
	}

	if true {
		if err := self.updateTxWithTx(tx, nil); err != nil {
			return err
		}
	}
	return nil
}

func (self *Client) toTx(tx *etypes.Transaction) *types.Transfer {
	tmpTx := &types.Transfer{Tx_hash: tx.Hash().String()}
	self.updateTxWithTx(tmpTx, tx)
	return tmpTx
}

func (self *Client) InsertRechargeAddress(address []string) (invalid []string) {
	self.lock()
	defer self.unlock()

	if self.addresslist == nil {
		self.addresslist = new(utils.FoldedStrings) // utils.FoldedStrings(make(string[], 0, 512))
	}

	if self.addresslist.Len() == 0 {
		for _, value := range address {
			self.addresslist.Insert(value)
		}
		self.addresslist.Sort()
		return
	}

	for _, value := range address {
		if !self.addresslist.Contains(value) {
			self.addresslist.Insert(value)
		}
		l4g.Trace(value)
	}

	// eth always return nil
	return
}

func (self *Client) saveConfigurations() {
	atomic.StoreUint64(
		&config.MainConfiger().Clientconfig[types.Chain_eth].Start_scan_Blocknumber,
		atomic.LoadUint64(&self.scanblock))

	config.MainConfiger().Save()
}

func (self *Client) Stop() {
	self.ctx_canncel()
	self.saveConfigurations()
}

// from is a crypted private key, if ok, may replace SendTx
func (self *Client) SendTxBySteps(chiperKey string, tx *types.Transfer) error {
	err := self.BuildTx(chiperKey, tx)
	if err != nil {
		l4g.Error("buildtx failed: %s", err.Error())
		return err
	}

	txByte, err := self.SignTx(chiperKey, tx)
	if err != nil {
		l4g.Error("signtx failed: %s", err.Error())
		return err
	}
	return self.SendSignedTx(txByte, tx)
}

func (self *Client) innerBuildTx(tx *types.Transfer) (etx *etypes.Transaction, err error) {
	var input []byte = nil

	from := common.HexToAddress(tx.From)
	to := common.HexToAddress(tx.To)

	if tkTx := tx.TokenTx; tkTx != nil {
		_, tkTx.From, err = ParseKey(tx.TokenFromKey)
		tokenFrom := common.HexToAddress(tx.TokenTx.From)
		tokenTo := common.HexToAddress(tx.TokenTx.To)
		value := tx.TokenTx.Contract.ToTokenDecimal(tkTx.Value)

		// we can construct input data, with no using of token.TOKENABI.Pack function
		if false {
			if tx.From == tx.TokenTx.From {
				input, err = token.TOKENABI.Pack("transferFrom", tokenFrom, tokenTo, value)
			} else {
				input, err = token.TOKENABI.Pack("transfer", tokenTo, value)
			}
		} else {
			if tx.From == tx.TokenTx.From {
				input = common.FromHex("0xa9059cbb")
			} else {
				input = common.FromHex("0x23b872dd")
				input = append(input, common.LeftPadBytes(tokenFrom.Bytes(), 32)[:]...)
			}
			input = append(input, common.LeftPadBytes(tokenTo.Bytes(), 32)[:]...)
			input = append(input, common.LeftPadBytes(value.Bytes(), 32)[:]...)
		}
		if err != nil {
			return nil, err
		}
	}

	return self.buildRawTx(from, to, tx.Value, input)
}

// build transaction
func (self *Client) BuildTx(fromkey string, tx *types.Transfer) (err error) {
	_, tx.From, err = ParseKey(fromkey)
	if err != nil {
		return
	}
	var etx *etypes.Transaction
	if etx, err = self.innerBuildTx(tx); err != nil {
		return err
	}

	if tx.Additional_data, err = etx.MarshalJSON(); err != nil {
		l4g.Error("ethereum BuildTx faild, message:%s", err.Error())
		return err
	} else {
		tx.State = types.Tx_state_BuildOk
	}

	return err
}

// sign transaction
func (self *Client) SignTx(chiperKey string, tx *types.Transfer) ([]byte, error) {
	if tx.State != types.Tx_state_BuildOk {
		return nil, fmt.Errorf("Cannot sign Tx, which state not Tx_state_BuildOk")
	}

	key, from, err := ParseKey(chiperKey)
	if err != nil {
		return nil, err
	}

	// check pubkkey is prikey's address
	if strings.ToLower(from) != strings.ToLower(tx.From) {
		return nil, fmt.Errorf("tx.Form(%s) is not equal prikey's address(%s)", tx.From, from)
	}

	etx := &etypes.Transaction{}
	if err = etx.UnmarshalJSON(tx.Additional_data); err != nil {
		return nil, err
	}

	signer := etypes.HomesteadSigner{}
	signedTx, err := etypes.SignTx(etx, signer, key)
	if err != nil {
		l4g.Error("sign Transaction error:%s", err.Error())
		return nil, err
	}

	tx.Additional_data, err = signedTx.MarshalJSON()
	if err != nil {
		l4g.Error("SignTx-MarshalJSON error:%s", err.Error())
		return nil, err
	}

	l4g.Trace("eth sign Transaction: ", signedTx.String())
	tx.State = types.Tx_state_Signed
	return tx.Additional_data, nil
}

// send signed transaction
func (self *Client) SendSignedTx(txByte []byte, tx *types.Transfer) error {
	var signedTx etypes.Transaction
	err := signedTx.UnmarshalJSON(txByte)
	if err != nil {
		l4g.Error("SendSignedTx UnmarshalJSON: %s", err.Error())
		return err
	}

	if err := self.c.SendTransaction(context.TODO(), &signedTx); err != nil {
		l4g.Trace("Transaction gas * price + value = %d", signedTx.Cost().Uint64())
		l4g.Error("SendTransaction error: %s", err.Error())
		return err
	}

	l4g.Trace("SendSignedTx Tx information:%s", tx.String())
	tx.State = types.Tx_state_pending
	return nil
}
