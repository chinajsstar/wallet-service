package eth

import (
	//"github.com/ethereum/go-ethereum/crypto"
	//"github.com/ethereum/go-ethereum/ethclient"
	"testing"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	//"encoding/json"
	//"encoding/json"
	//"github.com/ethereum/go-ethereum/common/hexutil"
	//"github.com/ethereum/go-ethereum/ethclient"
	"blockchain_server/conf"
	"blockchain_server/types"
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
)

/*
func TestNewAccount(t *testing.T) {
	fmt.Println("*****************Testing NewAccount")
	account, _ := NewAccount()
	fmt.Printf("\"%s\", \n\"%s\"\n", account.PrivateKey, account.Address)
	//cryptKey := utils.String_cat_prefix(account.PrivateKey, "0x")
	//keyData, _ := hex.DecodeString(cryptKey)
	//keyPainData, _:= blockchain_server.Decrypto(keyData)

	//fmt.Printf("DecryptedPrivateKeyString: 0x%x\n", keyPainData)

	key, _ := ParseChiperkey(account.PrivateKey)
	fmt.Printf("decrypt----------------------\n")
	fmt.Printf("address:%s\n", crypto.PubkeyToAddress(key.PublicKey).String())
	fmt.Printf("publickey:%s\n", key.PublicKey.X.String())
	fmt.Printf("privatekey:0x%x\n", key.D.Bytes())
}
*/

/* this is a erc20 token Transaction! */
//> eth.getTransaction("0x8b5a6e64cec81a10fd40d0ade27ab19032efe38f9f94be5b51282f015b79e1b5")
//{
//  blockHash: "0x46a7cbe3c29680ee77543d438cda8dd677f6f1c126fd475d68a9577e669c87a0",
//  blockNumber: 5835,
//  from: "0xa40f6bf261914447987959ce26880d22eddf7dc6",
//  gas: 36588,
//  gasPrice: 18000000000,
//  hash: "0x8b5a6e64cec81a10fd40d0ade27ab19032efe38f9f94be5b51282f015b79e1b5",
//  input: "0xa9059cbb000000000000000000000000498d8306dd26ab45d8b7dd4f07a40d2c744f54bc000000000000000000000000000000000000000000000000000000000000000a",
//  nonce: 120,
//  r: "0x4a737cd5605235eeb9901820db866e31190b064e0b8ddd6aa673809be4c5801",
//  s: "0x56642a0f1c13bf9f9c133656ea217db1fb5e9160af85f9b6ac35120691cee665",
//  to: "0x27e7be9eaf092f27125ef867b87ed0adcce1431c",
//  transactionIndex: 0,
//  v: "0x42",
//  value: 0
//}
//> eth.getTransactionReceipt("0x8b5a6e64cec81a10fd40d0ade27ab19032efe38f9f94be5b51282f015b79e1b5")
//{
//  blockHash: "0x46a7cbe3c29680ee77543d438cda8dd677f6f1c126fd475d68a9577e669c87a0",
//  blockNumber: 5835,
//  contractAddress: null,
//  cumulativeGasUsed: 36588,
//  from: "0xa40f6bf261914447987959ce26880d22eddf7dc6",
//  gasUsed: 36588,
//  logs: [{
//      address: "0x27e7be9eaf092f27125ef867b87ed0adcce1431c",
//      blockHash: "0x46a7cbe3c29680ee77543d438cda8dd677f6f1c126fd475d68a9577e669c87a0",
//      blockNumber: 5835,
//      data: "0x000000000000000000000000000000000000000000000000000000000000000a",
//      logIndex: 0,
//      removed: false,
//      topics: ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "0x000000000000000000000000a40f6bf261914447987959ce26880d22eddf7dc6", "0x000000000000000000000000498d8306dd26ab45d8b7dd4f07a40d2c744f54bc"],
//      transactionHash: "0x8b5a6e64cec81a10fd40d0ade27ab19032efe38f9f94be5b51282f015b79e1b5",
//      transactionIndex: 0
//  }],
//  logsBloom: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000020000000008000000000000000000000000000800000000000000800080000000000000000000000000020000000000000000000010000000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000080000000000000000000000000000000000000000000000000",
//  root: "0xa32aeb6f508dd7efc7641d3bb0fe1b0536e729c67f9e7b8ba65565710ca2a87a",
//  to: "0x27e7be9eaf092f27125ef867b87ed0adcce1431c",
//  transactionHash: "0x8b5a6e64cec81a10fd40d0ade27ab19032efe38f9f94be5b51282f015b79e1b5",
//  transactionIndex: 0
//}
//transaction result: {"blockHash":"0x46a7cbe3c29680ee77543d438cda8dd677f6f1c126fd475d68a9577e669c87a0","blockNumber":"0x16cb","from":"0xa40f6bf261914447987959ce26880d22eddf7dc6","gas":"0x8eec","gasPrice":"0x430e23400","hash":"0x8b5a6e64cec81a10fd40d0ade27ab19032efe38f9f94be5b51282f015b79e1b5","input":"0xa9059cbb000000000000000000000000498d8306dd26ab45d8b7dd4f07a40d2c744f54bc000000000000000000000000000000000000000000000000000000000000000a","nonce":"0x78","to":"0x27e7be9eaf092f27125ef867b87ed0adcce1431c","transactionIndex":"0x0","value":"0x0","v":"0x42","r":"0x4a737cd5605235eeb9901820db866e31190b064e0b8ddd6aa673809be4c5801","s":"0x56642a0f1c13bf9f9c133656ea217db1fb5e9160af85f9b6ac35120691cee665"}
//Transaction receipt rpc result:{"blockHash":"0x46a7cbe3c29680ee77543d438cda8dd677f6f1c126fd475d68a9577e669c87a0","blockNumber":"0x16cb","contractAddress":null,"cumulativeGasUsed":"0x8eec","from":"0xa40f6bf261914447987959ce26880d22eddf7dc6","gasUsed":"0x8eec","logs":[{"address":"0x27e7be9eaf092f27125ef867b87ed0adcce1431c","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000a40f6bf261914447987959ce26880d22eddf7dc6","0x000000000000000000000000498d8306dd26ab45d8b7dd4f07a40d2c744f54bc"],"data":"0x000000000000000000000000000000000000000000000000000000000000000a","blockNumber":"0x16cb","transactionHash":"0x8b5a6e64cec81a10fd40d0ade27ab19032efe38f9f94be5b51282f015b79e1b5","transactionIndex":"0x0","blockHash":"0x46a7cbe3c29680ee77543d438cda8dd677f6f1c126fd475d68a9577e669c87a0","logIndex":"0x0","removed":false}],"logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000020000000008000000000000000000000000000800000000000000800080000000000000000000000000020000000000000000000010000000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000080000000000000000000000000000000000000000000000000","root":"0xa32aeb6f508dd7efc7641d3bb0fe1b0536e729c67f9e7b8ba65565710ca2a87a","to":"0x27e7be9eaf092f27125ef867b87ed0adcce1431c","transactionHash":"0x8b5a6e64cec81a10fd40d0ade27ab19032efe38f9f94be5b51282f015b79e1b5","transactionIndex":"0x0"}

// not token transaction!!!
// result : {"blockHash":"0xc0b8ad608d27af3789bd28d1454d5547e007c2c5686b9b5c577fb35a778dcca3","blockNumber":"0x16d0","from":"0xa40f6bf261914447987959ce26880d22eddf7dc6","gas":"0x15f90","gasPrice":"0x430e23400","hash":"0xa7be60a93699047b4d3b145529d9fb9bba01ab6e79e65d6581d0f5cb671b650a","input":"0x","nonce":"0x79","to":"0x498d8306dd26ab45d8b7dd4f07a40d2c744f54bc","transactionIndex":"0x0","value":"0x56bc75e2d63100000","v":"0x41","r":"0x4a0192cd411900eec828eb55b1365fa146590616e08495d1f5838ecc5d790ff7","s":"0x24098cff484acbf116cf979b0cf6848e45333b30e0a6b66ff1720ceea87bde4d"}
// > eth.getTransaction("0xa7be60a93699047b4d3b145529d9fb9bba01ab6e79e65d6581d0f5cb671b650a")
//{
//  blockHash: "0xc0b8ad608d27af3789bd28d1454d5547e007c2c5686b9b5c577fb35a778dcca3",
//  blockNumber: 5840,
//  from: "0xa40f6bf261914447987959ce26880d22eddf7dc6",
//  gas: 90000,
//  gasPrice: 18000000000,
//  hash: "0xa7be60a93699047b4d3b145529d9fb9bba01ab6e79e65d6581d0f5cb671b650a",
//  input: "0x",
//  nonce: 121,
//  r: "0x4a0192cd411900eec828eb55b1365fa146590616e08495d1f5838ecc5d790ff7",
//  s: "0x24098cff484acbf116cf979b0cf6848e45333b30e0a6b66ff1720ceea87bde4d",
//  to: "0x498d8306dd26ab45d8b7dd4f07a40d2c744f54bc",
//  transactionIndex: 0,
//  v: "0x41",
//  value: 100000000000000000000
//}
//Transaction receipt rpc result:{"blockHash":"0xc0b8ad608d27af3789bd28d1454d5547e007c2c5686b9b5c577fb35a778dcca3","blockNumber":"0x16d0","contractAddress":null,"cumulativeGasUsed":"0x5208","from":"0xa40f6bf261914447987959ce26880d22eddf7dc6","gasUsed":"0x5208","logs":[],"logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","root":"0xeb7bdd3a7ff33c7750a8f0a2e2ff2cadfb3f55fb770bcce42b18c27ef68243c0","to":"0x498d8306dd26ab45d8b7dd4f07a40d2c744f54bc","transactionHash":"0xa7be60a93699047b4d3b145529d9fb9bba01ab6e79e65d6581d0f5cb671b650a","transactionIndex":"0x0"}

//> eth.getTransactionReceipt("0xa7be60a93699047b4d3b145529d9fb9bba01ab6e79e65d6581d0f5cb671b650a")
//{
//  blockHash: "0xc0b8ad608d27af3789bd28d1454d5547e007c2c5686b9b5c577fb35a778dcca3",
//  blockNumber: 5840,
//  contractAddress: null,
//  cumulativeGasUsed: 21000,
//  from: "0xa40f6bf261914447987959ce26880d22eddf7dc6",
//  gasUsed: 21000,
//  logs: [],
//  logsBloom: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
//  root: "0xeb7bdd3a7ff33c7750a8f0a2e2ff2cadfb3f55fb770bcce42b18c27ef68243c0",
//  to: "0x498d8306dd26ab45d8b7dd4f07a40d2c744f54bc",
//  transactionHash: "0xa7be60a93699047b4d3b145529d9fb9bba01ab6e79e65d6581d0f5cb671b650a",
//  transactionIndex: 0
//}

func TestReceipt(t *testing.T) {
	client, err := ethclient.Dial(config.GetConfiger().Clientconfig[types.Chain_eth].RPC_url)
	if nil!=err {
		return
	}

	txHash := "0xa7be60a93699047b4d3b145529d9fb9bba01ab6e79e65d6581d0f5cb671b650a"
	txTokenHash := "0x8b5a6e64cec81a10fd40d0ade27ab19032efe38f9f94be5b51282f015b79e1b5"

	var hash string
	if false {
		hash=txHash
		fmt.Println("Not a erc20 transaction hash!!")
	} else {
		hash=txTokenHash
		fmt.Println("It is a erc20 transaction hash!!")
	}

	tx, err := client.TransactionByHash(context.TODO(), common.HexToHash(hash))
	if err==nil {
		data := tx.Data()
		if data!=nil {
			fmt.Printf("Transacton data:\n0x%x\n", data)
		}
		fmt.Printf("tranaction information:%s\n", tx.String())
	} else {
		fmt.Println("get Transaction by hash error:", err.Error())
	}

	receiptTx, err := client.TransactionReceipt(context.TODO(),
		common.HexToHash(hash))

	if err!=nil {
		fmt.Println("!!!!!!!!!!error message!!!!!!!!!!")
		fmt.Println(err.Error())
	}

	fmt.Println("Transaction receipt information:%s\n", receiptTx.String())
}

//type Receipt struct {
//	// Consensus fields
//	PostState         []byte `json:"root"`
//	Status            uint   `json:"status"`
//	CumulativeGasUsed hexutil.Uint64 `json:"cumulativeGasUsed"`
//	TxHash          common.Hash    `json:"transactionHash,string,omitempty""`
//	ContractAddress string `json:"contractAddress,string"`
//	GasUsed           hexutil.Uint64         `json:"gasUsed"`
//
//
//	From			  common.Address 	`json:"from"`
//	To				  common.Address	`json:"to"`
//	BlockNumber       hexutil.Uint64	`json:"blockNumber"`
//	BlockHash		  common.Hash 		`json:"blockHash"`
//	TxIndex			  hexutil.Uint64	`json:"transactionIndex"`
//
//}
//
//func TestUnmarshal(t *testing.T) {
//	receipt := &Receipt{}
//
//	jsonstring := `{"blockHash":"0xf6d3182a189699c663cef39d0b6d8af0281d28275e3946c03518243f815a2b3c","blockNumber":"0x1639","contractAddress":null,"cumulativeGasUsed":"0x5208","from":"0xa40f6bf261914447987959ce26880d22eddf7dc6","gasUsed":"0x5208","logs":[],"logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","root":"0x2d1114f35afb9c4e8a96496aee7c259261716da9b63a50deb0ce5b021c60e366","to":"0x498d8306dd26ab45d8b7dd4f07a40d2c744f54bc","transactionHash":"0x2f8f91210ae367750dcbad0cbed823ecd68ace2d0974549813c8cc0faea0156a","transactionIndex":"0x0"}`
//
//	if err:=json.Unmarshal([]byte(jsonstring), receipt); err!=nil {
//		fmt.Printf("err:%s\n", err)
//	}
//
//	fmt.Printf("block hash : 0x%x, blocknumber:%d\n", receipt.BlockHash, receipt.BlockNumber)
//	fmt.Println("ok")
//}

/*
func TestPendingNonceAt(t *testing.T) {
	fmt.Println("*****************Testing PendingNonceAt")
	client, err := ethclient.Dial("ws://127.0.0.1:8500")
	if err != nil {
		fmt.Printf("error:%s", err)
		return
	}
	fmt.Printf("address is : 0x54B2E44D40D3Df64e38487DD4e145b3e6Ae25927")
	nonce, err := client.PendingNonceAt(context.TODO(), common.HexToAddress("0x54B2E44D40D3Df64e38487DD4e145b3e6Ae25927"))
	if err != nil {
		fmt.Printf("error:%s", err)
		return
	}
	fmt.Printf("nonce is %d\n", nonce)
}
*/
