package notice

import (
	"blockchain_server/types"
	"business_center/def"
	"context"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"sync"
)

type Notice struct {
	rechargeChannel types.RechargeTxChannel
	cmdTxChannel    types.CmdTxChannel
	waitGroup       sync.WaitGroup
	ctx             context.Context
}

func NewNotice(ctx context.Context, rechargeChannel types.RechargeTxChannel, cmdTxChannel types.CmdTxChannel) *Notice {
	instance := new(Notice)
	instance.rechargeChannel = rechargeChannel
	instance.cmdTxChannel = cmdTxChannel
	instance.ctx = ctx
	return instance
}

func (ntc *Notice) Start() {
	ntc.recvRechargeTxChannel()
	ntc.recvCmdTxChannel()
}

func (ntc *Notice) Stop() {
	ntc.waitGroup.Wait()
}

func (ntc *Notice) recvRechargeTxChannel() {
	ntc.waitGroup.Add(1)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		c, err := redis.Dial("tcp", "127.0.0.1:6379")
		if err != nil {
			fmt.Printf("recvRechargeTxChannel Redis Dial Error: %s", err.Error())
			return
		}
		defer c.Close()

		for {
			select {
			case rct := <-channel:
				{
					userAddr, err := findUserAddress(c, rct.Coin_name, rct.Tx.To)
					if err != nil {
						fmt.Println(err.Error())
						continue
					}

					lastTrans, err := findTransaction(c, rct.Tx.Tx_hash)
					if err != nil {
						fmt.Println(err.Error())
						continue
					}

					var trans def.TransactionDetail
					trans.UserID = userAddr.UserID
					trans.AssetID = userAddr.AssetID

					trans.TxHash = rct.Tx.Tx_hash
					trans.From = rct.Tx.From
					trans.To = rct.Tx.To
					trans.Value = rct.Tx.Value
					trans.Gase = rct.Tx.Gase
					trans.Gaseprice = rct.Tx.Gaseprice
					trans.Total = rct.Tx.Total
					trans.Fee = 0
					trans.State = int(rct.Tx.State)
					trans.OnBlock = rct.Tx.OnBlock
					trans.PresentBlock = rct.Tx.PresentBlock
					trans.ConfirmationsNumber = rct.Tx.Confirmationsnumber
					trans.CreateTime = rct.Tx.Time
					trans.UpdateTime = rct.Tx.Time

					if lastTrans != nil {
						if lastTrans.PresentBlock-lastTrans.OnBlock >= 10 {
							continue
						}
						trans.CreateTime = lastTrans.CreateTime
					}

					reply, err := json.Marshal(trans)
					if err != nil {
						fmt.Println(err.Error())
						continue
					}
					c.Do("hset", "transaction", trans.TxHash, reply)

					if trans.PresentBlock-trans.OnBlock >= 10 {
						reply, err = redis.Bytes(c.Do("hget", "user_account", trans.UserID+"_"+rct.Coin_name))
						if err != nil {
							continue
						}
						var userAccount def.UserAccount
						err = json.Unmarshal(reply, &userAccount)
						if err != nil {
							continue
						}
						userAccount.AvailableAmount += float64(trans.Value) / 1000000000000000000
						reply, err = json.Marshal(userAccount)
						if err != nil {
							continue
						}
						c.Do("hset", "user_account", trans.UserID+"_"+rct.Coin_name, reply)
					}
				}
			case <-ctx.Done():
				{
					fmt.Println("RechangeTx context done, because : ", ctx.Err())
					ntc.waitGroup.Done()
					return
				}
			}
		}
	}(ntc.ctx, ntc.rechargeChannel)
}

func (ntc *Notice) recvCmdTxChannel() {
	ntc.waitGroup.Add(1)
	go func(ctx context.Context, channel types.CmdTxChannel) {
		for {
			select {
			case cmdTx := <-channel:
				{
					fmt.Printf("Transaction state changed, transaction information:%s\n",
						cmdTx.Tx.String())

					if cmdTx.Tx.State == types.Tx_state_confirmed {
						fmt.Println("Transaction is confirmed! success!!!")
					}

					if cmdTx.Tx.State == types.Tx_state_unconfirmed {
						fmt.Println("Transaction is unconfirmed! failed!!!!")
					}
				}
			case <-ctx.Done():
				fmt.Println("TxState context done, because : ", ctx.Err())
				ntc.waitGroup.Done()
				return
			}
		}
	}(ntc.ctx, ntc.cmdTxChannel)
}

func findUserAddress(c redis.Conn, assetID string, address string) (*def.UserAddress, error) {
	reply, err := redis.Bytes(c.Do("hget", "user_address_"+assetID, address))
	if err != nil {
		return nil, err
	}

	userAddress := new(def.UserAddress)
	err = json.Unmarshal(reply, userAddress)
	if err != nil {
		return nil, err
	}

	return userAddress, nil
}

func findTransaction(c redis.Conn, txHash string) (*def.TransactionDetail, error) {
	r, err := c.Do("hget", "transaction", txHash)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, nil
	}

	reply, err := redis.Bytes(r, err)
	if err != nil {
		return nil, err
	}

	trans := new(def.TransactionDetail)
	err = json.Unmarshal(reply, trans)
	if err != nil {
		return nil, err
	}

	return trans, nil
}
