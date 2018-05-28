package backend

type (
	ReqPostTransaction struct {
		AssetName string  `json:"asset_name"`
		From      string  `json:"from"`
		To        string  `json:"to"`
		Amount    float64 `json:"amount"`
	}
)
