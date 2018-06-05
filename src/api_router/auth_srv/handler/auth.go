package handler

import (
	"api_router/auth_srv/db"
	"bastionpay_base/data"
	//"api_router/base/service"
	service "bastionpay_base/service2"
	"io/ioutil"
	"sync"
	l4g "github.com/alecthomas/log4go"
	"bastionpay_base/config"
	"bastionpay_api/apibackend"
)

type Auth struct{
	node *service.ServiceNode

	privateKey []byte

	rwmu     sync.RWMutex
	usersLevel map[string]*db.UserLevel
}

var defaultAuth = &Auth{}

func AuthInstance() *Auth{
	return defaultAuth
}

func (auth * Auth)Init(dir string, node *service.ServiceNode) {
	var err error
	auth.privateKey, err = ioutil.ReadFile(dir + "/"+ config.BastionPayPrivateKey)
	if err != nil {
		l4g.Crashf("", err)
	}

	auth.node = node
	auth.usersLevel = make(map[string]*db.UserLevel)
}

func (auth * Auth)getUserLevel(userKey string) (*db.UserLevel, error)  {
	ul := func() *db.UserLevel{
		auth.rwmu.RLock()
		defer auth.rwmu.RUnlock()

		return auth.usersLevel[userKey]
	}()
	if ul != nil {
		return ul,nil
	}

	return func() (*db.UserLevel, error){
		auth.rwmu.Lock()
		defer auth.rwmu.Unlock()

		ul := auth.usersLevel[userKey]
		if ul != nil {
			return ul, nil
		}
		ul, err := db.ReadUserLevel(userKey)
		if err != nil {
			return nil, err
		}
		auth.usersLevel[userKey] = ul
		return ul, nil
	}()
}

func (auth * Auth)reloadUserLevel(userKey string) (*db.UserLevel, error)  {
	auth.rwmu.Lock()
	defer auth.rwmu.Unlock()

	ul, err := db.ReadUserLevel(userKey)
	if err != nil {
		return nil, err
	}
	auth.usersLevel[userKey] = ul
	return ul, nil
}

func (auth * Auth)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	func(){
		service.RegisterApi(&nam,
			"authdata", data.APILevel_client, auth.AuthData)
	}()

	func(){
		service.RegisterApi(&nam,
			"encryptdata", data.APILevel_client, auth.EncryptData)
	}()

	return nam
}

func (auth * Auth)HandleNotify(req *data.SrvRequest){
	if req.Method.Srv == "account" {
		//reqUpdateProfile := v1.ReqUserUpdateProfile{}
		//err := json.Unmarshal([]byte(req.Argv.Message), &reqUpdateProfile)
		//if err != nil {
		//	l4g.Error("HandleNotify-Unmarshal: %s", err.Error())
		//	return
		//}
		if req.Method.Function == "updateprofile" || req.Method.Function == "updatefrozen" {
			// reload profile
			ndata, err := auth.reloadUserLevel(req.Argv.SubUserKey)
			if err != nil {
				l4g.Error("HandleNotify-reloadUserLevel: %s", err.Error())
				return
			}

			l4g.Info("HandleNotify-reloadUserLevel: ", req.Argv.SubUserKey, ndata)
		}
	}
}

// 验证数据
func (auth *Auth)AuthData(req *data.SrvRequest, res *data.SrvResponse) {
	ul, err := auth.getUserLevel(req.Argv.UserKey)
	if err != nil {
		l4g.Error("(%s) get user level failed: %s",req.Argv.UserKey, err.Error())
		res.Err = apibackend.ErrAuthSrvNoUserKey
		return
	}

	if ul.PublicKey == "" {
		l4g.Error("(%s-%s) failed: no public key", req.Argv.UserKey, req.Method.Function)
		res.Err = apibackend.ErrAuthSrvNoPublicKey
		return
	}

	if req.Context.ApiLever > ul.Level{
		l4g.Error("(%s-%s) failed: no api level", req.Argv.UserKey, req.Method.Function)
		res.Err = apibackend.ErrAuthSrvNoApiLevel
		return
	}

	if req.Context.ApiLever > ul.Level || ul.IsFrozen != 0{
		l4g.Error("(%s-%s) failed: user frozen", req.Argv.UserKey, req.Method.Function)
		res.Err = apibackend.ErrAuthSrvUserFrozen
		return
	}

	if(req.Context.DataFrom == apibackend.DataFromUser || req.Context.DataFrom == apibackend.DataFromAdmin){
		if ul.UserClass != data.UserClass_Admin {
			l4g.Error("%s illegally call data type", req.Argv.UserKey)
			res.Err = apibackend.ErrAuthSrvIllegalDataType
			return
		}
	}

	originData, err := data.DecryptionAndVerifyData(&req.Argv, []byte(ul.PublicKey), auth.privateKey)
	if err != nil {
		l4g.Error("DecryptionAndVerifyData: %s", err.Error())
		res.Err = apibackend.ErrAuthSrvIllegalData
		return
	}

	res.Value.Message = string(originData)
}

// 打包数据
func (auth *Auth)EncryptData(req *data.SrvRequest, res *data.SrvResponse) {
	ul, err := auth.getUserLevel(req.Argv.UserKey)
	if err != nil {
		l4g.Error("(%s) get user level failed: %s",req.Argv.UserKey, err.Error())
		res.Err = apibackend.ErrAuthSrvNoUserKey
		return
	}

	if ul.PublicKey == "" {
		l4g.Error("(%s-%s) failed: no public key", req.Argv.UserKey, req.Method.Function)
		res.Err = apibackend.ErrAuthSrvNoPublicKey
		return
	}

	srvData, err := data.EncryptionAndSignData([]byte(req.Argv.Message), req.Argv.UserKey, []byte(ul.PublicKey), auth.privateKey)
	if err != nil {
		l4g.Error("EncryptionAndSignData: %s", err.Error())
		res.Err = apibackend.ErrInternal
		return
	}

	// ok
	res.Value.Message = srvData.Message
	res.Value.Signature = srvData.Signature
}