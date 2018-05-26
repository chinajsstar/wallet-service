package v1

// 获取支持币种
type (
	ReqNewAddress struct {
		AssetName string `json:"asset_name" doc:"币种简称"`
		Count     int    `json:"count" doc:"数量"`
	}

	AckNewAddressList struct {
		AssetName string   `json:"asset_name" doc:"币种简称"`
		Data      []string `json:"data" doc:"新生成的地址"`
	}

	ReqWithdrawal struct {
		AssetName   string  `json:"asset_name" doc:"币种简称"`
		Amount      float64 `json:"" doc:"数量"`
		Address     string  `json:"" doc:"提币地址"`
		UserOrderID string  `json:"" doc:"用户自定义序号"`
	}

	AckWithdrawal struct {
		OrderID     string `json:"" doc:"交易订单号"`
		UserOrderID string `json:"" doc:"用户自定义序号"`
	}

	ReqSupportAssets struct{}

	AckSupportAssetList struct {
		Data []string `json:"data" doc:"支持的币种列表"`
	}

	// 获取币种属性
	ReqAssetsAttributeList struct {
		AssetNames   []string `json:"asset_names" doc:"需要查询属性的币种列表，不空表示精确查找"`
		IsToken      int      `json:"is_token" doc:"是否代币，-1:所有，0：不是代币，非0：代币"`
		TotalLines   int      `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex    int      `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines int      `json:"max_disp_lines" doc:"页最大数,100以下"`
	}

	AckAssetsAttribute struct {
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

	AckAssetsAttributeList struct {
		Data         []AckAssetsAttribute `json:"data" doc:"币种属性列表"`
		TotalLines   int                  `json:"total_lines" doc:"总数"`
		PageIndex    int                  `json:"page_index" doc:"页索引"`
		MaxDispLines int                  `json:"max_disp_lines" doc:"页最大数"`
	}

	ReqSetAssetAttribute struct {
		AssetName       string  `json:"asset_name" doc:"币种简称"`
		FullName        string  `json:"full_name" doc:"币种全称"`
		IsToken         int     `json:"is_token" doc:"是否代币，0：不是代币，非0：代币"`
		ParentName      string  `json:"parent_name" doc:"公链平台"`
		Logo            string  `json:"logo" doc:"图标"`
		DepositMin      float64 `json:"deposit_min" doc:"最小充值"`
		WithdrawalRate  float64 `json:"withdrawal_rate" doc:"单笔费率"`
		WithdrawalValue float64 `json:"withdrawal_value" doc:"单笔费用"`
		ConfirmationNum int     `json:"confirmation_num" doc:"确认数"`
		Decimals        int     `json:"decimals" doc:"精度"`
	}

	// 获取用户余额
	ReqUserBalance struct {
		AssetNames   []string `json:"asset_names" doc:"需要查询余额的币种列表"`
		TotalLines   int      `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex    int      `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines int      `json:"max_disp_lines" doc:"页最大数,100以下"`
	}

	AckUserBalance struct {
		AssetName       string  `json:"asset_name" doc:"币种简称"`
		AvailableAmount float64 `json:"available_amount" doc:"可用余额"`
		FrozenAmount    float64 `json:"frozen_amount" doc:"冻结余额"`
		Time            int64   `json:"time" doc:"刷新时间"`
	}

	AckUserBalanceList struct {
		Data         []AckUserBalance `json:"data" doc:"币种余额列表"`
		TotalLines   int              `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex    int              `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines int              `json:"max_disp_lines" doc:"页最大数,100以下"`
	}

	// 获取用户地址
	ReqUserAddress struct {
		AssetNames        []string `json:"asset_names" doc:"币种"`
		MaxAllocationTime int64    `json:"max_allocation_time" doc:"分配地址时间"`
		MinAllocationTime int64    `json:"min_allocation_time" doc:"分配地址时间"`
		Address           string   `json:"address" doc:"地址"`
		TotalLines        int      `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex         int      `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines      int      `json:"max_disp_lines" doc:"页最大数，100以下"`
	}

	AckUserAddress struct {
		AssetName      string `json:"asset_name" doc:"币种"`
		Address        string `json:"address" doc:"地址"`
		AllocationTime int64  `json:"allocation_time" doc:"分配时间"`
	}

	AckUserAddressList struct {
		Data         []AckUserAddress `json:"data" doc:"用户地址列表"`
		TotalLines   int              `json:"total_lines" doc:"总数"`
		PageIndex    int              `json:"page_index" doc:"页索引"`
		MaxDispLines int              `json:"max_disp_lines" doc:"页最大数"`
	}

	// 历史交易订单
	ReqHistoryTransactionBill struct {
		ID             int64   `json:"id" doc:"流水号"`
		OrderID        string  `json:"order_id" doc:"订单号"`
		AssetName      string  `json:"asset_name" doc:"币种"`
		Address        string  `json:"address" doc:"地址"`
		TransType      int     `json:"trans_type" doc:"交易类型"`
		Status         int     `json:"status" doc:"交易状态"`
		Hash           string  `json:"hash" doc:"交易哈希"`
		MaxAmount      float64 `json:"max_amount" doc:"最大金额"`
		MinAmount      float64 `json:"min_amount" doc:"最小金额"`
		MaxConfirmTime int64   `json:"max_confirm_time" doc:"开始时间"`
		MinConfirmTime int64   `json:"min_confirm_time" doc:"结束时间"`
		TotalLines     int     `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex      int     `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines   int     `json:"max_disp_lines" doc:"页最大数，100以下"`
	}

	AckHistoryTransactionBill struct {
		ID              int64   `json:"id" doc:"流水号"`
		OrderID         string  `json:"order_id" doc:"交易订单"`
		UserOrderID     string  `json:"user_order_id" doc:"用户订单号"`
		AssetName       string  `json:"asset_name" doc:"币种"`
		Address         string  `json:"address" doc:"地址"`
		TransType       int     `json:"trans_type" doc:"交易类型"`
		Amount          float64 `json:"amount" doc:"数量"`
		PayFee          float64 `json:"pay_fee" doc:"交易费用"`
		Balance         float64 `json:"balance"doc:"当前余额"`
		Hash            string  `json:"hash"`
		Status          int     `json:"status" doc:"交易状态"`
		BlockinHeight   int64   `json:"blockin_height" doc:"入块高度"`
		CreateOrderTime int64   `json:"create_order_time" doc:"订单创建时间"`
		BlockinTime     int64   `json:"blockin_time" doc:"入块时间"`
		ConfirmTime     int64   `json:"confirm_time" doc:"确认时间"`
	}

	AckHistoryTransactionBillList struct {
		Data         []AckHistoryTransactionBill `json:"data" doc:"历史交易订单列表"`
		TotalLines   int                         `json:"total_lines" doc:"总数"`
		PageIndex    int                         `json:"page_index" doc:"页索引"`
		MaxDispLines int                         `json:"max_disp_lines" doc:"页最大数"`
	}

	// 历史交易消息
	ReqHistoryTransactionMessage struct {
		MaxMessageID int64 `json:"max_msg_id" doc:"最大消息id"`
		MinMessageID int64 `json:"min_msg_id" doc:"最小消息id"`
		TotalLines   int   `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex    int   `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines int   `json:"max_disp_lines" doc:"页最大数，100以下"`
	}

	AckHistoryTransactionMessage struct {
		MsgID         int64   `json:"msg_id" doc:"消息id"`
		TransType     int     `json:"trans_type" doc:"交易类型"`
		Status        int     `json:"status" doc:"交易状态"`
		BlockinHeight int64   `json:"blockin_height" doc:"入块高度"`
		AssetName     string  `json:"asset_name" doc:"币种"`
		Address       string  `json:"address" doc:"地址"`
		Amount        float64 `json:"amount" doc:"数量"`
		PayFee        float64 `json:"pay_fee" doc:"交易费用"`
		Balance       float64 `json:"balance" doc:"当前余额"`
		Hash          string  `json:"hash" doc:"交易哈希"`
		OrderId       string  `json:"order_id" doc:"交易订单"`
		Time          int64   `json:"time" doc:"交易时间"`
	}

	AckHistoryTransactionMessageList struct {
		Data         []AckHistoryTransactionMessage `json:"data" doc:"历史交易消息列表"`
		TotalLines   int                            `json:"total_lines" doc:"总数"`
		PageIndex    int                            `json:"page_index" doc:"页索引"`
		MaxDispLines int                            `json:"max_disp_lines" doc:"页最大数"`
	}

	PushTransactionMessage struct {
		MsgID         int64   `json:"msg_id"`
		TransType     int     `json:"trans_type"`
		Status        int     `json:"status"`
		BlockinHeight int64   `json:"blockin_height"`
		AssetName     string  `json:"asset_name"`
		Address       string  `json:"address"`
		Amount        float64 `json:"amount"`
		PayFee        float64 `json:"pay_fee"`
		Balance       float64 `json:"balance"`
		Hash          string  `json:"hash"`
		OrderID       string  `json:"order_id"`
		Time          int64   `json:"time"`
	}

	ReqTransactionBillDaily struct {
		AssetName    string `json:"asset_name" doc:"币种"`
		MaxPeriod    int    `json:"max_period" doc:"最大周期值"`
		MinPeriod    int    `json:"min_period" doc:"最小周期值"`
		TotalLines   int    `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex    int    `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines int    `json:"max_disp_lines" doc:"页最大数，100以下"`
	}

	AckTransactionBillDaily struct {
		Period      int     `json:"period"`
		AssetName   string  `json:"asset_name"`
		SumDPAmount float64 `json:"sum_dp_amount"`
		SumWDAmount float64 `json:"sum_wd_amount"`
		SumPayFee   float64 `json:"sum_pay_fee"`
		PreBalance  float64 `json:"pre_balance"`
		LastBalance float64 `json:"last_balance"`
	}

	AckTransactionBillDailyList struct {
		Data         []AckTransactionBillDaily `json:"data" doc:"历史日结帐单"`
		TotalLines   int                       `json:"total_lines" doc:"总数"`
		PageIndex    int                       `json:"page_index" doc:"页索引"`
		MaxDispLines int                       `json:"max_disp_lines" doc:"页最大数"`
	}

	ReqPayAddress struct {
		AssetNames []string `json:"asset_names" doc:"需要查询属性的币种列表，不空表示精确查找"`
	}

	AckPayAddress struct {
		AssetName  string  `json:"asset_name"`
		Address    string  `json:"address"`
		Amount     float64 `json:"amount"`
		UpdateTime int64   `json:"update_time"`
	}

	AckPayAddressList struct {
		Data []AckPayAddress
	}
)
