package business

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestHandleMsg(t *testing.T) {
	svr := NewBusinessSvr()
	svr.InitAndStart()
	s := "{\"user_id\":\"abc\",\"method\":\"new_address\",\"params\":{\"id\":\"1\",\"symbol\":\"eth\",\"count\":2}}"
	//s = "{\"user_id\":\"abc\",\"method\":\"withdrawal\",\"params\":{\"user_order_id\":\"1\",\"symbol\":\"eth\",\"amount\":1,\"to_address\":\"0x00000\",\"user_timestamp\":\"0xaaaaa\"}}"
	var reply string
	svr.HandleMsg(s, &reply)
	fmt.Println(reply)

	time.Sleep(time.Second * 10)
	svr.Stop()

}
