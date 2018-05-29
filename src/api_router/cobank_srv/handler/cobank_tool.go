package handler

import (
	"api_router/base/data"
	"bastionpay_api/api/v1"
	"bastionpay_api/apibackend"
	"bastionpay_tools/common"
	"bastionpay_tools/handler"
	bservice "blockchain_server/service"
	"blockchain_server/types"
	"business/def"
	"business/mysqlpool"
	"encoding/json"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////
const (
	btc_cmd = "/opt/btc_app/bitcoin-0.16.0/bin/bitcoin-cli"
	btc_arg = "-conf=/data/btc_data/bitcoin.conf"
	//btc_cmd = "bitcoin-cli"
	//btc_arg = ""
)

func (x *Cobank) recharge(req *data.SrvRequest, res *data.SrvResponse) {
	res.Err = apibackend.NoErr

	l4g.Debug("argv: %s", req.Argv)

	rc := v1.ReqRecharge{}
	err := json.Unmarshal([]byte(req.Argv.Message), &rc)
	if err != nil {
		l4g.Error("json err: ", err)
	}

	if rc.Coin == "eth" {
		go func(req *v1.ReqRecharge) {
			clientManager := x.business.GetWallet()
			//tmp_account_ztoken := &types.Account{
			//	"0x04e2b6c9bfeacd4880d99790a03a3db4ad8d87c82bb7d72711b277a9a03e49743077f3ae6d0d40e6bc04eceba67c2b3ec670b22b30d57f9d6c42779a05fba097536c412af73be02d1642aecea9fa7082db301e41d1c3c2686a6a21ca431e7e8605f761d8e12d61ca77605b31d707abc3f17bc4a28f4939f352f283a48ed77fc274b039590cc2c43ef739bd3ea13e491316",
			//	"0x54b2e44d40d3df64e38487dd4e145b3e6ae25927"}

			tmp_account_ztoken := &types.Account{
				"0x04bac76a1adee95417514a449efe0f1e6ccfb856b1f077a2b61dd0bdf8fd71defb4b9761d27bd0bf3f3f80094844edb19b55c54d159254d4b4230a88a2f19202093df5687bc03eff86875baf1fa45710073b3496fb8a4edcee9f54af9213ec41b5d0380454435ec4361892b1ae8dca64f9343b2fd46a492b2bd9623b1f23757b5398de5b33606835e68824c9bd0d8f516c",
				"0x5c0ebcbaa341147c1b98d097bed99356f8b8340f"}

			var err error
			var txCmd *types.CmdSendTx
			if req.Token == "" {
				txCmd, err = bservice.NewSendTxCmd("", req.Coin, tmp_account_ztoken.PrivateKey, req.To, req.Token, tmp_account_ztoken.PrivateKey, req.Value)
			} else {
				txCmd, err = bservice.NewSendTxCmd("", req.Coin, tmp_account_ztoken.PrivateKey, req.To, req.Token, tmp_account_ztoken.PrivateKey, req.Value)
			}

			if err != nil {
				res.Err = 1
				res.ErrMsg = "NewSendTxCmd-" + err.Error()
			} else {
				clientManager.SendTx(txCmd)
			}
		}(&rc)
	} else if rc.Coin == "btc" {
		amount := fmt.Sprintf("%.8f", rc.Value)
		l4g.Debug(btc_cmd, btc_arg, "sendtoaddress", rc.To, amount)

		var c *exec.Cmd
		if btc_arg != "" {
			c = exec.Command(btc_cmd, btc_arg, "sendtoaddress", rc.To, amount)
		} else {
			c = exec.Command(btc_cmd, "sendtoaddress", rc.To, amount)
		}

		if c != nil {
			if err := c.Run(); err != nil {
				fmt.Println(err)
				res.Err = 1
				res.ErrMsg = err.Error()
			}
		} else {
			res.Err = 1
			res.ErrMsg = "no bitcoin-cli command"
		}
	}

	l4g.Debug("value: %s", res.Value)
}

func (x *Cobank) generate(req *data.SrvRequest, res *data.SrvResponse) {
	res.Err = apibackend.NoErr

	l4g.Debug("argv: %s", req.Argv)

	rc := v1.ReqGenerate{}
	err := json.Unmarshal([]byte(req.Argv.Message), &rc)
	if err != nil {
		l4g.Error("json err: ", err)
	}

	if rc.Coin == "eth" {
		res.Err = 1
		res.ErrMsg = "not support eth"
	} else if rc.Coin == "btc" {
		var c *exec.Cmd
		if btc_arg != "" {
			c = exec.Command(btc_cmd, btc_arg, "generate", strconv.Itoa(rc.Count))
		} else {
			c = exec.Command(btc_cmd, "generate", strconv.Itoa(rc.Count))
		}

		if c != nil {
			if err := c.Run(); err != nil {
				fmt.Println(err)
				res.Err = 1
				res.ErrMsg = err.Error()
			}
		} else {
			res.Err = 1
			res.ErrMsg = "no bitcoin-cli command"
		}
	}

	l4g.Debug("value: %s", res.Value)
}

// 导入地址
type ImportAddress struct {
	UniName string `json:"uniname" comment:"DB文件名"`
}

func (x *Cobank) importAddress(req *data.SrvRequest, res *data.SrvResponse) {
	res.Err = apibackend.NoErr

	l4g.Debug("argv: %s", req.Argv)

	ia := ImportAddress{}
	err := json.Unmarshal([]byte(req.Argv.Message), &ia)
	if err != nil {
		l4g.Error("json err: ", err)
	}

	// read db, import to free addrss
	uniDbName := ia.UniName + "@" + common.GetOnlineExtension()
	aCcs, err := handler.LoadAddress(x.dataDir, uniDbName)
	if err != nil {
		res.Err = 1
		l4g.Error("load online address failed: %s", err.Error())
		return
	}

	uniNameTags := strings.Split(ia.UniName, "@")
	if len(uniNameTags) != 3 {
		res.Err = 1
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
		res.Err = 1
		l4g.Error("没有找到币种")
		return
	}

	Tx, err := mysqlpool.Get().Begin()
	if err != nil {
		res.Err = 1
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
		res.Err = 1
		l4g.Error("写入失败：%s", err.Error())
		return
	}

	Tx.Commit()
	l4g.Debug("写入完成")

	l4g.Debug("value: %s", res.Value)
}

// 导入地址
type SendSignedTx struct {
	TxSignedName string `json:"txsignedname" comment:"签名交易文件名"`
}

func (x *Cobank) sendSignedTx(req *data.SrvRequest, res *data.SrvResponse) {
	res.Err = apibackend.NoErr

	l4g.Debug("argv: %s", req.Argv)

	ia := SendSignedTx{}
	err := json.Unmarshal([]byte(req.Argv.Message), &ia)
	if err != nil {
		res.Err = 1
		l4g.Error("json err: ", err)
		return
	}

	txSignedFilePath := x.dataDir + "/" + ia.TxSignedName
	err = handler.SendSignedTx(x.business.GetWallet(), txSignedFilePath)
	if err != nil {
		res.Err = 1
		l4g.Error("send signed tx failed: ", err)
		return
	}

	l4g.Debug("value: %s", res.Value)
}
