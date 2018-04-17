package handler

import (
	"../../base/data"
	"../../base/service"
	"business_center/business"
	l4g "github.com/alecthomas/log4go"
)

type Cobank struct{
	business *business.Business
}

func NewCobank() (*Cobank) {
	x := &Cobank{}
	x.business = &business.Business{}
	return x
}

func (x *Cobank)Init() error {
	return x.business.InitAndStart()
}

func (x *Cobank)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	apiInfo := data.ApiInfo{Name:"new_address", Level:data.APILevel_client}
	apiInfo.Example = "{\"user_id\":\"0001\",\"method\":\"new_address\",\"params\":{\"id\":\"1\",\"symbol\":\"eth\",\"count\":5}}"
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:x.NewAddress, ApiInfo:apiInfo}

	return nam
}

func (x *Cobank)NewAddress(req *data.SrvRequestData, res *data.SrvResponseData){
	res.Data.Err = data.NoErr

	l4g.Debug("message: %s", req.Data.Argv.Message)

	var reply string
	err := x.business.HandleMsg(req.Data.Argv.Message, &reply)
	l4g.Debug("reply: %s", reply)

	if err != nil {
		res.Data.Err = data.ErrDataCorrupted
		res.Data.ErrMsg = data.ErrDataCorruptedText
		return
	}

	res.Data.Value.Message = reply
	res.Data.Value.Signature = ""
	res.Data.Value.LicenseKey = req.Data.Argv.LicenseKey
}