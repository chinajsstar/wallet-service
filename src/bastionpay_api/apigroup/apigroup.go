package apigroup

import (
	"bastionpay_api/apibackend"
	"bastionpay_api/apidoc"
	"bastionpay_api/apidoc/v1"
	"fmt"
)

var (
	apiDocGroupInfo     map[string]*apidoc.ApiGroupInfo
	apiDocGroupHandlers map[string][]apidoc.ApiDocHandler
)

func init() {
	apiDocGroupInfo = make(map[string]*apidoc.ApiGroupInfo)
	apiDocGroupHandlers = make(map[string][]apidoc.ApiDocHandler)

	apiDocGroupInfo[apibackend.HttpRouterApi] = &apidoc.ApiGroupInfo{
		Description: `This api document is for developers to access BastionPay service, 
					all api request and response json body is not real body data,
					developers need convert json to string, and then package string to common json,
					you can go to github.com to download golang sdk.`,
	}

	apiDocGroupInfo[apibackend.HttpRouterUser] = &apidoc.ApiGroupInfo{
		Description: "",
	}
	apiDocGroupInfo[apibackend.HttpRouterAdmin] = &apidoc.ApiGroupInfo{
		Description: "",
	}

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
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocGetBalance})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocQueryUserAddress})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocHistoryTransactionBill})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocHistoryTransactionMessage})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocTransactionBillDaily})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocNewAddress})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocWithdrawal})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocSetPayAddress})

	// bastionpay_tool
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocRecharge})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocGenerate})

	// push
}

func GetApiGroupInfo(apiGroup string) (*apidoc.ApiGroupInfo, error) {
	if v1, ok := apiDocGroupInfo[apiGroup]; ok {
		return v1, nil
	}

	return nil, fmt.Errorf("Not find %s", apiGroup)
}

func RegisterApiDocHandler(apiProxy *apidoc.ApiDocHandler) error {
	apiDocInfo := apiProxy.ApiDocInfo
	_, err := FindApiBySrvFunction(apiDocInfo.VerName, apiDocInfo.SrvName, apiDocInfo.FuncName)
	if err == nil {
		return fmt.Errorf("%s.%s.%s exist!", apiDocInfo.VerName, apiDocInfo.SrvName, apiDocInfo.FuncName)
	}

	apiGroup := apiDocGroupHandlers[apiDocInfo.VerName+"."+apiDocInfo.SrvName]
	apiGroup = append(apiGroup, *apiProxy)
	apiDocGroupHandlers[apiDocInfo.VerName+"."+apiDocInfo.SrvName] = apiGroup

	return nil
}

func ListApiGroup() map[string][]apidoc.ApiDocHandler {
	return apiDocGroupHandlers
}

func ListApiGroupBySrv(ver string, srv string) ([]apidoc.ApiDocHandler, error) {
	if apiGroup, ok := apiDocGroupHandlers[ver+"."+srv]; ok {
		return apiGroup, nil
	}

	return nil, fmt.Errorf("% not exist!", srv)
}

func FindApiBySrvFunction(ver string, srv string, function string) (*apidoc.ApiDocHandler, error) {
	apiGroup := apiDocGroupHandlers[ver+"."+srv]
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
