package apibackend

// router's message format
const (
	ApiTypeString 		= 0		// from /api/ver/srv/function real message
	ApiTypeUserMessage 	= 1		// from /user/ver/srv/function UserMessage(subuserkey+real message)
	ApiTypeAdminMessage = 2		// from /admin/ver/srv/function AdminMessage(subuserkey+real message)

	HttpRouterApi		= "api"
	HttpRouterApiTest	= "apitest"
	HttpRouterUser		= "user"
	HttpRouterAdmin		= "admin"
)

// 后台user message的格式
type UserMessage struct {
	SubUserKey 	string `json:"sub_user_key" doc:"指定用户请求的唯一key"`
	Message 	string `json:"message" doc:"实际的请求信息"`
}

// admin message的格式
type AdminMessage struct {
	SubUserKey 	string `json:"sub_user_key" doc:"指定用户请求的唯一key"`
	Message 	string `json:"message" doc:"实际的请求信息"`
}
