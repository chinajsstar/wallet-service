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
	Tx(tx_hash string)(*types.Transfer, error)
	//TxRecipt(ctx context.Context, tx_hash string)(*types.Transfer, error)
	BlockHeight() (uint64)
	InsertRechageAddress(address []string)

	SubscribeRechageTx(txRechChannel types.RechargeTxChannel)

	Start() error
	Stop()
}