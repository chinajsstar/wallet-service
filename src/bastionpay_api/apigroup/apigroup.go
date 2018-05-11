package apigroup

import (
	"errors"
	"bastionpay_api/apidoc/v1/account"
	"bastionpay_api/apidoc"
)

var (
	apiProxys map[string]apidoc.ApiProxy
)

func init()  {
	apiProxys = make(map[string]apidoc.ApiProxy)

	RegisterApiProxy("account.register", new(account.ApiRegister))
}

func ListAll() (map[string]apidoc.ApiProxy) {
	return apiProxys
}

func RegisterApiProxy(name string, apiProxy apidoc.ApiProxy) error {
	if _, ok := apiProxys[name]; ok {
		return errors.New("repeat api name")
	}
	apiProxys[name] = apiProxy
	return nil
}

func FindApiProxy(name string) (apidoc.ApiProxy, error) {
	if apiHanlder, ok := apiProxys[name]; ok {
		return apiHanlder, nil
	}
	return nil, errors.New("not find api handler")
}
