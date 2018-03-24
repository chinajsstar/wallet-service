package eth

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"testing"
	//"context"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

func TestNewAccount(t *testing.T) {
	fmt.Println("*****************Testing NewAccount")
	account, _ := NewAccount()
	fmt.Printf("\"%s\", \n\"%s\"\n", account.PrivateKey, account.Address)
	//cryptKey := utils.String_cat_prefix(account.PrivateKey, "0x")
	//keyData, _ := hex.DecodeString(cryptKey)
	//keyPainData, _:= blockchain_server.Decrypto(keyData)

	//fmt.Printf("DecryptedPrivateKeyString: 0x%x\n", keyPainData)

	key, _ := ParseChiperkey(account.PrivateKey)
	fmt.Printf("decrypt----------------------\n")
	fmt.Printf("address:%s\n", crypto.PubkeyToAddress(key.PublicKey).String())
	fmt.Printf("publickey:%s\n", key.PublicKey.X.String())
	fmt.Printf("privatekey:0x%x\n", key.D.Bytes())
}

func TestPendingNonceAt(t *testing.T) {
	fmt.Println("*****************Testing PendingNonceAt")
	client, err := ethclient.Dial("ws://127.0.0.1:8500")
	if err != nil {
		fmt.Printf("error:%s", err)
		return
	}
	fmt.Printf("address is : 0x54B2E44D40D3Df64e38487DD4e145b3e6Ae25927")
	nonce, err := client.PendingNonceAt(context.TODO(), common.HexToAddress("0x54B2E44D40D3Df64e38487DD4e145b3e6Ae25927"))
	if err != nil {
		fmt.Printf("error:%s", err)
		return
	}
	fmt.Printf("nonce is %d\n", nonce)
}
