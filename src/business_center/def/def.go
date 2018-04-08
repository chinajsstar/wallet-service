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

type UserAddress struct {
	UserID     string `json:"user_id"`
	AssetID    int    `json:"asset_id"`
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
	Enabled    bool   `json:"enabled"`
	CreateTime uint64 `json:"create_time"`
}

type UserAccount struct {
	UserID          string  `json:"user_id"`
	AssetID         int     `json:"asset_id"`
	AvailableAmount float64 `json:"available_amount"`
	FrozenAmount    float64 `json:"frozen_amount"`
	CreateTime      uint64  `json:"create_time"`
	UpdateTime      uint64  `json:"update_time"`
}

type TransactionDetail struct {
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
