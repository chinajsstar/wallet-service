package service

import (
	"sync"
	"../../data"
	"../nethelper"
	"fmt"
	"context"
	"strings"
	"log"
	"net/rpc"
	"time"
	"errors"
	"net"
)

// 服务节点回调接口
type CallNodeApi func(req *data.SrvRequestData, res *data.SrvResponseData)

// 服务节点信息
type ServiceNode struct{
	// 注册的信息
	RegisterData data.SrvRegisterData
	// 回掉
	Handler CallNodeApi
	// 服务中心
	ServiceCenterAddr string
	// 等待
	wg *sync.WaitGroup
	// 连接组管理
	clientGroup *ConnectionGroup

	// 注册到gateway的rpc客户端
	client *rpc.Client
}

// 生成一个服务节点
func NewServiceNode(srvName string, versionName string) (*ServiceNode, error){
	serviceNode := &ServiceNode{}

	serviceNode.RegisterData.Srv = srvName
	serviceNode.RegisterData.Version = versionName
	serviceNode.clientGroup = NewConnectionGroup()

	return serviceNode, nil
}

// 启动服务节点
func (ni *ServiceNode)Start(ctx context.Context, wg *sync.WaitGroup) error {
	ni.wg = wg

	err := func()error{
		s :=strings.Split(ni.RegisterData.Addr, ":")
		if len(s) != 2{
			fmt.Println("#Error: Node addr is not ip:port format")
			return errors.New("#Addr is error format")
		}

		listener, err := nethelper.CreateTcpServer(":"+s[1])
		if err != nil {
			log.Println("#ListenTCP Error: ", err.Error())
			return err
		}

		go func() {
			ni.wg.Add(1)
			defer ni.wg.Done()

			log.Println("Tcp server routine running... ")
			go func(){
				for{
					conn, err := listener.Accept();
					if err != nil {
						log.Println("Error: ", err.Error())
						continue
					}

					log.Println("Tcp server Accept a client: ", conn.RemoteAddr())
					rc := ni.clientGroup.Register(conn)

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

	if err != nil {
		return err
	}

	ni.startToServiceCenter(ctx)

	return err
}

// RPC 方法
// 服务节点RPC--调用节点方法ServiceNodeInstance.Call
func (ni *ServiceNode) Call(req *data.SrvRequestData, res *data.SrvResponseData) error {
	if ni.Handler != nil {
		ni.Handler(req, res)
	}else{
		fmt.Println("Error function call (no handler)--function=" , req.Data.Method.Function, ",argv=", req.Data.Argv)

		res.Data.Err = data.ErrSrvInternalErr
		res.Data.ErrMsg = data.ErrSrvInternalErrText
	}

	return nil
}

// diaptch
func (ni *ServiceNode) Dispatch(req *data.UserRequestData, res *data.UserResponseData) error {
	var err error
	if ni.client != nil {
		err = nethelper.CallJRPCToTcpServerOnClient(ni.client, data.MethodCenterDispatch, req, res)
	}else{
		err = nethelper.CallJRPCToTcpServer(ni.ServiceCenterAddr, data.MethodCenterDispatch, req, res)
	}
	return err
}

// push
func (ni *ServiceNode) Push(req *data.UserResponseData, res *data.UserResponseData) error {
	var err error
	if ni.client != nil {
		err = nethelper.CallJRPCToTcpServerOnClient(ni.client, data.MethodCenterPush, req, res)
	}else{
		err = nethelper.CallJRPCToTcpServer(ni.ServiceCenterAddr, data.MethodCenterPush, req, res)
	}
	return err
}

// 内部方法
func (ni *ServiceNode)connectToCenter() (*Connection, error){
	var err error

	conn, err := net.Dial("tcp", ni.ServiceCenterAddr)
	if err != nil {
		log.Println("#connectToCenter Error: ", err.Error())
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
		err = ni.client.Call(data.MethodCenterRegister, ni.RegisterData, &res)
	}else{
		errors.New("connection is closed")
	}
	return err
}

func (ni *ServiceNode)unRegistToCenter() error{
	var err error
	var res string
	if ni.client != nil {
		err = nethelper.CallJRPCToTcpServerOnClient(ni.client, data.MethodCenterUnRegister, ni.RegisterData, &res)
	}else{
		err = nethelper.CallJRPCToTcpServer(ni.ServiceCenterAddr, data.MethodCenterUnRegister, ni.RegisterData, &res)
	}

	return err
}

func (ni *ServiceNode)startToServiceCenter(ctx context.Context) error {
	go func() {
		ni.wg.Add(1)
		defer ni.wg.Done()

		go func() {
			err := errors.New("not connect")
			var cn *Connection
			for {
				if err != nil {
					log.Println("Tcp client connect...")
					cn, err = ni.connectToCenter()
				}

				if err == nil {
					ni.client = rpc.NewClient(cn)

					ni.registToCenter()

					log.Println("Tcp client connected...")
					<-cn.Done
					log.Println("Tcp client close... ")

					err = errors.New("not connect")
				}

				time.Sleep(time.Second*5)
			}
		}()

		<-ctx.Done()
		ni.unRegistToCenter()
		fmt.Println("UnRegist to center ok...", ni.RegisterData)
	}()

	return nil
}
