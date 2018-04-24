package config

import (
	"blockchain_server/utils"
	"encoding/json"
	"io/ioutil"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"os"
	"blockchain_server/types"
	"sync"
)

var (
	configer            Configer
	Debugmode           = true
	SaveConfiguarations = true
)

type ClientConfig struct {
	RPC_url                string `json:"rpc_url"`
	Name                   string `json:"name"`
	Start_scan_Blocknumber uint64 `json:"start_sacn_blocknumber,string,omitempty"`
	TxConfirmNumber        uint64 `json:"confirmnumber,string,omitempty"`
	Tokens				   map[string]*types.Token `json:"Tokens"`

	SubConfigs map[string]interface{} `josn:"SubConfigs, omitempty"`
}

type Configer struct {
	Online_mode		string  `json:"online_mode"`
	Cryptofile  	string	`json:"crypto_file"`
	Log_conf_file	string  `json:"log_conf_file"`
	Log_path		string  `json:"log_path"`
	Clientconfig    map[string]*ClientConfig

	mutx 			sync.Mutex
}


func (self *ClientConfig) Save() error {
	return configer.Save()
}


func (self *ClientConfig) String() string {
	str := fmt.Sprintf(`
	client config informations: %s,
	rpc_url:			%s,
	start_scan_block: 	%d,
	confirmnumber: 		%d`,
		self.Name, self.RPC_url, self.Start_scan_Blocknumber, self.TxConfirmNumber)
	for _, value := range self.Tokens {
		str += value.String()
	}
	return str
}

func (self *Configer) Lock() {
	self.mutx.Lock()
}

func (self *Configer) Unlock() {
	self.mutx.Unlock()
}

func (self *Configer) Save() error {

	if !SaveConfiguarations {return nil}

	var d []byte
	var err error = nil

	self.Lock()
	if d, err = json.Marshal(configer); err!=nil {
		l4g.Error("Save configer error, message:%s", err.Error())
	} else {
		if err = ioutil.WriteFile(GetConfigFilePath(), d, os.ModePerm); err!=nil {
			l4g.Error("Save configer error, message:%s", err.Error())
		}
	}
	self.Unlock()

	return err
}

func (self *Configer)Trace() {
	l4g.Trace(`
	global config informations: 
	crypte_key_file:			%s,
	log_config_file:			%s,
	log_path:					%s`,
		self.Cryptofile, self.Log_conf_file, self.Log_path)

	for _, c := range self.Clientconfig {
		l4g.Trace(c.String())
	}
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
	configfile := GetConfigFilePath()
	dat, err := ioutil.ReadFile(configfile)
	check(err)
	err = json.Unmarshal(dat, &configer)
	check(err)

	configer.Trace()
}

func MainConfiger() (*Configer) {
	return &configer
}

func GetConfigFilePath() string {
	if Debugmode {
		l4g.Trace("running as debug version!")
		return "/Users/cengliang/code/wallet-service/src/blockchain_server/res/app_debug.config"
	}
	l4g.Trace("running as release version!")
	return utils.CurrentRuningFileDir() + "./../res/app.config"
}

