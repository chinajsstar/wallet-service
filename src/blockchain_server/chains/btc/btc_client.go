package btc

import (
	"github.com/btcsuite/btcd/rpcclient"
	"sync"
	"log"
	//"github.com/btcsuite/btcwallet/waddrmgr"
	"fmt"
	"net/http"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"context"
	"encoding/json"
	"blockchain_server/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/btcec"
	"errors"
	l4g "github.com/alecthomas/log4go"
)

type BTCClient struct{
	*rpcclient.Client

	addressList 	[]string
	blockHeight 	uint64
	scanBlock 		uint64

	///
	connConfig  *rpcclient.ConnConfig

	blockNotification 	chan interface{}
	walletNotification 	chan interface{}
	//currentBlock        chan *waddrmgr.BlockStamp

	wg      sync.WaitGroup
	stopped bool
	quitMtx sync.Mutex

	httpAddr string
}

func NewBTCClient(connect, user, pass string, httpAddr string) (*BTCClient, error) {
	//connCfg := &rpcclient.ConnConfig{
	//	Host:         "localhost:18444",
	//	User:         "henly",
	//	Pass:         "henly123456",
	//	HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
	//	DisableTLS:   true, // Bitcoin core does not provide TLS by default
	//}

	client := &BTCClient{
		connConfig: &rpcclient.ConnConfig{
			Host:                 connect,
			User:                 user,
			Pass:                 pass,
			HTTPPostMode: 		  true, // Bitcoin core only supports HTTP POST mode
			DisableTLS:   		  true, // Bitcoin core does not provide TLS by default
		},
		blockNotification: 	 make(chan interface{}),
		walletNotification:  make(chan interface{}),
		//currentBlock:        make(chan *waddrmgr.BlockStamp),
		stopped:             true,
		httpAddr:			 httpAddr,
	}

	rpcClient, err := rpcclient.New(client.connConfig, nil)
	if err != nil {
		return nil, err
	}
	client.Client = rpcClient
	return client, nil
}

func (c *BTCClient) Start(ctx context.Context) error {
	c.startHttpServer(ctx, c.httpAddr)

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