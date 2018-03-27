package types

import (
	"fmt"
	"blockchain_server/conf"
)

//-32700	Parse error	Invalid JSON was received by the server.
//An error occurred on the server while parsing the JSON text.
//-32600	Invalid Request	The JSON sent is not a valid Request object.
//-32601	Method not found	The method does not exist / is not available.
//-32602	Invalid params	Invalid method parameter(s).
//-32603	Internal error	Internal JSON-RPC error.
//-32000 to -32099	Server error	Reserved for implementation-defined server-errors.

type Account struct {
	PrivateKey string // 加密后的私钥字节流的16进制字符串
	Address    string // 私钥对应的地址
}

const (
	Tx_state_unkown = iota
	Tx_state_commited             	// transaction was sended(call SendTransaction)
	Tx_state_pending				// pending Transaction on node
	Tx_state_mined					// transaction mined on a block!!
	Tx_state_confirmed              // transaction was mined and stored on a block. confirmed number is 1 or biger
	Tx_state_unconfirmed            // some error happened...need to re send

	Chain_eth = "eth"
	Chain_bitcoin = "btc"

	NetCmdCode_success = iota
	NetCmdCode_failed
)



type CmdTx struct {
	NetCmd
	Chiperkey string
	Tx        *Transfer
}

type CmdAccounts struct {
	NetCmd
	Amount uint32
}

type CmdRechargeAddress struct {
	NetCmd
	Recall_url string
	Addresses  []string
}

type RechargeTxChannel chan *RechargeTx
type CmdTxChannel chan *CmdTx
type Addresswatcher_Channel chan *Transfer
type TxState int

// Transaction of Recharge
type RechargeTx struct {
	Coin_name string
	Tx        *Transfer
}

type Transfer struct {
	Tx_hash             string
	From                string
	To                  string
	Value               uint64
	Gase                uint64
	Gaseprice           uint64
	Total				uint64
	State               TxState
	OnBlock             uint64
	PresentBlock        uint64
	Confirmationsnumber uint64
	Time                uint64 //TODO
	////fmt.Println("dd-mm-yyyy : ", current.Format("02-01-2006"))
}

func (tx *Transfer) Minerfee() uint64 {
	return 	tx.Gase * tx.Gaseprice
}

func (tx *Transfer) Tatolcost() uint64 {
	return tx.Minerfee() + tx.Value
}

func TxStateString(state TxState) string {
	switch state {
	case Tx_state_unkown: {
		return "unkown"
	}
	case Tx_state_commited: {
		return "commit"
	}
	case Tx_state_pending: {
		return "pending"
	}
	case Tx_state_mined: {
		return "mined"
	}
	case Tx_state_confirmed: {
		return "confirmed"
	}
	case Tx_state_unconfirmed: {
		return "unconfirmed"
	}
	default:
		return "unkown"
	}
}

func (tx *Transfer)String() string {
	return fmt.Sprintf(`
	TX (%s)
	From:	%s
	To:		%s
	State:	%s
	Value:	%d
	gasfee:	%d 
	OnBlock:%d
	CurrentBlock:	%d`,
		tx.Tx_hash,
		tx.From,
		tx.To,
		TxStateString(tx.State),
		tx.Value,
		tx.Minerfee(),
		tx.OnBlock, tx.PresentBlock)
}

type NotFound struct {
	message string
}

func (self *NotFound)Error() string {
	return self.message
}

func NewTxNotFoundErr(tx_hash string) *NotFound {
	//return &NotFound{tx_info: fmt.Sprintf("Transaction not found, detail:%s", tx.String())}
	return &NotFound{message:fmt.Sprintf("Transaction(%s) not found!", tx_hash)}
}

type NetCmdErr struct {
	Code 		int32
	Message		string
	Data 		interface{}
}

type NetCmd struct  {
	MsgId    string
	Coinname string
	Method   string
	Result   interface{}
	Error    *NetCmdErr
}

func NewNetCmdErr(code int32, message string, data interface{}) *NetCmdErr {
	return &NetCmdErr{Code:code, Message:message, Data:data}
}

func NewAccountCmd(msgId, coinname string, amount uint32) *CmdAccounts {
	return &CmdAccounts{
		NetCmd:NetCmd{MsgId: msgId, Coinname:coinname, Method:"new_account", Result:nil, Error:nil},
		Amount:amount}
}

func NewTxCmd(msgId, coinname, chiperKey, to string, amount uint64) (*CmdTx) {
	return &CmdTx{ NetCmd:NetCmd{MsgId: msgId, Coinname:coinname, Method:"send_transaction", Result:nil, Error:nil},
		Chiperkey:chiperKey, Tx:&Transfer{To: to, Value:amount,
		Confirmationsnumber: config.GetConfiger().Clientconfig[coinname].TxConfirmNumber,
		OnBlock:0, PresentBlock:0}}
}

func NewRechargeAddressCmd(msgId, coin string, address []string) (*CmdRechargeAddress) {
	return &CmdRechargeAddress{
		NetCmd:NetCmd{MsgId: msgId, Coinname: coin, Method:"watch_addresses", Result:nil, Error:nil},
		Addresses:address }
}

