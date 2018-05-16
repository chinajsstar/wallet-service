package token

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

//18160ddd -> totalSupply()
//70a08231 -> balanceOf(address)
//dd62ed3e -> allowance(address,address)
//a9059cbb -> transfer(address,uint256)
//095ea7b3 -> approve(address,uint256)
//23b872dd -> transferFrom(address,address,uint256)
func ParseTokenTxInput (input []byte) (from, to string, value *big.Int, err error) {
	length := len(input)
	if length < 4 {
		err = fmt.Errorf("Invalid TokenTx input data!")
		return
	}

	switch common.Bytes2Hex(input[0:4]) {
	case "a9059cbb": {// 'transfer'
		if length!=68 {
			err = fmt.Errorf("Invalid tokenTx input data")
			return
		}
		to = common.ToHex(input[16:36])
		value = big.NewInt(0).SetBytes(input[36:68])
	}
	case "23b872dd": {//'transferFrom'
		if length!=100 {
			err = fmt.Errorf("Invalid tokenTx input data")
			return
		}
		from = common.ToHex(input[16:36])
		to = common.ToHex(input[48:68])
		value = big.NewInt(0).SetBytes(input[68:100])
	}
	}
	return from, to, value, nil
}

func BuildTransferInput1 (to string, value *big.Int) (input []byte) {
	input = common.FromHex("0xa9059cbb")
	input = append(input, common.LeftPadBytes(
		common.FromHex(to), 32)[:]...)
	input = append(input, common.LeftPadBytes(value.Bytes(), 32)[:]...)
	return
}

func BuildTransferInput2 (from, to string, value *big.Int) (input[]byte)  {
	input = common.FromHex("0x23b872dd")
	input = append(input, common.LeftPadBytes(common.FromHex(from), 32)[:]...)
	input = append(input, common.LeftPadBytes(common.FromHex(to), 32)[:]...)
	input = append(input, common.LeftPadBytes(value.Bytes(), 32)[:]...)
	return
}

func BuildApproveInput(address string, value *big.Int) (input[]byte) {
	input = common.FromHex("0x095ea7b3")
	input = append(input, common.LeftPadBytes(common.FromHex(address), 32)[:]...)
	input = append(input, common.LeftPadBytes(value.Bytes(), 32)[:]...)
	return
}

