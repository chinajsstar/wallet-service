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
	"net/rpc/jsonrpc"
	"io"
	"bytes"
	"net"
	"time"
	"errors"
)

type ServiceCenter struct{
	// 名称
	name string

	// rpc服务
	rpcServer *rpc.Server

	// http
	httpPort string

	// tcp
	tcpPort string

	// 节点信息
	Rwmu sync.RWMutex
	SrvNodeNameMapSrvNodeGroup map[string]*SrvNodeGroup // name+version mapto srvnodegroup
	ApiInfo map[string]*data.ApiInfo

	// 等待
	wg *sync.WaitGroup
}

// 生成一个服务中心
func NewServiceCenter(rootName string, httpPort string, tcpPort string) (*ServiceCenter, error){
	serviceCenter := &ServiceCenter{}

	serviceCenter.name = rootName
	serviceCenter.httpPort = httpPort
	serviceCenter.tcpPort = tcpPort
	serviceCenter.rpcServer = rpc.NewServer()
	serviceCenter.rpcServer.Register(serviceCenter)

	serviceCenter.SrvNodeNameMapSrvNodeGroup = make(map[string]*SrvNodeGroup)

	return serviceCenter, nil
}

// 启动服务中心
func (mi *ServiceCenter)Start(ctx context.Context, wg *sync.WaitGroup) error{
	mi.wg = wg

	mi.rpcServer.HandleHTTP("/wallet", "/wallet_debug")
	http.Handle("/rpc", http.HandlerFunc(mi.handleRpc))
	http.Handle("/restful/", http.HandlerFunc(mi.handleRestful))

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
		log.Println("Start Tcp server on ", mi.tcpPort)
		listener, err := nethelper.CreateTcpServer(mi.tcpPort)
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
					go mi.rpcServer.ServeConn(conn)
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
func (mi *ServiceCenter) Register(reg *data.ServiceCenterRegisterData, res *string) error {
	mi.Rwmu.Lock()
	defer mi.Rwmu.Unlock()

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
func (mi *ServiceCenter) UnRegister(reg *data.ServiceCenterRegisterData, res *string) error {
	mi.Rwmu.Lock()
	defer mi.Rwmu.Unlock()

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
func (mi *ServiceCenter) Dispatch(req *data.ServiceCenterDispatchData, ack *data.ServiceCenterDispatchAckData) error {
	mi.wg.Add(1)
	defer mi.wg.Done()

	// 禁止直接调用auth
	if req.Srv == "auth" {
		return errors.New("not support this srv")
	}

	api := mi.getApiInfo(req)
	if api == nil {
		return errors.New("not find this api")
	}

	var rpcAuth data.SrvDispatchData
	rpcAuth.SrvArgv = *req
	rpcAuth.Api = *api

	var rpcAuthAck data.SrvDispatchAckData

	// 验证数据
	if mi.authData(&rpcAuth, &rpcAuthAck) != nil || rpcAuthAck.SrvAck.Err != data.NoErr{
		// 失败
		*ack = rpcAuthAck.SrvAck
		return nil
	}

	// 请求服务
	var rpcSrv data.SrvDispatchData
	rpcSrv.SrvArgv = *req
	rpcSrv.SrvArgv.Argv.Message = rpcAuthAck.SrvAck.Value.Message
	var rpcSrvAck data.SrvDispatchAckData
	func(){
		versionSrvName := strings.ToLower(rpcSrv.SrvArgv.Srv + "." + rpcSrv.SrvArgv.Version)
		//fmt.Println("Center dispatch...", versionSrvName, ".", rpcSrv.SrvArgv.Function)

		mi.Rwmu.RLock()
		defer mi.Rwmu.RUnlock()

		srvNodeGroup := mi.SrvNodeNameMapSrvNodeGroup[versionSrvName]
		if srvNodeGroup == nil{
			ack.Err = data.ErrNotFindSrv
			ack.ErrMsg = data.ErrNotFindSrvText
			return
		}

		srvNodeGroup.Dispatch(&rpcSrv, &rpcSrvAck)
	}()

	// 打包数据
	var reqEncrypted data.SrvDispatchData
	reqEncrypted.SrvArgv = *req
	reqEncrypted.SrvArgv.Argv.Message = rpcSrvAck.SrvAck.Value.Message

	var reqEncryptedAck data.SrvDispatchAckData

	mi.encryptData(&reqEncrypted, &reqEncryptedAck)

	// 返回
	*ack = reqEncryptedAck.SrvAck

	return nil
}

func (mi *ServiceCenter) authData(req *data.SrvDispatchData, ack *data.SrvDispatchAckData) error{
	reqAuth := *req
	reqAuth.SrvArgv.Srv = "auth"
	reqAuth.SrvArgv.Function = "AuthData"

	reqAck := data.SrvDispatchAckData{}

	versionSrvName := strings.ToLower(reqAuth.SrvArgv.Srv + "." + reqAuth.SrvArgv.Version)
	//fmt.Println("Center auth data...", versionSrvName)

	return func() error {
		mi.Rwmu.RLock()
		defer mi.Rwmu.RUnlock()

		srvNodeGroup := mi.SrvNodeNameMapSrvNodeGroup[versionSrvName]
		if srvNodeGroup == nil{
			ack.SrvAck.Err = data.ErrNotFindAuth
			ack.SrvAck.ErrMsg = data.ErrNotFindAuthText
			return errors.New("not find this auth service")
		}

		err := srvNodeGroup.Dispatch(&reqAuth, &reqAck)
		ack.SrvAck = reqAck.SrvAck

		return err
	}()
}

func (mi *ServiceCenter) encryptData(req *data.SrvDispatchData, ack *data.SrvDispatchAckData) error{
	reqEnc := *req
	reqEnc.SrvArgv.Srv = "auth"
	reqEnc.SrvArgv.Function = "EncryptData"

	reqAck := data.SrvDispatchAckData{}

	//fmt.Println("Center encrypt data...")
	versionSrvName := strings.ToLower(reqEnc.SrvArgv.Srv + "." + reqEnc.SrvArgv.Version)

	return func() error {
		mi.Rwmu.RLock()
		defer mi.Rwmu.RUnlock()

		srvNodeGroup := mi.SrvNodeNameMapSrvNodeGroup[versionSrvName]
		if srvNodeGroup == nil{
			ack.SrvAck.Err = data.ErrNotFindAuth
			ack.SrvAck.ErrMsg = data.ErrNotFindAuthText
			return errors.New("not find this auth service")
		}

		err := srvNodeGroup.Dispatch(&reqEnc, &reqAck)
		ack.SrvAck = reqAck.SrvAck

		return err
	}()
}

func (mi *ServiceCenter) getApiInfo(req *data.ServiceCenterDispatchData) (*data.ApiInfo) {
	mi.Rwmu.RLock()
	defer mi.Rwmu.RUnlock()

	name := strings.ToLower(req.Srv + "." + req.Version + "." + req.Function)
	return mi.ApiInfo[name]

}

// http 处理
// rpcRequest represents a RPC request.
// rpcRequest implements the io.ReadWriteCloser interface.
type rpcRequest struct {
	r    io.Reader     // holds the JSON formated RPC request
	rw   io.ReadWriter // holds the JSON formated RPC response
	done chan bool     // signals then end of the RPC request
}

// Read implements the io.ReadWriteCloser Read method.
func (r *rpcRequest) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

// Write implements the io.ReadWriteCloser Write method.
func (r *rpcRequest) Write(p []byte) (n int, err error) {
	return r.rw.Write(p)
}

// Close implements the io.ReadWriteCloser Close method.
func (r *rpcRequest) Close() error {
	r.done <- true
	return nil
}

// NewRPCRequest returns a new rpcRequest.
func newRPCRequest(r io.Reader) *rpcRequest {
	var buf bytes.Buffer
	done := make(chan bool)
	return &rpcRequest{r, &buf, done}
}
func (mi *ServiceCenter) handleRpc(w http.ResponseWriter, req *http.Request) {
	log.Println("Http server Accept a rpc client: ", req.RemoteAddr)
	defer req.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	rpcReq := newRPCRequest(req.Body)

	// go and wait
	go mi.rpcServer.ServeCodec(jsonrpc.NewServerCodec(rpcReq))
	<-rpcReq.done

	io.Copy(w, rpcReq.rw)
}
func (mi *ServiceCenter) handleRestful(w http.ResponseWriter, req *http.Request) {
	log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	defer req.Body.Close()

	fmt.Println("path=", req.URL.Path)

	path := req.URL.Path
	path = strings.Replace(path, "restful", "", -1)
	path = strings.TrimLeft(path, "/")
	path = strings.TrimRight(path, "/")

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println("#HandleRequest Error: ", err.Error())
		return
	}

	body := string(b)
	fmt.Println("body=", body)

	// 重组rpc结构json
	dispatchData := data.ServiceCenterDispatchData{}
	err = json.Unmarshal(b, &dispatchData);
	if err != nil {
		fmt.Println("#HandleRequest Error: ", err.Error())
		return;
	}

	// 分割参数
	paths := strings.Split(path, "/")
	for i := 0; i < len(paths); i++ {
		if i == 0 {
			dispatchData.Version = paths[i]
		}else if i == 1{
			dispatchData.Srv = paths[i]
		} else{
			if dispatchData.Function != "" {
				dispatchData.Function += "."
			}
			dispatchData.Function += paths[i]
		}
	}

	dispatchAckData := data.ServiceCenterDispatchAckData{}
	mi.Dispatch(&dispatchData, &dispatchAckData)

	w.Header().Set("Content-Type", "application/json")

	b, err = json.Marshal(dispatchAckData)
	if err != nil {
		fmt.Println("#HandleRequest Error: ", err.Error())
		return;
	}

	w.Write(b)
	return
}

func (mi *ServiceCenter)keepAlive(){
	mi.Rwmu.RLock()
	defer mi.Rwmu.RUnlock()

	for _, v := range mi.SrvNodeNameMapSrvNodeGroup{
		v.KeepAlive()
	}
}

func (mi *ServiceCenter)startToKeepAlive(ctx context.Context) error{
	timeout := make(chan bool)
	go func(){
		for ; ; {
			timeout <- true
			time.Sleep(time.Second*10)
		}
	}()

	go func() {
		mi.wg.Add(1)
		defer mi.wg.Done()

		for ; ; {
			select{
			case <-ctx.Done():
				fmt.Println("Keep alive quit...")
				return
			case <-timeout:
				mi.keepAlive()
			}
		}

	}()

	return  nil
}