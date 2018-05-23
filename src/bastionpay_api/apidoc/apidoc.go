package apidoc

import (
	"bastionpay_api/api"
	"bastionpay_api/gateway"
)

type (
	ApiGroupInfo struct {
		Description string
	}

	ApiDoc struct{
		Group 			[]string
		Name  			string
		Description 	string
		VerName 		string
		SrvName 		string
		FuncName 		string
		Input 			interface{}
		Output 			interface{}
	}

	ApiDocHandler struct{
		ApiDocInfo *ApiDoc
	}
)

func (ad *ApiDoc)Path() string {
	return "/" + ad.VerName + "/" + ad.SrvName + "/" + ad.FuncName
}

func (this *ApiDocHandler)Help() (*ApiDoc) {
	return this.ApiDocInfo
}

func (this *ApiDocHandler)RunApi(req interface{}, ack interface{}) (*api.Error) {
	apiErr := gateway.RunApi(this.ApiDocInfo.Path(), req, ack)
	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (this *ApiDocHandler)RunUser(subUserKey string, req interface{}, ack interface{}) (*api.Error) {
	apiErr := gateway.RunUser(this.ApiDocInfo.Path(), subUserKey, req, ack)
	if apiErr != nil {
		return apiErr
	}

	return nil
}