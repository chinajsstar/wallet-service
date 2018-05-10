package def

import "fmt"

const (
	ErrorSuccess = 0
	ErrorFailed  = 20000 + iota
)

var errorInfoMap map[int]string

func init() {
	errorInfoMap = make(map[int]string)
	errorInfoMap[ErrorSuccess] = "成功"
	errorInfoMap[ErrorFailed] = "失败"
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
