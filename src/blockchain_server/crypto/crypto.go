package crypto
import (
	"github.com/pki-io/core/crypto"
	"crypto/ecdsa"
	"blockchain_server/conf"
	"blockchain_server/l4g"
	"os"
	"crypto/x509"
	"io/ioutil"
	"fmt"
	"reflect"
	"time"
	"blockchain_server/utils"
	"blockchain_server/types"
)

var (
	cryptoKey *ecdsa.PrivateKey
	L4g = L4G.GetL4g(types.Chain_eth)
)

func l4g_fatalln(v error) {
	if v==nil {return}
	L4g.Trace(`
------------------Config init faild------------------
message : %s
appliaction will exit in 1 second!
------------------Config init faild------------------`, v)
	time.Sleep(time.Second)
	os.Exit(1)
}

func init () {
	configer := config.MainConfiger()
	isExist, err := utils.PathExists(configer.Cryptokeyfile())

	l4g_fatalln(err)

	if isExist {
		priKeyBuffer, err := ioutil.ReadFile(configer.Cryptokeyfile())
		l4g_fatalln(err)

		cryptoKey, err = x509.ParseECPrivateKey(priKeyBuffer)
		l4g_fatalln(err)
	} else {
		newCryptoKey()
	}
}

func newCryptoKey() () {
	keyfile := config.MainConfiger().Cryptokeyfile()
	if false {
		if nil!=cryptoKey {
			return
		}

		if is_exist, err := utils.PathExists(keyfile); err!=nil {
			if is_exist {
				// load exist key file
			} else {
				// create new key file
			}
		}
	}

	cryptoKey, err := crypto.GenerateECKey()
	if err!=nil {
		l4g_fatalln(err)
	}

	priKeyBuffer, err := x509.MarshalECPrivateKey(cryptoKey)
	if err:=ioutil.WriteFile(keyfile, priKeyBuffer, 0444); err!=nil {
		l4g_fatalln(err)
	}
}

// 此函数设计为用来加密和解密账户的秘钥文件
func Encrypt(plaintext []byte) (chiper []byte, err error) {
	if nil==cryptoKey {
		return nil, fmt.Errorf("CryptKey is nil. keyfile:%s", config.MainConfiger().Cryptofile)
	}

	return crypto.Encrypt(plaintext, &cryptoKey.PublicKey)
}

func type_display(i interface{}) {
	switch k := i.(type) {
	default:
		fmt.Printf("key type : %s\n", reflect.TypeOf(k))
	}
}


// 此函数设计为用来加密和解密账户的秘钥文件
func Decrypto(chiper []byte) (plaintext []byte, err error) {
	if nil==cryptoKey {
		return nil, fmt.Errorf("CryptKey is nil. keyfile:%s", config.MainConfiger().Cryptofile)
	}

	return crypto.Decrypt(chiper, cryptoKey)
}

