package backend

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/apibackend"
	"bastionpay_api/apibackend/v1/backend"
)

var ApiDocListSrv = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterAdmin},
	VerName:"v1",
	SrvName:"gateway",
	FuncName:"listsrv",
	Input:nil,
	Output:&backend.ServiceInfoList{[]backend.ServiceInfo{backend.ServiceInfo{}}},
}