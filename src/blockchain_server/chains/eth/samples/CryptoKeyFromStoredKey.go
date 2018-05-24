package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"fmt"
	"blockchain_server/crypto"
	"blockchain_server/chains/eth"
	"time"
)

var (
	key_stores = []string{
`{"address":"69f7337302aec7f6ae7915db3f31da865214e771","crypto":{"cipher":"aes-128-ctr","ciphertext":"5a024f248878bd284fcc1c6c4b48f4f95f8f089c6138efc5a0ab73e6ab60072f","cipherparams":{"iv":"7d4d18bcaf386a85180d5e4cc5a68b2b"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"b186c5238aec0ea7b787f7cdadf2e9b91b79241d10dbbe31fe9f33730183582e"},"mac":"68d59d96aee233f96dd81550e855d7f1b4f9396f8d4c4621c759d5048df4f59c"},"id":"e5e42425-1ddd-457f-b232-ee4f46e057c4","version":3}
`,
`{"address":"c1cfd731f440fa8fb8ed1b24056a514ffd36ef32","crypto":{"cipher":"aes-128-ctr","ciphertext":"24f9d61c7c7d011809ed04106edcf1f46eb4aa892be077232bd99bc16ce0e8c0","cipherparams":{"iv":"0d30fcdb9fd904cbb4767a9d3223670a"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"844043a148c2bc32ee7489f885ca6de86de9ebb8d8eb78b3b20a6b29f071e687"},"mac":"2d711be4c36f785ff9386c2a27790d2ccc969a0b481cd615db03cb9cd66a2546"},"id":"840821c3-3f8b-4bdf-bfcf-f6987fea9d47","version":3}
`,
`{"address":"43a957847d88c019c621847b1bd1b917741dbe3a","crypto":{"cipher":"aes-128-ctr","ciphertext":"2e2a305546e119c807a5d792cd96a14543e20563bcc689e2efee5221369259a5","cipherparams":{"iv":"0b53a66380b839e6edb3c6c93a06a577"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"08a8377f83c2a0f2fb0c05082df52c8b8173cd2dbc16e3bc643dabf7c62dc53a"},"mac":"521cdbf3fd7c3482826ee29c5bf05e9d4213119a92d1b38070d47306bf4d9d90"},"id":"5929b17a-46f1-4e0a-9b1c-d8c3295a7a96","version":3}
`, }
)
func main() {
	for _, s := range key_stores {
		_ = ParseKeystore(s)
	}
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
