package service

import (
	"sync"
	"api_router/base/data"
	"api_router/base/nethelper"
	"api_router/base/config"
	"context"
	"strings"
	"log"
	"net/rpc"
	"time"
	"errors"
	"net"
	l4g "github.com/alecthomas/log4go"
)

// node api interface
type NodeApiHandler func(req *data.SrvRequestData, res *data.SrvResponseData)
type NodeApi struct{
	ApiInfo 	data.ApiInfo
	ApiHandler 	NodeApiHandler
}
type NodeApiGroup interface {
	GetApiGroup()(map[string]NodeApi)
}

// service node
type ServiceNode struct{
	// register data
	registerData data.SrvRegisterData

	// callback
	apiHandler map[string]*NodeApi

	// center addr
	serviceCenterAddr string

	// wait group
	wg sync.WaitGroup

	// connection group
	clientGroup *ConnectionGroup

	// connection to center
	client *rpc.Client
}

// New a service node
func NewServiceNode(confPath string) (*ServiceNode, error){
	cfgNode := config.ConfigNode{}
	cfgNode.Load(confPath)

	serviceNode := &ServiceNode{}

	serviceNode.apiHandler = make(map[string]*NodeApi)

	// node info
	serviceNode.registerData.Srv = cfgNode.SrvName
	serviceNode.registerData.Version = cfgNode.SrvVersion
	serviceNode.registerData.Addr = cfgNode.SrvAddr

	// center info
	serviceNode.serviceCenterAddr = cfgNode.CenterAddr

	serviceNode.clientGroup = NewConnectionGroup()

	return serviceNode, nil
}

// register api group
func RegisterNodeApi(ni *ServiceNode, nodeApiGroup NodeApiGroup) {
	nam := nodeApiGroup.GetApiGroup()

	for k, v := range nam{
		if ni.apiHandler[k] != nil {
			log.Fatal("#Error api repeat:", k)
		}
		ni.apiHandler[k] = &NodeApi{ApiInfo:v.ApiInfo, ApiHandler:v.ApiHandler}
		ni.registerData.Functions = append(ni.registerData.Functions, v.ApiInfo)
	}
}

// Start the service node
func StartNode(ctx context.Context, ni *ServiceNode) {
	ni.startTcpServer(ctx)

	ni.startToServiceCenter(ctx)
}

// Stop the service node
func StopNode(ni *ServiceNode)  {
	ni.wg.Wait()
}

// RPC -- call
func (ni *ServiceNode) Call(req *data.SrvRequestData, res *data.SrvResponseData) error {
	h := ni.apiHandler[strings.ToLower(req.Data.Method.Function)]
	if h != nil {
		h.ApiHandler(req, res)
	}else{
		res.Data.Err = data.ErrNotFindFunction
	}
	if res.Data.Err != data.NoErr {
		l4g.Error("call failed: %s", res.Data.ErrMsg)
	}
	return nil
}

// dispatch a request to center
func (ni *ServiceNode) ListSrv(req *data.UserRequestData, res *data.UserResponseData) error {
	var err error
	if ni.client != nil {
		err = nethelper.CallJRPCToTcpServerOnClient(ni.client, data.MethodCenterListSrv, req, res)
	}else{
		err = nethelper.CallJRPCToTcpServer(ni.serviceCenterAddr, data.MethodCenterListSrv, req, res)
	}
	return err
}

// dispatch a request to center
func (ni *ServiceNode) Dispatch(req *data.UserRequestData, res *data.UserResponseData) error {
	var err error
	if ni.client != nil {
		err = nethelper.CallJRPCToTcpServerOnClient(ni.client, data.MethodCenterDispatch, req, res)
	}else{
		err = nethelper.CallJRPCToTcpServer(ni.serviceCenterAddr, data.MethodCenterDispatch, req, res)
	}
	return err
}

// push a data to center
func (ni *ServiceNode) Push(req *data.UserRequestData, res *data.UserResponseData) error {
	var err error
	if ni.client != nil {
		err = nethelper.CallJRPCToTcpServerOnClient(ni.client, data.MethodCenterPush, req, res)
	}else{
		err = nethelper.CallJRPCToTcpServer(ni.serviceCenterAddr, data.MethodCenterPush, req, res)
	}
	return err
}

func (ni *ServiceNode) startTcpServer(ctx context.Context) {
	s :=strings.Split(ni.registerData.Addr, ":")
	if len(s) != 2{
		l4g.Crashf("#Error: Node addr is not ip:port format")
	}

	listener, err := nethelper.CreateTcpServer(":"+s[1])
	if err != nil {
		l4g.Crashf("", err)
	}

	go func() {
		ni.wg.Add(1)
		defer ni.wg.Done()

		l4g.Debug("Tcp server routine running... ")
		go func(){
			for{
				conn, err := listener.Accept();
				if err != nil {
					l4g.Error("%s", err.Error())
					continue
				}

				l4g.Info("Tcp server Accept a client: %s", conn.RemoteAddr().String())
				rc := ni.clientGroup.Register(conn)

				go func() {
					go rpc.ServeConn(rc)
					<- rc.Done
					l4g.Info("Tcp server close a client: %s", conn.RemoteAddr().String())
				}()
			}
		}()

		<- ctx.Done()
		l4g.Debug("Tcp server routine stoped... ")
	}()
}

// 内部方法
func (ni *ServiceNode)connectToCenter() (*Connection, error){
	var err error

	conn, err := net.Dial("tcp", ni.serviceCenterAddr)
	if err != nil {
		l4g.Error("%s", err.Error())
		return nil, err
	}

	cn := &Connection{}
	cn.Cg = nil
	cn.Conn = conn
	cn.Done = make(chan bool)

	return cn, err
}

func (ni *ServiceNode)registToCenter() error{
	var err error
	var res string
	if ni.client != nil {
		err = ni.client.Call(data.MethodCenterRegister, ni.registerData, &res)
	}else{
		errors.New("connection is closed")
	}
	return err
}

func (ni *ServiceNode)unRegistToCenter() error{
	var err error
	var res string
	if ni.client != nil {
		err = nethelper.CallJRPCToTcpServerOnClient(ni.client, data.MethodCenterUnRegister, ni.registerData, &res)
	}else{
		err = nethelper.CallJRPCToTcpServer(ni.serviceCenterAddr, data.MethodCenterUnRegister, ni.registerData, &res)
	}

	return err
}

func (ni *ServiceNode)startToServiceCenter(ctx context.Context) {
	go func() {
		ni.wg.Add(1)
		defer ni.wg.Done()

		go func() {
			err := errors.New("not connect")
			var cn *Connection
			for {
				if err != nil {
					l4g.Info("Tcp client connect...")
					cn, err = ni.connectToCenter()
				}

				if err == nil {
					ni.client = rpc.NewClient(cn)

					ni.registToCenter()

					l4g.Info("Tcp client connected...")
					<-cn.Done
					l4g.Info("Tcp client close... ")

					err = errors.New("not connect")
				}

				time.Sleep(time.Second*5)
			}
		}()

		<-ctx.Done()
		ni.unRegistToCenter()
		l4g.Info("UnRegist to center ok %s", ni.registerData.String())
	}()
}
