package handler

import (
	"api_router/base/data"
	//"api_router/base/service"
	"api_router/base/config"
	service "api_router/base/service2"
	"bastionpay_api/apibackend"
	"business/business"
	l4g "github.com/alecthomas/log4go"
)

type Cobank struct {
	business *business.Business

	node *service.ServiceNode

	// http 服务的数据目录
	dataDir string
}

func NewCobank(dataDir string) *Cobank {
	x := &Cobank{}
	x.business = &business.Business{}
	x.dataDir = dataDir
	return x
}

func (x *Cobank) Start(node *service.ServiceNode) error {
	x.node = node
	return x.business.InitAndStart(x.callBack)
}

func (x *Cobank) Stop() {
	x.business.Stop()
}

func (x *Cobank) callBack(userID string, callbackMsg string) {
	pData := data.SrvRequest{}
	pData.Method.Version = "v1"
	pData.Method.Srv = "push"
	pData.Method.Function = "pushdata"

	pData.Argv.UserKey = userID
	pData.Argv.Message = callbackMsg

	res := data.SrvResponse{}
	x.node.InnerCallByEncrypt(&pData, &res)
	l4g.Info("push return: ", res)
}

func (x *Cobank) GetApiGroup() map[string]service.NodeApi {
	nam := make(map[string]service.NodeApi)

	func() {
		service.RegisterApi(&nam,
			"new_address", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"query_address", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"withdrawal", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"support_assets", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"asset_attribute", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"get_balance", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"transaction_bill", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"transaction_bill_daily", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"transaction_message", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"block_height", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"sp_get_asset_attribute", data.APILevel_admin, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"sp_set_asset_attribute", data.APILevel_admin, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"sp_get_pay_address", data.APILevel_admin, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"sp_set_pay_address", data.APILevel_admin, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"sp_post_transaction", data.APILevel_admin, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"sp_query_address", data.APILevel_admin, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"sp_get_chain_balance", data.APILevel_admin, x.handler)
	}()

	////////////////////////////////////////////////////////////////
	// 以下方法liuheng添加测试
	var testTool int
	err := config.LoadJsonNode(config.GetBastionPayConfigDir()+"/cobank.json", "test_tool", &testTool)
	l4g.Info("test tool mode: %d", testTool)
	if err == nil && testTool == 1 {
		func() {
			service.RegisterApi(&nam,
				"recharge", data.APILevel_client, x.recharge)
		}()

		func() {
			service.RegisterApi(&nam,
				"generate", data.APILevel_client, x.generate)
		}()

		func() {
			service.RegisterApi(&nam,
				"importaddress", data.APILevel_admin, x.importAddress)
		}()

		func() {
			service.RegisterApi(&nam,
				"sendsignedtx", data.APILevel_admin, x.sendSignedTx)
		}()
	}

	return nam
}

func (x *Cobank) HandleNotify(req *data.SrvRequest) {
	l4g.Info("HandleNotify-reloadUserLevel: do nothing")
}

func (x *Cobank) handler(req *data.SrvRequest, res *data.SrvResponse) {
	res.Err = apibackend.NoErr

	err := x.business.HandleMsg(req, res)
	if err != nil {
		l4g.Error("err: %s", err.Error())
	}
	if res.Err != apibackend.NoErr {
		l4g.Error("res err: %d-%s", res.Err, res.ErrMsg)
	}

	l4g.Info("res message: %s", res.Value.Message)
}
