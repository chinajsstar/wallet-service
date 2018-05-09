package service2

import (
	"sync"
	"api_router/base/data"
	"api_router/base/nethelper"
	"api_router/base/config"
	"encoding/json"
	"context"
	"net/http"
	"io/ioutil"
	"strings"
	l4g "github.com/alecthomas/log4go"
	"github.com/cenkalti/rpc2"
)

const (
	HttpApi = "api"
	HttpApiTest = "apitest"
)

type ServiceGateway struct{
	// rpc2
	*rpc2.Server

	// config
	cfgGateway config.ConfigGateway

	// srv nodes
	rwmu                       sync.RWMutex
	srvNodeNameMapSrvNodeGroup map[string]*SrvNodeGroup // name+version mapto srvnodegroup
	clientMapSrvNodeGroup 	   map[*rpc2.Client]*SrvNodeGroup // name+version mapto srvnodegroup
	ApiInfo                    map[string]*data.ApiInfo

	// wait group
	wg sync.WaitGroup

	// center's apis
	registerData data.SrvRegisterData
	apiHandler map[string]*NodeApi
}

// new a center
func NewServiceGateway(confPath string) (*ServiceGateway, error){
	serviceGateway := &ServiceGateway{}

	serviceGateway.cfgGateway.Load(confPath)

	serviceGateway.srvNodeNameMapSrvNodeGroup = make(map[string]*SrvNodeGroup)
	serviceGateway.clientMapSrvNodeGroup = make(map[*rpc2.Client]*SrvNodeGroup)

	serviceGateway.registerData.Srv = serviceGateway.cfgGateway.GatewayName
	serviceGateway.registerData.Version = serviceGateway.cfgGateway.GatewayVersion

	serviceGateway.apiHandler = make(map[string]*NodeApi)

	func(){
		// api listsrv
		apiInfo := data.ApiInfo{Name:"listsrv", Level:data.APILevel_admin}

		example := "{}"
		icomment := "无参数"

		var oargv []data.SrvRegisterData
		oargv = append(oargv, data.SrvRegisterData{Version:"v1", Srv:"srv"})
		ocomment := data.FieldTag(oargv)

		apiDoc := data.ApiDoc{Name:apiInfo.Name, Level:apiInfo.Level, Doc:"列出所有服务", Example:example, InComment:icomment, OutComment:ocomment}
		serviceGateway.apiHandler[apiInfo.Name] = &NodeApi{ApiHandler:serviceGateway.listSrv, ApiInfo:apiInfo, ApiDoc:apiDoc}
		serviceGateway.registerData.Functions = append(serviceGateway.registerData.Functions, apiInfo)
		serviceGateway.registerData.ApiDocs = append(serviceGateway.registerData.ApiDocs, apiDoc)
	}()

	// register
	var res string
	serviceGateway.register(nil, &serviceGateway.registerData, &res)

	// rpc2
	serviceGateway.Server = rpc2.NewServer()

	return serviceGateway, nil
}

// start the service center
func StartCenter(ctx context.Context, mi *ServiceGateway) {
	mi.startHttpServer(ctx)

	mi.startTcpServer(ctx)
}

// Stop the service center
func StopCenter(mi *ServiceGateway)  {
	mi.wg.Wait()
}

func (mi *ServiceGateway) register(client *rpc2.Client, reg *data.SrvRegisterData, res *string) error {
	err := func()error {
		mi.rwmu.Lock()
		defer mi.rwmu.Unlock()

		versionSrvName := strings.ToLower(reg.Srv + "." + reg.Version)
		srvNodeGroup := mi.srvNodeNameMapSrvNodeGroup[versionSrvName]
		if srvNodeGroup == nil {
			srvNodeGroup = &SrvNodeGroup{}
			mi.srvNodeNameMapSrvNodeGroup[versionSrvName] = srvNodeGroup
		}

		mi.clientMapSrvNodeGroup[client] = srvNodeGroup

		err := srvNodeGroup.RegisterNode(client, reg)
		if err == nil {
			if mi.ApiInfo == nil {
				mi.ApiInfo = make(map[string]*data.ApiInfo)
			}

			for _, v := range reg.Functions{
				mi.ApiInfo[strings.ToLower(versionSrvName+"."+v.Name)] = &data.ApiInfo{v.Name, v.Level}
			}
		}

		l4g.Info("%d-%d", len(mi.srvNodeNameMapSrvNodeGroup), len(mi.clientMapSrvNodeGroup))

		*res = "ok"
		return err
	}()

	return err
}

func (mi *ServiceGateway) unRegister(client *rpc2.Client, reg *data.SrvRegisterData, res *string) error {
	mi.disconnectClient(client)
	*res = "ok"
	return nil
}

func (mi *ServiceGateway) innerCall(client *rpc2.Client, req *data.UserRequestData, res *data.UserResponseData) error {
	mi.wg.Add(1)
	defer mi.wg.Done()

	mi.srvCall(req, res)

	// make sure no data if err
	if res.Err != data.NoErr {
		res.Value.Message = ""
		res.Value.Signature = ""
	}
	return nil
}

func (mi *ServiceGateway) innerCallByEncrypt(client *rpc2.Client, req *data.UserRequestData, res *data.UserResponseData) error {
	mi.wg.Add(1)
	defer mi.wg.Done()

	mi.srvCallByEncrypt(req, res)

	// make sure no data if err
	if res.Err != data.NoErr {
		res.Value.Message = ""
		res.Value.Signature = ""
	}
	return nil
}

// start http server
func (mi *ServiceGateway) startHttpServer(ctx context.Context) {
	// http
	l4g.Debug("Start http server on %s", mi.cfgGateway.Port)

	http.Handle("/" + HttpApi + "/", http.HandlerFunc(mi.handleApi))

	// test mode
	if mi.cfgGateway.TestMode != 0 {
		http.Handle("/" + HttpApiTest + "/", http.HandlerFunc(mi.handleApiTest))
	}

	go func() {
		l4g.Info("Http server routine running... ")
		err := http.ListenAndServe(":"+mi.cfgGateway.Port, nil)
		if err != nil {
			l4g.Crashf("", err)
		}
	}()
}

func (mi *ServiceGateway)disconnectClient(client *rpc2.Client)  {
	mi.rwmu.Lock()
	defer mi.rwmu.Unlock()

	srvNodeGroup, ok := mi.clientMapSrvNodeGroup[client]
	if srvNodeGroup == nil || !ok {
		return
	}

	srvNodeGroup.UnRegisterNode(client)
	if srvNodeGroup.GetSrvNodes() == 0 {
		// remove srv node group
		reg := srvNodeGroup.GetSrvInfo()
		versionSrvName := strings.ToLower(reg.Srv + "." + reg.Version)
		if mi.ApiInfo != nil {
			for _, v := range reg.Functions{
				delete(mi.ApiInfo, strings.ToLower(versionSrvName+"."+v.Name))
			}
		}

		delete(mi.srvNodeNameMapSrvNodeGroup, versionSrvName)
	}

	delete(mi.clientMapSrvNodeGroup, client)
}

// start tcp server
func (mi *ServiceGateway) startTcpServer(ctx context.Context) {
	mi.Server.OnConnect(func(client *rpc2.Client) {
		l4g.Info("rpc2 client connect...")
	})

	mi.Server.OnDisconnect(func(client *rpc2.Client) {
		l4g.Info("rpc2 client disconnect...")

		mi.disconnectClient(client)
	})

	mi.Server.Handle(data.MethodCenterRegister, mi.register)
	mi.Server.Handle(data.MethodCenterUnRegister, mi.unRegister)
	mi.Server.Handle(data.MethodCenterInnerCall, mi.innerCall)
	mi.Server.Handle(data.MethodCenterInnerCallByEncrypt, mi.innerCallByEncrypt)

	l4g.Debug("Start Tcp server on %s", mi.cfgGateway.GatewayPort)

	listener, err := nethelper.CreateTcpServer(":"+mi.cfgGateway.GatewayPort)
	if err != nil {
		l4g.Crashf("", err)
	}
	go func() {
		mi.wg.Add(1)
		defer mi.wg.Done()

		l4g.Info("Tcp server routine running... ")

		go mi.Server.Accept(listener)
		<- ctx.Done()

		l4g.Info("Tcp server routine stoped... ")
	}()
}

func (mi *ServiceGateway) srvCall(req *data.UserRequestData, res *data.UserResponseData) {
	api := mi.getApiInfo(req)
	if api == nil {
		res.Err = data.ErrNotFindSrv
		l4g.Error("%s %d", req.String(), res.Err)
		return
	}

	// call function
	var rpcSrv data.SrvRequestData
	rpcSrv.Data = *req
	rpcSrv.Context.ApiLever = api.Level
	var rpcSrvRes data.SrvResponseData
	if mi.callFunction(&rpcSrv, &rpcSrvRes); rpcSrvRes.Data.Err != data.NoErr{
		*res = rpcSrvRes.Data
		return
	}

	*res = rpcSrvRes.Data
}

func (mi *ServiceGateway) srvCallByEncrypt(req *data.UserRequestData, res *data.UserResponseData) {
	// encode and sign data
	var reqEncrypted data.SrvRequestData
	reqEncrypted.Data.Method = req.Method
	reqEncrypted.Data.Argv = req.Argv
	var reqEncryptedRes data.SrvResponseData
	if mi.encryptData(&reqEncrypted, &reqEncryptedRes); reqEncryptedRes.Data.Err != data.NoErr{
		*res = reqEncryptedRes.Data
		return
	}

	// push encode and sign data
	var reqPush data.SrvRequestData
	reqPush.Data.Method = req.Method
	reqPush.Data.Argv = req.Argv
	reqPush.Data.Argv.Message = reqEncryptedRes.Data.Value.Message
	reqPush.Data.Argv.Signature = reqEncryptedRes.Data.Value.Signature
	var reqPushRes data.SrvResponseData

	mi.callFunction(&reqPush, &reqPushRes)
	*res = reqPushRes.Data
}

// user call by user
func (mi *ServiceGateway) userCall(req *data.UserRequestData, res *data.UserResponseData) {
	// can not call auth service
	if req.Method.Srv == "auth" {
		res.Err = data.ErrIllegallyCall
		l4g.Error("%s %d", req.String(), res.Err)
		return
	}

	// find api
	api := mi.getApiInfo(req)
	if api == nil {
		res.Err = data.ErrNotFindSrv
		l4g.Error("%s %d", req.String(), res.Err)
		return
	}

	// decode and verify data
	var rpcAuth data.SrvRequestData
	rpcAuth.Data = *req
	rpcAuth.Context.ApiLever = api.Level
	var rpcAuthRes data.SrvResponseData
	if mi.authData(&rpcAuth, &rpcAuthRes); rpcAuthRes.Data.Err != data.NoErr{
		*res = rpcAuthRes.Data
		return
	}

	// call real srv
	var rpcSrv data.SrvRequestData
	rpcSrv.Data = *req
	rpcSrv.Context.ApiLever = api.Level
	rpcSrv.Data.Argv.Message = rpcAuthRes.Data.Value.Message
	var rpcSrvRes data.SrvResponseData
	if mi.callFunction(&rpcSrv, &rpcSrvRes); rpcSrvRes.Data.Err != data.NoErr{
		*res = rpcSrvRes.Data
		return
	}

	// encode and sign data
	var reqEncrypted data.SrvRequestData
	reqEncrypted.Data = *req
	reqEncrypted.Context.ApiLever = api.Level
	reqEncrypted.Data.Argv.Message = rpcSrvRes.Data.Value.Message
	var reqEncryptedRes data.SrvResponseData
	if mi.encryptData(&reqEncrypted, &reqEncryptedRes); reqEncryptedRes.Data.Err != data.NoErr{
		*res = reqEncryptedRes.Data
		return
	}

	*res = reqEncryptedRes.Data
}

// auth data
func (mi *ServiceGateway) authData(req *data.SrvRequestData, res *data.SrvResponseData) {
	reqAuth := *req
	reqAuth.Data.Method.Srv = "auth"
	reqAuth.Data.Method.Function = "AuthData"
	reqAuthRes := data.SrvResponseData{}

	mi.callFunction(&reqAuth, &reqAuthRes)

	*res = reqAuthRes
}

// package data
func (mi *ServiceGateway) encryptData(req *data.SrvRequestData, res *data.SrvResponseData) {
	reqEnc := *req
	reqEnc.Data.Method.Srv = "auth"
	reqEnc.Data.Method.Function = "EncryptData"

	reqEncRes := data.SrvResponseData{}

	mi.callFunction(&reqEnc, &reqEncRes)

	*res = reqEncRes
}

//  call a srv node
func (mi *ServiceGateway) callFunction(req *data.SrvRequestData, res *data.SrvResponseData) {
	centerVersionSrvName := strings.ToLower(mi.registerData.Srv + "." + mi.registerData.Version)
	versionSrvName := strings.ToLower(req.Data.Method.Srv + "." + req.Data.Method.Version)
	l4g.Debug("call %s", req.Data.String())

	mi.rwmu.RLock()
	defer mi.rwmu.RUnlock()

	if centerVersionSrvName == versionSrvName {
		h := mi.apiHandler[strings.ToLower(req.Data.Method.Function)]
		if h != nil {
			h.ApiHandler(req, res)
		}else{
			res.Data.Err = data.ErrNotFindFunction
		}
		if res.Data.Err != data.NoErr {
			l4g.Error("call failed: %d", res.Data.Err)
		}
		return
	}else{
		srvNodeGroup := mi.srvNodeNameMapSrvNodeGroup[versionSrvName]
		if srvNodeGroup == nil{
			res.Data.Err = data.ErrNotFindSrv
			l4g.Error("%s %d", req.Data.String(), res.Data.Err)
			return
		}

		srvNodeGroup.Call(req, res)
	}
}

func (mi *ServiceGateway) getApiInfo(req *data.UserRequestData) (*data.ApiInfo) {
	mi.rwmu.RLock()
	defer mi.rwmu.RUnlock()

	name := strings.ToLower(req.Method.Srv + "." + req.Method.Version + "." + req.Method.Function)
	return mi.ApiInfo[name]
}

func (mi *ServiceGateway) handleApi(w http.ResponseWriter, req *http.Request) {
	l4g.Debug("Http server Accept a client: %s", req.RemoteAddr)
	//defer req.Body.Close()

	//w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	//w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型

	mi.wg.Add(1)
	defer mi.wg.Done()

	resData := data.UserResponseData{}
	func (){
		//fmt.Println("path=", req.URL.Path)

		path := req.URL.Path
		path = strings.Replace(path, HttpApi, "", -1)
		path = strings.TrimLeft(path, "/")
		path = strings.TrimRight(path, "/")

		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			l4g.Error("http handler: %s", err.Error())
			resData.Err = data.ErrDataCorrupted
			return
		}

		//body := string(b)
		//fmt.Println("body=", body)

		// make data
		reqData := data.UserRequestData{}
		err = json.Unmarshal(b, &reqData.Argv);
		if err != nil {
			l4g.Error("http handler: %s", err.Error())
			resData.Err = data.ErrDataCorrupted
			return
		}

		// get method
		paths := strings.Split(path, "/")
		for i := 0; i < len(paths); i++ {
			if i == 0 {
				reqData.Method.Version = paths[i]
			}else if i == 1{
				reqData.Method.Srv = paths[i]
			} else{
				if reqData.Method.Function != "" {
					reqData.Method.Function += "."
				}
				reqData.Method.Function += paths[i]
			}
		}

		mi.userCall(&reqData, &resData)
		resData.Method = reqData.Method
	}()

	if resData.Err != data.NoErr && resData.ErrMsg == "" {
		resData.ErrMsg = data.GetErrMsg(resData.Err)
	}

	// write back http
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(resData)
	w.Write(b)
	return
}

func (mi *ServiceGateway) handleApiTest(w http.ResponseWriter, req *http.Request) {
	l4g.Debug("Http server test Accept a client: %s", req.RemoteAddr)
	//defer req.Body.Close()

	//w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	//w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型

	mi.wg.Add(1)
	defer mi.wg.Done()

	resData := data.UserResponseData{}
	func (){
		//fmt.Println("path=", req.URL.Path)

		path := req.URL.Path
		path = strings.Replace(path, HttpApiTest, "", -1)
		path = strings.TrimLeft(path, "/")
		path = strings.TrimRight(path, "/")

		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			l4g.Error("http handler: %s", err.Error())
			resData.Err = data.ErrDataCorrupted
			return
		}

		//body := string(b)
		//fmt.Println("body=", body)

		// make data
		reqData := data.UserRequestData{}
		err = json.Unmarshal(b, &reqData.Argv);
		if err != nil {
			l4g.Error("http handler: %s", err.Error())
			resData.Err = data.ErrDataCorrupted
			return
		}

		// get method
		paths := strings.Split(path, "/")
		for i := 0; i < len(paths); i++ {
			if i == 0 {
				reqData.Method.Version = paths[i]
			}else if i == 1{
				reqData.Method.Srv = paths[i]
			} else{
				if reqData.Method.Function != "" {
					reqData.Method.Function += "."
				}
				reqData.Method.Function += paths[i]
			}
		}

		mi.srvCall(&reqData, &resData)
		resData.Method = reqData.Method
	}()

	if resData.Err != data.NoErr && resData.ErrMsg == "" {
		resData.ErrMsg = data.GetErrMsg(resData.Err)
	}

	// write back http
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(resData)
	w.Write(b)
	return
}

// RPC -- listsrv
func (mi *ServiceGateway) listSrv(req *data.SrvRequestData, res *data.SrvResponseData) {
	mi.rwmu.RLock()
	defer mi.rwmu.RUnlock()

	var nodes []data.SrvRegisterData
	for _, v := range mi.srvNodeNameMapSrvNodeGroup{
		node := v.GetSrvInfo()
		nodes = append(nodes, node)
	}

	b, err := json.Marshal(nodes)
	if err != nil {
		res.Data.Err = data.ErrDataCorrupted
		res.Data.Value.Message = ""
		res.Data.Value.Signature = ""
		return
	}
	res.Data.Value.Message = string(b)

	// make sure no data if err
	if res.Data.Err != data.NoErr {
		res.Data.Value.Message = ""
		res.Data.Value.Signature = ""
	}
}