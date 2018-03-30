package handler

import (
	"../../data"
	"../../base/service"
	"errors"
)

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}
type Arith int

func (arith *Arith)RegisterApi(apis *[]data.ApiInfo, apisfunc *map[string]service.CallNodeApi) error  {
	regapi := func(name string, caller service.CallNodeApi, level int) error {
		if (*apisfunc)[name] != nil {
			return errors.New("api is already exist...")
		}

		*apis = append(*apis, data.ApiInfo{name, level})
		(*apisfunc)[name] = service.CallNodeApi(caller)
		return nil
	}

	if err := regapi("add", arith.Add, data.APILevel_boss); err != nil {
		return err
	}

	return nil
}

func (arith *Arith)Add(req *data.SrvDispatchData, ack *data.SrvDispatchAckData) error{
	return nil
}