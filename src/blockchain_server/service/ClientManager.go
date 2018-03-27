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
	"time"
)

const (
	max_once_account_number = 100
)

type ErrorInvalideParam struct {
	message string
}

type ErrSendTx struct {
	types.NetCmdErr
}

type ClientManager struct {
	txCmdChannel types.CmdTxChannel			// txCmdChannel 用于接收外部调用的命令并执行
	txRchChannel types.RechargeTxChannel	// txRchChannel 接收所有Client监听地址充值事件

	txCmdClose   chan bool
	clients      map[string]blockchain_server.ChainClient

	loopRechageRuning bool
	loopTxCmdRuning   bool
	txCmdFeed         event.Feed
	rechTxFeed        event.Feed
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

func (self *ClientManager) AddClient(client blockchain_server.ChainClient) {
	if self.clients==nil {
		self.clients = make(map[string]blockchain_server.ChainClient)
	}
	client.SubscribeRechageTx(self.txRchChannel)
	self.clients[client.Name()] = client
}

func (self *ClientManager) loopTxCmd() {
	if self.loopTxCmdRuning {return} else {self.loopTxCmdRuning = true}

	if self.txCmdChannel == nil {
		fmt.Printf("self.txCmdChannel is nil , create new")
		self.txCmdChannel = make(types.CmdTxChannel)
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
		defer func() {self.loopTxCmdRuning = false}()
	}()
}

func (self *ClientManager) loopRechargeTxMessage () {
	if self.loopRechageRuning {return} else {self.loopRechageRuning = true}

	go func() {
		defer func(){self.loopRechageRuning = false}()
		for {
			select {
			case rechTx := <-self.txRchChannel:{
				self.rechTxFeed.Send(rechTx)
			}
			case close := <-self.txCmdClose: {
				if close {
					goto endfor
				}
			}
			}
		}
		endfor:

	}()

}

func (self *ClientManager) Start() {
	self.loopRechargeTxMessage()
	self.loopTxCmd()
	self.startAllClient()
}

func (self *ClientManager) startAllClient() error {
	if self.clients==nil || len(self.clients)==0 {
		return fmt.Errorf("There are 0 client instance. add client instance first!")
	}

	for _, instance := range self.clients {
		instance.SubscribeRechageTx(self.txRchChannel)
		instance.Start()
	}
	return nil
}

func (self *ClientManager) SubscribeTxCmdState (txCmdChannel types.CmdTxChannel) event.Subscription{
	return self.txCmdFeed.Subscribe(txCmdChannel)
}

func (self *ClientManager) SubscribeTxRecharge(txRechageChannel types.RechargeTxChannel) event.Subscription {
	subscribe := self.rechTxFeed.Subscribe(txRechageChannel)
	return subscribe
}

//func (slef *ClientManager) SubscribeRechargeTx(rctChannel types.RechargeTxChannel) *event.Subscription {
//	for _, instance := range slef.clients {
//		instance.SubscribeRechargeTx(rctChannel)
//	}
//}

func (self *ClientManager) innerInsertRechargeAddress(coin string, addresses []string,
	) (error) {
	client := self.clients[coin]
	if nil==client {
		return fmt.Errorf("coin not supported:%s", coin)
	}
	client.InsertRechageAddress(addresses)
	return nil
}

func (self *ClientManager) InsertRechargeAddress(cmdRchAddress *types.CmdRechargeAddress) error {
	return self.innerInsertRechargeAddress(cmdRchAddress.Coinname, cmdRchAddress.Addresses)
}

func (self *ClientManager) innerSendTx(txCmd *types.CmdTx) {
	l4g.Trace("------------sendTransaction begin------------")
	instance := self.clients[txCmd.Coinname]

	err := instance.SendTx(txCmd.Chiperkey, txCmd.Tx)
	if nil!=err {
		// -32000 to -32099	Server error Reserved for implementation-defined server-errors.
		txCmd.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
		l4g.Error("Send Transaction error:%s", txCmd.Error.Message)
		self.txCmdFeed.Send(txCmd)
		
	}

	txCmd.Tx.State = types.Tx_state_commited
	self.txCmdFeed.Send(txCmd)

	escapeloop := false
	// Transaction state change : committed->waite confirm number -> confirmed/unconfirmed ??
	for !escapeloop {
		time.Sleep(time.Second)
		tx, err := instance.Tx(txCmd.Tx.Tx_hash)
		tx.Confirmationsnumber = txCmd.Tx.Confirmationsnumber

		if err != nil {
			if _, ok := err.(*types.NotFound); ok {
				l4g.Trace("Transaction: %s have not found on node, please wait.....!", txCmd.Tx.Tx_hash)
			} else {
				escapeloop = true
				txCmd.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
				l4g.Error(err.Error())
				self.txCmdFeed.Send(txCmd)
			}
			continue
		}

		if tx.State != txCmd.Tx.State {
			if tx.State==types.Tx_state_confirmed {
				escapeloop = true
			}

			l4g.Trace("Transaction state changed from:%s to %s",
				types.TxStateString(txCmd.Tx.State), types.TxStateString(tx.State))

			txCmd.Tx = tx
			self.txCmdFeed.Send(txCmd)

		} else if tx.State==types.Tx_state_mined {

			if tx.PresentBlock !=txCmd.Tx.PresentBlock {

				if tx.PresentBlock-tx.PresentBlock >tx.Confirmationsnumber {

					tx.State = types.Tx_state_confirmed
					escapeloop = true

					l4g.Trace("Transaction state from:%s to %s",
						types.TxStateString(types.Tx_state_mined), types.TxStateString(types.Tx_state_confirmed))
				}

				txCmd.Tx = tx
				self.txCmdFeed.Send(txCmd)
			}
		}
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
	self.txCmdChannel = make(types.CmdTxChannel, 256)
	self.txRchChannel = make(types.RechargeTxChannel, 256)
	self.clients = make(map[string]blockchain_server.ChainClient, 256)
	self.txCmdClose = make(chan bool)
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

func (self *ClientManager)Close() {
	close(self.txCmdChannel)
	close(self.txCmdClose)
	for _, client := range self.clients {
		client.Stop()
	}
}

