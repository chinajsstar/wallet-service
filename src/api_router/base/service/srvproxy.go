package service

import (
	"api_router/base/data"
	"api_router/base/nethelper"
	"net/rpc"
	"sync"
	"sync/atomic"
	"errors"
	l4g "github.com/alecthomas/log4go"
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
		client, err := rpc.Dial("tcp", "")
		if err != nil {
			l4g.Error("connect srv node: %s", err.Error())
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

	l4g.Debug("reg-%s, all-%d", reg.String(), len(sng.addrMapSrvNode))
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

	l4g.Debug("unreg-%s, all-%d", reg.String(), len(sng.addrMapSrvNode))
	return nil
}

// list all service node
func (sng *SrvNodeGroup) ListSrv(nodes *[]data.SrvRegisterData) {
	sng.rwmu.RLock()
	defer sng.rwmu.RUnlock()

	if sng.addrMapSrvNode == nil{
		return
	}

	for _, v := range sng.addrMapSrvNode {
		*nodes = append(*nodes, v.registerData)
	}
}

// dispatch a request to service node
func (sng *SrvNodeGroup) Dispatch(req *data.SrvRequestData, res *data.SrvResponseData) (string,error) {
	sng.rwmu.RLock()
	defer sng.rwmu.RUnlock()

	nodeAddr := ""

	// check has srv nodes
	if sng.addrMapSrvNode == nil || len(sng.addrMapSrvNode) == 0 {
		res.Data.Err = data.ErrNotFindSrv
		return nodeAddr, errors.New(res.Data.ErrMsg)
	}

	// get a free srv node
	var srvNode *SrvNode
	srvNode = sng.getFreeNode()
	if srvNode == nil{
		res.Data.Err = data.ErrNotFindSrv
		return nodeAddr, errors.New(res.Data.ErrMsg)
	}

	// share on a client
	// check client is nil
	if srvNode.client == nil {
		srvNode.connect()
	}

	// call
	var err error
	if srvNode.client != nil {
		err = srvNode.call(data.MethodNodeCall, req, res)
		if err != nil {
			l4g.Error("#Call srv:%s", err.Error())

			res.Data.Err = data.ErrCallFailed

			nodeAddr = srvNode.registerData.Addr
			err = errors.New(res.Data.ErrMsg)
		}
	}else{
		res.Data.Err = data.ErrConnectSrvFailed

		nodeAddr = srvNode.registerData.Addr
		err = errors.New(res.Data.ErrMsg)
	}

	return nodeAddr, err
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