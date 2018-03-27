package business

import (
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/context"
	"sync"
)

func NewBusinessSvr() *Business {
	return new(Business)
}

type Business struct {
	wallet    *service.ClientManager
	ctx       context.Context
	cancel    context.CancelFunc
	waitGroup *sync.WaitGroup
}

func (self *Business) Init() error {
	self.ctx, self.cancel = context.WithCancel(context.Background())
	self.waitGroup = new(sync.WaitGroup)
	self.wallet = new(service.ClientManager)

	//实例化以太坊客户端
	client, err := eth.NewClient()
	if err != nil {
		fmt.Printf("create client:%s error:%s\n", types.Chain_eth, err.Error())
		return err
	}
	self.wallet.AddClient(client)

	return nil
}

func (self *Business) Start() {
	// 创建监控充币地址channael
	self.waitGroup.Add(1)
	defer self.waitGroup.Done()

	rctChannel := make(types.RechargeTxChannel)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		for {
			select {
			case rct := <-channel:
				{
					fmt.Printf("Recharge Transaction : cointype:%s, information:%s.", rct.Coin_name, rct.Tx.String())
				}
			case <-ctx.Done():
				{
					fmt.Printf("Business正常退出")
					return
				}
			}
		}
	}(self.ctx, rctChannel)

	/*********开启服务!!!!!*********/
	ctx, _ := context.WithCancel(self.ctx)
	go self.wallet.Start(ctx, rctChannel)
}

func (self *Business) Stop() {
	self.cancel()
	self.waitGroup.Wait()
}

func (self *Business) HandleMsg(args string, reply *string) error {
	var head ReqHead
	err := json.Unmarshal([]byte(args), &head)
	if err != nil {
		fmt.Printf("HandleMsg Unmarshal Error: %s/n", err.Error())
		return err
	}

	replyChan := make(chan string)
	switch head.Method {
	case "new_address":
		go self.HandleNewAddress(args, replyChan)
	case "withdrawal":
		go self.HandleWithdrawal(args, replyChan)
		*reply = args
	default:
		return errors.New("invalid command")
	}
	*reply = <-replyChan
	return nil
}

func (self *Business) HandleNewAddress(args string, replyChan chan string) error {
	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Printf("HandleNewAddress Redis Error: %s", err.Error())
		return err
	}
	defer c.Close()

	req := new(ReqNewAddress)
	err = json.Unmarshal([]byte(args), req)
	if err != nil {
		fmt.Printf("HandleNewAddress Json Unmarshal Error:%s", err.Error())
		return err
	}

	var symbol string
	switch req.Params.Symbol {
	case types.Chain_eth:
		symbol = types.Chain_eth
	case types.Chain_bitcoin:
		symbol = types.Chain_bitcoin
	default:
		err = errors.New("No Symble " + req.Params.Symbol)
		return err
	}

	//判断剩余的地址
	sum, err := redis.Int(c.Do("scard", symbol+"_unuse_addr"))
	if err != nil {
		fmt.Printf("%s", err.Error())
		return err
	}

	if sum < 500 {
		//需要去请求地址了
		accs, err := self.getAddress(symbol, 999-uint32(sum))
		if err != nil {
			fmt.Printf("%s", err.Error())
			return err
		}

		for _, account := range accs {
			acc, err := json.Marshal(account)
			if err != nil {
				return err
			}
			c.Do("sadd", symbol+"_unuse_addr", acc)
		}
		sum, err = redis.Int(c.Do("scard", symbol+"_unuse_addr"))
	}

	if sum < req.Params.Count {
		return errors.New("地址不够用了")
	}

	accs, _ := redis.Strings(c.Do("spop", symbol+"_unuse_addr", req.Params.Count))

	/*********添加监控地址示例*********/
	var acc types.Account
	json.Unmarshal([]byte(accs[0]), &acc)

	var addresses []string
	addresses = append(addresses, acc.Address[2:])
	rcaCmd := service.NewRechargeAddressCmd("message id", symbol, addresses)
	self.wallet.SetRechargeAddress(rcaCmd)

	rsp := new(RspNewAddress)
	rsp.Result.ID = req.UserID
	rsp.Result.Symbol = req.Params.Symbol
	for _, v := range accs {
		var acc types.Account
		json.Unmarshal([]byte(v), &acc)
		rsp.Result.Address = append(rsp.Result.Address, acc.Address)
	}
	rsp.Status.Code = 0
	rsp.Status.Msg = ""

	reply, _ := json.Marshal(rsp)

	replyChan <- string(reply)

	return nil
}

func (self *Business) HandleWithdrawal(args string, replyChan chan string) error {
	req := new(ReqWithdrawal)
	err := json.Unmarshal([]byte(args), req)
	if err != nil {
		fmt.Printf("HandleWithdrawal Json Unmarshal Error:%s", err.Error())
		return err
	}

	rsp := new(RspWithdrawal)
	rsp.Result.UserOrderID = req.Params.UserOrderID
	rsp.Result.Timestamp = "0"
	rsp.Status.Code = 0
	rsp.Status.Msg = ""

	reply, _ := json.Marshal(rsp)

	replyChan <- string(reply)

	return nil
}

func (self *Business) getAddress(symbol string, count uint32) ([]*types.Account, error) {
	accCmd := service.NewAccountCmd("message id", symbol, uint32(count))
	return self.wallet.NewAccounts(accCmd)

}
