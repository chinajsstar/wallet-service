package factory

import (
	"errors"
	"fmt"
	"unsafe"
	"z_wallet/client"
	"z_wallet/client/eth"
)

const (
	type_client_min = iota
	type_ethereum_client

	type_client_max = 1024
)

func NewClient(client_type int, rpc_url string) (*client.Rpc_client, error) {
	if client_type <= type_client_min || client_type >= type_client_max {
		return nil, fmt.Errorf("client_type:%d ", "not supported!")
	}

	var rpc_client *client.Rpc_client = nil
	var err error = nil
	if type_ethereum_client == client_type {
		rpc_client, err = new_eth_client(rpc_url)
	}

	if nil == rpc_client {
		if err == nil {
			err = errors.New(fmt.Sprintf("client_type:%d ", "not supported!"))
		}
	}
	return rpc_client, err
}

func new_eth_client(rpc_url string) (*client.Rpc_client, error) {
	rpc_client := (*client.Rpc_client)(unsafe.Pointer(new(eth.Client)))
	return rpc_client, nil
}
