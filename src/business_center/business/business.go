package business

import (
	"api_router/base/data"
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"business_center/address"
	. "business_center/def"
	"context"
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

func (b *Business) InitAndStart(callback PushMsgCallback) error {
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.wallet = service.NewClientManager()
	b.address = &address.Address{}

	//实例化以太坊客户端
	client, err := eth.ClientInstance()
	if err != nil {
		fmt.Printf("InitAndStart ClientInstance %s Error : %s\n", types.Chain_eth, err.Error())
		return err
	}
	b.wallet.AddClient(client)

	b.address.Run(b.ctx, b.wallet, callback)
	b.wallet.Start()

	return nil
}

func (b *Business) Stop() {
	b.cancel()
	b.address.Stop()
}

func (b *Business) HandleMsg(req *data.SrvRequestData, res *data.SrvResponseData) error {
	switch req.Data.Method.Function {
	case "new_address":
		{
			return b.address.NewAddress(req, res)
		}
	case "withdrawal":
		{
			return b.address.Withdrawal(req, res)
		}
	case "query_user_address":
		{
			return b.address.QueryUserAddress(req, res)
		}
	}
	return errors.New("invalid command")
}
