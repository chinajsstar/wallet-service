package service

import (
	"sync"
	"net/rpc"
	"api_router/base/data"
	"api_router/base/nethelper"
	"api_router/base/config"
	"encoding/json"
	"context"
	"net/http"
	"io/ioutil"
	"strings"
	"golang.org/x/net/websocket"
	"errors"
	l4g "github.com/alecthomas/log4go"
)

type WsClient struct{
	// ws client license key
	licenseKey string
}

type ServiceCenter struct{
	// config
	cfgCenter config.ConfigCenter

	// srv nodes
	rwmu                       sync.RWMutex
	SrvNodeNameMapSrvNodeGroup map[string]*SrvNodeGroup // name+version mapto srvnodegroup
	ApiInfo                    map[string]*data.ApiInfo

	// websocket
	rwmuws sync.RWMutex
	// valid clients
	licenseKey2wsClients map[string]*websocket.Conn
	wsClients map[*websocket.Conn]*WsClient

	// wait group
	wg sync.WaitGroup

	// connection group
	clientGroup *ConnectionGroup

	// center's apis
	registerData data.SrvRegisterData
	apiHandler map[string]*NodeApi
}

// new a center
func NewServiceCenter(confPath string) (*ServiceCenter, error){
	serviceCenter := &ServiceCenter{}
	serviceCenter.cfgCenter.Load(confPath)

	serviceCenter.SrvNodeNameMapSrvNodeGroup = make(map[string]*SrvNodeGroup)

	serviceCenter.wsClients = make(map[*websocket.Conn]*WsClient)
	serviceCenter.licenseKey2wsClients = make(map[string]*websocket.Conn)

	serviceCenter.clientGroup = NewConnectionGroup()

	serviceCenter.registerData.Srv = serviceCenter.cfgCenter.CenterName
	serviceCenter.registerData.Version = serviceCenter.cfgCenter.CenterVersion
	serviceCenter.registerData.Addr = ""

	serviceCenter.apiHandler = make(map[string]*NodeApi)
	// api listsrv
	apiInfo := data.ApiInfo{Name:"listsrv", Level:data.APILevel_admin}
	apiInfo.Example = "none"
	serviceCenter.apiHandler[apiInfo.Name] = &NodeApi{ApiHandler:serviceCenter.listSrv, ApiInfo:apiInfo}
	serviceCenter.registerData.Functions = append(serviceCenter.registerData.Functions, apiInfo)

	// register
	var res string
	serviceCenter.Register(&serviceCenter.registerData, &res)

	return serviceCenter, nil
}

// start the service center
func StartCenter(ctx context.Context, mi *ServiceCenter) {
	mi.startHttpServer(ctx)

	//mi.startWsServer(ctx)

	mi.startTcpServer(ctx)
}

// Stop the service center
func StopCenter(mi *ServiceCenter)  {
	mi.wg.Wait()
}

// RPC -- register
func (mi *ServiceCenter) Register(reg *data.SrvRegisterData, res *string) error {
	err := func()error {
		mi.rwmu.Lock()
		defer mi.rwmu.Unlock()

		versionSrvName := strings.ToLower(reg.Srv + "." + reg.Version)
		srvNodeGroup := mi.SrvNodeNameMapSrvNodeGroup[versionSrvName]
		if srvNodeGroup == nil {
			srvNodeGroup = &SrvNodeGroup{}
			mi.SrvNodeNameMapSrvNodeGroup[versionSrvName] = srvNodeGroup
		}

		err := srvNodeGroup.RegisterNode(reg)
		if err == nil {
			if mi.ApiInfo == nil {
				mi.ApiInfo = make(map[string]*data.ApiInfo)
			}

			for _, v := range reg.Functions{
				mi.ApiInfo[strings.ToLower(versionSrvName+"."+v.Name)] = &data.ApiInfo{v.Name, v.Level, v.Example}
			}
		}

		return err
	}()

	return err
}

// RPC -- unregister
func (mi *ServiceCenter) UnRegister(reg *data.SrvRegisterData, res *string) error {
	err := func() error {
		mi.rwmu.Lock()
		defer mi.rwmu.Unlock()

		versionSrvName := strings.ToLower(reg.Srv + "." + reg.Version)
		srvNodeGroup := mi.SrvNodeNameMapSrvNodeGroup[versionSrvName]
		if srvNodeGroup == nil {
			return nil
		}

		err := srvNodeGroup.UnRegisterNode(reg)
		if err == nil {
			if mi.ApiInfo != nil {
				for _, v := range reg.Functions{
					delete(mi.ApiInfo, strings.ToLower(versionSrvName+"."+v.Name))
				}
			}
		}
		return err
	}()

	return err
}

// RPC -- dispatch
func (mi *ServiceCenter) Dispatch(req *data.UserRequestData, res *data.UserResponseData) error {
	mi.wg.Add(1)
	defer mi.wg.Done()

	mi.innerCall(req, res)

	// make sure no data if err
	if res.Err != data.NoErr {
		res.Value.Message = ""
		res.Value.Signature = ""
	}
	return nil
}

// RPC -- push
func (mi *ServiceCenter) Push(req *data.UserRequestData, res *data.UserResponseData) error {
	mi.wg.Add(1)
	defer mi.wg.Done()

	mi.pushCall(req, res)

	// make sure no data if err
	if res.Err != data.NoErr {
		res.Value.Message = ""
		res.Value.Signature = ""
	}
	return nil
}

// start http server
func (mi *ServiceCenter) startHttpServer(ctx context.Context) {
	// http
	l4g.Debug("Start http server on %s", mi.cfgCenter.Port)

	http.Handle("/wallet/", http.HandlerFunc(mi.handleWallet))

	go func() {
		l4g.Info("Http server routine running... ")
		err := http.ListenAndServe(":"+mi.cfgCenter.Port, nil)
		if err != nil {
			l4g.Crashf("", err)
		}
	}()
}

// start websocket server
func (mi *ServiceCenter) startWsServer(ctx context.Context) {
	// websocket
	l4g.Debug("Start ws server on %s", mi.cfgCenter.WsPort)

	http.Handle("/ws", websocket.Handler(mi.handleWebSocket))

	go func() {
		l4g.Info("ws server routine running... ")
		err := http.ListenAndServe(":"+mi.cfgCenter.WsPort, nil)
		if err != nil {
			l4g.Crashf("", err)
		}
	}()
}

// start tcp server
func (mi *ServiceCenter) startTcpServer(ctx context.Context) {
	l4g.Debug("Start Tcp server on ", mi.cfgCenter.CenterPort)

	listener, err := nethelper.CreateTcpServer(":"+mi.cfgCenter.CenterPort)
	if err != nil {
		l4g.Crashf("", err)
	}
	go func() {
		mi.wg.Add(1)
		defer mi.wg.Done()

		l4g.Info("Tcp server routine running... ")
		go func(){
			for{
				conn, err := listener.Accept();
				if err != nil {
					l4g.Error("%s", err.Error())
					continue
				}

				l4g.Info("Tcp server Accept a client: %s", conn.RemoteAddr().String())
				rc := mi.clientGroup.Register(conn)

				go func() {
					go rpc.ServeConn(rc)
					<- rc.Done
					l4g.Info("Tcp server close a client: %s", conn.RemoteAddr().String())
				}()
			}
		}()

		<- ctx.Done()
		l4g.Info("Tcp server routine stoped... ")
	}()
}

// inner call by srv node
func (mi *ServiceCenter) innerCall(req *data.UserRequestData, res *data.UserResponseData) {
	api := mi.getApiInfo(req)
	if api == nil {
		res.Err = data.ErrNotFindSrv
		res.ErrMsg = data.ErrNotFindSrvText
		l4g.Error("%s %s", req.String(), res.ErrMsg)
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

// user call by user
func (mi *ServiceCenter) userCall(req *data.UserRequestData, res *data.UserResponseData) {
	// can not call auth service
	if req.Method.Srv == "auth" {
		res.Err = data.ErrIllegallyCall
		res.ErrMsg = data.ErrIllegallyCallText
		l4g.Error("%s %s", req.String(), res.ErrMsg)
		return
	}

	// find api
	api := mi.getApiInfo(req)
	if api == nil {
		res.Err = data.ErrNotFindSrv
		res.ErrMsg = data.ErrNotFindSrvText
		l4g.Error("%s %s", req.String(), res.ErrMsg)
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

// push call by srv node
func (mi *ServiceCenter) pushCall(req *data.UserRequestData, res *data.UserResponseData) {
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
	reqPush.Data.Argv = reqEncryptedRes.Data.Value
	var reqPushRes data.SrvResponseData

	mi.callFunction(&reqPush, &reqPushRes)
	*res = reqPushRes.Data

	// push data to user
	//userPushData := data.UserResponseData{}
	//userPushData.Method = req.Method
	//userPushData.Value = reqEncryptedRes.Data.Value
	//err := mi.pushWsData(&userPushData)
	//if err != nil {
	//	res.Err = data.ErrPushDataFailed
	//	res.ErrMsg = data.ErrPushDataFailedText
	//}else{
	//	res.Err = data.NoErr
	//}
}

// auth data
func (mi *ServiceCenter) authData(req *data.SrvRequestData, res *data.SrvResponseData) {
	reqAuth := *req
	reqAuth.Data.Method.Srv = "auth"
	reqAuth.Data.Method.Function = "AuthData"
	reqAuthRes := data.SrvResponseData{}

	mi.callFunction(&reqAuth, &reqAuthRes)

	*res = reqAuthRes
}

// package data
func (mi *ServiceCenter) encryptData(req *data.SrvRequestData, res *data.SrvResponseData) {
	reqEnc := *req
	reqEnc.Data.Method.Srv = "auth"
	reqEnc.Data.Method.Function = "EncryptData"

	reqEncRes := data.SrvResponseData{}

	mi.callFunction(&reqEnc, &reqEncRes)

	*res = reqEncRes
}

//  call a srv node
func (mi *ServiceCenter) callFunction(req *data.SrvRequestData, res *data.SrvResponseData) {
	var err error
	var nodeAddr string

	func() {
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
				res.Data.ErrMsg = data.ErrNotFindFunctionText
			}
			if res.Data.Err != data.NoErr {
				l4g.Error("call failed: %s", res.Data.ErrMsg)
			}
			return
		}else{
			srvNodeGroup := mi.SrvNodeNameMapSrvNodeGroup[versionSrvName]
			if srvNodeGroup == nil{
				res.Data.Err = data.ErrNotFindSrv
				res.Data.ErrMsg = data.ErrNotFindSrvText
				l4g.Error("%s %s", req.Data.String(), res.Data.ErrMsg)
				err = errors.New(res.Data.ErrMsg)
				return
			}

			nodeAddr, err = srvNodeGroup.Dispatch(req, res)
		}
	}()

	// failed, remove this node
	if err != nil && nodeAddr != ""{
		regData := data.SrvRegisterData{}
		regData.Srv = req.Data.Method.Srv
		regData.Version = req.Data.Method.Version
		regData.Addr = nodeAddr
		var rs string
		mi.UnRegister(&regData, &rs)
	}
}

func (mi *ServiceCenter) getApiInfo(req *data.UserRequestData) (*data.ApiInfo) {
	mi.rwmu.RLock()
	defer mi.rwmu.RUnlock()

	name := strings.ToLower(req.Method.Srv + "." + req.Method.Version + "." + req.Method.Function)
	return mi.ApiInfo[name]
}

// http handler
func (mi *ServiceCenter) handleWallet(w http.ResponseWriter, req *http.Request) {
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
		path = strings.Replace(path, "wallet", "", -1)
		path = strings.TrimLeft(path, "/")
		path = strings.TrimRight(path, "/")

		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			l4g.Error("http handler: %s", err.Error())
			resData.Err = data.ErrDataCorrupted
			resData.ErrMsg = data.ErrDataCorruptedText
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
			resData.ErrMsg = data.ErrDataCorruptedText
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

	// write back http
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(resData)
	w.Write(b)
	return
}

// ws handler
func (mi *ServiceCenter)handleWebSocket(conn *websocket.Conn) {
	for {
		l4g.Debug("ws handle data...")
		var err error
		var data string
		err = websocket.Message.Receive(conn, &data)
		if err == nil{
			err = mi.handleWsData(conn, data)
		}

		if err != nil {
			//移除出错的链接
			mi.removeWsClient(conn)
			l4g.Error("ws read failed, remove client:%s", err.Error())
			break
		}
	}
}

func (mi *ServiceCenter)addWsClient(conn *websocket.Conn, client *WsClient) error{
	var err error

	mi.rwmuws.Lock()
	defer mi.rwmuws.Unlock()

	mi.wsClients[conn] = client
	mi.licenseKey2wsClients[client.licenseKey] = conn

	l4g.Debug("add, ws client = %d", len(mi.wsClients))
	return err
}

func (mi *ServiceCenter)removeWsClient(conn *websocket.Conn) error{
	var err error

	conn.Close()

	mi.rwmuws.Lock()
	defer mi.rwmuws.Unlock()

	licenseKey := ""
	v := mi.wsClients[conn]
	if v != nil {
		licenseKey = v.licenseKey
	}

	delete(mi.wsClients, conn)
	if licenseKey != ""{
		delete(mi.licenseKey2wsClients, licenseKey)
	}

	l4g.Debug("remove, ws client = %d", len(mi.wsClients))
	return err
}

func (mi *ServiceCenter)handleWsData(conn *websocket.Conn, msg string) error{
	mi.wg.Add(1)
	defer mi.wg.Done()

	// only handle login request
	resData := data.UserResponseData{}
	err := func () error {
		// 重组rpc结构json
		reqData := data.UserRequestData{}
		err := json.Unmarshal([]byte(msg), &reqData);
		if err != nil {
			resData.Err = data.ErrDataCorrupted
			resData.ErrMsg = data.ErrDataCorruptedText
			l4g.Error("ws parse failed:%s", err.Error())
			return err
		}
		resData.Method = reqData.Method

		if reqData.Method.Srv != "account" {
			resData.Err = data.ErrIllegallyCall
			resData.ErrMsg = data.ErrIllegallyCallText
			l4g.Error("ws illegally call:%s", reqData.Method.Srv)
			return errors.New(resData.ErrMsg)
		}

		if reqData.Method.Function != "login" && reqData.Method.Function != "logout" {
			resData.Err = data.ErrIllegallyCall
			resData.ErrMsg = data.ErrIllegallyCallText
			l4g.Error("ws illegally call:%s", reqData.Method.Function)
			return errors.New(resData.ErrMsg)
		}

		mi.userCall(&reqData, &resData)
		resData.Method = reqData.Method
		return nil
	}()

	// write back
	b, _ := json.Marshal(resData)
	websocket.Message.Send(conn, string(b))

	if err == nil && resData.Err == data.NoErr && resData.Value.UserKey != ""{
		wsc := &WsClient{licenseKey:resData.Value.UserKey}
		mi.addWsClient(conn, wsc)

		return nil
	}

	if resData.Err != data.NoErr {
		err = errors.New(resData.ErrMsg)
	}

	return err
}

func (mi *ServiceCenter)pushWsData(d *data.UserResponseData) error {
	b, err := json.Marshal(*d)
	if err != nil {
		return err
	}

	mi.rwmuws.RLock()
	defer mi.rwmuws.RUnlock()

	conn := mi.licenseKey2wsClients[d.Value.UserKey]
	if conn != nil {
		err = websocket.Message.Send(conn, string(b))
	}else{
		err = errors.New("no client login")
	}

	return err
}

// RPC -- listsrv
func (mi *ServiceCenter) listSrv(req *data.SrvRequestData, res *data.SrvResponseData) {
	mi.rwmu.RLock()
	defer mi.rwmu.RUnlock()

	var nodes []data.SrvRegisterData
	for _, v := range mi.SrvNodeNameMapSrvNodeGroup{
		v.ListSrv(&nodes)
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