package handler

import (
	"fmt"
	"api_router/router/db"
	"api_router/base/utils"
	"api_router/base/data"
	"crypto/sha512"
	"crypto"
	"io/ioutil"
	"encoding/base64"
	"sync"
	"api_router/account_srv/user"
	l4g "github.com/alecthomas/log4go"
)

type Auth struct{
	privateKey []byte

	rwmu     sync.RWMutex
	usersKey map[string]*user.UserLevel
}

var defaultAuth = &Auth{}

func AuthInstance() *Auth{
	return defaultAuth
}

func (auth * Auth)Init(dir string) {
	var err error
	auth.privateKey, err = ioutil.ReadFile(dir+"/private.pem")
	if err != nil {
		l4g.Crashf("", err)
	}

	auth.usersKey = make(map[string]*user.UserLevel)
}

func (auth * Auth)getUserLevel(userKey string) (*user.UserLevel, error)  {
	ul := func() *user.UserLevel{
		auth.rwmu.RLock()
		defer auth.rwmu.RUnlock()

		return auth.usersKey[userKey]
	}()
	if ul != nil {
		return ul,nil
	}

	return func() (*user.UserLevel, error){
		auth.rwmu.Lock()
		defer auth.rwmu.Unlock()

		ul := auth.usersKey[userKey]
		if ul != nil {
			return ul, nil
		}
		ul, err := db.ReadUserLevel(userKey)
		if err != nil {
			return nil, err
		}
		auth.usersKey[userKey] = ul
		return ul, nil
	}()
}

// 验证数据
func (auth *Auth)AuthData(req *data.SrvRequestData, res *data.SrvResponseData) {
	ul, err := auth.getUserLevel(req.Data.Argv.UserKey)
	if err != nil {
		l4g.Error("(%s) get user level failed: %s",req.Data.Argv.UserKey, err.Error())
		res.Data.Err = data.ErrAuthSrvNoUserKey
		return
	}

	if ul.PublicKey == "" {
		l4g.Error("(%s-%s) failed: no public key", req.Data.Argv.UserKey, req.Data.Method.Function)
		res.Data.Err = data.ErrAuthSrvNoPublicKey
		return
	}

	if req.Context.ApiLever > ul.Level{
		l4g.Error("(%s-%s) failed: no api level", req.Data.Argv.UserKey, req.Data.Method.Function)
		res.Data.Err = data.ErrAuthSrvNoApiLevel
		return
	}

	if req.Context.ApiLever > ul.Level || ul.IsFrozen != 0{
		l4g.Error("(%s-%s) failed: user frozen", req.Data.Argv.UserKey, req.Data.Method.Function)
		res.Data.Err = data.ErrAuthSrvUserFrozen
		return
	}

	bencrypted, err := base64.StdEncoding.DecodeString(req.Data.Argv.Message)
	if err != nil {
		l4g.Error("error base64: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	bsignature, err := base64.StdEncoding.DecodeString(req.Data.Argv.Signature)
	if err != nil {
		l4g.Error("error base64: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write(bencrypted)
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature, []byte(ul.PublicKey))
	if err != nil {
		l4g.Error("verify: %s", err.Error())
		res.Data.Err = data.ErrAuthSrvIllegalData
		return
	}

	// 解密数据
	var originData []byte
	originData, err = utils.RsaDecrypt(bencrypted, auth.privateKey, utils.RsaDecodeLimit2048)
	if err != nil {
		l4g.Error("decrypt: %s", err.Error())
		res.Data.Err = data.ErrAuthSrvIllegalData
		return
	}

	// ok
	res.Data.Value.Message = string(originData)
}

// 打包数据
func (auth *Auth)EncryptData(req *data.SrvRequestData, res *data.SrvResponseData) {
	ul, err := auth.getUserLevel(req.Data.Argv.UserKey)
	if err != nil {
		l4g.Error("(%s) get user level failed: %s",req.Data.Argv.UserKey, err.Error())
		res.Data.Err = data.ErrAuthSrvNoUserKey
		return
	}

	if ul.PublicKey == "" {
		l4g.Error("(%s-%s) failed: no public key", req.Data.Argv.UserKey, req.Data.Method.Function)
		res.Data.Err = data.ErrAuthSrvNoPublicKey
		return
	}

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
		l4g.Error("encrypt: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

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
		l4g.Error("sign: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// ok
	res.Data.Value.Message = base64.StdEncoding.EncodeToString(bencrypted)
	res.Data.Value.Signature = base64.StdEncoding.EncodeToString(bsignature)
}