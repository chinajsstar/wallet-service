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
type CallNodeApi func(req *data.ServiceCenterDispatchData, ack *data.ServiceCenterDispatchAckData)

// 服务节点信息
type ServiceNode struct{
	// 注册的信息
	RegisterData data.ServiceCenterRegisterData
	// 回掉
	Handler CallNodeApi
	// 服务中心
	ServiceCenterAddr string
	// 等待
	wg *sync.WaitGroup
}

// 生成一个服务节点
func NewServiceNode(serviceName string, serviceVersion string) (*ServiceNode, error){
	serviceNode := &ServiceNode{}

	serviceNode.RegisterData.Name = serviceName
	serviceNode.RegisterData.Version = serviceVersion

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

	err = ni.startToServiceCenter(ctx)

	return err
}

// RPC 方法
// 服务节点RPC--调用节点方法ServiceNodeInstance.Call
func (ni *ServiceNode) Call(req *data.ServiceCenterDispatchData, ack *data.ServiceCenterDispatchAckData) error {
	ack.Err = data.ServiceDispatchErrOk
	ack.ErrMsg = ""
	if ni.Handler != nil {
		ni.Handler(req, ack)
	}else{
		fmt.Println("Error api call (no handler)--api=" , req.Api, ",argv=", req.Argv)

		ack.Err = data.ServiceDispatchErrNotFindHanlder
		ack.ErrMsg = "Not find handler"
	}

	return nil
}

const(
	ToCenterStatusOk = 0
	ToCenterStatusCLose = 1
	ToCenterStatusStop = 2
)
// 内部方法
func (ni *ServiceNode)keepAlive(status int) int{
	var err error
	var res string
	if status == ToCenterStatusCLose {
		err = nethelper.CallJRPCToTcpServer(ni.ServiceCenterAddr, data.MethodServiceCenterRegister, ni.RegisterData, &res)
		if err == nil {
			status = ToCenterStatusOk
		}
	}

	if status == ToCenterStatusOk{
		err = nethelper.CallJRPCToTcpServer(ni.ServiceCenterAddr, data.MethodServiceCenterPingpong, "ping", &res)
		if err == nil && res == "pong" {
			status = ToCenterStatusOk
		}else{
			status = ToCenterStatusCLose
		}
	}

	return status
}

func (ni *ServiceNode)startToServiceCenter(ctx context.Context) error{
	timeout := make(chan bool)
	go func(){
		for ; ; {
			timeout <- true
			time.Sleep(time.Second*10)
		}
	}()

	go func() {
		status := ToCenterStatusCLose
		for ; ; {
			select{
			case <-ctx.Done():
				status = ToCenterStatusStop
			case <-timeout:
				status = ni.keepAlive(status)
			}

			if status == ToCenterStatusStop{
				break
			}
		}
	}()

	return  nil
}
