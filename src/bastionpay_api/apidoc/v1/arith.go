package v1

import (
	"bastionpay_api/apidoc"
	"api_router/base/data"
	"bastionpay_api/api/v1"
)

var ApiDocAdd = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"arith",
	FuncName:"add",
	Level:data.APILevel_client,
	Comment:"加法功能",
	Input:&v1.Args{},
	Output:&v1.AckArgs{},
}