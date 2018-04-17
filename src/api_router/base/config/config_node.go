package config

import (
	"io/ioutil"
	"encoding/json"
	l4g "github.com/alecthomas/log4go"
)

// srv node config
type ConfigNode struct{
	SrvName 	string `json:"srv_name"`		// service name
	SrvVersion 	string `json:"srv_version"`		// service version
	SrvAddr 	string `json:"srv_addr"`		// service addr ip:port
	CenterAddr 	string `json:"center_addr"`		// center addr ip:port
}

// load srv node config from absolution path
func (cn *ConfigNode)Load(absPath string) {
	var err error
	var data []byte
	data, err = ioutil.ReadFile(absPath)
	if err != nil {
		l4g.Crashf("", err)
	}

	err = json.Unmarshal(data, cn)
	if err != nil {
		l4g.Crashf("", err)
	}
}