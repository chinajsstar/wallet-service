package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"fmt"
	"blockchain_server/crypto"
	"blockchain_server/chains/eth"
	"time"
)

var (
	key_stores = []string{`
{"address":"5c0ebcbaa341147c1b98d097bed99356f8b8340f","crypto":{"cipher":"aes-128-ctr","ciphertext":"1c26978ebfe3b29a57cbcb9e211ed6c19f49fe4bbc398926f43644e15b0d2da5","cipherparams":{"iv":"19ff0fec4c2669222abfec9cd79990a6"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"3f8a0c6ff3dc687626da342c95c6dfbb25d4df036afca4c99f085a7a774cdd7a"},"mac":"cbf34e1e52b93d96915b1136a3ff3e23e4b8a2ded88275385228699ea824437b"},"id":"37814409-bb2c-4bae-b8b4-8073eb03fbc6","version":3}`,

`{"address":"0b9cb828028306470994a09b584bc7bad196131a","crypto":{"cipher":"aes-128-ctr","ciphertext":"2d66ee5412a1639af37404e744e45194e9e885d8acdf1014b8c3a6055d359e0a","cipherparams":{"iv":"097aad73d03056c03fe3c145ae7b1a49"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"0c5153383c88dd0f160043dee308b44e74a653972d8cc04ac5ed73e80647d4e7"},"mac":"3560f92f1c7ef6703a24fd8b2d65a90de9d9fd422aa8e9c6241648a9ce30446e"},"id":"baa8497a-45cd-4871-9a1b-813b9e425c71","version":3}`,

`{"address":"b97205bf3831b2c45a561e77d10e575a78453a3a","crypto":{"cipher":"aes-128-ctr","ciphertext":"4c056b936d2c69ac9c7338e6aa867e19ba64b2e07e1cc9d584a666fe227f74e8","cipherparams":{"iv":"2090ba54958c59eaaa339b4fc244fa24"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"cd25d72e8a99b88e7b41ae81c01e309ef7615f74d1cc12f3d655cf25ea85ed05"},"mac":"1ec5195511762988da5ee45e47f11d3ccc3e8792f8a090ccfc68291b476ae367"},"id":"993366c8-0506-4280-a4e6-baa16c193d9b","version":3}`, }
)
func main() {
	ParseKeystore(key_stores[0])
	ParseKeystore(key_stores[1])
	time.Sleep(time.Second * 3)
}

func ParseKeystore(keystr string) *keystore.Key {
	key, err := keystore.DecryptKey([]byte(keystr), "ko2005,./123eth")
	if err!=nil {
		fmt.Printf("error:%s", err.Error())
		return nil
	}

	cryptPrivKey, err := crypto.Encrypt(key.PrivateKey.D.Bytes())
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
