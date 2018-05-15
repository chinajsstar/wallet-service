package v1

// 获取支持币种
type ReqSupportAssets struct {

}
type AckSupportAssetList struct {
	Data []string	`json:"data" comment:"支持的币种列表"`
}

// 获取币种属性--分页
type ReqAssetsAttributeList struct {
	Assets 			[]string	`json:"assets" comment:"需要查询属性的币种列表，不空表示精确查找"`

	IsToken         int     	`json:"is_token" comment:"是否代币，-1:所有，0：不是代币，非0：代币"`

	TotalLines 		int 		`json:"total_lines" comment:"总数,0：表示首次查询"`
	PageIndex 		int 		`json:"page_index" comment:"页索引,1开始"`
	MaxDispLines 	int 		`json:"max_disp_lines" comment:"页最大数"，100以下`
}
type AckAssetsAttribute struct{
	AssetId				  int 	  `json:"id" comment:"唯一ID"`
	AssetLogo             string  `json:"asset_logo" comment:"LOGO，未实现"`
	AssetName             string  `json:"asset_name" comment:"币种简称"`
	FullName              string  `json:"full_name" comment:"币种全称"`
	IsToken               int     `json:"is_token" comment:"是否代币，0：不是代币，非0：代币"`
	ParentName            string  `json:"parent_name" comment:"公链平台"`
	DepositMin            float64 `json:"deposit_min" comment:"最小充值"`
	WithdrawalRate        float64 `json:"withdrawal_rate" comment:"单笔费率"`
	WithdrawalValue       float64 `json:"withdrawal_value" comment:"单笔费用"`
	ConfirmationNum       int     `json:"confirmation_num" comment:"确认数"`
	Decimal               int     `json:"decimal" comment:"精度"`
}
type AckAssetsAttributeList struct {
	Data []AckAssetsAttribute `json:"data" comment:"币种属性列表"`

	TotalLines 		int 		`json:"total_lines" comment:"总数"`
	PageIndex 		int 		`json:"page_index" comment:"页索引"`
	MaxDispLines 	int 		`json:"max_disp_lines" comment:"页最大数"`
}

// 获取用户余额-不用分页
type ReqUserBalance struct{
	Assets []string	`json:"assets" comment:"需要查询余额的币种列表"`
}
type AckUserBalance struct{
	AssetName           string  `json:"asset_name" comment:"币种简称"`
	AvailableAmount     float64 `json:"available_amount" comment:"可用余额"`
	FrozenAmount       	float64 `json:"frozen_amount" comment:"冻结余额"`
}
type AckUserBalanceList struct{
	Data []AckUserBalance `json:"data" comment:"币种余额列表"`
}

// 获取用户地址-分页
type ReqUserAddress struct {
	BeginTime 		int64 	`json:"begin_time" comment:"开始时间, 0表示不限制"`
	EndTime 		int64 	`json:"eng_time" comment:"结束时间, 0表示不限制"`
	AssetName 		string 	`json:"asset_name" comment:"币种"`
	Address         string  `json:"address" comment:"地址"`

	TotalLines 		int 		`json:"total_lines" comment:"总数,0：表示首次查询"`
	PageIndex 		int 		`json:"page_index" comment:"页索引,1开始"`
	MaxDispLines 	int 		`json:"max_disp_lines" comment:"页最大数"，100以下`
}

type AckUserAddress struct {
	AssetName       string  `json:"asset_name" comment:"币种"`
	Address         string  `json:"address" comment:"地址"`
	AllocationTime  int64   `json:"allocation_time" comment:"分配时间"`
}

type AckUserAddressList struct {
	Data 			[]AckUserAddress 	`json:"data" comment:"用户地址列表"`

	TotalLines 		int 				`json:"total_lines" comment:"总数"`
	PageIndex 		int 				`json:"page_index" comment:"页索引"`
	MaxDispLines 	int 				`json:"max_disp_lines" comment:"页最大数"`
}

// 历史交易订单
type ReqHistoryTransactionOrder struct {
	SerialId 		string 	`json:"serial_id" comment:"流水号"`
	OrderId 		string 	`json:"order_id" comment:"订单号"`
	AssetName 		string 	`json:"asset_name" comment:"币种"`
	TransType 		int 	`json:"trans_type" comment:"交易类型"`
	//Status 			int 	`json:"status" comment:"交易状态"`
	Hash 			string 	`json:"hash" comment:"交易哈希"`
	MaxAmount		int64   `json:"max_amount" comment:"最大金额"`
	MinAmount		int64   `json:"min_amount" comment:"最小金额"`
	MaxUpdateTime 	int64 	`json:"max_update_time" comment:"开始时间"`
	MinUpdateTime 	int64 	`json:"min_update_time" comment:"结束时间"`

	TotalLines 		int 		`json:"total_lines" comment:"总数,0：表示首次查询"`
	PageIndex 		int 		`json:"page_index" comment:"页索引,1开始"`
	MaxDispLines 	int 		`json:"max_disp_lines" comment:"页最大数"，100以下`
}

type AckHistoryTransactionOrder struct {
	AssetName 		string 	`json:"asset_name" comment:"币种"`
	TransType 		int 	`json:"trans_type" comment:"交易类型"`
	Status 			int 	`json:"status" comment:"交易状态"`
	Amount 			int64 	`json:"amount" comment:"数量"`
	PayFee 			int64 	`json:"pay_fee" comment:"交易费用"`
	Hash 			string 	`json:"hash" comment:"交易哈希"`
	OrderId 		string 	`json:"order_id" comment:"交易订单"`
	Time 			int64 	`json:"time" comment:"交易时间"`
}

type AckHistoryTransactionOrderList struct {
	Data []AckHistoryTransactionOrder `json:"data" comment:"历史交易订单列表"`

	TotalLines 		int 	`json:"total_lines" comment:"总数"`
	PageIndex 		int 	`json:"page_index" comment:"页索引"`
	MaxDispLines 	int 	`json:"max_disp_lines" comment:"页最大数"`
}

// 历史交易消息
type ReqHistoryTransactionMessage struct {
	MaxMsgId 	int64 	`json:"max_msg_id" comment:"最大消息id"`
	MinMsgId 	int64 	`json:"min_msg_id" comment:"最小消息id"`
}

type AckHistoryTransactionMessage struct {
	MsgId 			int64 	`json:"msg_id" comment:"消息id"`
	TransType 		int 	`json:"trans_type" comment:"交易类型"`
	Status 			int 	`json:"status" comment:"交易状态"`
	BlockinHeight 	int64 	`json:"blockin_height" comment:"入块高度"`
	AssetName 		string 	`json:"asset_name" comment:"币种"`
	Address         string  `json:"address" comment:"地址"`
	Amount 			int64 	`json:"amount" comment:"数量"`
	PayFee 			int64 	`json:"pay_fee" comment:"交易费用"`
	Hash 			string 	`json:"hash" comment:"交易哈希"`
	OrderId 		string 	`json:"order_id" comment:"交易订单"`
	Time 			int64 	`json:"time" comment:"交易时间"`
}

type AckHistoryTransactionMessageList struct {
	Data []AckHistoryTransactionMessage `json:"data" comment:"历史交易消息列表"`
}