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
	"blockchain_server/utils"
	"sort"
	"github.com/ethereum/go-ethereum"
	"sync"
	"sync/atomic"
	"strings"
)

type Client struct {
	c                *ethclient.Client
	addresslist      SortedString
	//pendingTxChannel chan *types.Transfer
	// TODO:由于blockheight被两个routine使用, 应该加上同步!!
	blockHeight      uint64
	scanblock        uint64
	rctChannel       types.RechargeTxChannel
	ctx				 context.Context
	ctx_canncel		 context.CancelFunc

	address_locker   *sync.RWMutex

	rechargelist	 types.RechargeTxChannel
}

type SortedString []string

func insertByOrder(strSorted SortedString, s string) {
	if !sort.StringsAreSorted(strSorted) {
		sort.Strings(strSorted)
	}
	s = strings.ToLower(s)
	index := sort.SearchStrings(strSorted, s)
	if strSorted[index]!=s {
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
	self.scanblock = config.GetConfiger().Clientconfig[types.Chain_eth].Start_scan_Blocknumber
	self.ctx, self.ctx_canncel = context.WithCancel(context.Background())
	self.address_locker = new(sync.RWMutex)
	// 如果为0, 则默认从最新的开始扫描
	if self.scanblock==0 {
		self.scanblock = self.blockHeight
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

// from is a crypted private key
func (self *Client)SendTx(chiperKey string, tx *types.Transfer) error {
	key, err := ParseChiperkey(chiperKey)
	if err!= nil {
		return err
	}
	tx.From = crypto.PubkeyToAddress(key.PublicKey).String()
	etx, err := self.newEthTx(tx)
	if err!=nil {
		return err
	}

	//l4g.Trace("tx.gas=%d, tx.gasprice=%d, tx.value=%d, tx.cost=%d\n", etx.Gas(), etx.GasPrice().Uint64(), etx.Value().Uint64(), etx.Cost().Uint64())

	//signer := etypes.NewEIP155Signer()

	//l4g.Trace("transaction from address:%s\n", crypto.PubkeyToAddress(key.PublicKey).String())

	//types.SignTx(tx, types.NewEIP155Signer(chainID), key.PrivateKey)
	// return types.SignTx(tx, types.HomesteadSigner{}, key.PrivateKey)

	chainId := big.NewInt(15)
	signedTx, err := etypes.SignTx(etx, etypes.NewEIP155Signer(chainId), key)

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
	tx.State = types.Tx_state_commited
	return nil
}


func (self *Client)newEthTx(tx *types.Transfer) (*etypes.Transaction, error) {
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	gaslimit := uint64(0x2fefd8)
	gasprice, err := self.c.SuggestGasPrice(context.TODO())
	if err!=nil {
		return nil, err
	}

	address := common.HexToAddress(tx.From)
	nonce, err := self.c.PendingNonceAt(context.TODO(), address)
	if nil!= err {
		return nil, err
	}

	//l4g.Trace("^^^^^^^^^^^^^^^^^^\n" +
	//	"getPendingNonce at (%s) returns nonce : %d\n" +
	//	"^^^^^^^^^^^^^^^^^^",
	//	tx.From, nonce)

	//tx := types.NewTransaction(nonce, toaddress, amount, uint64(gaslimit), gasprice, nil)
	//fmt.Printf("tx.amount ; %d, tx.realamount :%d\n", tx.Value, big.NewInt(int64(tx.Value)))
	return etypes.NewTransaction(nonce, address, big.NewInt(int64(tx.Value)), gaslimit, gasprice, nil), nil
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

func (self *Client)Tx(tx_hash string)(*types.Transfer, error) {
	hash := common.HexToHash(tx_hash)
	tx, blocknumber, err := self.c.TransactionByHash(self.ctx, hash)
	if err!= nil {
		if err==ethereum.NotFound {
			return nil, types.NewTxNotFoundErr(tx_hash)
		}
		return nil, err
	}

	var big_blocknumber *big.Int = nil
	if blocknumber!=nil {
		big_blocknumber , _ = utils.Hex_string_to_big_int(*blocknumber)
	} else {
		big_blocknumber = big.NewInt(0)
	}

	tmp_tx := txToTx(tx, big_blocknumber.Uint64(), atomic.LoadUint64(&self.blockHeight))

	if tmp_tx.State==types.Tx_state_confirmed {
		// 当状态变更为确认时, 需要确认是否交易已经上链, 但是由于矿工费太少,
		// 交易并没有生效, 并且还浪费了矿工费, 这种情况一般不会发生!
		rctTx, err := self.c.TransactionReceipt(self.ctx, hash)
		if err!=nil {
			l4g.Error("Query Receipt Transaction error message:%s", err.Error())

		} else {
			// gasused > gas , 矿工费被白收, 交易失效!
			if rctTx.GasUsed >= tx.Gas() {
				tmp_tx.State = types.Tx_state_unconfirmed
			}
		}
	}

	return tmp_tx, nil

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
				l4g.Trace( "transaction onblock : %d, tx information:%s", block.NumberU64(),
					tx.String())

				l4g.Trace("recharge tx to(%s)", tx.To().String())
				if !self.hasAddress(tx.To().String()) {
					continue
				}

				// TODO : check if tx.TO().String() hax prefix "0x" originally
				l4g.Trace("Transaction to address in wallet storage: %s", tx.To().String())
				self.rctChannel <- &types.RechargeTx{types.Chain_eth, txToTx(tx,  block.NumberU64(), height), nil}
			}

			self.scanblock++

			select {
			case <-self.ctx.Done(): {
				l4g.Trace("stop scaning blocks! for message:%s", self.ctx.Err().Error())
				goto endfor
			}
			default: {
				time.Sleep(time.Second * 5)
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

func txToTx(tx *etypes.Transaction, inblock uint64, lastblock uint64) *types.Transfer {
	var state types.TxState

	if inblock ==0 {
		state = types.Tx_state_pending
	} else if lastblock-inblock >= config.GetConfiger().Clientconfig[types.Chain_eth].TxConfirmNumber {
		state = types.Tx_state_confirmed
	} else {
		state = types.Tx_state_mined
	}

	return &types.Transfer{
		Tx_hash:      tx.Hash().String(),
		From:         tx.From(),
		To :          tx.To().String(),
		Value:        tx.Value().Uint64(),
		Gase :        tx.Gas(),
		Gaseprice:    tx.GasPrice().Uint64(),
		Total :       tx.Cost().Uint64(),
		OnBlock:      inblock,
		PresentBlock: lastblock,
		State:        state,
	}
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
