package types

import (
	"fmt"
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
	Tx_state_mined					// transaction mined on a block!!
	Tx_state_confirmed              // transaction was mined and stored on a block. confirmed number is 1 or biger
	Tx_state_unconfirmed            // some error happened...need to re send

	Chain_eth = "eth"
	Chain_bitcoin = "btc"

	NetCmdCode_success = iota
	NetCmdCode_failed

)

type TxState int

// Transaction of Recharge
type RechargeTx struct {
	Coin_name string
	Tx        *Transfer
}
type RechargeTxChannel chan *RechargeTx

type Transfer struct {
	Tx_hash 			string
	From				string
	To					string

	Amount  			uint64
	Gase				uint64
	Gaseprice			uint64
	Total				uint64	// amount + gas *gasprice = total

	State				TxState
	OnBlocknumber 		uint64
	PresentBlocknumber 	uint64
	Confirmationsnumber uint64
	Times 				uint64  //TODO
	////fmt.Println("dd-mm-yyyy : ", current.Format("02-01-2006"))
}

func TxStateString(state TxState) string {
	switch state {
	case Tx_state_unkown: {
		return "unkown"
	}
	case Tx_state_commited: {
		return "commit"
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
	return fmt.Sprintf("from:%s, to:%s, amount:%s, state:%s, minerfee:%d, onblocknumber:%d, present block number:%d", tx.From, tx.To, tx.Amount,
		TxStateString(tx.State), tx.Gaseprice * tx.Gase, tx.OnBlocknumber, tx.PresentBlocknumber)
}

type TxNotFoundErr struct {
	tx_info string
}

func (self *TxNotFoundErr)Error() string {
	return self.tx_info
}

func NewTxNotFoundErr(tx *Transfer) *TxNotFoundErr {
	return &TxNotFoundErr{tx_info: fmt.Sprintf("Transaction not found, detail:%s", tx.String())}
}


type NetCmdErr struct {
	Code 		int32
	Message		string
	Data 		interface{}
}

//type NetCmdRlt struct {
//
//}

type NetCmd struct  {
	MsgId  string
	Coin   string
	Method string
	Result interface{}
	Error  *NetCmdErr
}

func NewNetCmdErr(code int32, message string, data interface{}) *NetCmdErr {
	return &NetCmdErr{Code:code, Message:message, Data:data}
}
