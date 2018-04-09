package handler

import (
	"../../data"
	"../../base/service"
	"business_center/business"
	"fmt"
)

type Xxxx struct{
	business *business.Business
}

func NewXxxx() (*Xxxx) {
	x := &Xxxx{}
	x.business = &business.Business{}
	return x
}

func (x *Xxxx)Init() error {
	return x.business.InitAndStart()
}

func (x *Xxxx)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	apiInfo := data.ApiInfo{Name:"new_address", Level:data.APILevel_client}
	apiInfo.Example = "{\"user_id\":\"0001\",\"method\":\"new_address\",\"params\":{\"id\":\"1\",\"symbol\":\"eth\",\"count\":5}}"
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:x.NewAddress, ApiInfo:apiInfo}

	return nam
}

func (x *Xxxx)NewAddress(req *data.SrvRequestData, res *data.SrvResponseData){
	res.Data.Err = data.NoErr

	fmt.Println("message: ", req.Data.Argv.Message)

	var reply string
	err := x.business.HandleMsg(req.Data.Argv.Message, &reply)
	fmt.Println("reply: ", reply)

	if err != nil {
		res.Data.Err = data.ErrDataCorrupted
		res.Data.ErrMsg = data.ErrDataCorruptedText
		return
	}

	res.Data.Value.Message = reply
	res.Data.Value.Signature = ""
	res.Data.Value.LicenseKey = req.Data.Argv.LicenseKey
}