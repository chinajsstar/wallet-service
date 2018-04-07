package service

import (
	"sync"
	"net/rpc"
	"../../data"
	"../nethelper"
	"../config"
	"log"
	"encoding/json"
	"fmt"
	"context"
	"net/http"
	"io/ioutil"
	"strings"
	"golang.org/x/net/websocket"
	"errors"
)

type WsClient struct{
	// ws client license key
	licenseKey string
}

type ServiceCenter struct{
	// name
	name string

	// http
	httpPort string

	// websocket
	wsPort string

	// center
	centerPort string

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
}

// new a center
func NewServiceCenter(confPath string) (*ServiceCenter, error){
	cfgCenter := config.ConfigCenter{}
	if err := cfgCenter.Load(confPath); err != nil{
		return nil, err
	}
	serviceCenter := &ServiceCenter{}

	serviceCenter.name = cfgCenter.CenterName
	serviceCenter.httpPort = ":"+cfgCenter.Port
	serviceCenter.wsPort = ":"+cfgCenter.WsPort
	serviceCenter.centerPort = ":"+cfgCenter.CenterPort

	serviceCenter.SrvNodeNameMapSrvNodeGroup = make(map[string]*SrvNodeGroup)

	serviceCenter.wsClients = make(map[*websocket.Conn]*WsClient)
	serviceCenter.licenseKey2wsClients = make(map[string]*websocket.Conn)

	serviceCenter.clientGroup = NewConnectionGroup()

	return serviceCenter, nil
}

// start the service center
func StartCenter(ctx context.Context, mi *ServiceCenter) error{
	if err := mi.startHttpServer(ctx); err != nil {
		return err
	}

	if err := mi.startWsServer(ctx); err != nil {
		return err
	}

	if err := mi.startTcpServer(ctx); err != nil {
		return err
	}

	return nil
}

// Stop the service center
func StopCenter(mi *ServiceCenter)  {
	mi.wg.Wait()
}

// RPC -- register
func (mi *ServiceCenter) Register(reg *data.SrvRegisterData, res *string) error {
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
			mi.ApiInfo[strings.ToLower(versionSrvName+"."+v.Name)] = &data.ApiInfo{v.Name, v.Level}
		}
	}

	return err
}

// RPC -- unregister
func (mi *ServiceCenter) UnRegister(reg *data.SrvRegisterData, res *string) error {
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

// RPC -- dispatch
func (mi *ServiceCenter) Push(req *data.UserResponseData, res *data.UserResponseData) error {
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
func (mi *ServiceCenter) startHttpServer(ctx context.Context) error {
	// http
	log.Println("Start http server on ", mi.httpPort)

	http.Handle("/wallet/", http.HandlerFunc(mi.handleWallet))

	go func() {
		log.Println("Http server routine running... ")
		err := http.ListenAndServe(mi.httpPort, nil)
		if err != nil {
			fmt.Println("#Error:", err)
			return
		}
	}()

	return nil
}

// start websocket server
func (mi *ServiceCenter) startWsServer(ctx context.Context) error {
	// websocket
	log.Println("Start ws server on ", mi.wsPort)

	http.Handle("/ws", websocket.Handler(mi.handleWebSocket))

	go func() {
		log.Println("ws server routine running... ")
		err := http.ListenAndServe(mi.wsPort, nil)
		if err != nil {
			fmt.Println("#Error:", err)
			return
		}
	}()

	return nil
}

// start tcp server
func (mi *ServiceCenter) startTcpServer(ctx context.Context) error {
	log.Println("Start Tcp server on ", mi.centerPort)

	listener, err := nethelper.CreateTcpServer(mi.centerPort)
	if err != nil {
		log.Println("#ListenTCP Error: ", err.Error())
		return err
	}
	go func() {
		mi.wg.Add(1)
		defer mi.wg.Done()

		log.Println("Tcp server routine running... ")
		go func(){
			for{
				conn, err := listener.Accept();
				if err != nil {
					log.Println("Error: ", err.Error())
					continue
				}

				log.Println("Tcp server Accept a client: ", conn.RemoteAddr())
				rc := mi.clientGroup.Register(conn)

				go func() {
					go rpc.ServeConn(rc)
					<- rc.Done
					log.Println("Tcp server close a client: ", conn.RemoteAddr())
				}()

			}
		}()

		<- ctx.Done()
		log.Println("Tcp server routine stoped... ")
	}()

	return nil
}

// inner call by srv node
func (mi *ServiceCenter) innerCall(req *data.UserRequestData, res *data.UserResponseData) {
	api := mi.getApiInfo(req)
	if api == nil {
		res.Err = data.ErrNotFindSrv
		res.ErrMsg = data.ErrNotFindSrvText
		fmt.Println("#Error: ", res.ErrMsg)
		return
	}

	// call function
	var rpcSrv data.SrvRequestData
	rpcSrv.Data = *req
	rpcSrv.Context.Api = *api
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
		fmt.Println("#Error: ", res.ErrMsg)
		return
	}

	// find api
	api := mi.getApiInfo(req)
	if api == nil {
		res.Err = data.ErrNotFindSrv
		res.ErrMsg = data.ErrNotFindSrvText
		fmt.Println("#Error: ", res.ErrMsg)
		return
	}

	// decode and verify data
	var rpcAuth data.SrvRequestData
	rpcAuth.Data = *req
	rpcAuth.Context.Api = *api
	var rpcAuthRes data.SrvResponseData
	if mi.authData(&rpcAuth, &rpcAuthRes); rpcAuthRes.Data.Err != data.NoErr{
		*res = rpcAuthRes.Data
		return
	}

	// call real srv
	var rpcSrv data.SrvRequestData
	rpcSrv.Data = *req
	rpcSrv.Context.Api = *api
	rpcSrv.Data.Argv.Message = rpcAuthRes.Data.Value.Message
	var rpcSrvRes data.SrvResponseData
	if mi.callFunction(&rpcSrv, &rpcSrvRes); rpcSrvRes.Data.Err != data.NoErr{
		*res = rpcSrvRes.Data
		return
	}

	// encode and sign data
	var reqEncrypted data.SrvRequestData
	reqEncrypted.Data = *req
	reqEncrypted.Context.Api = *api
	reqEncrypted.Data.Argv.Message = rpcSrvRes.Data.Value.Message
	var reqEncryptedRes data.SrvResponseData
	if mi.encryptData(&reqEncrypted, &reqEncryptedRes); reqEncryptedRes.Data.Err != data.NoErr{
		*res = reqEncryptedRes.Data
		return
	}

	*res = reqEncryptedRes.Data
}

// push call by srv node
func (mi *ServiceCenter) pushCall(req *data.UserResponseData, res *data.UserResponseData) {
	// encode and sign data
	var reqEncrypted data.SrvRequestData
	reqEncrypted.Data.Method = req.Method
	reqEncrypted.Data.Argv = req.Value
	var reqEncryptedRes data.SrvResponseData
	if mi.encryptData(&reqEncrypted, &reqEncryptedRes); reqEncryptedRes.Data.Err != data.NoErr{
		*res = reqEncryptedRes.Data
		return
	}

	// push data to user
	userPushData := data.UserResponseData{}
	userPushData.Method = req.Method
	userPushData.Value = reqEncryptedRes.Data.Value
	err := mi.pushWsData(&userPushData)
	if err != nil {
		res.Err = data.ErrPushDataFailed
		res.ErrMsg = data.ErrPushDataFailedText
	}else{
		res.Err = data.NoErr
	}
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
	versionSrvName := strings.ToLower(req.Data.Method.Srv + "." + req.Data.Method.Version)
	fmt.Println("Center dispatch function...", versionSrvName, ".", req.Data.Method.Function)

	mi.rwmu.RLock()
	defer mi.rwmu.RUnlock()

	srvNodeGroup := mi.SrvNodeNameMapSrvNodeGroup[versionSrvName]
	if srvNodeGroup == nil{
		res.Data.Err = data.ErrNotFindSrv
		res.Data.ErrMsg = data.ErrNotFindSrvText
		fmt.Println("#Error: Center dispatch function...", res.Data.ErrMsg)
		return
	}

	srvNodeGroup.Dispatch(req, res)
}

func (mi *ServiceCenter) getApiInfo(req *data.UserRequestData) (*data.ApiInfo) {
	mi.rwmu.RLock()
	defer mi.rwmu.RUnlock()

	name := strings.ToLower(req.Method.Srv + "." + req.Method.Version + "." + req.Method.Function)
	return mi.ApiInfo[name]
}

// http handler
func (mi *ServiceCenter) handleWallet(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

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
			fmt.Println("#Error, handleRestful: ", err.Error())
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
			fmt.Println("#Error, handleRestful: ", err.Error())
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
		fmt.Println("ws handle data...")
		var err error
		var data string
		err = websocket.Message.Receive(conn, &data)
		if err == nil{
			err = mi.handleWsData(conn, data)
		}

		if err != nil {
			//移除出错的链接
			mi.removeWsClient(conn)
			fmt.Println("ws read failed, remove client...", err)
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

	fmt.Println("add, ws client = ", len(mi.wsClients))
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

	fmt.Println("remove, ws client = ", len(mi.wsClients))
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
			fmt.Println("#Error, handlews: ", err.Error())
			resData.Err = data.ErrDataCorrupted
			resData.ErrMsg = data.ErrDataCorruptedText
			return err
		}
		resData.Method = reqData.Method

		if reqData.Method.Srv != "account" {
			fmt.Println("#Error, handlews: illegally call: ", reqData.Method)
			resData.Err = data.ErrIllegallyCall
			resData.ErrMsg = data.ErrIllegallyCallText
			return errors.New(resData.ErrMsg)
		}

		if reqData.Method.Function != "login" && reqData.Method.Function != "logout" {
			fmt.Println("#Error, handlews: illegally call: ", reqData.Method)
			resData.Err = data.ErrIllegallyCall
			resData.ErrMsg = data.ErrIllegallyCallText
			return errors.New(resData.ErrMsg)
		}

		mi.userCall(&reqData, &resData)
		resData.Method = reqData.Method
		return nil
	}()

	// write back
	b, _ := json.Marshal(resData)
	websocket.Message.Send(conn, string(b))

	if err == nil && resData.Err == data.NoErr && resData.Value.LicenseKey != ""{
		wsc := &WsClient{licenseKey:resData.Value.LicenseKey}
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

	conn := mi.licenseKey2wsClients[d.Value.LicenseKey]
	if conn != nil {
		err = websocket.Message.Send(conn, string(b))
	}else{
		err = errors.New("no client login")
	}

	return err
}