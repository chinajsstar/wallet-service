package data

// 错误码
const(
	// 0-1000 通用
	NoErr 						= 0				// no error
	ErrCall						= 1				// call error,调用错误
	ErrCallText					= "内部调用错误"

	ErrNotFindAuth				= 2
	ErrNotFindAuthText			= "没有找到验证服务"

	ErrNotFindSrv				= 100
	ErrNotFindSrvText			= "没有找到服务"

	ErrClientConn				= 101
	ErrClientConnText 			= "服务连接异常"

	ErrSrvInternalErr           = 102
	ErrSrvInternalErrText		= "服务内部错误"

	// 1001-2000 auth_srv使用
	ErrAuthSrvIllegalData 		= 1001
	ErrAuthSrvIllegalDataText 	= "非法数据"

	// 2001-3000 account_srv使用
	ErrAccountSrvRegister		= 2001
	ErrAccountSrvRegisterText	= "用户注册失败"

)
