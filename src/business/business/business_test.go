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

	testType := 23
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
		req.Argv.Message = "{\"asset_name\":\"eth\",\"count\":1}"
	case 6:
		time.Sleep(time.Second * 3)
		req.Method.Function = "withdrawal"
		req.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		params := fmt.Sprintf("{\"asset_name\":\"btc\",\"amount\":1.5, \"address\":\"mrEfgUBMUM5zjmzSdoBQuodTz16kyZ1tnD\","+
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
		req.Method.Function = "sp_set_asset_attribute"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		//req.Argv.Message = "{\"asset_name\":\"BTC\",\"full_name\":\"Bitcoin\",\"is_token\":0,\"parent_name\":\"\",\"logo\":\"\",\"deposit_min\":1,\"withdrawal_rate\":0.001,\"withdrawal_value\":0.002,\"withdrawal_reserve_rate\":0.001,\"withdrawal_alert_rate\":1,\"withdrawal_stategy\":5,\"confirmation_num\":2,\"decimals\":8,\"gas_factor\":0.00005,\"debt\":2,\"park_amount\":1,\"enabled\":1}"
		req.Argv.Message = "{\"enabled\":0,\"deposit_min\":2,\"is_token\":0,\"confirmation_num\":2,\"decimals\":2,\"withdrawal_alert_rate\":2,\"withdrawal_reserve_rate\":2,\"withdrawal_rate\":2,\"withdrawal_stategy\":2,\"withdrawal_value\":2,\"logo\":\"\",\"asset_name\":\"2\",\"full_name\":\"2\",\"parent_name\":\"\"}"
	case 13:
		req.Method.Function = "sp_get_chain_balance"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = "{\"asset_name\":\"eth\", \"address\":\"0x5563eaB8a68D36156E15621b7D85Ac215C477434\"}"
	case 20:
		req.Method.Function = "sp_get_asset_attribute"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = ""
	case 21:
		req.Method.Function = "sp_post_transaction"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = "{\"asset_name\":\"eth\", \"address\":\"0x5563eaB8a68D36156E15621b7D85Ac215C477434\"}"
	case 22:
		req.Method.Function = "block_height"
		req.Argv.UserKey = "795b587d-2ee7-4979-832d-5d0ea64205d5"
		req.Argv.Message = ""
	case 23:
		req.Method.Function = "sp_transaction_bill"
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
