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
	"context"
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
	txCmdChannel types.CmdTxChannel			// txCmdChannel    用于接收外部调用的命令并执行
	txRchChannel types.RechargeTxChannel	// txRchChannel    接收所有Client监听地址充值事件
	txCmdqTxChannel types.CmdqTxChannel		// txCmdqTxChannel 通过hash值查询tx的channel

	clients      map[string]blockchain_server.ChainClient

	loopRechageRuning bool
	loopTxCmdRuning   bool
	txCmdFeed         event.Feed
	rechTxFeed        event.Feed

	ctx 			context.Context
	ctx_cannel 		context.CancelFunc


	// TODO: try do all net command within one go routine
	cmdChannel types.NetCmdChannel
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
		defer func() {self.loopTxCmdRuning = false}()
		l4g.Trace("start transaction command loop!!")
		for {
			select {
			case txCmd := <- self.txCmdChannel: {
				if txCmd==nil {
					l4g.Trace("Transaction Cmd Channel was closed!")
					goto endfor
				} else {
					l4g.Trace("recived TxCommand: %s \n", txCmd.MsgId)
					go self.innerSendTx(txCmd)
				}
			}
			case <-self.ctx.Done(): {
				goto endfor
			}
			//default: {
			//	fmt.Printf("looping Transaction command.....\n")
			//	time.Sleep(time.Second * time.Duration(3))
			//}
			}
		}
		endfor:
			l4g.Trace("exit transaction command loop!!")
	}()
}

func (self *ClientManager) loopRechargeTxMessage () {
	if self.loopRechageRuning {return} else {self.loopRechageRuning = true}

	go func() {
		defer func(){self.loopRechageRuning = false}()
		l4g.Trace("start recharge transaction loop!")
		for {
			select {
			case rechTx := <-self.txRchChannel:{
				if rechTx!=nil {
					go self.trackRechargeTx(rechTx)
				} else {
					l4g.Trace("Recharge Transaction channel was closed!")
				}
			}
			case <-self.ctx.Done():{
				l4g.Trace(self.ctx.Err().Error())
				goto endfor
			}
			}
		}
		endfor:
			l4g.Trace("exit recharge transaction loop!")
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

	if self.txRchChannel == nil {
		return fmt.Errorf("Recharge Transaction channel is nil, subscribe first!")
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

func (self *ClientManager) trackRechargeTx(rechTx *types.RechargeTx) {
	tx_channel := make(chan *types.Transfer)
	err_channel := make(chan error)

	go self.trackTxState(rechTx.Coin_name, rechTx.Tx, tx_channel, err_channel)

	for {
		select {
		case tx := <-tx_channel:{
			rechTx.Tx = tx
			l4g.Trace("Send Recharge Transaction to channel********")
			self.rechTxFeed.Send(rechTx)
			if tx.State==types.Tx_state_confirmed || tx.State==types.Tx_state_unconfirmed {
				goto endfor
			}
		}
		case rechTx.Err = <-err_channel:{
			self.rechTxFeed.Send(rechTx)
			goto endfor
		}
		case <-self.ctx.Done():{
			rechTx.Err = self.ctx.Err()
			self.rechTxFeed.Send(rechTx)
			goto endfor
		}
		}
	}
endfor:
	l4g.Trace("Recharge Transaction(%s) done!!!", rechTx.Tx.Tx_hash)
}

func (self *ClientManager) trackTxCmd(txCmd *types.CmdTx) {
	tx_channel := make(chan *types.Transfer)
	err_channel := make(chan error)
	go self.trackTxState(txCmd.Coinname, txCmd.Tx, tx_channel, err_channel)
	for {
		select {
		case tx := <-tx_channel:{
			txCmd.Tx = tx
			self.txCmdFeed.Send(txCmd)
			if tx.State==types.Tx_state_confirmed || tx.State==types.Tx_state_unconfirmed {
				goto endfor
			}
		}
		case err := <-err_channel:{
			txCmd.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
			self.txCmdFeed.Send(txCmd)
			goto endfor
		}
		case <-self.ctx.Done():{
			txCmd.Error = types.NewNetCmdErr(-32000, self.ctx.Err().Error(), nil)
			self.txCmdFeed.Send(txCmd)
			goto endfor
		}
		}
	}
endfor:
}

/*  if qtxchannel is nil, block and returns tx or error
	else no block, returns directly, with 2 nil,
	then, send back cmdqtx from qtxchannel,
	as tx information stored in the netcmd.result member*/
func (self *ClientManager) QuryTx(cmdqTx *types.CmdqueryTx, qTxChannel types.CmdqTxChannel) (tx *types.Transfer, err error) {

	instance := self.clients[cmdqTx.Coinname]
	if instance==nil {
		return nil, fmt.Errorf("query on not supported coin type(%s)", cmdqTx.Coinname)
	}
	if qTxChannel==nil {
		tx, err = instance.Tx(cmdqTx.Hash)
	} else {
		go func() {
			cmdqTx.Result, err = instance.Tx(cmdqTx.Hash)
			if err!=nil {
				cmdqTx.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
			}
			qTxChannel <- cmdqTx
		}()
	}
	return tx, err
}

// TODO:
//func (self *ClientManager) loopNetCmd(){
//	if self.cmdChannel==nil {
//		self.cmdChannel = make(types.NetCmdChannel)
//	}
//
//	for {
//		select {
//		case cmd := <- self.cmdChannel: {
//			if value, ok := cmd.(*types.CmdqueryTx); ok {
//				// do CmdqueryTx
//			}
//			if value, ok := cmd.(*types.CmdTx); ok {
//
//			}
//
//			if value, ok := cmd.(*types.CmdNewAccounts); ok {
//
//			}
//		}
//		case self.ctx.Done(){
//
//		}
//		}
//	}
//	endfor:
//}

func (self *ClientManager) trackTxState(clientName string,
	tx *types.Transfer, tx_channel chan *types.Transfer, err_channel chan error) {

	l4g.Trace("********start trace transaction(%s)", tx.Tx_hash)
	instance := self.clients[clientName]
	tx.State = types.Tx_state_commited

	// Transaction state change : committed->waite confirm number -> confirmed/unconfirmed ??
	for {
		time.Sleep(time.Second * 2)
		tmp_tx, err := instance.Tx(tx.Tx_hash)

		if err != nil {
			if _, ok := err.(*types.NotFound); ok {
				l4g.Trace("Transaction:(%s) have not found on node, please wait.....!", tx.Tx_hash)
				continue
			} else {
				err_channel <- err
				l4g.Error(err.Error())
				goto endfor
			}
		}

		if tx.State != tmp_tx.State{
			l4g.Trace("Transaction state changed from:%s to %s",
				types.TxStateString(tx.State), types.TxStateString(tmp_tx.State))

			tx.State =  tmp_tx.State
			tx_channel <- tmp_tx

			if tmp_tx.State==types.Tx_state_confirmed ||
				tmp_tx.State==types.Tx_state_unconfirmed {
				l4g.Trace("Transaction(%s) success done!!!", tmp_tx.Tx_hash)
				goto endfor
			}
		} else if tx.State==types.Tx_state_mined {
			if tx.PresentBlock != tmp_tx.PresentBlock{
				tx.PresentBlock = tmp_tx.PresentBlock
				tx_channel <- tmp_tx
			}
		}
	}
endfor:
	l4g.Trace("********stop trace transaction(%s)", tx.Tx_hash)
}

func (self *ClientManager) innerSendTx(txCmd *types.CmdTx) {
	l4g.Trace("------------send transaction begin------------")
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
	self.trackTxCmd(txCmd)

	var message string
	if txCmd.Error!=nil {
		message = txCmd.Error.Message
	} else {
		message = ""
	}

	l4g.Trace("send transaction(%s), result: %s, message:", types.TxStateString(txCmd.Tx.State), message)
	l4g.Trace("------------send transaction end------------")
}

/*
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
*/

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
	self.ctx, self.ctx_cannel = context.WithCancel(context.Background())
}

func (self *ClientManager)NewAccounts(cmd *types.CmdNewAccounts) ([]*types.Account, error) {
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



func (self *ClientManager) SendTx(cmdTx *types.CmdTx) {
	if self.txCmdChannel==nil {
		fmt.Print("txCmdChannel is nil, create new")
		self.txCmdChannel = make(chan *types.CmdTx)
	}
	self.txCmdChannel <- cmdTx
}

func (self *ClientManager)Close() {
	self.ctx_cannel()

	close(self.txCmdChannel)
	close(self.txRchChannel)

	for _, client := range self.clients {
		client.Stop()
	}
}

