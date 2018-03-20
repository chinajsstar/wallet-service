package eth

import (
	"blockchain_server"
	mrand "math/rand"
	"time"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"crypto/rand"
	"blockchain_server/utils"
	wtypes "blockchain_server/types"
	"github.com/ethereum/go-ethereum/crypto"
	//lg4 "github.com/alecthomas/log4go"
	"encoding/hex"
	"crypto/x509"
)

func init() {
}

func getPassphrase(length uint32) string {
   const str = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
   bytes := []byte(str)
   var result []byte

	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))

   for i := 0; uint32(i) < length; i++ {
      result = append(result, bytes[r.Intn(int(len(bytes)))])
   }
   return string(result)
}

func generateECKey() (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Can't create ECDSA keys: %s", err)
	}
	return key, nil
}

func NewAccount() (*wtypes.Account, error) {
	priKey, err := generateECKey()
	if err!=nil {
		return nil, err
	}

	keyChiper, err := blockchain_server.Encrypt(priKey.D.Bytes())
	if err!=nil {
		utils.Faltal_error(err)
	}

	keyChiperString := fmt.Sprintf("0x%x", keyChiper)

	//decryptBytes, err := wallet.Decrypto(keyChiper)
	//if err!=nil {
	//	return nil, err
	//}
	//prikey_string = hex.EncodeToString(decryptBytes)

	address := crypto.PubkeyToAddress(priKey.PublicKey)
	addressString := fmt.Sprintf("0x%x", address)

	account := wtypes.Account{Private_key:keyChiperString, Address:addressString}
	//account := types.Account{keyChiperString, address}

	return &account, nil
}

func ParseChiperkey (keyChiper string) (*ecdsa.PrivateKey, error) {
	keyChiperStr := utils.String_cat_prefix(keyChiper, "0x")
	chiper, err := hex.DecodeString(keyChiperStr)
	if err!=nil {
		return nil, fmt.Errorf("Invalid private key")
	}

	decryptBytes, err := blockchain_server.Decrypto(chiper)
	if err!=nil {
		return nil, err
	}

	privateKey, err := x509.ParseECPrivateKey(decryptBytes)
	if err!=nil {
		return nil, err
	}
	//address := crypto.PubkeyToAddress(privateKey.PublicKey)
	//addressHexStr := "0x" + address.String()

	return privateKey, nil
}

//func CreateaAccountfromChiperkey(keyChiperStr string) (*wtypes.Account, error) {
//	keyChiperStr = utils.String_cat_prefix(keyChiperStr, "0x")
//	chiper, err := hex.DecodeString(keyChiperStr)
//	if err!=nil {
//		return nil, fmt.Errorf("Invalid private key")
//	}
//
//	decryptBytes, err := wallet.Decrypto(chiper)
//	if err!=nil {
//		return nil, err
//	}
//}


