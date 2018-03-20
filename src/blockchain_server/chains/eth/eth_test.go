package eth

import (
	"testing"
	"fmt"
)

func TestNewAccount(t *testing.T) {
	account, _ := NewAccount()
	fmt.Printf("new account result: %s\n", account.Private_key)
}





