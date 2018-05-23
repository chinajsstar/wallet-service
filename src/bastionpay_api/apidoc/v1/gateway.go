package v1

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/api/v1"
	"bastionpay_api/apibackend"
)

var ApiDocListSrv = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	Name:"获取服务列表",
	Description:"查看当前服务，管理员权限api",
	VerName:"v1",
	SrvName:"gateway",
	FuncName:"listsrv",
	Input:nil,
	Output:&v1.SrvRegisterData{Functions:[]v1.ApiInfo{v1.ApiInfo{}}},
}