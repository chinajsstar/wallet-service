package business

import (
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"business_center/address"
	"business_center/def"
	"business_center/notice"
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
	ntc     *notice.Notice
	address *address.Address
}

func (busi *Business) InitAndStart() error {
	busi.ctx, busi.cancel = context.WithCancel(context.Background())
	busi.wallet = service.NewClientManager()
	busi.address = &address.Address{}

	//实例化以太坊客户端
	client, err := eth.NewClient()
	if err != nil {
		fmt.Printf("InitAndStart NewClient %s Error : %s\n", types.Chain_eth, err.Error())
		return err
	}
	busi.wallet.AddClient(client)

	rechargeChannel := make(types.RechargeTxChannel)
	cmdTxChannel := make(types.CmdTxChannel)

	busi.wallet.SubscribeTxRecharge(rechargeChannel)
	busi.wallet.SubscribeTxCmdState(cmdTxChannel)

	busi.ntc = notice.NewNotice(busi.ctx, rechargeChannel, cmdTxChannel)
	busi.address.Init(busi.wallet)
	busi.ntc.Start()
	busi.wallet.Start()

	return nil
}

func (busi *Business) Stop() {
	busi.cancel()
	busi.ntc.Stop()
}

func (busi *Business) HandleMsg(args string, reply *string) error {
	var head def.ReqHead
	err := json.Unmarshal([]byte(args), &head)
	if err != nil {
		fmt.Printf("HandleMsg Unmarshal Error: %s/n", err.Error())
		return err
	}

	switch head.Method {
	case "new_address":
		{
			return busi.address.AllocationAddress(args, reply)
		}
	case "withdrawal":
		{
		}
	}
	return errors.New("invalid command")
}
