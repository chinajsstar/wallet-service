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
	Version      string `json:"version"`    // service version
	Srv          string `json:"srv"`		// service name
	Addr         string `json:"addr"`		// service ip address
	Functions  []string `json:"functions"`  // service functions
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

/////////////////////////////////////////////////////////////////////////
// 用户输入输出信息，作为请求和应答的实际信息
// 用户数据
type UserData struct {
	LicenseKey string `json:"license_key"` 	// 用户标示
	Message    string `json:"message"`		// 数据
	Signature  string `json:"signature"`	// 签名
}

// 请求信息，作为请求数据
// json like: {"version":"v1", "srv":"arith", "function":"add", "argv":""}
type ServiceCenterDispatchData struct{
	Version     string `json:"version"`   // 版本号
	Srv     	string `json:"srv"`	  	  // 服务名称
	Function  	string `json:"function"`  // 服务功能
	Argv 		string `json:"argv"` 	  // UserData string
}

// 应答信息，作为应答数据
// json like: {"err":0, "errmsg":"", "value":""}
type ServiceCenterDispatchAckData struct{
	Err     int    `json:"err"`     // 错误码
	ErrMsg  string `json:"errmsg"`  // 错误信息
	Value   string `json:"value"`   // UserData string
}