package main

import (
	consulapi "github.com/hashicorp/consul/api"
	"fmt"
	"time"
	"net/http"
	"log"
	"github.com/benschw/dns-clb-go/clb"
)

// http handler
func handleCheck(w http.ResponseWriter, req *http.Request) {
	return
}
// start http server
func startHttpServer() error {
	// http
	log.Println("Start http server on ", "8000")

	http.Handle("/check", http.HandlerFunc(handleCheck))

	go func() {
		log.Println("Http server routine running... ")
		err := http.ListenAndServe(":8000", nil)
		if err != nil {
			fmt.Println("#Error:", err)
			return
		}
	}()

	return nil
}

type NodeRegistration struct{
	id string

	check *consulapi.AgentServiceCheck
	config *consulapi.Config
	client *consulapi.Client
}

func NewNodeRegistration(id string, config *consulapi.Config) (*NodeRegistration, error) {
	nr := &NodeRegistration{}
	nr.id = id
	if config != nil {
		nr.config = config
	}else{
		nr.config = consulapi.DefaultConfig()
	}

	nr.check = new(consulapi.AgentServiceCheck)

	var err error
	nr.client, err = consulapi.NewClient(nr.config)
	if err != nil {
		fmt.Println("#Error NewNodeRegistration: ", nr.config)
		return nil, err
	}
	return nr, nil
}

func (self *NodeRegistration) Register(name, addr, tags string, port int) error {
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = self.id
	registration.Name = name
	registration.Port = port
	registration.Tags = []string{tags}
	registration.Address = addr

	//check.Args = []string{"sh", "-c", "sleep 1 && exit 0"}
	//check.HTTP = fmt.Sprintf("http://%s:%d%s", registration.Address, registration.Port, "/check")
	//设置超时 5s。
	//check.Timeout = "5s"
	//设置间隔 5s。
	//check.Interval = "5s"
	//注册check服务。
	registration.Check = self.check

	return self.client.Agent().ServiceRegister(registration)
}

func (self *NodeRegistration)Deregister() error {
	return self.client.Agent().ServiceDeregister(self.id)
}


func main(){
	startHttpServer()

	config := consulapi.DefaultConfig()
	fmt.Println("default: ", config)

	client, err := consulapi.NewClient(config)
	if err != nil {
		fmt.Println("err: ", config)
		return
	}

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = "test"
	registration.Name = "web"
	registration.Port = 8000
	registration.Tags = []string{"rails"}
	registration.Address = "127.0.0.1"

	//增加check。
	check := new(consulapi.AgentServiceCheck)

	//check.Args = []string{"sh", "-c", "sleep 1 && exit 0"}
	check.HTTP = fmt.Sprintf("http://%s:%d%s", registration.Address, registration.Port, "/check")
	//设置超时 5s。
	check.Timeout = "5s"
	//设置间隔 5s。
	check.Interval = "5s"
	//注册check服务。
	registration.Check = check

	err = client.Agent().ServiceRegister(registration)

	time.Sleep(time.Second*1)
	for ; ;  {
		fmt.Println("Input 'quit' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "quit" {
			break;
		}else if input == "scan" {
			testClient()
		}
	}

	err = client.Agent().ServiceDeregister("test")
	if err != nil {
		fmt.Println("register server error : ", err)
	}
}

const (
	consulHost = "127.0.0.1"
	consulPort = "8600"
	srvName = "web.service.dc1.consul"
)

func testClient()  {
	c := clb.NewClb(consulHost, consulPort, clb.RoundRobin)
	address, err := c.GetAddress(srvName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("addr:", address)
}
