package blockchain_server

import (
	"blockchain_server/types"
)

type Signer interface {
	SignTx(acount *types.Account, tx interface{})(signed_tx interface{}, err error)
}


type ChainClient interface {
	Name() string
	NewAccount()(*types.Account, error)
	// from is a crypted private key
	SendTx(privkey string, transfer *types.Transfer) error
	UpdateTx(tx *types.Transfer) error
	BlockHeight() (uint64)
	SubscribeRechageTx(txRechChannel types.RechargeTxChannel)

	InsertRechageAddress(address []string)
	GetBalance(address string, tokenname *string) (uint64, error)
	Tx(tx_hash string)(*types.Transfer, error)

	Start() error
	Stop()
}