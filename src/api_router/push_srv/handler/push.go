package handler

import (
	"api_router/push_srv/db"
	"bastionpay_base/data"
	//"api_router/base/service"
	service "bastionpay_base/service2"
	"sync"
	l4g "github.com/alecthomas/log4go"
	"bastionpay_base/nethelper"
	"encoding/json"
	"bastionpay_api/api"
	"bastionpay_api/apibackend"
	"bastionpay_api/apibackend/v1/backend"
	"io/ioutil"
	"bastionpay_base/config"
)

type Push struct{
	privateKey []byte

	rwmu sync.RWMutex
	usersCallbackUrl map[string]*backend.AckUserReadProfile
}

var defaultPush = &Push{}

func PushInstance() *Push{
	return defaultPush
}

func (push * Push)Init(dir string) {
	var err error
	push.privateKey, err = ioutil.ReadFile(dir + "/"+ config.BastionPayPrivateKey)
	if err != nil {
		l4g.Crashf("", err)
	}

	push.usersCallbackUrl = make(map[string]*backend.AckUserReadProfile)
}

func (push * Push)getUserProfile(userKey string) (*backend.AckUserReadProfile, error)  {
	rp := func() *backend.AckUserReadProfile{
		push.rwmu.RLock()
		defer push.rwmu.RUnlock()

		return push.usersCallbackUrl[userKey]
	}()
	if rp != nil {
		return rp, nil
	}

	return func() (*backend.AckUserReadProfile, error){
		push.rwmu.Lock()
		defer push.rwmu.Unlock()

		rp := push.usersCallbackUrl[userKey]
		if rp != nil {
			return rp, nil
		}
		rp, err := db.ReadProfile(userKey)
		if err != nil {
			return nil, err
		}
		push.usersCallbackUrl[userKey] = rp
		return rp, nil
	}()
}

func (push * Push)reloadUserCallbackUrl(userKey string) (*backend.AckUserReadProfile, error)  {
	push.rwmu.Lock()
	defer push.rwmu.Unlock()

	rp, err := db.ReadProfile(userKey)
	if err != nil {
		return nil, err
	}
	push.usersCallbackUrl[userKey] = rp
	return rp, nil
}

func (push *Push)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	func(){
		service.RegisterApi(&nam,
			"pushdata", data.APILevel_client, push.PushData)
	}()

	return nam
}

func (push *Push)HandleNotify(req *data.SrvRequest){
	if req.Method.Srv == "account" && req.Method.Function == "updateprofile" {
		//reqUpdateProfile := v1.ReqUserUpdateProfile{}
		//err := json.Unmarshal([]byte(req.Argv.Message), &reqUpdateProfile)
		//if err != nil {
		//	l4g.Error("HandleNotify-Unmarshal: %s", err.Error())
		//	return
		//}

		// reload profile
		rp, err := push.reloadUserCallbackUrl(req.Argv.SubUserKey)
		if err != nil {
			l4g.Error("HandleNotify-reloadUserCallbackUrl: %s", err.Error())
			return
		}

		l4g.Info("HandleNotify-reloadUserCallbackUrl: ", rp)
	}
}

// 推送数据
func (push *Push)PushData(req *data.SrvRequest, res *data.SrvResponse) {
	rp, err := push.getUserProfile(req.Argv.UserKey)
	if err != nil {
		l4g.Error("(%s) no user callback: %s",req.Argv.UserKey, err.Error())
		res.Err = apibackend.ErrPushSrvPushData
		return
	}

	l4g.Info("push %s to %s-%s", req.Argv.Message, req.Argv.UserKey, rp.CallbackUrl)

	func(){
		// encrypt
		srvData, err := data.EncryptionAndSignData([]byte(req.Argv.Message), req.Argv.UserKey, []byte(rp.PublicKey), push.privateKey)
		if err != nil {
			l4g.Error("EncryptionAndSignData: %s", err.Error())
			res.Err = apibackend.ErrInternal
			return
		}

		pushData := api.UserResponseData{}
		srvData.ToApiData(&pushData.Value)

		// call url
		b, err := json.Marshal(pushData)
		if err != nil {
			l4g.Error("error json message: %s", err.Error())
			res.Err = apibackend.ErrDataCorrupted
			return
		}

		l4g.Info("push data: %s", string(b))

		httpCode, ret, err := nethelper.CallToHttpServer(rp.CallbackUrl, "", string(b))
		if err != nil {
			l4g.Error("push http: %s", err.Error())
			res.Err = apibackend.ErrPushSrvPushData
			return
		}
		res.Value.Message = ret

		l4g.Info("push status:%d-%s", httpCode, ret)
	}()

	l4g.Info("push fin to %s", req.Argv.UserKey)
}