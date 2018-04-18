package handler

import (
	"api_router/base/data"
	"api_router/base/service"
	"business_center/business"
	l4g "github.com/alecthomas/log4go"
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

func (x *Cobank)Init(node *service.ServiceNode) error {
	x.node = node
	return x.business.InitAndStart(x.callBack)
}

func (x *Cobank)callBack(userID string, callbackMsg string){
	pData := data.UserRequestData{}
	pData.Method.Version = "v1"
	pData.Method.Srv = "push"
	pData.Method.Function = "pushdata"

	pData.Argv.UserKey = userID
	pData.Argv.Message = callbackMsg

	res := data.UserResponseData{}
	x.node.Push(&pData, &res)
}

func (x *Cobank)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	apiInfo := data.ApiInfo{Name:"new_address", Level:data.APILevel_client}
	apiInfo.Example = "{\"id\":\"1\",\"symbol\":\"eth\",\"count\":5}"
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:x.NewAddress, ApiInfo:apiInfo}

	return nam
}

func (x *Cobank)NewAddress(req *data.SrvRequestData, res *data.SrvResponseData){
	res.Data.Err = data.NoErr

	l4g.Debug("argv: %s", req.Data.Argv)

	err := x.business.HandleMsg(req, res)
	l4g.Debug("value: %s", res.Data.Value)

	if err != nil {
		return
	}

	res.Data.Value.Signature = ""
	res.Data.Value.UserKey = req.Data.Argv.UserKey
}