package business

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
