package business

import (
	"api_router/base/data"
	"business_center/mysqlpool"
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

	//mysqlpool.QueryUserPropertyByKey("f223c88b-102a-485d-a5da-f96bb55f0bdf")
	//mysqlpool.QueryUserAccount("737205c4-af3c-426d-973d-165a0bf46c71")
	//mysqlpool.QueryUserAccountByUserKey("737205c4-af3c-426d-973d-165a0bf46c71")
	mysqlpool.QueryPayAddress("eth")

	testType := 6
	switch testType {
	case 1:
		req.Data.Method.Function = "new_address"
		req.Data.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Data.Argv.Message = "{\"asset_name\":\"eth\",\"count\":1}"
	case 2:
		req.Data.Method.Function = "withdrawal"
		req.Data.Argv.UserKey = "737205c4-af3c-426d-973d-165a0bf46c71"
		req.Data.Argv.Message = "{\"asset_name\":\"eth\",\"amount\":0.1,\"address\":\"0x357859b176a72f3167e867b1cf0c1e04abba1ce1\"}"
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
		req.Data.Argv.Message = "[\"btc\", \"eth\"]"
	}

	if testType > 0 {
		svr.HandleMsg(&req, &res)
		fmt.Println(res.Data.Value.Message)
	}

	time.Sleep(time.Second * 60 * 60)
	svr.Stop()
}
