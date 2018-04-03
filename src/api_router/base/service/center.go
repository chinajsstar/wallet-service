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
)

type ServiceCenter struct{
	// 名称
	name string

	// http
	httpPort string

	// center
	centerPort string

	// 节点信息
	rwmu                       sync.RWMutex
	SrvNodeNameMapSrvNodeGroup map[string]*SrvNodeGroup // name+version mapto srvnodegroup
	ApiInfo                    map[string]*data.ApiInfo

	// 等待
	wg *sync.WaitGroup

	// 连接组管理
	clientGroup *ConnectionGroup
}

// 生成一个服务中心
func NewServiceCenter(rootName string, httpPort string, centerPort string) (*ServiceCenter, error){
	serviceCenter := &ServiceCenter{}

	serviceCenter.name = rootName
	serviceCenter.httpPort = httpPort
	serviceCenter.centerPort = centerPort
	rpc.Register(serviceCenter)

	serviceCenter.SrvNodeNameMapSrvNodeGroup = make(map[string]*SrvNodeGroup)

	serviceCenter.clientGroup = NewConnectionGroup()

	return serviceCenter, nil
}

// 启动服务中心
func (mi *ServiceCenter)Start(ctx context.Context, wg *sync.WaitGroup) error{
	mi.wg = wg

	http.Handle("/wallet/", http.HandlerFunc(mi.handleWallet))

	// http server
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
	if mi.doFunction(&rpcSrv, &rpcSrvRes); rpcSrvRes.Data.Err != data.NoErr{
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
	if req.Srv == "auth" {
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
	if mi.doFunction(&rpcSrv, &rpcSrvRes); rpcSrvRes.Data.Err != data.NoErr{
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

func (mi *ServiceCenter) doFunction(req *data.SrvRequestData, res *data.SrvResponseData) {
	versionSrvName := strings.ToLower(req.Data.Srv + "." + req.Data.Version)
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

func (mi *ServiceCenter) authData(req *data.SrvRequestData, res *data.SrvResponseData) {
	reqAuth := *req
	reqAuth.Data.Srv = "auth"
	reqAuth.Data.Function = "AuthData"
	reqAuthRes := data.SrvResponseData{}

	mi.doFunction(&reqAuth, &reqAuthRes)

	*res = reqAuthRes
}

func (mi *ServiceCenter) encryptData(req *data.SrvRequestData, res *data.SrvResponseData) {
	reqEnc := *req
	reqEnc.Data.Srv = "auth"
	reqEnc.Data.Function = "EncryptData"

	reqEncRes := data.SrvResponseData{}

	mi.doFunction(&reqEnc, &reqEncRes)

	*res = reqEncRes
}

func (mi *ServiceCenter) getApiInfo(req *data.UserRequestData) (*data.ApiInfo) {
	mi.rwmu.RLock()
	defer mi.rwmu.RUnlock()

	name := strings.ToLower(req.Srv + "." + req.Version + "." + req.Function)
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
			return;
		}

		// 分割参数
		paths := strings.Split(path, "/")
		for i := 0; i < len(paths); i++ {
			if i == 0 {
				reqData.Version = paths[i]
			}else if i == 1{
				reqData.Srv = paths[i]
			} else{
				if reqData.Function != "" {
					reqData.Function += "."
				}
				reqData.Function += paths[i]
			}
		}

		mi.userCall(&reqData, &resData)
	}()

	// write back http
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(resData)
	w.Write(b)
	return
}