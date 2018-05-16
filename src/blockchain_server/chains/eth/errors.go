package eth

import "github.com/ethereum/go-ethereum/core/types"

type TxNoNeedApprove struct {
	messsage	string
	tx_information string
}

func (e TxNoNeedApprove) Error () string {
	return e.messsage
}

func (e TxNoNeedApprove) TxInfo() string {
	return e.tx_information
}

func NewTxNoNeedApprove(tx *types.Transaction) TxNoNeedApprove {
	return TxNoNeedApprove{
		messsage:tx.Hash().String(),
		tx_information:tx.String() }
}
