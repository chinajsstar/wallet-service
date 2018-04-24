package data

// error code and message
var(
	err_msg map[int]string
)

func init()  {
	err_msg = make(map[int]string)
	err_msg[NoErr] 									= ""
	err_msg[ErrInternal] 							= "internal error"
	err_msg[ErrDataCorrupted] 						= "data corrupted"
	err_msg[ErrCallFailed] 							= "call failed"
	err_msg[ErrIllegallyCall] 						= "illegally call"
	err_msg[ErrNotFindAuth] 						= "not find auth service"
	err_msg[ErrNotFindSrv] 							= "not find service"
	err_msg[ErrNotFindFunction] 					= "not find function"
	err_msg[ErrConnectSrvFailed] 					= "connect service failed"

	err_msg[ErrAccountSrvUsernameRepeated] 			= "user name is repeated"
	err_msg[ErrAccountSrvPhoneRepeated] 			= "user phone is repeated"
	err_msg[ErrAccountSrvEmailRepeated] 			= "user email is repeated"

	err_msg[ErrAccountSrvNoUser] 					= "no user"
	err_msg[ErrAccountSrvWrongPassword] 			= "wrong password"

	err_msg[ErrAccountSrvUpdatePassword] 			= "update password failed"

	err_msg[ErrAccountSrvListUsers] 				= "list users failed"

	err_msg[ErrAuthSrvNoUserKey] 					= "no user key"
	err_msg[ErrAuthSrvNoPublicKey] 					= "no public key"
	err_msg[ErrAuthSrvNoApiLevel] 					= "no api level"
	err_msg[ErrAuthSrvUserFrozen] 					= "user key is frozen"
	err_msg[ErrAuthSrvIllegalData] 					= "illegal data"

	err_msg[ErrPushSrvPushData] 					= "push failed"
}

func GetErrMsg(errId int) string {
	if msg, ok := err_msg[errId]; ok{
		return msg
	}
	return "service internal error"
}

const(
	// /////////////////////////////////////////////////////
	// 0-1000 common errors
	// /////////////////////////////////////////////////////
	// no error
	NoErr 						= 0

	// internal err
	ErrInternal					= 1

	// data corrupted
	ErrDataCorrupted			= 2

	// call failed
	ErrCallFailed				= 3

	// illegally call
	ErrIllegallyCall			= 4

	// not find auth service
	ErrNotFindAuth				= 5

	// not find service
	ErrNotFindSrv				= 100

	// not find function
	ErrNotFindFunction			= 101

	// connect service failed
	ErrConnectSrvFailed			= 102

	// /////////////////////////////////////////////////////
	// 1001-2000 account_srv errors
	// /////////////////////////////////////////////////////
	// register - user name is repeated
	ErrAccountSrvUsernameRepeated		= 1001

	// register - user phone is repeated
	ErrAccountSrvPhoneRepeated			= 1002

	// register - user email is repeated
	ErrAccountSrvEmailRepeated			= 1003

	// login - no user
	ErrAccountSrvNoUser					= 1101

	// login - wrong password
	ErrAccountSrvWrongPassword			= 1102

	// updatepassword - failed
	ErrAccountSrvUpdatePassword			= 1201

	// listusers - failed
	ErrAccountSrvListUsers				= 1301

	// /////////////////////////////////////////////////////
	// 2001-3000 auth_srv errors
	// /////////////////////////////////////////////////////
	// no user key
	ErrAuthSrvNoUserKey				= 2001

	// no public key
	ErrAuthSrvNoPublicKey  			= 2002

	// no api level
	ErrAuthSrvNoApiLevel			= 2003

	// user frozen
	ErrAuthSrvUserFrozen			= 2004

	// illegal data
	ErrAuthSrvIllegalData 			= 2005

	// /////////////////////////////////////////////////////
	// 3001-4000 push_srv errors
	// /////////////////////////////////////////////////////
	// illegal data
	ErrPushSrvPushData 			= 3001
)
