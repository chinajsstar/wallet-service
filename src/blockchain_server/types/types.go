package types

import (
	"blockchain_server/utils"
	"fmt"
	"math"
	"math/big"
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

func (self *Account)String() string {
	return fmt.Sprintf(`
PrivateKey:	"%s",
Address:	"%s"`, self.PrivateKey, self.Address)
}

const (
	Tx_state_ToBuild = iota // 尚未准备好, 调用buildTx后, 状态成为BuildOk
	Tx_state_BuildOk        // ready to send
	Tx_state_Signed         // build, and signed, ready for send!

	Tx_state_pending   		// pending Transaction on node
	Tx_state_mined     		// transaction mined on a block!!
	Tx_state_confirmed 		// transaction was mined and stored on a block. confirmed number is 1 or biger

	Tx_state_unconfirmed 	// some error happened...need to re send

	Chain_eth     = "ETH"
	Chain_bitcoin = "BTC"

	NetCmdCode_success = iota
	NetCmdCode_failed

	// online  模式不允许保存私钥
	// offline 模式可以都保存
	Onlinemode_offline = "offline"
	Onlinemode_online  = "online"
)

var txStateNames = map[TxState]string{
	Tx_state_ToBuild:     "to build",
	Tx_state_BuildOk:     "build ok",
	Tx_state_Signed:      "signed",
	Tx_state_pending:     "pending",
	Tx_state_mined:       "mined",
	Tx_state_confirmed:   "confirmed",
	Tx_state_unconfirmed: "unconfirmed",
}

type Token struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals,string,omitempty"`
}

// Transaction中的TokenTx
type TokenTx struct {
	From  string
	To    string
	Value float64
	Token *Token
}

func (self *TokenTx) String() string {
	return fmt.Sprintf(
		`TokenTransaction inormation: 
		TokenAddress: %s
		Symbol		: %s
		From		: %s
		To  		: %s
		Value		: %f`,
		self.Token.Address, self.Token.Symbol, self.From, self.To, self.Value)
}

func (self *TokenTx) Value_decimaled() *big.Int {
	if !self.IsValid() {
		return nil
	}
	return self.Token.Dodecimal(self.Value)
}

func (self *TokenTx) SetValue_by_decimaled(value *big.Int) float64 {
	if isValid := self.IsValid(); isValid == false {
		return 0
	}
	self.Value = self.Token.Undecimal(value)
	return self.Value
}

func (self *TokenTx) IsValid() bool {
	return self.Token != nil
}

func (self *TokenTx) Name() string {
	if self.Token != nil {
		return self.Token.Name
	}
	return ""
}

func (self *TokenTx) Symbol() string {
	if self.Token != nil {
		return self.Token.Symbol
	}
	return ""
}

func (self *TokenTx) ContractAddress() string {
	if self.Token != nil {
		return self.Token.Address
	}
	return ""
}

func (self *Token) Dodecimal(v float64) *big.Int {
	val := v * math.Pow10(self.Decimals)
	ival, _ := new(big.Int).SetString(fmt.Sprintf("%.0f", val), 10)
	return ival
}

func (self *Token) Undecimal(v *big.Int) float64 {
	val := new(big.Float).SetInt(v)
	val = val.Mul(val, big.NewFloat(math.Pow10(-self.Decimals)))
	f, _ := val.Float64()
	return utils.PrecisionN(f, 6)
}

func (self *Token) String() string {
	return fmt.Sprintf(`
		symbol:%-5s, decimals:%-2d, address:%s`,
		self.Symbol, self.Decimals, self.Address)
}

type CmdSendTx struct {
	NetCmd
	FromKey string
	Tx      *Transfer
	// liuheng add
	SignedTxString string // 已签名交易(数据[]byte经base64编码过)，空表示没有签名
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
	Address     string
	TokenSymbol string
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
	Err       error
}

type Transfer struct {
	Tx_hash string
	From    string
	To      string
	Value   float64 // 交易金额
	Fee     float64
	//Gas                 uint64	// deprecated!!!!
	Total               float64 // 总花费金额
	State               TxState
	InBlock             uint64 // 所在块高
	ConfirmatedHeight   uint64 // 确认块高
	Confirmationsnumber uint64 // 需要的确认数
	Time                uint64
	TokenFromKey        string
	TokenTx             *TokenTx

	// 根据不同种类的币种, 有不同!只有其自己才能理解
	Additional_data []byte
	////fmt.Println("dd-mm-yyyy : ", current.Format("02-01-2006"))
}

func (tx *Transfer) Tatolcost() float64 {
	return tx.Fee + tx.Value
}

func TxStateString(state TxState) (name string) {
	name = txStateNames[state]
	if name=="" {
		name="unknow state"
	}
	return
}

func (tx *Transfer) IsTokenTx() bool {
	return !(tx.TokenTx == nil)
}

func (tx *Transfer) String() string {
	var token_str string
	if tx.IsTokenTx() {
		token_str = tx.TokenTx.String()
	} else {
		token_str = "not a token"
	}

	return fmt.Sprintf(`
	TxHash: %s
	From:   %s
	To:     %s
	State:  %s
	Value:  %f
	Fee: 	%f 
	InBlock:%d
	ConfirmtedBlockHeight: %d
	%s`,
		tx.Tx_hash,
		tx.From,
		tx.To,
		TxStateString(tx.State),
		tx.Value,
		tx.Fee,
		tx.InBlock, tx.ConfirmatedHeight, token_str)
}

type NetCmdErr struct {
	Code    int32
	Message string
	Data    interface{}
}

type NetCmd struct {
	MsgId    string
	Coinname string
	Method   string
	Result   interface{}
	Error    *NetCmdErr
}

func NewNetCmdErr(code int32, message string, data interface{}) *NetCmdErr {
	return &NetCmdErr{Code: code, Message: message, Data: data}
}
