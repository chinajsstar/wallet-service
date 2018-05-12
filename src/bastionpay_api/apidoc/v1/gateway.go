package v1

import (
	"bastionpay_api/apidoc"
	"api_router/base/data"
	"bastionpay_api/api/v1"
)

var ApiDocListSrv = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"gateway",
	FuncName:"listsrv",
	Level:data.APILevel_admin,
	Comment:"查看当前服务",
	Path:"/api/v1/gateway/listsrv",
	Input:nil,
	Output:v1.SrvRegisterData{Functions:[]v1.ApiInfo{v1.ApiInfo{}}},
}