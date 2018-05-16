package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/ethereum/go-ethereum"
	"testing"
	"context"
	"blockchain_server/types"
	etype "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"blockchain_server/conf"
	l4g "github.com/alecthomas/log4go"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"blockchain_server/chains/eth/token"
)

var (
	tmp_account = &types.Account{
		"0x04e2b6c9bfeacd4880d99790a03a3db4ad8d87c82bb7d72711b277a9a03e49743077f3ae6d0d40e6bc04eceba67c2b3ec670b22b30d57f9d6c42779a05fba097536c412af73be02d1642aecea9fa7082db301e41d1c3c2686a6a21ca431e7e8605f761d8e12d61ca77605b31d707abc3f17bc4a28f4939f352f283a48ed77fc274b039590cc2c43ef739bd3ea13e491316",
		"0x54b2e44d40d3df64e38487dd4e145b3e6ae25927"}
	tmp_toaddress = "0x0c14120e179f7dc6571467448fb3a7f7b14f889d"
	token_address = "0x27e7be9eaf092f27125ef867b87ed0adcce1431c"
)

func TestSendTokenTx(t *testing.T) {
	client, err := ethclient.Dial(config.MainConfiger().Clientconfig[types.Chain_eth].RPC_url)
	if nil!=err {
		l4g.Error("create eth client error! message:%s", err.Error())
		return
	}

	if false {
		tmp_tk, err := token.NewToken(common.HexToAddress(token_address), client);
		if err!=nil {
			l4g.Trace("SendTransactrion error:%s", err.Error())
			return
		}

		key, _ := ParseKey(tmp_account.PrivateKey)
		opts := bind.NewKeyedTransactor(key)
		tx, err := tmp_tk.Transfer(opts, common.HexToAddress(tmp_toaddress), big.NewInt(10))
		if err!=nil {
			l4g.Trace("SendTransactrion error:%s", err.Error())
			return
		}

		l4g.Trace("Tx information:%s", tx.String())

		//signedTx, _ := etype.SignTx(tx, etype.HomesteadSigner{}, key)
		//if err := client.SendTransaction(context.TODO(), signedTx); err!=nil {
		//	l4g.Trace("SendTransactrion error:%s", err.Error())
		//}
	}

	if true {
		if err:=transferToken (client,
			common.HexToAddress(token_address),
			tmp_account,
			common.HexToAddress(tmp_toaddress),
			big.NewInt(10)); err!=nil {
			l4g.Error("send transaction error:%s", err)
		}
	}
}

func transferToken(client * ethclient.Client, tokenAddress common.Address, from *types.Account, reciverAddress common.Address, value *big.Int) error {

	ABI, _:= abi.JSON(strings.NewReader(token.TokenABI))
	input, err := ABI.Pack("transfer", reciverAddress, value)
	if err!=nil {return err}

	value = new(big.Int)

	key, err := ParseKey(from.PrivateKey)
	if err!= nil {
		return err
	}
	fromAddress := crypto.PubkeyToAddress(key.PublicKey)
	nonce, err := client.PendingNonceAt(context.TODO(), fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(context.TODO())
	if err != nil {
		return err
	}

	code, err := client.PendingCodeAt(context.TODO(), tokenAddress)
	if err!=nil { return err } else if len(code) ==0 { return bind.ErrNoCode }
	l4g.Trace("code:0x%x\n", code)

	msg := ethereum.CallMsg{From: fromAddress, To: &tokenAddress, Value: value, Data: input}
	gasLimit, err := client.EstimateGas(context.TODO(), msg)
	if err != nil {
		return err
	}

	l4g.Trace("nonce:%d, to:0x%x, value:%d gaslimit:%d, gasprice:%d\ninput:0x%x",
		nonce, tokenAddress, value.Uint64(), gasLimit, gasPrice.Uint64(), input)
	rawTx := etype.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, input)
	fmt.Println("TX information:%s", rawTx.String())

	signer := etype.HomesteadSigner{}
	txhash := signer.Hash(rawTx).Bytes()
	l4g.Trace("tx_hash:0x%x", txhash)
	signature, err := crypto.Sign(signer.Hash(rawTx).Bytes(), key)
	if err != nil {
		return nil
	}
	signedTx, err := rawTx.WithSignature(signer, signature)

	//signedTx, err := etype.SignTx(rawTx, etype.HomesteadSigner{}, key)
	if err != nil {
		return err
	}
	return client.SendTransaction(context.TODO(), signedTx)
}


