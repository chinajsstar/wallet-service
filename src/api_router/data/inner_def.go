package data

const(
	MethodCenterRegister   = "ServiceCenter.Register"	// 节点向服务中心注册请求
	MethodCenterUnRegister = "ServiceCenter.UnRegister"	// 节点向服务中心反注册请求

	MethodCenterPush   	   = "ServiceCenter.Push"		// 节点向服务中心推送消息
	MethodCenterDispatch   = "ServiceCenter.Dispatch"	// 节点向服务中心发送请求
	MethodNodeCall         = "ServiceNode.Call"			// 服务中心向节点发送请求
)

const(
	// 一般用户
	APILevel_client = 0

	// 一般后台管理员
	APILevel_admin = 100

	// 创世管理员
	APILevel_genesis = 200
)

// API信息
type ApiInfo struct{
	Name 	string 	`json:"name"`
	Level 	int		`json:"level"`
}
// 注册信息
type SrvRegisterData struct {
	Version      string `json:"version"`    // service version
	Srv          string `json:"srv"`		// service name
	Addr         string `json:"addr"`		// service ip address
	Functions []ApiInfo `json:"functions"`  // service functions
}

// 内部请求的上下文数据
type SrvRequestContext struct{
	Api ApiInfo `json:"api"`
	// others
}

// 内部RPC结构，在center中转时，增加请求权限信息
type SrvRequestData struct{
	Data UserRequestData 		`json:"data"`
	Context SrvRequestContext 	`json:"context"`
}
type SrvResponseData struct{
	Data UserResponseData 		`json:"data"`
}
