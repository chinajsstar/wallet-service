package common

import (
	"reflect"
	"strings"
)

const(
	MethodServiceCenterRegister = "ServiceCenter.Register"	// 服务向服务中心注册请求，对内
	MethodServiceCenterDispatch = "ServiceCenter.Dispatch"	// 客户向服务中心发送请求，对外
	MethodServiceCenterPingpong = "ServiceCenter.Pingpong"	// 服务向服务中心发送心跳，对内
	MethodServiceNodeCall       = "ServiceNode.Call"		// 服务中心向服务发送请求，对内
)

// 注册信息
type ServiceCenterRegisterData struct {
	Name    string `json:"name"`		// service node name
	Version string `json:"version"`		// service node version
	Addr    string `json:"addr"`		// service node ip address
	Apis  []string `json:"apis""`  		// service node apis
}

// 注册API
func (rd *ServiceCenterRegisterData)RegisterApi(api interface{})  {
	t := reflect.TypeOf(api)
	v := reflect.ValueOf(api)

	tName := reflect.Indirect(v).Type().Name()
	for m := 0; m < t.NumMethod(); m++ {
		method := t.Method(m)
		mName := method.Name

		//rd.Apis = append(rd.Apis, tName+"."+mName)
		rd.Apis = append(rd.Apis, strings.ToLower(tName) + "." + strings.ToLower(mName))
	}
}

func (rd *ServiceCenterRegisterData)GetVersionName() string {
	return strings.ToLower(rd.Version) + "." + strings.ToLower(rd.Name)
}

// 请求信息，作为rpc请求的params数据
// json like: {"version":"v1", "api":"Arith.Add", "argv":""}
type ServiceCenterDispatchData struct{
	Version string `json:"version"` // like "v1"
	Api  	string `json:"api"`  	// like "xxx.xxx"
	Argv 	string `json:"argv"` 	// json string
}

func (sd *ServiceCenterDispatchData)GetVersionApi() string {
	return strings.ToLower(sd.Version) + "." + strings.ToLower(sd.Api)
}

// 应答信息，作为rpc应答的result数据
// json like: {"version":"v1", "api":"Arith.Add", "err":0, "errmsg":"", "value":""}
type ServiceCenterDispatchAckData struct{
	Version string `json:"version"` // like "v1"
	Api     string `json:"api"`     // like "xxx.xxx"
	Err     int    `json:"err"`     // like 100
	ErrMsg  string `json:"errmsg"`  // string
	Value   string `json:"value"`   // json string
}