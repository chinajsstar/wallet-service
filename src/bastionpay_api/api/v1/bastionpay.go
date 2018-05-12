package v1

// TODO:xuliang 添加
// 获取支持币种
//type ReqSupportAssets string
//type AckSupportAssets []string

// 获取币种属性
//type ReqAssetsAttribute []string
type AckAssetsAttribute struct{
	AssetName             string  `json:"asset_name" comment:"币种简称"`
	FullName              string  `json:"full_name" comment:"币种全称"`
	IsToken               int     `json:"is_token" comment:"是否代币"`
	ParentName            string  `json:"parent_name" comment:"爸爸币"`
	DepositMin            float64 `json:"deposit_min" comment:"币种简称"`
	WithdrawalRate        float64 `json:"withdrawal_rate" comment:"币种简称"`
	WithdrawalValue       float64 `json:"withdrawal_value" comment:"币种简称"`
	ConfirmationNum       int     `json:"confirmation_num" comment:"币种简称"`
	Decimal               int     `json:"decaimal" comment:"币种简称"`
}
type AckAssetsAttributes []AckAssetsAttribute