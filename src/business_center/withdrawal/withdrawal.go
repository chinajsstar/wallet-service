package withdrawal

import (
	"blockchain_server/service"
	"blockchain_server/types"
	"business_center/def"
	"business_center/redispool"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

type Withdrawal struct {
	wallet *service.ClientManager
}

func (r *Withdrawal) Init(wallet *service.ClientManager) {
	r.wallet = wallet
}

func (r *Withdrawal) HandleWithdrawal(req string, ack *string) error {
	reqInfo := new(def.ReqWithdrawal)
	err := json.Unmarshal([]byte(req), reqInfo)
	if err != nil {
		fmt.Printf("Withdrawal Json Unmarshal Error:%s", err.Error())
		return err
	}

	c := redispool.Get()
	defer c.Close()

	jsonInfo, err := redis.Bytes(c.Do("hget", "user_account", reqInfo.UserID+"_"+reqInfo.Params.Symbol))
	if err != nil {
		return err
	}

	var userAccount def.UserAccount
	err = json.Unmarshal(jsonInfo, &userAccount)
	if err != nil {
		return err
	}

	if userAccount.AvailableAmount >= reqInfo.Params.Amount {
		userAccount.AvailableAmount = userAccount.AvailableAmount - reqInfo.Params.Amount
		userAccount.FrozenAmount = userAccount.FrozenAmount + reqInfo.Params.Amount
		userAccount.UpdateTime = uint64(time.Now().Unix())

		jsonInfo, err := json.Marshal(userAccount)
		if err != nil {
			return err
		}

		c.Do("hset", "user_account", reqInfo.UserID+"_"+reqInfo.Params.Symbol, jsonInfo)

		txCmd := types.NewTxCmd("message id", types.Chain_eth, "0x040cae69092e07c8c7f788ed072ec630e50f899588727bfc7855ff5c2a8c3dad2f30ad6538996baf3e2b35fb7f98d6218a20b60e1c57d1edfe5364e948c840b477143087ff6481053d5e504735961d756932998f587fc269e601bc32ca3b7cb8374355bb4d38491a260a0947e72e2c8228281380dad7d8a50f738ca89cc82410ebb7083e6ec79e1001443232e6bcd96450", "0x8128b33eb9d5b5fc975f42eb944a24292db09ec5", 10000000000000000000) //0x00f55b34Ae3Ec318fDE10846a74B4e40f6cc5614
		r.wallet.SendTx(txCmd)
	}

	//查询商户帐户，判断能不能交易
	//如果能交易，冻结要交易的资金
	//如果不能交易，返回错误信息

	return nil
}
