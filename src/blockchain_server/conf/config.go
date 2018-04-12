package config

import (
	"blockchain_server/utils"
	"encoding/json"
	"io/ioutil"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"os"
	"blockchain_server/types"
)

var (
	configer    Configer
	DebugRuning = true
	SaveConfiguarations = false
)

type ClientConfig struct {
	Name                   string `json:"name"`
	RPC_url                string `json:"rpc_url"`
	Start_scan_Blocknumber uint64 `json:"start_sacn_blocknumber,string,omitempty"`
	TxConfirmNumber        uint64 `json:"confirmnumber,string,omitempty"`
	Tokens				   map[string]*types.Token `json:Tokens`
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

func (self *Configer) Save() error {
	if !SaveConfiguarations {return nil}

	jsondata, _ := json.Marshal(configer)
	err := ioutil.WriteFile(getConfigFilePath(), jsondata, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
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
	configfile := getConfigFilePath()
	dat, err := ioutil.ReadFile(configfile)
	check(err)
	err = json.Unmarshal(dat, &configer)
	check(err)

	l4g.LoadConfiguration(configer.Log_conf_file)

	configer.Trace()
}

func GetConfiger() (*Configer) {
	return &configer
}

func getConfigFilePath() string {
	if DebugRuning {
		l4g.Trace("running as debug version!")
		return "/Users/cengliang/code/wallet-service/src/blockchain_server/res/app_debug.config"
	}
	l4g.Trace("running as release version!")
	return utils.CurrentRuningFileDir() + "./../res/app.config"
}

