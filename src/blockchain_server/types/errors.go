package types

import (
	"github.com/ethereum/go-ethereum/core/types"
	"fmt"
)

type NotFound struct {
	message string
}

func (self *NotFound) Error() string {
	return self.message
}

func NewNotFound(message string) *NotFound {
	return &NotFound{message: message}
}

func NewTxNotFoundErr(tx_hash string) *NotFound {
	//return &NotFound{tx_info: fmt.Sprintf("Transaction not found, detail:%s", tx.String())}
	return &NotFound{message: fmt.Sprintf("Transaction(%s) not found!", tx_hash)}
}

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

