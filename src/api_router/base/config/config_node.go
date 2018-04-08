package config

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

// srv node config
type ConfigNode struct{
	SrvName 	string `json:"srv_name"`		// service name
	SrvVersion 	string `json:"srv_version"`	// service version
	SrvAddr 	string `json:"srv_addr"`		// service addr ip:port
	CenterAddr 	string `json:"center_addr""`	// center addr ip:port
}

// load srv node config from absolution path
func (cn *ConfigNode)Load(absPath string) error {
	var err error
	var data []byte
	data, err = ioutil.ReadFile(absPath)
	if err != nil {
		fmt.Println("#Error: ", err)
		return err
	}

	err = json.Unmarshal(data, cn)
	if err != nil {
		fmt.Println("#Error: ", err)
		return err
	}

	return err
}