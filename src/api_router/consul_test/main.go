package main

import (
	consulapi "github.com/hashicorp/consul/api"
	"fmt"
	"net/http"
	"log"
	"github.com/benschw/dns-clb-go/clb"
	"../base/consulhelper"
	"../base/nethelper"
	"net/rpc"
	"strings"
	"../base/utils"
	"strconv"
)

// http handler
func handleCheck(w http.ResponseWriter, req *http.Request) {
	return
}
// start http server
func startHttpServer(port string) error {
	// http
	log.Println("Start http server on ", port)

	http.Handle("/check", http.HandlerFunc(handleCheck))

	go func() {
		log.Println("Http server routine running... ")
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			fmt.Println("#Error:", err)
			return
		}
	}()

	return nil
}

func startTcpServer(port string) {
	listener, err := nethelper.CreateTcpServer(":"+port)
	if err != nil {
		fmt.Println("#Error:", err)
		return
	}

	go func() {
		fmt.Println("Tcp server routine running... ")
		go func(){
			for{
				conn, err := listener.Accept();
				if err != nil {
					fmt.Println("%s", err.Error())
					continue
				}

				fmt.Println("Tcp server Accept a client: %s", conn.RemoteAddr().String())
				go func() {
					go rpc.ServeConn(conn)
					fmt.Println("Tcp server close a client: %s", conn.RemoteAddr().String())
				}()
			}
		}()
		fmt.Println("Tcp server routine stoped... ")
	}()
}

func main(){
	var nodes []*consulhelper.ConsulClient

	for ; ;  {
		fmt.Println("Input 'quit' to quit...")
		fmt.Println("Please input command: ")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		if argv[0] == "quit" {
			break;
		}else if argv[0] == "node" {
			cc := startNode(argv)
			if cc != nil{
				cc.Start()
				nodes = append(nodes, cc)
			}
		}else if argv[0] == "stop" {
			for _, cc := range nodes  {
				cc.Stop()
			}
		}else if argv[0] == "scan" {
			testClient(argv[1])
		}
	}
}

func startNode(node []string) *consulhelper.ConsulClient {
	if len(node) < 5 {
		fmt.Println("id name port tcp/http")
		return nil
	}
	cc := consulhelper.NewConsulClient()

	cc.Registration.ID = node[1] //"test"
	cc.Registration.Name = node[2] //"web"
	port, _ := strconv.Atoi(node[3])
	cc.Registration.Port = port //8000
	cc.Registration.Tags = []string{node[1], node[2]}
	cc.Registration.Address = "127.0.0.1"

	cc.Config = consulapi.DefaultConfig()
	//check.Args = []string{"sh", "-c", "sleep 1 && exit 0"}
	if node[4] == "tcp"{
		startTcpServer(node[3])

		cc.Check.TCP = fmt.Sprintf("%s:%d", cc.Registration.Address, cc.Registration.Port)
	}else if node[4] == "http"{
		startHttpServer(node[3])

		cc.Check.HTTP = fmt.Sprintf("http://%s:%d%s", cc.Registration.Address, cc.Registration.Port, "/check")
	}

	//设置超时 5s。
	cc.Check.Timeout = "5s"
	//设置间隔 5s。
	cc.Check.Interval = "5s"

	return cc
}

const (
	consulHost = "127.0.0.1"
	consulPort = "8600"
	srvName = ".service.dc1.consul"
)

func testClient(name string)  {
	c := clb.NewClb(consulHost, consulPort, clb.RoundRobin)
	address, err := c.GetAddress(name + srvName)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("addr:", address)
}
