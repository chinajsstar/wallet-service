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

	err_msg[ErrAccountSrvNoUser] 					= "no user"
	err_msg[ErrAccountSrvUpdateProfile] 			= "update profile failed"
	err_msg[ErrAccountSrvListUsers] 				= "list users failed"

	err_msg[ErrAuthSrvNoUserKey] 					= "no user key"
	err_msg[ErrAuthSrvNoPublicKey] 					= "no public key"
	err_msg[ErrAuthSrvNoApiLevel] 					= "no api level"
	err_msg[ErrAuthSrvUserFrozen] 					= "user key is frozen"
	err_msg[ErrAuthSrvIllegalData] 					= "illegal data"
	err_msg[ErrAuthSrvIllegalDataType] 				= "illegal data type"

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
	// 0, success
	// /////////////////////////////////////////////////////
	// no error
	NoErr 							= 0

	// /////////////////////////////////////////////////////
	// 10001-11100 common errors
	// /////////////////////////////////////////////////////
	// internal err
	ErrInternal						= 10001

	// data corrupted
	ErrDataCorrupted				= 10002

	// call failed
	ErrCallFailed					= 10003

	// illegally call
	ErrIllegallyCall				= 10004

	// not find auth service
	ErrNotFindAuth					= 10005

	// not find service
	ErrNotFindSrv					= 10006

	// not find function
	ErrNotFindFunction				= 10007

	// connect service failed
	ErrConnectSrvFailed				= 10008

	// /////////////////////////////////////////////////////
	// 11101-11200 account_srv errors
	// /////////////////////////////////////////////////////
	// no user
	ErrAccountSrvNoUser				= 11101

	// updateprofile - failed
	ErrAccountSrvUpdateProfile		= 11102

	// listusers - failed
	ErrAccountSrvListUsers			= 11103

	// /////////////////////////////////////////////////////
	// 11201-11300 auth_srv errors
	// /////////////////////////////////////////////////////
	// no user key
	ErrAuthSrvNoUserKey				= 11201

	// no public key
	ErrAuthSrvNoPublicKey			= 11202

	// no api level
	ErrAuthSrvNoApiLevel			= 11203

	// user frozen
	ErrAuthSrvUserFrozen			= 11204

	// illegal data
	ErrAuthSrvIllegalData			= 11205

	// illegal data type
	ErrAuthSrvIllegalDataType		= 11206

	// /////////////////////////////////////////////////////
	// 11301-11400 push_srv errors
	// /////////////////////////////////////////////////////
	// illegal data
	ErrPushSrvPushData 				= 11301
)
