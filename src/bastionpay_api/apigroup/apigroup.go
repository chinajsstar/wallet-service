package apigroup

import (
	"errors"
	"bastionpay_api/apidoc/v1"
	"bastionpay_api/apidoc"
)

var (
	apiDocHandlers map[string]*apidoc.ApiDocHandler
)

func init()  {
	apiDocHandlers = make(map[string]*apidoc.ApiDocHandler)

	RegisterApiDocHandler("account.register", &apidoc.ApiDocHandler{&v1.ApiDocRegister})
	RegisterApiDocHandler("account.updateprofile", &apidoc.ApiDocHandler{&v1.ApiDocUpdateProfile})

	RegisterApiDocHandler("bastionpay.support_assets", &apidoc.ApiDocHandler{&v1.ApiDocSupportAssets})
	RegisterApiDocHandler("bastionpay.asset_attribute", &apidoc.ApiDocHandler{&v1.ApiDocAssetAttribute})
}

func ListAllApiDocHandlers() (map[string]*apidoc.ApiDocHandler) {
	return apiDocHandlers
}

func RegisterApiDocHandler(name string, apiProxy *apidoc.ApiDocHandler) error {
	if _, ok := apiDocHandlers[name]; ok {
		return errors.New("repeat api name")
	}
	apiDocHandlers[name] = apiProxy
	return nil
}

func FindApiDocHandler(name string) (*apidoc.ApiDocHandler, error) {
	if apiHanlder, ok := apiDocHandlers[name]; ok {
		return apiHanlder, nil
	}
	return nil, errors.New("not find api doc handler")
}
