package eth

import (
	mrand "math/rand"
	"time"
	"crypto/ecdsa"
	"fmt"
	"crypto/rand"
	"blockchain_server/utils"
	wtypes "blockchain_server/types"
	"github.com/ethereum/go-ethereum/crypto"
	"encoding/hex"
	bcrypto "blockchain_server/crypto"

)

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
	key, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
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

	//keydata := crypto.FromECDSA(priKey)
	//fmt.Printf("1: 0x%x\n", keydata)

	cryptPrivKey, err := bcrypto.Encrypt(priKey.D.Bytes())
	if err!=nil {
		utils.Faltal_error(err)
	}
	cryptKeyString := fmt.Sprintf("0x%x", cryptPrivKey)
	//decryptBytes, err := wallet.Decrypto(cryptedPrivKeyByte)
	//if err!=nil {
	//	return nil, err
	//}
	//prikey_string = hex.EncodeToString(decryptBytes)

	account := wtypes.Account{PrivateKey: cryptKeyString, Address:crypto.PubkeyToAddress(priKey.PublicKey).String()}

	//fmt.Printf("account.privatekey:	0x%x\n", priKey.D.Bytes())
	//fmt.Printf("account.publickey:	%s\n", priKey.PublicKey.X.String())
	//fmt.Printf("account.address:	%s\n", account.Address)
	return &account, nil
}

func ParseChiperkey (keyChiper string) (*ecdsa.PrivateKey, error) {
	//fmt.Printf("crypt key : %s\n", keyChiper)
	cryptKeyHexString := utils.String_cat_prefix(keyChiper, "0x")
	cryptKeyData, err := hex.DecodeString(cryptKeyHexString )

	if err!=nil {
		return nil, fmt.Errorf("Invalid private key")
	}
	decryptBytes, err := bcrypto.Decrypto(cryptKeyData)
	if err!=nil {
		return nil, err
	}

	privateKey, err := crypto.ToECDSA(decryptBytes)
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


