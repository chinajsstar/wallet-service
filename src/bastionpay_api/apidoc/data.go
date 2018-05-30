package apidoc

import (
	"bastionpay_api/api"
)

var ApiDocDataEntry = ApiDoc{
	Group:[]string{},
	VerName:"version",
	SrvName:"service",
	FuncName:"function",
	Input:&api.UserData{},
	Output:&api.UserResponseData{},
}