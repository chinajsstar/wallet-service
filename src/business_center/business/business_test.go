package business

import (
	"api_router/base/data"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestHandleMsg(t *testing.T) {
	svr := NewBusinessSvr()
	svr.InitAndStart(nil)
	var req data.SrvRequestData
	var res data.SrvResponseData

	testType := 0
	switch testType {
	case 1:
		req.Data.Method.Function = "new_address"
		req.Data.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Data.Argv.Message = "{\"asset_name\":\"eth\",\"count\":1}"
	case 2:
		req.Data.Method.Function = "withdrawal"
		req.Data.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Data.Argv.Message = "{\"asset_name\":\"eth\",\"amount\":0.9,\"address\":\"0x357859b176a72f3167e867b1cf0c1e04abba1ce1\"}"
	case 3:
		req.Data.Method.Function = "query_user_address"
		req.Data.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Data.Argv.Message = "{\"page_index\":1,\"max_disp_lines\":9,\"min_create_time\":1523656800}"
	case 4:
		req.Data.Method.Function = "support_assets"
		req.Data.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Data.Argv.Message = ""
	case 5:
		req.Data.Method.Function = "asset_attributie"
		req.Data.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Data.Argv.Message = "[\"btc\", \"eth\"]"
	case 6:
		req.Data.Method.Function = "get_balance"
		req.Data.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Data.Argv.Message = ""
	case 7:
		req.Data.Method.Function = "history_transaction_order"
		req.Data.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Data.Argv.Message = ""
	case 8:
		req.Data.Method.Function = "history_transaction_message"
		req.Data.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Data.Argv.Message = ""
	case 9:
		req.Data.Method.Function = "set_pay_address"
		req.Data.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Data.Argv.Message = "{\"asset_name\":\"eth\", \"address\":\"0xC4CD9AA94a7F13dAF7Ff18DA9c830BaA71D41d17\"}"
	}

	if testType > 0 {
		svr.HandleMsg(&req, &res)
		fmt.Println(res.Data.Value.Message)
	}

	time.Sleep(time.Second * 60 * 60)
	svr.Stop()
}
