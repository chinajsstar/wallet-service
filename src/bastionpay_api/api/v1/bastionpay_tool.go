package v1

// 模拟充值
type ReqRecharge struct {
	Coin  string `json:"coin" comment:"币种"`
	To    string `json:"to" comment:"充值地址"`
	Value uint64 `json:"value" comment:"数量，为币种的单位的10^-8"`
}

// 模拟挖矿
type ReqGenerate struct {
	Coin  string `json:"coin" comment:"币种，支持btc"`
	Count int `json:"count" comment:"块数量"`
}