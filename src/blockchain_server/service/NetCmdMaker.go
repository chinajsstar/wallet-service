package service

import (
	"blockchain_server/conf"
	"blockchain_server/types"
	"fmt"
)

func NewQueryTxCmd(msgId, coinname, hash string) *types.CmdqueryTx {
	return &types.CmdqueryTx{
		NetCmd: types.NetCmd{MsgId: msgId, Coinname: coinname, Method: "get_transaction", Result: nil, Error: nil},
		Hash:   hash}
}

func NewAccountCmd(msgId, coinname string, amount uint32) *types.CmdNewAccounts {
	return &types.CmdNewAccounts{
		NetCmd: types.NetCmd{MsgId: msgId, Coinname: coinname, Method: "new_account", Result: nil, Error: nil},
		Amount: amount}
}

func NewQueryBalanceCmd(msgId, coinname, address string, tokenName *string) *types.CmdqueryBalance {
	return &types.CmdqueryBalance{
		NetCmd:  types.NetCmd{MsgId: msgId, Coinname: coinname, Method: "get_balance", Result: nil, Error: nil},
		Address: address, Token: tokenName}
}

// 交易和充值中的单位都是10^8为一个单位, 即1^8数量为一个bitcoin或者eth
// 发送时, 内部会执行转换!
func NewSendTxCmd(msgId, coinName, fromKey, to, tkname, tokenFromkey string, value uint64) (*types.CmdSendTx, error) {
	config := config.MainConfiger().Clientconfig[coinName]
	if config == nil {
		return nil, fmt.Errorf("Coin[%s] not supported", coinName)
	}

	var tokenTx *types.TokenTx
	if tkname != "" {
		if tk := config.Tokens[tkname]; tk == nil {
			return nil, fmt.Errorf("Not Supported %s:%s", coinName, tkname)
		} else {
			tokenTx = &types.TokenTx{
				From:"", 	// 这里设置为"", 后面在Client中, 通过tokenFromKey来计算
				To: to,		// 接收代币的地址
				Value: value,
				Contract: tk}

			// 如果是token代币
			// fromkey 是 执行erc20合约的地址 from, Transfer.To设置为代币合约地址
			// tokenfromKey 是erc20合约代币的转出地址(TokenTx.From)
			// TokenTx.To 设置为接收代币的地址
			// Transfer.Value 设置为0, TokenTx.Value才是代币的数量
			// 所以在这里需要把value的值设置为0, to设置为tk.Address
			value = 0
			to = tk.Address
		}
	}

	return &types.CmdSendTx{
		NetCmd: types.NetCmd{
			MsgId:    msgId,
			Coinname: coinName,
			Method:   "send_transaction",
			Result:   nil,
			Error:    nil},

		FromKey: fromKey,

		Tx: &types.Transfer{
			From:                "", // from 需要在client的内部, 解析CmdSendTx.Fromkey来得到
			To:                  to, // 如果是代币, to指向的是合约的地址
			Value:               value,
			Confirmationsnumber: config.TxConfirmNumber,
			InBlock:             0,
			ConfirmatedHeight:   0,

			State:				types.Tx_state_ToBuild,
			TokenFromKey:        tokenFromkey,
			TokenTx:             tokenTx},
	}, nil
}

func NewRechargeAddressCmd(msgId, coin string, address []string) *types.CmdRechargeAddress {
	return &types.CmdRechargeAddress{
		NetCmd:    types.NetCmd{MsgId: msgId, Coinname: coin, Method: "watch_addresses", Result: nil, Error: nil},
		Addresses: address}
}
