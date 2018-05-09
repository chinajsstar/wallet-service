package business

import (
	"api_router/base/config"
	"api_router/base/data"
	"blockchain_server/chains/btc"
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

// 模拟充值 add by liuheng
func (b *Business) GetWallet() *service.ClientManager {
	return b.wallet
}

func (b *Business) InitAndStart(callback PushMsgCallback) error {
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.wallet = service.NewClientManager()
	b.address = &address.Address{}

	var chains []string
	err := config.LoadJsonNode(config.GetBastionPayConfigDir()+"/cobank.json", "chains", &chains)
	if err != nil {
		return err
	}

	for _, value := range chains {
		switch value {
		case "btc":
			//实例化比特币客户端
			btcClient, err := btc.ClientInstance()
			if err == nil {
				b.wallet.AddClient(btcClient)
			} else {
				fmt.Printf("InitAndStart btcClientInstance %s Error : %s\n", types.Chain_bitcoin, err.Error())
			}

		case "eth":
			//实例化以太坊客户端
			ethClient, err := eth.ClientInstance()
			if err == nil {
				b.wallet.AddClient(ethClient)
			} else {
				fmt.Printf("InitAndStart ethClientInstance %s Error : %s\n", types.Chain_eth, err.Error())
			}

		}
	}

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
		return b.address.NewAddress(req, res)

	case "withdrawal":
		return b.address.Withdrawal(req, res)

	case "support_assets":
		return b.address.SupportAssets(req, res)

	case "asset_attribute":
		return b.address.AssetAttribute(req, res)

	case "get_balance":
		return b.address.GetBalance(req, res)

	case "history_transaction_order":
		return b.address.HistoryTransactionOrder(req, res)

	case "history_transaction_message":
		return b.address.HistoryTransactionMessage(req, res)

	case "query_user_address":
		return b.address.QueryUserAddress(req, res)

	case "set_pay_address":
		return b.address.SetPayAddress(req, res)
	}
	return errors.New("invalid command")
}
