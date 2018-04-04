package service

import (
	"sync"
	"net/rpc"
	"../../data"
	"../nethelper"
	"log"
	"encoding/json"
	"fmt"
	"context"
	"net/http"
	"io/ioutil"
	"strings"
	"net"
	"golang.org/x/net/websocket"
	"errors"
)

type WsClient struct{
	licenseKey string
}

type ServiceCenter struct{
	// 名称
	name string

	// http
	httpPort string

	// websocket
	wsPort string

	// center
	centerPort string

	// 节点信息
	rwmu                       sync.RWMutex
	SrvNodeNameMapSrvNodeGroup map[string]*SrvNodeGroup // name+version mapto srvnodegroup
	ApiInfo                    map[string]*data.ApiInfo

	// ws 节点信息
	rwmuws sync.RWMutex
	// 已验证用户
	licenseKey2wsClients map[string]*websocket.Conn
	wsClients map[*websocket.Conn]*WsClient

	// 等待
	wg *sync.WaitGroup

	// 连接组管理
	clientGroup *ConnectionGroup
}

// 生成一个服务中心
func NewServiceCenter(rootName string, httpPort, wsPort, centerPort string) (*ServiceCenter, error){
	serviceCenter := &ServiceCenter{}

	serviceCenter.name = rootName
	serviceCenter.httpPort = httpPort
	serviceCenter.wsPort = wsPort
	serviceCenter.centerPort = centerPort
	rpc.Register(serviceCenter)

	serviceCenter.SrvNodeNameMapSrvNodeGroup = make(map[string]*SrvNodeGroup)

	serviceCenter.wsClients = make(map[*websocket.Conn]*WsClient)
	serviceCenter.licenseKey2wsClients = make(map[string]*websocket.Conn)

	serviceCenter.clientGroup = NewConnectionGroup()

	return serviceCenter, nil
}

// 启动服务中心
func (mi *ServiceCenter)Start(ctx context.Context, wg *sync.WaitGroup) error{
	mi.wg = wg

	// http server
	http.Handle("/wallet/", http.HandlerFunc(mi.handleWallet))
	err := func() error {
		log.Println("Start Http server on ", mi.httpPort)
		listener, err := net.Listen("tcp", mi.httpPort)
		if err != nil {
			fmt.Println("#Http listen Error:", err.Error())
			return err
		}
		go func() {
			mi.wg.Add(1)
			defer mi.wg.Done()

			log.Println("Http server routine running... ")
			srv := http.Server{Handler:nil}
			go srv.Serve(listener)

			<-ctx.Done()
			listener.Close()

			log.Println("Http server routine stoped... ")
		}()

		return nil
	}()
	if err != nil {
		return err
	}

	// websocket
	http.Handle("/ws", websocket.Handler(mi.handleWebSocket))
	err = func() error {
		log.Println("Start ws server on ", mi.wsPort)
		/*
		listener, err := net.Listen("tcp", mi.wsPort)
		if err != nil {
			fmt.Println("#Http listen Error:", err.Error())
			return err
		}*/

		go func() {
			//mi.wg.Add(1)
			//defer mi.wg.Done()

			log.Println("ws server routine running... ")
			err := http.ListenAndServe(mi.wsPort, nil)
			if err != nil {
				fmt.Println("#Error:", err)
				return
			}
			//srv := http.Server{Handler:nil}
			//go srv.Serve(listener)

			//<-ctx.Done()
			//listener.Close()

			log.Println("ws server routine stoped... ")
		}()

		return nil
	}()
	if err != nil {
		return err
	}

	// tcp server
	err = func() error {
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
	}()

	return err
}

// RPC 方法
// 服务中心方法--注册到服务中心
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

// 服务中心方法--注册到服务中心
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

// 派发命令
func (mi *ServiceCenter) Dispatch(req *data.UserRequestData, res *data.UserResponseData) error {
	mi.wg.Add(1)
	defer mi.wg.Done()

	mi.innerCall(req, res)

	// 确保错误的情况下，没有实际数据
	if res.Err != data.NoErr {
		res.Value.Message = ""
		res.Value.Signature = ""
	}
	return nil
}

// 推送命令
func (mi *ServiceCenter) Push(req *data.UserResponseData, res *data.UserResponseData) error {
	mi.wg.Add(1)
	defer mi.wg.Done()

	mi.pushCall(req, res)

	// 确保错误的情况下，没有实际数据
	if res.Err != data.NoErr {
		res.Value.Message = ""
		res.Value.Signature = ""
	}
	return nil
}

// 内部调用
func (mi *ServiceCenter) innerCall(req *data.UserRequestData, res *data.UserResponseData) {
	api := mi.getApiInfo(req)
	if api == nil {
		res.Err = data.ErrNotFindSrv
		res.ErrMsg = data.ErrNotFindSrvText
		fmt.Println("#Error: ", res.ErrMsg)
		return
	}

	// 请求具体服务
	var rpcSrv data.SrvRequestData
	rpcSrv.Data = *req
	rpcSrv.Context.Api = *api
	var rpcSrvRes data.SrvResponseData
	if mi.callFunction(&rpcSrv, &rpcSrvRes); rpcSrvRes.Data.Err != data.NoErr{
		// 失败
		*res = rpcSrvRes.Data
		return
	}

	// 返回
	*res = rpcSrvRes.Data
}

// 用户调用
func (mi *ServiceCenter) userCall(req *data.UserRequestData, res *data.UserResponseData) {
	// 禁止直接调用auth
	if req.Method.Srv == "auth" {
		res.Err = data.ErrIllegalCall
		res.ErrMsg = data.ErrIllegalCallText
		fmt.Println("#Error: ", res.ErrMsg)
		return
	}

	api := mi.getApiInfo(req)
	if api == nil {
		res.Err = data.ErrNotFindSrv
		res.ErrMsg = data.ErrNotFindSrvText
		fmt.Println("#Error: ", res.ErrMsg)
		return
	}

	// 验证数据
	var rpcAuth data.SrvRequestData
	rpcAuth.Data = *req
	rpcAuth.Context.Api = *api
	var rpcAuthRes data.SrvResponseData
	if mi.authData(&rpcAuth, &rpcAuthRes); rpcAuthRes.Data.Err != data.NoErr{
		// 失败
		*res = rpcAuthRes.Data
		return
	}

	// 请求具体服务
	var rpcSrv data.SrvRequestData
	rpcSrv.Data = *req
	rpcSrv.Context.Api = *api
	rpcSrv.Data.Argv.Message = rpcAuthRes.Data.Value.Message
	var rpcSrvRes data.SrvResponseData
	if mi.callFunction(&rpcSrv, &rpcSrvRes); rpcSrvRes.Data.Err != data.NoErr{
		// 失败
		*res = rpcSrvRes.Data
		return
	}

	// 打包数据
	var reqEncrypted data.SrvRequestData
	reqEncrypted.Data = *req
	reqEncrypted.Context.Api = *api
	reqEncrypted.Data.Argv.Message = rpcSrvRes.Data.Value.Message
	var reqEncryptedRes data.SrvResponseData
	if mi.encryptData(&reqEncrypted, &reqEncryptedRes); reqEncryptedRes.Data.Err != data.NoErr{
		// 失败
		*res = reqEncryptedRes.Data
		return
	}

	// 返回
	*res = reqEncryptedRes.Data
}

// 推送调用
func (mi *ServiceCenter) pushCall(req *data.UserResponseData, res *data.UserResponseData) {
	// 打包数据
	var reqEncrypted data.SrvRequestData
	reqEncrypted.Data.Method = req.Method
	reqEncrypted.Data.Argv = req.Value
	var reqEncryptedRes data.SrvResponseData
	if mi.encryptData(&reqEncrypted, &reqEncryptedRes); reqEncryptedRes.Data.Err != data.NoErr{
		// 失败
		*res = reqEncryptedRes.Data
		return
	}

	// 推送数据
	userPushData := data.UserResponseData{}
	userPushData.Method = req.Method
	userPushData.Value = reqEncryptedRes.Data.Value

	// 推送
	err := mi.pushWsData(&userPushData)
	if err != nil {
		res.Err = data.ErrPush
		res.ErrMsg = data.ErrPushText
	}else{
		res.Err = data.NoErr
	}
}

func (mi *ServiceCenter) authData(req *data.SrvRequestData, res *data.SrvResponseData) {
	reqAuth := *req
	reqAuth.Data.Method.Srv = "auth"
	reqAuth.Data.Method.Function = "AuthData"
	reqAuthRes := data.SrvResponseData{}

	mi.callFunction(&reqAuth, &reqAuthRes)

	*res = reqAuthRes
}

func (mi *ServiceCenter) encryptData(req *data.SrvRequestData, res *data.SrvResponseData) {
	reqEnc := *req
	reqEnc.Data.Method.Srv = "auth"
	reqEnc.Data.Method.Function = "EncryptData"

	reqEncRes := data.SrvResponseData{}

	mi.callFunction(&reqEnc, &reqEncRes)

	*res = reqEncRes
}

func (mi *ServiceCenter) callFunction(req *data.SrvRequestData, res *data.SrvResponseData) {
	versionSrvName := strings.ToLower(req.Data.Method.Srv + "." + req.Data.Method.Version)
	//fmt.Println("Center dispatch function...", versionSrvName, ".", req.Data.Function)

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
			resData.Err = data.ErrData
			resData.ErrMsg = data.ErrDataText
			return
		}

		//body := string(b)
		//fmt.Println("body=", body)

		// 重组rpc结构json
		reqData := data.UserRequestData{}
		err = json.Unmarshal(b, &reqData.Argv);
		if err != nil {
			fmt.Println("#Error, handleRestful: ", err.Error())
			resData.Err = data.ErrData
			resData.ErrMsg = data.ErrDataText
			return
		}

		// 分割参数
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

// ws
func (mi *ServiceCenter)handleWebSocket(conn *websocket.Conn) {
	for {
		// 连接...
		fmt.Println("开始解析数据...")
		var err error
		var data string
		err = websocket.Message.Receive(conn, &data)
		if err == nil{
			err = mi.handleWsData(conn, data)
		}

		fmt.Println("data:", data)

		if err != nil {
			//移除出错的链接
			mi.removeWsClient(conn)
			fmt.Println("读取数据出错...", err)
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

	fmt.Println("ws client = ", len(mi.wsClients))
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

	fmt.Println("ws client = ", len(mi.wsClients))
	return err
}

func (mi *ServiceCenter)handleWsData(conn *websocket.Conn, msg string) error{
	mi.wg.Add(1)
	defer mi.wg.Done()

	// 只处理登陆登出，其他的消息推送
	resData := data.UserResponseData{}
	err := func () error {
		// 重组rpc结构json
		reqData := data.UserRequestData{}
		err := json.Unmarshal([]byte(msg), &reqData);
		if err != nil {
			fmt.Println("#Error, handlews: ", err.Error())
			resData.Err = data.ErrData
			resData.ErrMsg = data.ErrDataText
			return err
		}
		resData.Method = reqData.Method

		if reqData.Method.Srv != "account" {
			fmt.Println("#Error, handlews: 非法")
			resData.Err = data.ErrIllegalCall
			resData.ErrMsg = data.ErrIllegalCallText
			return errors.New("非法调用")
		}

		if reqData.Method.Function != "login" && reqData.Method.Function != "logout" {
			fmt.Println("#Error, handlews: 非法")
			resData.Err = data.ErrIllegalCall
			resData.ErrMsg = data.ErrIllegalCallText
			return errors.New("非法调用")
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