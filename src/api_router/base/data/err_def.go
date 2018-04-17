package data

// error code and message
const(
	// /////////////////////////////////////////////////////
	// 0-1000 common errors
	// /////////////////////////////////////////////////////
	// no error
	NoErr 						= 0

	// data corrupted
	ErrDataCorrupted			= 1
	ErrDataCorruptedText		= "data corrupted"

	// call failed
	ErrCallFailed				= 2
	ErrCallFailedText			= "call failed"

	// illegally call
	ErrIllegallyCall			= 3
	ErrIllegallyCallText		= "illegally call"

	// not find auth service
	ErrNotFindAuth				= 4
	ErrNotFindAuthText			= "not find auth service"

	// not find service
	ErrNotFindSrv				= 100
	ErrNotFindSrvText			= "not find service"

	// not find function
	ErrNotFindFunction			= 101
	ErrNotFindFunctionText		= "not find function"

	// connect service failed
	ErrConnectSrvFailed			= 102
	ErrConnectSrvFailedText 	= "connect service failed"

	// push data failed
	ErrPushDataFailed           = 103
	ErrPushDataFailedText		= "push data failed"

	// /////////////////////////////////////////////////////
	// 1001-2000 account_srv errors
	// /////////////////////////////////////////////////////
	// register - user register failed
	ErrAccountSrvRegisterFailed			= 1001
	ErrAccountSrvRegisterFailedText		= "user register failed"

	// register - user name is repeated
	ErrAccountSrvUsernameRepeated		= 1002
	ErrAccountSrvUsernameRepeatedText	= "user name is repeated"

	// register - user phone is repeated
	ErrAccountSrvPhoneRepeated			= 1003
	ErrAccountSrvPhoneRepeatedText		= "user phone is repeated"

	// register - user email is repeated
	ErrAccountSrvEmailRepeated			= 1004
	ErrAccountSrvEmailRepeatedText		= "user email is repeated"

	// login - login failed
	ErrAccountSrvLogin					= 1101
	ErrAccountSrvLoginText				= "user login failed"

	// login - no user
	ErrAccountSrvNoUser					= 1102
	ErrAccountSrvNoUserText				= "no user"

	// login - wrong password
	ErrAccountSrvWrongPassword			= 1103
	ErrAccountSrvWrongPasswordText		= "wrong password"

	// updatepassword - failed
	ErrAccountSrvUpdatePassword			= 1201
	ErrAccountSrvUpdatePasswordText		= "update password failed"

	// listusers - failed
	ErrAccountSrvListUsers				= 1301
	ErrAccountSrvListUsersText			= "list users failed"

	// /////////////////////////////////////////////////////
	// 2001-3000 auth_srv errors
	// /////////////////////////////////////////////////////
	// illegal data
	ErrAuthSrvIllegalData 			= 2001
	ErrAuthSrvIllegalDataText 		= "illegal data"

	// no permission api
	ErrAuthSrvNoPermissionApi 		= 2002
	ErrAuthSrvNoPermissionApiText 	= "no permission api"
)
