package btc

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"encoding/json"
	"net/http"
	"blockchain_server/types"
)

// Notifications returns a channel of parsed notifications sent by the remote
// bitcoin RPC server.  This channel must be continually read or the process
// may abort for running out memory, as unread notifications are queued for
// later reads.
//func (c *Client) Notifications() <-chan interface{} {
//	return c.dequeueNotification
//}

//func (c *Client) handleNotifications(ctx context.Context) {
//	for notify := range c.Notifications() {
//		var err error
//		switch x := notify.(type) {
//		case chain.ClientConnected:{}
//		case chain.BlockConnected:{}
//		case chain.BlockDisconnected:{}
//		case chain.RelevantTx: {
//		}
//		// The following are handled by the wallet's rescan
//		// goroutines, so just pass them there.
//		case *chain.RescanProgress:{}
//		case *chain.RescanFinished:{
//			atomic.StoreUint64(&c.startScanHeight, uint64(x.Height))
//		}
//		}
//		if err != nil {
//			l4g.Error("Cannot handle chain server notification: %v", err)
//		}
//	}
//}

func (c *Client) getHandlers() *rpcclient.NotificationHandlers {
	return nil
	//return &rpcclient.NotificationHandlers {
	//	OnRedeemingTx		  : nil, //c.onRedeemingTx,
	//	OnClientConnected	  : nil,
	//	OnBlockConnected 	  : nil, // c.onBlockConnected,
	//	OnBlockDisconnected   : nil, // c.onBlockDisconnected,
	//	OnRescanProgress 	  : nil,
	//	OnRescanFinished 	  : nil,
	//	OnRelevantTxAccepted  : nil,
	//	OnTxAccepted 		  : nil,
	//	OnRecvTx 			  : c.onRecvTx,
	//}
}

// handler maintains a queue of notifications and the current state (best
// block) of the chain.
func (c *Client) handler() {
	height, err := c.GetBlockCount()
	if err != nil {
		l4g.Trace("Failed to receive best block from chain server: ", err)
		c.Stop()
		c.wg.Done()
		return
	}

	l4g.Trace("first height=", height)

out:
	for {
		select {
		case n, ok := <- c.blockNotification:
			if !ok { continue }

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
					l4g.Error(err)
				}
				l4g.Trace("Block count: %d", blockCount)

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
			if !ok { continue }

			go func(n interface{}) {
				txHash, ok := n.(string);
				if ok == false {
					return
				}

				l4g.Trace("new txid, hash = ", txHash)

				// Get ...
				hs, err := chainhash.NewHashFromStr(txHash)

				if err != nil {
					l4g.Error("err:", err)
					return
				}

				//tx, err := c.GetRawTransaction(hs)
				if btx, err := c.GetTransaction(hs); err!=nil {
					return
				} else if tx, err := c.toTx(btx);err!=nil {
					l4g.Error("err:", err)
					return
				} else {
					c.rechargeTxNotification <- &types.RechargeTx{Tx:tx, Coin_name:types.Chain_bitcoin, Err:nil}
				}
				return
			}(n)
		case <- c.quit:
			break out
		}
	}

	c.Stop()
	c.wg.Done()
}

// start http server
func (c *Client)startHttpServer() error {
	// http
	http.Handle("/walletnotify", http.HandlerFunc(c.handleWalletNotify))
	http.Handle("/blocknotify",	 http.HandlerFunc(c.handleBlockNotify))
	http.Handle("/alertnotify",  http.HandlerFunc(c.handleAlertNotify))

	// TODO: shutdown http server
	go func() {
		defer c.wg.Done()
		err := http.ListenAndServe(c.rpc_settings.Http_server, nil)
		if err != nil {
			l4g.Trace("#Error:", err)
			return
		}
	}()

	return nil
}

// http handler
func (c *Client)handleWalletNotify(w http.ResponseWriter, req *http.Request) {
	vv := req.URL.Query()
	data := vv.Get("data")
	l4g.Trace("txid=", data)

	c.walletNotification <- data
}

// http handler
func (c *Client)handleBlockNotify(w http.ResponseWriter, req *http.Request) {
	vv := req.URL.Query()
	data := vv.Get("data")
	l4g.Trace("blsh=", data)

	c.blockNotification <- data
}

// http handler -- chain alert
func (c *Client)handleAlertNotify(w http.ResponseWriter, req *http.Request) {
	vv := req.URL.Query()

	data := vv.Get("data")
	l4g.Trace("alert=", data)
}
