package data

const(
	MethodServiceCenterRegister   = "ServiceCenter.Register"	// 服务向服务中心注册请求，对内
	MethodServiceCenterUnRegister = "ServiceCenter.UnRegister"	// 服务向服务中心反注册请求，对内
	MethodServiceCenterDispatch   = "ServiceCenter.Dispatch"	// 客户向服务中心发送请求，对外
	MethodServiceNodePingpong     = "ServiceNode.Pingpong"		// 服务中心向服务发送心跳，对内
	MethodServiceNodeCall         = "ServiceNode.Call"			// 服务中心向服务发送请求，对内
)

const(
	// 一般用户
	APILevel_client = 0

	// 一般后台管理员
	APILevel_admin = 100

	// 创世管理员
	APILevel_boss = 200
)

// API信息
type ApiInfo struct{
	Name 	string 	`json:"name"`
	Level 	int		`json:"level"`
}
// 注册信息
type ServiceCenterRegisterData struct {
	Version      string `json:"version"`    // service version
	Srv          string `json:"srv"`		// service name
	Addr         string `json:"addr"`		// service ip address
	Functions []ApiInfo `json:"functions"`  // service functions
}

// 内部RPC结构
type SrvDispatchData struct{
	SrvArgv ServiceCenterDispatchData `json:"data"`
	Api ApiInfo `json:"api"`
}
type SrvDispatchAckData struct{
	SrvAck ServiceCenterDispatchAckData `json:"data"`
}
