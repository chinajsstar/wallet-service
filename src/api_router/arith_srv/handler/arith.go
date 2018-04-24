package handler

import (
	"api_router/base/data"
	"api_router/base/service"
	"encoding/json"
	"strconv"
	l4g "github.com/alecthomas/log4go"
	"business_center/def"
)

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}
type Arith int

func (arith *Arith)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	func(){
		apiInfo := data.ApiInfo{Name:"add", Level:data.APILevel_client}
		example := Args{}
		b, _ := json.Marshal(example)
		apiInfo.Example = string(b)
		nam[apiInfo.Name] = service.NodeApi{ApiHandler:arith.Add, ApiInfo:apiInfo}
	}()

	return nam
}

func (arith *Arith)Add(req *data.SrvRequestData, res *data.SrvResponseData){
	res.Data.Err = data.NoErr

	// from req
	din := Args{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &din)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Data.Err = data.ErrDataCorrupted
		return
	}

	res.Data.Value.Message = strconv.Itoa(din.A+din.B)
}