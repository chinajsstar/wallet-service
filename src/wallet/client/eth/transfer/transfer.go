package transfer

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"time"
	"wallet/utils"
	"bytes"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	//"github.com/ethereum/go-ethereum/internal/ethapi"
	//"io"
	//"github.com/ethereum/go-ethereum/params"
	//"github.com/ethereum/go-ethereum/eth"
	//"github.com/ethereum/go-ethereum/node"
	//"github.com/ethereum/go-ethereum/rpc"
	//"math"
	//"github.com/ethereum/go-ethereum/crypto"
	//"github.com/ethereum/go-ethereum/accounts"
	//"bytes"
	//"github.com/ethereum/go-ethereum"
	//"github.com/ethereum/go-ethereum/accounts/keystore"
	//"os"
	//"strconv"
	//"encoding/json"
	//"time"
	//"github.com/ethereum/go-ethereum/accounts"
	//"github.com/ethereum/go-ethereum"
	//"bufio"
	//"github.com/ethereum/go-ethereum/accounts"
	//"os"
	//"encoding/json"
)

const (
	Unkown                 = iota
	Commited               // transaction was sended(call SendTransaction)
	Mined                  // transaction was mined and stored on a block. confirmed number is 1
	WaitConfirmationNumber // ethereum need 12 conformation number
	Confirmed              // .....
	Unconfirmed            // some error happened...need to re send
)

type ErrorTxDisappear struct {
	tx_hash  common.Hash
	blocknum *big.Int
}

func (self ErrorTxDisappear) Error() string {
	return fmt.Sprintf("can not found transaction:%s on block:%d",
		self.tx_hash, self.blocknum.Int64())
}

type ErrorTimeout struct {
	cmd          string
	max_duration time.Duration
}

func (e ErrorTimeout) Error() string {
	return fmt.Sprintf("%s, time out:%d", e.cmd, e.max_duration)
}

type ErrorTxUnconfirmed struct {
	transactionHash string
	gasUsed         uint64
	gas             uint64
}

func (e ErrorTxUnconfirmed) Error() string {
	return fmt.Sprintf("transaction:%s is unconfirmed, not enough gas, gasUsed:%d, gas:%d.",
		e.transactionHash, e.gasUsed, e.gas)
}

//type Callback_tx_state_changed func(transfer *Transfer, err error)

type Transfer struct {
	tx                  *types.Transaction
	blocknumber         *big.Int
	state               uint32
	confirmed_number    uint64
	confirmation_number uint16
	message             string
	identifer           string
}

func hasSignature(tx *types.Transaction) bool {
	if nil == tx {
		return false
	}

	if v, r, s := tx.RawSignatureValues(); v != nil && r != nil && s != nil && 0 != v.Int64() && 0 != r.Int64() && 0 != s.Int64() {
		return true
	}
	return false
}

func NewTransfer(tx *types.Transaction, identifer string) (transfer *Transfer) {
	return &Transfer{tx: tx, blocknumber: nil, state: Unkown, confirmed_number: 0, confirmation_number: 12, identifer: identifer}
}

func GetTransfer(client *ethclient.Client, tx_hash common.Hash) (*Transfer, error) {
	ts, hex_num_string, err := client.TransactionByHash(context.TODO(), tx_hash)
	if err != nil {
		return nil, err
	}

	bigint, err := utils.Hex_string_to_big_int(*hex_num_string)
	if err != nil {
		return nil, err
	}

	transfer := &Transfer{tx: ts, blocknumber: bigint, state: Mined}
	return transfer, nil
}

func (self *Transfer) HasSignatrue() bool {
	return hasSignature(self.tx)
}

/* parse address from key store data
var (
	Adr struct {
		Address string `json:"address"`
	}
)
Adr.Address = ""
err := json.NewDecoder(ks_data).Decode(&Adr)
addr := common.HexToAddress(Adr.Address)
*/

func (self *Transfer) SignTx(ks_data *bytes.Buffer, chain_id *big.Int, passpharse string) error {
	if self.HasSignatrue() {
		return nil
	}

	key, err := keystore.DecryptKey(ks_data.Bytes(), passpharse)
	if err != nil {
		return err
	}

	signer := types.NewEIP155Signer(chain_id)

	self.tx, err = types.SignTx(self.tx, signer, key.PrivateKey)
	return err
}

func (self *Transfer) Send(ctx context.Context, client *ethclient.Client) error {
	if !self.HasSignatrue() {
		return fmt.Errorf("transaction must be signed first")
	}
	if false {
		dur := time.Second
		ctx, cancel := context.WithTimeout(context.Background(), dur)
		defer cancel()

		if err := client.SendTransaction(ctx, self.tx); err != nil {
			return err
		} else {
			select {
			case <-time.After(dur):
				return &ErrorTimeout{cmd: "SendTransaction", max_duration: dur}
			case <-ctx.Done():
				return ctx.Err()
			default:
				self.state = Commited
				return nil
			}
		}
		return nil
	}

	if err := client.SendTransaction(ctx, self.tx); err != nil {
		return err
	}
	self.state = Commited
	return nil
}

func (self *Transfer) refresh_confrim_number(ctx context.Context, client *ethclient.Client) error {
	if self.tx == nil {
		self.state = Unkown
		return fmt.Errorf("self.transaction must not be nil")
	}

	block, err := client.BlockByNumber(ctx, self.blocknumber)
	if err != nil {
		return err
	}
	ts := block.Transaction(self.tx.Hash())
	if ts == nil {
		return ErrorTxDisappear{self.tx.Hash(), block.Number()}
	}

	lastblock, err := client.BlockByNumber(ctx, nil)
	if err != nil {
		return err
	}

	if 1 == lastblock.Number().Cmp(self.blocknumber) {
		self.confirmed_number = lastblock.Number().Sub(lastblock.Number(), self.blocknumber).Uint64()
		//fmt.Printf("current comfirmed number is %d\n", self.confirmed_number)
	}

	return nil
}

// is state changed
// is conformed number changed
func (self *Transfer) RefreshState(ctx context.Context, client *ethclient.Client) error {
	if !self.HasSignatrue() {
		self.state = Unkown
		return fmt.Errorf("transaction must be signed first")
	}

	switch self.state {
	case Commited:
		// TransactionByHash checks the pool of pending transactions in addition to the
		// blockchain. The isPending return value indicates whether the transaction has been
		// mined yet. Note that the transaction may not be part of the canonical chain even if
		// it's not pending.
		// TransactionByHash is changed by zl
		// the second return value form bool(isPending) to block number(which Transaction store at)
		tx, blocknum_hex_string, err := client.TransactionByHash(ctx, self.tx.Hash())

		if err != nil || nil == tx {
			return err
		}

		if blocknum_hex_string == nil {
			//fmt.Println("waiting for mined!!!")
			return nil
		}

		fmt.Printf("tx:%v, blocknum_hex_string:%s\n", *tx, *blocknum_hex_string)

		bigint, err := utils.Hex_string_to_big_int(*blocknum_hex_string)
		if err != nil {
			return err
		}

		self.tx = tx
		self.blocknumber = bigint
		self.state = Mined

		fmt.Println("tx is mined!!!!!")
		return nil
	case Mined:
		// TransactionReceipt returns the receipt of a mined transaction. Note that the
		// transaction may not be included in the current canonical chain even if a receipt
		// exists.
		recepit, err := client.TransactionReceipt(ctx, self.tx.Hash())
		if err != nil {
			return err
		}

		// out of gas, not confirmed!
		if recepit.GasUsed >= self.tx.Gas() {
			self.state = Unconfirmed
			return ErrorTxUnconfirmed{transactionHash: recepit.TxHash.Str(),
				gasUsed: recepit.GasUsed, gas: self.tx.Gas()}
		}
		self.state = WaitConfirmationNumber
		self.confirmed_number = 1
		return nil
	case WaitConfirmationNumber:
		if err := self.refresh_confrim_number(ctx, client); err != nil {
			if _, ok := err.(ErrorTxDisappear); ok {
				self.blocknumber.SetInt64(0)
				self.state = Unkown
				return err
			}
			return err
		} else if uint16(self.confirmed_number) >= self.confirmation_number {
			self.state = Confirmed
			return nil
		}
	default:
		return fmt.Errorf("wrong transaction state:%d", self.state)
	}
	return nil
}

func (self *Transfer) Identifer() string {
	return self.identifer
}
