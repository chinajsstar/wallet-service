package business

import (
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"business_center/address"
	"business_center/def"
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

func NewBusinessSvr() *Business {
	return new(Business)
}

type Business struct {
	wallet  *service.ClientManager
	ctx     context.Context
	cancel  context.CancelFunc
	address *address.Address
}

func (b *Business) InitAndStart() error {
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.wallet = service.NewClientManager()
	b.address = &address.Address{}

	//实例化以太坊客户端
	client, err := eth.NewClient()
	if err != nil {
		fmt.Printf("InitAndStart NewClient %s Error : %s\n", types.Chain_eth, err.Error())
		return err
	}
	b.wallet.AddClient(client)

	b.address.Run(b.ctx, b.wallet)
	b.wallet.Start()

	return nil
}

func (b *Business) Stop() {
	b.cancel()
	b.address.Stop()
}

func (b *Business) HandleMsg(args string, reply *string) error {
	var head def.ReqHead
	err := json.Unmarshal([]byte(args), &head)
	if err != nil {
		fmt.Printf("HandleMsg Unmarshal Error: %s/n", err.Error())
		return err
	}

	switch head.Method {
	case "new_address":
		{
			return b.address.AllocationAddress(args, reply)
		}
	case "withdrawal":
		{
			return b.address.Withdrawal(args, reply)
		}
	}
	return errors.New("invalid command")
}
