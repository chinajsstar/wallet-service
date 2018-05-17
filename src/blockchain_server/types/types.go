package types

import (
	"blockchain_server/utils"
	"fmt"
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

const (
	Tx_state_ToBuild = iota // 尚未准备好, 调用buildTx后, 状态成为BuildOk
	Tx_state_BuildOk   // ready to send
	Tx_state_Signed    // build, and signed, ready for send!

	Tx_state_pending   // pending Transaction on node
	Tx_state_mined     // transaction mined on a block!!
	Tx_state_confirmed // transaction was mined and stored on a block. confirmed number is 1 or biger

	Tx_state_unconfirmed // some error happened...need to re send

	Chain_eth     = "eth"
	Chain_bitcoin = "btc"

	NetCmdCode_success = iota
	NetCmdCode_failed

	// online  模式不允许保存私钥
	// offline 模式可以都保存
	Onlinemode_offline = "offline"
	Onlinemode_online  = "online"

	StandardDecimal = 8
)

type Token struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals uint   `json:"decimals,string,omitempty"`
}

// Transaction中的TokenTx
type TokenTx struct {
	From     string
	To       string
	Value    uint64
	Contract *Token
}

func (self *TokenTx) String() string {
	return fmt.Sprintf(
`TokenTransaction inormation: 
	Symbol:			: %s
	Contract		: %s
	From:			: %s
	To				: %s
	Value			: %d`,
	self.Contract.Symbol, self.Contract.Address, self.From, self.To, self.Value)
}

func (self *TokenTx) TokenDecimalValue() uint64 {
	return utils.DecimalCvt_i_i(self.Value, 8, 0).Uint64()
	//return self.Contract.ToTokenDecimal(self.Value).Uint64()
}

func (self *TokenTx) IsValid() bool {
	return self.Contract != nil
}

func (self *TokenTx) Name() string {
	if self.Contract != nil {
		return self.Contract.Name
	}
	return ""
}

func (self *TokenTx) Symbol() string {
	if self.Contract != nil {
		return self.Contract.Symbol
	}
	return ""
}

func (self *TokenTx) ContractAddress() string {
	if self.Contract != nil {
		return self.Contract.Address
	}
	return ""
}

// 把StandardDecimal 表示的数量, 转换为币种内部使用的数量
func (self *Token) ToTokenDecimal(v uint64) *big.Int {
	return utils.DecimalCvt_i_i(v, 8, 0)
	//i := int(self.Decimals) - 8
	//ibig := big.NewInt(int64(v))
	//if i > 0 {
	//	return ibig.Mul(ibig, big.NewInt(int64(math.Pow10(i))))
	//} else {
	//	return ibig.Div(ibig, big.NewInt(int64(math.Pow10(-i))))
	//}
}

// 把币种内部使用的精度表示为外部标准使用过的精度!
func (self *Token) ToStandardDecimal(v uint64) uint64 {
	return utils.DecimalCvt_i_i(v, int(self.Decimals), 8).Uint64()
}

func (self *Token) ToStandardDecimalWithBig(ibig *big.Int) uint64 {
	return ibig.Uint64()
	//v := ibig.Uint64()
	//return utils.DecimalCvt_i_i(v, int(self.Decimals), 8).Uint64()
}

func (self *Token) String() string {
	return fmt.Sprintf(`
		name:%-8s, symbol:%-8s, decimals:%-8d, address:%s`,
		self.Name, self.Symbol, self.Decimals, self.Address)
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
	Address string
	Token   *string
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
	Tx_hash             string
	From                string
	To                  string
	Value               uint64 // 交易金额
	Fee                 uint64
	Gaseprice           uint64
	Gas                 uint64
	Total               uint64 // 总花费金额
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

func (tx *Transfer) Tatolcost() uint64 {
	return tx.Fee + tx.Value
}

func TxStateString(state TxState) string {
	switch state {
	case Tx_state_ToBuild:
		{
			return "To build"
		}
	case Tx_state_BuildOk:
		{
			return "Build Ok"
		}
	case Tx_state_pending:
		{
			return "Pending"
		}
	case Tx_state_mined:
		{
			return "mined"
		}
	case Tx_state_confirmed:
		{
			return "confirmed"
		}
	case Tx_state_unconfirmed:
		{
			return "unconfirmed"
		}
	default:
		return "unkown"
	}
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
