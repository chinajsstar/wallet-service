package main

import (
	"fmt"
	"math/big"
	"time"
)

func convertWeiToEther(w *big.Int) float64 {
	bigfloat := new(big.Float).SetInt(w)
	bigfloat = bigfloat.Mul(bigfloat, big.NewFloat(1.0e-18))

	f, _ := bigfloat.Float64()
	return f
}

func convertEtherToWei(e float64) *big.Int {
	tf := new(big.Float).SetFloat64(e)
	tf = tf.Mul(tf, big.NewFloat(1.0e+18))

	f, _ := tf.Float64()
	s := fmt.Sprintf("%.0f", f)

	ib, _ := new(big.Int).SetString(s, 10)

	return ib
}

func convertWeiToGWei(w *big.Int) float64 {

	bigfloat := new(big.Float).SetInt(w)
	bigfloat = bigfloat.Mul(bigfloat, big.NewFloat(1.0e-9))
	f, _ := bigfloat.Float64()
	return f
}

func main() {
	var nonces map[string]uint64
	nonces = make(map[string]uint64, 256)
	nonces["a"] = 1
	nonces["b"] = 1
	nonces["c"] = 1

	a := nonces["a"]
	b := nonces["b"]
	c := nonces["c"]
	d := nonces["d"]

	fmt.Printf("%d, %d, %d, %d", a, b, c, d)
	for i, nonce := range nonces {
		fmt.Printf("key: %s, nonce: %d", i, nonce)
	}

	if false {
		bigint := convertEtherToWei(10000000.00000001)
		fmt.Printf("%d\n", bigint.Uint64())

		min_gasprice := int64(141e8) // 14.1GWei
		price := big.NewInt(11e8)
		if -1==price.Cmp(big.NewInt(min_gasprice)) {
			price.SetInt64(min_gasprice)
		}
		gwei := convertWeiToGWei(price)
		fmt.Printf("%f", gwei)
	}
	time.Sleep(3 * time.Second)
}
