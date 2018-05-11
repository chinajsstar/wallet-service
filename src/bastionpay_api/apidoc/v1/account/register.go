package account

import (
	"bastionpay_api/gateway"
	"bastionpay_api/api"
	"bastionpay_api/api/v1/account"
	"bastionpay_api/apidoc"
)

var (
	doc = apidoc.ApiDoc{
		Name:"register a user",
		Comment:"register a user",
		Path:"/api/v1/account/register",
		Input:account.Register{},
		Output:account.AckRegister{},
	}
)

type ApiRegister struct{

}

func ( *ApiRegister)Help() (*apidoc.ApiDoc) {
	return &doc
}

func ( *ApiRegister)Run(req interface{}, ack interface{}) (*api.Error) {
	apiErr := gateway.Run(doc.Path, req, ack)
	if apiErr != nil {
		return apiErr
	}

	return nil
}

func ( *ApiRegister)Output(message string, out *string) (*api.Error) {
	resByte, apiErr := gateway.Output(doc.Path, []byte(message))
	if apiErr != nil {
		return apiErr
	}

	*out = string(resByte)
	return nil
}