package service

import (
	"sync"
	"net/rpc"
	"../../data"
	"../nethelper"
	"log"
	"encoding/json"
	"fmt"
	"errors"
	"context"
	"net/http"
	"io/ioutil"
	"strings"
	"net/rpc/jsonrpc"
	"io"
	"bytes"
	"net"
)

type ServiceNodeInfo struct{
	RegisterData data.ServiceCenterRegisterData

	Rwmu sync.RWMutex
	Client *rpc.Client
}

type ServiceNodeBusiness struct{
	AddrMapServiceNodeInfo map[string]*ServiceNodeInfo
}

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
	ApiMapServiceName map[string]string // api+version mapto name+version
	ServiceNameMapBusiness map[string]*ServiceNodeBusiness // name+version mapto allservicenode

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

	serviceCenter.ApiMapServiceName = make(map[string]string)
	serviceCenter.ServiceNameMapBusiness = make(map[string]*ServiceNodeBusiness)

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
	err := mi.registerServiceNodeInfo(reg)
	if err != nil {
		log.Println("#Register Error: ", err.Error())
		return err
	}

	*res = "ok"
	fmt.Println("Addr ", reg.Addr, " register in gateway...")
	return nil
}

// 派发明亮
func (mi *ServiceCenter) Dispatch(req *data.ServiceCenterDispatchData, ack *data.ServiceCenterDispatchAckData) error {
	mi.wg.Add(1)
	defer mi.wg.Done()

	ack.Id = req.Id
	versionApi := req.GetVersionApi()
	fmt.Println("Center dispatch versionApi...", versionApi)

	nodeInfo := mi.getServiceNodeInfoByApi(versionApi)
	if nodeInfo == nil {
		fmt.Println("#Error: not find versionApi")
		ack.Err = data.ServiceDispatchErrNotFindApi
		ack.ErrMsg = "Not find api"
		return nil
	}

	if nodeInfo.Client == nil {
		mi.openClient(nodeInfo)
	}

	err := func() error {
		if nodeInfo.Client != nil {
			nodeInfo.Rwmu.RLock()
			defer nodeInfo.Rwmu.RUnlock()

			return nethelper.CallJRPCToTcpServerOnClient(nodeInfo.Client, data.MethodServiceNodeCall, req, ack)
		}
		return errors.New("Srv is not online")
	}()

	if err != nil {
		fmt.Println("#Call versionApi failed, close client, ", err.Error())

		mi.closeClient(nodeInfo)

		ack.Err = data.ServiceDispatchErrNotFindApi
		ack.ErrMsg = "Others error"
		return err;
	}

	return err
}

// 服务中心方法--与服务中心心跳
func (mi *ServiceCenter) Pingpong(req *string, res * string) error {
	if *req == "ping" {
		*res = "pong"
	}else{
		*res = *req
	}
	return nil;
}

// 内部方法
func (mi *ServiceCenter)registerServiceNodeInfo(registerData *data.ServiceCenterRegisterData) error{
	mi.Rwmu.Lock()
	defer mi.Rwmu.Unlock()

	//version := registerData.Version
	version := strings.ToLower(registerData.Version)

	versionName := registerData.GetVersionName()

	business := mi.ServiceNameMapBusiness[versionName]
	if business == nil {
		business = new(ServiceNodeBusiness)
		mi.ServiceNameMapBusiness[versionName] = business;
	}

	for i := 0; i < len(registerData.Apis); i++ {
		//api := registerData.Apis[i]
		api := version + "." + strings.ToLower(registerData.Apis[i])
		mi.ApiMapServiceName[api] = versionName;
	}

	if business.AddrMapServiceNodeInfo == nil {
		business.AddrMapServiceNodeInfo = make(map[string]*ServiceNodeInfo)
	}

	if business.AddrMapServiceNodeInfo[registerData.Addr] == nil {
		business.AddrMapServiceNodeInfo[registerData.Addr] = &ServiceNodeInfo{RegisterData:*registerData, Client:nil};
	}

	fmt.Println("nodes = ", len(business.AddrMapServiceNodeInfo))
	return nil
}

func (mi *ServiceCenter)getServiceNodeInfoByApi(versionApi string) *ServiceNodeInfo{
	mi.Rwmu.RLock()
	defer mi.Rwmu.RUnlock()

	versionName := mi.ApiMapServiceName[versionApi]
	if versionName == ""{
		return nil
	}

	business := mi.ServiceNameMapBusiness[versionName]
	if business == nil || business.AddrMapServiceNodeInfo == nil{
		return nil
	}

	var nodeInfo *ServiceNodeInfo
	nodeInfo = nil
	for _, v := range business.AddrMapServiceNodeInfo{
		nodeInfo = v
		break
	}

	// first we return index 0
	return nodeInfo
}

func (mi *ServiceCenter)removeServiceNodeInfo(nodeInfo *ServiceNodeInfo) error{
	mi.Rwmu.Lock()
	defer mi.Rwmu.Unlock()

	business := mi.ServiceNameMapBusiness[nodeInfo.RegisterData.Name]
	if business == nil || business.AddrMapServiceNodeInfo == nil{
		return nil
	}

	delete(business.AddrMapServiceNodeInfo, nodeInfo.RegisterData.Addr)

	fmt.Println("nodes = ", len(business.AddrMapServiceNodeInfo))
	return nil
}

func (mi *ServiceCenter)openClient(nodeInfo *ServiceNodeInfo) error{
	nodeInfo.Rwmu.Lock()
	defer nodeInfo.Rwmu.Unlock()

	if nodeInfo.Client == nil{
		client, err := rpc.Dial("tcp", nodeInfo.RegisterData.Addr)
		if err != nil {
			log.Println("Error Open client: ", err.Error())
			return err
		}

		nodeInfo.Client = client
	}

	return nil
}

func (mi *ServiceCenter)closeClient(nodeInfo *ServiceNodeInfo) error{
	nodeInfo.Rwmu.Lock()
	defer nodeInfo.Rwmu.Unlock()

	if nodeInfo.Client != nil{
		nodeInfo.Client.Close()
		nodeInfo.Client = nil
	}

	//mi.RemoveServiceNodeInfo(nodeInfo)

	return nil
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

	versionApi := req.URL.Path
	versionApi = strings.Replace(versionApi, "restful", "", -1)

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

	var version = true
	paths := strings.Split(versionApi, "/")
	for i := 0; i < len(paths); i++ {
		if paths[i] == "" {
			continue;
		}
		if version {
			dispatchData.Version = paths[i]
			version = false
		}else{
			dispatchData.Api += paths[i] + "."
		}
	}
	dispatchData.Api = strings.TrimRight(dispatchData.Api, ".")

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