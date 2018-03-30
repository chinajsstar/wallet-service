package service

import (
	"../../data"
	"../nethelper"
	"net/rpc"
	"sync"
	"fmt"
	"time"
	"sync/atomic"
)

type SrvNode struct{
	RegisterData data.ServiceCenterRegisterData

	Rwmu sync.RWMutex
	Client *rpc.Client

	LastOperationTime time.Time
}

// 内部方法
func (srvNode *SrvNode)openClient() error{
	srvNode.Rwmu.Lock()
	defer srvNode.Rwmu.Unlock()

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

func (srvNode *SrvNode)closeClient() error{
	srvNode.Rwmu.Lock()
	defer srvNode.Rwmu.Unlock()

	if srvNode.Client != nil{
		srvNode.Client.Close()
		srvNode.Client = nil
	}

	return nil
}

func (srvNode *SrvNode)sendData(method string, params interface{}, res interface{}) error {
	srvNode.Rwmu.RLock()
	defer srvNode.Rwmu.RUnlock()

	return nethelper.CallJRPCToTcpServerOnClient(srvNode.Client, method, params, res)
}

type SrvNodeGroup struct{
	Srv	string
	Rwmu sync.RWMutex

	AddrMapSrvNode map[string]*SrvNode

	index int64
	nodes []*SrvNode
}

func (sng *SrvNodeGroup) RegisterNode(reg *data.ServiceCenterRegisterData) error {
	sng.Rwmu.Lock()
	defer sng.Rwmu.Unlock()

	if sng.AddrMapSrvNode == nil {
		sng.index = 0
		sng.Srv = reg.Srv
		sng.AddrMapSrvNode = make(map[string]*SrvNode)
	}

	if sng.AddrMapSrvNode[reg.Addr] == nil {
		sng.AddrMapSrvNode[reg.Addr] = &SrvNode{RegisterData:*reg, Client:nil}
	}

	sng.nodes = append(sng.nodes, sng.AddrMapSrvNode[reg.Addr])

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

	for i, v := range sng.nodes {
		if v == srvNode {
			sng.nodes = append(sng.nodes[:i], sng.nodes[i+1:]...)
		}
	}

	if srvNode != nil {
		srvNode.closeClient()
	}
	delete(sng.AddrMapSrvNode, reg.Addr)


	fmt.Println("srv-", sng.Srv, ",unregister node-", reg.Addr, ",all-", len(sng.AddrMapSrvNode))
	return nil
}

func (sng *SrvNodeGroup) Dispatch(req *data.SrvDispatchData, ack *data.SrvDispatchAckData) error {
	sng.Rwmu.RLock()
	defer sng.Rwmu.RUnlock()

	if sng.AddrMapSrvNode == nil || len(sng.AddrMapSrvNode) == 0 {
		ack.SrvAck.Err = data.ErrNotFindSrv
		ack.SrvAck.ErrMsg = data.ErrNotFindSrvText
		return nil
	}

	var srvNode *SrvNode
	srvNode = sng.getNode()
	if srvNode == nil{
		ack.SrvAck.Err = data.ErrNotFindSrv
		ack.SrvAck.ErrMsg = data.ErrNotFindSrvText
		return nil
	}

	// 检查是否连接
	if srvNode.Client == nil {
		srvNode.openClient()
	}

	// 发送数据
	if srvNode.Client != nil {
		err := srvNode.sendData(data.MethodServiceNodeCall, req, ack)
		if err != nil {
			fmt.Println("#Call srv failed")

			srvNode.closeClient()

			ack.SrvAck.Err = data.ErrCall
			ack.SrvAck.ErrMsg = err.Error()
		}
	}else{
		ack.SrvAck.Err = data.ErrClientConn
		ack.SrvAck.ErrMsg = data.ErrClientConnText
	}

	return nil
}

func (sng *SrvNodeGroup) getNode() *SrvNode {
	// TODO:根据算法获取空闲的
	// NOTE:go map 多次range会从随机位置开始迭代
	/*
		for _, v := range sng.AddrMapSrvNode{
		srvNode = v
		break
	}
	 */
	length := int64(len(sng.nodes))
	if length == 0 {
		return nil
	}

	atomic.AddInt64(&sng.index, 1)
	atomic.CompareAndSwapInt64(&sng.index, length, 0)

	index := sng.index % length
	return sng.nodes[index]
}

func (sng *SrvNodeGroup)KeepAlive() {
	// 是否有断开连接
	var rgQuit []data.ServiceCenterRegisterData

	func(){
		sng.Rwmu.RLock()
		defer sng.Rwmu.RUnlock()

		var res string
		for _, b := range sng.AddrMapSrvNode{
			if b.Client != nil{
				b.LastOperationTime = time.Now()
				res = ""
				err := b.sendData(data.MethodServiceNodePingpong, "ping", &res)
				if err != nil || res != "pong" {
					rgQuit = append(rgQuit, b.RegisterData)
				}
			}
		}
	}()

	// 去掉断开的
	for _, v := range rgQuit {
		sng.UnRegisterNode(&v)
	}
}