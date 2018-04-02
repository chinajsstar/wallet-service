package config

import (
	"io/ioutil"
	"encoding/json"
)

type ConfigNode struct{
	SrvName string `json:"srv_name"`
	SrvVersion string `json:"srv_version"`
	SrvAddr string `json:"srv_addr"`
	CenterAddr string `json:"center_addr""`
}

func (cn *ConfigNode)Load(path string) error {
	var err error
	var data []byte
	data, err = ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, cn)
	if err != nil {
		return err
	}

	return err
}