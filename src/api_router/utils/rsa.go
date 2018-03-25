package utils

import (
	"crypto"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"errors"
)

func RsaGen(bits int, priPath string, pubPath string) error {
	// pri
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil{
		return err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:"私钥",
		Bytes:derStream,
	}

	file, err := os.Create(priPath)
	if err != nil {
		return err
	}

	err = pem.Encode(file, block)
	if err != nil {
		return err
	}

	// pub
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	block = &pem.Block{
		Type:"公钥",
		Bytes:derPkix,
	}
	file, err = os.Create(pubPath)
	if err != nil {
		return err
	}

	err = pem.Encode(file, block)
	if err != nil {
		return err
	}

	return nil
}

func RsaEncrypt(originData []byte, pubKey []byte)([]byte, error){
	block, _ := pem.Decode(pubKey)
	if block == nil {
		return nil, errors.New("pub key error")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pub := pubInterface.(*rsa.PublicKey)

	return rsa.EncryptPKCS1v15(rand.Reader, pub, originData)
}

func RsaDecrypt(cipherData []byte, priKey []byte)([]byte, error){
	block, _ := pem.Decode(priKey)
	if block == nil {
		return nil, errors.New("pri key error")
	}

	priInterface, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, priInterface, cipherData)
}

func RsaSign(hash crypto.Hash, hashData []byte, priKey []byte)([]byte, error){
	block, _ := pem.Decode(priKey)
	if block == nil {
		return nil, errors.New("pri key error")
	}

	priInterface, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsa.SignPKCS1v15(rand.Reader, priInterface, hash, hashData)
}

func RsaVerify(hash crypto.Hash, hashData []byte, signData []byte, pubKey []byte)(error){
	block, _ := pem.Decode(pubKey)
	if block == nil {
		return errors.New("pub key error")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	pub := pubInterface.(*rsa.PublicKey)

	return rsa.VerifyPKCS1v15(pub, hash, hashData, signData)
}