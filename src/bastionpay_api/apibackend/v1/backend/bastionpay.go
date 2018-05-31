package backend

type (
	SpReqPostTransaction struct {
		AssetName string  `json:"asset_name"`
		From      string  `json:"from"`
		To        string  `json:"to"`
		Amount    float64 `json:"amount"`
	}

	// 获取币种属性
	SpReqAssetsAttributeList struct {
		AssetNames   []string `json:"asset_names" doc:"需要查询属性的币种列表，不空表示精确查找"`
		IsToken      int      `json:"is_token" doc:"是否代币，-1:所有，0：不是代币，非0：代币"`
		Enabled      int      `json:"enabled" doc:"是否支持服务，-1:所有，0：不支持， 1：支持"`
		TotalLines   int      `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex    int      `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines int      `json:"max_disp_lines" doc:"页最大数,100以下"`
	}

	SpAckAssetsAttribute struct {
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

	SpAckAssetsAttributeList struct {
		Data         []SpAckAssetsAttribute `json:"data" doc:"币种属性列表"`
		TotalLines   int                    `json:"total_lines" doc:"总数"`
		PageIndex    int                    `json:"page_index" doc:"页索引"`
		MaxDispLines int                    `json:"max_disp_lines" doc:"页最大数"`
	}

	SpReqSetAssetAttribute struct {
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

	// 获取用户地址
	SpReqUserAddress struct {
		UserKey           string   `json:"user_key" doc:"用户Key"`
		AssetNames        []string `json:"asset_names" doc:"币种"`
		MaxAllocationTime int64    `json:"max_allocation_time" doc:"分配地址时间"`
		MinAllocationTime int64    `json:"min_allocation_time" doc:"分配地址时间"`
		Address           string   `json:"address" doc:"地址"`
		TotalLines        int      `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex         int      `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines      int      `json:"max_disp_lines" doc:"页最大数，100以下"`
	}

	SpAckUserAddress struct {
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

	SpAckUserAddressList struct {
		Data         []SpAckUserAddress `json:"data" doc:"用户地址列表"`
		TotalLines   int                `json:"total_lines" doc:"总数"`
		PageIndex    int                `json:"page_index" doc:"页索引"`
		MaxDispLines int                `json:"max_disp_lines" doc:"页最大数"`
	}

	// 获取用户余额
	SpReqChainBalance struct {
		AssetName    string `json:"asset_name" doc:"需要查询余额的币种列表"`
		Address      string `json:"address"`
		TotalLines   int    `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex    int    `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines int    `json:"max_disp_lines" doc:"页最大数,100以下"`
	}

	SpAckChainBalance struct {
		AssetName string  `json:"asset_name" doc:"币种简称"`
		Address   string  `json:"address"`
		Amount    float64 `json:"available_amount" doc:"可用余额"`
	}

	SpAckChainBalanceList struct {
		Data         []SpAckChainBalance `json:"data" doc:"币种余额列表"`
		TotalLines   int                 `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex    int                 `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines int                 `json:"max_disp_lines" doc:"页最大数,100以下"`
	}

	// 历史交易订单
	SpReqTransactionBill struct {
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

	SpAckTransactionBill struct {
		ID              int64   `json:"id" doc:"流水号"`
		OrderID         string  `json:"order_id" doc:"交易订单"`
		UserOrderID     string  `json:"user_order_id" doc:"用户订单号"`
		AssetName       string  `json:"asset_name" doc:"币种"`
		Address         string  `json:"address" doc:"地址"`
		TransType       int     `json:"trans_type" doc:"交易类型"`
		Amount          float64 `json:"amount" doc:"数量"`
		PayFee          float64 `json:"pay_fee" doc:"交易费用"`
		MinerFee        float64 `json:"miner_fee" doc:"矿工费"`
		Balance         float64 `json:"balance"doc:"当前余额"`
		Hash            string  `json:"hash"`
		Status          int     `json:"status" doc:"交易状态"`
		BlockinHeight   int64   `json:"blockin_height" doc:"入块高度"`
		CreateOrderTime int64   `json:"create_order_time" doc:"订单创建时间"`
		BlockinTime     int64   `json:"blockin_time" doc:"入块时间"`
		ConfirmTime     int64   `json:"confirm_time" doc:"确认时间"`
	}

	SpAckTransactionBillList struct {
		Data         []SpAckTransactionBill `json:"data" doc:"历史交易订单列表"`
		TotalLines   int                    `json:"total_lines" doc:"总数"`
		PageIndex    int                    `json:"page_index" doc:"页索引"`
		MaxDispLines int                    `json:"max_disp_lines" doc:"页最大数"`
	}

	SpReqTransactionBillDaily struct {
		AssetName    string `json:"asset_name" doc:"币种"`
		MaxPeriod    int    `json:"max_period" doc:"最大周期值"`
		MinPeriod    int    `json:"min_period" doc:"最小周期值"`
		TotalLines   int    `json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex    int    `json:"page_index" doc:"页索引,1开始"`
		MaxDispLines int    `json:"max_disp_lines" doc:"页最大数，100以下"`
	}

	SpAckTransactionBillDaily struct {
		Period      int     `json:"period"`
		AssetName   string  `json:"asset_name"`
		SumDPAmount float64 `json:"sum_dp_amount"`
		SumWDAmount float64 `json:"sum_wd_amount"`
		SumPayFee   float64 `json:"sum_pay_fee"`
		SumMinerFee float64 `json:"sum_miner_fee"`
		PreBalance  float64 `json:"pre_balance"`
		LastBalance float64 `json:"last_balance"`
	}

	SpAckTransactionBillDailyList struct {
		Data         []SpAckTransactionBillDaily `json:"data" doc:"历史日结帐单"`
		TotalLines   int                         `json:"total_lines" doc:"总数"`
		PageIndex    int                         `json:"page_index" doc:"页索引"`
		MaxDispLines int                         `json:"max_disp_lines" doc:"页最大数"`
	}
)
