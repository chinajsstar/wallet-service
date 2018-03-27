package business

import (
	//"encoding/json"
	//"fmt"
	//"reflect"
	//"database/sql"
	//"bytes"
	//"encoding/binary"
	"fmt"
	//"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	//"os/exec"
	"testing"
	//"time"
	"time"
)

type Info struct {
	Name string
	Age  int
}

type MyInfo struct {
	ID       string `json:"ID"`
	InfoList []Info `json:"InfoList"`
}

type Modu int

func (*Modu) String() {
	fmt.Println("test")
}

func TestHandleMsg(t *testing.T) {
	//req := "new_address"
	//var rsp string
	//HandleMsg(&req, &rsp)
	//fmt.Println(rsp)
	//strJson := "{\"ID\":\"1000\",\"InfoList\":[{\"Name\":\"liuxuliang\", \"Age\":35},{\"Name\":\"liuheng\", \"Age\":35}]}"
	//var myInfo MyInfo
	//json.Unmarshal([]byte(strJson), &myInfo)
	//fmt.Println(myInfo)
	//
	//j, _ := json.Marshal(myInfo)
	//fmt.Println(string(j))
	//
	//var md Modu
	//mdV := reflect.ValueOf(&md)
	//mdV.MethodByName("String").Call(nil)

	//out, err := exec.Command("uuidgen").Output()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("%s", out)

	//db, err := sql.Open("mysql", "dev_ws:D31c427e!@tcp(35.171.57.246:31379)/btc_wallet_deposit") //对应数据库的用户名和密码
	//db, err := sql.Open("mysql", "root:command@tcp(127.0.0.1:3306)/test") //对应数据库的用户名和密码
	//defer db.Close()
	//
	//if err != nil {
	//	panic(err)
	//} else {
	//	fmt.Println("success")
	//}
	//rows, err := db.Query("select iden, display from address;")
	//if err != nil {
	//	panic(err)
	//	return
	//}
	//for rows.Next() {
	//	var iden int
	//	var name string
	//	err = rows.Scan(&iden, &name)
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println(iden, name)
	//}

	//c, err := redis.Dial("tcp", "127.0.0.1:6379")
	//if err != nil {
	//	fmt.Println("Connect to redis error", err)
	//	return
	//}
	//defer c.Close()
	//
	//buf := new(bytes.Buffer)
	//info := Info{"liu", 23}
	//binary.Write(buf, binary.LittleEndian, info)
	//
	//fmt.Println(buf.Bytes())
	//
	//c.Do("sadd", "myinfo", buf)
	//
	//outbuf, err := redis.Bytes(c.Do("smembers", "myinfo"))
	//buf = bytes.NewBuffer(outbuf)
	//var infoout Info
	//binary.Read(buf, binary.LittleEndian, &infoout)
	//
	//fmt.Println(infoout)

	//clientManager := &service.ClientManager{}
	//client, err := eth.NewClient()
	//
	//if nil != err {
	//	fmt.Printf("create client:%s error:%s", types.Chain_eth, err.Error())
	//	return
	//}
	//
	//// add client instance to manager
	//clientManager.AddClient(client)
	//
	////ctx, cancel := context.WithCancel(context.Background())
	////defer cancel()
	//
	///*********批量创建账号示例*********/
	//accCmd := service.NewAccountCmd("message id", types.Chain_eth, 10)
	//var accs []*types.Account
	//accs, err = clientManager.NewAccounts(accCmd)
	//for i, account := range accs {
	//	fmt.Printf("account[%d], crypt private key:%s, address:%s\n",
	//		i, account.PrivateKey, account.Address)
	//}

	svr := NewBusinessSvr()
	svr.InitAndStart()
	s := "{\"user_id\":\"abc\",\"method\":\"new_address\",\"params\":{\"id\":\"1\",\"symbol\":\"eth\",\"count\":2}}"
	//s = "{\"user_id\":\"abc\",\"method\":\"withdrawal\",\"params\":{\"user_order_id\":\"1\",\"symbol\":\"eth\",\"amount\":1,\"to_address\":\"0x00000\",\"user_timestamp\":\"0xaaaaa\"}}"
	var reply string
	svr.HandleMsg(s, &reply)
	fmt.Println(reply)

	time.Sleep(time.Second * 60 * 60)
	svr.Stop()

}
