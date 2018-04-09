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
	Tx_state_notfound
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

type Token struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals,string,omitempty"`
}

func (self *Token) String() string {
	return fmt.Sprintf(`
		name:%-8s, symbol:%-8s, decimals:%-8d, address:%s`,
		self.Name, self.Symbol, self.Decimals, self.Address)
}

type CmdSendTx struct {
	NetCmd
	Chiperkey 	string
	Tx       	*Transfer
}

type CmdNewAccounts struct {
	NetCmd
	Amount uint32
}

type CmdRechargeAddress struct {
	NetCmd
	Recall_url string
	Addresses  []string
}

type CmdqueryTx struct {
	NetCmd
	Hash string
}

type CmdqueryBalance struct {
	NetCmd
	Address	string
	Token	*string
}

type NetCmdChannel chan interface{}
type CmdqTxChannel chan *CmdqueryTx
type RechargeTxChannel chan *RechargeTx
type CmdTxChannel chan *CmdSendTx
type TxChannel chan *Transfer
type TxState int

// Transaction of Recharge
type RechargeTx struct {
	Coin_name string
	Tx        *Transfer
	Err		  error
}

type Transfer struct {
	Tx_hash             string
	From                string
	To                  string
	Value               uint64	// 交易金额
	Gase                uint64
	Gaseprice           uint64
	GasUsed             uint64
	Total               uint64	// 总花费金额
	State               TxState
	InBlock             uint64	// 所在块高
	ConfirmatedHeight   uint64	// 确认块高
	Confirmationsnumber uint64	// 需要的确认数
	Time                uint64
	Token               *Token
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

func (tx* Transfer) IsTokenTx() bool {
	return !(tx.Token==nil)
}

func (tx *Transfer)String() string {
	var token_str string
	if tx.IsTokenTx() {
		token_str = tx.Token.String()
	}else {
		token_str = "not a token"
	}

	return fmt.Sprintf(`
	TX      %s
	From:   %s
	To:     %s
	State:  %s
	Value:  %d
	gasfee: %d 
	InBlock:%d
	ConfirmtedBlockHeight: %d
	Token information: %s`,
		tx.Tx_hash,
		tx.From,
		tx.To,
		TxStateString(tx.State),
		tx.Value,
		tx.Minerfee(),
		tx.InBlock, tx.ConfirmatedHeight, token_str)
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

