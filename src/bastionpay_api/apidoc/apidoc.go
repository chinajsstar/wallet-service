package apidoc

import (
	"bastionpay_api/api"
	"bastionpay_api/gateway"
)

type (
	ApiDoc struct{
		Name 	string
		Comment string
		Path 	string
		Input 	interface{}
		Output 	interface{}
	}

	ApiDocHandler struct{
		ApiDocInfo *ApiDoc
	}
)

func (this *ApiDocHandler)Help() (*ApiDoc) {
	return this.ApiDocInfo
}

func (this *ApiDocHandler)Run(req interface{}, ack interface{}) (*api.Error) {
	apiErr := gateway.Run(this.ApiDocInfo.Path, req, ack)
	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (this *ApiDocHandler)Output(message string, out *string) (*api.Error) {
	resByte, apiErr := gateway.Output(this.ApiDocInfo.Path, []byte(message))
	if apiErr != nil {
		return apiErr
	}

	*out = string(resByte)
	return nil
}