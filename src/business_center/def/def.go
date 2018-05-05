package def

type PushMsgCallback func(userID string, callbackMsg string)

const (
	TimeFormat = "2006-01-02 15:04:05"
)

const (
	TypeDeposit = iota
	TypeWithdrawal
	TypeChange
)

const (
	StatusBlockin = iota
	StatusConfirm
	StatusFail
)

type ParamsMapping struct {
	UserKey   string   `json:"user_key"`
	AssetName string   `json:"asset_name"`
	Address   string   `json:"address"`
	Amount    int64    `json:"amount"`
	Count     int      `json:"count"`
	Params    []string `json:"params"`
}

type ReqHead struct {
	UserID string `json:"user_id"`
	Method string `json:"method"`
}

type ReqNewAddress struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Count  int    `json:"count"`
}

type RspNewAddress struct {
	ID      string   `json:"id"`
	Symbol  string   `json:"symbol"`
	Address []string `json:"address"`
}

type ReqWithdrawal struct {
	UserOrderID   string  `json:"user_order_id"`
	Symbol        string  `json:"symbol"`
	Amount        float64 `json:"amount"`
	ToAddress     string  `json:"to_address"`
	UserTimestamp int64   `json:"user_timestamp"`
}

type RspWithdrawal struct {
	OrderID     string `json:"order_id"`
	UserOrderID string `json:"user_order_id"`
	Timestamp   int64  `json:"timestamp"`
}

type UserAccount struct {
	UserKey         string `json:"user_key"`
	UserClass       int    `json:"user_class"`
	AssetName       string `json:"asset_name"`
	AvailableAmount int64  `json:"available_amount"`
	FrozenAmount    int64  `json:"frozen_amount"`
	CreateTime      uint64 `json:"create_time"`
	UpdateTime      uint64 `json:"update_time"`
}

type TransDetail struct {
	UserID              string `json:"user_id"`
	AssetID             int    `json:"asset_id"`
	TxHash              string `json:"tx_hash"`
	From                string `json:"from"`
	To                  string `json:"to"`
	Value               uint64 `json:"value"`
	Gase                uint64 `json:"gase"`
	Gaseprice           uint64 `json:"gase_price"`
	Total               uint64 `json:"total"`
	Fee                 uint64 `json:"fee"`
	State               int    `json:"state"`
	OnBlock             uint64 `json:"onblock"`
	PresentBlock        uint64 `json:"present_block"`
	ConfirmationsNumber uint64 `json:"confirmations_number"`
	CreateTime          uint64 `json:"create_time"`
	UpdateTime          uint64 `json:"update_time"`
}

type UserProperty struct {
	UserKey       string `json:"user_key"`
	UserName      string `json:"user_name"`
	UserClass     int    `json:"user_class"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Salt          string `json:"salt"`
	Password      string `json:"password"`
	GoogleAuth    string `json:"google_auth"`
	PublicKey     string `json:"public_key"`
	CallbackUrl   string `json:"callback_url"`
	Level         int    `json:"level"`
	LastLoginTime string `json:"last_login_time"`
	LastLoginIp   string `json:"last_login_ip"`
	LastLoginMac  string `json:"last_login_mac"`
	CreateTime    int64  `json:"create_date"`
	UpdateTime    int64  `json:"update_date"`
	IsFrozen      int    `json:"is_frozen"`
	TimeZone      int    `json:"time_zone"`
	Conutry       string `json:"conutry"`
	Language      string `json:"language"`
}

type AssetProperty struct {
	AssetName             string  `json:"asset_name"`
	FullName              string  `json:"full_name"`
	IsToken               int     `json:"is_token"`
	ParentName            string  `json:"parent_name"`
	Logo                  string  `json:"logo"`
	DepositMin            float64 `json:"deposit_min"`
	WithdrawalRate        float64 `json:"withdrawal_rate"`
	WithdrawalValue       float64 `json:"withdrawal_value"`
	WithdrawalReserveRate float64 `json:"withdrawal_reserve_rate"`
	WithdrawalAlertRate   float64 `json:"withdrawal_alert_rate"`
	WithdrawalStategy     float64 `json:"withdrawal_stategy"`
	ConfirmationNum       int     `json:"confirmation_num"`
	Decaimal              int     `json:"decaimal"`
	GasFactor             float64 `json:"gas_factor"`
	Debt                  float64 `json:"debt"`
	ParkAmount            float64 `json:"park_amount"`
}

type UserAddress struct {
	UserKey         string `json:"user_key"`
	UserClass       int    `json:"user_class"`
	AssetName       string `json:"asset_name"`
	Address         string `json:"address"`
	PrivateKey      string `json:"private_key"`
	AvailableAmount int64  `json:"available_amount"`
	FrozenAmount    int64  `json:"frozen_amount"`
	Enabled         int    `json:"enabled"`
	CreateTime      int64  `json:"create_time"`
	AllocationTime  int64  `json:"allocation_time"`
	UpdateTime      int64  `json:"update_time"`
}

type TransactionBlockin struct {
	AssetName     string              `json:"asset_name"`
	Status        int                 `json:"status"`
	MinerFee      int64               `json:"miner_fee"`
	BlockinHeight int64               `json:"blockin_height"`
	BlockinTime   int64               `json:"blockin_time"`
	OrderID       string              `json:"order_id"`
	Hash          string              `json:"hash"`
	Detail        []TransactionDetail `json:"detail"`
}

type TransactionBlockin2 struct {
	AssetName     string              `json:"asset_name"`
	Hash          string              `json:"hash"`
	Status        int                 `json:"status"`
	MinerFee      int64               `json:"miner_fee"`
	BlockinHeight int64               `json:"blockin_height"`
	OrderID       string              `json:"order_id"`
	Time          int64               `json:"blockin_time"`
	Detail        []TransactionDetail `json:"detail"`
}

type TransactionDetail struct {
	AssetName string `json:"asset_name"`
	Address   string `json:"address"`
	TransType string `json:"trans_type"`
	Amount    int64  `json:"amount"`
	Hash      string `json:"hash"`
	DetailID  string `json:"detail_id"`
}

type TransactionStatus struct {
	AssetName     string `json:"asset_name"`
	Status        int    `json:"status"`
	ConfirmHeight int64  `json:"confirm_height"`
	ConfirmTime   int64  `json:"confirm_time"`
	UpdateTime    int64  `json:"update_time"`
	OrderID       string `json:"order_id"`
	Hash          string `json:"hash"`
}

type TransactionNotic struct {
	UserKey       string `json:"user_key"`
	MsgID         int64  `json:"msg_id"`
	Type          int    `json:"type"`
	Status        int    `json:"status"`
	BlockinHeight int64  `json:"blockin_height"`
	AssetName     string `json:"asset_name"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	PayFee        int64  `json:"pay_fee"`
	Hash          string `json:"hash"`
	OrderID       string `json:"order_id"`
	Time          int64  `json:"time"`
}

type TransactionOrder struct {
	UserKey   string `json:"user_key"`
	AssetName string `json:"asset_name"`
	TransType int    `json:"trans_type"`
	Amount    int64  `json:"amount"`
	PayFee    int64  `json:"pay_fee"`
	Hash      string `json:"hash"`
	OrderID   string `json:"order_id"`
	Status    int    `json:"status"`
	Time      int64  `json:"time"`
}

type TransactionMessage struct {
	UserKey       string `json:"user_key"`
	MsgID         int64  `json:"msg_id"`
	TransType     int    `json:"trans_type"`
	Status        int    `json:"status"`
	BlockinHeigth int64  `json:"blockin_height"`
	AssetName     string `json:"asset_name"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	PayFee        int64  `json:"pay_fee"`
	Hash          string `json:"hash"`
	OrderID       string `json:"order_id"`
	Time          int64  `json:"time"`
}
