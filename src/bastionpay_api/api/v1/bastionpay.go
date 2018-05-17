package v1

// 获取支持币种
type ReqSupportAssets struct {
}
type AckSupportAssetList struct {
	Data []string `json:"data" doc:"支持的币种列表"`
}

// 获取币种属性
type ReqAssetsAttributeList struct {
	AssetNames   []string `json:"asset_names" doc:"需要查询属性的币种列表，不空表示精确查找"`
	IsToken      int      `json:"is_token" doc:"是否代币，-1:所有，0：不是代币，非0：代币"`
	TotalLines   int      `json:"total_lines" doc:"总数,0：表示首次查询"`
	PageIndex    int      `json:"page_index" doc:"页索引,1开始"`
	MaxDispLines int      `json:"max_disp_lines" doc:"页最大数,100以下"`
}

type AckAssetsAttribute struct {
	AssetName       string  `json:"asset_name" doc:"币种简称"`
	FullName        string  `json:"full_name" doc:"币种全称"`
	IsToken         int     `json:"is_token" doc:"是否代币，0：不是代币，非0：代币"`
	ParentName      string  `json:"parent_name" doc:"公链平台"`
	DepositMin      float64 `json:"deposit_min" doc:"最小充值"`
	WithdrawalRate  float64 `json:"withdrawal_rate" doc:"单笔费率"`
	WithdrawalValue float64 `json:"withdrawal_value" doc:"单笔费用"`
	ConfirmationNum int     `json:"confirmation_num" doc:"确认数"`
	Decimals        int     `json:"decimals" doc:"精度"`
}
type AckAssetsAttributeList struct {
	Data []AckAssetsAttribute `json:"data" doc:"币种属性列表"`

	TotalLines   int `json:"total_lines" doc:"总数"`
	PageIndex    int `json:"page_index" doc:"页索引"`
	MaxDispLines int `json:"max_disp_lines" doc:"页最大数"`
}

// 获取用户余额
type ReqUserBalance struct {
	Assets []string `json:"assets" doc:"需要查询余额的币种列表"`
}
type AckUserBalance struct {
	AssetName       string  `json:"asset_name" doc:"币种简称"`
	AvailableAmount float64 `json:"available_amount" doc:"可用余额"`
	FrozenAmount    float64 `json:"frozen_amount" doc:"冻结余额"`
}
type AckUserBalanceList struct {
	Data []AckUserBalance `json:"data" doc:"币种余额列表"`
}

// 获取用户地址
type ReqUserAddress struct {
	BeginTime int64  `json:"begin_time" doc:"开始时间, 0表示不限制"`
	EndTime   int64  `json:"eng_time" doc:"结束时间, 0表示不限制"`
	AssetName string `json:"asset_name" doc:"币种"`
	Address   string `json:"address" doc:"地址"`

	TotalLines   int `json:"total_lines" doc:"总数,0：表示首次查询"`
	PageIndex    int `json:"page_index" doc:"页索引,1开始"`
	MaxDispLines int `json:"max_disp_lines" doc:"页最大数，100以下"`
}

type AckUserAddress struct {
	AssetName      string `json:"asset_name" doc:"币种"`
	Address        string `json:"address" doc:"地址"`
	AllocationTime int64  `json:"allocation_time" doc:"分配时间"`
}

type AckUserAddressList struct {
	Data []AckUserAddress `json:"data" doc:"用户地址列表"`

	TotalLines   int `json:"total_lines" doc:"总数"`
	PageIndex    int `json:"page_index" doc:"页索引"`
	MaxDispLines int `json:"max_disp_lines" doc:"页最大数"`
}

// 历史交易订单
type ReqHistoryTransactionOrder struct {
	SerialId  string `json:"serial_id" doc:"流水号"`
	OrderId   string `json:"order_id" doc:"订单号"`
	AssetName string `json:"asset_name" doc:"币种"`
	TransType int    `json:"trans_type" doc:"交易类型"`
	//Status 			int 	`json:"status" doc:"交易状态"`
	Hash          string  `json:"hash" doc:"交易哈希"`
	MaxAmount     float64 `json:"max_amount" doc:"最大金额"`
	MinAmount     float64 `json:"min_amount" doc:"最小金额"`
	MaxUpdateTime int64   `json:"max_update_time" doc:"开始时间"`
	MinUpdateTime int64   `json:"min_update_time" doc:"结束时间"`

	TotalLines   int `json:"total_lines" doc:"总数,0：表示首次查询"`
	PageIndex    int `json:"page_index" doc:"页索引,1开始"`
	MaxDispLines int `json:"max_disp_lines" doc:"页最大数，100以下"`
}

type AckHistoryTransactionOrder struct {
	AssetName string  `json:"asset_name" doc:"币种"`
	TransType int     `json:"trans_type" doc:"交易类型"`
	Status    int     `json:"status" doc:"交易状态"`
	Amount    float64 `json:"amount" doc:"数量"`
	PayFee    float64 `json:"pay_fee" doc:"交易费用"`
	Hash      string  `json:"hash" doc:"交易哈希"`
	OrderId   string  `json:"order_id" doc:"交易订单"`
	Time      int64   `json:"time" doc:"交易时间"`
}

type AckHistoryTransactionOrderList struct {
	Data []AckHistoryTransactionOrder `json:"data" doc:"历史交易订单列表"`

	TotalLines   int `json:"total_lines" doc:"总数"`
	PageIndex    int `json:"page_index" doc:"页索引"`
	MaxDispLines int `json:"max_disp_lines" doc:"页最大数"`
}

// 历史交易消息
type ReqHistoryTransactionMessage struct {
	MaxMsgId int64 `json:"max_msg_id" doc:"最大消息id"`
	MinMsgId int64 `json:"min_msg_id" doc:"最小消息id"`
}

type AckHistoryTransactionMessage struct {
	MsgId         int64   `json:"msg_id" doc:"消息id"`
	TransType     int     `json:"trans_type" doc:"交易类型"`
	Status        int     `json:"status" doc:"交易状态"`
	BlockinHeight int64   `json:"blockin_height" doc:"入块高度"`
	AssetName     string  `json:"asset_name" doc:"币种"`
	Address       string  `json:"address" doc:"地址"`
	Amount        float64 `json:"amount" doc:"数量"`
	PayFee        float64 `json:"pay_fee" doc:"交易费用"`
	Hash          string  `json:"hash" doc:"交易哈希"`
	OrderId       string  `json:"order_id" doc:"交易订单"`
	Time          int64   `json:"time" doc:"交易时间"`
}

type AckHistoryTransactionMessageList struct {
	Data []AckHistoryTransactionMessage `json:"data" doc:"历史交易消息列表"`
}
