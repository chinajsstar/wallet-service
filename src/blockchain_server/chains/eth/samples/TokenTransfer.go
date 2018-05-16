package main

import (
	"github.com/ethereum/go-ethereum/ethclient"
	l4g "github.com/alecthomas/log4go"
	"github.com/ethereum/go-ethereum/crypto"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"blockchain_server/chains/eth"
	bcrypto "blockchain_server/crypto"
)


var (
	key_stores = []string{`
{"address":"5c0ebcbaa341147c1b98d097bed99356f8b8340f","crypto":{"cipher":"aes-128-ctr","ciphertext":"1c26978ebfe3b29a57cbcb9e211ed6c19f49fe4bbc398926f43644e15b0d2da5","cipherparams":{"iv":"19ff0fec4c2669222abfec9cd79990a6"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"3f8a0c6ff3dc687626da342c95c6dfbb25d4df036afca4c99f085a7a774cdd7a"},"mac":"cbf34e1e52b93d96915b1136a3ff3e23e4b8a2ded88275385228699ea824437b"},"id":"37814409-bb2c-4bae-b8b4-8073eb03fbc6","version":3}`,

		`{"address":"0b9cb828028306470994a09b584bc7bad196131a","crypto":{"cipher":"aes-128-ctr","ciphertext":"2d66ee5412a1639af37404e744e45194e9e885d8acdf1014b8c3a6055d359e0a","cipherparams":{"iv":"097aad73d03056c03fe3c145ae7b1a49"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"0c5153383c88dd0f160043dee308b44e74a653972d8cc04ac5ed73e80647d4e7"},"mac":"3560f92f1c7ef6703a24fd8b2d65a90de9d9fd422aa8e9c6241648a9ce30446e"},"id":"baa8497a-45cd-4871-9a1b-813b9e425c71","version":3}`,

		`{"address":"b97205bf3831b2c45a561e77d10e575a78453a3a","crypto":{"cipher":"aes-128-ctr","ciphertext":"4c056b936d2c69ac9c7338e6aa867e19ba64b2e07e1cc9d584a666fe227f74e8","cipherparams":{"iv":"2090ba54958c59eaaa339b4fc244fa24"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"cd25d72e8a99b88e7b41ae81c01e309ef7615f74d1cc12f3d655cf25ea85ed05"},"mac":"1ec5195511762988da5ee45e47f11d3ccc3e8792f8a090ccfc68291b476ae367"},"id":"993366c8-0506-4280-a4e6-baa16c193d9b","version":3}`, }
	key0, key1, key2             *ecdsa.PrivateKey
	address0, address1, address2 string
)

func init() {
	key0 = ParseKeystore(key_stores[0]).PrivateKey
	key1 = ParseKeystore(key_stores[1]).PrivateKey
	key2 = ParseKeystore(key_stores[2]).PrivateKey

	address0 = crypto.PubkeyToAddress(key0.PublicKey).String()
	address1 = crypto.PubkeyToAddress(key1.PublicKey).String()
	address2 = crypto.PubkeyToAddress(key2.PublicKey).String()

	fmt.Printf(`
address0[%s]
address1[%s]
address2[%s]
`, address0, address1, address2)
}


//18160ddd -> totalSupply()
//70a08231 -> balanceOf(address)
//dd62ed3e -> allowance(address,address)
//a9059cbb -> transfer(address,uint256)
//095ea7b3 -> approve(address,uint256)
//23b872dd -> transferFrom(address,address,uint256)

func approveInput(address string, value int64) []byte {
	input := common.FromHex("0x095ea7b3")
	input = append(input, common.LeftPadBytes(common.FromHex(address), 32)[:]...)
	input = append(input, common.LeftPadBytes(big.NewInt(value).Bytes(), 32)[:]...)
	return input
}

func main() {
	var (
		err error
		nonce, gaslimit uint64
		gasprice *big.Int
		pendingcode []byte
	)

	value := 1024
	bankKey := key0
	bankaddress := common.HexToAddress(address0)
	approved := common.HexToAddress(address2)
	input := approveInput(address2, int64(value))
	contract := common.HexToAddress("0xf4df4932879d6034ff0a0d61f574669763d659c5")

	fmt.Printf("approve [%s] withdraw [%d]token from[%s]",
		approved.String(), value, bankaddress.String())

	c, err := ethclient.Dial("ws://127.0.0.1:8500")
	if nil != err {
		l4g.Error("create eth client error! message:%s", err.Error())
		return
	}

	if nonce, err = c.PendingNonceAt(context.TODO(), bankaddress); err != nil {
		l4g.Error("error:%#v", err.Error())
		return
	}

	if gasprice, err = c.SuggestGasPrice(context.TODO()); err != nil {
		fmt.Printf("failed to suggest gas price: %v", err)
		return
	}

	if pendingcode, err = c.PendingCodeAt(context.TODO(), contract); err != nil {
	} else {
		if len(pendingcode) == 0 { // 这是一个普通账号!!!
		} else { // 这是一个合约地址!!!
		}
	}

	msg := ethereum.CallMsg{
		From:bankaddress,
		To:&contract, Value:big.NewInt(0), Data:input}

	gaslimit, err = c.EstimateGas(context.TODO(), msg)
	if err != nil {
	}
	tx, err := etypes.NewTransaction(nonce, contract, big.NewInt(0), gaslimit, gasprice, input), nil
	if err!=nil {
		fmt.Printf("error :%#v\n", err)
		return
	}
	signedTx, err := etypes.SignTx(tx,
		etypes.HomesteadSigner{},
		bankKey)
	if err!=nil {
		fmt.Printf("error :%#v\n", err)
		return
	}

	err = c.SendTransaction(context.TODO(), signedTx)
	if err!=nil {
		fmt.Printf("error :%#v\n", err)
		return
	}

	fmt.Printf("transaction:%s", signedTx.String())
}

func ParseKeystore(keystr string) *keystore.Key {
	key, err := keystore.DecryptKey([]byte(keystr), "ko2005,./123eth")
	if err!=nil {
		fmt.Printf("error:%s", err.Error())
		return nil
	}
	cryptPrivKey, err := bcrypto.Encrypt(key.PrivateKey.D.Bytes())
	if err!=nil {
		fmt.Printf("error:%s", err.Error())
	}
	cryptKeyString := fmt.Sprintf("0x%x", cryptPrivKey)
	_, address, err := eth.ParseKey(cryptKeyString)
	if err!=nil {
		fmt.Printf("error:%s", err.Error())
	}
	fmt.Printf(`key : %s, address : %s`, cryptKeyString,address)
	return key
}

