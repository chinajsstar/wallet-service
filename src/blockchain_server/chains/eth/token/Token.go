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
		to = common.BytesToAddress(input[16:36]).String()
		value = big.NewInt(0).SetBytes(input[36:68])
	}
	case "23b872dd": {//'transferFrom'
		if length!=100 {
			err = fmt.Errorf("Invalid tokenTx input data")
			return
		}
		from = common.BytesToAddress(input[16:36]).String()
		to = common.BytesToAddress(input[48:68]).String()
		value = big.NewInt(0).SetBytes(input[68:100])
	}
	default:{
		err = fmt.Errorf("TX doesn't transfer token")
	}
	}
	return
}

func BuildTokenTxInput(from *common.Address, to common.Address, value *big.Int) (input[]byte)  {
	if from!=nil {
		input = common.FromHex("0x23b872dd")
		input = append(input, common.LeftPadBytes(from.Bytes(), 32)[:]...)
	} else {
		input = common.FromHex("0xa9059cbb")
	}
	input = append(input, common.LeftPadBytes(to.Bytes(), 32)[:]...)
	input = append(input, common.LeftPadBytes(value.Bytes(), 32)[:]...)
	return
}

func BuildTokenApproveInput(address common.Address, value *big.Int) (input[]byte) {
	input = common.FromHex("0x095ea7b3")
	input = append(input, common.LeftPadBytes(address.Bytes(), 32)[:]...)
	input = append(input, common.LeftPadBytes(value.Bytes(), 32)[:]...)

	return
}

func BuildAllowanceInput(owner, spender common.Address) (input[] byte) {
	L4g.Trace("To implementaion!!")
	input = nil
	return
}

