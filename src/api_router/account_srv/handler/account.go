package handler

import (
	"api_router/account_srv/db"
	"api_router/base/data"
	"io/ioutil"
	//service "api_router/base/service"
	service "api_router/base/service2"
	"encoding/json"
	"github.com/satori/go.uuid"
	l4g "github.com/alecthomas/log4go"
	"api_router/base/config"
	"bastionpay_api/utils"
	"bastionpay_api/apibackend"
	"bastionpay_api/apibackend/v1/backend"
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

	func(){
		service.RegisterApi(&nam,
			"updatefrozen", data.APILevel_admin, s.UpdateFrozen)

	}()

	return nam
}

func (s *Account)HandleNotify(req *data.SrvRequest){
	l4g.Info("HandleNotify-reloadUserLevel: do nothing")
}

// 创建账号
func (s *Account) Register(req *data.SrvRequest, res *data.SrvResponse) {
	// from req
	reqUserRegister := backend.ReqUserRegister{}
	err := json.Unmarshal([]byte(req.Argv.Message), &reqUserRegister)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Err = apibackend.ErrDataCorrupted
		return
	}

	// userkey
	uuid, err := uuid.NewV4()
	if err != nil {
		l4g.Error("error create user key: %s", err.Error())
		res.Err = apibackend.ErrInternal
		return
	}
	userKey := uuid.String()

	// db
	err = db.Register(&reqUserRegister, userKey)
	if err != nil {
		l4g.Error("error create user: %s", err.Error())
		res.Err = apibackend.ErrInternal
		return
	}

	// to ack
	ackUserCreate := backend.AckUserRegister{}
	ackUserCreate.UserKey = userKey

	dataAck, err := json.Marshal(ackUserCreate)
	if err != nil {
		db.Delete(userKey)
		l4g.Error("error Marshal: %s", err.Error())
		res.Err = apibackend.ErrInternal
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
	reqUserList := struct{
		TotalLines 		int 		`json:"total_lines" doc:"总数,0：表示首次查询"`
		PageIndex 		int 		`json:"page_index" doc:"页索引,1开始"`
		MaxDispLines 	int 		`json:"max_disp_lines" doc:"页最大数，100以下"`

		Condition map[string]interface{}  `json:"condition" doc:"条件查询"`
	}{}
	err := json.Unmarshal([]byte(req.Argv.Message), &reqUserList)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Err = apibackend.ErrDataCorrupted
		return
	}

	var (
		pageNum = reqUserList.MaxDispLines
		totalLine = reqUserList.TotalLines
		pageIndex = reqUserList.PageIndex

		beginIndex = 0
	)

	if totalLine == 0 {
		//totalLine, err = db.ListUserCount()
		totalLine, err = db.ListUserCountByBasic(reqUserList.Condition)
		if err != nil {
			l4g.Error("error json message: %s", err.Error())
			res.Err = apibackend.ErrAccountSrvListUsersCount
			return
		}
	}

	if pageNum < 1 || pageNum > 100 {
		pageNum = 50
	}

	beginIndex = pageNum * (pageIndex-1)

	//ackUserList, err := db.ListUsers(beginIndex, pageNum)
	ackUserList, err := db.ListUsersByBasic(beginIndex, pageNum, reqUserList.Condition)
	if err != nil {
		l4g.Error("error ListUsers: %s", err.Error())
		res.Err = apibackend.ErrAccountSrvListUsers
		return
	}

	ackUserList.PageIndex = pageIndex
	ackUserList.MaxDispLines = pageNum
	ackUserList.TotalLines = totalLine

	// to ack
	dataAck, err := json.Marshal(ackUserList)
	if err != nil {
		l4g.Error("error Marshal: %s", err.Error())
		res.Err = apibackend.ErrInternal
		return
	}

	// ok
	res.Value.Message = string(dataAck)
	l4g.Info("list users: %s", res.Value.Message)
}

// 获取key
func (s * Account) ReadProfile(req *data.SrvRequest, res *data.SrvResponse) {
	// from req
	//reqReadProfile := v1.ReqUserReadProfile{}
	//err := json.Unmarshal([]byte(req.Argv.Message), &reqReadProfile)
	//if err != nil {
	//	l4g.Error("error json message: %s", err.Error())
	//	res.Err = data.ErrDataCorrupted
	//	return
	//}

	// load profile
	ackReadProfile, err := db.ReadProfile(req.Argv.SubUserKey)
	if err != nil {
		l4g.Error("error ReadProfile: %s", err.Error())
		res.Err = apibackend.ErrAccountSrvNoUser
		return
	}

	if ackReadProfile.PublicKey != "" && ackReadProfile.CallbackUrl != "" && ackReadProfile.SourceIP != "" {
		ackReadProfile.ServerPublicKey = string(s.serverPublicKey)
	}

	// to ack
	dataAck, err := json.Marshal(ackReadProfile)
	if err != nil {
		l4g.Error("error Marshal: %s", err.Error())
		res.Err = apibackend.ErrInternal
		return
	}

	// ok
	res.Value.Message = string(dataAck)
	l4g.Info("read a user profile: %s", res.Value.Message)
}

// 更新key
func (s * Account) UpdateProfile(req *data.SrvRequest, res *data.SrvResponse) {
	// from req
	reqUpdateProfile := backend.ReqUserUpdateProfile{}
	err := json.Unmarshal([]byte(req.Argv.Message), &reqUpdateProfile)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Err = apibackend.ErrDataCorrupted
		return
	}

	err = utils.RsaVerifyPubKey([]byte(reqUpdateProfile.PublicKey))
	if err != nil {
		l4g.Error("pub key parse: %s", err.Error())
		res.Err = apibackend.ErrAccountPubKeyParse
		return
	}

	// load old key
	oldUserReadProfile, err := db.ReadProfile(req.Argv.SubUserKey)
	if err != nil {
		l4g.Error("error ReadProfile: %s", err.Error())
		res.Err = apibackend.ErrAccountSrvNoUser
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
	if err := db.UpdateProfile(req.Argv.SubUserKey, &reqUpdateProfile); err != nil {
		l4g.Error("error update profile: %s", err.Error())
		res.Err = apibackend.ErrAccountSrvUpdateProfile
		return
	}

	// to ack
	ackUpdateProfile := backend.AckUserUpdateProfile{
		ServerPublicKey:string(s.serverPublicKey),
	}
	dataAck, err := json.Marshal(ackUpdateProfile)
	if err != nil {
		// 写回去
		oldUserUpdateProfile := backend.ReqUserUpdateProfile{}
		oldUserUpdateProfile.PublicKey = oldUserReadProfile.PublicKey
		oldUserUpdateProfile.SourceIP = oldUserReadProfile.SourceIP
		oldUserUpdateProfile.CallbackUrl = oldUserReadProfile.CallbackUrl
		db.UpdateProfile(req.Argv.SubUserKey, &oldUserUpdateProfile)
		l4g.Error("error Marshal: %s", err.Error())
		res.Err = apibackend.ErrInternal
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
		notifyReq.Argv.UserKey = ""
		notifyReq.Argv.SubUserKey = req.Argv.SubUserKey

		notifyRes := data.SrvResponse{}
		s.node.InnerNotify(&notifyReq, &notifyRes)

		l4g.Info("notify a user profile: %s", req.Argv.SubUserKey)
	}()
}

// 设置冻结
func (s * Account) UpdateFrozen(req *data.SrvRequest, res *data.SrvResponse) {
	// from req
	reqFrozeUser := backend.ReqFrozenUser{}
	err := json.Unmarshal([]byte(req.Argv.Message), &reqFrozeUser)
	if err != nil {
		l4g.Error("error json message: %s", err.Error())
		res.Err = apibackend.ErrDataCorrupted
		return
	}

	// set frozen
	err = db.UpdateFrozen(reqFrozeUser.UserKey, reqFrozeUser.IsFrozen)
	if err != nil {
		l4g.Error("error UpdateFrozen: %s", err.Error())
		res.Err = apibackend.ErrAccountSrvSetFrozen
		return
	}

	// get frozen
	ackFrozenUser := backend.AckFrozenUser{}
	ackFrozenUser.UserKey = reqFrozeUser.UserKey
	ackFrozenUser.IsFrozen, err = db.ReadFrozen(ackFrozenUser.UserKey)
	if err != nil {
		l4g.Error("error UpdateFrozen: %s", err.Error())
		res.Err = apibackend.ErrAccountSrvSetFrozen
		return
	}

	// to ack
	dataAck, err := json.Marshal(ackFrozenUser)
	if err != nil {
		l4g.Error("error Marshal: %s", err.Error())
		res.Err = apibackend.ErrInternal
		return
	}

	// ok
	res.Value.Message = string(dataAck)
	l4g.Info("update a user frozen: %s", res.Value.Message)

	// notify
	func(){
		notifyReq := data.SrvRequest{}
		notifyReq.Method.Version = "v1"
		notifyReq.Method.Srv = "account"
		notifyReq.Method.Function = "updatefrozen"
		notifyReq.Argv.UserKey = ""
		notifyReq.Argv.SubUserKey = ackFrozenUser.UserKey

		notifyRes := data.SrvResponse{}
		s.node.InnerNotify(&notifyReq, &notifyRes)

		l4g.Info("notify a user frozen: %s", req.Argv.SubUserKey)
	}()
}