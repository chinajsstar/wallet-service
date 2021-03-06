package client

import (
	"blockchain_server/types"
)

type Signer interface {
	SignTx(acount *types.Account, tx interface{})(signed_tx interface{}, err error)
}


type ChainClient interface {
	Name() string
	NewAccount(c uint32)([]*types.Account, error)
	// from is a crypted private key
	SendTx(privkey string, transfer *types.Transfer) error
	UpdateTx(tx *types.Transfer) error
	BlockHeight() (uint64)
	SubscribeRechargeTx(txRechChannel types.RechargeTxChannel)

	InsertWatchingAddress(address []string) (invalid []string)

	// split SendTx to 3 steps: BuildTx, SignTx, SendSignedTx
	// liuheng add
	// TODO: zl review
	BuildTx(fromkey string, tx *types.Transfer) (error)
	SignTx(fromkey string, tx *types.Transfer) ([]byte, error)
	SendSignedTx(txByte []byte, tx *types.Transfer) (error)

	GetBalance(address string, tokenSymbol string) (float64, error)
	Tx(tx_hash string)(*types.Transfer, error)

	SetNotifyChannel(ch chan interface{})

	Start() error
	Stop()
}