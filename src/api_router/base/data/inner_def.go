package data

import (
	"fmt"
	"bastionpay_api/apibackend"
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
	// normal client
	UserClass_Client = 0

	// hot
	UserClass_Hot = 1

	// admin
	UserClass_Admin = 2

	// client
	APILevel_client = 0

	// common administrator
	APILevel_admin = 100

	// genesis administrator
	APILevel_genesis = 200
)

// API info
type ApiInfo struct{
	Name 	string 	`json:"name"`    	// api name
	Level 	int		`json:"level"`		// api level, refer APILevel_*
}

// srv register data
type SrvRegisterData struct {
	Version      string `json:"version"`    // srv version
	Srv          string `json:"srv"`		// srv name
	Functions []ApiInfo `json:"functions"`  // srv functions
}

// srv context
type SrvContext struct{
	ApiLever int `json:"apilevel"`	// api info level
	DataFrom int `json:"datafrom"`	// data from router
	// future...
}

// srv data
type SrvData struct {
	// user unique key
	UserKey string `json:"user_key"`
	// sub user key
	SubUserKey string `json:"sub_user_key"`
	// user request message
	Message string `json:"message"`
	// signature = origin data -> sha512 -> rsa sign -> base64
	Signature  string `json:"signature"`
}

// input/output method
type SrvMethod struct {
	Version     string `json:"version"`   // srv version
	Srv     	string `json:"srv"`	  	  // srv name
	Function  	string `json:"function"`  // srv function
}

// srv request
type SrvRequest struct{
	Context 	SrvContext 	`json:"context"`	// api info
	Method		SrvMethod 	`json:"method"`		// request method
	Argv 		SrvData 	`json:"argv"` 		// request argument
}

// srv response/push
type SrvResponse struct{
	Err     	int    		`json:"err"`    // error code
	ErrMsg  	string 		`json:"errmsg"` // error message
	Value   	SrvData 	`json:"value"` 	// response data
}

//////////////////////////////////////////////////////////////////////
func (sr *SrvRequest)GetAccessUserKey() (string) {
	if sr.Context.DataFrom == apibackend.DataFromUser {
		return sr.Argv.SubUserKey
	} else if sr.Context.DataFrom == apibackend.DataFromAdmin {
		return sr.Argv.SubUserKey
	} else {
		return sr.Argv.UserKey
	}
}

func (urd SrvRequest)String() string {
	return fmt.Sprintf("%s %s-%s-%s", urd.Argv.UserKey, urd.Method.Srv, urd.Method.Version, urd.Method.Function)
}