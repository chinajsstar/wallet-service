package def

import "fmt"

const (
	ErrorSuccess  = 0            //成功
	ErrorParse    = 20000 + iota //指令解析错误
	ErrorParam                   //指令参数错误
	ErrorWallet                  //钱包错误
	ErrorDataBase                //数据库错误
)

var errorInfoMap map[int]string

func init() {
	errorInfoMap = make(map[int]string)
	errorInfoMap[ErrorSuccess] = "成功"
	errorInfoMap[ErrorParse] = "指令解析错误"
	errorInfoMap[ErrorParam] = "指令参数错误"
	errorInfoMap[ErrorDataBase] = "数据库错误"
}

func GetErrorMsg(errorID int) string {
	if v, ok := errorInfoMap[errorID]; ok {
		return v
	}
	return "未知错误"
}

func CheckError(errID int, extraMsg string) (int, string) {
	errMsg := ""
	if len(extraMsg) > 0 {
		errMsg = fmt.Sprintf("%s: %s", GetErrorMsg(errID), extraMsg)
	} else {
		errMsg = GetErrorMsg(errID)
	}
	return errID, errMsg
}
