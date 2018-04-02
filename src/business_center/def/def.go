package def

type ReqHead struct {
	UserID string `json:"user_id"`
	Method string `json:"method"`
}

type ReqNewAddress struct {
	UserID string `json:"user_id"`
	Method string `json:"method"`
	Params struct {
		ID     string `json:"id"`
		Symbol string `json:"symbol"`
		Count  int    `json:"count"`
	} `json:"params"`
}

type RspNewAddress struct {
	Result struct {
		ID      string   `json:"id"`
		Symbol  string   `json:"symbol"`
		Address []string `json:"address"`
	} `json:"result"`
	Status struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"status"`
}

type ReqWithdrawal struct {
	UserID string `json:"user_id"`
	Method string `json:"method"`
	Params struct {
		UserOrderID   string `json:"user_order_id"`
		Symbol        string `json:"symbol"`
		Amount        int    `json:"amount"`
		ToAddress     string `json:"to_address"`
		UserTimestamp string `json:"user_timestamp"`
	} `json:"params"`
}

type RspWithdrawal struct {
	Result struct {
		UserOrderID string `json:"user_order_id"`
		Timestamp   string `json:"timestamp"`
	} `json:"result"`
	Status struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"status"`
}

type UserAssetAddress struct {
	UserID     string `json:"user_id"`
	AssetID    int    `json:"asset_id"`
	Address    string `json:"address"`
	Enabled    bool   `json:"enabled"`
	CreateTime string `json:"create_time"`
}

type UserAssets struct {
	UserID          string
	AssetID         int
	AvailableAmount float64
	FrozenAmount    float64
	MaxTxBlock      uint64
	CreateTime      string
	UpdateTime      string
}

type Transfer struct {
	AssetID             int
	TxHash              string
	From                string
	To                  string
	Value               uint64
	Gase                uint64
	Gaseprice           uint64
	Total               uint64
	Fee                 uint64
	State               int
	OnBlock             uint64
	PresentBlock        uint64
	Confirmationsnumber uint64
	Time                uint64
}
