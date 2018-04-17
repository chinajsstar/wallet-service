package config

import (
	"io/ioutil"
	"encoding/json"
	l4g "github.com/alecthomas/log4go"
)

// api gateway center config
type ConfigCenter struct{
	Port 		string `json:"port"`			// http port
	WsPort 		string `json:"ws_port"`			// websocket port
	CenterName	string `json:"center_name"`		// center name
	CenterPort 	string `json:"center_port"`		// center rpc port
}

// load center config from absolution path
func (cc *ConfigCenter)Load(absPath string) {
	var err error
	var data []byte
	data, err = ioutil.ReadFile(absPath)
	if err != nil {
		l4g.Crashf("", err)
	}

	err = json.Unmarshal(data, cc)
	if err != nil {
		l4g.Crashf("", err)
	}
}