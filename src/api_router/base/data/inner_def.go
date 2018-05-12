package data

import (
	"fmt"
	"bastionpay_api/api"
)

// /////////////////////////////////////////////////////
// internal api gateway and service RPC data define
// /////////////////////////////////////////////////////

const(
	MethodCenterRegister   			= "ServiceCenter.Register"				// register to center
	MethodCenterUnRegister 			= "ServiceCenter.UnRegister"			// unregister to center
	MethodCenterInnerCall  			= "ServiceCenter.InnerCall"				// call a api to center
	MethodCenterInnerCallByEncrypt 	= "ServiceCenter.InnerCallByEncrypt"	// call a api by encrypt data to center

	MethodNodeCall         			= "ServiceNode.Call"					// center call a srv node function
)

const(
	// client
	APILevel_client = 0

	// common administrator
	APILevel_admin = 100

	// genesis administrator
	APILevel_genesis = 200
)

// API doc
type ApiDoc struct{
	Name 		string 	`json:"name"`    	// api name
	Level 		int		`json:"level"`		// api level, refer APILevel_*
	Doc 		string 	`json:"doc"`    	// api doc
	Example 	string  `json:"example"`	// api example string
	InComment 	string  `json:"incomment"`	// api input comment string
	OutComment 	string  `json:"outcomment"`	// api output comment string
}

// API info
type ApiInfo struct{
	Name 	string 	`json:"name"`    	// api name
	Level 	int		`json:"level"`		// api level, refer APILevel_*
}

// register data
type SrvRegisterData struct {
	Version      string `json:"version"`    // srv version
	Srv          string `json:"srv"`		// srv name
	Functions []ApiInfo `json:"functions"`  // srv functions
	ApiDocs    []ApiDoc `json:"apidocs"`  	// srv apidocs
}

// srv context
type SrvContext struct{
	ApiLever int `json:"apilevel"`	// api info level
	// future...
}

// user request data
type UserRequestData struct{
	Method		api.UserMethod 	`json:"method"`	// request method
	Argv 		api.UserData 	`json:"argv"` 	// request argument
}

// rpc srv request data
type SrvRequestData struct{
	Context SrvContext 		`json:"context"`	// api info
	Data 	UserRequestData `json:"data"`		// user request data
}

// rpc srv response data
type SrvResponseData struct{
	Data 	api.UserResponseData `json:"data"`		// user response data
}

//////////////////////////////////////////////////////////////////////
func (urd UserRequestData)String() string {
	return fmt.Sprintf("%s %s-%s-%s", urd.Argv.UserKey, urd.Method.Srv, urd.Method.Version, urd.Method.Function)
}

func (srd SrvRegisterData)String() string {
	return fmt.Sprintf("%s-%s", srd.Srv, srd.Version)
}