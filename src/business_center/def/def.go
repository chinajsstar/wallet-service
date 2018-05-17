package def

type PushMsgCallback func(userID string, callbackMsg string)

const (
	TimeFormat = "2006-01-02 15:04:05"
)

const (
	TypeDeposit = iota
	TypeWithdrawal
)

const (
	StatusBlockin = iota
	StatusConfirm
	StatusFail
)

type UserAccount struct {
	UserKey         string `json:"user_key"`
	UserClass       int    `json:"user_class"`
	AssetName       string `json:"asset_name"`
	AvailableAmount int64  `json:"available_amount"`
	FrozenAmount    int64  `json:"frozen_amount"`
	CreateTime      uint64 `json:"create_time"`
	UpdateTime      uint64 `json:"update_time"`
}

type UserProperty struct {
	UserKey    string `json:"user_key"`
	UserClass  int    `json:"user_class"`
	IsFrozen   int    `json:"is_frozen"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
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
	Decimals              int     `json:"decimals"`
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
	AssetName     string `json:"asset_name"`
	Hash          string `json:"hash"`
	Status        int    `json:"status"`
	MinerFee      int64  `json:"miner_fee"`
	BlockinHeight int64  `json:"blockin_height"`
	OrderID       string `json:"order_id"`
	Time          int64  `json:"blockin_time"`
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

type TransactionNotice struct {
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
