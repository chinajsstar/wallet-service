package def

type PushMsgCallback func(userID string, callbackMsg string)

const (
	TimeFormat = "2006-01-02 15:04:05"
	DateFormat = "20060102"
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

type (
	UserAccount struct {
		UserKey         string  `json:"user_key"`
		UserClass       int     `json:"user_class"`
		AssetName       string  `json:"asset_name"`
		AvailableAmount float64 `json:"available_amount"`
		FrozenAmount    float64 `json:"frozen_amount"`
		CreateTime      int64   `json:"create_time"`
		UpdateTime      int64   `json:"update_time"`
	}

	UserProperty struct {
		UserKey    string `json:"user_key"`
		UserName   string `json:"user_name"`
		UserClass  int    `json:"user_class"`
		IsFrozen   int    `json:"is_frozen"`
		CreateTime int64  `json:"create_time"`
		UpdateTime int64  `json:"update_time"`
	}

	AssetProperty struct {
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
		Enabled               int     `json:"enabled"`
	}

	UserAddress struct {
		UserKey         string  `json:"user_key"`
		UserClass       int     `json:"user_class"`
		AssetName       string  `json:"asset_name"`
		Address         string  `json:"address"`
		PrivateKey      string  `json:"private_key"`
		AvailableAmount float64 `json:"available_amount"`
		FrozenAmount    float64 `json:"frozen_amount"`
		Enabled         int     `json:"enabled"`
		CreateTime      int64   `json:"create_time"`
		AllocationTime  int64   `json:"allocation_time"`
		UpdateTime      int64   `json:"update_time"`
	}

	TransactionBlockin struct {
		AssetName     string  `json:"asset_name"`
		Hash          string  `json:"hash"`
		Status        int     `json:"status"`
		MinerFee      float64 `json:"miner_fee"`
		BlockinHeight int64   `json:"blockin_height"`
		OrderID       string  `json:"order_id"`
		Time          int64   `json:"blockin_time"`
	}

	TransactionDetail struct {
		AssetName string  `json:"asset_name"`
		Address   string  `json:"address"`
		TransType string  `json:"trans_type"`
		Amount    float64 `json:"amount"`
		Hash      string  `json:"hash"`
		DetailID  string  `json:"detail_id"`
	}

	TransactionStatus struct {
		AssetName     string `json:"asset_name"`
		Status        int    `json:"status"`
		ConfirmHeight int64  `json:"confirm_height"`
		ConfirmTime   int64  `json:"confirm_time"`
		UpdateTime    int64  `json:"update_time"`
		OrderID       string `json:"order_id"`
		Hash          string `json:"hash"`
	}

	TransactionNotice struct {
		UserKey       string  `json:"user_key"`
		MsgID         int64   `json:"msg_id"`
		TransType     int     `json:"trans_type"`
		Status        int     `json:"status"`
		BlockinHeight int64   `json:"blockin_height"`
		AssetName     string  `json:"asset_name"`
		Address       string  `json:"address"`
		Amount        float64 `json:"amount"`
		PayFee        float64 `json:"pay_fee"`
		MinerFee      float64 `json:"miner_fee"`
		Balance       float64 `json:"balance"`
		Hash          string  `json:"hash"`
		OrderID       string  `json:"order_id"`
		Time          int64   `json:"time"`
	}

	TransactionBill struct {
		ID              int64   `json:"id"`
		UserKey         string  `json:"user_key"`
		OrderID         string  `json:"order_id"`
		UserOrderID     string  `json:"user_order_id"`
		TransType       int     `json:"trans_type"`
		AssetName       string  `json:"asset_name"`
		Address         string  `json:"address"`
		Amount          float64 `json:"amount"`
		PayFee          float64 `json:"pay_fee"`
		MinerFee        float64 `json:"miner_fee"`
		Balance         float64 `json:"balance"`
		Hash            string  `json:"hash"`
		Status          int     `json:"status"`
		BlockinHeight   int64   `json:"blockin_height"`
		ConfirmHeight   int64   `json:"confirm_height"`
		CreateOrderTime int64   `json:"create_order_time"`
		BlockinTime     int64   `json:"blockin_time"`
		ConfirmTime     int64   `json:"confirm_time"`
	}

	TransactionBillDaily struct {
		Period      int     `json:"period"`
		UserKey     string  `json:"user_key"`
		AssetName   string  `json:"asset_name"`
		SumDPAmount float64 `json:"sum_dp_amount"`
		SumWDAmount float64 `json:"sum_wd_amount"`
		SumPayFee   float64 `json:"sum_pay_fee"`
		SumMinerFee float64 `json:"sum_miner_fee"`
		PreBalance  float64 `json:"pre_balance"`
		LastBalance float64 `json:"last_balance"`
	}

	ProfitBill struct {
		ProfitUserKey string  `json:"profit_user_key"`
		UserKey       string  `json:"user_key"`
		TransType     int     `json:"trans_type"`
		AssetName     string  `json:"asset_name"`
		OrderID       string  `json:"order_id"`
		Hash          string  `json:"hash"`
		Amount        float64 `json:"amount"`
		PayFee        float64 `json:"pay_fee"`
		MinerFee      float64 `json:"miner_fee"`
		Profit        float64 `json:"profit"`
		Time          int64   `json:"time"`
	}

	ProfitBillDaily struct {
		Period        string  `json:"period"`
		ProfitUserKey string  `json:"profit_user_key"`
		AssetName     string  `json:"asset_name"`
		SumProfit     float64 `json:"sum_profit"`
		Time          int64   `json:"time"`
	}
)
