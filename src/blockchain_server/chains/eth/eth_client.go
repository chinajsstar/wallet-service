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
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum"
)

type Client struct {
	c                *ethclient.Client
	addresses        SortedString
	pendingTxChannel chan *types.Transfer

	lastBlocknumber 	uint64
	beginScanBlock		uint64

	ctx 				context.Context
	cancelfun 			context.CancelFunc
	ctxCancel 			context.CancelFunc
	closeChannel		chan bool

	//subscribeFeed 		event.Fee
	// d
	rctChannel			types.RechargeTxChannel
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
	return self[sort.SearchStrings(self, s)]==s
}

func NewClient() (*Client, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second * 5)

	rpc_client, err := rpc.DialContext(ctx, config.GetConfiger().Clientconfig[types.Chain_eth].RPC_url)
	if err!=nil {
		return nil, err
	}
	client := ethclient.NewClient(rpc_client)
	//rpc_client, err := ethclient.Dial(config.GetConfiger().Clientconfig[types.Chain_eth].RPC_url)
	//if nil!=err {
	//	l4g.Error("create eth client error! message:%s", err.Error())
	//	return nil, err
	//}
	c := &Client{
			c:				  client,
			closeChannel:     make(chan bool),
			pendingTxChannel: make(chan *types.Transfer),
			lastBlocknumber:  0}

	c.ctx, c.cancelfun  = context.WithCancel(context.Background())

	c.beginScanBlock = config.GetConfiger().Clientconfig[types.Chain_eth].Start_scan_Blocknumber
	c.addresses = make([]string, 0, 1024)
	return c, nil
}

//func (self *Client) SubscribeRechargeTx(rctChannel types.RechargeTxChannel) {
//	self.rctChannel = rctChannel
//	//return self.subscribeFeed.Subscribe(rctChannel)
//}

func (self *Client) Start(rcTxChannel types.RechargeTxChannel) error {
	//if self.txChannel==nil {
	//	fmt.Println( "self.txchannel is nil")
	//} else {
	//	fmt.Println( "self.txchannel is not nil")
	//}
	go self.subscribeNewBlockheader()
	go self.StartScanBlock(rcTxChannel)
	return nil
}

func (self *Client) Name() string {
	return types.Chain_eth
}

func (self *Client) NewAccount()(*types.Account, error) {
	return NewAccount()
}

// from is a crypted private key
func (self *Client)SendTx(ctx context.Context,  chiperKey string, tx *types.Transfer) error {
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

func (self *Client)blocknumber() uint64 {
	blocknumber, err := self.c.BlockByNumber(context.TODO(), nil)
	if err!=nil {
		return 0
	}
	return blocknumber.NumberU64()
}

func (self *Client)Tx(ctx context.Context, tx_hash string)(*types.Transfer, error) {
	tx, blocknumber, err := self.c.TransactionByHash(ctx, common.HexToHash(tx_hash))
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

	return txToTx(tx, big_blocknumber.Uint64(), self.blocknumber()), nil
}

func (self *Client)subscribeNewBlockheader() {
	header_ch := make(chan *etypes.Header)
	ctx, _ := context.WithCancel(self.ctx)
	subscription, err := self.c.SubscribeNewHead(ctx, true, header_ch)
	if err!=nil || nil==subscription {
		fmt.Printf("subscribefiler error:%s\n", err.Error())
		return
	}

	fmt.Printf("subcribe newblock ok, listenling...\n")

	tobreak := false
	for {
		select {
		case <- ctx.Done():
			fmt.Printf("error happened:%v\n", ctx.Err())
			tobreak = true
		case header := <-header_ch: {
			if self.beginScanBlock == header.Number.Uint64() {

			}
			self.lastBlocknumber = header.Number.Uint64()
			l4g.Trace("new block header : blocknumber : %d, hashvalue:%s\n",
				self.lastBlocknumber, header.Hash().String())
		}
		}

		if tobreak {
			break
		}
	}
	subscription.Unsubscribe()
}


//TxRecipt(ctx context.Context, tx_hash string)(*types.Transfer, error)
func (self *Client)Blocknumber(ctx context.Context) (uint64, error) {
	n, err := self.c.BlockByNumber(ctx, nil)
	if err!=nil {
		return 0, err
	}
	return n.NumberU64(), nil
}

func (self *Client) StartScanBlock(rtc types.RechargeTxChannel) error {
	if nil==self.addresses || len(self.addresses)==0 {
		return fmt.Errorf("address length is 0, please add care address")
	}
	var close = false
	for !close {
		ctx, _ := context.WithCancel(self.ctx)

		big_bigenblock := big.NewInt(int64(self.beginScanBlock))
		block, err := self.c.BlockByNumber(ctx, big_bigenblock)

		if err!= nil {
			l4g.Error("MinitorAddress get block error, message:%s", err.Error())
			break
		}

		txs := block.Transactions()

		for _, tx := range txs {
			l4g.Trace( "transaction onblock : %d, tx information:%s", block.NumberU64(),
				tx.String())

			if !self.addresses.containString("0x" + tx.To().String()) {
				continue
			}
			l4g.Trace("Transaction to address in wallet storage: %s", tx.To().String())

			rtc <- &types.RechargeTx{types.Chain_eth, txToTx(tx, big_bigenblock.Uint64(), uint64(self.lastBlocknumber))}
		}

		self.lastBlocknumber++
		select {
		case <-self.ctx.Done(): {
			close = true
		}
		case close = <-self.closeChannel: {
			l4g.Trace("stop Scan blocks, close channel get true value!")
		}
		default: {
			time.Sleep(time.Second)
		}
		}

	}
	return nil
}

func txToTx(tx *etypes.Transaction, blocknumber uint64, lastnumber uint64) *types.Transfer {
	var state types.TxState

	if blocknumber==0 {
		state = types.Tx_state_pending
	} else if lastnumber-blocknumber >= config.GetConfiger().Clientconfig[types.Chain_eth].CoinConfirmNumber {
		state = types.Tx_state_confirmed
	} else {
		state = types.Tx_state_mined
	}

	return &types.Transfer{
		Tx_hash:      tx.Hash().String(),
		From:		  tx.From(),
		To :          tx.To().String(),
		Value:        tx.Value().Uint64(),
		Gase :        tx.Gas(),
		Gaseprice:    tx.GasPrice().Uint64(),
		Total :       tx.Cost().Uint64(),
		OnBlock:      blocknumber,
		PresentBlock: lastnumber,
		State:        state,
	}
}

func (self *Client)InsertCareAddress(address []string) {
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

func (self *Client) Stop(ctx context.Context,  duration time.Duration) {
	self.closeChannel <- true
	self.ctxCancel()

	config.GetConfiger().Clientconfig[types.Chain_eth].Start_scan_Blocknumber = self.lastBlocknumber
}
