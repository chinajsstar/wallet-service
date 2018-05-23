package apidoc

import (
	"bastionpay_api/api"
)

var ApiDocDataEntry = ApiDoc{
	Group:[]string{},
	Name:"统一数据接口",
	Description:"所有API的输入输出数据结构必须先序列化成字符串，经过加密和签名之后，将加密，签名和UserKey填入此数据结构，加密签名使用RSA2048位非对称加密",
	VerName:"version",
	SrvName:"service",
	FuncName:"function",
	Input:&api.UserData{},
	Output:&api.UserResponseData{},
}