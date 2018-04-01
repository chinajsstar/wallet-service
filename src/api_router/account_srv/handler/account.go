package handler

import (
	"fmt"
	"../db"
	"../../data"
	"crypto/rand"
	"io/ioutil"
	"../user"
	"../../base/service"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"encoding/json"
	"github.com/satori/go.uuid"
	"errors"
)

const (
	//x = "cruft123"
	x = "super999"
)
var (
	alphanum = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)
func random(i int) string {
	bytes := make([]byte, i)
	for {
		rand.Read(bytes)
		for i, b := range bytes {
			bytes[i] = alphanum[b%byte(len(alphanum))]
		}
		return string(bytes)
	}
	return "ughwhy?!!!"
}

///////////////////////////////////////////////////////////////////////
// 账号管理
type Account struct{
	privateKey []byte
	serverPublicKey []byte
}

// 默认实例
var defaultAccount = &Account{}
func AccountInstance() *Account{
	return defaultAccount
}

// 初始化
func (s *Account)Init() error {
	var err error
	s.privateKey, err = ioutil.ReadFile("/Users/henly.liu/workspace/private_wallet.pem")
	if err != nil {
		fmt.Println(err)
		return err
	}
	s.serverPublicKey, err = ioutil.ReadFile("/Users/henly.liu/workspace/public_wallet.pem")
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (s *Account)RegisterApi(apis *[]data.ApiInfo, apisfunc *map[string]service.CallNodeApi) error  {
	regapi := func(name string, caller service.CallNodeApi, level int) error {
		if (*apisfunc)[name] != nil {
			fmt.Println("#error: api is already exist...", name)
			return errors.New("api is already exist...")
		}

		*apis = append(*apis, data.ApiInfo{name, level})
		(*apisfunc)[name] = caller
		return nil
	}

	if err := regapi("create", service.CallNodeApi(s.Create), data.APILevel_boss); err != nil {
		return err
	}

	if err := regapi("listusers", service.CallNodeApi(s.ListUsers), data.APILevel_admin); err != nil {
		return err
	}

	if err := regapi("login", service.CallNodeApi(s.Login), data.APILevel_admin); err != nil {
		return err
	}

	if err := regapi("updatepassword", service.CallNodeApi(s.UpdatePassword), data.APILevel_admin); err != nil {
		return err
	}

	return nil
}

// 创建账号
func (s *Account) Create(req *data.SrvRequestData, res *data.SrvResponseData) {
	err := func() error {
		// from req
		din := user.ReqUserCreate{}
		err := json.Unmarshal([]byte(req.Data.Argv.Message), &din)
		if err != nil {
			return err
		}

		// password
		salt := random(16)
		h, err := bcrypt.GenerateFromPassword([]byte(x+salt+din.Password), 10)
		if err != nil {
			return err
		}
		pp := base64.StdEncoding.EncodeToString(h)

		// licencekey
		u, err := uuid.NewV4()
		if err != nil {
			return err
		}
		licenseKey := u.String()

		// db
		err = db.Create(&din, licenseKey, salt, pp)
		if err != nil {
			return err
		}

		// to ack
		dout := user.AckUserCreate{}
		dout.LicenseKey = licenseKey
		dout.ServerPublicKey = string(s.serverPublicKey)

		d, err := json.Marshal(dout)
		if err != nil {
			db.DeleteByLicenseKey(dout.LicenseKey)
			return err
		}

		res.Data.Value.Message = string(d)
		return err
	}()

	if err != nil {
		res.Data.Err = data.ErrAccountSrvRegister
		res.Data.ErrMsg = data.ErrAccountSrvRegisterText
	}
}

// 登入
func (s *Account) Login(req *data.SrvRequestData, res *data.SrvResponseData) {
	err := func()error{
		// from req
		din := user.ReqUserLogin{}
		err := json.Unmarshal([]byte(req.Data.Argv.Message), &din)
		if err != nil {
			return err
		}

		dout, salt, hashed, err := db.ReadPassword(din.UserName, din.Phone, din.Email)
		if err != nil {
			return err
		}

		hh, err := base64.StdEncoding.DecodeString(hashed)
		if err != nil {
			return err
		}

		if err := bcrypt.CompareHashAndPassword(hh, []byte(x+salt+din.Password)); err != nil {
			return err
		}

		// TODO: add session
		// save session
		/*
		sess := &account.Session{
			Id:       random(128),
			Username: username,
			Created:  time.Now().Unix(),
			Expires:  time.Now().Add(time.Hour * 24 * 7).Unix(),
		}

		if err := db.CreateSession(sess); err != nil {
			return errors.InternalServerError("go.micro.srv.user.Login", err.Error())
		}
		rsp.Session = sess
		*/

		// to ack
		data, err := json.Marshal(dout)
		if err != nil {
			return err
		}

		res.Data.Value.Message = string(data)
		return nil
	}()

	if err != nil {
		res.Data.Err = data.ErrAccountSrvLogin
		res.Data.ErrMsg = data.ErrAccountSrvLoginText
	}
}

// 登陆
func (s *Account) Logout(req *data.SrvRequestData, res *data.SrvResponseData)  {
	// TODO: 登出session处理
	//return db.DeleteSession(req.SessionId)
	return
}

// 更新密码
func (s * Account) UpdatePassword(req *data.SrvRequestData, res *data.SrvResponseData) {
	err := func() error {
		// from req
		din := user.ReqUserUpdatePassword{}
		err := json.Unmarshal([]byte(req.Data.Argv.Message), &din)
		if err != nil {
			return err
		}

		d, salt, hashed, err := db.ReadPassword(din.UserName, din.Phone, din.Email)
		if err != nil {
			return err
		}

		hh, err := base64.StdEncoding.DecodeString(hashed)
		if err != nil {
			return err
		}

		if err := bcrypt.CompareHashAndPassword(hh, []byte(x+salt+din.OldPassword)); err != nil {
			return err
		}

		// reset new
		newSalt := random(16)
		h, err := bcrypt.GenerateFromPassword([]byte(x+newSalt+din.NewPassword), 10)
		if err != nil {
			return err
		}
		pp := base64.StdEncoding.EncodeToString(h)

		if err := db.UpdatePassword(d.Id, newSalt, pp); err != nil {
			return err
		}

		// to ack
		dout := user.AckUserUpdatePassword{Status:"ok"}
		data, err := json.Marshal(dout)
		if err != nil {
			// 写回去
			db.UpdatePassword(d.Id, salt, hashed)
			return err
		}

		res.Data.Value.Message = string(data)
		return nil
	}()

	if err != nil {
		res.Data.Err = data.ErrAccountSrvUpdatePassword
		res.Data.ErrMsg = data.ErrAccountSrvUpdatePasswordText
	}
}

// 获取用户列表
// 登入
func (s *Account) ListUsers(req *data.SrvRequestData, res *data.SrvResponseData) {
	err := func() error {
		// from req
		din := user.ReqUserList{}
		err := json.Unmarshal([]byte(req.Data.Argv.Message), &din)
		if err != nil {
			return err
		}

		const listnum = 10
		dout, err := db.ListUsers(din.Id, listnum)
		if err != nil {
			return err
		}

		// to ack
		data, err := json.Marshal(dout)
		if err != nil {
			return err
		}

		res.Data.Value.Message = string(data)
		return nil
	}()

	if err != nil {
		res.Data.Err = data.ErrAccountSrvListUsers
		res.Data.ErrMsg = data.ErrAccountSrvListUsersText
	}
}