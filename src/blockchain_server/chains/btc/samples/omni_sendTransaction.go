package main

import (
	"github.com/btcsuite/btcd/rpcclient"
	"blockchain_server/types"
	"log"
	"blockchain_server/l4g"
)

//----------account[0] information:---------------
//child index:			1,
//virtual private key:	1af172762746ae5722948e342a0abddb01000000xGU55VythhpMhHkHcJ9tP2x1,
//real private key:		cTwx1uw9eRXh13zpFt28Y6P2k1crKfBx31F6cg7Z4F2mGZWueN7J,
//address string:		mmztibmBsXh3ezygzAMN52CtfCUwPWzXMy
//[00:51:25 CST 2018/06/04] [TRAC] (main.main:112)
//----------account[1] information:---------------
//child index:			2,
//virtual private key:	52e5f1cb31fec078b4891c2c05a9661501000000PHovvvnmhs1FS9Ev7ZFNeEv2,
//real private key:		cVpJKAxu8GhBLdyufUp2XcLJuxznm96qG2KA2REqBgKEyCARwCpX,
//address string:		mfetgXxCwsh9U19v3Kv5j77FdgbchHjkzK
//[00:51:25 CST 2018/06/04] [TRAC] (main.main:112)
//----------account[2] information:---------------
//child index:			3,
//virtual private key:	8dcd7c7cbae3bb60d1d89f6d6aed5ff301000000hcqzyyOIwoUXUsm6IgTl7Or3,
//real private key:		cPjkT4hJGRtj6MFuGKNg35BcMb9vvLg7t8Bo9NobZw3sncajy3KA,
//address string:		mwBgJASmXyh3TF8PEkoTY8yePfi44P6aNA
//[00:51:25 CST 2018/06/04] [TRAC] (main.main:112)
//----------account[3] information:---------------
//child index:			4,
//virtual private key:	1c1da3f9d7b66c855f0fb2f97f0c26f201000000QKdbN5ICgA5orsHkCzHG9Pv4,
//real private key:		cR2s6Rdt9cbsLNSgsksHw1Ha1PduJkWMEj8efFHHjFSKHyg9rpea,
//address string:		muLBXrSNdAsPBtRM9fS73p8QHZCnFwKKDi
//[00:51:25 CST 2018/06/04] [TRAC] (main.main:112)
//----------account[4] information:---------------
//child index:			5,
//virtual private key:	aff7117bf1eab04914602ed27e2a9d64010000006hoKDbrmFTiBBaeJ7y7lfCj5,
//real private key:		cTMiKnQZ8ZJeh16Cw98JNuxchiZ5hzSGsmk837ZALf35rD9QNhDG,
//address string:		mnXM8CCyXmzBZM1d11rUax4XBv4bYDNhJ6
var (
	// the following account create fro extend publick key
	// child index from 1 : to 5
	accs = []types.Account{
		{"mmztibmBsXh3ezygzAMN52CtfCUwPWzXMy", "1af172762746ae5722948e342a0abddb01000000xGU55VythhpMhHkHcJ9tP2x1"},
		{"mfetgXxCwsh9U19v3Kv5j77FdgbchHjkzK", "52e5f1cb31fec078b4891c2c05a9661501000000PHovvvnmhs1FS9Ev7ZFNeEv2"},
		{"mwBgJASmXyh3TF8PEkoTY8yePfi44P6aNA", "8dcd7c7cbae3bb60d1d89f6d6aed5ff301000000hcqzyyOIwoUXUsm6IgTl7Or3"},
		{"muLBXrSNdAsPBtRM9fS73p8QHZCnFwKKDi", "1c1da3f9d7b66c855f0fb2f97f0c26f201000000QKdbN5ICgA5orsHkCzHG9Pv4"},
		{"mnXM8CCyXmzBZM1d11rUax4XBv4bYDNhJ6", "aff7117bf1eab04914602ed27e2a9d64010000006hoKDbrmFTiBBaeJ7y7lfCj5"},
	}
	l4g = L4G.GetL4g(types.Chain_bitcoin)
	client *rpcclient.Client
)

func init () {
	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:18443",
		User:         "zengl",
		Pass:         "123456",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	l4g.Info("Block count: %d", blockCount)
}


func main() {
	//client.ListUnspentMinMaxAddresses()
}

