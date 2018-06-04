package handler

import (
	"api_router/push_srv/db"
	"api_router/base/data"
	//"api_router/base/service"
	service "api_router/base/service2"
	"sync"
	l4g "github.com/alecthomas/log4go"
	"api_router/base/nethelper"
	"encoding/json"
	"bastionpay_api/api"
	"bastionpay_api/apibackend"
)

type Push struct{
	rwmu sync.RWMutex
	usersCallbackUrl map[string]string
}

var defaultPush = &Push{}

func PushInstance() *Push{
	return defaultPush
}

func (push * Push)Init() {
	push.usersCallbackUrl = make(map[string]string)
}

func (push * Push)getUserCallbackUrl(userKey string) (string, error)  {
	url := func() string{
		push.rwmu.RLock()
		defer push.rwmu.RUnlock()

		return push.usersCallbackUrl[userKey]
	}()
	if url != "" {
		return url,nil
	}

	return func() (string, error){
		push.rwmu.Lock()
		defer push.rwmu.Unlock()

		url := push.usersCallbackUrl[userKey]
		if url != "" {
			return url, nil
		}
		url, err := db.ReadUserCallbackUrl(userKey)
		if err != nil {
			return "", err
		}
		push.usersCallbackUrl[userKey] = url
		return url, nil
	}()
}

func (push * Push)reloadUserCallbackUrl(userKey string) (string, error)  {
	push.rwmu.Lock()
	defer push.rwmu.Unlock()

	url, err := db.ReadUserCallbackUrl(userKey)
	if err != nil {
		return "", err
	}
	push.usersCallbackUrl[userKey] = url
	return url, nil
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
		ndata, err := push.reloadUserCallbackUrl(req.Argv.SubUserKey)
		if err != nil {
			l4g.Error("HandleNotify-reloadUserCallbackUrl: %s", err.Error())
			return
		}

		l4g.Info("HandleNotify-reloadUserCallbackUrl: ", ndata)
	}
}

// 推送数据
func (push *Push)PushData(req *data.SrvRequest, res *data.SrvResponse) {
	url, err := push.getUserCallbackUrl(req.Argv.UserKey)
	if err != nil {
		l4g.Error("(%s) no user callback: %s",req.Argv.UserKey, err.Error())
		res.Err = apibackend.ErrPushSrvPushData
		return
	}

	l4g.Info("push %s to %s-%s", req.Argv.Message, req.Argv.UserKey, url)

	func(){
		pushData := api.UserResponseData{}

		req.Argv.ToApiData(&pushData.Value)

		// call url
		b, err := json.Marshal(pushData)
		if err != nil {
			l4g.Error("error json message: %s", err.Error())
			res.Err = apibackend.ErrDataCorrupted
			return
		}

		l4g.Info("push data: %s", string(b))

		httpCode, ret, err := nethelper.CallToHttpServer(url, "", string(b))
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