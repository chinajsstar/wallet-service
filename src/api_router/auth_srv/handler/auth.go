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
	"../../account_srv/user"
)

type Auth struct{
	privateKey []byte

	rwmu sync.RWMutex
	usersLicenseKey map[string]*user.UserLevel
}

var defaultAuth = &Auth{}

func AuthInstance() *Auth{
	return defaultAuth
}

func (auth * Auth)getUserLevel(licenseKey string) (*user.UserLevel, error)  {
	ul := func() *user.UserLevel{
		auth.rwmu.RLock()
		defer auth.rwmu.RUnlock()

		return auth.usersLicenseKey[licenseKey]
	}()
	if ul != nil {
		return ul,nil
	}

	return func() (*user.UserLevel, error){
		auth.rwmu.Lock()
		defer auth.rwmu.Unlock()

		ul := auth.usersLicenseKey[licenseKey]
		if ul != nil {
			return ul, nil
		}
		ul, err := db.ReadUserLevel(licenseKey)
		if err != nil {
			return nil, err
		}
		return ul, nil
	}()
}

func (auth * Auth)Init(dir string) error {
	var err error
	auth.privateKey, err = ioutil.ReadFile(dir+"/private.pem")
	if err != nil {
		fmt.Println(err)
		return err
	}

	auth.usersLicenseKey = make(map[string]*user.UserLevel)

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
func (auth *Auth)AuthData(req *data.SrvRequestData, res *data.SrvResponseData) {
	err := func() error{
		ul, err := auth.getUserLevel(req.Data.Argv.LicenseKey)
		if err != nil {
			fmt.Println("#Error AuthData--", err.Error())
			return err
		}

		if req.Context.Api.Level > ul.Level || ul.IsFrozen != 0{
			fmt.Println("#Error AuthData--", err)
			return errors.New("没权限或者被冻结")
		}

		bencrypted, err := base64.StdEncoding.DecodeString(req.Data.Argv.Message)
		if err != nil {
			fmt.Println("#Error AuthData--", err.Error())
			return err
		}

		bsignature, err := base64.StdEncoding.DecodeString(req.Data.Argv.Signature)
		if err != nil {
			fmt.Println("#Error AuthData--", err.Error())
			return err
		}

		// 验证签名
		var hashData []byte
		hs := sha512.New()
		hs.Write(bencrypted)
		hashData = hs.Sum(nil)

		err = utils.RsaVerify(crypto.SHA512, hashData, bsignature, []byte(ul.PublicKey))
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

		res.Data.Value.Message = string(originData)
		res.Data.Value.Signature = ""
		res.Data.Value.LicenseKey = req.Data.Argv.LicenseKey

		return nil
	}()

	if err != nil {
		res.Data.Err = data.ErrAuthSrvIllegalData
		res.Data.ErrMsg = data.ErrAuthSrvIllegalDataText
	}
}

// 打包数据
func (auth *Auth)EncryptData(req *data.SrvRequestData, res *data.SrvResponseData) {
	err := func() error{
		ul, err := auth.getUserLevel(req.Data.Argv.LicenseKey)
		if err != nil {
			fmt.Println("#Error EncryptData--", err.Error())
			return err
		}

		// 加密数据不需要判断权限
		/*
		if req.Context.Api.Level > ul.Level || ul.IsFrozen != 0{
			fmt.Println("#Error AuthData--", err.Error())
			return errors.New("没权限或者被冻结")
		}*/

		// 加密
		bencrypted, err := func() ([]byte, error){
			// 用用户的pub加密message ->encrypteddata
			bencrypted, err := utils.RsaEncrypt([]byte(req.Data.Argv.Message), []byte(ul.PublicKey), utils.RsaEncodeLimit2048)
			if err != nil {
				return nil, err
			}

			return bencrypted, nil
		}()
		if err != nil {
			return err
		}
		res.Data.Value.Message = base64.StdEncoding.EncodeToString(bencrypted)

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
		res.Data.Value.Signature = base64.StdEncoding.EncodeToString(bsignature)

		// licensekey
		res.Data.Value.LicenseKey = req.Data.Argv.LicenseKey

		return nil
	}()

	if err != nil {
		res.Data.Err = data.ErrAuthSrvIllegalData
		res.Data.ErrMsg = data.ErrAuthSrvIllegalDataText
	}
}