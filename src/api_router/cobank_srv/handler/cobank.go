package handler

import (
	"api_router/base/data"
	"api_router/base/service"
	"bastionpay_tools/common"
	"bastionpay_tools/handler"
	bservice "blockchain_server/service"
	"blockchain_server/types"
	"business_center/business"
	"business_center/def"
	"business_center/mysqlpool"
	"encoding/json"
	l4g "github.com/alecthomas/log4go"
	"strings"
	"time"
)

type Cobank struct {
	business *business.Business

	node *service.ServiceNode

	// http 服务的数据目录
	dataDir string
}

func NewCobank(dataDir string) *Cobank {
	x := &Cobank{}
	x.business = &business.Business{}
	x.dataDir = dataDir
	return x
}

func (x *Cobank) Start(node *service.ServiceNode) error {
	x.node = node
	return x.business.InitAndStart(x.callBack)
}

func (x *Cobank) Stop() {
	x.business.Stop()
}

func (x *Cobank) callBack(userID string, callbackMsg string) {
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

func (x *Cobank) GetApiGroup() map[string]service.NodeApi {
	nam := make(map[string]service.NodeApi)

	func() {
		example := "{\"user_key\":\"\",\"asset_id\":2,\"count\":0}"
		service.RegisterApi(&nam,
			"new_address", data.APILevel_client, x.handler,
			"获取新地址", example, "", "")
	}()

	func() {
		example := "{\"user_key\":\"\"}"
		service.RegisterApi(&nam,
			"query_user_address", data.APILevel_client, x.handler,
			"查询用户地址", example, "", "")
	}()

	func() {
		input := def.ReqWithdrawal{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"withdrawal", data.APILevel_client, x.handler,
			"提币", string(b), input, "")
	}()

	func() {
		service.RegisterApi(&nam,
			"support_assets", data.APILevel_client, x.handler,
			"查询支持的币种", "", "", "")
	}()

	func() {
		service.RegisterApi(&nam,
			"asset_attributie", data.APILevel_client, x.handler,
			"查询币种属性", "", "", "")
	}()

	////////////////////////////////////////////////////////////////
	// 以下方法liuheng添加测试
	func() {
		input := Recharge{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"recharge", data.APILevel_client, x.recharge,
			"模拟充值", string(b), input, "")
	}()

	func() {
		input := ImportAddress{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"importaddress", data.APILevel_admin, x.importAddress,
			"导入地址", string(b), input, "")
	}()

	func() {
		input := SendSignedTx{}
		b, _ := json.Marshal(input)
		service.RegisterApi(&nam,
			"sendsignedtx", data.APILevel_admin, x.sendSignedTx,
			"发送签名交易", string(b), input, "")
	}()

	return nam
}

func (x *Cobank) handler(req *data.SrvRequestData, res *data.SrvResponseData) {
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
type Recharge struct {
	Coin  string `json:"coin"`
	To    string `json:"to"`
	Value uint64 `json:"value"`
}

func (x *Cobank) recharge(req *data.SrvRequestData, res *data.SrvResponseData) {
	res.Data.Err = data.NoErr

	l4g.Debug("argv: %s", req.Data.Argv)

	rc := Recharge{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &rc)
	if err != nil {
		l4g.Error("json err: ", err)
	}

	go func(req *Recharge) {
		clientManager := x.business.GetWallet()
		tmp_account := &types.Account{
			"0x04e2b6c9bfeacd4880d99790a03a3db4ad8d87c82bb7d72711b277a9a03e49743077f3ae6d0d40e6bc04eceba67c2b3ec670b22b30d57f9d6c42779a05fba097536c412af73be02d1642aecea9fa7082db301e41d1c3c2686a6a21ca431e7e8605f761d8e12d61ca77605b31d707abc3f17bc4a28f4939f352f283a48ed77fc274b039590cc2c43ef739bd3ea13e491316",
			"0x54b2e44d40d3df64e38487dd4e145b3e6ae25927"}

		var token *string
		token = nil
		privatekey := tmp_account.PrivateKey

		txCmd := bservice.NewSendTxCmd("message id", req.Coin, privatekey, req.To, token, req.Value)
		clientManager.SendTx(txCmd)
	}(&rc)

	l4g.Debug("value: %s", res.Data.Value)
}

// 导入地址
type ImportAddress struct {
	UniName string `json:"uniname" comment:"DB文件名"`
}

func (x *Cobank) importAddress(req *data.SrvRequestData, res *data.SrvResponseData) {
	res.Data.Err = data.NoErr

	l4g.Debug("argv: %s", req.Data.Argv)

	ia := ImportAddress{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &ia)
	if err != nil {
		l4g.Error("json err: ", err)
	}

	// read db, import to free addrss
	uniDbName := ia.UniName + "@" + common.GetOnlineDbNameSuffix()
	aCcs, err := handler.LoadAddress(x.dataDir, uniDbName)
	if err != nil {
		res.Data.Err = 1
		l4g.Error("load online address failed: %s", err.Error())
		return
	}

	uniNameTags := strings.Split(ia.UniName, "@")
	if len(uniNameTags) != 3 {
		res.Data.Err = 1
		l4g.Error("error filename format")
		return
	}

	db := mysqlpool.Get()

	coinType := uniNameTags[0]
	dataTime := uniNameTags[1]

	t, err := time.Parse(common.TimeFormat, dataTime)
	//uuid := uniNameTags[2]
	asset_id := -1
	row := db.QueryRow("select id from asset_property where name = ?", coinType)
	err = row.Scan(&asset_id)
	if err != nil {
		res.Data.Err = 1
		l4g.Error("没有找到币种")
		return
	}

	Tx, err := mysqlpool.Get().Begin()
	if err != nil {
		res.Data.Err = 1
		l4g.Error("数据库begin：%s", err.Error())
		return
	}

	for _, acc := range aCcs {
		_, err = Tx.Exec("insert free_address (asset_id, address, private_key, create_time) values (?, ?, ?, ?);",
			asset_id, acc.Address, acc.PrivateKey,
			t.Format(def.TimeFormat))
		if err != nil {
			l4g.Error("写入失败：%s", err.Error())
			break
		}
	}

	if err != nil {
		Tx.Rollback()
		res.Data.Err = 1
		l4g.Error("写入失败：%s", err.Error())
		return
	}

	Tx.Commit()
	l4g.Debug("写入完成")

	l4g.Debug("value: %s", res.Data.Value)
}

// 导入地址
type SendSignedTx struct {
	TxSignedName string `json:"txsignedname" comment:"签名交易文件名"`
}

func (x *Cobank) sendSignedTx(req *data.SrvRequestData, res *data.SrvResponseData) {
	res.Data.Err = data.NoErr

	l4g.Debug("argv: %s", req.Data.Argv)

	ia := SendSignedTx{}
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &ia)
	if err != nil {
		res.Data.Err = 1
		l4g.Error("json err: ", err)
		return
	}

	txSignedFilePath := x.dataDir + "/" + ia.TxSignedName
	err = handler.SendSignedTx(x.business.GetWallet(), txSignedFilePath)
	if err != nil {
		res.Data.Err = 1
		l4g.Error("send signed tx failed: ", err)
		return
	}

	l4g.Debug("value: %s", res.Data.Value)
}
