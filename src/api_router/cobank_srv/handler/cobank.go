package handler

import (
	"api_router/base/data"
	//"api_router/base/service"
	service "api_router/base/service2"
	"business_center/business"
	l4g "github.com/alecthomas/log4go"
	"bastionpay_api/api"
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
	pData := data.UserRequestData{}
	pData.Method.Version = "v1"
	pData.Method.Srv = "push"
	pData.Method.Function = "pushdata"

	pData.Argv.UserKey = userID
	pData.Argv.Message = callbackMsg

	res := api.UserResponseData{}
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
			"query_user_address", data.APILevel_client, x.handler)
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
			"history_transaction_order", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"history_transaction_message", data.APILevel_client, x.handler)
	}()

	func() {
		service.RegisterApi(&nam,
			"set_pay_address", data.APILevel_client, x.handler)
	}()

	////////////////////////////////////////////////////////////////
	// 以下方法liuheng添加测试
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

	return nam
}

func (x *Cobank) handler(req *data.SrvRequestData, res *data.SrvResponseData) {
	res.Data.Err = data.NoErr

	l4g.Debug("argv: %s", req.Data.Argv)

	err := x.business.HandleMsg(req, res)
	if err != nil {
		l4g.Error("err: ", err)
	}

	l4g.Debug("value: %s", res.Data.Value)
}
