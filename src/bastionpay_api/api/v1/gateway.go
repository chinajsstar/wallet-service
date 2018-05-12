package v1

// API info
type ApiInfo struct{
	Name 	string 	`json:"name" comment:"方法名称"`    	// api name
	Level 	int		`json:"level" comment:"方法权限"`		// api level, refer APILevel_*
}

// srv register data
type SrvRegisterData struct {
	Version      string `json:"version" comment:"服务版本"`    // srv version
	Srv          string `json:"srv" comment:"服务名称"`		// srv name
	Functions []ApiInfo `json:"functions" comment:"方法列表"`  // srv functions
}
