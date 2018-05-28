package backend

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/apibackend"
	"bastionpay_api/apibackend/v1/backend"
)

var ApiDocListSrv = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterAdmin},
	Name:"获取服务列表",
	Description:"查看当前服务，管理员权限api",
	VerName:"v1",
	SrvName:"gateway",
	FuncName:"listsrv",
	Input:nil,
	Output:&backend.ServiceInfoList{[]backend.ServiceInfo{backend.ServiceInfo{}}},
}