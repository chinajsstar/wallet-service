package btc

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"net/http"
	"bytes"
	"time"
	"io/ioutil"
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
//			L4G.Error("Cannot handle chain server notification: %v", err)
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

				L4g.Trace("new block, hash = %v ", blockHash)
				c.refresh_blockheight()

			}(n)

		case n, ok := <- c.walletNotification:
			if !ok { continue }

			if false {
				L4g.Trace("wallet notify is ignored! this is for testing!!!! %#v", n)
				continue
			}

			go func(n interface{}) {
				txHash, ok := n.(string)
				if ok == false {
					return
				}

				L4g.Trace("new txid, hash = %v", txHash)

				hs, err := chainhash.NewHashFromStr(txHash)

				if err != nil {
					L4g.Error("err:%v", err)
					return
				}

				if btx, err := c.GetTransaction(hs); err!=nil {
					L4g.Error("bitcoin get transaction error, message:%s", hs.String())
					return
				} else {
					if len(btx.Details)==0 || btx.Details[0].Category=="immature" {
						return
					}

					if tx, err := c.toTx(btx); err == nil {

						L4g.Trace("Bitcoin tx notified:%s", tx.String())

						c.rechargeTxNotification <- &types.RechargeTx{
							Tx: tx, Coin_name: types.Chain_bitcoin, Err: nil}
					} else {
						L4g.Error("bitcoin transaction to types.Transfer error:%v", err)
						return
					}
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
	http.Handle(c.rpc_settings.HttpCallback_wl, http.HandlerFunc(c.handleWaletNotify))
	http.Handle(c.rpc_settings.HttpCallback_bl, http.HandlerFunc(c.handleBlockNotify))
	http.Handle(c.rpc_settings.HttpCallback_al, http.HandlerFunc(c.handleAlertNotify))

	close := false

	http.Handle("/isok", http.HandlerFunc(func(r http.ResponseWriter, req *http.Request){
		bts, err := ioutil.ReadAll(req.Body)
		if err!=nil {
			L4g.Error("bitcoin http server handler /isok error, message:%s", err.Error())
		} else {
			message := string(bts)
			L4g.Trace("bitcoin http server get plain/text:%s", message)

			if message == "close" {
				close = true
				L4g.Trace("bitcoin http server close!!!")
			} else {
				r.Write([]byte("http server ok!!!!"))
			}
		}
	}) )

	go func() {
		defer c.wg.Done()
		err := http.ListenAndServe(c.rpc_settings.Http_server, nil)

		if close {
			L4g.Trace("========notic:bitcoin http server will shutdown!========")
			return
		}

		if err != nil {
			L4g.Error(`
==========start callback http service faild!!!!==========
error message:%s`,err.Error())
			return
		}
	}()

	time.Sleep(time.Second)

	if res, err:=http.Post("http://127.0.0.1" + c.rpc_settings.Http_server + "/isok",
		"text/plain; charset=utf-8", bytes.NewBuffer([]byte(`isok`))); err!=nil {
		L4g.Error("try post to http server faild, message:%s", err.Error())
		return err
	} else if bs, err := ioutil.ReadAll(res.Body); err==nil {
		L4g.Trace(string(bs))
		L4g.Trace(`
==========start callback http service success!!==========
http port	 :%s,
wallet notify:%s,
block  notify:%s, 
alert  notify:%s)`,
			c.rpc_settings.Http_server,
			c.rpc_settings.HttpCallback_wl,
			c.rpc_settings.HttpCallback_bl,
			c.rpc_settings.HttpCallback_al )
	}else  {
		return err
	}

	return nil
}

// http handler
func (c *Client) handleWaletNotify(w http.ResponseWriter, req *http.Request) {
	if err:=req.ParseForm(); err!=nil {
		if content, err := ioutil.ReadAll(req.Body); err==nil {
			L4g.Error("Bitcoin http server handle wallet notify, err content:%s",
				err.Error())
		}else {
			L4g.Error("Bitcoin http server cannot parse form data : %s", string(content))
		}
	}

	message := req.Form["txid"]
	L4g.Trace("get new transaction txId=%s", message)
	c.walletNotification <- message[0]
}

// http handler
func (c *Client)handleBlockNotify(w http.ResponseWriter, req *http.Request) {
	if err:=req.ParseForm(); err!=nil {
		if content, err := ioutil.ReadAll(req.Body); err==nil {
			L4g.Error("Bitcoin http server handle wallet notify, err content:%s",
				err.Error())
		}else {
			L4g.Error("Bitcoin http server cannot parse form data : %s", string(content))
		}
	}
	message := req.Form["blhash"]
	c.blockNotification <- message[0]
}

// http handler -- chain alert
func (c *Client)handleAlertNotify(w http.ResponseWriter, req *http.Request) {
	if err:=req.ParseForm(); err!=nil {
		if content, err := ioutil.ReadAll(req.Body); err==nil {
			L4g.Error("Bitcoin http server handle wallet notify, err content:%s",
				err.Error())
		}else {
			L4g.Error("Bitcoin http server cannot parse form data : %s", string(content))
		}
	}
	message := req.Form["alert"]
	L4g.Trace("alert=%s", message[0])
}

