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
	UserKey         string  `json:"user_key"`
	UserClass       int     `json:"user_class"`
	AssetID         int     `json:"asset_id"`
	AvailableAmount float64 `json:"available_amount"`
	FrozenAmount    float64 `json:"frozen_amount"`
	CreateTime      uint64  `json:"create_time"`
	UpdateTime      uint64  `json:"update_time"`
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
	Password      string `json:"password"`
	GoogleAuth    string `json:"google_auth"`
	PublicKey     string `json:"public_key"`
	Level         int    `json:"level"`
	LastLoginTime string `json:"last_login_time"`
	LastLoginIp   string `json:"last_login_ip"`
	LastLoginMac  string `json:"last_login_mac"`
	CreateTime    string `json:"create_date"`
	UpdateTime    string `json:"update_date"`
	IsFrozen      int    `json:"is_frozen"`
	TimeZone      int    `json:"time_zone"`
	Conutry       string `json:"conutry"`
}

type AssetProperty struct {
	ID                    int     `json:"id"`
	Name                  string  `json:"name"`
	FullName              string  `json:"full_name"`
	IsToken               int     `json:"is_token"`
	CoinName              string  `json:"coin_name"`
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
	UserName        string `json:"user_name"`
	UserClass       int    `json:"user_class"`
	AssetID         int    `json:"asset_id"`
	AssetName       string `json:"asset_name"`
	AssetFullName   string `json:"asset_full_name"`
	Address         string `json:"address"`
	PrivateKey      string `json:"private_key"`
	AvailableAmount int64  `json:"available_amount"`
	FrozenAmount    int64  `json:"frozen_amount"`
	Enabled         int    `json:"enabled"`
	CreateTime      int64  `json:"create_time"`
	UpdateTime      int64  `json:"update_time"`
}

type TransactionBlockin struct {
	AssetID       int                 `json:"asset_id"`
	Hash          string              `json:"hash"`
	Status        int                 `json:"status"`
	MinerFee      int64               `json:"miner_fee"`
	AssetName     string              `json:"asset_name"`
	BlockinHeight int64               `json:"blockin_height"`
	BlockinTime   int64               `json:"blockin_time"`
	OrderID       string              `json:"order_id"`
	Detail        []TransactionDetail `json:"detail"`
}

type TransactionDetail struct {
	AssetID   int    `json:"asset_id"`
	Address   string `json:"address"`
	TransType string `json:"trans_type"`
	Amount    int64  `json:"amount"`
	MinerFee  int64  `json:"miner_fee"`
	Hash      string `json:"hash"`
	DetailID  string `json:"detail_id"`
}

type TransactionStatus struct {
	AssetID       int    `json:"asset_id"`
	Hash          string `json:"hash"`
	AssetName     string `json:"asset_name"`
	Status        int    `json:"status"`
	ConfirmHeight int64  `json:"confirm_height"`
	ConfirmTime   int64  `json:"confirm_time"`
	UpdateTime    int64  `json:"update_time"`
	OrderID       string `json:"order_id"`
}

type TransactionNotic struct {
	UserKey       string `json:"user_key"`
	MsgID         int64  `json:"msg_id"`
	Type          int    `json:"type"`
	Status        int    `json:"status"`
	BlockinHeight int64  `json:"blockin_height"`
	AssetID       int    `json:"asset_id"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	WalletFee     int64  `json:"wallet_fee"`
	Hash          string `json:"hash"`
	Time          int64  `json:"time"`
}
