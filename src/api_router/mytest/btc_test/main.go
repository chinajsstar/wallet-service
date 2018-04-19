package main

import (
	"fmt"
	"time"
	"net/http"
	"log"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type BtcTest struct{
	client *rpcclient.Client
}

// start http server
func (bt *BtcTest)startHttpServer() error {
	// http
	log.Println("Start http server on ", "8076")

	http.Handle("/walletnotify", http.HandlerFunc(bt.handleWalletNotify))
	http.Handle("/blocknotify", http.HandlerFunc(bt.handleBlockNotify))
	http.Handle("/alertnotify", http.HandlerFunc(bt.handleAlertNotify))

	go func() {
		log.Println("Http server routine running... ")
		err := http.ListenAndServe(":8076", nil)
		if err != nil {
			fmt.Println("#Error:", err)
			return
		}
	}()

	return nil
}

// http handler
func (bt *BtcTest)handleWalletNotify(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	fmt.Println("path=", req.URL.Path)

	vv := req.URL.Query();

	data := vv.Get("data")
	fmt.Println("txid=", data)

	// Get ...
	hs, err := chainhash.NewHashFromStr(data)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	mb, err := bt.client.GetTransaction(hs)

	fmt.Println("tx info:", mb)
	return
}

// http handler
func (bt *BtcTest)handleBlockNotify(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	fmt.Println("path=", req.URL.Path)

	vv := req.URL.Query();

	data := vv.Get("data")
	fmt.Println("blockhash=", data)

	// Get the current block count.
	blockCount, err := bt.client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)

	// Get ...
	hs, err := chainhash.NewHashFromStr(data)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	mb, err := bt.client.GetBlock(hs)

	fmt.Println("block info:", mb)
	return
}

// http handler
func (bt *BtcTest)handleAlertNotify(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	fmt.Println("path=", req.URL.Path)

	vv := req.URL.Query();

	data := vv.Get("data")
	fmt.Println("alert=", data)

	return
}

func main() {
	var err error

	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:18444",
		User:         "henly",
		Pass:         "henly123456",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	bt := &BtcTest{}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	bt.client, err = rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer bt.client.Shutdown()

	// start notify...
	bt.startHttpServer()

	time.Sleep(time.Second*1)
	for ; ;  {
		fmt.Println("Input 'quit' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "quit" {
			break;
		}
	}

	fmt.Println("Waiting all routine quit...")
	fmt.Println("All routine is quit...")
}
