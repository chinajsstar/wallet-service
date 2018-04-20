package handler

import (
	"api_router/base/data"
	"api_router/base/service"
	"business_center/business"
	l4g "github.com/alecthomas/log4go"
	"business_center/def"
	"encoding/json"
)

type Cobank struct{
	business *business.Business

	node *service.ServiceNode
}

func NewCobank() (*Cobank) {
	x := &Cobank{}
	x.business = &business.Business{}
	return x
}

func (x *Cobank)Start(node *service.ServiceNode) error {
	x.node = node
	return x.business.InitAndStart(x.callBack)
}

func (x *Cobank)Stop() {
	x.business.Stop()
}

func (x *Cobank)callBack(userID string, callbackMsg string){
	pData := data.UserRequestData{}
	pData.Method.Version = "v1"
	pData.Method.Srv = "push"
	pData.Method.Function = "pushdata"

	pData.Argv.UserKey = userID
	pData.Argv.Message = callbackMsg

	res := data.UserResponseData{}
	x.node.InnerCallByEncrypt(&pData, &res)
	l4g.Info("push return: ", res)
}

func (x *Cobank)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	apiInfo := data.ApiInfo{Name:"new_address", Level:data.APILevel_client}
	reqNewAddress := def.ReqWithdrawal{}
	b1, _ := json.Marshal(reqNewAddress)
	apiInfo.Example = string(b1)
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:x.handler, ApiInfo:apiInfo}

	//"user_key": "string",
	//"user_name": "string",
	//"user_class": "int",
	//"asset_name": "string",
	//"address": "string",
	//"max_amount": "double",
	//"min_amount": "double",
	//"create_time_begin": "int64",
	//"create_time_end": "int64",
	//"page_index": "int",
	//"max_display": "int"

	apiInfo = data.ApiInfo{Name:"query_user_address", Level:data.APILevel_admin}
	apiInfo.Example = "{\"user_key\":\"\"}"
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:x.handler, ApiInfo:apiInfo}

	apiInfo = data.ApiInfo{Name:"withdrawal", Level:data.APILevel_client}
	reqWithDrawal := def.ReqWithdrawal{}
	b2, _ := json.Marshal(reqWithDrawal)
	apiInfo.Example = string(b2)
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:x.handler, ApiInfo:apiInfo}

	return nam
}

func (x *Cobank)handler(req *data.SrvRequestData, res *data.SrvResponseData){
	res.Data.Err = data.NoErr

	l4g.Debug("argv: %s", req.Data.Argv)

	err := x.business.HandleMsg(req, res)
	if err != nil {
		l4g.Error("err: ", err)
	}

	l4g.Debug("value: %s", res.Data.Value)
}