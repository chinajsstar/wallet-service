package handler

import (
	"fmt"
	"../db"
	"../../utils"
	"../../data"
	"crypto/sha512"
	"crypto"
	"io/ioutil"
	"../user"
	"encoding/base64"
	"sync"
)

type Auth struct{
	PrivateKey []byte

	Rwmu sync.RWMutex
	users map[string]*user.User
}

var defaultAuth = &Auth{}

func AuthInstance() *Auth{
	return defaultAuth
}

func (auth * Auth)getUser(licenseKey string) (*user.User, error)  {
	usr := func() *user.User{
		auth.Rwmu.RLock()
		defer auth.Rwmu.RUnlock()

		usr := auth.users[licenseKey]
		if usr != nil {
			return usr
		}
		return nil
	}()
	if usr != nil {
		return usr,nil
	}

	return func() (*user.User, error){
		auth.Rwmu.Lock()
		defer auth.Rwmu.Unlock()

		usr := auth.users[licenseKey]
		if usr != nil {
			return usr, nil
		}
		user, err := db.Read(licenseKey)
		if err != nil {
			fmt.Println("111--", err.Error())
			return nil, err
		}

		return user, nil
	}()
}

func (auth * Auth)Init() error {
	var err error
	auth.PrivateKey, err = ioutil.ReadFile("/Users/henly.liu/workspace/private.pem")
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// 创建
func (auth* Auth)CreateUser(licenseKey string, userName string, pubKey string)error{
	user := user.User{}
	user.LicenseKey = licenseKey
	user.Username = userName
	user.PubKey = pubKey

	return db.Create(&user)
}

// 验证数据
func (auth *Auth)AuthData(req *data.ServiceCenterDispatchData, ack *data.ServiceCenterDispatchAckData)  error{
	user, err := auth.getUser(req.Argv.LicenseKey)
	if err != nil {
		fmt.Println("111--", err.Error())
		return err
	}

	bmessage, err := base64.StdEncoding.DecodeString(req.Argv.Message)
	if err != nil {
		fmt.Println("222--", err.Error())
		return err
	}

	bsignature, err := base64.StdEncoding.DecodeString(req.Argv.Signature)
	if err != nil {
		fmt.Println("333--", err.Error())
		return err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write(bmessage)
	hashData = sha512.New().Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature, []byte(user.PubKey))
	if err != nil {
		fmt.Println("444--", err.Error())
		return err
	}

	// 解密数据
	var originData []byte
	originData, err = utils.RsaDecrypt(bmessage, auth.PrivateKey)
	if err != nil {
		fmt.Println("555--", err.Error())
		return err
	}

	ack.Value.Message = string(originData)
	ack.Value.Signature = ""
	ack.Value.LicenseKey = req.Argv.LicenseKey

	return nil
}

// 打包数据
func (auth *Auth)EncryptData(req *data.ServiceCenterDispatchData, ack *data.ServiceCenterDispatchAckData)  error{
	user, err := auth.getUser(req.Argv.LicenseKey)
	if err != nil {
		fmt.Println("111--", err.Error())
		return err
	}

	// 加密
	ack.Value.Message, err = func() (string, error){
		// 用用户的pub加密message ->encrypteddata
		encrypted, err := utils.RsaEncrypt([]byte(req.Argv.Message), []byte(user.PubKey))
		if err != nil {
			return "", err
		}

		return base64.StdEncoding.EncodeToString(encrypted), nil
	}()
	if err != nil {
		return err
	}

	// 签名
	ack.Value.Signature, err = func() (string, error){
		// 用服务器的pri签名encrypteddata ->signature
		var hashData []byte
		hs := sha512.New()
		hs.Write([]byte(ack.Value.Message))
		hashData = sha512.New().Sum(nil)

		var signData []byte
		signData, err = utils.RsaSign(crypto.SHA512, hashData, auth.PrivateKey)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		return base64.StdEncoding.EncodeToString(signData), nil
	}()
	if err != nil {
		return err
	}

	ack.Value.LicenseKey = req.Argv.LicenseKey

	return nil
}