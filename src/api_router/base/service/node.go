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
	// 中心
	client *rpc.Client
}

// 生成一个服务节点
func NewServiceNode(srvName string, versionName string) (*ServiceNode, error){
	serviceNode := &ServiceNode{}

	serviceNode.RegisterData.Srv = srvName
	serviceNode.RegisterData.Version = versionName

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
					go rpc.ServeConn(conn)
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
		fmt.Println("Error function call (no handler)--function=" , req.Data.Function, ",argv=", req.Data.Argv)

		res.Data.Err = data.ErrSrvInternalErr
		res.Data.ErrMsg = data.ErrSrvInternalErrText
	}

	return nil
}

// 服务节点RPC--与服务中心心跳
func (ni *ServiceNode) Pingpong(req *string, res * string) error {
	if *req == "ping" {
		*res = "pong"
	}else{
		*res = *req
	}
	return nil
}

// 内部方法
func (ni *ServiceNode)registToCenter() error{
	if ni.client != nil {
		ni.client.Close()
		ni.client = nil
	}

	var err error
	ni.client, err = rpc.Dial("tcp", ni.ServiceCenterAddr)
	if err != nil {
		log.Println("#registToCenter Error: ", err.Error())
		return err
	}

	var res string
	err = nethelper.CallJRPCToTcpServerOnClient(ni.client, data.MethodCenterRegister, ni.RegisterData, &res)
	fmt.Println("Regist to center...", ni.RegisterData, ",error--", err)
	if err != nil {
		ni.client.Close()
		ni.client = nil
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

func (ni *ServiceNode)doPingpong() error{
	var err error
	var res string

	if ni.client != nil {
		err = nethelper.CallJRPCToTcpServerOnClient(ni.client, data.MethodCenterPingpong, "ping", &res)
	}else{
		err = errors.New("client is close")
	}
	return err
}

func (ni *ServiceNode)startToServiceCenter(ctx context.Context) error {
	go func() {
		ni.wg.Add(1)
		defer ni.wg.Done()

		go func() {
			err := ni.registToCenter()
			for {
				if err == nil {
					time.Sleep(time.Second*60)
				}else{
					time.Sleep(time.Second*5)
				}

				if err == nil {
					err = ni.doPingpong()
				}else{
					err = ni.registToCenter()
				}

				fmt.Println("keepalive...sleep")
			}
		}()

		<-ctx.Done()
		ni.unRegistToCenter()
		fmt.Println("UnRegist to center ok...", ni.RegisterData)
	}()

	return nil
}
