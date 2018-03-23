package eth

import (
	"testing"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestNewAccount(t *testing.T) {
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





