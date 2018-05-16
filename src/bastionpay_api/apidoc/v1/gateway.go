package v1

import (
	"bastionpay_api/apidoc"
	"api_router/base/data"
	"bastionpay_api/api/v1"
)

var ApiDocListSrv = apidoc.ApiDoc{
	Name:"获取服务列表",
	Description:"查看当前服务，管理员权限api",
	Level:data.APILevel_admin,
	VerName:"v1",
	SrvName:"gateway",
	FuncName:"listsrv",
	Input:nil,
	Output:&v1.SrvRegisterData{Functions:[]v1.ApiInfo{v1.ApiInfo{}}},
}