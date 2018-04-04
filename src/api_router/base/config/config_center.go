package config

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type ConfigCenter struct{
	Port string `json:"port"`
	WsPort string `json:"ws_port"`
	CenterName string `json:"center_name"`
	CenterPort string `json:"center_port"`
}

func (cc *ConfigCenter)Load(path string) error {
	var err error
	var data []byte
	data, err = ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("#Error: ", err)
		return err
	}

	err = json.Unmarshal(data, cc)
	if err != nil {
		fmt.Println("#Error: ", err)
		return err
	}

	return err
}