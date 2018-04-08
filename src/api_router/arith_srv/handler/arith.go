package handler

import (
	"../../data"
	"../../base/service"
	"encoding/json"
	"strconv"
)

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}
type Arith int

func (arith *Arith)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	apiInfo := data.ApiInfo{Name:"add", Level:data.APILevel_client}
	apiInfo.Example = "{\"a\":1,\"b\":1}"
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:arith.Add, ApiInfo:apiInfo}

	return nam
}

func (arith *Arith)Add(req *data.SrvRequestData, res *data.SrvResponseData){
	res.Data.Err = data.NoErr

	// from req
	din := Args{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &din)
	if err != nil {
		res.Data.Err = data.ErrDataCorrupted
		res.Data.ErrMsg = data.ErrDataCorruptedText
		return
	}

	res.Data.Value.Message = strconv.Itoa(din.A+din.B)
	res.Data.Value.Signature = ""
	res.Data.Value.LicenseKey = req.Data.Argv.LicenseKey
}