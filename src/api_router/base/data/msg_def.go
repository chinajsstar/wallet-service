package data

// /////////////////////////////////////////////////////
// API Gateway layer data define for users
// /////////////////////////////////////////////////////

// input/output method
type UserMethod struct {
	Version     string `json:"version"`   // srv version
	Srv     	string `json:"srv"`	  	  // srv name
	Function  	string `json:"function"`  // srv function
}

// input/output data/value
// when input data, user encode and sign data, server decode and verify;
// when output value, server encode and sign data, user decode and verify;
type UserData struct {
	// user unique key
	UserKey string `json:"user_key"`
	// message = origin data -> rsa encode -> base64
	Message    string `json:"message"`
	// signature = origin data -> sha512 -> rsa sign -> base64
	Signature  string `json:"signature"`
}

// user request data
type UserRequestData struct{
	Method		UserMethod 	`json:"method"`	// request method
	Argv 		UserData 	`json:"argv"` 	// request argument
}

// user response/push data
type UserResponseData struct{
	Method		UserMethod 	`json:"method"`	// response method
	Err     	int    		`json:"err"`    // error code
	ErrMsg  	string 		`json:"errmsg"` // error message
	Value   	UserData 	`json:"value"` 	// response data
}