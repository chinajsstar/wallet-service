package config

import (
	"blockchain_server/utils"
	"encoding/json"
	"io/ioutil"
	L4g "github.com/alecthomas/log4go"
	"fmt"
	"os"
	"blockchain_server/types"
	"sync"
	apiutils "bastionpay_api/utils"
	"time"
)

var (
	configer            Configer
	Debugmode           = true
	SaveConfiguarations = true
	IsOnlinemode		= false
)

type ClientConfig struct {
	RPC_url                string `json:"rpc_url"`
	Name                   string `json:"name"`
	Start_scan_Blocknumber uint64 `json:"start_sacn_blocknumber"`
	TxConfirmNumber        uint64 `json:"confirmnumber,omitempty"`
	Tokens				   map[string]*types.Token `json:"Tokens"`

	SubConfigs map[string]interface{} `josn:"SubConfigs, omitempty"`
}

type Configer struct {
	Online_mode   string `json:"online_mode"`
	Cryptofile    string `json:"crypto_file"`
	LogconfigFile string `json:"log_conf_file"`
	LogsPath      string `json:"log_path"`
	Clientconfig  map[string]*ClientConfig

	mutx 			sync.Mutex
}

func (self *Configer) IsOnlineMode() bool {
	return self.Online_mode==types.Onlinemode_online
}

func (self *ClientConfig) Save() error {
	return configer.Save()
}

func (self *Configer) Cryptokeyfile() string {
	appDir, err := apiutils.GetAppDir()

	if err!=nil { return "" }
	return appDir + "/" + configer.Cryptofile
}

func (self *Configer) LogPath() string {
	appDir, err := apiutils.GetAppDir()

	if err!=nil { return "" }
	return  appDir + "/" + configer.LogsPath
}

func (self *Configer) L4gConfigFile() string {
	appDir, err := apiutils.GetAppDir()

	if err!=nil { return "" }
	return appDir + "/" + configer.LogconfigFile
}

func (self *ClientConfig) String() string {
	str := fmt.Sprintf(`
	client config informations: %s,
	rpc_url         : %s,
	start_scan_block: %d,
	confirmnumber   : %d`,
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
	if d, err = json.MarshalIndent(configer, "", "\t"); err!=nil {
		L4g.Error("Save configer error, message:%s", err.Error())
	} else {
		if err = ioutil.WriteFile(GetConfigFilePath(), d, os.ModePerm); err!=nil {
			L4g.Error("Save configer error, message:%s", err.Error())
		}
	}
	self.Unlock()

	return err
}

func (self *Configer)Trace() {
	L4g.Trace(`
	global config informations: 
	Keyfile      : %s,
	LogConfieFile: %s,
	LogsPath     : %s`,
		self.Cryptokeyfile(),
		self.L4gConfigFile(),
		self.LogPath() )

	for _, c := range self.Clientconfig {
		L4g.Trace(c.String())
	}
}
func (self *Configer)ClientConfiger(coinname string) *ClientConfig{
	return self.Clientconfig[coinname]
}

func l4g_fatalln(err error) {
	if err==nil {return}
	L4g.Trace(`
------------------Config init faild------------------
message : %s
appliaction will exit in 1 second!
------------------Config init faild------------------`, err)
	time.Sleep(time.Second)
	os.Exit(1)
}

func init () {
	configfile := GetConfigFilePath()
	dat, err := ioutil.ReadFile(configfile)
	l4g_fatalln(err)
	err = json.Unmarshal(dat, &configer)


	l4g_fatalln(err)

	IsOnlinemode = configer.Online_mode==types.Onlinemode_online

	configer.Trace()
}

func MainConfiger() (*Configer) {
	return &configer
}

func L4gConfigFile() string {
	return configer.L4gConfigFile()
}

func GetConfigFilePath() string {
	if Debugmode {
		L4g.Trace("running as debug version!")
		appDir, _:= apiutils.GetAppDir()
		return appDir + "/blockchain_server/res/app_debug.config"
	}
	L4g.Trace("running as release version!")
	return utils.CurrentRuningFileDir() + "./../res/app.config"
}

func LogPath() string {
	return configer.LogPath()
}
