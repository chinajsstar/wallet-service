package eth

import (
	"blockchain_server/chains/eth/token"
	"blockchain_server/conf"
	"blockchain_server/types"
	"blockchain_server/utils"
	"context"
	"fmt"
	"blockchain_server/l4g"
	"github.com/ethereum/go-ethereum"
	//"github.com/ethereum/go-ethereum/accounts/abi"
	"crypto/ecdsa"
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

var (L4g = L4G.BuildL4g(types.Chain_eth, "ethereum"))

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
	nonceMtx *sync.Mutex
	nonceMap map[string]uint64
}

const (
	max_nonce          = 1024 * 1024
	first_custom_nonce = max_nonce + 1
)

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
		L4g.Error("create eth client error! message:%s", err.Error())
		return nil, err
	}
	c := &Client{c: client}
	if err := c.init(); err != nil {
		return nil, err
	}
	instance = c
	return instance, nil
}

func (self *Client) nextNonce(from string) (nonce uint64) {
	self.nonceMtx.Lock()
	var err error
	nonce = self.nonceMap[from]
	if 0 == nonce {
		nonce, err = self.c.PendingNonceAt(context.TODO(), common.HexToAddress(from))
		if err != nil {
		}
	} else {
		if nonce == max_nonce {
			nonce, err = self.c.PendingNonceAt(context.TODO(), common.HexToAddress(from))
			if err != nil {
			}
		}
	}
	self.nonceMap[from] = nonce + 1
	self.nonceMtx.Unlock()

	return nonce
}

func (self *Client) removeNonc(from string) {
	self.nonceMtx.Lock()
	delete(self.nonceMap, from)
	self.nonceMtx.Unlock()
}

func (self *Client) init() error {
	// self.ctx must
	self.ctx, self.ctx_canncel = context.WithCancel(context.Background())

	self.nonceMtx = new(sync.Mutex)
	self.nonceMap = make(map[string]uint64, 256)

	self.addresslist = new(utils.FoldedStrings)

	// 起始扫描高度为 当前真实高度 - 确认数
	if _, err := self.refreshBlockHeight(); err != nil {
		//L4g.Trace("eth refresh height faild, message:%s", err)
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

	self.tokens = make(map[string]*types.Token, 256)

	L4g.Trace("----------------Init EthToken-----------------")
	// create configed tokens
	self.contracts = make(map[string]*token.Token)

	for symbol, tk := range configer.Tokens {
		tkaddress := common.HexToAddress(tk.Address)
		tmp_tk, err := token.NewToken(tkaddress, self.c)

		if err != nil {
			L4g.Trace("Create EthToken instance(%s) error:%s", symbol, err.Error())
			continue
		}

		tk.Name, err = tmp_tk.Name(nil)
		if err != nil {
			continue
		}

		if d, err := tmp_tk.Decimals(nil); err == nil {
			tk.Decimals = int(d)
		}

		tk.Symbol, _ = tmp_tk.Symbol(nil)

		self.contracts[tk.Symbol] = tmp_tk
		L4g.Trace("EthToken information: {%s}", tk.String())

		self.tokens[tkaddress.String()] = tk
	}

	//var err error
	//if self.abi, err = abi.JSON(strings.NewReader(token.TokenABI)); err != nil {
	//	// TODO:here may should print error message and exit process
	//	return err
	//	//L4g.Error("Create erc20 token abi error:%s", err.Error())
	//}
	return nil
}

func (self *Client) SetNotifyChannel(ch chan interface{}) {
	return
}

func (self *Client) Start() error {
	go self.loopRefreshBlockheight()
	go self.startScanBlock()
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
			L4g.Trace("%s Create New account error message:%s", err.Error())
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
	return self.contracts[tx.TokenTx.Token.Symbol]
}

func (self *Client) estimatTxFee(from, to common.Address, value *big.Int,
	input []byte) (gasprice *big.Int, gaslimit uint64, err error) {
	if gasprice, err = self.c.SuggestGasPrice(context.TODO()); err != nil {
		return
	}
	msg := ethereum.CallMsg{From: from, To: &to, Value: value, Data: input}
	gaslimit, err = self.c.EstimateGas(context.TODO(), msg)
	return
}

// 这个value是以wei为单位的
func (self *Client) buildRawTx(from, to common.Address, value *big.Int, input []byte) (*etypes.Transaction, error) {

	var (
		err             error
		nonce, gaslimit uint64
		gasprice        *big.Int
		pendingcode     []byte
	)

	//if nonce, err = self.c.PendingNonceAt(context.TODO(), from); err != nil {
	//	return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	//}
	nonce = self.nextNonce(from.String())

	L4g.Trace("PendingNonceAt(%s) = %d", from.String(), nonce)

	if gasprice, err = self.c.SuggestGasPrice(context.TODO()); err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	if pendingcode, err = self.c.PendingCodeAt(context.TODO(), to); err != nil {
	} else {
		if len(pendingcode) == 0 { // 这是一个普通账号!!!
		} else { // 这是一个合约地址!!!
		}
	}

	msg := ethereum.CallMsg{From: from, To: &to, Value: value, Data: input}

	gaslimit, err = self.c.EstimateGas(context.TODO(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}
	return etypes.NewTransaction(nonce, to, value, gaslimit, gasprice, input), nil
}

func (self *Client) blockTrackTx(tx *etypes.Transaction) (bool, error) {
	state := types.Tx_state_pending
	var tmptx *etypes.Transaction
	var err error

	for i := 0; i <240; i++ {

		L4g.Trace("blockTrackTx, txhash(%s), try (%d)times, cost (%d)seconds",
			tx.Hash().String(), i, i*30)

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
							L4g.Trace("Transaction(%s) state changed to 'mind' onblock=%d",
								tmptx.Hash().String(), tmptx.Inblock)
						} else if int(int64(self.realBlockHeight())-int64(tmptx.Inblock)) > int(self.confirm_count) {
							L4g.Trace("Transaction(%s) state changed to 'confirmed'", tmptx.Hash().String())
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

				if receipt.Status == etypes.ReceiptStatusFailed {
					state = types.Tx_state_unconfirmed
				} else {
					state = types.Tx_state_confirmed
				}

				return state == types.Tx_state_confirmed, nil
			}
		}
		time.Sleep(time.Second * 30)
		// TODO: to check quit channel and exit
	}
	return false, fmt.Errorf("trace tx(%s), time out", tx.Hash().String())
}

func (self *Client) approveTokenTx(ownerKey, spenderKey, contract_string string, value *big.Int) error {
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
	input := token.BuildTokenApproveInput(spender, value)

	var (
		tx, signedTx *etypes.Transaction
		nonce        uint64
		gasprice     *big.Int
		gaslimit     uint64
		isok         bool
	)

	gasprice, gaslimit, err = self.estimatTxFee(owner, contract, big.NewInt(0), input)
	if err != nil {
		return err
	}

	txfee := new(big.Int).Set(gasprice)
	txfee = txfee.Mul(txfee, big.NewInt(int64(gaslimit)))
	signer := etypes.HomesteadSigner{}

	// Spender需要充owner转走token, 首先需要Token的Owner调用合约给自己授权
	// 所以由Spender先发送txfee给owner, 作为执行授权的交易费
	// 因为Spender中没有'ether'作为txfee
	tx, err = self.buildRawTx(spender, owner, txfee, nil)
	//L4g.Trace("Build Fee ethereumTx:%s, txfee:(%d * %d) = %d",
	//	tx.String(), gasprice.Uint64(), gaslimit, txfee)
	if err != nil {
		goto Exception
	}

	signedTx, err = etypes.SignTx(tx, signer, privSpenderKey)
	if err != nil {
		goto Exception
	}
	if err = self.c.SendTransaction(context.TODO(), signedTx); err != nil {
		goto Exception
	}

	isok, err = self.blockTrackTx(signedTx)
	if err != nil {
		goto Exception
	}

	if !isok {
		err = fmt.Errorf("trace txfee send error!")
		goto Exception
	}

	nonce = self.nextNonce(owner_string)
	//nonce, err = self.c.PendingNonceAt(context.TODO(), owner);
	if err != nil {
		goto Exception
	}
	tx = etypes.NewTransaction(nonce, contract, big.NewInt(0), gaslimit, gasprice, input)

	signedTx, err = etypes.SignTx(tx, signer, privOwnerKey)
	if err != nil {
		goto Exception
	}

	err = self.c.SendTransaction(context.TODO(), signedTx)
	if err != nil {
		goto Exception
	} else {
		L4g.Trace("SendTxOk:%s", signedTx.Hash().String())
	}
	time.Sleep(time.Second)

Exception:

	if err != nil {
		L4g.Trace("Approve token error, message:%s", err.Error())
		return err
	}

	isok, err = self.blockTrackTx(signedTx)
	if err != nil {
		return err
	}
	if !isok {
		err = fmt.Errorf("approve TokenSymbol Tx(%s) faild", signedTx.Hash().String())
	}

	return err
}

// from is a crypted private key
func (self *Client) SendTx(fromkey string, tx *types.Transfer) error {

	fromPrivkey, input, err := self.initTxInfo(fromkey, tx)
	if err != nil {
		return err
	}

	// 如果tx.From不等于tx.TokenTx.From, 应该是从用户地址转出token
	// 用户地址上应该是没有ether的. 则需要授权token
	// 如果tx.From==tx.TokenTx.From, 说明是用户从热钱包地址提币,
	// 热钱包地址应该保留了一定数量的ether
	if tx.IsTokenTx() && tx.From != tx.TokenTx.From {
		if err := self.approveTokenTx(tx.TokenFromKey, fromkey,
			tx.TokenTx.ContractAddress(),
			tx.TokenTx.Value_decimaled()); err != nil {
			return err
		}
	}

	var (
		etx, signedTx *etypes.Transaction
	)

	etx, err = self.buildRawTx(
		common.HexToAddress(tx.From),
		common.HexToAddress(tx.To),
		EtherToWei(tx.Value), input)

	if err != nil {
		goto Exception
	}

	tx.State = types.Tx_state_BuildOk
	signedTx, err = etypes.SignTx(etx, etypes.HomesteadSigner{}, fromPrivkey)
	tx.Tx_hash = signedTx.Hash().String()
	if err != nil {
		L4g.Error("sign Transaction error:%s", err.Error())
		goto Exception
	}
	tx.State = types.Tx_state_Signed

	if err = self.c.SendTransaction(context.TODO(), signedTx); err != nil {
		L4g.Error("SendTransaction error: %s txinfo:%s", err.Error(), signedTx.String())
		goto Exception
	}
	tx.State = types.Tx_state_pending

Exception:
	return err
}

func (self *Client) GetBalance(addstr string, tokenSymbol string) (float64, error) {
	address := common.HexToAddress(addstr)

	if strings.Trim(tokenSymbol, " ") == "" {
		if balance, err := self.c.BalanceAt(self.ctx, address, nil); err != nil {
			return 0, err
		} else {
			return WeiToEther(balance), nil
		}
	} else {
		tmpToken := self.contracts[tokenSymbol]
		if nil == tmpToken {
			return 0, fmt.Errorf("GetBalance, Not Supported assert type!")
		}
		//	tmpToken := self.contracts[tmpToken.ContractAddress]
		if nil == tmpToken {
			return 0, fmt.Errorf("GetBalance, Not Supported assert type!")
		}

		if balance, err := tmpToken.BalanceOf(nil, address); err != nil {
			return 0, fmt.Errorf("Get EthToken[%s] Balance of %s error, message:%s",
				tokenSymbol, addstr, err.Error())
		} else {
			decimal, err := tmpToken.Decimals(nil)
			if err != nil {
				return 0, err
			}

			fb, _ := new(big.Float).SetString(balance.String())
			f, _ := fb.Mul(fb, big.NewFloat(math.Pow10(-int(decimal)))).Float64()
			f = utils.PrecisionN(f, 6)
			return f, nil
		}
	}
}

func (self *Client) refreshBlockHeight() (uint64, error) {
	block, err := self.c.BlockByNumber(self.ctx, nil)
	if err != nil {
		L4g.Error("ETH get block height faild, message:%s", err.Error())
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

func (self *Client) loopRefreshBlockheight() {
	break_for:
	for {
		select {
		case <-self.ctx.Done(): {
			L4g.Trace("ETH Stop refresh block height")
			break break_for
		}
		default : {
			if height, err := self.refreshBlockHeight(); err!=nil {
				L4g.Trace("ETH refresh block height faild, meesage:%s", err.Error())
			} else {
				L4g.Trace("ETH refresh blocknumber: %d", height)
			}
			time.Sleep(time.Second * 5)
		}
		}
	}
	L4g.Trace("ETH loop for refresh blockheight stoped!!!")
}

//func (self *Client) subscribeBlockHeader() error {
//	header_chan := make(chan *etypes.Header)
//	subscription, err := self.c.SubscribeNewHead(self.ctx, true, header_chan)
//
//	if err != nil || nil == subscription {
//		L4g.Error("subscribefiler error:%s\n", err.Error())
//		return err
//	}
//	L4g.Trace("eth Subscribe new block header, begin!")
//
//	go func(header_chan chan *etypes.Header) {
//		defer subscription.Unsubscribe()
//
//		var lastNotifyTime time.Time = time.Now()
//
//		for {
//			select {
//			case <-self.ctx.Done():
//				{
//					L4g.Error("subscribe block header exit loop for:%s", self.ctx.Err().Error())
//					close(header_chan)
//					goto endfor
//				}
//			case header := <-header_chan:
//				{
//					lastNotifyTime = time.Now()
//
//					h := header.Number.Uint64()
//					atomic.StoreUint64(&self.blockHeight, h)
//					L4g.Trace("ETH notify new block(%s), blocknumber=%d",
//						header.Hash().String(),
//						header.Number.Uint64() )
//				}
//			default: {
//				nowtime := time.Now()
//				if nowtime.Sub(lastNotifyTime).Seconds() > 60 {
//					L4g.Trace("ETH wallet did't notify new header duration 60 second, try reconnect!!!")
//					var err error
//
//					subscription.Unsubscribe()
//
//					subscription, err = self.c.SubscribeNewHead(self.ctx, true, header_chan)
//					if err!=nil {
//						L4g.Trace("Reconnect to ETH wallet faild, message:%s", err.Error())
//					}
//
//
//					lastNotifyTime = nowtime
//
//				} else {
//					time.Sleep(5 * time.Second)
//				}
//
//			}
//
//			}
//		}
//	endfor:
//		L4g.Trace("Will stop subscribe block header!")
//	}(header_chan)
//
//	return nil
//}

func (self *Client) startScanBlock() {
	var scanblock, top uint64
	for {
		// top block height
		top = self.virtualBlockHeight()
		scanblock = self.scanblock

		// the following express to make sure blockchain are not forked
		// height <= (self.scanblock-self.confirm_count)
		if nil == self.addresslist || len(*self.addresslist) == 0 ||
		//height <= self.scanblock - uint64(self.confirm_count) ||
			top < self.scanblock || self.rctChannel == nil {
			L4g.Trace("ETH Scanblock warning:\n[blockheight(%d)==scanblock(%d)],len(addresslit)=%d.",
				top, self.scanblock,
				self.addresslist.Len())
			time.Sleep(time.Second * 2)
			continue
		}

		block, err := self.c.BlockByNumber(self.ctx, big.NewInt(int64(scanblock)))

		if err != nil {
			L4g.Error("get block error, stop scanning block, message:%s", err.Error())
			time.Sleep(time.Second * 2)
			continue
		}

		L4g.Trace("Start Scaning block :%d", self.scanblock)

		txs := block.Transactions()

		for _, tx := range txs {
			L4g.Trace("find tx(%s)", tx.Hash().String())
			to := tx.To()
			if to == nil {
				continue
			}

			// from和to都需要监控, 所以这个地方用addresses数组来保存
			// 然后检查addresses中是否有地址在监控地址中
			// 只要检查到其中一个, 就需要把交易通知到外部
			addresses := []string{tx.From()}

			// check if 'to' is a cantract address
			if code, err := self.c.PendingCodeAt(context.TODO(), *to); err == nil && len(code) != 0 {
				tk := self.tokens[to.String()]
				if tk == nil {
					continue
				}

				tkowner, tkreciver, _, err := token.ParseTokenTxInput(tx.Data())

				if err!=nil {
					L4g.Error("ETH parsing TxInputData error:%s, TxInfo:%s",
						err.Error(), tx.String())
					continue
				} else {
					if tkowner != "" { addresses = append(addresses, tkowner) }
					if tkreciver != "" { addresses = append(addresses, tkreciver) }
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
					select {
					case <-self.ctx.Done():
						L4g.Trace("stop scaning blocks! for message:%s", self.ctx.Err().Error())
						goto exitfor
					case self.rctChannel <- &types.RechargeTx{types.Chain_eth, tmp_tx, nil}:
					}
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
		case <-self.ctx.Done(): {
			L4g.Trace("stop scaning blocks! for message:%s", self.ctx.Err().Error())
			goto exitfor
		}
		default: {
			time.Sleep(time.Millisecond * 100)
		}
		}
	}
exitfor:
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
	//height := self.virtualBlockHeight()

	receipt, err := self.c.TransactionReceipt(self.ctx, common.HexToHash(tx.Tx_hash))
	if err != nil {
		L4g.Error("Update Transaction(%s) with Receipt error : %s", tx.Tx_hash, err.Error())
		return err
	}

	// The jenises must set 'ByzantiumBlock' field, or
	// recipt.Status always be 'ReceiptStatusFailed'

	// if IsTokenTx()==true, len(receipt.Logs) will not be 0,
	// in practice testing, receipt transaction of cantract,
	// if success it's 'Logs' field will not be empty, or will faild ,
	// this conclusion need more proving, no offical document said that!!!!
	if receipt.Status == etypes.ReceiptStatusSuccessful {
		tx.State = types.Tx_state_confirmed
	} else {
		tx.State = types.Tx_state_unconfirmed
	}

	return nil
}

func (self *Client) updateTxWithTx(destTx *types.Transfer, srcTx *etypes.Transaction) error {
	if srcTx == nil {
		return fmt.Errorf("update tx error, message:sourceTx is nil")
	}
	hash := srcTx.Hash().String()

	if destTx.Tx_hash == "" {
		destTx.Tx_hash = hash
	} else if hash != destTx.Tx_hash {
		return fmt.Errorf("updateTx error, destTx.Hash(%s) != srcTx.Hash(%s)",
			destTx.Tx_hash, hash)
	}

	if destTx.Confirmationsnumber == 0 {
		destTx.Confirmationsnumber = uint64(self.confirm_count)
	}
	height := self.virtualBlockHeight()
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
		if to != nil {
			if code, err := self.c.PendingCodeAt(context.TODO(), *to); err == nil && len(code) != 0 {
				tkfrom, tkto, value, err := token.ParseTokenTxInput(srcTx.Data())
				if err != nil {
					return err
				}

				if destTx.TokenTx == nil {
					destTx.TokenTx = &types.TokenTx{
						Token: self.tokens[to.String()],
					}
				}

				if tkfrom != "" {
					destTx.TokenTx.From = tkfrom
				} else {
					destTx.TokenTx.From = srcTx.From()
				}

				destTx.TokenTx.Token = self.tokens[to.String()]
				destTx.TokenTx.To = tkto
				destTx.TokenTx.SetValue_by_decimaled(value)
			}
			destTx.To = to.String()
		} else {
			return fmt.Errorf("Transaction(%s) To address is nil?????", srcTx.Hash().String())
		}

		destTx.From = srcTx.From()
		destTx.Value = WeiToEther(srcTx.Value())
		destTx.Total = WeiToEther(srcTx.Cost())
		destTx.Fee = utils.PrecisionN(destTx.Total-destTx.Value, 6)
		destTx.InBlock = srcTx.Inblock
		//destTx.Gas = srcTx.Gas()
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

	if etx, err := self.c.TransactionByHash(self.ctx, common.HexToHash(tx.Tx_hash)); err != nil {
		return err
	} else if err := self.updateTxWithTx(tx, etx); err != nil {
		return err
	}
	return nil
}

func (self *Client) toTx(tx *etypes.Transaction) *types.Transfer {
	if tx == nil {
		return nil
	}

	tmpTx := &types.Transfer{}
	if err := self.updateTxWithTx(tmpTx, tx); err != nil {
		return nil
	}

	return tmpTx
}

func (self *Client) InsertWatchingAddress(address []string) (invalid []string) {
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
		L4g.Trace(value)
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
		L4g.Error("buildtx failed: %s", err.Error())
		return err
	}

	txByte, err := self.SignTx(chiperKey, tx)
	if err != nil {
		L4g.Error("signtx failed: %s", err.Error())
		return err
	}
	return self.SendSignedTx(txByte, tx)
}

func (self *Client) initTxInfo(fromKey string, tx *types.Transfer) (fromPrivkey *ecdsa.PrivateKey, input []byte, err error) {
	var from string
	fromPrivkey, from, err = ParseKey(fromKey)
	if err != nil {
		return
	}
	tx.From = from

	if tkTx := tx.TokenTx; tkTx != nil {
		_, tkTx.From, err = ParseKey(tx.TokenFromKey)
		tokenFrom := common.HexToAddress(tx.TokenTx.From)
		tokenTo := common.HexToAddress(tx.TokenTx.To)
		value := tkTx.Value_decimaled()

		// we can construct input data, with no using of token.TOKENABI.Pack function
		if false {
			if tx.From == tx.TokenTx.From {
				input, err = token.TOKENABI.Pack("transferFrom", tokenFrom, tokenTo, value)
			} else {
				input, err = token.TOKENABI.Pack("transfer", tokenTo, value)
			}
			if err != nil {
				return nil, nil, err
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
	}
	return
}

// build transaction
func (self *Client) BuildTx(fromKey string, tx *types.Transfer) (err error) {
	var etx *etypes.Transaction
	var input []byte

	if _, input, err = self.initTxInfo(fromKey, tx); err != nil {
		return err
	}

	etx, err = self.buildRawTx(common.HexToAddress(tx.From),
		common.HexToAddress(tx.To),
		EtherToWei(tx.Value), input)

	if tx.Additional_data, err = etx.MarshalJSON(); err != nil {
		L4g.Error("ethereum BuildTx faild, message:%s", err.Error())
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
		L4g.Error("sign Transaction error:%s", err.Error())
		return nil, err
	}

	tx.Additional_data, err = signedTx.MarshalJSON()
	if err != nil {
		L4g.Error("SignTx-MarshalJSON error:%s", err.Error())
		return nil, err
	}

	L4g.Trace("eth sign Transaction: ", signedTx.String())
	tx.State = types.Tx_state_Signed
	return tx.Additional_data, nil
}

// send signed transaction
func (self *Client) SendSignedTx(txByte []byte, tx *types.Transfer) error {
	var signedTx etypes.Transaction
	err := signedTx.UnmarshalJSON(txByte)
	if err != nil {
		L4g.Error("SendSignedTx UnmarshalJSON: %s", err.Error())
		return err
	}

	if err := self.c.SendTransaction(context.TODO(), &signedTx); err != nil {
		L4g.Trace("Transaction gas * price + value = %d", signedTx.Cost().Uint64())
		L4g.Error("SendTransaction error: %s", err.Error())
		return err
	}

	L4g.Trace("SendSignedTx Tx information:%s", tx.String())
	tx.State = types.Tx_state_pending
	return nil
}
