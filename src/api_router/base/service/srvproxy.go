package service

import (
	"../../data"
	"../nethelper"
	"net/rpc"
	"sync"
	"fmt"
	"sync/atomic"
)

type SrvNode struct{
	RegisterData data.SrvRegisterData

	Rwmu sync.RWMutex
	Client *rpc.Client
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

//////////////////////////////////////////////////////////////////
type SrvNodeGroup struct{
	Srv	string
	Rwmu sync.RWMutex

	AddrMapSrvNode map[string]*SrvNode

	index int64
	nodes []*SrvNode
}

func (sng *SrvNodeGroup) RegisterNode(reg *data.SrvRegisterData) error {
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

func (sng *SrvNodeGroup) UnRegisterNode(reg *data.SrvRegisterData) error {
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

func (sng *SrvNodeGroup) Dispatch(req *data.SrvRequestData, res *data.SrvResponseData) {
	sng.Rwmu.RLock()
	defer sng.Rwmu.RUnlock()

	if sng.AddrMapSrvNode == nil || len(sng.AddrMapSrvNode) == 0 {
		res.Data.Err = data.ErrNotFindSrv
		res.Data.ErrMsg = data.ErrNotFindSrvText
		return
	}

	var srvNode *SrvNode
	srvNode = sng.getFreeNode()
	if srvNode == nil{
		res.Data.Err = data.ErrNotFindSrv
		res.Data.ErrMsg = data.ErrNotFindSrvText
		return
	}

	// 检查是否连接
	if srvNode.Client == nil {
		srvNode.openClient()
	}

	// 发送数据
	if srvNode.Client != nil {
		err := srvNode.sendData(data.MethodNodeCall, req, res)
		if err != nil {
			fmt.Println("#Call srv failed...", err)

			srvNode.closeClient()

			res.Data.Err = data.ErrCall
			res.Data.ErrMsg = data.ErrCallText
		}
	}else{
		res.Data.Err = data.ErrClientConn
		res.Data.ErrMsg = data.ErrClientConnText
	}

	return
}

func (sng *SrvNodeGroup) getFreeNode() *SrvNode {
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