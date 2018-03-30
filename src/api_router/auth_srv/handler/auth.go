package handler

import (
	"fmt"
	"../db"
	"../../base/utils"
	"../../data"
	"../../base/service"
	"crypto/sha512"
	"crypto"
	"io/ioutil"
	"encoding/base64"
	"sync"
	"errors"
)

type Auth struct{
	privateKey []byte

	rwmu sync.RWMutex
	usersLicenseKey map[string][]byte
}

var defaultAuth = &Auth{}

func AuthInstance() *Auth{
	return defaultAuth
}

func (auth * Auth)getUserPubKey(licenseKey string) ([]byte, error)  {
	key := func() []byte{
		auth.rwmu.RLock()
		defer auth.rwmu.RUnlock()

		key := auth.usersLicenseKey[licenseKey]
		if key != nil {
			return key
		}
		return key
	}()
	if key != nil {
		return key,nil
	}

	return func() ([]byte, error){
		auth.rwmu.Lock()
		defer auth.rwmu.Unlock()

		key := auth.usersLicenseKey[licenseKey]
		if key != nil {
			return key, nil
		}
		keyStr, err := db.ReadPubKey(licenseKey)
		if err != nil {
			return nil, err
		}
		return []byte(keyStr), nil
	}()
}

func (auth * Auth)Init() error {
	var err error
	auth.privateKey, err = ioutil.ReadFile("/Users/henly.liu/workspace/private_wallet.pem")
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (auth *Auth)RegisterApi(apis *[]data.ApiInfo, apisfunc *map[string]service.CallNodeApi) error  {
	regapi := func(name string, caller service.CallNodeApi, level int) error {
		if (*apisfunc)[name] != nil {
			return errors.New("api is already exist...")
		}

		*apis = append(*apis, data.ApiInfo{name, level})
		(*apisfunc)[name] = caller
		return nil
	}

	if err := regapi("authdata", service.CallNodeApi(auth.AuthData), data.APILevel_client); err != nil {
		return err
	}

	if err := regapi("encryptdata", service.CallNodeApi(auth.EncryptData), data.APILevel_client); err != nil {
		return err
	}

	return nil
}

// 验证数据
func (auth *Auth)AuthData(req *data.SrvDispatchData, ack *data.SrvDispatchAckData)  error{
	pubKey, err := auth.getUserPubKey(req.SrvArgv.Argv.LicenseKey)
	if err != nil {
		fmt.Println("#Error AuthData--", err.Error())
		return err
	}

	bencrypted, err := base64.StdEncoding.DecodeString(req.SrvArgv.Argv.Message)
	if err != nil {
		fmt.Println("#Error AuthData--", err.Error())
		return err
	}

	bsignature, err := base64.StdEncoding.DecodeString(req.SrvArgv.Argv.Signature)
	if err != nil {
		fmt.Println("#Error AuthData--", err.Error())
		return err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write(bencrypted)
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature, []byte(pubKey))
	if err != nil {
		fmt.Println("#Error AuthData--", err.Error())
		return err
	}

	// 解密数据
	var originData []byte
	originData, err = utils.RsaDecrypt(bencrypted, auth.privateKey, utils.RsaDecodeLimit2048)
	if err != nil {
		fmt.Println("#Error AuthData--", err.Error())
		return err
	}

	ack.SrvAck.Value.Message = string(originData)
	ack.SrvAck.Value.Signature = ""
	ack.SrvAck.Value.LicenseKey = req.SrvArgv.Argv.LicenseKey

	return nil
}

// 打包数据
func (auth *Auth)EncryptData(req *data.SrvDispatchData, ack *data.SrvDispatchAckData)  error{
	pubKey, err := auth.getUserPubKey(req.SrvArgv.Argv.LicenseKey)
	if err != nil {
		fmt.Println("#Error EncryptData--", err.Error())
		return err
	}

	// 加密
	bencrypted, err := func() ([]byte, error){
		// 用用户的pub加密message ->encrypteddata
		bencrypted, err := utils.RsaEncrypt([]byte(req.SrvArgv.Argv.Message), []byte(pubKey), utils.RsaEncodeLimit2048)
		if err != nil {
			return nil, err
		}

		return bencrypted, nil
	}()
	if err != nil {
		return err
	}
	ack.SrvAck.Value.Message = base64.StdEncoding.EncodeToString(bencrypted)

	// 签名
	bsignature, err := func() ([]byte, error){
		// 用服务器的pri签名encrypteddata ->signature
		var hashData []byte
		hs := sha512.New()
		hs.Write(bencrypted)
		hashData = hs.Sum(nil)

		bsignature, err := utils.RsaSign(crypto.SHA512, hashData, auth.privateKey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		return bsignature, nil
	}()
	if err != nil {
		return err
	}
	ack.SrvAck.Value.Signature = base64.StdEncoding.EncodeToString(bsignature)

	// licensekey
	ack.SrvAck.Value.LicenseKey = req.SrvArgv.Argv.LicenseKey

	return nil
}