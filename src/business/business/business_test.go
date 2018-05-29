package business

import (
	"api_router/base/data"
	"business/monitor"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestHandleMsg(t *testing.T) {
	svr := NewServer()
	svr.InitAndStart(nil)
	var req data.SrvRequest
	var res data.SrvResponse

	testType := 9
	switch testType {
	case 1:
		req.Method.Function = "support_assets"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = ""
	case 2:
		req.Method.Function = "asset_attribute"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = "{\"asset_names\":[\"btc\", \"eth\"], \"max_disp_lines\":2, \"total_lines\":0}"
	case 3:
		req.Method.Function = "query_address"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = "{\"page_index\":1,\"max_disp_lines\":2}"
	case 4:
		req.Method.Function = "get_balance"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = ""
	case 5:
		req.Method.Function = "new_address"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = "{\"asset_name\":\"btc\",\"count\":1}"
	case 6:
		time.Sleep(time.Second * 3)
		req.Method.Function = "withdrawal"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		params := fmt.Sprintf("{\"asset_name\":\"btc\",\"amount\":1.9, \"address\":\"mrEfgUBMUM5zjmzSdoBQuodTz16kyZ1tnD\","+
			"\"user_order_id\":\"%s\" }", monitor.GenerateUUID("UR"))
		req.Argv.Message = params
	case 7:
		req.Method.Function = "transaction_bill"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = "{\"trans_type\":0, \"id\":9}"
	case 8:
		req.Method.Function = "transaction_bill_daily"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = ""
	case 9:
		req.Method.Function = "transaction_message"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = ""
	case 10:
		req.Method.Function = "set_pay_address"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = "{\"asset_name\":\"eth\", \"address\":\"0xC4CD9AA94a7F13dAF7Ff18DA9c830BaA71D41d17\"}"
	case 11:
		req.Method.Function = "query_pay_address"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = "{\"asset_names\":[\"btc\"]}"
	case 12:
		req.Method.Function = "set_asset_attribute"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = ""
	case 13:
		req.Method.Function = "sp_get_chain_balance"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = ""
	case 20:
		req.Method.Function = "sp_get_asset_attribute"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = ""
	}

	if testType > 0 {
		svr.HandleMsg(&req, &res)
		fmt.Println(res.Value.Message)
	}

	time.Sleep(time.Second * 60 * 60)
	svr.Stop()
}
