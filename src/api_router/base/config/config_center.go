package config

import (
	l4g "github.com/alecthomas/log4go"
)

// api gateway center config
type ConfigCenter struct{
	Port 			string `json:"port"`			// http port
	CenterVersion	string `json:"center_version"`	// center version
	CenterName		string `json:"center_name"`		// center name
	CenterPort 		string `json:"center_port"`		// center rpc port
}

// load center config from absolution path
func (cc *ConfigCenter)Load(absPath string) {
	err := LoadJsonNode(absPath, "center", cc)
	if err != nil {
		l4g.Crashf("", err)
	}
}