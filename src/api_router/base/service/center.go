package service

import (
	"sync"
	"net/rpc"
	"../common"
	"../jrpc"
	wlutil "../util"
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
	RegisterData common.ServiceCenterRegisterData

	Rwmu sync.RWMutex
	Client *rpc.Client
}

type ServiceNodeBusiness struct{
	AddrMapServiceNodeInfo map[string]*ServiceNodeInfo
}

type ServiceCenter struct{
	// 名称
	Name string

	// rpc服务
	RpcServer *rpc.Server

	// http
	HttpPort string

	// tcp
	TcpPort string

	// 节点信息
	Rwmu sync.RWMutex
	ApiMapServiceName map[string]string // api+version mapto name+version
	ServiceNameMapBusiness map[string]*ServiceNodeBusiness // name+version mapto allservicenode

	// 等待
	Wg *sync.WaitGroup

	// 请求队列
	ApiRequestTracer *wlutil.ApiRequestTracer
}

// 生成一个服务中心
func NewServiceCenter(rootName string) (*ServiceCenter, error){
	serviceCenter := &ServiceCenter{}

	serviceCenter.Wg = &sync.WaitGroup{}
	serviceCenter.Name = rootName
	serviceCenter.ApiMapServiceName = make(map[string]string)
	serviceCenter.ServiceNameMapBusiness = make(map[string]*ServiceNodeBusiness)

	serviceCenter.ApiRequestTracer = wlutil.NewApiRequestTracer()

	return serviceCenter, nil
}

// 启动服务中心
func StartCenter(ctx context.Context, serviceCenter *ServiceCenter) {
	serviceCenter.RpcServer.HandleHTTP("/wallet", "/wallet_debug")
	http.Handle("/walletrpc", http.HandlerFunc(serviceCenter.walletRpc))
	http.Handle("/walletrest/", http.HandlerFunc(serviceCenter.walletRest))

	go serviceCenter.ApiRequestTracer.StartTracer(ctx, serviceCenter.Wg)
	go startHttpServiceCenter(ctx, serviceCenter)
	go startServiceCenter(ctx, serviceCenter)

	<-ctx.Done()
}

// RPC 方法
// 服务中心方法--注册到服务中心
func (mi *ServiceCenter) Register(req *string, res *string) error {
	RegisterData := common.ServiceCenterRegisterData{}
	err := json.Unmarshal([]byte(*req), &RegisterData);
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return err;
	}

	err = mi.registerServiceNodeInfo(&RegisterData)
	if err != nil {
		log.Println("Error: ", err.Error())
		return err
	}

	return nil
}

// 服务中心方法--发送请求到服务中心进行转发
/*
func (mi *ServiceCenter) Dispatch(req *string, res * string) error {
	dispatchData := &common.ServiceCenterDispatchData{}
	err := json.Unmarshal([]byte(*req), &dispatchData);
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return err;
	}

	fmt.Println("A module dispatch in...", *req)
	nodeInfo := mi.getServiceNodeInfoByApi(dispatchData.Api)
	if nodeInfo == nil {
		fmt.Println("Error: not find api")
		return errors.New("not find api")
	}

	if nodeInfo.Client == nil {
		mi.openClient(nodeInfo)
	}

	err = func() error {
		if nodeInfo.Client != nil {
			nodeInfo.Rwmu.RLock()
			defer nodeInfo.Rwmu.RUnlock()
			return jrpc.CallJRPCToTcpServerOnClient(nodeInfo.Client, common.MethodServiceNodeCall, dispatchData, res)
		}
		return nil
	}()

	if err != nil {
		fmt.Println("Call service api failed close client, ", err.Error())

		mi.closeClient(nodeInfo)
		return err;
	}

	fmt.Println("A module dispatch in callback")
	return err
}
*/
func (mi *ServiceCenter) Dispatch(ask *common.ServiceCenterDispatchData, ack *common.ServiceCenterDispatchAckData) error {
	mi.Wg.Add(1)
	defer mi.Wg.Done()

	versionApi := ask.GetVersionApi()

	fmt.Println("A module dispatch in versionApi...", versionApi)

	id := mi.ApiRequestTracer.TraceApi(versionApi)
	if id == "" {
		ack.Err = common.ServiceDispatchErrServiceStop
		ack.ErrMsg = "service is stop"
		return nil
	}
	defer mi.ApiRequestTracer.FinishTraceApi(id)

	nodeInfo := mi.getServiceNodeInfoByApi(versionApi)
	if nodeInfo == nil {
		fmt.Println("Error: not find versionApi")
		return errors.New("not find versionApi")
	}

	if nodeInfo.Client == nil {
		mi.openClient(nodeInfo)
	}

	err := func() error {
		if nodeInfo.Client != nil {
			nodeInfo.Rwmu.RLock()
			defer nodeInfo.Rwmu.RUnlock()
			return jrpc.CallJRPCToTcpServerOnClient(nodeInfo.Client, common.MethodServiceNodeCall, ask, ack)
		}
		return nil
	}()

	if err != nil {
		fmt.Println("Call service versionApi failed close client, ", err.Error())

		mi.closeClient(nodeInfo)
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
func startHttpServiceCenter(ctx context.Context, serviceCenter *ServiceCenter){
	serviceCenter.Wg.Add(1)
	defer serviceCenter.Wg.Done()

	log.Println("Start Http server on ", serviceCenter.HttpPort)
	listener, err := net.Listen("tcp", serviceCenter.HttpPort)
	if err != nil {
		fmt.Println("#startHttpServiceCenter Error:", err.Error())
		return
	}

	srv := http.Server{Handler:nil}
	go srv.Serve(listener)

	<-ctx.Done()
	listener.Close()

	fmt.Println("i am quit startHttpServiceCenter")
}

func startServiceCenter(ctx context.Context, serviceCenter *ServiceCenter){
	serviceCenter.Wg.Add(1)
	defer serviceCenter.Wg.Done()

	log.Println("Start JRPC Tcp server on ", serviceCenter.TcpPort)
	l, err := jrpc.CreateTcpServer(serviceCenter.TcpPort)
	if err != nil {
		fmt.Println("Error: Create tcp server, ", err.Error())
		return
	}

	go func(serviceCenter *ServiceCenter){
		for{
			conn, err := l.Accept();
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}

			log.Println("JRPC Tcp server Accept a client: ", conn.RemoteAddr())
			//go rpc.ServeConn(conn)
			go serviceCenter.RpcServer.ServeConn(conn)
		}
	}(serviceCenter)

	<- ctx.Done()
	fmt.Println("i am quit startServiceCenter")
}

func (mi *ServiceCenter)registerServiceNodeInfo(registerData *common.ServiceCenterRegisterData) error{
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
func (mi *ServiceCenter) walletRpc(w http.ResponseWriter, req *http.Request) {
	log.Println("Http server Accept a rpc client: ", req.RemoteAddr)
	defer req.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	rpcReq := newRPCRequest(req.Body)

	// go and wait
	go mi.RpcServer.ServeCodec(jsonrpc.NewServerCodec(rpcReq))
	<-rpcReq.done

	io.Copy(w, rpcReq.rw)
}
func (mi *ServiceCenter) walletRest(w http.ResponseWriter, req *http.Request) {
	log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	defer req.Body.Close()

	fmt.Println("path=", req.URL.Path)

	versionApi := req.URL.Path
	versionApi = strings.Replace(versionApi, "walletrest", "", -1)

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println("#HandleRequest Error: ", err.Error())
		return
	}

	body := string(b)
	fmt.Println("body=", body)

	// 重组rpc结构json
	dispatchData := common.ServiceCenterDispatchData{}
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

	dispatchAckData := common.ServiceCenterDispatchAckData{}
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