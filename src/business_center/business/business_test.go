package business

import (
	"api_router/base/data"
	"business_center/transaction"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestHandleMsg(t *testing.T) {
	svr := NewBusinessSvr()
	svr.InitAndStart(nil)
	var req data.SrvRequest
	var res data.SrvResponse

	testType := 0
	switch testType {
	case 1:
		req.Method.Function = "new_address"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = "{\"asset_name\":\"btc\",\"count\":2}"
	case 2:
		time.Sleep(time.Second * 3)
		req.Method.Function = "withdrawal"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		params := fmt.Sprintf("{\"asset_name\":\"btc\",\"amount\":30, \"address\":\"mxLju5VqXZR6f8aeFvk82Ltv6WmdnYg8KY\","+
			"\"user_order_id\":\"%s\" }", transaction.GenerateUUID("UR"))
		req.Argv.Message = params
	case 3:
		req.Method.Function = "query_user_address"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = "{\"page_index\":1,\"max_disp_lines\":2}"
	case 4:
		req.Method.Function = "support_assets"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = ""
	case 5:
		req.Method.Function = "asset_attribute"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = "{\"asset_names\":[\"btc\", \"eth\"], \"max_disp_lines\":2, \"total_lines\":0}"
	case 6:
		req.Method.Function = "get_balance"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = ""
	case 7:
		req.Method.Function = "history_transaction_bill"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = "{\"trans_type\":0, \"id\":9}"
	case 8:
		req.Method.Function = "history_transaction_message"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = ""
	case 9:
		req.Method.Function = "set_pay_address"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = "{\"asset_name\":\"eth\", \"address\":\"0xC4CD9AA94a7F13dAF7Ff18DA9c830BaA71D41d17\"}"
	case 10:
		req.Method.Function = "query_pay_address"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = "[\"btc\"]"
	case 11:
		req.Method.Function = "transaction_bill_daily"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Argv.Message = ""
	}

	if testType > 0 {
		svr.HandleMsg(&req, &res)
		fmt.Println(res.Value.Message)
	}

	time.Sleep(time.Second * 60 * 60)
	svr.Stop()
}
