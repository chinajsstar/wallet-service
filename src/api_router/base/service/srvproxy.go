package service

import (
	"../../data"
	"../nethelper"
	"net/rpc"
	"sync"
	"fmt"
	"sync/atomic"
)

// a service node
type SrvNode struct{
	registerData data.SrvRegisterData 	// register data

	rwmu sync.RWMutex					// read/write lock
	client *rpc.Client					// client
}

// try to connect service node
func (srvNode *SrvNode)connect() error{
	srvNode.rwmu.Lock()
	defer srvNode.rwmu.Unlock()

	if srvNode.client == nil{
		client, err := rpc.Dial("tcp", srvNode.registerData.Addr)
		if err != nil {
			fmt.Println("#Error connect srv node: ", err.Error())
			return err
		}

		srvNode.client = client
	}

	return nil
}

// close service node
func (srvNode *SrvNode)close() error{
	srvNode.rwmu.Lock()
	defer srvNode.rwmu.Unlock()

	if srvNode.client != nil{
		srvNode.client.Close()
		srvNode.client = nil
	}

	return nil
}

// call service node
func (srvNode *SrvNode)call(method string, params interface{}, res interface{}) error {
	srvNode.rwmu.RLock()
	defer srvNode.rwmu.RUnlock()

	return nethelper.CallJRPCToTcpServerOnClient(srvNode.client, method, params, res)
}

// service node group
type SrvNodeGroup struct{
	srv	string							// service name

	rwmu sync.RWMutex					// read/write lock
	addrMapSrvNode map[string]*SrvNode	// service node map

	index int64							// index for use
	nodes []*SrvNode					// service nodes [] for use
}

// register a service node
func (sng *SrvNodeGroup) RegisterNode(reg *data.SrvRegisterData) error {
	sng.rwmu.Lock()
	defer sng.rwmu.Unlock()

	if sng.addrMapSrvNode == nil {
		sng.index = 0
		sng.srv = reg.Srv
		sng.addrMapSrvNode = make(map[string]*SrvNode)
	}

	srvNode := sng.addrMapSrvNode[reg.Addr]
	if srvNode == nil {
		srvNode = &SrvNode{registerData:*reg, client:nil}
		sng.addrMapSrvNode[reg.Addr] = srvNode
	}

	sng.nodes = append(sng.nodes, srvNode)

	fmt.Println("srv-", sng.srv, ",register node-", reg.Addr, ",all-", len(sng.addrMapSrvNode))
	return nil
}

// unregister a service node
func (sng *SrvNodeGroup) UnRegisterNode(reg *data.SrvRegisterData) error {
	sng.rwmu.Lock()
	defer sng.rwmu.Unlock()

	if sng.addrMapSrvNode == nil{
		return nil
	}

	srvNode := sng.addrMapSrvNode[reg.Addr]
	delete(sng.addrMapSrvNode, reg.Addr)

	for i, v := range sng.nodes {
		if v == srvNode {
			sng.nodes = append(sng.nodes[:i], sng.nodes[i+1:]...)
			break
		}
	}

	if srvNode != nil {
		srvNode.close()
	}

	fmt.Println("srv-", sng.srv, ",unregister node-", reg.Addr, ",all-", len(sng.addrMapSrvNode))
	return nil
}

// dispatch a request to service node
func (sng *SrvNodeGroup) Dispatch(req *data.SrvRequestData, res *data.SrvResponseData) {
	sng.rwmu.RLock()
	defer sng.rwmu.RUnlock()

	// check has srv nodes
	if sng.addrMapSrvNode == nil || len(sng.addrMapSrvNode) == 0 {
		res.Data.Err = data.ErrNotFindSrv
		res.Data.ErrMsg = data.ErrNotFindSrvText
		return
	}

	// get a free srv node
	var srvNode *SrvNode
	srvNode = sng.getFreeNode()
	if srvNode == nil{
		res.Data.Err = data.ErrNotFindSrv
		res.Data.ErrMsg = data.ErrNotFindSrvText
		return
	}

	// check client is nil
	if srvNode.client == nil {
		srvNode.connect()
	}

	// call
	if srvNode.client != nil {
		err := srvNode.call(data.MethodNodeCall, req, res)
		if err != nil {
			fmt.Println("#Call srv failed...", err)

			srvNode.close()

			res.Data.Err = data.ErrCallFailed
			res.Data.ErrMsg = data.ErrCallFailedText
		}
	}else{
		res.Data.Err = data.ErrConnectSrvFailed
		res.Data.ErrMsg = data.ErrConnectSrvFailedText
	}

	return
}

// get a free node by index
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