package eth

import (
	"blockchain_server/types"
	l4g "github.com/alecthomas/log4go"
	"time"
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"blockchain_server/conf"
	"github.com/ethereum/go-ethereum/crypto"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"sort"
	"github.com/ethereum/go-ethereum"
	"sync"
	"sync/atomic"
	"strings"
	"fmt"
	"blockchain_server/chains/eth/token"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type Client struct {
	c                *ethclient.Client
	addresslist      SortedString
	blockHeight      uint64
	scanblock        uint64
	rctChannel       types.RechargeTxChannel
	ctx				 context.Context
	ctx_canncel		 context.CancelFunc

	address_locker   *sync.RWMutex

	rechargelist	 types.RechargeTxChannel

	confirm_count	 uint16

	// TODO:后面支持动态增加ERC20代币的时候, 这个map应该改成协程同步的!
	// TODO:这个map中保存的token, 实际上是config中的指针, 并通过abi更新了相关字段信息
	// TODO:这个token应该要重新写会到config文件中去
	tokens 			map[string]* types.Token
	erc20ABI		abi.ABI
	erc20Token		map[string]*token.Token
}

type SortedString []string

func insertByOrder(strSorted SortedString, s string) {
	if !sort.StringsAreSorted(strSorted) {
		sort.Strings(strSorted)
	}
	s = strings.ToLower(s)
	index := sort.SearchStrings(strSorted, s)

	if index>=len(strSorted) || strSorted[index]!=s {
		strSorted = append(strSorted[:index], append([]string{s}, strSorted[index:]...)...)
	}
}
func (self SortedString)containString(s string) bool {
	s = strings.ToLower(s)
	if index:=sort.SearchStrings(self, s); index>=0 && index<len(self) && self[index]==s {
		return true
	}
	return false
}

func (self *Client)lock() {
	self.address_locker.Lock()
}

func (self *Client)unlock() {
	self.address_locker.Unlock()
}

func NewClient() (*Client, error) {
	//ctx, _ := context.WithTimeout(context.Background(), time.Second * 5)
	//rpc_client, err := rpc.DialContext(ctx, config.GetConfiger().Clientconfig[types.Chain_eth].RPC_url)
	//if err!=nil {
	//	return nil, err
	//}
	//client := ethclient.NewClient(rpc_client)

	client, err := ethclient.Dial(config.GetConfiger().Clientconfig[types.Chain_eth].RPC_url)
	if nil!=err {
		l4g.Error("create eth client error! message:%s", err.Error())
		return nil, err
	}
	c := &Client{c:client}
	c.init()

	return c, nil
}

func  (self *Client) init() {
	self.addresslist = make([]string, 0, 512)
	atomic.StoreUint64(&self.blockHeight, self.getBlockHeight())

	configer := config.GetConfiger().Clientconfig[types.Chain_eth]

	self.scanblock = configer.Start_scan_Blocknumber
	self.confirm_count = uint16(configer.TxConfirmNumber)

	self.ctx, self.ctx_canncel = context.WithCancel(context.Background())
	self.address_locker = new(sync.RWMutex)
	// 如果为0, 则默认从最新的开始扫描
	if self.scanblock==0 {
		self.scanblock = self.blockHeight
	}

	self.tokens = make(map[string]*types.Token, 128)
	l4g.Trace("----------------Init %s Token-----------------")
	// create configed tokens
	self.erc20Token = make(map[string]*token.Token)

	for symbol, tk := range configer.Tokens {
		tmp_tk, err := token.NewToken(common.HexToAddress(tk.Address), self.c)

		if err!= nil {
			l4g.Trace("Create Token instance(%s) error:%s", symbol, err.Error())
			continue
		}

		tk.Name, err = tmp_tk.Name(nil)
		if err!=nil { continue }
		tk.Decimals, _ = tmp_tk.Decimals(nil)
		tk.Symbol, _ = tmp_tk.Symbol(nil)

		self.erc20Token[tk.Name] = tmp_tk
		l4g.Trace("Token information: {%s}", tk.String())

		self.tokens[strings.ToLower(tk.Address)] = tk
	}

	var err error
	if self.erc20ABI, err = abi.JSON(strings.NewReader(token.TokenABI)); err!=nil {
		// TODO:here may should print error message and exit process
		l4g.Error("Create erc20 token abi error:%s", err.Error())
	}
}

func (self *Client) Start() error {
	if err := self.beginSubscribeBlockHaders(); err!=nil {
		self.Stop()
		return err
	}

	if err := self.beginScanBlock(); err!=nil {
		self.Stop()
		return err
	}
	return nil
}

func (self *Client) Name() string {
	return types.Chain_eth
}

func (self *Client) NewAccount()(*types.Account, error) {
	return NewAccount()
}


func (self *Client)SubscribeRechageTx(txRechChannel types.RechargeTxChannel) {
	self.rctChannel = txRechChannel
}

func isTxToken(tx *types.Transfer) bool {
	if tx==nil {return false}
	return tx.Token!=nil
}

// from is a crypted private key
func (self *Client)SendTx(chiperKey string, tx *types.Transfer) error {
	key, err := ParseChiperkey(chiperKey)
	if err!= nil {
		return err
	}
	tx.From = crypto.PubkeyToAddress(key.PublicKey).String()

	if isTxToken(tx) {
		tk := self.erc20Token[tx.Token.Name]
		if tk==nil {
			return fmt.Errorf("Not supported token(%s) Transaction:%s", tx.Token.Name, tx.String() )
		}
		opts := bind.NewKeyedTransactor(key)
		tmpTx, err := tk.Transfer(opts, common.HexToAddress(tx.To), big.NewInt(10))
		if err!=nil {
			l4g.Trace("SendTransactrion error:%s", err.Error())
			return err
		}
		err = self.updateTxWithTx(tx, tmpTx)
		if err!=nil {
			l4g.Trace("updateTxWithTx error, message:%s", err.Error())
		}
		l4g.Trace("Tx information:%s", tx.String())
	} else {
		etx, err := self.newEthTx(tx)
		if err!=nil { return err }

		// chainID := big.NewInt(CHAIN_ID)
		// signer := types.NewEIP155Signer(chainID)
		signer := etypes.HomesteadSigner{}
		signedTx, err := etypes.SignTx(etx, signer, key)

		if err!=nil {
			l4g.Error("sign Transaction error:%s", err.Error())
			return err
		}

		l4g.Trace("eth Transaction information:%s", etx.String())

		tx.Tx_hash = signedTx.Hash().String()
		tx.State = types.Tx_state_unkown
		if err:=self.c.SendTransaction(context.TODO(), signedTx); err!=nil {
			l4g.Trace("Transaction gas * price + value = %d",  signedTx.Cost().Uint64())
			l4g.Error("SendTransaction error: %s", err.Error())
			return err
		}
	}
	tx.State = types.Tx_state_commited
	return nil
}

func (self *Client)GetBalance(addstr string, tokenname *string) (uint64, error){
	address := common.HexToAddress(addstr)
	var bl uint64 = 0
	if tokenname==nil {
		if balance, err := self.c.BalanceAt(self.ctx, address, nil); err!=nil {
			return 0, err
		} else { bl = balance.Uint64() }
	} else {
		tmp_Token := self.tokens[*tokenname]
		if nil==tmp_Token {
			return 0, fmt.Errorf("GetBalance, Not Supported assert type!")
		}
		tk := self.erc20Token[tmp_Token.Address]
		if nil==tk {
			return 0, fmt.Errorf("GetBalance, Not Supported assert type!")
		}
		if balance, err := tk.BalanceOf(nil, address); err!=nil {
			return 0, fmt.Errorf("Get Token[%s] Balance of %s error, message:%s",
				*tokenname, addstr, err.Error()  )
		}else {bl = balance.Uint64()}
	}
	return bl, nil

}

// TODO: 创建出来的tx,如果是token的话, 交易总是失败, 目前发送Token不用这个方法, 后面再检查
func (self *Client)newEthTx(tx *types.Transfer) (*etypes.Transaction, error) {

	from  := common.HexToAddress(tx.From)
	to 	  := common.HexToAddress(tx.To)
	value := big.NewInt(int64(tx.Value))
	var input []byte = nil

	// if tx.Token field not nil, then create erc20 token transaction
	// for erc20 Token Trasanction, 'to' address is contract address
	// and reciver address is packed to []byte 'input'
	if tx.Token!=nil {
		var err error
		if input, err = self.erc20ABI.Pack("transfer", to, value); err!=nil {
			l4g.Error("Create erc20 transaction error, message:%s", err.Error())
			return nil, err
		}

		value = value.SetInt64(0)
		to = common.HexToAddress(tx.Token.Address)
	}

	// gas limit
	msg := ethereum.CallMsg{From: from, To: &to, Value:value, Data: input}
	gaslimit, err := self.c.EstimateGas(context.TODO(), msg)
	if err!=nil { return nil, err }

	// gas price
	gasprice, err := self.c.SuggestGasPrice(context.TODO())
	if err!=nil {return nil, err}

	// nonce
	nonce, err := self.c.PendingNonceAt(context.TODO(), common.HexToAddress(tx.From))
	if nil!= err {
		return nil, err
	}

	return etypes.NewTransaction(nonce, to, value, gaslimit, gasprice, input), nil
}

func (self Client) getBlockHeight() uint64 {
	block, err := self.c.BlockByNumber(context.TODO(), nil)
	if err!=nil {
		return 0
	}
	atomic.StoreUint64(&self.blockHeight, block.NumberU64())
	return atomic.LoadUint64(&self.blockHeight)
}

func (self *Client) BlockHeight() uint64 {
	return atomic.LoadUint64(&self.blockHeight)
}

//func (self *Client)updateTxWithTx(transfer *types.Transfer, transaction *etypes.Transaction) bool {
//	var state types.TxState
//
//	if inblock ==0 {
//		state = types.Tx_state_pending
//	} else if lastblock-inblock >= config.GetConfiger().Clientconfig[types.Chain_eth].TxConfirmNumber {
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
//		Gase :             tx.Gas(),
//		Gaseprice:         tx.GasPrice().Uint64(),
//		Total :            tx.Cost().Uint64(),
//		blockNumber:       inblock,
//		ConfirmatedHeight: lastblock,
//		State:             state,
//	}
//}

// 如果是合约的交易, 需要使用TransactionReceipt来获取合约地址名称等!
func (self *Client)Tx(tx_hash string)(*types.Transfer, error) {
	hash := common.HexToHash(tx_hash)
	tx, err := self.c.TransactionByHash(self.ctx, hash)

	if err!= nil {
		if err==ethereum.NotFound {
			return nil, types.NewTxNotFoundErr(tx_hash)
		}
		return nil, err
	}

	height := atomic.LoadUint64(&self.blockHeight)
	return self.toTx(tx, height),nil
	}

func (self *Client) beginSubscribeBlockHaders() error {
	header_chan := make(chan *etypes.Header)
	subscription, err := self.c.SubscribeNewHead(self.ctx, true, header_chan)

	if err!=nil || nil==subscription {
		l4g.Error("subscribefiler error:%s\n", err.Error())
		return err
	}
	l4g.Trace("eth Subscribe new block header, begin!")

	go func(header_chan chan *etypes.Header) {
		defer subscription.Unsubscribe()
		for {
			select {
			case <-self.ctx.Done() : {
				l4g.Error("subscribe block header exit loop for:%s", self.ctx.Err().Error())
				close(header_chan)
				goto endfor
			}
			case header := <- header_chan: {
				atomic.StoreUint64(&self.blockHeight, header.Number.Uint64())
				l4g.Trace("get new block(%s) height:%d",
					header.Hash().String(), atomic.LoadUint64(&self.blockHeight))

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
		for {
			height := atomic.LoadUint64(&self.blockHeight)

			if nil==self.addresslist || len(self.addresslist)==0 ||
				height <= self.scanblock ||
				self.rctChannel == nil {
				l4g.Trace("Recharge channel is nil, or block height(%d)==scanblock(%d), or addresslit is empty!", height, self.scanblock)
				time.Sleep(time.Second * 5)
				continue
			}

			block, err := self.c.BlockByNumber(self.ctx, big.NewInt(int64(self.scanblock)))

			if err!= nil {
				l4g.Error("get block error, stop scanning block, message:%s", err.Error())
				goto endfor
			}

			l4g.Trace("scaning block :%d", block.NumberU64())

			txs := block.Transactions()

			for _, tx := range txs {
				l4g.Trace( "transaction onblock : %d, tx information:%s", block.NumberU64(), tx.String())

				to := tx.To()
				if to==nil { continue }

				if tk:=self.tokens[to.String()]; tk!=nil || self.hasAddress(to.String()) {
					tmp_tx := self.toTx(tx, block.NumberU64())
					self.rctChannel <- &types.RechargeTx{types.Chain_eth, tmp_tx, nil}
				}
			}

			self.scanblock++

			select {
			case <-self.ctx.Done(): {
				l4g.Trace("stop scaning blocks! for message:%s", self.ctx.Err().Error())
				goto endfor
			}
			default: {
				time.Sleep(time.Second * 1)
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
	return self.addresslist.containString(address)
}

func (self *Client) updateTxWithReceipt(tx *types.Transfer) error {
	height := atomic.LoadUint64(&self.blockHeight)

	receipt, err := self.c.TransactionReceipt(self.ctx, common.HexToHash(tx.Tx_hash))
	if err!=nil {
		l4g.Error("Update Transaction(%s) with Receipt error : %s", tx.Tx_hash, err.Error())
		return err
	}

	tx.GasUsed = receipt.GasUsed
	//if receipt.Status==etypes.ReceiptStatusFailed ||
	if tx.GasUsed > tx.Gase {
		tx.State = types.Tx_state_unconfirmed
	} else {
		if height - tx.InBlock > config.GetConfiger().Clientconfig[types.Chain_eth].TxConfirmNumber {
			tx.State = types.Tx_state_confirmed
			tx.ConfirmatedHeight = height
		} else {
			tx.State = types.Tx_state_mined
		}
	}
	return nil
}

// 使用srcTx更新destTx, 如果为空, 则使用TransactionByHash获取最新的tx
// 如果使用了已经存在的Transaction来更新Transfer, 程序不会从网络去获取Transaction, 不会检查txhash是否存在!
// 则无法判断Transaction是否被分叉的问题
func (self *Client) updateTxWithTx(destTx *types.Transfer, srcTx *etypes.Transaction) error {
	height := atomic.LoadUint64(&self.blockHeight)
	// if input srcTx is nil, try to get transaction from network node
	if srcTx ==nil {
		var err error
		srcTx, err = self.c.TransactionByHash(self.ctx, common.HexToHash(destTx.Tx_hash))
		if err!=nil {
			if err == ethereum.NotFound {
				// 如果之前的状态为已经上块, 现在又找不到了, 说明可能被分叉了
				// 把状态设置为unconfirmed
				if destTx.State==types.Tx_state_mined || destTx.State==types.Tx_state_confirmed {
					destTx.State=types.Tx_state_unconfirmed
				} else {
					// 可能为交易才被提交, 这里不做任何处理
					return types.NewTxNotFoundErr(destTx.Tx_hash)
				}
			}
		}
	}

	var state types.TxState
	if srcTx.Inblock==0 {
		state = types.Tx_state_pending
	} else if uint16(height-srcTx.Inblock) > self.confirm_count {
		state = types.Tx_state_confirmed
	} else {
		state = types.Tx_state_mined
	}

	to := srcTx.To()
	if to!=nil {
		// 则检查to是否为已经注册的ERC20 Token合约地址
		// 如果确定为合约地址, 需要把destTx.to设置为接收代币的地址, 并为其设置token成员

		// the data field like this,
		// for a token Transaction, the 32-73 bytes of Data feild means reciver's address, like following input data:
		// 0xa9059cbb000000000000000000000000 498d8306dd26ab45d8b7dd4f07a40d2c744f54bc 000000000000000000000000000000000000000000000000000000000000000a
		to_string := strings.ToLower(to.String())
		if tk := self.tokens[to_string]; tk!=nil {
			destTx.To = fmt.Sprintf("0x%x", srcTx.Data()[16:36])
			// TODO:why use the following express cause process crush!!
			// destTx.To = "0x" + hex.EncodeToString(recevier_to)
			destTx.Token = tk
		} else { // else just set destTx.To as reciver's address
			destTx.To = to.String()
		}
	}

	destTx.Tx_hash = srcTx.Hash().String()
	destTx.From = srcTx.From()
	destTx.Value = srcTx.Value().Uint64()
	destTx.Gase = srcTx.Gas()
	destTx.Gaseprice = srcTx.GasPrice().Uint64()
	destTx.Total = srcTx.Cost().Uint64()
	destTx.InBlock = srcTx.Inblock
	destTx.State = state

	if state==types.Tx_state_confirmed {
		if err := self.updateTxWithReceipt(destTx); err==nil {
			if destTx.State==types.Tx_state_confirmed {
				destTx.ConfirmatedHeight = height
			}
		}
	}

	return nil
}

func (self *Client) UpdateTx(tx *types.Transfer) error {
	if nil== tx || len(tx.Tx_hash)==0 {return fmt.Errorf("Invalid paramater!")}

	if tx.State==types.Tx_state_confirmed {
		if err:=self.updateTxWithReceipt(tx); err!=nil {
			return err
		}
		return nil
	} else { // tx.State==types.Tx_state_commited, Tx_state_pending, tx_state_unkown, tx_state_notfound
		if err:=self.updateTxWithTx(tx, nil); err!=nil {
			return err
		}
		return nil
	}

}

func (self *Client) toTx(tx *etypes.Transaction, height uint64) *types.Transfer {
	tmpTx := &types.Transfer{Tx_hash:tx.Hash().String()}
	self.updateTxWithTx(tmpTx, tx)

	return tmpTx
}

func (self *Client) InsertRechageAddress(address []string) {
	self.lock()
	defer self.unlock()

	if self.addresslist ==nil {
		self.addresslist = make([]string, 0, 512)
	}

	if len(self.addresslist)==0 {
		for _, value := range address {
			value = strings.ToLower(value)
			self.addresslist = append(self.addresslist, value)
		}
		sort.Strings(self.addresslist)
		return
	}

	for _, value := range address {
		if !self.addresslist.containString(value) {
			insertByOrder(self.addresslist, value)
		}
		l4g.Trace(value)
	}
}

func (self *Client) Stop() {
	self.ctx_canncel()

	atomic.StoreUint64(&config.GetConfiger().Clientconfig[types.Chain_eth].Start_scan_Blocknumber, atomic.LoadUint64(&self.blockHeight))

	config.GetConfiger().Save()

	//close(self.rctChannel)
	//close(self.pendingTxChannel)

}

