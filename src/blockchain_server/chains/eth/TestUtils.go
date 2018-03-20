package eth

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
)


const (
	veryLightScryptN = 2
	veryLightScryptP = 1
	tmpKeyStorePath = "/Users/cengliang/code/ethereum/data/00/keystore"
)
func tmpKeyStore(keystore_dir string) (*keystore.KeyStore) {
	//d, err := ioutil.TempDir("", "eth-keystore-test")
	//if err != nil {
	//	fmt.Printf("error:%s\n", err)
	//}
	newfunc := func (dir string) *keystore.KeyStore {
		return keystore.NewKeyStore(dir, veryLightScryptN, veryLightScryptP) }
	return newfunc(keystore_dir)
}

func DefualtKeyStore() (*keystore.KeyStore){
	return tmpKeyStore(tmpKeyStorePath)
}


