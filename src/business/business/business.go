package business

import (
	"bastionpay_base/config"
	adata "bastionpay_base/data"
	"blockchain_server/chains/btc"
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"business/chain"
	"business/data"
	. "business/def"
	"business/monitor"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	wallet  *service.ClientManager = nil
	funcMap map[string]reflect.Value
)

func init() {
	wallet = service.NewClientManager()
	chain.SetWallet(wallet)
}

func addFuncMap(cmdName string, funcV interface{}) {
	funcMap[cmdName] = reflect.ValueOf(funcV)
}

func initFunc() {
	funcMap = make(map[string]reflect.Value)

	//普通用户接口
	addFuncMap("support_assets", data.SupportAssets)
	addFuncMap("asset_attribute", data.AssetAttribute)
	addFuncMap("new_address", chain.NewAddress)
	addFuncMap("withdrawal", chain.Withdrawal)
	addFuncMap("query_address", data.QueryAddress)
	addFuncMap("get_balance", data.GetBalance)
	addFuncMap("transaction_bill", data.HistoryTransactionBill)
	addFuncMap("transaction_bill_daily", data.HistoryTransactionBillDaily)
	addFuncMap("transaction_message", data.HistoryTransactionMessage)
	addFuncMap("block_height", chain.BlockHeight)

	//管理员接口
	addFuncMap("sp_get_asset_attribute", data.SpGetAssetAttribute)
	addFuncMap("sp_set_asset_attribute", data.SpSetAssetAttribute)
	addFuncMap("sp_query_address", data.SpQueryAddress)
	addFuncMap("sp_get_chain_balance", chain.SpGetChainBalance)
	addFuncMap("sp_post_transaction", chain.SpPostTransaction)
	addFuncMap("sp_get_pay_address", data.SpGetPayAddress)
	addFuncMap("sp_set_pay_address", data.SpSetPayAddress)
	addFuncMap("sp_transaction_bill", data.SpHistoryTransactionBill)
	addFuncMap("sp_transaction_bill_daily", data.SpHistoryTransactionBillDaily)
	addFuncMap("sp_get_balance", data.SpGetBalance)
}

func NewServer() *Business {
	return new(Business)
}

type Business struct {
	ctx     context.Context
	cancel  context.CancelFunc
	monitor *monitor.Monitor
}

// 模拟充值 add by liuheng
func (b *Business) GetWallet() *service.ClientManager {
	return wallet
}

func (b *Business) InitAndStart(callback PushMsgCallback) error {
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.monitor = &monitor.Monitor{}

	var chains []string
	err := config.LoadJsonNode(config.GetBastionPayConfigDir()+"/cobank.json", "chains", &chains)
	if err != nil {
		return err
	}

	initFunc()

	for _, value := range chains {
		switch strings.ToUpper(value) {
		case "BTC":
			//实例化比特币客户端
			btcClient, err := btc.ClientInstance()
			if err == nil {
				wallet.AddClient(btcClient)
			} else {
				fmt.Printf("InitAndStart btcClientInstance %s Error : %s\n", types.Chain_bitcoin, err.Error())
			}
		case "ETH":
			//实例化以太坊客户端
			ethClient, err := eth.ClientInstance()
			if err == nil {
				wallet.AddClient(ethClient)
			} else {
				fmt.Printf("InitAndStart ethClientInstance %s Error : %s\n", types.Chain_eth, err.Error())
			}
		}
	}

	b.monitor.Run(b.ctx, wallet, callback)
	wallet.Start()

	return nil
}

func (b *Business) Stop() {
	b.cancel()
	b.monitor.Stop()
}

func (b *Business) HandleMsg(req *adata.SrvRequest, res *adata.SrvResponse) error {
	if v, ok := funcMap[req.Method.Function]; ok {
		params := make([]reflect.Value, 0)
		params = append(params, reflect.ValueOf(req), reflect.ValueOf(res))
		if e, ok := v.Call(params)[0].Interface().(error); ok {
			return e
		}
		return nil
	}
	return errors.New("invalid command")
}
