package admin

const (
	AuthDataTypeString = 0
	AuthDataTypeUserMessage = 1
)

// 后台user message的格式
type UserMessage struct {
	SubUserKey 	string `json:"sub_user_key" doc:"指定用户请求的唯一key"`
	Message 	string `json:"message" doc:"实际的请求信息"`
}
