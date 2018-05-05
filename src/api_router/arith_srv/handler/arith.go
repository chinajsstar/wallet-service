package handler

import (
	"api_router/base/data"
	//"api_router/base/service"
	service "api_router/base/service2"
	"encoding/json"
	l4g "github.com/alecthomas/log4go"
)

type Args struct {
	A int `json:"a" comment:"加数1"`
	B int `json:"b" comment:"加数2"`
}
type AckArgs struct {
	C int `json:"c" comment:"和"`
}
type Arith int

func (arith *Arith)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	func(){
		input := Args{}
		output := AckArgs{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"add", data.APILevel_client, arith.Add,
			"加法运算", string(b), input, output)
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

	out := AckArgs{C:din.A+din.B}
	b, _ := json.Marshal(out)
	res.Data.Value.Message = string(b)
}