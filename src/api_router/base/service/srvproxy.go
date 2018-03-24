package service

import (
	"../../data"
	"../nethelper"
	"net/rpc"
	"sync"
	"fmt"
	"time"
)

type SrvNode struct{
	RegisterData data.ServiceCenterRegisterData

	Client *rpc.Client

	LastOperationTime time.Time
}

type SrvNodeGroup struct{
	Srv	string
	Rwmu sync.RWMutex
	AddrMapSrvNode map[string]*SrvNode
}

func (sng *SrvNodeGroup) RegisterNode(reg *data.ServiceCenterRegisterData) error {
	sng.Rwmu.Lock()
	defer sng.Rwmu.Unlock()

	if sng.AddrMapSrvNode == nil {
		sng.Srv = reg.Srv
		sng.AddrMapSrvNode = make(map[string]*SrvNode)
	}

	if sng.AddrMapSrvNode[reg.Addr] == nil {
		sng.AddrMapSrvNode[reg.Addr] = &SrvNode{RegisterData:*reg, Client:nil}
	}

	time.Now()

	fmt.Println("srv-", sng.Srv, ",register node-", reg.Addr, ",all-", len(sng.AddrMapSrvNode))
	return nil
}

func (sng *SrvNodeGroup) UnRegisterNode(reg *data.ServiceCenterRegisterData) error {
	sng.Rwmu.Lock()
	defer sng.Rwmu.Unlock()

	if sng.AddrMapSrvNode == nil{
		return nil
	}

	srvNode := sng.AddrMapSrvNode[reg.Addr]
	if srvNode != nil {
		sng.closeClient(srvNode)
	}
	delete(sng.AddrMapSrvNode, reg.Addr)

	fmt.Println("srv-", sng.Srv, ",unregister node-", reg.Addr, ",all-", len(sng.AddrMapSrvNode))
	return nil
}

func (sng *SrvNodeGroup) Dispatch(req *data.ServiceCenterDispatchData, ack *data.ServiceCenterDispatchAckData) error {
	if sng.AddrMapSrvNode == nil || len(sng.AddrMapSrvNode) == 0 {
		ack.Err = data.ServiceDispatchErrServiceStop
		ack.ErrMsg = "No service online"
		return nil
	}

	// TODO:根据算法获取空闲的
	// NOTE:go map 多次range会从随机位置开始迭代
	var srvNode *SrvNode
	for _, v := range sng.AddrMapSrvNode{
		srvNode = v
		break
	}
	if srvNode == nil{
		return nil
	}

	// 检查是否连接
	if srvNode.Client == nil {
		sng.openClient(srvNode)
	}

	func() {
		if srvNode.Client != nil {
			sng.Rwmu.RLock()
			defer sng.Rwmu.RUnlock()

			srvNode.LastOperationTime = time.Now()
			if nil != nethelper.CallJRPCToTcpServerOnClient(srvNode.Client, data.MethodServiceNodeCall, req, ack){
				fmt.Println("#Call versionApi failed, close client")

				sng.closeClient(srvNode)

				ack.Err = data.ServiceDispatchErrNotFindApi
				ack.ErrMsg = "Others error"
			}
		}else{
			ack.Err = data.ServiceDispatchErrServiceStop
			ack.ErrMsg = "Service is stop"
		}
	}()

	return nil
}

func (sng *SrvNodeGroup)KeepAlive() {
	sng.Rwmu.RLock()
	defer sng.Rwmu.RUnlock()

	var res string
	for _, b := range sng.AddrMapSrvNode{
		if b.Client != nil{
			b.LastOperationTime = time.Now()
			res = ""
			err := nethelper.CallJRPCToTcpServerOnClient(b.Client, data.MethodServiceNodePingpong, "ping", &res)
			if err != nil || res != "pong" {
				// close this client
				fmt.Println("#keep alive failed, remove...")
				sng.UnRegisterNode(&b.RegisterData)
				break
			}
		}
	}
}

// 内部方法
func (sng *SrvNodeGroup)openClient(srvNode *SrvNode) error{
	sng.Rwmu.Lock()
	defer sng.Rwmu.Unlock()

	if srvNode.Client == nil{
		client, err := rpc.Dial("tcp", srvNode.RegisterData.Addr)
		if err != nil {
			fmt.Println("Error Open client: ", err.Error())
			return err
		}

		srvNode.LastOperationTime = time.Now()
		srvNode.Client = client
	}

	return nil
}

func (sng *SrvNodeGroup)closeClient(srvNode *SrvNode) error{
	//sng.Rwmu.Lock()
	//defer sng.Rwmu.Unlock()

	if srvNode.Client != nil{
		srvNode.Client.Close()
		srvNode.Client = nil
	}

	return nil
}