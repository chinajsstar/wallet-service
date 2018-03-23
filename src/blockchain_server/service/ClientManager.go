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
	subcribeFeed event.Feed
	//cmdHandler *ClientManager
)

func init() {
	//cmdHandler = &ClientManager{}
}

type ErrorInvalideParam struct {
	message string
}

type ErrSendTx struct {
	types.NetCmdErr
}


type ClientManager struct {
	txCmdChannel chan *types.CmdTx
	txCmdClose   chan bool
	clients      map[string]blockchain_server.ChainClient
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
func newTxSendError(txCmd *types.CmdTx, message string, code int32)*ErrSendTx {
	return &ErrSendTx{types.NetCmdErr{
		Message:fmt.Sprintf("Send %s Transaction error:%s\ntx detail:%s",
		txCmd.Coinname, message, txCmd.Tx.String()), Code: code, Data: nil}}
}

func (e ErrSendTx) Error() string {
	return fmt.Sprintf("SendTransaction error, code:%d, message:%s", e.Code, e.Message)
}

func (self ErrorInvalideParam)Error() string {
	return self.message
}

// subscribe Sample
/*
func TestFeed(t *testing.TxStateString) {
	var subcribeFeed Feed
	var done, subscribed sync.WaitGroup
	subscriber := func(i int) {
		defer done.Done()

		subchan := make(chan int)
		sub := subcribeFeed.Subscribe(subchan)
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
	if nsent := subcribeFeed.Send(1); nsent != n {
		t.Errorf("first send delivered %d times, want %d", nsent, n)
	}
	if nsent := subcribeFeed.Send(2); nsent != 0 {
		t.Errorf("second send delivered %d times, want 0", nsent)
	}
	done.Wait()
}
*/

func (self *ClientManager) AddClient(client blockchain_server.ChainClient) {
	if self.clients==nil {
		self.clients = make(map[string]blockchain_server.ChainClient)
	}
	self.clients[client.Name()] = client
}

func (self *ClientManager) Start(ctx context.Context, rctxChannel types.RechargeTxChannel) {
	self.loopTxCmd()
	self.startAllClient(ctx, rctxChannel)
}

func (self *ClientManager) startAllClient(ctx context.Context, rctChannel types.RechargeTxChannel) error {
	if self.clients==nil || len(self.clients)==0 {
		return fmt.Errorf("there are 0 client instance. add client instance first!")
	}
	for _, instance := range self.clients {
		instance.Start(rctChannel)
	}
	return nil
}

func (self *ClientManager) SubscribeTxStateChange(txStateChannel types.TxStateChange_Channel) event.Subscription{
	subscribe := subcribeFeed.Subscribe(txStateChannel)
	return subscribe
}

//func (slef *ClientManager) SubscribeRechargeTx(rctChannel types.RechargeTxChannel) *event.Subscription {
//	for _, instance := range slef.clients {
//		instance.SubscribeRechargeTx(rctChannel)
//	}
//}

func (self *ClientManager) innerSetRechargeAddress(coin string, addresses []string,
	) (error) {
	client := self.clients[coin]
	if nil==client {
		return fmt.Errorf("coin not supported:%s", coin)
	}
	client.InsertCareAddress(addresses)
	return nil
}

func (self *ClientManager) SetRechargeAddress(cmdRchAddress *types.CmdRechargeAddress) error {
	return self.innerSetRechargeAddress(cmdRchAddress.Coinname, cmdRchAddress.Addresses)
}

//func (self *ClientManager) SubscribeIncomingTx (types. incomeTxChannel *Addresswatcher_Channel)  *event.Subscription {
//	subscribe := subcribeFeed.Subscribe(incomeTxChannel)
//	return &subscribe
//}

func (self *ClientManager) innerSendTx(txCmd *types.CmdTx) {
	l4g.Trace("------------sendTransaction begin------------")
	instance := self.clients[txCmd.Coinname]

	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)

	err := instance.SendTx(ctx, txCmd.Chiperkey, txCmd.Tx)
	if nil!=err {
		// -32000 to -32099	Server error Reserved for implementation-defined server-errors.
		txCmd.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
		l4g.Error("Send Transaction error:%s", txCmd.Error.Message)
		subcribeFeed.Send(txCmd)
		
	}

	txCmd.Tx.State = types.Tx_state_commited
	subcribeFeed.Send(txCmd)

	escapeloop := false
	// Transaction state change : committed->waite confirm number -> confirmed/unconfirmed ??
	for !escapeloop {
		time.Sleep(time.Second)
		tx, err := instance.Tx(ctx, txCmd.Tx.Tx_hash)

		if err != nil {
			if _, ok := err.(*types.TxNotFoundErr); ok {
				l4g.Trace("Transaction: %s have not found on node, please wait.....!", txCmd.Tx.Tx_hash)
			} else {
				escapeloop = true
				txCmd.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
				l4g.Error(err.Error())
				subcribeFeed.Send(txCmd)
			}
			continue
		}

		if tx.State != txCmd.Tx.State {
			if tx.State==types.Tx_state_confirmed {
				escapeloop = true
			}

			l4g.Trace("Transaction state from:%s to %s, Transaction information:%s",
				types.TxStateString(txCmd.Tx.State), types.TxStateString(tx.State), tx.String())

			txCmd.Tx = tx
			subcribeFeed.Send(txCmd)

		} else if tx.State==types.Tx_state_mined {

			if tx.PresentBlocknumber!=txCmd.Tx.PresentBlocknumber {

				if tx.PresentBlocknumber-tx.PresentBlocknumber>tx.Confirmationsnumber {

					tx.State = types.Tx_state_confirmed
					escapeloop = true

					l4g.Trace("Transaction state from:%s to %s, Transaction information:%s",
						types.TxStateString(types.Tx_state_mined), types.TxStateString(types.Tx_state_confirmed),
							tx.String())
				}

				txCmd.Tx = tx
				subcribeFeed.Send(txCmd)
			}
		}
		/*
		switch tx.State {
		case types.Tx_state_commited: {
			txCmd.tx = tx
			subcribeFeed.Send(txCmd)
		}
		case types.Tx_state_mined: {
			if txCmd.tx.State != tx.State {
				l4g.Trace("Transaction mined, Transaction infromation:%s", tx.String())
				txCmd.tx = tx
				subcribeFeed.Send(txCmd)
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
			subcribeFeed.Send(txCmd)
		}
		case types.Tx_state_unconfirmed: {
			txCmd.tx = tx
			l4g.Trace("Transaction is unconfimred! tx information:%s", tx.String())
			subcribeFeed.Send(txCmd)
		}
		default: {
			message = fmt.Sprintf("Transaction state looks unusual, the state changed from:'%s' to:'%s' Transaction information:%s", types.TxStateString(txCmd.tx.State), types.TxStateString(tx.State))
			txCmd.Error = types.NewNetCmdErr(-32000, message, nil)
			l4g.Warn(message)
			escapeloop = true
			subcribeFeed.Send(txCmd)
		}
		}
		*/
	}
	l4g.Trace("------------SendTransaction   end------------")
}

func NewClientManager() *ClientManager {
	//clientManager := &ClientManager{txCmdChannel:make(chan *types.CmdTx),
	//	clients : make(map[string]blockchain_server.ChainClient),
	//	txCmdClose : make(chan bool)}

	clientManager := &ClientManager{}
	clientManager.init()

	return clientManager
}

func (self *ClientManager) init () {
	self.txCmdChannel = make(chan *types.CmdTx)
	self.clients = make(map[string]blockchain_server.ChainClient)
	self.txCmdClose = make(chan bool)
}

func (self *ClientManager) loopTxCmd() {
	if self.txCmdChannel == nil {
		fmt.Printf("self.txCmdChannel is nil , create new")
		self.txCmdChannel = make(chan *types.CmdTx)
	}
	go func() {
		closeloop := false
		for !closeloop {
			select {
			case txCmd := <- self.txCmdChannel: {
				fmt.Printf("recived TxCommand: %s \n", txCmd.MsgId)
				go self.innerSendTx(txCmd)
			}
			case closeloop = <- self.txCmdClose: {
				break
			}
			default: {
				fmt.Printf("looping Transaction command.....\n")
				time.Sleep(time.Second * time.Duration(3))
			}

			}
		}
	}()
}

func (self *ClientManager)NewAccounts(cmd *types.CmdAccounts) ([]*types.Account, error) {
	if cmd.Amount==0 || cmd.Amount>max_once_account_number {
		return nil, newInvalidParamError(fmt.Sprintf("the count of account must >0 and <%d", max_once_account_number))
	}
	accs := make([]*types.Account,cmd.Amount)

	client := self.clients[cmd.Coinname]

	if nil==client {
		return nil, fmt.Errorf("not found '%s' client!", cmd.Coinname)
	}

	for i:=0; i<int(cmd.Amount); i++ {
		acc, err := client.NewAccount()
		if err!=nil {
			l4g.Error("new %s account error, messafge", cmd.Coinname, err.Error())
			return nil, err
		}
		accs[i] = acc
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


func (self *ClientManager) closeTransferloop() {
	self.txCmdClose <-true
}

func (self *ClientManager) SendTx(cmdTx *types.CmdTx) {
	if self.txCmdChannel==nil {
		fmt.Print("txCmdChannel is nil, create new")
		self.txCmdChannel = make(chan *types.CmdTx)
	}

	self.txCmdChannel <- cmdTx
}

//func (self *ClientManager)SendTx(msgid, coin, chiperkey, to string, amount uint64)  {
//	transferCmd := NewTxCmd(msgid, coin, chiperkey, to, amount)
//	self.txCmdChannel <- transferCmd
//}

//privatekey, err := privatekeyFromChiperHexString(chiperKey)
//if err!=nil {
//	return nil, err
//}
//gaslimit := 0x2fefd8			// gas limit 可以设置尽量大
//big_amount := big.NewInt(int64(amounte))
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

func (self *ClientManager)TxInfo(tx_hash string)(*types.Transfer, error) {
	return nil, nil
}

func (self *ClientManager) Blocknumber() uint64 {
	return 0
}

func (self *ClientManager)Close() {
	// TODO !!!!!
	for _, client := range self.clients {
		client.Stop(context.TODO(), time.Second * 5)
	}
}

func init () {

}

