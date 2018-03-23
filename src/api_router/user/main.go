package main

import (
	"../base/nethelper"
	"../data"
	"fmt"
	"encoding/json" // for json get
	"sync/atomic"
	"net/rpc"
	"log"
	"time"
	"strconv"
)

var timeBegin,timeEnd time.Time

func DoTest(params interface{}, str *string, count *int64, right *int64, times int64){
	ackData := data.ServiceCenterDispatchAckData{}
	err := nethelper.CallJRPCToHttpServer("127.0.0.1:8080", "/wallet", data.MethodServiceCenterDispatch, params, &ackData)

	atomic.AddInt64(count, 1)
	if  err == nil && ackData.Err==0{
		atomic.AddInt64(right, 1)
	}

	if atomic.CompareAndSwapInt64(count, times, times) {
		cost := time.Now().Sub(timeBegin)
		fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
	}
}

func DoTest2(client *rpc.Client, params interface{}, str *string, count *int64, right *int64, times int64){
	ackData := data.ServiceCenterDispatchAckData{}
	err := nethelper.CallJRPCToHttpServerOnClient(client, data.MethodServiceCenterDispatch, params, &ackData)

	atomic.AddInt64(count, 1)
	if  err == nil && ackData.Err==0{
		atomic.AddInt64(right, 1)
	}

	if atomic.CompareAndSwapInt64(count, times, times) {
		cost := time.Now().Sub(timeBegin)
		fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
	}
}

func DoTestTcp(params interface{}, str *string, count *int64, right *int64, times int64){
	ackData := data.ServiceCenterDispatchAckData{}

	err := nethelper.CallJRPCToTcpServer("127.0.0.1:8090", data.MethodServiceNodeCall, params, &ackData)

	atomic.AddInt64(count, 1)
	if  err == nil && ackData.Err==0{
		atomic.AddInt64(right, 1)
	}

	if atomic.CompareAndSwapInt64(count, times, times) {
		cost := time.Now().Sub(timeBegin)
		fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
	}
}

func DoTestTcp2(client *rpc.Client, params interface{}, str *string, count *int64, right *int64, times int64){
	ackData := data.ServiceCenterDispatchAckData{}
	err := nethelper.CallJRPCToTcpServerOnClient(client, data.MethodServiceNodeCall, params, &ackData)

	atomic.AddInt64(count, 1)
	if  err == nil && ackData.Err==0{
		atomic.AddInt64(right, 1)
	}

	if atomic.CompareAndSwapInt64(count, times, times) {
		cost := time.Now().Sub(timeBegin)
		fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
	}
}

// http rpc风格
// curl -d '{"method":"ServiceCenter.Dispatch", "params":[{"version":"v1", "api":"arith.add","argv":"[{\"a\":\"hello, \", \"b\":\"world\"}]", "id":1}], "id": 1}' http://localhost:8080/rpc
// curl -d '{
// "method":"ServiceCenter.Dispatch",
// "params":[{"version":"v1", "api":"arith.add","argv":"[{\"a\":\"hello, \", \"b\":\"world\"}]}],
// "id": 1
// }'
// http://localhost:8080/rpc

// http restful风格
// curl -d '{"argv":"[{\"a\":2, \"b\":1}]", "id":1}' http://localhost:8080/restful/v1/arith/add
func main() {


	const times = 100;
	var count, right int64
	count = 0
	right = 0

	var testdata string
	for i := 0; i < 1000; i++ {
		testdata += strconv.Itoa(i)
	}
	testdata = "hello, world"

	dispatchData := data.ServiceCenterDispatchData{}
	dispatchData.Version = "v1"
	dispatchData.Api = "Arith.Add"
	dispatchData.Argv = "[{\"a\":1, \"b\":2}]"
	b,err := json.Marshal(dispatchData);
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return;
	}

	fmt.Println("argv:", string(b[:]))

	ackData := data.ServiceCenterDispatchAckData{}

	for ; ;  {
		fmt.Println("Please input command: ")
		var input string
		fmt.Scanln(&input)

		fmt.Println("Execute input command: ")
		count = 0
		right = 0
		timeBegin = time.Now();

		if input == "quit" {
			fmt.Println("I do quit")
			break;
		}else if input == "d1" {
			nethelper.CallJRPCToHttpServer("127.0.0.1:8080", "/wallet", data.MethodServiceCenterDispatch, dispatchData, &ackData)
			fmt.Println("ack==", ackData)
		}else if input == "d2" {
			nethelper.CallJRPCToTcpServer("127.0.0.1:8090", data.MethodServiceNodeCall, dispatchData, &ackData)
			fmt.Println("ack==", ackData)
		}else if input == "d3" {
			for i := 0; i < times; i++ {
				go DoTestTcp(dispatchData, &testdata, &count, &right, times)
			}
		} else if input == "d33" {

			client, err := rpc.Dial("tcp", "127.0.0.1:8090")
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}

			for i := 0; i < times*times*2; i++ {
				go DoTestTcp2(client, dispatchData, &testdata, &count, &right, times*times*2)
			}
		}else if input == "d4" {
			for i := 0; i < times; i++ {
				go DoTest(dispatchData, &testdata, &count, &right, times)
			}
		} else if input == "d44" {

			addr := "127.0.0.1:8080"
			log.Println("Call JRPC to Http server...", addr)

			client, err := rpc.DialHTTPPath("tcp", addr, "/wallet")
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}
			for i := 0; i < times*times*2; i++ {
				go DoTest2(client, dispatchData, &testdata, &count, &right, times*times*2)
			}
		}
	}
}
