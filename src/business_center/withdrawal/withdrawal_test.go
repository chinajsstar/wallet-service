package withdrawal

import "testing"

func TestWithdrawal_HandleWithdrawal(t *testing.T) {
	w := &Withdrawal{}
	w.Init(nil)
	req := "{\"user_id\":\"0001\",\"method\":\"withdrawal\",\"params\":{\"user_order_id\":\"1\",\"symbol\":\"eth\",\"amount\":0,\"to_address\":\"0x00000\",\"user_timestamp\":\"0xaaaaa\"}}"
	var rsp string
	w.HandleWithdrawal(req, &rsp)
}
