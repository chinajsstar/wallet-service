package service

import (
	"blockchain_server/conf"
	"blockchain_server/types"
	"github.com/alecthomas/log4go"
)

func NewQueryTxCmd(msgId, coinname, hash string) *types.CmdqueryTx {
	return &types.CmdqueryTx{
		NetCmd:types.NetCmd{MsgId:msgId, Coinname:coinname, Method:"get_transaction", Result:nil, Error:nil},
		Hash: hash}
}

func NewAccountCmd(msgId, coinname string, amount uint32) *types.CmdNewAccounts {
	return &types.CmdNewAccounts{
		NetCmd:types.NetCmd{MsgId: msgId, Coinname:coinname, Method:"new_account", Result:nil, Error:nil},
		Amount:amount}
}

func NewQueryBalanceCmd(msgId, coinname, address string, tokenName *string) *types.CmdqueryBalance{
	return &types.CmdqueryBalance{
		NetCmd:types.NetCmd{MsgId: msgId, Coinname:coinname, Method:"get_balance", Result:nil, Error:nil},
		Address:address, Token:tokenName}
}

func NewSendTxCmd(msgId, coinname, chiperKey, to string, tkname *string, amount uint64) (*types.CmdSendTx) {
	config := config.GetConfiger().Clientconfig[coinname]
	if config == nil {
		log4go.Error("Coin[%s] not supported", coinname)
		return nil
	}

	// Token针指向的对象, 就是config的Token指针的同一个对象,
	// 不应该被修改, 只能够被使用
	var tk *types.Token
	if nil!=tkname {
		tk = config.Tokens[*tkname]
	}

	return &types.CmdSendTx{ NetCmd:types.NetCmd{MsgId: msgId, Coinname:coinname, Method:"send_transaction", Result:nil, Error:nil},
		Chiperkey:chiperKey,
		Tx:&types.Transfer{
			To: to, Value: amount, Confirmationsnumber: config.TxConfirmNumber, InBlock: 0, ConfirmatedHeight: 0, Token:tk},
	}
}

func NewRechargeAddressCmd(msgId, coin string, address []string) (*types.CmdRechargeAddress) {
	//config := config.GetConfiger().Clientconfig[coin]
	//if config == nil {
	//	log4go.Error("Coin[%s] not supported",coin)
	//	return nil
	//}
	//
	//token := config.Tokens[tkname]
	return &types.CmdRechargeAddress{
		NetCmd:types.NetCmd{MsgId: msgId, Coinname: coin, Method:"watch_addresses", Result:nil, Error:nil},
		Addresses:address }
}
