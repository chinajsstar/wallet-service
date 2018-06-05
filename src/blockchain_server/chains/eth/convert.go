package eth

import (
	"math/big"
	"fmt"
	"blockchain_server/utils"
)

func WeiToEther(w *big.Int) float64 {
	bigfloat := new(big.Float).SetInt(w)
	bigfloat = bigfloat.Mul(bigfloat, big.NewFloat(1.0e-18))
	f, _ := bigfloat.Float64()
	return utils.PrecisionN(f, 6)
}

func WeiToGWei(w *big.Int) float64 {
	bigfloat := new(big.Float).SetInt(w)
	bigfloat = bigfloat.Mul(bigfloat, big.NewFloat(1.0e-9))
	f, _ := bigfloat.Float64()
	return utils.PrecisionN(f, 6)
}

func EtherToWei(e float64) *big.Int {
	tf := new(big.Float).SetFloat64(e)
	tf = tf.Mul(tf, big.NewFloat(1.0e+18))

	f, _ := tf.Float64()
	s := fmt.Sprintf("%.0f", f)

	ib, _ := new(big.Int).SetString(s, 10)

	return ib
}

