package eth

import (
	"sync"
	"strings"
	"github.com/ethereum/go-ethereum/ethclient"
	"time"
	"golang.org/x/net/context"
	"github.com/ethereum/go-ethereum/common"
)

// Nonces 类, 专门用于管理以太坊对应地址的 nonce
// Nonces.nextNonce用于返回下一次发送交易可以使用的nonce,
// 当用于在 maxUpdateSecond 定义的时间(单位:秒)内请求nonce
// 会直接返回Nonces[from].nonce, 然后把Nonces[from].nonce+1
// 如果超过了 maxUpdateSecond 的时间间隔, Nonces 会调用refreshNonce,
// 从钱包节点重新获取新的nonce来设置Nonces[from].nonce,
// 然后把Nonces[from].nonce+1

const (
	maxUpdateSecond = 60
	hour = time.Second * 60 * 60
)

type Nonce struct {
	nonce    uint64
	lastTime time.Time
}

func (self *Nonce) refreshNonce(client * ethclient.Client,
	from string) (err error) {
	var nonce uint64
	address := common.HexToAddress(from)
	nonce, err = client.PendingNonceAt(context.TODO(), address)
	if err!=nil {
		return
	}
	self.lastTime = time.Now()
	self.nonce = nonce
	return
}

func (self *Nonce) lastRefreshTimeDiffNow() float64 {
	return time.Now().Sub(self.lastTime).Seconds()
}

type Nonces struct {
	nonceMap map[string]*Nonce
	mutx     *sync.Mutex
	quit     chan bool
}

func newNonces() *Nonces {
	return &Nonces{
		nonceMap: make(map[string]*Nonce),
		mutx:     new(sync.Mutex),
	}
}

func (self *Nonces)stop() {
	self.quit <- true
}

// 定期清理 map内 长时间不使用的元素
func (self *Nonces) loopUnusedNonce() {

	L4g.Trace("begin unusedNonceloop......")
	defer L4g.Trace("end unusedNonceloop......")

	exit_for:
	for {
		select {
		case <-time.After(hour): {
			L4g.Info("start clear unused nonces....")
			self.mutx.Lock()
			for from, theNonce := range self.nonceMap {
				if theNonce.lastRefreshTimeDiffNow() < hour.Seconds() {
					continue
				}
				delete(self.nonceMap, from)
			}
			self.mutx.Unlock()
			L4g.Info("clear unused nonces complete....")
		}
		case <- self.quit:{
			break exit_for
		}
		}
	}
}

func (self *Nonces)nextNonce(client *ethclient.Client, from string) (nonce uint64, err error) {
	from = strings.ToLower(from)

	self.mutx.Lock()
	defer self.mutx.Unlock()

	theNonce := self.nonceMap[from]

	if nil== theNonce {
		theNonce = &Nonce{
			nonce:0,
			lastTime:time.Time{},
		}
		self.nonceMap[from] = theNonce
	}

	if theNonce.lastRefreshTimeDiffNow() >= maxUpdateSecond {
		err = theNonce.refreshNonce(client, from);
		if err!=nil {
			return
		}
	}
	nonce = theNonce.nonce
	theNonce.nonce++
	return
}
