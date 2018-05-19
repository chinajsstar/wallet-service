package btc

import (
	"blockchain_server/chains/btc/btc_settings"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/chaincfg"
	l4g "github.com/alecthomas/log4go"
	"blockchain_server/conf"
	"errors"
	"fmt"
	"sync/atomic"
)

var instance *Client = nil

func (c *Client)init_rpc_settings() error {
	if !config.MainConfiger().IsOnlineMode() {
		l4g.Trace("btc client working on offline mode, will not init rpc!")
		return nil
	}

	if c.rpc_settings!=nil {return nil}

	if c.rpc_settings == nil {
		var err error
		if c.rpc_settings, err = btc_settings.RPCSettings_from_MainConfig(); err!=nil {
			return nil
		}
	}
	return nil
}

func (c *Client)init_rpc_config() error {
	if c.rpc_settings==nil {
		 if err:= c.init_rpc_settings(); err!=nil {
		 	return err
		 }
	}

	c.rpc_config = &rpcclient.ConnConfig {
		Host:                 c.rpc_settings.Rpc_url,
		//Endpoint:             c.rpc_settings.Endpoint,
		User:                 c.rpc_settings.Username,
		Pass:                 c.rpc_settings.Password,
		//DisableAutoReconnect: false,
		//DisableConnectOnNew:  true,
		HTTPPostMode:         true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:           true, // Bitcoin core does not provide TLS by default
	}

	return nil
}

// rpc_setings -> rpc_config -> rpc_client
func (c *Client) init_rpc_client() error {
	if !config.IsOnlinemode {
		return errors.New("btc client not config to 'online' mode")
	}
	var err error
	if err = c.init_rpc_settings(); err!=nil {
		l4g.Error("btc get network settings faild, message:%v", err)
		return err
	}

	if err = c.init_rpc_config(); err!=nil {
		l4g.Error("btc client online mode, rpc_config faild, message:%v", err)
		return err
	}

	if c.Client, err = rpcclient.New(c.rpc_config, c.getHandlers()); err!=nil {
		l4g.Error("btc client online mode, create rpc client error!, message:%s",
			err.Error())
		return err
	}

	return nil
}

func (c *Client) init_key_settings() error {
	if c.key_settings!=nil {return nil}

	var err error
	if c.key_settings, err = btc_settings.KeySettings_from_MainConfig(); err!=nil {
		return err
	}
	return nil
}

// init_chain_params, 根据配置文件获取当前连接主网还是测试网路或者其他网络!
// 生成pay-to-pubkey-hash address, 设置rpcclient配置的时候, 需要用到
func  (c *Client) init_chain_params() error {
	if !config.IsOnlinemode { return nil }

	if c.rpc_settings==nil {
		if err := c.init_rpc_settings(); err!=nil {
			return nil
		}
	}

	if c.rpc_settings.NetworkMode == btc_settings.Networktype_main {
		c.chain_params = &chaincfg.MainNetParams
	} else if c.rpc_settings.NetworkMode==btc_settings.Networktype_test {
		c.chain_params = &chaincfg.TestNet3Params
	} else {
		c.chain_params = &chaincfg.RegressionNetParams
	}

	c.confirminationNumber = btc_settings.Client_config().TxConfirmNumber

	return nil
}


func ClientInstance() (*Client, error) {
	if !btc_settings.InitOk() {
		message := "btc settings init faild."
		l4g.Trace(message)
		return nil, fmt.Errorf(message)
	}

	if instance!=nil { return instance, nil }

	inst := &Client{
		started:            false,
		blockNotification:  make(chan interface{}, 256),
		walletNotification: make(chan interface{}, 256),
		quit:               make(chan struct{}),
	}
	if err:=inst.init(); err!=nil {
		l4g.Error("btc client instance initialize faild, message:%v", err)
		return nil, err
	}

	return inst, nil
}

func (c *Client) init () error {
	if err := c.init_rpc_client(); err!=nil {return err}
	if err := c.init_chain_params(); err!=nil {return err}
	if _, err := c.refresh_blockheight(); err!=nil { return err }
	if err := c.init_key_settings(); err!=nil {return nil}

	l4g.Trace("coinname:%s, blockheight:%d", c.Name(), c.BlockHeight())

	configer := btc_settings.Client_config()
	atomic.StoreUint64(&c.startScanHeight, configer.Start_scan_Blocknumber)
	c.importAddressLabelName = "watchonly_addresses"

	return nil
}


