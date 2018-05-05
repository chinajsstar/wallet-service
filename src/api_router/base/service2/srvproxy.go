package service2

import (
	"api_router/base/data"
	"sync"
	"sync/atomic"
	l4g "github.com/alecthomas/log4go"
	"github.com/cenkalti/rpc2"
)

// service node group
type SrvNodeGroup struct{
	registerData data.SrvRegisterData 	// register data

	rwmu sync.RWMutex					// read/write lock
	index int64							// index for use
	nodes []*rpc2.Client					// service nodes [] for use
}

// register a service node
func (sng *SrvNodeGroup) RegisterNode(client *rpc2.Client, reg *data.SrvRegisterData) error {
	sng.rwmu.Lock()
	defer sng.rwmu.Unlock()

	sng.registerData = *reg

	sng.nodes = append(sng.nodes, client)

	l4g.Debug("reg-%s, all-%d", reg.String(), len(sng.nodes))
	return nil
}

// unregister a service node
func (sng *SrvNodeGroup) UnRegisterNode(client *rpc2.Client) error {
	sng.rwmu.Lock()
	defer sng.rwmu.Unlock()

	for i, v := range sng.nodes {
		if v == client {
			sng.nodes = append(sng.nodes[:i], sng.nodes[i+1:]...)
			break
		}
	}

	l4g.Debug("unreg-%s, all-%d", sng.registerData.String(), len(sng.nodes))
	return nil
}

func (sng *SrvNodeGroup) GetSrvInfo() (data.SrvRegisterData) {
	sng.rwmu.RLock()
	defer sng.rwmu.RUnlock()

	return sng.registerData
}

func (sng *SrvNodeGroup) GetSrvNodes() (int) {
	sng.rwmu.RLock()
	defer sng.rwmu.RUnlock()

	return len(sng.nodes)
}

func (sng *SrvNodeGroup) Call(req *data.SrvRequestData, res *data.SrvResponseData) {
	sng.rwmu.RLock()
	defer sng.rwmu.RUnlock()

	// get a free srv node
	node := sng.getFreeNode()
	if node == nil{
		res.Data.Err = data.ErrNotFindSrv
		return
	}

	// call node
	err := node.Call(data.MethodNodeCall, req, res)
	if err != nil {
		l4g.Error("#Call srv:%s", err.Error())

		res.Data.Err = data.ErrCallFailed
		return
	}
}

// get a free node by index
func (sng *SrvNodeGroup) getFreeNode() *rpc2.Client {
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