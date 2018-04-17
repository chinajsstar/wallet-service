package handler

import (
	"../../base/data"
	"../../base/service"
	"business_center/business"
	l4g "github.com/alecthomas/log4go"
	"business_center/def"
	"strconv"
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
	var cb def.PushMsgCallback
	cb = def.PushMsgCallback(x.callBack)
	return x.business.InitAndStart(&cb)
}

func (x *Cobank)callBack(userID string, callbackMsg string){
	pData := data.UserRequestData{}
	pData.Method.Version = "v1"
	pData.Method.Srv = "push"
	pData.Method.Function = "pushdata"

	pData.Argv.LicenseKey = userID
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

	l4g.Debug("message: %s", req.Data.Argv.Message)

	err := x.business.HandleMsg(req, res)
	l4g.Debug("reply: %s", res.Data.Value.Message)

	if err != nil {
		res.Data.Err = data.ErrDataCorrupted
		res.Data.ErrMsg = data.ErrDataCorruptedText
		return
	}

	res.Data.Value.Signature = ""
	res.Data.Value.LicenseKey = req.Data.Argv.LicenseKey
}