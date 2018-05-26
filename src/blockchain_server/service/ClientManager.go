package service

import (
	chainclient "blockchain_server/chains/client"
	"blockchain_server/chains/event"
	"blockchain_server/crypto"
	"blockchain_server/types"
	"blockchain_server/utils"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"blockchain_server/l4g"
	"time"
)

var L4g = L4G.GetL4g("default")
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
	txCmdChannel    types.CmdTxChannel      // txCmdChannel    用于接收外部调用的命令并执行
	txRchChannel    types.RechargeTxChannel // txRchChannel    接收所有Client监听地址充值事件
	txCmdqTxChannel types.CmdqTxChannel     // txCmdqTxChannel 通过hash值查询tx的channel

	clients map[string]chainclient.ChainClient

	loopRechageRuning bool
	loopTxCmdRuning   bool
	txCmdFeed         event.Feed
	rechTxFeed        event.Feed

	ctx        context.Context
	ctx_cannel context.CancelFunc

	// TODO: try do all net command within one go routine
	cmdChannel types.NetCmdChannel
}

func newInvalidParamError(msg string) *ErrorInvalideParam {
	return &ErrorInvalideParam{message: msg}
}

//-32700	Parse error	Invalid JSON was received by the server.
//An error occurred on the server while parsing the JSON text.
//-32600	Invalid Request	The JSON sent is not a valid Request object.
//-32601	Method not found	The method does not exist / is not available.
//-32602	Invalid params	Invalid method parameter(s).
//-32603	Internal error	Internal JSON-RPC error.
//-32000 to -32099	Server error	Reserved for implementation-defined server-errors.
func newTxSendError(txCmd *types.CmdSendTx, message string, code int32) *ErrSendTx {
	return &ErrSendTx{types.NetCmdErr{
		Message: fmt.Sprintf("Send %s Transaction error:%s\ntx detail:%s",
			txCmd.Coinname, message, txCmd.Tx.String()), Code: code, Data: nil}}
}

func (e ErrSendTx) Error() string {
	return fmt.Sprintf("SendTransaction error, code:%d, message:%s", e.Code, e.Message)
}

func (self ErrorInvalideParam) Error() string {
	return self.message
}

func (self *ClientManager) AddClient(client chainclient.ChainClient) {
	if self.clients == nil {
		self.clients = make(map[string]chainclient.ChainClient)
	}
	client.SubscribeRechargeTx(self.txRchChannel)
	self.clients[client.Name()] = client
}

func (self *ClientManager) loopTxCmd() {
	if self.loopTxCmdRuning {
		return
	} else {
		self.loopTxCmdRuning = true
	}

	if self.txCmdChannel == nil {
		L4g.Trace("self.txCmdChannel is nil , create new")
		self.txCmdChannel = make(types.CmdTxChannel, 512)
	}

	go func() {
		defer func() { self.loopTxCmdRuning = false }()
		L4g.Trace("start transaction command loop!!")
		for {
			select {
			case txCmd := <-self.txCmdChannel:
				{
					if txCmd == nil {
						L4g.Trace("!!!!!!!!!!!!!txcmd is nil, maybe Transaction Cmd Channel was closed, exit loop !!!!!!!!!!!!!")
						goto endfor
					} else {
						go self.innerSendTx(txCmd)
					}
				}
			case <-self.ctx.Done():
				{
					goto endfor
				}
				//default: {
				//	fmt.Printf("looping Transaction command.....\n")
				//	time.Sleep(time.Second * time.Duration(3))
				//}
			}
		}
	endfor:
		L4g.Trace("exit transaction command loop!!")
	}()
}


func (self *ClientManager) Client(coinName string) chainclient.ChainClient {
	return self.clients[coinName]
}

func (self *ClientManager) loopRechargeTxMessage() {
	if self.loopRechageRuning {
		return
	} else {
		self.loopRechageRuning = true
	}

	go func() {
		defer func() { self.loopRechageRuning = false }()
		L4g.Trace("start recharge transaction loop!")
		for {
			select {
			// txRchChannel 通过Client.SubscribeRecharge, 传递给Client
			// 并在此用于接收Recharge的交易通知
			case rechTx := <-self.txRchChannel:
				{
					if rechTx != nil {
						go self.trackRechargeTx(rechTx)
					} else {
						L4g.Trace("Recharge Transaction channel was closed!")
					}
				}
			case <-self.ctx.Done():
				{
					L4g.Trace(self.ctx.Err().Error())
					goto endfor
				}
			}
		}
	endfor:
		L4g.Trace("exit recharge transaction loop!")
	}()

}

func (self *ClientManager) Start() {
	self.loopRechargeTxMessage()
	self.loopTxCmd()
	self.startAllClient()
}

func (self *ClientManager) startAllClient() error {
	if self.clients == nil || len(self.clients) == 0 {
		return fmt.Errorf("There are 0 client instance. add client instance first!")
	}

	if self.txRchChannel == nil {
		return fmt.Errorf("Recharge Transaction channel is nil, subscribe first!")
	}

	var invalid_inst []chainclient.ChainClient

	for _, instance := range self.clients {
		L4g.Trace(`-----------start ***%s*** instance-----------`,
			instance.Name())

		instance.SubscribeRechargeTx(self.txRchChannel)

		if err := instance.Start(); err != nil {
			invalid_inst = append(invalid_inst, instance)
			L4g.Trace("start client instance : %s, faild, message:%s", instance.Name(), err.Error())
		} else {
			L4g.Trace("start client instance :  !!!!!>-%s,success<-!!!!!", instance.Name())
		}
	}

	for _, inst := range invalid_inst {
		delete(self.clients, inst.Name())
	}

	return nil
}

func (self *ClientManager) SubscribeTxCmdState(txCmdChannel types.CmdTxChannel) event.Subscription {
	return self.txCmdFeed.Subscribe(txCmdChannel)
}

func (self *ClientManager) SubscribeTxRecharge(txRechageChannel types.RechargeTxChannel) event.Subscription {
	subscribe := self.rechTxFeed.Subscribe(txRechageChannel)
	return subscribe
}

func (self *ClientManager) innerInsertRechargeAddress(coin string, addresses []string,
) error {
	client := self.clients[coin]
	if nil == client {
		return fmt.Errorf("coin not supported:%s", coin)
	}
	client.InsertWatchingAddress(addresses)
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
		case tx := <-tx_channel:
			{
				rechTx.Tx = tx
				L4g.Trace(`
ψ(｀∇´)ψψ(｀∇´)ψ [Track Recharge Tx information] ψ(｀∇´)ψψ(｀∇´)ψ
%s`, tx.String())
				self.rechTxFeed.Send(rechTx)
				if tx.State == types.Tx_state_confirmed || tx.State == types.Tx_state_unconfirmed {
					goto breakfor
				}
			}
		case rechTx.Err = <-err_channel:
			{
				self.rechTxFeed.Send(rechTx)
				goto breakfor
			}
		case <-self.ctx.Done():
			{
				rechTx.Err = self.ctx.Err()
				self.rechTxFeed.Send(rechTx)
				goto breakfor
			}
		}
	}
breakfor:
	L4g.Trace("Recharge Transaction(%s) done!!!", rechTx.Tx.Tx_hash)
}

func (self *ClientManager) trackTxCmd(txCmd *types.CmdSendTx) {
	tx_channel := make(chan *types.Transfer)
	err_channel := make(chan error)

	go self.trackTxState(txCmd.Coinname, txCmd.Tx, tx_channel, err_channel)

	for {
		select {
		case tx := <-tx_channel:
			{
				txCmd.Tx = tx
				L4g.Trace(`(＠。ε。＠)(＠。ε。＠) [[TrackTxCmd information]] (*≧∪≦)(*≧∪≦)(*≧∪≦)
%s`, tx.String())
				self.txCmdFeed.Send(txCmd)
				if tx.State == types.Tx_state_confirmed || tx.State == types.Tx_state_unconfirmed {
					goto break_for
				}
			}
		case err := <-err_channel:
			{
				txCmd.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
				self.txCmdFeed.Send(txCmd)
				goto break_for
			}
		case <-self.ctx.Done():
			{
				txCmd.Error = types.NewNetCmdErr(-32000, self.ctx.Err().Error(), nil)
				self.txCmdFeed.Send(txCmd)
				goto break_for
			}
		}
	}
break_for:
}

/* 查询地址中的资产余额, 使用 NewQueryBalanceCmd() 创建CmdqueryBalance对象, 作为参数传入.
如果doneChannel参数为空, 此函数为阻塞模式, 直接返回查询结果
如果doneChannel有值, 则此函数会立即返回0, nil, 并通过doneChanne通知外部,
如果doneChannel触发值为true, 表示查询成功, 通过CmdqueryBalance.Result可以获取资产余额
如果doneChannel触发值为false, 表示查询失败, 通过CmdQueryBalance.Error获取失败信息! */
func (self *ClientManager) GetBalance(ctx context.Context, cmdBalance *types.CmdqueryBalance, doneChannel chan bool) (float64, error) {
	if nil == cmdBalance {
		return 0, fmt.Errorf("GetBalance error, Invalid paramater!")
	}

	instance := self.clients[cmdBalance.Coinname]
	if instance == nil {
		return 0, fmt.Errorf("Not supported assert!")
	}

	if nil == doneChannel {
		return instance.GetBalance(cmdBalance.Address, cmdBalance.TokenSymbol)
	} else {
		go func() {
			var err error
			cmdBalance.Result, err = instance.GetBalance(cmdBalance.Address, cmdBalance.TokenSymbol)
			if err != nil {
				cmdBalance.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
				doneChannel <- false
			}
			doneChannel <- true
		}()
	}
	return 0, nil
}

/*  if qtxchannel is nil, block and returns tx or error
else no block, returns directly, with 2 nil,
then, send back cmdqtx from qtxchannel,
as tx information stored in the netcmd.result member*/
func (self *ClientManager) QuryTx(cmdqTx *types.CmdqueryTx, qTxChannel types.CmdqTxChannel) (tx *types.Transfer, err error) {

	instance := self.clients[cmdqTx.Coinname]
	if instance == nil {
		return nil, fmt.Errorf("query on not supported coin type(%s)", cmdqTx.Coinname)
	}
	if qTxChannel == nil {
		tx, err = instance.Tx(cmdqTx.Hash)
	} else {
		go func() {
			cmdqTx.Result, err = instance.Tx(cmdqTx.Hash)
			if err != nil {
				cmdqTx.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
			}
			qTxChannel <- cmdqTx
		}()
	}
	return tx, err
}

func (self *ClientManager) BlockHeight(assert string) uint64 {
	instance := self.clients[assert]
	if instance == nil {
		L4g.Error("query on not supported coin type(%s)", assert)
		return 0
	}
	return instance.BlockHeight()
}

// 检查Transaction状态, 先记录当前块高, 如果发现块高增加1以上, 则查询Tx是否入块,
// 如果入块, 则开始检查Tx确认状态,
// 如果Tx确认已经入块, 但在检查确认状态时, 返回了NotFound错误, 则可能由于区块链分叉, 交易可能会被重新打包
// 则goto到第一步, 重新开始检查确认状态...直到Tx的状态变成了 confimred 或者 uncomfirmed!
func (self *ClientManager) trackTxState(clientName string,
	tx *types.Transfer, tx_channel chan *types.Transfer, err_channel chan error) {

	L4g.Trace("********start trace transaction(%s)", tx.Tx_hash)
	instance := self.clients[clientName]
	tx_channel <- tx

	max_try_count := 30
	i := 0

	if tx.State==types.Tx_state_unconfirmed || tx.State==types.Tx_state_confirmed {
		goto exitfor
	}

	// 每次有新块高增加, 才会去检查transaction的状态是否发生改变
	// 如果状态没有改变, 说明tansaction还没有被打包到新的块中
	// 直到transaction的状态变为了 mined, 再进入 确认的流程中
	for i := 0; i < max_try_count; i++ {
		curBlockHeight := instance.BlockHeight()
		for {
			time.Sleep(time.Second)
			if curBlockHeight < instance.BlockHeight() {
				break
			}
		}

		state := tx.State

		err := instance.UpdateTx(tx)
		if err != nil {
			if _, ok := err.(*types.NotFound); !ok {
				L4g.Error("update Tx faild, message:%s", err.Error())
				err_channel <- err
				goto exitfor
			}
		}

		if state != tx.State {
			tx_channel <- tx
		}

		if tx.State == types.Tx_state_unconfirmed || tx.State == types.Tx_state_confirmed {
			tx_channel <- tx
			goto exitfor
		}
		i++
	}

	if i == max_try_count {
		message := "Update Transaction state upto max count, still can not confirm!!!"
		L4g.Error(message)
		err_channel <- fmt.Errorf(message)
	}

exitfor:
	L4g.Trace("********stop trace transaction(%s)", tx.Tx_hash)
}

func (self *ClientManager) innerSendTx(txCmd *types.CmdSendTx) {
	//	L4g.Trace(`
	//------------send transaction begin------------
	//Asset:%s , Crypted Key:%s
	//TxInfo : %s`, txCmd.Coinname, txCmd.FromKey, txCmd.Tx.String() )

	instance := self.clients[txCmd.Coinname]

	// liuheng add
	// TODO: zl review
	var err error
	if txCmd.SignedTxString != "" {
		err = func() error {
			txCmdSignedByte, err := base64.StdEncoding.DecodeString(txCmd.SignedTxString)
			if nil != err {
				L4g.Error("Signed tx DecodeString error:%s", err.Error())
				return err
			}

			err = instance.SendSignedTx(txCmdSignedByte, txCmd.Tx)
			if nil != err {
				L4g.Error("SendSignedTx Transaction error:%s", err.Error())
				return err
			}

			return nil
		}()
	} else {
		err = instance.SendTx(txCmd.FromKey, txCmd.Tx)
	}

	if nil != err {
		// -32000 to -32099	Server error Reserved for implementation-defined server-errors.
		txCmd.Error = types.NewNetCmdErr(-32000, err.Error(), nil)
		L4g.Error("Send Transaction error:%s", txCmd.Error.Message)
		self.txCmdFeed.Send(txCmd)
		return
	}

	txCmd.Tx.State = types.Tx_state_pending
	self.trackTxCmd(txCmd)

	var message string
	if txCmd.Error != nil {
		message = txCmd.Error.Message
	} else {
		message = ""
	}

	L4g.Trace("send transaction(%s), result: %s, message:", types.TxStateString(txCmd.Tx.State), message)
	L4g.Trace("------------send transaction end------------")
}

func NewClientManager() *ClientManager {
	//clientManager := &ClientManager{txCmdChannel:make(chan *types.CmdSendTx),
	//	clients : make(map[string]blockchain_server.ChainClient),
	//	txCmdClose : make(chan bool)}

	clientManager := &ClientManager{}
	clientManager.init()

	return clientManager
}

func (self *ClientManager) init() {
	self.txCmdChannel = make(types.CmdTxChannel, 256)
	self.txRchChannel = make(types.RechargeTxChannel, 256)
	self.clients = make(map[string]chainclient.ChainClient, 256)
	self.ctx, self.ctx_cannel = context.WithCancel(context.Background())
}

func (self *ClientManager) NewAccounts(cmd *types.CmdNewAccounts) ([]*types.Account, error) {
	if cmd.Amount == 0 || cmd.Amount > max_once_account_number {
		return nil, newInvalidParamError(fmt.Sprintf("the count of account must >0 and <%d", max_once_account_number))
	}

	client := self.clients[cmd.Coinname]
	if nil == client {
		return nil, fmt.Errorf("not found '%s' client!", cmd.Coinname)
	}

	return client.NewAccount(cmd.Amount)
}

func privatekeyFromChiperHexString(chiper string) (*ecdsa.PrivateKey, error) {
	chiper = utils.String_cat_prefix(chiper, "0x")
	chiper_bytes, err := hex.DecodeString(chiper)
	if nil != err {
		return nil, err
	}
	plainKey, err := crypto.Decrypto(chiper_bytes)
	if err != nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(plainKey)
}

// 交易和充值中的单位都是10^8为一个单位
// 即1^8 单位为一个bitcoin或者eth
func (self *ClientManager) SendTx(cmdTx *types.CmdSendTx) {
	if self.loopTxCmdRuning == false {
		L4g.Error("TxCommandLoop is not runing!!!")
		return
	}

	L4g.Trace("Recived one SendTx command:%s", cmdTx.MsgId)

	if self.txCmdChannel == nil {
		L4g.Trace("txCmdChannel is nil, create new")
		self.txCmdChannel = make(chan *types.CmdSendTx, 512)
	}

	self.txCmdChannel <- cmdTx
}

func (self *ClientManager) Close() {
	self.ctx_cannel()
	for _, client := range self.clients {
		client.Stop()
	}

	close(self.txCmdChannel)
	close(self.txRchChannel)

}

// build transaction, return types.CmdSendTx 's json string
func (self *ClientManager) BuildTx(txCmd *types.CmdSendTx) (string, error) {
	L4g.Trace("------------Build transaction begin------------")
	instance := self.clients[txCmd.Coinname]

	err := instance.BuildTx(txCmd.FromKey, txCmd.Tx)
	if nil != err {
		L4g.Error("Build Transaction error:%s", err.Error())
		return "", err
	}

	txCmdByte, err := json.Marshal(txCmd)
	if nil != err {
		L4g.Error("Build Transaction Marshal error:%s", err.Error())
		return "", err
	}

	txCmdString := base64.StdEncoding.EncodeToString(txCmdByte)
	if nil != err {
		L4g.Error("Build Transaction base64 EncodeToString error:%s", err.Error())
		return "", err
	}

	L4g.Trace("------------Build transaction end------------")

	return txCmdString, nil
}

// sing transaction, return types.CmdSendTx 's json string
func (self *ClientManager) SignTx(chiperPrikey string, txCmdString string) (string, error) {
	L4g.Trace("------------Sign transaction begin------------")

	// 解包
	txCmdByte, err := base64.StdEncoding.DecodeString(txCmdString)
	if nil != err {
		L4g.Error("Sign Transaction base64 DecodeString error:%s", err.Error())
		return "", err
	}

	var txCmd types.CmdSendTx
	err = json.Unmarshal(txCmdByte, &txCmd)
	if err != nil {
		L4g.Error("Sign Transaction Unmarshal error:%s", err.Error())
		return "", err
	}

	// 签名
	instance := self.clients[txCmd.Coinname]

	txCmdSignedByte, err := instance.SignTx(chiperPrikey, txCmd.Tx)
	if nil != err {
		L4g.Error("Sign Transaction SignTx error:%s", err.Error())
		return "", err
	}

	txCmdSignedByteString := base64.StdEncoding.EncodeToString(txCmdSignedByte)
	if nil != err {
		L4g.Error("Sign Transaction base64 error:%s", err.Error())
		return "", err
	}

	// 重新打包
	txCmd.SignedTxString = txCmdSignedByteString
	txCmdByte2, err := json.Marshal(txCmd)
	if nil != err {
		L4g.Error("Sign Transaction Marshal error:%s", err.Error())
		return "", err
	}

	txCmdString2 := base64.StdEncoding.EncodeToString(txCmdByte2)
	if nil != err {
		L4g.Error("Sign Transaction base64 EncodeToString error:%s", err.Error())
		return "", err
	}

	L4g.Trace("------------Sign transaction end------------")

	return txCmdString2, nil
}

// send transaction, txCmdString may signed, or unsigned, if unsigned, chiperprikey need real prikey
func (self *ClientManager) SendSignedTx(txCmdString string) error {
	L4g.Trace("------------SendSignedTx transaction begin------------")

	txCmdByte, err := base64.StdEncoding.DecodeString(txCmdString)
	if nil != err {
		L4g.Error("SendSignedTx Transaction base64 DecodeString error:%s", err.Error())
		return err
	}

	txCmd := &types.CmdSendTx{}
	err = json.Unmarshal(txCmdByte, txCmd)
	if err != nil {
		L4g.Error("SendSignedTx Transaction Unmarshal error:%s", err.Error())
		return err
	}

	self.SendTx(txCmd)

	L4g.Trace("------------SendSignedTx transaction end------------")
	return nil
}
