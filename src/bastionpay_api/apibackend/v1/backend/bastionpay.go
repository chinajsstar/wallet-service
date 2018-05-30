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
)
