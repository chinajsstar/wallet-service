package data

// /////////////////////////////////////////////////////
// internal api gateway and service RPC data define
// /////////////////////////////////////////////////////

const(
	MethodCenterRegister   = "ServiceCenter.Register"	// srv node register to center
	MethodCenterUnRegister = "ServiceCenter.UnRegister"	// srv node unregister to center

	MethodCenterPush   	   = "ServiceCenter.Push"		// srv node push data to center
	MethodCenterDispatch   = "ServiceCenter.Dispatch"	// srv node dispatch a request to center
	MethodNodeCall         = "ServiceNode.Call"			// center call a srv node function
)

const(
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

// register data
type SrvRegisterData struct {
	Version      string `json:"version"`    // srv version
	Srv          string `json:"srv"`		// srv name
	Addr         string `json:"addr"`		// srv ip address
	Functions []ApiInfo `json:"functions"`  // srv functions
}

// srv context
type SrvContext struct{
	Api ApiInfo `json:"api"`	// api info
	// future...
}

// rpc srv request data
type SrvRequestData struct{
	Context SrvContext 		`json:"context"`	// api info
	Data 	UserRequestData `json:"data"`		// user request data
}

// rpc srv response data
type SrvResponseData struct{
	Data 	UserResponseData `json:"data"`		// user response data
}