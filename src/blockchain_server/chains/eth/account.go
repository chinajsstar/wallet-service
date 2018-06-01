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
	bcrypto "blockchain_server/crypto"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/common"
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

func CheckAccount(account *wtypes.Account) bool {
	if account==nil {
		return false
	}

	if _, address, err := ParseKey(account.PrivateKey); err==nil {
		if address!=account.Address {
			return false
		}
		return true
	}
	return false
}

func NewAccount() (*wtypes.Account, error) {
	priKey, err := generateECKey()
	if err!=nil {
		return nil, err
	}

	priKeyData := math.PaddedBigBytes(priKey.D, 32)
	cryptPrivKey, err := bcrypto.Encrypt(priKeyData)
	if err!=nil {
		utils.Faltal_error(err)
	}
	cryptKeyString := common.ToHex(cryptPrivKey)

	account := wtypes.Account{PrivateKey: cryptKeyString,
		Address:crypto.PubkeyToAddress(priKey.PublicKey).String()}

	if CheckAccount(&account) {
		return &account, nil
	}

	return nil, fmt.Errorf("invalid account:%s", account.String())
}

func ParseKey(keyChiper string) (*ecdsa.PrivateKey, string, error) {
	cryptKeyData := common.FromHex(keyChiper)
	decryptBytes, err := bcrypto.Decrypto(cryptKeyData)
	if err!=nil {
		return nil, "",  err
	}

	privKey, err := crypto.ToECDSA(decryptBytes)
	if err!=nil {
		return nil, "",  err
	}

	return privKey, crypto.PubkeyToAddress(privKey.PublicKey).String(), nil
}

//func Chiperkey2Account(key string) (*wtypes.Account, error) {
//	privkey, err := ParseKey(key)
//	if err!=nil {
//		return nil, err
//	}
//
//	return &wtypes.Account{
//		PrivateKey: privkey,
//		Address: 	crypto.PubkeyToAddress(privkey.PublicKey)
//	}, nil
//}

