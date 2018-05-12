package handler

import (
	"api_router/account_srv/db"
	"api_router/base/data"
	"io/ioutil"
	"bastionpay_api/api/v1"
	//service "api_router/base/service"
	service "api_router/base/service2"
	"encoding/json"
	"github.com/satori/go.uuid"
	l4g "github.com/alecthomas/log4go"
	"api_router/base/config"
)

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
	s.privateKey, err = ioutil.ReadFile(dir + "/" + config.BastionPayPrivateKey)
	if err != nil {
		l4g.Crashf("", err)
	}
	s.serverPublicKey, err = ioutil.ReadFile(dir + "/" + config.BastionPayPublicKey)
	if err != nil {
		l4g.Crashf("", err)
	}
}

func (s * Account)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	func(){
		input := v1.ReqUserRegister{}
		output := v1.AckUserRegister{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"register", data.APILevel_genesis, s.Register,
			"注册用户", string(b), input, output)
	}()

	func(){
		input := v1.ReqUserUpdateProfile{}
		output := v1.AckUserUpdateProfile{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"updateprofile", data.APILevel_admin, s.UpdateProfile,
			"更新开发者配置信息", string(b), input, output)
	}()

	func(){
		input := v1.ReqUserReadProfile{}
		output := v1.AckUserReadProfile{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"readprofile", data.APILevel_admin, s.ReadProfile,
			"读取开发者配置信息", string(b), input, output)
	}()

	func(){
		input := v1.ReqUserList{}
		output := v1.AckUserList{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"listusers", data.APILevel_admin, s.ListUsers,
			"列出所有用户", string(b), input, output)

	}()

	return nam
}

// 创建账号
func (s *Account) Register(req *data.SrvRequestData, res *data.SrvResponseData) {
	// from req
	reqUserRegister := v1.ReqUserRegister{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &reqUserRegister)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Data.Err = data.ErrDataCorrupted
		return
	}

	// userkey
	uuid, err := uuid.NewV4()
	if err != nil {
		l4g.Error("error create user key: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}
	userKey := uuid.String()

	// db
	err = db.Register(&reqUserRegister, userKey)
	if err != nil {
		l4g.Error("error create user: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// to ack
	ackUserCreate := v1.AckUserRegister{}
	ackUserCreate.UserKey = userKey

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

// 获取用户列表
// 登入
func (s *Account) ListUsers(req *data.SrvRequestData, res *data.SrvResponseData) {
	// from req
	reqUserList := v1.ReqUserList{}
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
	l4g.Info("list users: %s", res.Data.Value.Message)
}

// 获取key
func (s * Account) ReadProfile(req *data.SrvRequestData, res *data.SrvResponseData) {
	// from req
	reqReadProfile := v1.ReqUserReadProfile{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &reqReadProfile)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Data.Err = data.ErrDataCorrupted
		return
	}

	// load profile
	ackReadProfile, err := db.ReadProfile(reqReadProfile.UserKey)
	if err != nil {
		l4g.Error("error ReadKey: %s", err.Error())
		res.Data.Err = data.ErrAccountSrvNoUser
		return
	}

	// to ack
	dataAck, err := json.Marshal(ackReadProfile)
	if err != nil {
		l4g.Error("error Marshal: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// ok
	res.Data.Value.Message = string(dataAck)
	l4g.Info("update a user key: %s", res.Data.Value.Message)
}

// 更新key
func (s * Account) UpdateProfile(req *data.SrvRequestData, res *data.SrvResponseData) {
	// from req
	reqUpdateProfile := v1.ReqUserUpdateProfile{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &reqUpdateProfile)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Data.Err = data.ErrDataCorrupted
		return
	}

	// load old key
	oldUserReadProfile, err := db.ReadProfile(reqUpdateProfile.UserKey)
	if err != nil {
		l4g.Error("error ReadProfile: %s", err.Error())
		res.Data.Err = data.ErrAccountSrvNoUser
		return
	}

	if reqUpdateProfile.PublicKey == ""{
		reqUpdateProfile.PublicKey = oldUserReadProfile.PublicKey
	}
	if reqUpdateProfile.SourceIP == ""{
		reqUpdateProfile.SourceIP = oldUserReadProfile.SourceIP
	}
	if reqUpdateProfile.CallbackUrl == ""{
		reqUpdateProfile.CallbackUrl = oldUserReadProfile.CallbackUrl
	}

	// update key
	if err := db.UpdateProfile(&reqUpdateProfile); err != nil {
		l4g.Error("error update profile: %s", err.Error())
		res.Data.Err = data.ErrAccountSrvUpdateProfile
		return
	}

	// to ack
	ackUpdateProfile := v1.AckUserUpdateProfile{Status:"ok"}
	dataAck, err := json.Marshal(ackUpdateProfile)
	if err != nil {
		// 写回去
		oldUserUpdateProfile := v1.ReqUserUpdateProfile{}
		oldUserUpdateProfile.UserKey = oldUserReadProfile.UserKey
		oldUserUpdateProfile.PublicKey = oldUserReadProfile.PublicKey
		oldUserUpdateProfile.SourceIP = oldUserReadProfile.SourceIP
		oldUserUpdateProfile.CallbackUrl = oldUserReadProfile.CallbackUrl
		db.UpdateProfile(&oldUserUpdateProfile)
		l4g.Error("error Marshal: %s", err.Error())
		res.Data.Err = data.ErrInternal
		return
	}

	// ok
	res.Data.Value.Message = string(dataAck)
	l4g.Info("update a user key: %s", res.Data.Value.Message)
}