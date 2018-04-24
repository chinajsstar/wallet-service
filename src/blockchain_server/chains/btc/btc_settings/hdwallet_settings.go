package btc_settings

import (
	"github.com/btcsuite/btcutil/hdkeychain"
	"fmt"
	"blockchain_server/conf"
	"blockchain_server/types"
	l4g "github.com/alecthomas/log4go"
	"strconv"
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	"blockchain_server/chains/btc"
)

const (
	Name_KeySettings = "key_settings"
	Name_RPCSettings = "rpc_settings"

	Networktype_test = "test"
	Networktype_main = "main"
)

type RPCSettings struct {
	Rpc_url  		string 	`json:"rpc_url, string"`
	Username 		string	`json:"name, string"`
	Password 		string	`json:"password, string"`
	Endpoint		string	`json:"endpoint, string"`
	NetworkMode		string	`json:"network_mode, string"`
	// test or main network
}

func (self *RPCSettings) Isvalid() bool {
	return self.Rpc_url=="" || self.Endpoint=="" || self.Username=="" ||
		self.Password=="" || self.NetworkMode==""
}

func RPCSettings_from_MainConfig() (*RPCSettings, error) {
	client_config := Client_config()
	if client_config ==nil { return nil, fmt.Errorf("BTC get client config faild.")}

	if client_config.SubConfigs[Name_RPCSettings]==nil {
		message := "BTC sub-config not Founded"
		l4g.Trace(message)
		return nil, fmt.Errorf(message)
	}

	rpc_settings_tmp := new(RPCSettings)
	rpcsettings_tmp_json := client_config.SubConfigs[Name_RPCSettings]

	if v, isok := rpcsettings_tmp_json.(map[string]interface{}); isok {
		if e := v["rpc_url"]; e != nil {
			if t, isok := e.(string); isok {
				rpc_settings_tmp.Rpc_url = t
			}
		}
		if e := v["name"]; e != nil {
			if t, isok := e.(string); isok {
				rpc_settings_tmp.Username = t
			}
		}
		if e := v["password"]; e != nil {
			if t, isok := e.(string); isok {
				rpc_settings_tmp.Password = t
			}
		}

		if e := v["endpoint"]; e != nil {
			if t, isok := e.(string); isok {
				rpc_settings_tmp.Endpoint = t
			}
		}
		if e := v["network_mode"]; e!=nil {
			if t, isok := e.(string); isok {
				if t!= Networktype_main && t!= Networktype_test {
					l4g.Trace("Not supported network mode:[%s], set to:[%s]", t, Networktype_test)
					rpc_settings_tmp.NetworkMode = Networktype_test
				} else { rpc_settings_tmp.NetworkMode = t }
			}
		}
	}
	if rpc_settings_tmp.Isvalid() {
		return rpc_settings_tmp, nil
	}
	return nil, fmt.Errorf("BTC rpc settings invalid!")
}

type KeySettings struct {
	SeedValue        []byte
	Ext_pri          *hdkeychain.ExtendedKey
	Ext_pub          *hdkeychain.ExtendedKey
	Child_upto_index uint32
}

type keySettingsUnmarshal struct {
	SeedValue        string `json:"SeedValue,string"`
	Ext_pri          string `json:"extpri_primarykey,string"`
	Ext_pub          string `json:"extpub_primarykey,string"`
	Child_upto_index uint32 `json:"Child_upto_index,string"`
}

//func (self *KeySettings) UnmarshalJSON(input []byte) error {}

func (this *keySettingsUnmarshal) keySettings() (*KeySettings) {
	config := new(KeySettings)

	config.SeedValue, _= hex.DecodeString(this.SeedValue)
	config.Ext_pri, _ = hdkeychain.NewKeyFromString(this.Ext_pri)
	config.Ext_pub, _ = hdkeychain.NewKeyFromString(this.Ext_pub)
	config.Child_upto_index = this.Child_upto_index

	if config.IsValid() { return config }
	return nil
}

func (this *KeySettings) save() {
	clientconfig := Client_config()
	clientconfig.SubConfigs[Name_KeySettings] = this.keySettingsUnmarshal()
	clientconfig.Save()
}

func (this *KeySettings) keySettingsUnmarshal() *keySettingsUnmarshal {
	config := new(keySettingsUnmarshal)
	if nil!=this.Ext_pub { config.Ext_pub = this.Ext_pub.String() }
	if nil!=this.Ext_pri { config.Ext_pri = this.Ext_pri.String() }
	config.SeedValue = hex.EncodeToString(this.SeedValue)
	config.Child_upto_index = this.Child_upto_index
	return config
}

func (this *KeySettings) String() string {
	return fmt.Sprintf(`"
		SeedValue:%s
		Ext_pri : %s
		Ext_pub : %s
		Child_upto_index : %d"`, hex.EncodeToString(this.SeedValue),
		this.Ext_pri.String(), this.Ext_pub.String(), this.Child_upto_index)
}

func (this *KeySettings) IsValid() bool {
	return this.Ext_pri !=nil || this.Ext_pub !=nil
}

func Client_config() *config.ClientConfig {
	return config.MainConfiger().Clientconfig[types.Chain_bitcoin]
}

func KeySettings_from_MainConfig() (*KeySettings, error) {
	client_config := Client_config()
	if client_config ==nil { return nil, fmt.Errorf("BTC Client config not found!") }

	if client_config.SubConfigs[Name_KeySettings]==nil {
		return nil, fmt.Errorf("BTC key-settings not Founded")
	}

	keysettings_tmp := new(KeySettings)
	keysettings_tmp_json := client_config.SubConfigs[Name_KeySettings]

	if v, isok := keysettings_tmp_json.(map[string]interface{}); isok {
		if e := v["extpri_primarykey"]; e!=nil {
			if t, isok := e.(string); isok {
				keysettings_tmp.Ext_pri, _ = hdkeychain.NewKeyFromString(t)
			}
		}
		if e := v["extpub_primarykey"]; e!=nil {
			if t, isok := e.(string); isok {
				keysettings_tmp.Ext_pub, _ = hdkeychain.NewKeyFromString(t)
			}
		}
		if e := v["SeedValue"]; e!=nil {
			if t, isok := e.(string); isok {
				keysettings_tmp.SeedValue, _ = hex.DecodeString(t)
			}
		}
		if e := v["Child_upto_index"]; e!=nil {
			if t, isok := e.(string); isok {
				if index, err := strconv.ParseUint(t, 10, 32); err!=nil {
					return nil, err
				} else {
					keysettings_tmp.Child_upto_index = uint32(index)
				}

			}
		}
	}

	if keysettings_tmp.IsValid() {
		return keysettings_tmp,nil
	}
	return nil, fmt.Errorf("BTC key-setitngs Invalid!")
}

/*
func load_keysettings() error {
	var parent struct {
		For_debug bool `json:"debug"`
		sub struct {
			sub struct {
				Config *keySettingsUnmarshal `json:Key_settings`
			} `josn:"btc"`
		}`json:"Clientconfig"`
	}
	if Key_settings !=nil { return nil }

	if Key_settings = KeySettings_from_MainConfig(); Key_settings ==nil {
		message := "Load btc_settings settings faild!"
		l4g.Trace(message)
		return fmt.Errorf(message)
	} else { l4g.Trace("Load btc_settings settings success!") }

	// if try reload from settings from configuration file
	if false {
		cltconfig := Client_config()
		if cltconfig == nil {
			return fmt.Errorf("btc client is nil, no more to saying the key settings!")
		}
		configfile := config.GetConfigFilePath()
		if dat, err := ioutil.ReadFile(configfile); err!=nil {
			return err
		} else {
			err = json.Unmarshal(dat, &parent)
			if err!=nil { return err }
			if parent.sub.sub.Config!=nil {
				Key_settings = parent.sub.sub.Config.keySettings()

				if nil== Key_settings {
					return fmt.Errorf("Invalid btc sub-config.")
				}

				cltconfig.SubConfigs[Name_KeySettings] = Key_settings
			} else { return types.NewNotFound("btc sub-config not found!")}
			l4g.Trace("BTC sub configurations : %s", Key_settings.String())
		}
	}
	return nil
}
*/
func new_majorkey() error {
	// Generate a random SeedValue at the recommended length.
	seed, err := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	if err != nil {
		return err
	}

	var netparam *chaincfg.Params = nil
	// if for_debug is defined, we use test network parameters
	// else use main bitcoin network parameters
	if config.Debugmode {
		netparam = &chaincfg.RegressionNetParams
	} else {
		netparam = &chaincfg.MainNetParams
	}

	tmp_keysettings := KeySettings{SeedValue:seed}

	// Generate a new master node using the SeedValue.
	if key, err := hdkeychain.NewMaster(seed, netparam); err!=nil {
		return err
	} else {
		tmp_keysettings.Ext_pri = key
		if extpub_key, err := key.Neuter(); err == nil {
			tmp_keysettings.Ext_pub = extpub_key
			Client_config().SubConfigs[Name_KeySettings] = tmp_keysettings.keySettingsUnmarshal()
		} else { return err }
	}

	l4g.Trace(`
		--------------------------------------
		This is Very Very Very importent notic:
		--------------------------------------
		the extended master private key is generated, application will not keep it,
		it must be stored at safe/security place!
		--------------------------------------
		seed_value	  :[%s],
		extend_private:[%s],
		extend_public :[%s],
		--------------------------------------`,		hex.EncodeToString(tmp_keysettings.SeedValue),
		tmp_keysettings.Ext_pri.String(),
		tmp_keysettings.Ext_pub.String() )

	// TODO:这里可能应该把秘钥配置导出到一个单独的文件中
	// TODO:已方便备份
	config.MainConfiger().Save()
	return nil
}

