package address

import (
	"blockchain_server/service"
	"blockchain_server/types"
	"business_center/def"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type Address struct {
	wallet        *service.ClientManager
	symbolNameMap map[string]string
}

func (addr *Address) Init(wallet *service.ClientManager) {
	addr.wallet = wallet
	addr.symbolNameMap = make(map[string]string)
	addr.symbolNameMap[types.Chain_eth] = types.Chain_eth
	addr.symbolNameMap[types.Chain_bitcoin] = types.Chain_bitcoin
}

func (addr *Address) AllocationAddress(req string, ack *string) error {
	jsonInfo := &def.ReqNewAddress{}
	err := json.Unmarshal([]byte(req), jsonInfo)
	if err != nil {
		fmt.Printf("AllocationAddress Unmarshal Error : %s/n", err.Error())
		return err
	}

	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Printf("AllocationAddress Redis Dial Error: %s", err.Error())
		return err
	}
	defer c.Close()

	symbolName, ok := addr.symbolNameMap[jsonInfo.Params.Symbol]
	if !ok {
		err := errors.New("AllocationAddress SymbolName Invalid")
		fmt.Println(err.Error())
		return err
	}

	//查询剩余的地址
	idleNumber, err := redis.Int(c.Do("scard", symbolName+"_unuse_addr"))
	if err != nil {
		fmt.Printf("AllocationAddress Redis Scard Error: %s\n", err.Error())
		return err
	}

	//补充地址
	if idleNumber < 500 {
		err := addr.generateAddress(c, symbolName, 1000-idleNumber)
		if err != nil {
			return err
		}
	}

	accounts, err := redis.Strings(c.Do("spop", symbolName+"_unuse_addr", jsonInfo.Params.Count))
	if err != nil {
		fmt.Printf("AllocationAddress Redis SPOP Error : %s/n", err.Error())
		return err
	}

	var addrs []string
	for _, v := range accounts {
		acc := &types.Account{}
		err := json.Unmarshal([]byte(v), acc)
		if err != nil {
			fmt.Printf("AllocationAddress Accounts Unmarshal Error : %s/n", err.Error())
			return nil
		}
		addrs = append(addrs, acc.Address)
	}

	if len(addrs) > 0 {
		rcaCmd := types.NewRechargeAddressCmd("message id", symbolName, addrs)
		addr.wallet.InsertRechargeAddress(rcaCmd)
	}

	rsp := new(def.RspNewAddress)
	rsp.Result.ID = jsonInfo.UserID
	rsp.Result.Symbol = jsonInfo.Params.Symbol
	rsp.Result.Address = addrs
	rsp.Status.Code = 0
	rsp.Status.Msg = ""

	byteRsp, err := json.Marshal(rsp)
	if err != nil {
		fmt.Printf("AllocationAddress RspNewAddress Marshal Error : %s/n", err.Error())
		return err
	}

	*ack = string(byteRsp)
	return nil
}

func (addr *Address) generateAddress(c redis.Conn, symbolName string, count int) error {
	accCmd := types.NewAccountCmd("message id", symbolName, 1)

	for i := 0; i < count; i++ {
		accounts, err := addr.wallet.NewAccounts(accCmd)
		if err != nil {
			fmt.Printf("generateAddress NewAccounts Error : %s\n", err.Error())
			return err
		}
		jsonInfo, err := json.Marshal(accounts[0])
		if err != nil {
			fmt.Printf("generateAddress Marshal Error : %s\n", err.Error())
			return err
		}
		c.Do("sadd", symbolName+"_unuse_addr", jsonInfo)
	}
	return nil
}

//func (busi *Business) HandleWithdrawal(args string, replyChan chan string) error {
//	req := new(def.ReqWithdrawal)
//	err := json.Unmarshal([]byte(args), req)
//	if err != nil {
//		fmt.Printf("HandleWithdrawal Json Unmarshal Error:%s", err.Error())
//		return err
//	}
//
//	rsp := new(def.RspWithdrawal)
//	rsp.Result.UserOrderID = req.Params.UserOrderID
//	rsp.Result.Timestamp = "0"
//	rsp.Status.Code = 0
//	rsp.Status.Msg = ""
//
//	reply, _ := json.Marshal(rsp)
//
//	replyChan <- string(reply)
//
//	return nil
//}
