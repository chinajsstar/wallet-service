package handler

import (
	"../db"
	"api_router/base/data"
	"crypto/rand"
	"io/ioutil"
	"api_router/account_srv/user"
	"api_router/base/service"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"encoding/json"
	"github.com/satori/go.uuid"
	l4g "github.com/alecthomas/log4go"
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
func (s *Account)Init(dir string) {
	var err error
	s.privateKey, err = ioutil.ReadFile(dir+"/private.pem")
	if err != nil {
		l4g.Crashf("", err)
	}
	s.serverPublicKey, err = ioutil.ReadFile(dir+"/public.pem")
	if err != nil {
		l4g.Crashf("", err)
	}
}

func (s * Account)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	apiInfo := data.ApiInfo{Name:"create", Level:data.APILevel_genesis}
	apiInfo.Example = ""
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:s.Create, ApiInfo:apiInfo}

	apiInfo = data.ApiInfo{Name:"listusers", Level:data.APILevel_admin}
	apiInfo.Example = "{\"id\":-1}"
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:s.ListUsers, ApiInfo:apiInfo}

	apiInfo = data.ApiInfo{Name:"login", Level:data.APILevel_client}
	apiInfo.Example = ""
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:s.Login, ApiInfo:apiInfo}

	apiInfo = data.ApiInfo{Name:"updatepassword", Level:data.APILevel_admin}
	apiInfo.Example = ""
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:s.UpdatePassword, ApiInfo:apiInfo}

	return nam
}

// 创建账号
func (s *Account) Create(req *data.SrvRequestData, res *data.SrvResponseData) {
	// from req
	reqUserCreate := user.ReqUserCreate{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &reqUserCreate)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Data.Err = data.ErrDataCorrupted
		return
	}

	// password
	salt := random(16)
	pswByte, err := bcrypt.GenerateFromPassword([]byte(x+salt+reqUserCreate.Password), 10)
	if err != nil {
		l4g.Error("error create salt password: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}
	psw := base64.StdEncoding.EncodeToString(pswByte)

	// userkey
	uuid, err := uuid.NewV4()
	if err != nil {
		l4g.Error("error create user key: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}
	userKey := uuid.String()

	// find if exist a user
	foundUser, err := db.ReadUser(userKey, reqUserCreate.UserName, reqUserCreate.Phone, reqUserCreate.Email)
	if foundUser != nil {
		if foundUser.UserKey == userKey {
			l4g.Error("error create a repeat user key: %s", userKey)
			res.Data.Err = data.ErrInternal
			return
		}

		if foundUser.UserName == reqUserCreate.UserName {
			l4g.Error("error create a repeat user name: %s", reqUserCreate.UserName)
			res.Data.Err = data.ErrAccountSrvUsernameRepeated
			return
		}

		if foundUser.Phone == reqUserCreate.Phone {
			l4g.Error("error create a repeat user phone: %s", reqUserCreate.Phone)
			res.Data.Err = data.ErrAccountSrvPhoneRepeated
			return
		}

		if foundUser.Email == reqUserCreate.Email {
			l4g.Error("error create a repeat user name: %s", reqUserCreate.Email)
			res.Data.Err = data.ErrAccountSrvEmailRepeated
			return
		}
	}

	// db
	err = db.Create(&reqUserCreate, userKey, salt, psw)
	if err != nil {
		l4g.Error("error create user: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// to ack
	ackUserCreate := user.AckUserCreate{}
	ackUserCreate.UserKey = userKey
	ackUserCreate.ServerPublicKey = string(s.serverPublicKey)

	dataAck, err := json.Marshal(ackUserCreate)
	if err != nil {
		db.Delete(userKey)
		l4g.Error("error Marshal: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// ok
	res.Data.Value.Message = string(dataAck)
	l4g.Info("create a new user: %s", res.Data.Value.Message)
}

// 登入
func (s *Account) Login(req *data.SrvRequestData, res *data.SrvResponseData) {
	// from req
	reqUserLogin := user.ReqUserLogin{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &reqUserLogin)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Data.Err = data.ErrDataCorrupted
		return
	}

	// read password
	user, salt, hashed, err := db.ReadPassword(reqUserLogin.UserName, reqUserLogin.Phone, reqUserLogin.Email)
	if err != nil {
		l4g.Error("error no user: %s", err.Error())
		res.Data.Err = data.ErrAccountSrvNoUser
		return
	}

	oldPswHash, err := base64.StdEncoding.DecodeString(hashed)
	if err != nil {
		l4g.Error("error base64: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// compare password
	if err := bcrypt.CompareHashAndPassword(oldPswHash, []byte(x+salt+reqUserLogin.Password)); err != nil {
		l4g.Error("error password: %s", err.Error())
		res.Data.Err = data.ErrAccountSrvWrongPassword
		return
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
	dataAck, err := json.Marshal(user)
	if err != nil {
		l4g.Error("error Marshal: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// ok
	res.Data.Value.Message = string(dataAck)
	l4g.Info("login a user: %s", res.Data.Value.Message)
}

// 登陆
func (s *Account) Logout(req *data.SrvRequestData, res *data.SrvResponseData)  {
	// TODO: 登出session处理
	//return db.DeleteSession(req.SessionId)
	return
}

// 更新密码
func (s * Account) UpdatePassword(req *data.SrvRequestData, res *data.SrvResponseData) {
	// from req
	reqUpdatePsw := user.ReqUserUpdatePassword{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &reqUpdatePsw)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Data.Err = data.ErrDataCorrupted
		return
	}

	userInfo, salt, hashed, err := db.ReadPassword(reqUpdatePsw.UserName, reqUpdatePsw.Phone, reqUpdatePsw.Email)
	if err != nil {
		l4g.Error("error no user: %s", err.Error())
		res.Data.Err = data.ErrAccountSrvNoUser
		return
	}

	oldPswHash, err := base64.StdEncoding.DecodeString(hashed)
	if err != nil {
		l4g.Error("error base64: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// compare old password
	if err := bcrypt.CompareHashAndPassword(oldPswHash, []byte(x+salt+reqUpdatePsw.OldPassword)); err != nil {
		l4g.Error("error password: %s", err.Error())
		res.Data.Err = data.ErrAccountSrvWrongPassword
		return
	}

	// reset new
	newSalt := random(16)
	newPswByte, err := bcrypt.GenerateFromPassword([]byte(x+newSalt+reqUpdatePsw.NewPassword), 10)
	if err != nil {
		l4g.Error("error create salt password: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}
	newPsw := base64.StdEncoding.EncodeToString(newPswByte)

	if err := db.UpdatePassword(userInfo.UserKey, newSalt, newPsw); err != nil {
		l4g.Error("error update password: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// to ack
	ackUpdatePsw := user.AckUserUpdatePassword{Status:"ok"}
	dataAck, err := json.Marshal(ackUpdatePsw)
	if err != nil {
		// 写回去
		db.UpdatePassword(userInfo.UserKey, salt, hashed)
		l4g.Error("error Marshal: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// ok
	res.Data.Value.Message = string(dataAck)
	l4g.Info("update a user password: %s", res.Data.Value.Message)
}

// 获取用户列表
// 登入
func (s *Account) ListUsers(req *data.SrvRequestData, res *data.SrvResponseData) {
	// from req
	reqUserList := user.ReqUserList{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &reqUserList)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Data.Err = data.ErrDataCorrupted
		return
	}

	const listnum = 10
	ackUserList, err := db.ListUsers(reqUserList.Id, listnum)
	if err != nil {
		l4g.Error("error ListUsers: %s", err.Error())
		res.Data.Err = data.ErrAccountSrvListUsers
		return
	}

	// to ack
	dataAck, err := json.Marshal(ackUserList)
	if err != nil {
		l4g.Error("error Marshal: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// ok
	res.Data.Value.Message = string(dataAck)
	l4g.Info("update a user password: %s", res.Data.Value.Message)
}