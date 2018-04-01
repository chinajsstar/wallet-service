package nethelper

import (
	"log"
	"net/rpc"
	"bytes"
	"io/ioutil"
	"net/http"
)

// Call a JRPC to Http server
// @parameter: addr string, like "127.0.0.1:8080"
// @parameter: method string
// @parameter: params string
// @parameter: res *string
// @return: error
func CallJRPCToHttpServer(addr string, path string, method string, params interface{}, res interface{}) error {
	realPath := path
	if realPath == "" {
		realPath = rpc.DefaultRPCPath
	}
	client, err := rpc.DialHTTPPath("tcp", addr, realPath)
	if err != nil {
		log.Println("#CallJRPCToHttpServer Error: ", err.Error())
		return err
	}
	defer client.Close()

	return CallJRPCToHttpServerOnClient(client, method, params, res)
	if err != nil {
		log.Println("#CallJRPCToHttpServer Error: ", err.Error())
		return err
	}

	return nil
}

// Call a JRPC to Http server on a client
// @parameter: client
// @parameter: method string
// @parameter: params string
// @parameter: res *string
// @return: error
func CallJRPCToHttpServerOnClient(client *rpc.Client, method string, params interface{}, res interface{}) error {
	err := client.Call(method, params, res)
	if err != nil {
		log.Println("#CallJRPCToHttpServerOnClient Error: ", err.Error())
		return err
	}

	return nil
}

// Call a JRPC to Http server
// @parameter: addr string, like "127.0.0.1:8080"
// @parameter: path string
// @parameter: res *string
// @return: error
func CallToHttpServer(addr string, path string, body string, res *string) error {
	url := addr + path
	contentType := "application/json;charset=utf-8"

	b := []byte(body)
	b2 := bytes.NewBuffer(b)

	resp, err := http.Post(url, contentType, b2)
	if err != nil {
		log.Println("#CallToHttpServer Post failed:", err)
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("#CallToHttpServer Read failed:", err)
		return err
	}

	*res = string(content)
	return nil
}