package apigroup

import (
	"bastionpay_api/apibackend"
	"bastionpay_api/apidoc"
	"bastionpay_api/apidoc/v1"
	"bastionpay_api/apidoc/v1/backend"
	"fmt"
)

var (
	apiDocGroupInfo     map[string]*apidoc.ApiGroupInfo
	apiDocGroupHandlers map[string][]apidoc.ApiDocHandler
)

func init() {
	apiDocGroupInfo = make(map[string]*apidoc.ApiGroupInfo)
	apiDocGroupHandlers = make(map[string][]apidoc.ApiDocHandler)

	// api group
	apiDocGroupInfo[apibackend.HttpRouterApi] = &apidoc.ApiGroupInfo{
		Description: "{{.ApiDocDescription}}",
	}

	// user group
	apiDocGroupInfo[apibackend.HttpRouterUser] = &apidoc.ApiGroupInfo{
		Description: "{{.UserApiDocDescription}}",
	}

	// admin group
	apiDocGroupInfo[apibackend.HttpRouterAdmin] = &apidoc.ApiGroupInfo{
		Description: "{{.AdminApiDocDescription}}",
	}

	// gateway
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocListSrv})

	// account
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocRegister})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocUpdateProfile})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocReadProfile})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocListUsers})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocUpdateFrozen})

	// auth

	// bastionpay
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocSupportAssets})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocAssetAttribute})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocNewAddress})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocWithdrawal})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocGetBalance})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocQueryAddress})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocTransactionBill})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocTransactionBillDaily})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocTransactionMessage})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocBlockHeight})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&v1.ApiDocDataPush})

	// backend
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocSpPostTransaction})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocSpReqAssetsAttributeList})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocSetPayAddress})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocSetAssetAttribute})
	RegisterApiDocHandler(&apidoc.ApiDocHandler{&backend.ApiDocSpGetbalance})

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

	if apiDocInfo.Tag == "" {
		apiDocInfo.Tag = "{{."
		apiDocInfo.Tag += apiDocInfo.SrvName
		apiDocInfo.Tag += "Tag"
		apiDocInfo.Tag += "}}"
	}
	if apiDocInfo.Name == "" {
		apiDocInfo.Name = "{{."
		apiDocInfo.Name += apiDocInfo.FuncName
		apiDocInfo.Name += "Name"
		apiDocInfo.Name += "}}"
	}
	if apiDocInfo.Description == "" {
		apiDocInfo.Description = "{{."
		apiDocInfo.Description += apiDocInfo.FuncName
		apiDocInfo.Description += "Description"
		apiDocInfo.Description += "}}"
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
