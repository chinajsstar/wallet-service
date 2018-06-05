package data

import (
	"crypto/sha512"
	"encoding/base64"
	"bastionpay_api/utils"
	"crypto"
)

func EncryptionAndSignData(message []byte, userKey string, userPubKey []byte, basPrivKey []byte) (*SrvData, error) {
	var (
		err error
		encMessage []byte
		signature []byte
	)

	// encrypt
	encMessage, err = utils.RsaEncrypt(message, userPubKey, utils.RsaEncodeLimit2048)
	if err != nil {
		return nil, err
	}

	// signature
	hs := sha512.New()
	hs.Write(encMessage)
	hashData := hs.Sum(nil)

	signature, err = utils.RsaSign(crypto.SHA512, hashData, basPrivKey)
	if err != nil {
		return nil, err
	}

	srvData := &SrvData{}
	srvData.UserKey = userKey
	srvData.Message = base64.StdEncoding.EncodeToString(encMessage)
	srvData.Signature = base64.StdEncoding.EncodeToString(signature)

	return srvData, nil
}

func DecryptionAndVerifyData(srvData *SrvData, userPubKey []byte, basPrivKey []byte) ([]byte, error) {
	var (
		err error
		encMessage 	[]byte
		signature 	[]byte
	)

	encMessage, err = base64.StdEncoding.DecodeString(srvData.Message)
	if err != nil {
		return nil, err
	}

	signature, err = base64.StdEncoding.DecodeString(srvData.Signature)
	if err != nil {
		return nil, err
	}

	// verify
	hs := sha512.New()
	hs.Write([]byte(encMessage))
	hashData := hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, signature, userPubKey)
	if err != nil {
		return nil, err
	}

	// decrypt
	return utils.RsaDecrypt(encMessage, basPrivKey, utils.RsaDecodeLimit2048)
}
