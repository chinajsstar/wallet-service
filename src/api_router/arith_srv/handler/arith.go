package handler

import (
	"api_router/base/data"
	//"api_router/base/service"
	service "api_router/base/service2"
	"encoding/json"
	l4g "github.com/alecthomas/log4go"
	"bastionpay_api/api/v1"
)

type Arith int

func (arith *Arith)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	func(){
		service.RegisterApi(&nam,
			"add", data.APILevel_client, arith.Add)
	}()

	return nam
}

func (arith *Arith)HandleNotify(req *data.SrvRequest){
	l4g.Info("HandleNotify-reloadUserLevel: do nothing")
}

func (arith *Arith)Add(req *data.SrvRequest, res *data.SrvResponse){
	res.Err = data.NoErr

	// from req
	din := v1.Args{}
	err := json.Unmarshal([]byte(req.Argv.Message), &din)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Err = data.ErrDataCorrupted
		return
	}

	out := v1.AckArgs{C:din.A+din.B}
	b, _ := json.Marshal(out)
	res.Value.Message = string(b)
}