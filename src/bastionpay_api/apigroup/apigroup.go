package apigroup

import (
	"bastionpay_api/apidoc/v1"
	"bastionpay_api/apidoc"
	"fmt"
	"encoding/json"
	"reflect"
	"bastionpay_api/utils"
)

var (
	apiDocGroupHandlers map[string][]apidoc.ApiDocHandler
)

func init()  {
	apiDocGroupHandlers = make(map[string][]apidoc.ApiDocHandler)

	// gateway
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocListSrv})

	// account
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocRegister})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocUpdateProfile})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocReadProfile})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocListUsers})

	// auth

	// bastionpay
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocSupportAssets})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocAssetAttribute})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocNewAddress})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocQueryUserAddress})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocWithdrawal})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocGetBalance})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocHistoryTransactionOrder})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocHistoryTransactionMessage})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocSetPayAddress})

	// bastionpay_tool
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocRecharge})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocGenerate})

	// push
}

func RegisterApiDocHandler(apiProxy *apidoc.ApiDocHandler) error {
	apiDocInfo := apiProxy.ApiDocInfo
	_, err := FindApiBySrvFunction(apiDocInfo.VerName, apiDocInfo.SrvName, apiDocInfo.FuncName)
	if err == nil {
		return fmt.Errorf("%s.%s.%s exist!", apiDocInfo.VerName, apiDocInfo.SrvName, apiDocInfo.FuncName)
	}

	if(reflect.TypeOf(apiProxy.ApiDocInfo.Input) == nil){
		apiProxy.ApiDocInfo.Example = ""
	} else if reflect.TypeOf(apiProxy.ApiDocInfo.Input).Kind() == reflect.String {
		apiProxy.ApiDocInfo.Example = apiProxy.ApiDocInfo.Input.(string)
	}else{
		example, _ := json.Marshal(apiProxy.ApiDocInfo.Input)
		apiProxy.ApiDocInfo.Example = string(example)
	}

	apiProxy.ApiDocInfo.InputComment = utils.FieldTag(apiProxy.Help().Input, 0)
	apiProxy.ApiDocInfo.OutputComment = utils.FieldTag(apiProxy.Help().Output, 0)

	apiGroup := apiDocGroupHandlers[apiDocInfo.VerName + "." + apiDocInfo.SrvName]
	apiGroup = append(apiGroup, *apiProxy)
	apiDocGroupHandlers[apiDocInfo.VerName + "." + apiDocInfo.SrvName] = apiGroup

	return nil
}

func ListApiGroup() (map[string][]apidoc.ApiDocHandler) {
	return apiDocGroupHandlers
}

func ListApiGroupBySrv(ver string, srv string) ([]apidoc.ApiDocHandler, error) {
	if apiGroup, ok := apiDocGroupHandlers[ver + "." + srv]; ok {
		return apiGroup, nil
	}

	return nil, fmt.Errorf("% not exist!", srv)
}

func FindApiBySrvFunction(ver string, srv string, function string) (*apidoc.ApiDocHandler, error) {
	apiGroup := apiDocGroupHandlers[ver + "." + srv]
	if len(apiGroup) == 0 {
		return nil, fmt.Errorf("%s not exist!", srv)
	}

	for _, apiProxy := range apiGroup {
		if apiProxy.ApiDocInfo.FuncName == function {
			return &apiProxy, nil
		}
	}

	return nil, fmt.Errorf("%.%s not exist!", srv, function)
}
