package main

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
)

func main() {
	// Generate a random SeedValue at the recommended length.
	seed, err := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	if err != nil {
		return
	}
	// if for_debug is defined, we use test network parameters
	// else use main bitcoin network parameters
	netparam := &chaincfg.RegressionNetParams
	// Generate a new master node using the SeedValue.
	if key, err := hdkeychain.NewMaster(seed, netparam); err != nil {
		return
	} else {
		if extpub_key, err := key.Neuter(); err == nil {
			for i := 0; i < 10; i++ {
				exk, _ := extpub_key.Child(uint32(i))
				k, _ := exk.ECPubKey()
				address, _ := exk.Address(netparam)

				fmt.Printf("pub key hex string is %s\n, encode_address is : %s\n string_address  is : %s\n",
					// import to btcwallet should use this string!!!
					hex.EncodeToString(k.SerializeCompressed()),
					address.EncodeAddress(),
					address.String())

			}
		}
	}
}
