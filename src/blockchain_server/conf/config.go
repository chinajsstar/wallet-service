package config

import (
	"blockchain_server/utils"
	"encoding/json"
	"io/ioutil"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"os"
)

var (
	configer Configer
	fordebug = true
)

type ClientConfig struct {
	Name                   string `json:"name"`
	RPC_url                string `json:"rpc_url"`
	Start_scan_Blocknumber uint64 `json:"start_sacn_blocknumber,string,omitempty"`
	TxConfirmNumber        uint64 `json:"confirmnumber,string,omitempty"`
}

type Configer struct {
	Cryptofile  	string	`json:"crypto_file"`
	Log_conf_file	string  `json:"log_conf_file"`
	Log_path		string  `json:"log_path"`
	Clientconfig    map[string]*ClientConfig
}


func (self *ClientConfig) Save() error {
	return configer.Save()
}

func (self *Configer) Save() error {
	jsondata, _ := json.Marshal(configer)
	err := ioutil.WriteFile(getConfigFilePath(), jsondata, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (self *Configer)ClientConfiger(coinname string) *ClientConfig{
	return self.Clientconfig[coinname]
}

func check(err error) {
    if err != nil {
    	fmt.Printf("error :%s\n", err)
    	os.Exit(1)
    }
}

func init () {
	configfile := getConfigFilePath()
	dat, err := ioutil.ReadFile(configfile)
	check(err)
	err = json.Unmarshal(dat, &configer)
	check(err)
	l4g.LoadConfiguration(configer.Log_conf_file)
}

func GetConfiger() (*Configer) {
	return &configer
}

func getConfigFilePath() string {
	if fordebug {
		return "/Users/cengliang/code/wallet-service/src/blockchain_server/res/app_debug.config"
	}
	return utils.CurrentRuningFileDir() + "/../res/app.config"
}

