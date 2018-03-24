package data

import (
	"reflect"
	"strings"
)

const(
	MethodServiceCenterRegister   = "ServiceCenter.Register"	// 服务向服务中心注册请求，对内
	MethodServiceCenterUnRegister = "ServiceCenter.UnRegister"	// 服务向服务中心反注册请求，对内
	MethodServiceCenterDispatch   = "ServiceCenter.Dispatch"	// 客户向服务中心发送请求，对外
	MethodServiceNodePingpong     = "ServiceNode.Pingpong"		// 服务中心向服务发送心跳，对内
	MethodServiceNodeCall         = "ServiceNode.Call"			// 服务中心向服务发送请求，对内
)

// 注册信息
type ServiceCenterRegisterData struct {
	Srv          string `json:"srv"`		// service arith_srv name=version+srvname
	Addr         string `json:"addr"`		// service arith_srv ip address
	Functions  []string `json:"functions"`  // service arith_srv functions
}

// 注册API
func (rd *ServiceCenterRegisterData)RegisterFunction(handler interface{})  {
	t := reflect.TypeOf(handler)
	//v := reflect.ValueOf(api)

	//tName := reflect.Indirect(v).Type().Name()
	for m := 0; m < t.NumMethod(); m++ {
		method := t.Method(m)
		mName := method.Name

		//rd.Functions = append(rd.Functions, tName+"."+mName)
		rd.Functions = append(rd.Functions, strings.ToLower(mName))
	}
}

func (rd *ServiceCenterRegisterData)GetVersionSrvName() string {
	return strings.ToLower(rd.Srv)
}

// 请求信息，作为rpc请求的params数据
// json like: {"command":"v1.Arith.Add", "argv":""}
type ServiceCenterDispatchData struct{
	Srv     	string `json:"srv"`	  	  // like "v1.arith"
	Function  	string `json:"function"`  // like "add"
	Argv 		string `json:"argv"` 	  // json string
}

func (sd *ServiceCenterDispatchData)GetVersionSrvName() string {
	return strings.ToLower(sd.Srv)
}

// 应答信息，作为rpc应答的result数据
// json like: {"err":0, "errmsg":"", "value":""}
type ServiceCenterDispatchAckData struct{
	Err     int    `json:"err"`     // like 0
	ErrMsg  string `json:"errmsg"`  // string
	Value   string `json:"value"`   // json string
}