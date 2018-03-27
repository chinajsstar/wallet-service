package business

import (
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"sync"
)

func NewBusinessSvr() *Business {
	return new(Business)
}

type Business struct {
	rechargeChannel types.RechargeTxChannel
	txStateChannel  types.TxStateChange_Channel
	wallet          *service.ClientManager
	ctx             context.Context
	cancel          context.CancelFunc
	waitGroup       *sync.WaitGroup
}

func (busi *Business) InitAndStart() error {
	busi.rechargeChannel = make(types.RechargeTxChannel)
	busi.txStateChannel = make(types.TxStateChange_Channel)
	busi.ctx, busi.cancel = context.WithCancel(context.Background())
	busi.waitGroup = new(sync.WaitGroup)
	busi.wallet = new(service.ClientManager)

	//实例化以太坊客户端
	client, err := eth.NewClient(busi.rechargeChannel)
	if err != nil {
		fmt.Printf("create client:%s error:%s\n", types.Chain_eth, err.Error())
		return err
	}
	busi.wallet.AddClient(client)
	busi.wallet.SubscribeTxStateChange(busi.txStateChannel)

	busi.recvRechargeTxChannel()
	busi.recvTxStateChannel()
	busi.startWalletSever()

	return nil
}

func (busi *Business) recvRechargeTxChannel() {
	busi.waitGroup.Add(1)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		for {
			select {
			case rct := <-channel:
				{
					fmt.Printf("Recharge Transaction : cointype:%s, information:%s.", rct.Coin_name, rct.Tx.String())
				}
			case <-ctx.Done():
				{
					fmt.Println("RechangeTx context done, because : ", ctx.Err())
					busi.waitGroup.Done()
					return
				}
			}
		}
	}(busi.ctx, busi.rechargeChannel)
}

func (busi *Business) recvTxStateChannel() {
	busi.waitGroup.Add(1)
	go func(ctx context.Context, channel types.TxStateChange_Channel) {
		for {
			select {
			case cmdTx := <-channel:
				{
					fmt.Printf("Transaction state changed, transaction information:%s\n",
						cmdTx.Tx.String())

					if cmdTx.Tx.State == types.Tx_state_confirmed {
						fmt.Println("Transaction is confirmed! success!!!")
					}

					if cmdTx.Tx.State == types.Tx_state_unconfirmed {
						fmt.Println("Transaction is unconfirmed! failed!!!!")
					}
				}
			case <-ctx.Done():
				fmt.Println("TxState context done, because : ", ctx.Err())
				busi.waitGroup.Done()
				return
			}
		}
	}(busi.ctx, busi.txStateChannel)
}

func (busi *Business) startWalletSever() {
	busi.wallet.Start()
}

func (busi *Business) Stop() {
	busi.cancel()
	busi.waitGroup.Wait()
}

func (busi *Business) HandleMsg(args string, reply *string) error {
	var head ReqHead
	err := json.Unmarshal([]byte(args), &head)
	if err != nil {
		fmt.Printf("HandleMsg Unmarshal Error: %s/n", err.Error())
		return err
	}

	replyChan := make(chan string)
	switch head.Method {
	case "new_address":
		{
			go busi.HandleNewAddress(args, replyChan)
		}
	case "withdrawal":
		{
			go busi.HandleWithdrawal(args, replyChan)
		}
		*reply = args
	default:
		return errors.New("invalid command")
	}
	*reply = <-replyChan
	return nil
}

func (busi *Business) HandleNewAddress(args string, replyChan chan string) error {
	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Printf("handle new address redis error: %s", err.Error())
		return err
	}
	defer c.Close()

	req := new(ReqNewAddress)
	err = json.Unmarshal([]byte(args), req)
	if err != nil {
		fmt.Printf("handle new address json Unmarshal Error:%s", err.Error())
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
		accs, err := busi.getAddress(symbol, 999-uint32(sum))
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
	addresses = append(addresses, acc.Address)
	rcaCmd := types.NewRechargeAddressCmd("message id", symbol, addresses)
	busi.wallet.SetRechargeAddress(rcaCmd)
	fmt.Println(acc.Address)

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

func (busi *Business) HandleWithdrawal(args string, replyChan chan string) error {
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

func (busi *Business) getAddress(symbol string, count uint32) ([]*types.Account, error) {
	accCmd := types.NewAccountCmd("message id", symbol, uint32(count))
	return busi.wallet.NewAccounts(accCmd)

}
