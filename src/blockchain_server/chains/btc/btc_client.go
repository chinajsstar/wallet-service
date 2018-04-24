package btc

import (
	"github.com/btcsuite/btcd/rpcclient"
	"sync"
	"log"
	//"github.com/btcsuite/btcwallet/waddrmgr"
	"net/http"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"context"
	"encoding/json"
	"blockchain_server/types"
	l4g "github.com/alecthomas/log4go"
	"blockchain_server/chains/btc/btc_settings"
	"blockchain_server/conf"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"strconv"
	"blockchain_server/utils"
	"encoding/binary"
	"encoding/hex"
	"github.com/btcsuite/btcutil"
)

type BTCClient struct {
	*rpcclient.Client
	rpcclient.NotificationHandlers


	addressList 	[]string
	blockHeight 	uint64
	scanBlock 		uint64

	connConfig  *rpcclient.ConnConfig

	blockNotification 	chan interface{}
	walletNotification 	chan interface{}
	//currentBlock        chan *waddrmgr.BlockStamp

	stopped bool
	wg      sync.WaitGroup
	quitMtx sync.Mutex

	key_settings *btc_settings.KeySettings
	rpc_settings *btc_settings.RPCSettings

	// 分层确定钱包使用扩散公钥, 和index, 生成子公钥, 并由公钥得到
	// pay-to-pubkey-hash
	// 作为收币地址, 可以说一个index代表一个收币地址,
	// accindexmtx, 用于index, 可能在多协程的情况下的访问同步
	accIndexMtx sync.Mutex

}


func get_rpc_settings() (*rpcclient.ConnConfig, error) {
	rpc_settings, err := btc_settings.RPCSettings_from_MainConfig()
	if err!=nil {return nil, err}

	conn_cfg := &rpcclient.ConnConfig{
		Host:         rpc_settings.Rpc_url,
		Endpoint:	  rpc_settings.Endpoint,
		User:         rpc_settings.Username,
		Pass:         rpc_settings.Password,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	return conn_cfg, nil
}

func NewBTCClient(connect, user, pass string, httpAddr string) (*BTCClient, error) {
	key_settings, err := btc_settings.KeySettings_from_MainConfig();
	if err!=nil {
		l4g.Error("btc client get key settings faild, message :%s", err.Error())
		return nil, err
	}

	// if online mode, create rpc client
	var rpc_client *rpcclient.Client = nil
	if config.MainConfiger().Online_mode==types.Onlinemode_online{
		if cfg, err := get_rpc_settings(); err!=nil {

			l4g.Error("btc client online mode, get_rpc_settings faild, message:%s",
				err.Error())
			return nil, err

			// TODO: implementaion notification handler
		} else if rpc_client, err = rpcclient.New(cfg, nil); err!=nil {
			l4g.Error("btc client online mode, create rpc client error!, message:%s",
				err.Error())
			return nil, err
		}
	}

	btcclient := &BTCClient{
		key_settings: key_settings,
		Client: 	rpc_client,
	}
	//btcclient.OnRelevantTxAccepted = btcclient.on_relevantTxAccepted
	//btcclient.On
	return btcclient, nil
}
//
//func (self )
// OnRelevantTxAccepted is invoked when an unmined transaction passes
// the client's transaction filter.
func (self *BTCClient) on_relevantTxAccepted (transaction []byte) {

}

// OnTxAccepted is invoked when a transaction is accepted into the
// memory pool.  It will only be invoked if a preceding call to
// NotifyNewTransactions with the verbose flag set to false has been
// made to register for the notification and the function is non-nil.
func (self *BTCClient)OnTxAccepted (hash *chainhash.Hash, amount btcutil.Amount) {
}



func (self *BTCClient) init () error {
	// online mode must has network configurations
	if config.MainConfiger().Online_mode==types.Onlinemode_online{
		var err error
		if self.rpc_settings, err = btc_settings.RPCSettings_from_MainConfig();err!=nil {
			l4g.Error("btc rpc-settings faild, message:%s", err.Error())
			return err
		}
	} else {
		l4g.Trace("BTC client init with 'offline' mode!")
	}
	return nil
}

func (c *BTCClient) Start(ctx context.Context) error {
	//c.startHttpServer(ctx, c.httpAddr)

	c.quitMtx.Lock()
	c.stopped = false
	c.quitMtx.Unlock()

	c.wg.Add(1)
	go c.handler(ctx)
	return nil
}

func (c *BTCClient) Stop() {
	c.quitMtx.Lock()

	if c.stopped == false{
		c.Client.Shutdown()

		l4g.Trace("stop...")
		close(c.blockNotification)
		close(c.walletNotification)

		c.stopped = true
	}

	c.quitMtx.Unlock()
}

func (c *BTCClient) WaitForShutdown() {
	c.Client.WaitForShutdown()
	c.wg.Wait()
}

func (c *BTCClient) Name() string {
	return types.Chain_bitcoin
}

// netparams, 根据配置文件获取当前连接主网还是测试网路或者其他网络!
// 生成pay-to-pubkey-hash address, 设置rpcclient配置的时候, 需要用到
func (self *BTCClient) netparams() *chaincfg.Params {
	if self.rpc_settings.NetworkMode==btc_settings.Networktype_main {
		return &chaincfg.MainNetParams
	} else {return &chaincfg.RegressionNetParams}
}


func indexToKey (index uint32, tolen uint, slat string) (string, error) {
	if tolen < 64 {tolen=64}

	index_hex := strconv.FormatUint(uint64(index), 16)

	md5 := utils.MD5(slat + strconv.FormatInt(int64(index), 16))

	index_len := len(index_hex)
	bs := make([]byte, 4)
	// 用32位, 4个字节, 转换成16进制, 形成字符串, 表示index的字符串的位数
	// 字符串需要站8个字符的位置!!
	binary.LittleEndian.PutUint32(bs, uint32(index_len))

	rdlen := tolen - (uint(len(index_hex + md5)) + 8)
	rdstr := utils.RandString(int(rdlen))

	return md5 + fmt.Sprintf("%08x", bs) + rdstr + index_hex, nil
}

func keyToIndex (index_str, slat string) (uint32, error) {
	// TODO: to check if data has been change by some one!
	index_len_hex := index_str[32:40]

	if index_len_bs, err := hex.DecodeString(index_len_hex); err!=nil {
		return 0, err
	} else {
		index_len := binary.LittleEndian.Uint32(index_len_bs)
		real_index_str := string(index_str[64-index_len:])
		if index, err := strconv.ParseUint(real_index_str, 16, 32); err!=nil {
			return 0, err
		} else { return uint32(index), nil }
	}
}

func (self *BTCClient) NewAccount(c uint32) ([]*types.Account, error) {
	var index_from, index_to uint32

	self.accIndexMtx.Lock()
	index_from = self.key_settings.Child_upto_index
	index_to = index_from + c
	self.accIndexMtx.Unlock()

	accs := make([]*types.Account, c)

	for i:=uint32(0); i<(index_to-index_from); i++ {
		if childpub, err := self.key_settings.Ext_pub.Child(index_from + i); err==nil {
			// converts the extended key to a standard bitcoin pay-to-pubkey-hash
			// address for the passed network.
			// AddressPubKeyHash is an Address for a pay-to-pubkey-hash (P2PKH)
			// transaction.
			if hash, err := childpub.Address(self.netparams()); err!=nil {
				l4g.Error("Convert child-extended-pub-key to address faild, message:%s", err.Error())
				return nil, err
			} else {
				if key, err := indexToKey(i, 64, "BTC_HD_Child_PubKey"); err==nil {
					accs[i] = &types.Account{Address:hash.String(), PrivateKey:key}
				} else {
					l4g.Error("BTC Convert index to child 'private key' faild, message:%s", err.Error())
					return nil, err
					}
			}
		} else {
			l4g.Error("BTC Get child public key faild, message:%s", err.Error())
			return nil, err
		}
	}

	return accs, nil
}

//func (c *BTCClient) NewAccount() (*types.Account, error) {
//	// Generate a random 256 bit seed
//	seed, err := hd_wallet_tmp.GenSeed(256)
//	// Create a master private key
//	masterprv := hd_wallet_tmp.MasterKey(seed)
//	// Convert a private key to public key
//	masterpub := masterprv.Pub()
//	// Generate new child key based on private or public key
//	childprv, err := masterprv.Child(0)
//	childpub, err := masterpub.Child(0)
//	// Create bitcoin address from public key
//	address := childpub.Address()
//
//	// Convenience string -> string Child and Address functions
//	walletstring := childpub.String()
//	childstring, err := hd_wallet_tmp.StringChild(walletstring,0)
//	childaddress, err := hd_wallet_tmp.StringAddress(childstring)
//}

/*
func (c *BTCClient) NewAccount()(*types.Account, error) {
	//// Generate a random seed at the recommended length.
	//seed, err := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	//if err != nil {
	//	l4g.Trace(err)
	//	return nil, err
	//}
	//
	//netParams := &chaincfg.MainNetParams
	netParams := &chaincfg.RegressionNetParams
	//
	//// Generate a new master node using the seed.
	//key, err := hdkeychain.NewMaster(seed, netParams)
	//if err != nil {
	//	l4g.Trace(err)
	//	return nil, err
	//}

	curve := btcec.S256()
	priv, err := btcec.NewPrivateKey(curve)
	if err != nil {
		l4g.Trace("%s: error:", err)
		return nil, err
	}
	if !curve.IsOnCurve(priv.PublicKey.X, priv.PublicKey.Y) {
		l4g.Trace("%s: public key invalid")
		return nil, errors.New("public key is invaild")
	}

	//priv, err := key.ECPrivKey()
	if err != nil {
		l4g.Trace("err:", err)
		return nil, err
	}

	wif, err := btcutil.NewWIF(priv, netParams, true)

	//bb, err := key.Address(netParams)
	pkHash := btcutil.Hash160(priv.PubKey().SerializeCompressed())
	bb , err := btcutil.NewAddressPubKeyHash(pkHash, netParams)

	fmt.Printf("account.privatekey:	%s\n", wif.String())
	fmt.Printf("account.publickey:	%s\n", bb.String())
	//fmt.Printf("account.address:	%s\n", account.Address)

	account := types.Account{PrivateKey:wif.String(), Address:bb.String()}
	return &account, nil
}
*/

// handler maintains a queue of notifications and the current state (best
// block) of the chain.
func (c *BTCClient) handler(ctx context.Context) {
	height, err := c.GetBlockCount()
	if err != nil {
		log.Println("Failed to receive best block from chain server: ", err)
		c.Stop()
		c.wg.Done()
		return
	}

	//bs := &waddrmgr.BlockStamp{Hash: *hash, Height: height}
	l4g.Trace("first height=", height)

out:
	for {
		select {
		case n, ok := <- c.blockNotification:
			if !ok {
				continue
			}

			go func(n interface{}) {
				// handler block
				blockHash, ok := n.(string)
				if ok ==false {
					return
				}

				l4g.Trace("new block, hash = ", blockHash)

				// Get the current block count.
				blockCount, err := c.GetBlockCount()
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Block count: %d", blockCount)

				// Get block by hash
				hs, err := chainhash.NewHashFromStr(blockHash)
				if err != nil {
					l4g.Trace("err:", err)
					return
				}
				mb, err := c.GetBlock(hs)
				b, err := json.Marshal(mb)

				l4g.Trace("block info:", string(b))
				return
			}(n)

		case n, ok := <- c.walletNotification:
			if !ok {
				continue
			}

			go func(n interface{}) {
				// handler wallet
				txHash, ok := n.(string);
				if ok == false {
					return
				}

				l4g.Trace("new txid, hash = ", txHash)

				// Get ...
				hs, err := chainhash.NewHashFromStr(txHash)
				if err != nil {
					l4g.Trace("err:", err)
					return
				}
				tx, err := c.GetRawTransaction(hs)

				b, err := json.Marshal(tx)

				l4g.Trace("tx info:", string(b))
				return
			}(n)

		//case c.currentBlock <- bs:
		//	l4g.Trace("new bs: ", c.currentBlock)
		case <-ctx.Done():
			l4g.Trace("ctx done...")
			break out
		}
	}

	c.Stop()
	c.wg.Done()
}

// start http server
func (c *BTCClient)startHttpServer(ctx context.Context, addr string) error {
	// http
	log.Println("Start http server on ", addr)

	http.Handle("/walletnotify", http.HandlerFunc(c.handleWalletNotify))
	http.Handle("/blocknotify", http.HandlerFunc(c.handleBlockNotify))
	http.Handle("/alertnotify", http.HandlerFunc(c.handleAlertNotify))

	go func() {
		log.Println("Http server routine running... ")
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			l4g.Trace("#Error:", err)
			return
		}
	}()

	return nil
}

// http handler
func (c *BTCClient)handleWalletNotify(w http.ResponseWriter, req *http.Request) {
	vv := req.URL.Query();
	data := vv.Get("data")
	l4g.Trace("txid=", data)

	c.walletNotification <- data
}

// http handler
func (c *BTCClient)handleBlockNotify(w http.ResponseWriter, req *http.Request) {
	vv := req.URL.Query();
	data := vv.Get("data")
	l4g.Trace("blockhash=", data)

	c.blockNotification <- data
}

// http handler -- chain alert
func (c *BTCClient)handleAlertNotify(w http.ResponseWriter, req *http.Request) {
	vv := req.URL.Query();

	data := vv.Get("data")
	l4g.Trace("alert=", data)
}