package v1

import (
	"bastionpay_api/apidoc"
	"api_router/base/data"
	"bastionpay_api/api/v1"
)

var ApiDocAdd = apidoc.ApiDoc{
	Name:"加法",
	Description:"加法功能",
	Level:data.APILevel_client,
	VerName:"v1",
	SrvName:"arith",
	FuncName:"add",
	Input:&v1.Args{},
	Output:&v1.AckArgs{},
}