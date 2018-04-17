package business

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestHandleMsg(t *testing.T) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))

	svr := NewBusinessSvr()
	svr.InitAndStart()
	//s := "{\"user_id\":\"737205c4-af3c-426d-973d-165a0bf46c71\",\"method\":\"new_address\",\"params\":{\"id\":\"1\",\"symbol\":\"eth\",\"count\":3}}"
	s := "{\"user_id\":\"737205c4-af3c-426d-973d-165a0bf46c71\",\"method\":\"withdrawal\",\"params\":{\"user_order_id\":\"1\",\"symbol\":\"eth\",\"amount\":0.1,\"to_address\":\"0x00000\",\"user_timestamp\":0}}"
	var reply string
	svr.HandleMsg(s, &reply)
	fmt.Println(reply)

	time.Sleep(time.Second * 60 * 60)
	svr.Stop()
}
