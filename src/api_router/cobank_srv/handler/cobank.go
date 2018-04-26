package handler

import (
	"api_router/base/data"
	"api_router/base/service"
	"business_center/business"
	l4g "github.com/alecthomas/log4go"
	"business_center/def"
	"encoding/json"
	bservice "blockchain_server/service"
	"blockchain_server/types"
)

type Cobank struct{
	business *business.Business

	node *service.ServiceNode
}

func NewCobank() (*Cobank) {
	x := &Cobank{}
	x.business = &business.Business{}
	return x
}

func (x *Cobank)Start(node *service.ServiceNode) error {
	x.node = node
	return x.business.InitAndStart(x.callBack)
}

func (x *Cobank)Stop() {
	x.business.Stop()
}

func (x *Cobank)callBack(userID string, callbackMsg string){
	pData := data.UserRequestData{}
	pData.Method.Version = "v1"
	pData.Method.Srv = "push"
	pData.Method.Function = "pushdata"

	pData.Argv.UserKey = userID
	pData.Argv.Message = callbackMsg

	res := data.UserResponseData{}
	x.node.InnerCallByEncrypt(&pData, &res)
	l4g.Info("push return: ", res)
}

func (x *Cobank)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	func() {
		input := def.ReqNewAddress{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"new_address", data.APILevel_client, x.handler,
			"获取新地址", string(b), input, "")
	}()

	func(){
		example := "{\"user_key\":\"\"}"
		service.RegisterApi(&nam,
			"query_user_address", data.APILevel_client, x.handler,
			"查询用户地址", example, "", "")
	}()

	func(){
		input := def.ReqWithdrawal{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"withdrawal", data.APILevel_client, x.handler,
			"提币", string(b), input, "")
	}()

	func(){
		input := Recharge{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"recharge", data.APILevel_client, x.recharge,
			"模拟充值", string(b), input, "")
	}()

	return nam
}

func (x *Cobank)handler(req *data.SrvRequestData, res *data.SrvResponseData){
	res.Data.Err = data.NoErr

	l4g.Debug("argv: %s", req.Data.Argv)

	err := x.business.HandleMsg(req, res)
	if err != nil {
		l4g.Error("err: ", err)
	}

	l4g.Debug("value: %s", res.Data.Value)
}

////////////////////////////////////////////////////////////
// 模拟充值
type Recharge struct{
	Coin string `json:"coin"`
	To string `json:"to"`
	Value uint64 `json:"value"`
}

func (x *Cobank)testSendTokenTx(req* Recharge) {

	clientManager := x.business.GetWallet()
	tmp_account := &types.Account{
		"0x04e2b6c9bfeacd4880d99790a03a3db4ad8d87c82bb7d72711b277a9a03e49743077f3ae6d0d40e6bc04eceba67c2b3ec670b22b30d57f9d6c42779a05fba097536c412af73be02d1642aecea9fa7082db301e41d1c3c2686a6a21ca431e7e8605f761d8e12d61ca77605b31d707abc3f17bc4a28f4939f352f283a48ed77fc274b039590cc2c43ef739bd3ea13e491316",
		"0x54b2e44d40d3df64e38487dd4e145b3e6ae25927"}

	var token *string
	token = nil
	privatekey := tmp_account.PrivateKey

	txCmd := bservice.NewSendTxCmd("message id", req.Coin, privatekey, req.To, token, req.Value)
	clientManager.SendTx(txCmd)

	/*********监控提币交易的channel*********/
	//txStateChannel := make(types.CmdTxChannel)

	/*
	// 创建并发送Transaction, 订阅只需要调用一次, 所有的Send的交易都会通过这个订阅channel传回来
	subscribe := clientManager.SubscribeTxCmdState(txStateChannel)

	txok_channel := make(chan bool)

	subCtx, cancel := context.WithCancel(ctx)

	go func(ctx context.Context, txstateChannel types.CmdTxChannel) {
		defer subscribe.Unsubscribe()
		defer close(txstateChannel)
		close := false
		for !close {
			select {
			case cmdTx := <-txStateChannel:
				{
					if cmdTx == nil {
						l4g.Trace("Transaction Command Channel is closed!")
						txok_channel <- false
					} else {
						l4g.Trace("Transaction state changed, transaction information:%s\n",
							cmdTx.Tx.String())

						if cmdTx.Tx.State == types.Tx_state_confirmed {
							l4g.Trace("Transaction is confirmed! success!!!")
							txok_channel <- true
						}

						if cmdTx.Tx.State == types.Tx_state_unconfirmed {
							l4g.Trace("Transaction is unconfirmed! failed!!!!")
							txok_channel <- false
						}
					}
				}
			case <-ctx.Done():
				{
					close = true
				}
			}
		}
	}(subCtx, txStateChannel)

	select {
	case <-txok_channel:
		{
			cancel()
		}
	}
	done <- true
	*/
}

func (x *Cobank)recharge(req *data.SrvRequestData, res *data.SrvResponseData){
	res.Data.Err = data.NoErr

	l4g.Debug("argv: %s", req.Data.Argv)

	rc := Recharge{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &rc)
	if err != nil {
		l4g.Error("json err: ", err)
	}

	go x.testSendTokenTx(&rc)

	l4g.Debug("value: %s", res.Data.Value)
}