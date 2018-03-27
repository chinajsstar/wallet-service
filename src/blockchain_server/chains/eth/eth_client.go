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
	"fmt"
	"sort"
	"github.com/ethereum/go-ethereum"
)

type Client struct {
	c                *ethclient.Client
	addresses        SortedString
	pendingTxChannel chan *types.Transfer
	// TODO:由于blockheight被两个routine使用, 应该加上同步!!
	blockHeight      uint64
	scanblock        uint64
	rctChannel       types.RechargeTxChannel

	ctx					context.Context
	ctx_canncel			context.CancelFunc
}

type SortedString []string

func insertByOrder(strSorted SortedString, s string) {
	if !sort.StringsAreSorted(strSorted) {
		sort.Strings(strSorted)
	}

	index := sort.SearchStrings(strSorted, s)
	if strSorted[index]!=s {
		strSorted = append(strSorted[:index], append([]string{s}, strSorted[index:]...)...)
	}
}
func (self SortedString)containString(s string) bool {
	if index:=sort.SearchStrings(self, s); index>=0 && index<len(self) && self[index]==s {
		return true
	}
	return false
}

func NewClient(rctChannel types.RechargeTxChannel) (*Client, error) {
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
	c := &Client{c:client, rctChannel:rctChannel}
	c.init()

	return c, nil
}

func  (self *Client) init() {
	self.addresses = make([]string, 0, 1024)
	self.blockHeight = self.Blocknumber()
	self.scanblock = config.GetConfiger().Clientconfig[types.Chain_eth].Start_scan_Blocknumber
	self.ctx, self.ctx_canncel = context.WithCancel(context.Background())
}



func (self *Client) Start() error {
	if err := self.beginSubscribeBlockHaders(); err!=nil {
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

func (self *Client)Blocknumber() uint64 {
	blocknumber, err := self.c.BlockByNumber(context.TODO(), nil)
	if err!=nil {
		return 0
	}
	return blocknumber.NumberU64()
}

func (self *Client)Tx(tx_hash string)(*types.Transfer, error) {
	tx, blocknumber, err := self.c.TransactionByHash(self.ctx, common.HexToHash(tx_hash))
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

	return txToTx(tx, big_blocknumber.Uint64(), self.Blocknumber()), nil
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
		for {
			select {
			case <-self.ctx.Done() : {
				l4g.Error("subscribe block header exit loop for:%s", self.ctx.Err().Error())
				subscription.Unsubscribe()
				close(header_chan)
				goto endfor
			}
			case header := <- header_chan: {
				// TODO: 订阅到新区块, 可以搞一些事情
				self.blockHeight = header.Number.Uint64()
			}
			}
		}

	endfor:
		l4g.Trace("Will stop subscribe block header!")
	}(header_chan)

	return nil
}


//TxRecipt(ctx context.Context, tx_hash string)(*types.Transfer, error)
//func (self *Client)Blocknumber(ctx context.Context) (uint64, error) {
//	n, err := self.c.BlockByNumber(ctx, nil)
//	if err!=nil {
//		return 0, err
//	}
//	return n.NumberU64(), nil
//}

func (self *Client) beginScanBlock() error {
	if nil==self.addresses || len(self.addresses)==0 {
		return fmt.Errorf("address length is 0, please add cared address")
	}

	go func() {
		for {
			if self.blockHeight <= self.scanblock {
				time.Sleep(time.Second)
				continue
			}

			block, err := self.c.BlockByNumber(self.ctx, big.NewInt(int64(self.scanblock)))

			//block.Time()

			if err!= nil {
				l4g.Error("get block error, stop scanning block, message:%s", err.Error())
				goto endfor
			}

			l4g.Trace("scaning block :%d", block.NumberU64())

			txs := block.Transactions()

			for _, tx := range txs {
				l4g.Trace( "transaction onblock : %d, tx information:%s", block.NumberU64(),
					tx.String())

				time.Sleep(time.Second)

				if !self.addresses.containString(tx.To().String()) {
					continue
				}
				// TODO : check if tx.TO().String() hax prefix "0x" originally
				l4g.Trace("Transaction to address in wallet storage: %s", tx.To().String())
				self.rctChannel <- &types.RechargeTx{types.Chain_eth, txToTx(tx,  block.NumberU64(), uint64(self.blockHeight))}
			}

			self.scanblock++

			select {
			case <-self.ctx.Done(): {
				l4g.Trace("stop scaning blocks! for message:%s", self.ctx.Err().Error())
				goto endfor
			}
			default: {
			}
			}

		}
	endfor:
	}()
	return nil
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

func (self *Client)InsertCareAddress(address []string) {
	// TODO : 这个在修改address时, 应该使用同步的方式
	if self.addresses==nil {
		self.addresses = make([]string, 0, 1024)
	}

	if len(self.addresses)==0 {
		for _, value := range address {
			self.addresses = append(self.addresses, value)
		}
		sort.Strings(self.addresses)
		return
	}

	for _, value := range address {
		if !self.addresses.containString(value) {
			insertByOrder(self.addresses, value)
		}
	}
}

func (self *Client) Stop() {
	self.ctx_canncel()
	config.GetConfiger().Clientconfig[types.Chain_eth].Start_scan_Blocknumber = self.blockHeight
	config.GetConfiger().Save(types.Chain_eth)
}
