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
	node *service.ServiceNode

	privateKey []byte
	serverPublicKey []byte
}

// 默认实例
var defaultAccount = &Account{}
func AccountInstance() *Account{
	return defaultAccount
}

// 初始化
func (s *Account)Init(dir string, node *service.ServiceNode) {
	var err error
	s.privateKey, err = ioutil.ReadFile(dir + "/" + config.BastionPayPrivateKey)
	if err != nil {
		l4g.Crashf("", err)
	}
	s.serverPublicKey, err = ioutil.ReadFile(dir + "/" + config.BastionPayPublicKey)
	if err != nil {
		l4g.Crashf("", err)
	}

	s.node = node
}

func (s * Account)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	func(){
		service.RegisterApi(&nam,
			"register", data.APILevel_genesis, s.Register)
	}()

	func(){
		service.RegisterApi(&nam,
			"updateprofile", data.APILevel_admin, s.UpdateProfile)
	}()

	func(){
		service.RegisterApi(&nam,
			"readprofile", data.APILevel_admin, s.ReadProfile)
	}()

	func(){
		service.RegisterApi(&nam,
			"listusers", data.APILevel_admin, s.ListUsers)

	}()

	return nam
}

func (s *Account)HandleNotify(req *data.SrvRequest){
	l4g.Info("HandleNotify-reloadUserLevel: do nothing")
}

// 创建账号
func (s *Account) Register(req *data.SrvRequest, res *data.SrvResponse) {
	// from req
	reqUserRegister := v1.ReqUserRegister{}
	err := json.Unmarshal([]byte(req.Argv.Message), &reqUserRegister)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Err = data.ErrDataCorrupted
		return
	}

	// userkey
	uuid, err := uuid.NewV4()
	if err != nil {
		l4g.Error("error create user key: %s", err.Error())
		res.Err = data.ErrInternal
		return
	}
	userKey := uuid.String()

	// db
	err = db.Register(&reqUserRegister, userKey)
	if err != nil {
		l4g.Error("error create user: %s", err.Error())
		res.Err = data.ErrInternal
		return
	}

	// to ack
	ackUserCreate := v1.AckUserRegister{}
	ackUserCreate.UserKey = userKey

	dataAck, err := json.Marshal(ackUserCreate)
	if err != nil {
		db.Delete(userKey)
		l4g.Error("error Marshal: %s", err.Error())
		res.Err = data.ErrInternal
		return
	}

	// ok
	res.Value.Message = string(dataAck)
	l4g.Info("create a new user: %s", res.Value.Message)
}

// 获取用户列表
// 登入
func (s *Account) ListUsers(req *data.SrvRequest, res *data.SrvResponse) {
	// from req
	reqUserList := v1.ReqUserList{}
	err := json.Unmarshal([]byte(req.Argv.Message), &reqUserList)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Err = data.ErrDataCorrupted
		return
	}

	const listnum = 10
	ackUserList, err := db.ListUsers(reqUserList.Id, listnum)
	if err != nil {
		l4g.Error("error ListUsers: %s", err.Error())
		res.Err = data.ErrAccountSrvListUsers
		return
	}

	// to ack
	dataAck, err := json.Marshal(ackUserList)
	if err != nil {
		l4g.Error("error Marshal: %s", err.Error())
		res.Err = data.ErrInternal
		return
	}

	// ok
	res.Value.Message = string(dataAck)
	l4g.Info("list users: %s", res.Value.Message)
}

// 获取key
func (s * Account) ReadProfile(req *data.SrvRequest, res *data.SrvResponse) {
	// from req
	reqReadProfile := v1.ReqUserReadProfile{}
	err := json.Unmarshal([]byte(req.Argv.Message), &reqReadProfile)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Err = data.ErrDataCorrupted
		return
	}

	// load profile
	ackReadProfile, err := db.ReadProfile(reqReadProfile.UserKey)
	if err != nil {
		l4g.Error("error ReadProfile: %s", err.Error())
		res.Err = data.ErrAccountSrvNoUser
		return
	}

	// to ack
	dataAck, err := json.Marshal(ackReadProfile)
	if err != nil {
		l4g.Error("error Marshal: %s", err.Error())
		res.Err = data.ErrInternal
		return
	}

	// ok
	res.Value.Message = string(dataAck)
	l4g.Info("read a user profile: %s", res.Value.Message)
}

// 更新key
func (s * Account) UpdateProfile(req *data.SrvRequest, res *data.SrvResponse) {
	// from req
	reqUpdateProfile := v1.ReqUserUpdateProfile{}
	err := json.Unmarshal([]byte(req.Argv.Message), &reqUpdateProfile)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Err = data.ErrDataCorrupted
		return
	}

	// load old key
	oldUserReadProfile, err := db.ReadProfile(reqUpdateProfile.UserKey)
	if err != nil {
		l4g.Error("error ReadProfile: %s", err.Error())
		res.Err = data.ErrAccountSrvNoUser
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
		res.Err = data.ErrAccountSrvUpdateProfile
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
		res.Err = data.ErrInternal
		return
	}

	// ok
	res.Value.Message = string(dataAck)
	l4g.Info("update a user profile: %s", res.Value.Message)

	// notify
	func(){
		notifyReq := data.SrvRequest{}
		notifyReq.Method.Version = "v1"
		notifyReq.Method.Srv = "account"
		notifyReq.Method.Function = "updateprofile"
		notifyReq.Argv.UserKey =""

		notifyData := v1.ReqUserUpdateProfile{}
		notifyData.UserKey = reqUpdateProfile.UserKey
		message, _ := json.Marshal(notifyData)

		notifyReq.Argv.Message = string(message)

		notifyRes := data.SrvResponse{}
		s.node.InnerNotify(&notifyReq, &notifyRes)

		l4g.Info("notify a user profile: %s", notifyReq.Argv.Message)
	}()
}