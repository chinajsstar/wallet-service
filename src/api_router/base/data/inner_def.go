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
	MethodCenterInnerNotify  		= "ServiceCenter.InnerNotify"			// notify data to center to nodes
	MethodCenterInnerCall  			= "ServiceCenter.InnerCall"				// call a api to center
	MethodCenterInnerCallByEncrypt 	= "ServiceCenter.InnerCallByEncrypt"	// call a api by encrypt data to center

	MethodNodeCall         			= "ServiceNode.Call"					// center call a srv node function
	MethodNodeNotify         		= "ServiceNode.Notify"					// center notify to a srv node function
)

const(
	// client
	APILevel_client = 0

	// common administrator
	APILevel_admin = 100

	// genesis administrator
	APILevel_genesis = 200
)

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