package blockchain_server

import (
	"blockchain_server/types"
	"context"
	"time"
)

type Signer interface {
	SignTx(acount *types.Account, tx interface{})(signed_tx interface{}, err error)
}


type ChainClient interface {
	Name() string
	NewAccount()(*types.Account, error)
	// from is a crypted private key
	SendTx(ctx context.Context, privkey string, transfer *types.Transfer) error
	Tx(ctx context.Context, tx_hash string)(*types.Transfer, error)
	//TxRecipt(ctx context.Context, tx_hash string)(*types.Transfer, error)
	Blocknumber(ctx context.Context) (uint64, error)
	InsertCareAddress(address []string)

	Start(rcTxchannel types.RechargeTxChannel) error
	Stop(ctx context.Context,  duration time.Duration)
}