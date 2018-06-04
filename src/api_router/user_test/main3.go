package main

import (
	"fmt"
	"bastionpay_base/nethelper"
	"log"
	"net/rpc"
	"net"
	"strings"
	"bastionpay_api/utils"
	"bastionpay_base/service"
	l4g "github.com/alecthomas/log4go"
)

type Args3 struct {
	A string `json:"a"`
	B string `json:"b"`
}
type Arith3 int

func (arith *Arith3)Add(args *Args3, res *string) error  {
	fmt.Println("args:", args)
	return nil
}

var s_c *service.ConnectionGroup = service.NewConnectionGroup()
var c_c *service.ConnectionGroup = service.NewConnectionGroup()
func startServer() error {
	// tcp server
	err := func() error {
		listener, err := nethelper.CreateTcpServer(":8010")
		if err != nil {
			log.Println("#ListenTCP Error: ", err.Error())
			return err
		}
		go func() {
			log.Println("Tcp server routine running... ")
			go func(){
				for{
					conn, err := listener.Accept();
					if err != nil {
						log.Println("Error: ", err.Error())
						continue
					}

					log.Println("Tcp server Accept a client: ", conn.RemoteAddr())

					rc := s_c.Register(conn)
					go rpc.ServeConn(rc)
					<- rc.Done
					log.Println("Tcp server close a client: ", conn.RemoteAddr())
				}
			}()

		}()

		return nil
	}()

	return err
}

func main()  {
	rpc.Register(new(Arith3))

	var err error
	var client *rpc.Client
	var client1 *rpc.Client
	client1 = nil
	client = nil

	curDir, _ := utils.GetAppDir()
	l4g.LoadConfiguration(curDir + "/superwallet/log.xml")

	for ; ; {
		fmt.Println("Please input command: ")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		if argv[0]=="q"{
			break
		} else if argv[0] == "w" {
			go startServer()
			//_, _ = StartWsServer2()
		}else if argv[0] == "c" {
			a := &Args3{}
			a.B = "a"
			a.A = "b"
			var res string
			nethelper.CallJRPCToTcpServer("127.0.0.1:8010", "Arith3.Add", &a, &res)
			fmt.Println("res:", res)
		}else if argv[0] == "cc" {
			client, err = rpc.Dial("tcp", "127.0.0.1:8010")
			if err != nil {
				log.Println("#CallJRPCToTcpServer Error: ", err.Error())
				continue
			}
		} else if argv[0] == "ccc" {
			if client != nil {
				a := Args3{}
				a.B = "a"
				a.A = "b"
				var res string
				client.Call("Arith3.Add", &a, &res)
			}
		} else if argv[0] == "dd" {
			conn, err := net.Dial("tcp", "127.0.0.1:8010")
			if err != nil {
				fmt.Println(err)
				continue
			}

			rc := c_c.Register(conn)
			client1 = rpc.NewClient(rc)

			go func() {
				log.Println("Tcp client create a client: ", conn.RemoteAddr())
				<-rc.Done
				log.Println("Tcp client close a client: ", conn.RemoteAddr())
			}()

		}else if argv[0] == "ddd" {
			a := Args3{}
			a.B = "a"
			a.A = "b"
			var res string
			if client1 != nil {
				client1.Call("Arith3.Add", &a, &res)
			}
		}else if argv[0] == "log"{
			for i := 0; i < 1; i++ {
				l4g.Debug("i am log debug: %d", i)
			}
			l4g.Close()
		}
	}
}