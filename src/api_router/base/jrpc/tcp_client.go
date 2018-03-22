package jrpc

import (
	"net/rpc"
	"log"
)

// Call a JRPC to Tcp server
// @parameter: addr string, like "127.0.0.1:8080"
// @parameter: method string
// @parameter: params interface{}
// @parameter: res *string
// @return: error
func CallJRPCToTcpServer(addr string, method string, params interface{}, res interface{}) error {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Println("#CallJRPCToTcpServer Error: ", err.Error())
		return err
	}
	defer client.Close()

	return CallJRPCToTcpServerOnClient(client, method, params, res)
}

// Call a JRPC to Tcp server on a client
// @parameter: client rpc.Client
// @parameter: addr string, like "127.0.0.1:8080"
// @parameter: method string
// @parameter: params interface{}
// @parameter: res *string
// @return: error
func CallJRPCToTcpServerOnClient(client *rpc.Client, method string, params interface{}, res interface{}) error {
	err := client.Call(method, params, res)
	if err != nil {
		log.Println("#CallJRPCToTcpServerOnClient Error: ", err.Error())
		return err
	}

	return nil
}