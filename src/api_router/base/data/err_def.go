package data

import (
	"os"
	"fmt"
	"bastionpay_api/apibackend"
)

// error code and message
type ErrorInfo struct {
	Code 	int
	Msg 	string
	Groups  []string
}

var(
	err_msg map[int]*ErrorInfo
)

func AddErrMsg(errId int, errMsg string, groups []string)  {
	if _, ok := err_msg[errId]; ok{
		fmt.Printf("Error code %d exist!", errId)
		os.Exit(1)
	}
	err_msg[errId] = &ErrorInfo{errId,errMsg, groups}
}

func GetErrMsg(errId int) string {
	if msgInfo, ok := err_msg[errId]; ok{
		return msgInfo.Msg
	}
	return "service internal error"
}

func GetGroupErrMsg(group string) map[int]*ErrorInfo {
	if group == ""{
		return err_msg
	}

	errMsgs := make(map[int]*ErrorInfo)
	for _, v := range err_msg {
		for _, g := range v.Groups {
			if g == group {
				errMsgs[v.Code] = v
				break
			}
		}
	}

	return errMsgs
}

func init()  {
	err_msg = make(map[int]*ErrorInfo)

	AddErrMsg(NoErr, "成功", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrInternal,"内部错误", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrDataCorrupted, "数据损坏", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrCallFailed, "调用服务失败", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrIllegallyCall, "非法调用", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrNotFindAuth, "未发现验证服务", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrNotFindSrv, "未找到服务", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrNotFindFunction, "未找到方法", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrConnectSrvFailed, "连接服务失败", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})

	AddErrMsg(ErrAccountSrvNoUser, "用户不存在", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrAccountSrvUpdateProfile, "更新设置失败", []string{apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrAccountSrvListUsers, "获取用户列表失败", []string{apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrAccountSrvListUsersCount, "获取用户列表数量失败", []string{apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrAccountPubKeyParse, "公钥解析失败", []string{apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})

	AddErrMsg(ErrAuthSrvNoUserKey, "未找到用户id", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrAuthSrvNoPublicKey, "未找到公钥", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrAuthSrvNoApiLevel, "没有权限", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrAuthSrvUserFrozen, "用户被冻结", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrAuthSrvIllegalData, "非法数据", []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
	AddErrMsg(ErrAuthSrvIllegalDataType, "非法数据解析", []string{apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})

	AddErrMsg(ErrPushSrvPushData, "推送失败", []string{apibackend.HttpRouterUser, apibackend.HttpRouterAdmin})
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

	// listusers count - failed
	ErrAccountSrvListUsersCount		= 11104

	// pub key parse
	ErrAccountPubKeyParse			= 11105

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
