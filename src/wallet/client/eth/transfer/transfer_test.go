package transfer

import (
	//"fmt"
	"testing"
	//"github.com/ethereum/go-ethereum/accounts/keystore"
	//"io/ioutil"
	//"golang.org/x/crypto/scrypt"
	//"io/ioutil"
	//"github.com/ethereum/go-ethereum/accounts"
	//"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/ethclient"
	"fmt"
	"math/big"
	//"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"context"
	"math"
	"time"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"wallet/client/eth"
)

func doTransfer(t *testing.T, ctx context.Context, client *ethclient.Client, transferchan chan *Transfer) error {
	t.Log("------------------do transfer begin!!!")
	ks := eth.DefaultTestKeyStore()
	//0x73a0d60b77f1045bac1ad25d681c51acbe679ff1

	if len(ks.Accounts())==0 {
		message := fmt.Sprintf("no account in key stroe!")
		t.Log(message)
		return fmt.Errorf(message)
	}

	ac := ks.Accounts()[0]
	gaslimit := 100000
	amount := big.NewInt(int64(math.Pow10(18)))
	toaddress := common.HexToAddress("0x73a0d60b77f1045bac1ad25d681c51acbe679ff1")

	gasprice, err := client.SuggestGasPrice(ctx)
	//gasprice, err := big.NewInt(int64(math.Pow10(18))), func() error {return nil}()
	if nil!=err {
		return err
	}
	nonce, err := client.PendingNonceAt(ctx, ac.Address)
	if nil!=err {
		return err
	}

	tx := types.NewTransaction(nonce, toaddress, amount, uint64(gaslimit), gasprice, nil)
	ks.Unlock(ac, "ko2005,./123eth")
	tx, err = ks.SignTx(ac, tx, big.NewInt(15))
	if err!=nil {
		return err
	}

	//fmt.Println("dd-mm-yyyy : ", current.Format("02-01-2006"))
	transfer := NewTransfer(tx, time.Now().Format("02-01-2006"))

	if err:=transfer.Send(ctx, client); err!=nil {
		return err
	}

	for {
		state := transfer.state
		if err:=transfer.RefreshState(ctx, client); err!=nil {
			fmt.Printf("error:%s\n", err.Error())
			if value, is := err.(ErrorTxUnconfirmed); is {
				return value
			} else if value, is := err.(ErrorTxDisappear); is {
				return value
			}
		} else{
			if state!=transfer.state {
				fmt.Printf("Transaction state changed, from:%d, to %d\n", state, transfer.state)
			}

			//fmt.Printf("current state is :%d\n", transfer.state)

			switch transfer.state {
			case Mined:
				fmt.Printf("trnasaction was mined on block number:%d", transfer.blocknumber.Uint64())
			case WaitConfirmationNumber:
				fmt.Printf("waiteconfirmationnumber: trnasaction was mined on block number:%d, waiting for confirmed nubmer, current is:%d\n", transfer.blocknumber.Uint64(), transfer.confirmed_number)
			case Confirmed:
				fmt.Printf("Confirmed:transaction confirmed, on block:%d, current confirmed number:%d\n", transfer.blocknumber.Uint64(), transfer.confirmed_number)
			case Unconfirmed:
				fmt.Println("Unconfirmed:transaction is unconfirmed!!")
				break
			default:
				t.Logf("it's imporsible to run here!, state is: %d\n", transfer.state)
			}
		}

		if (transfer.state==Confirmed && uint16(transfer.confirmed_number)>=transfer.confirmation_number) || transfer.state==Unconfirmed {
			fmt.Printf("transaction: %s, hash: 0x%x, is confirmed! ", transfer.identifer, transfer.tx.Hash())
			transferchan<- transfer
			break
		}

		isbreak := false
		select {
		case <- ctx.Done():
			isbreak = true
		default:
			time.Sleep(time.Second)
		}

		if isbreak {
			fmt.Printf("escape refresh state!")
			break
		}
	}
	return  nil
}

func tmpClient(rpc_url string) (client *ethclient.Client, err error) {
	rpc_client, err := rpc.Dial(rpc_url, )

	if err != nil {
	fmt.Println("can not dail to : ", rpc_url)
	return nil, err
	}

	//rpc.LatestBlockNumber
	if rpc_client == nil {
	return nil, fmt.Errorf("rpc client is nil")
	}

	client = ethclient.NewClient(rpc_client)
	return
}

func listen_new_transfer(t *testing.T, transferCmd chan interface{}) {
	for {
		// t.Log("listening new coming in transfer command")
		time.Sleep(5*time.Second)
	}
}

func TestTransfer(t *testing.T) {
	//client, err := tmpClient("http://127.0.0.1:8100")
	client, err := eth.NewWsClient("ws://127.0.0.1:8500")
	if err!=nil {
		t.Log("error:", err)
		return
	}
	t.Log("ethclient ok!!!")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 60 * 5)
	defer cancel()


	// 模型说明
	// 创建监听交易指令的channel(支持缓冲类型)
	// 创建交易指令完成的channel(支持缓冲类型)
	// select 读取以上两个channel
	// 任何一个channnel被触发, 就处理相关的任务, 循环

	// 监听服务的流程
	// 开启监听服务 -> 收到交易指令 -> 向交易指令channel队列输出具体命令

	// 交易goroutine
	transferCmdChan := make(chan interface{}, 1024)
	//go listen_new_transfer(t, transferCmdChan)

	transfer_chan := make(chan *Transfer, 1024)


	go func() {
		err := doTransfer(t, ctx, client, transfer_chan)
		if err!=nil {
			t.Log("do transfer error:", err)
			cancel()
		}
	}()

	select {
	case <- transferCmdChan:
		t.Log("new transfer command come in, add to transfer list")
	case transfer := <-transfer_chan:
		fmt.Printf("Transfer is returned from goroutine, Tx{id:%s, tx_hash:0x%x, state:%d}\n",
			transfer.identifer, transfer.tx.Hash(), transfer.state)
		t.Log("transfer ok!!!")
	case <-ctx.Done():
		t.Log("transaction error:", ctx.Err())
	}
}


func TestKeyStore(t *testing.T) {
	t.Log("do transfer start!")
	tks := eth.DefaultTestKeyStore()

	if tks==nil {
		t.Log("create key store failed")
		return
	}
	acs := tks.Accounts()

	t.Log("------------keystore accounts information------------")
	t.Logf("there are %d accounts in keystore\n", len(acs))
	t.Log("------------keystore accounts information------------")
	passpharse := "ko2005,./123eth"

	for _, c := range acs {
		t.Logf("account address:[0x%x]", c.Address)

		if err := tks.Unlock(c, passpharse); err!=nil {
			t.Log("\tunlock account error:", err)
		} else {
			t.Log("\tunlock ok.")
			tks.Lock(c.Address)
		}
	}

	t.Log("")
	t.Log("------------wallets accounts information------------")
	wlts := tks.Wallets()
	t.Logf("there are %d wallet in keystore\n", len(wlts))
	t.Log("------------wallets accounts information------------")
	for _, wlt :=  range wlts {

		walletDisplay(t, wlt)
	}

}

func walletDisplay(t *testing.T, wallet accounts.Wallet) {

	//wls := tks.Wallets()
	//t.Logf("keytore has %d wallets\n", len(wls))
	accs := wallet.Accounts()
	t.Logf("there are %d accounts in wallet\n", len(accs))
	//passpharse := "ko2005,./123eth"

	for _, ac := range accs {
		t.Logf("account address:[0x%x]", ac.Address)
	}

}

