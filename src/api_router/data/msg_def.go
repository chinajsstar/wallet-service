package data

// API Gateway层定义用户请求和应答数据结构
//

// 用户输入输出功能号
type UserMethod struct {
	Version     string `json:"version"`   // 版本号
	Srv     	string `json:"srv"`	  	  // 服务名称
	Function  	string `json:"function"`  // 服务功能
}

// 用户输入输出信息，作为请求和应答的实际信息;
// 当作为用户请求时，需要用户对数据进行加密签名，服务器来解析验证
// 当作为用户应答时，服务器会对数据进行加密签名，用户来解析验证
type UserData struct {
	LicenseKey string `json:"license_key"` 	// 用户唯一标示
	Message    string `json:"message"`		// 消息：原始数据--对方公钥加密--base64编码
	Signature  string `json:"signature"`	// 签名：原始数据--sha512哈希--自己私钥加密--base64编码
}

// 用户请求数据
// json like: {"method":"", "argv":{...}}
type UserRequestData struct{
	Method		UserMethod 	`json:"method"`	// 功能
	Argv 		UserData 	`json:"argv"` 	// 参数
}

// 用户应答/推送信息
// json like: {"method":"", "err":0, "errmsg":"", "value":{...}}
type UserResponseData struct{
	Method		UserMethod 	`json:"method"`	// 功能
	Err     	int    		`json:"err"`    // 错误码
	ErrMsg  	string 		`json:"errmsg"` // 错误信息
	Value   	UserData 	`json:"value"` 	// 结果
}