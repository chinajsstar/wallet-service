package handler

import (
	"../db"
	"api_router/base/data"
	"api_router/base/service"
	"sync"
	l4g "github.com/alecthomas/log4go"
	"api_router/base/nethelper"
	"encoding/json"
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

func (push * Push)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	apiInfo := data.ApiInfo{Name:"pushdata", Level:data.APILevel_client}
	nam[apiInfo.Name] = service.NodeApi{ApiHandler:push.PushData, ApiInfo:apiInfo}

	return nam
}

// 推送数据
func (push *Push)PushData(req *data.SrvRequestData, res *data.SrvResponseData) {
	err := func() error{
		url, err := push.getUserCallbackUrl(req.Data.Argv.UserKey)
		if err != nil {
			l4g.Error("(%s) failed: %s",req.Data.Argv.UserKey, err.Error())
			return err
		}

		// call url
		b, err := json.Marshal(req.Data.Argv)
		if err != nil {
			l4g.Error("(%s) marshal failed: %s",req.Data.Argv.UserKey, err.Error())
			return err
		}
		var ret string
		err = nethelper.CallToHttpServer(url, "", string(b), &ret)
		if err != nil {
			l4g.Error("%s", err.Error())
			return err
		}

		res.Data.Value.Message = ret
		res.Data.Value.Signature = ""
		res.Data.Value.UserKey = req.Data.Argv.UserKey

		return nil
	}()

	if err != nil {
		res.Data.Err = data.ErrPushSrvPushData
		res.Data.ErrMsg = data.ErrPushSrvPushDataText
	}
}