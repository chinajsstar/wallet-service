package handler

import (
	"../../data"
	"../../base/service"
	"errors"
	"fmt"
	"encoding/json"
	"strconv"
)

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}
type Arith int

func (arith *Arith)RegisterApi(apis *[]data.ApiInfo, apisfunc *map[string]service.CallNodeApi) error  {
	regapi := func(name string, caller service.CallNodeApi, level int) error {
		if (*apisfunc)[name] != nil {
			fmt.Println("#error: api is already exist...", name)
			return errors.New("api is already exist...")
		}

		*apis = append(*apis, data.ApiInfo{name, level})
		(*apisfunc)[name] = service.CallNodeApi(caller)
		return nil
	}

	if err := regapi("add", arith.Add, data.APILevel_client); err != nil {
		return err
	}

	return nil
}

func (arith *Arith)Add(req *data.SrvRequestData, res *data.SrvResponseData){
	res.Data.Err = data.NoErr

	// from req
	din := Args{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &din)
	if err != nil {
		res.Data.Err = data.ErrSrvInternalErr
		res.Data.ErrMsg = data.ErrSrvInternalErrText
		return
	}

	res.Data.Value.Message = strconv.Itoa(din.A+din.B)
	res.Data.Value.Signature = ""
	res.Data.Value.LicenseKey = req.Data.Argv.LicenseKey
}