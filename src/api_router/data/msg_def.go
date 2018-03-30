package data

// 对外结构
// 用户输入输出信息，作为请求和应答的实际信息
// 用户数据
type UserData struct {
	LicenseKey string `json:"license_key"` 	// 用户标示
	Message    string `json:"message"`		// 数据
	Signature  string `json:"signature"`	// 签名
}

// 请求信息，作为请求数据
// json like: {"version":"v1", "srv":"arith", "function":"add", "argv":{...}}
type ServiceCenterDispatchData struct{
	Version     string `json:"version"`   // 版本号
	Srv     	string `json:"srv"`	  	  // 服务名称
	Function  	string `json:"function"`  // 服务功能
	Argv 		UserData `json:"argv"` 	  // UserData
}

// 应答信息，作为应答数据
// json like: {"err":0, "errmsg":"", "value":{...}}
type ServiceCenterDispatchAckData struct{
	Err     int    `json:"err"`     // 错误码
	ErrMsg  string `json:"errmsg"`  // 错误信息
	Value   UserData `json:"value"` // UserData
}