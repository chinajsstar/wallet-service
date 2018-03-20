package service

import (
	"blockchain_server/types"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"encoding/hex"
	"blockchain_server/utils"
	"crypto/ecdsa"
	"crypto/x509"
	"blockchain_server"
	"blockchain_server/chains/event"
	"context"
	"time"
)

const (
	max_once_account_number = 100
)

var (
	feed event.Feed
	//cmdHandler *NetcmdHandler
)

type TxCmdStateSubscribe_Channel chan *CmdTx
type Addresswatcher_Channel chan *types.Transfer

func init() {
	//cmdHandler = &NetcmdHandler{}
}

type ErrorInvalideParam struct {
	message string
}

type ErrSendTx struct {
	types.NetCmdErr
}

func newInvalidParamError(msg string) *ErrorInvalideParam{
	return &ErrorInvalideParam{message:msg}
}

//-32700	Parse error	Invalid JSON was received by the server.
//An error occurred on the server while parsing the JSON text.
//-32600	Invalid Request	The JSON sent is not a valid Request object.
//-32601	Method not found	The method does not exist / is not available.
//-32602	Invalid params	Invalid method parameter(s).
//-32603	Internal error	Internal JSON-RPC error.
//-32000 to -32099	Server error	Reserved for implementation-defined server-errors.
func newTxSendError(txCmd *CmdTx, message string, code int32)*ErrSendTx {
	return &ErrSendTx{types.NetCmdErr{
		Message:fmt.Sprintf("Send %s Transaction error:%s\ntx detail:%s",
		txCmd.Coin, message, txCmd.tx.String()), Code:code, Data:nil}}
}

func (e ErrSendTx) Error() string {
	return fmt.Sprintf("SendTransaction error, code:%d, message:%s", e.Code, e.Message)
}

func (self ErrorInvalideParam)Error() string {
	return self.message
}

type CmdTx struct {
	types.NetCmd
	chiperkey string
	tx        *types.Transfer
}

type CmdNewAccount struct {
	types.NetCmd
	amount			uint32
}

type CmdMonitorAddresses struct {
	types.NetCmd
	recall_url string
	addresses []string
}

type NetcmdHandler struct {
	txCmdChannel chan *CmdTx
	txCmdClose   chan bool
	clients      map[string]blockchain_server.ChainClient
}

// subscribe Sample
/*
func TestFeed(t *testing.TxStateString) {
	var feed Feed
	var done, subscribed sync.WaitGroup
	subscriber := func(i int) {
		defer done.Done()

		subchan := make(chan int)
		sub := feed.Subscribe(subchan)
		timeout := time.NewTimer(2 * time.Second)
		subscribed.Done()

		select {
		case v := <-subchan:
			if v != 1 {
				t.Errorf("%d: received value %d, want 1", i, v)
			}
		case <-timeout.C:
			t.Errorf("%d: receive timeout", i)
		}

		sub.Unsubscribe()
		select {
		case _, ok := <-sub.Err():
			if ok {
				t.Errorf("%d: error channel not closed after unsubscribe", i)
			}
		case <-timeout.C:
			t.Errorf("%d: unsubscribe timeout", i)
		}
	}

	const n = 1000
	done.Add(n)
	subscribed.Add(n)
	for i := 0; i < n; i++ {
		go subscriber(i)
	}
	subscribed.Wait()
	if nsent := feed.Send(1); nsent != n {
		t.Errorf("first send delivered %d times, want %d", nsent, n)
	}
	if nsent := feed.Send(2); nsent != 0 {
		t.Errorf("second send delivered %d times, want 0", nsent)
	}
	done.Wait()
}
*/

func (self *NetcmdHandler) SubscribeTxStateChange(txCmdStateChangeChannel *TxCmdStateSubscribe_Channel) *event.Subscription{
	subscribe := feed.Subscribe(txCmdStateChangeChannel)
	return &subscribe
}

func (self *NetcmdHandler) SubscribeInComeTxWithAddresses(coin string, addresses []string,
	) (error) {
	client := self.clients[coin]
	if(nil==client) {
		return fmt.Errorf("coin not supported:%s", coin)
	}
	client.InsertCareAddress(addresses)
	return nil
}

//func (self *NetcmdHandler) SubscribeIncomingTx (types. incomeTxChannel *Addresswatcher_Channel)  *event.Subscription {
//	subscribe := feed.Subscribe(incomeTxChannel)
//	return &subscribe
//}

func (self *NetcmdHandler)sendTx(txCmd *CmdTx) {
	l4g.Trace("------------SendTransaction begin------------")
	handler := self.clients[txCmd.Coin]

	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)

	err := handler.SendTx(ctx, txCmd.chiperkey, txCmd.tx)
	if nil!=err {
		// -32000 to -32099	Server error Reserved for implementation-defined server-errors.
		txCmd.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
		l4g.Error("Send Transaction error:%s", txCmd.Error.Message)
		feed.Send(txCmd)
	}

	txCmd.tx.State = types.Tx_state_commited
	feed.Send(txCmd)

	escapeloop := false
	// Transaction state change : committed->waite confirm number -> confirmed/unconfirmed ??
	for !escapeloop {
		time.Sleep(time.Second)
		tx, err := handler.Tx(ctx, txCmd.tx.Tx_hash)

		if err != nil {
			if nfe, ok := err.(*types.TxNotFoundErr); ok {
				l4g.Trace(nfe)
			} else {
				escapeloop = true
				txCmd.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
				l4g.Error(err.Error())
				feed.Send(txCmd)
			}
			continue
		}

		if tx.State != txCmd.tx.State {
			if tx.State==types.Tx_state_confirmed {
				escapeloop = true
			}

			l4g.Trace("Transaction state from:%s to %s, Transaction information:%s",
				types.TxStateString(txCmd.tx.State), types.TxStateString(tx.State), tx.String())

			txCmd.tx = tx
			feed.Send(txCmd)

		} else if tx.State==types.Tx_state_mined {

			if tx.PresentBlocknumber!=txCmd.tx.PresentBlocknumber {

				if tx.PresentBlocknumber-tx.PresentBlocknumber>tx.Confirmationsnumber {

					tx.State = types.Tx_state_confirmed
					escapeloop = true

					l4g.Trace("Transaction state from:%s to %s, Transaction information:%s",
						types.TxStateString(types.Tx_state_mined), types.TxStateString(types.Tx_state_confirmed),
							tx.String())
				}

				txCmd.tx = tx
				feed.Send(txCmd)
			}
		}
		/*
		switch tx.State {
		case types.Tx_state_commited: {
			txCmd.tx = tx
			feed.Send(txCmd)
		}
		case types.Tx_state_mined: {
			if txCmd.tx.State != tx.State {
				l4g.Trace("Transaction mined, Transaction infromation:%s", tx.String())
				txCmd.tx = tx
				feed.Send(txCmd)
			}
		}
		case types.Tx_state_confirmed: {
			if tx.PresentBlocknumber != txCmd.tx.PresentBlocknumber {
				if tx.PresentBlocknumber-tx.OnBlocknumber > tx.Confirmationsnumber {
					l4g.Trace("Transaction success done! Transaction information:%s", tx.String())
					escapeloop = true
				}
			} else if tx.PresentBlocknumber == txCmd.tx.PresentBlocknumber {
				break
			} else {
				message = "It's imporsible that: old PresentBlockNumber biger than new PresentBlockNumber, check this situation!"
				txCmd.Error = types.NewNetCmdErr(-32000, message, nil)
				l4g.Error(message)
				escapeloop = true
			}

			txCmd.tx = tx
			feed.Send(txCmd)
		}
		case types.Tx_state_unconfirmed: {
			txCmd.tx = tx
			l4g.Trace("Transaction is unconfimred! tx information:%s", tx.String())
			feed.Send(txCmd)
		}
		default: {
			message = fmt.Sprintf("Transaction state looks unusual, the state changed from:'%s' to:'%s' Transaction information:%s", types.TxStateString(txCmd.tx.State), types.TxStateString(tx.State))
			txCmd.Error = types.NewNetCmdErr(-32000, message, nil)
			l4g.Warn(message)
			escapeloop = true
			feed.Send(txCmd)
		}
		}
		*/
	}
	l4g.Trace("------------SendTransaction   end------------")
}

func (self *NetcmdHandler)StartClient(coin string, client blockchain_server.ChainClient,
	rctChannel types.RechargeTxChannel) error {
	if self.clients[coin]!=nil {
		return fmt.Errorf("already exist : %s", coin)
	}

	if self.clients==nil {
		self.clients = make(map[string]blockchain_server.ChainClient)
	}
	self.clients[coin] = client
	if err:=self.clients[coin].Start(rctChannel); err!=nil {
		fmt.Printf("start client:%s, error:%s\n", coin, err.Error())
		return err
	}
	return nil
}

func (self *NetcmdHandler)loopTransferCmdChan() {
	//var done, subscribed sync.WaitGroup
	transferCmdChan	:= make(chan *CmdTx)
	sub := feed.Subscribe(transferCmdChan)
	defer sub.Unsubscribe()

	closeloop := false

	for !closeloop {
		select {
		case txCmd := <- self.txCmdChannel: {
			go self.sendTx(txCmd)
		}
		case closeloop = <- self.txCmdClose: {
			break
		}
		}
	}
}

func (self *NetcmdHandler)NewAccounts(cmd *CmdNewAccount) ([]*types.Account, error) {
	if cmd.amount==0 || cmd.amount>max_once_account_number {
		return nil, newInvalidParamError(fmt.Sprintf("the count of account must >0 and <%d", max_once_account_number))
	}
	accs := make([]*types.Account,cmd.amount)

	cmdhandler := self.clients[cmd.Coin]

	for i:=0; i<max_once_account_number; i++ {
		acc, err := cmdhandler.NewAccount()
		if err!=nil {
			l4g.Error("new %s account error, messafge", cmd.Coin, err.Error())
			return nil, err
		}
		accs = append(accs, acc)
	}
	return accs, nil
}

func privatekeyFromChiperHexString(chiper string) (*ecdsa.PrivateKey, error) {
	chiper = utils.String_cat_prefix(chiper, "0x")
	chiper_bytes, err := hex.DecodeString(chiper);
	if nil!=err {
		return nil, err
	}
	plainKey, err := blockchain_server.Decrypto(chiper_bytes)
	if err!=nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(plainKey)
}

func newTransferCmd(msgId, coinname, chiperKey, to string, amount uint64) (*CmdTx) {
	return &CmdTx{ NetCmd:types.NetCmd{MsgId: msgId, Coin:coinname, Method:"send_transaction", Result:nil, Error:nil},
		chiperkey:chiperKey, tx:&types.Transfer{To: to, Amount:amount}}
}

func newCareAddressCmd(msgId, coin string, address []string) (*CmdMonitorAddresses) {
	return &CmdMonitorAddresses{
		NetCmd:types.NetCmd{MsgId: msgId, Coin: coin, Method:"watch_addresses", Result:nil, Error:nil},
		addresses:address }
}

func (self *NetcmdHandler) waitTxCmd() *CmdTx {
	return <-self.txCmdChannel
}

func (self *NetcmdHandler) waitTxCmdClose() bool {
	return <-self.txCmdClose
}

func (self *NetcmdHandler) closeTransferloop() {
	self.txCmdClose <-true
}

func (self *NetcmdHandler)SendTx(msgid, coin, chiperkey, to string, amount uint64)  {
	transferCmd := newTransferCmd(msgid, coin, chiperkey, to, amount)
	self.txCmdChannel <- transferCmd
	//privatekey, err := privatekeyFromChiperHexString(chiperKey)
	//if err!=nil {
	//	return nil, err
	//}
	//gaslimit := 0x2fefd8			// gas limit 可以设置尽量大
	//big_amount := big.NewInt(int64(amount))
	//toaddress := common.HexToAddress(to)
	//gasprice, err := self.client.SuggestGasPrice()
	////gasprice, err := big.NewInt(int64(math.Pow10(18))), func() error {return nil}()
	//if nil!=err {
	//	return err
	//}
	//nonce, err := self.client.PendingNonceAt(ctx, ac.Address)
	//if nil!=err {
	//	return err
	//}
	//
	//tx := types.NewTransaction(nonce, toaddress, amount, uint64(gaslimit), gasprice, nil)
	//ks.Unlock(ac, "ko2005,./123eth")
	//tx, err = ks.SignTx(ac, tx, big.NewInt(15))
	//if err!=nil {
	//	return err
	//}
	//
	////fmt.Println("dd-mm-yyyy : ", current.Format("02-01-2006"))
	//tx := NewTransfer(tx, time.Now().Format("02-01-2006"))
	//
	//if err:=tx.Send(ctx, client); err!=nil {
	//	return err
	//}
}

func (self *NetcmdHandler)TxInfo(tx_hash string)(*types.Transfer, error) {
	return nil, nil
}

func (self *NetcmdHandler) Blocknumber() uint64 {
	return 0
}

func (self *NetcmdHandler)Close() {
	// TODO !!!!!
	for _, client := range self.clients {
		client.Stop(context.TODO(), time.Second * 5)
	}
}


func init () {

}

