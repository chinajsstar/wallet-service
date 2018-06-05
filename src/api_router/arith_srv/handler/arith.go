package handler

import (
	"bastionpay_base/data"
	//"api_router/base/service"
	service "bastionpay_base/service2"
	"encoding/json"
	l4g "github.com/alecthomas/log4go"
	"bastionpay_api/api/v1"
	"bastionpay_api/apibackend"
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
	res.Err = apibackend.NoErr

	// from req
	din := v1.Args{}
	err := json.Unmarshal([]byte(req.Argv.Message), &din)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Err = apibackend.ErrDataCorrupted
		return
	}

	out := v1.AckArgs{C:din.A+din.B}
	b, _ := json.Marshal(out)
	res.Value.Message = string(b)
}