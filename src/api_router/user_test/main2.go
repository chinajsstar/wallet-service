package main

import (
	//"golang.org/x/net/websocket"
	"fmt"
	"golang.org/x/net/websocket"
	"encoding/json"
	"io/ioutil"
	//l4g "github.com/alecthomas/log4go"
	//"api_router/base/config"
	"reflect"
)

var G_admin_prikey2 []byte
var G_admin_pubkey2 []byte
var G_admin_licensekey2 string

var G_server_pubkey2 []byte

func LoadRsaKeys2() error {
	var err error
	G_admin_prikey2, err = ioutil.ReadFile("/Users/henly.liu/workspace/private_admin.pem")
	if err != nil {
		return err
	}

	G_admin_pubkey2, err = ioutil.ReadFile("/Users/henly.liu/workspace/public_admin.pem")
	if err != nil {
		return err
	}

	G_server_pubkey2, err = ioutil.ReadFile("/Users/henly.liu/wallet-service/src/api_router/account_srv/worker/public.pem")
	if err != nil {
		return err
	}

	G_admin_licensekey2 = "25143234-b958-44a8-a87f-5f0f4ef46eb5"

	return nil
}

func StartWsClient() *websocket.Conn {
	conn, err := websocket.Dial("ws://127.0.0.1:8040/ws", "", "test://wallet/")
	if err != nil {
		fmt.Println("#error", err)
		return nil
	}

	go func(conn *websocket.Conn) {
		var data []byte
		for ; ; {
			_, err := conn.Read(data)
			if err != nil {
				fmt.Println("read failed:", err)
				break
			}

			fmt.Println("read:", string(data))
		}
	}(conn)
	return conn
}
/*
func StartWsServer2() (*rpc2.Server, net.Listener) {
	startServer := func(addr string) (*rpc2.Server, net.Listener) {
		srv := rpc2.NewServer()
		srv.RegisterName("arith", new(Arith))
		l, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Println(err)
		}
		go http.Serve(l, srv.WebsocketHandler([]string{"*"}))
		return srv, l
	}

	srv, l1 := startServer("127.0.0.1:8300")
	fmt.Println(l1.Addr().String())
	return srv, l1
}

func StartWsClient2() *rpc2.Client {
	client, err := rpc2.Dial("ws://127.0.0.1:8300")
	if err != nil {
		fmt.Println("can't dial", err)
		return nil
	}
	return client
}
*/

type Args2 struct {
	A int `json:"-" comment:"加数1"`
	B int `json:"b" comment:"加数2"`
}

func FieldTag(v reflect.Value) string {
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var out string
	if t.Kind() == reflect.Struct {
		out += "\n{"
		n := t.NumField()
		fmt.Println(n)
		for i := 0; i < n; i++ {
			tagJson := t.Field(i).Tag.Get("json")
			if tagJson == "-" {
				continue
			}
			out += "\n" + tagJson + " // " + t.Field(i).Tag.Get("comment")
		}
		out += "\n}"
	}else if t.Kind() == reflect.Slice || t.Kind() == reflect.Array{
		n := v.Len()

		out += "\n["
		for i := 0; i < n && i < 1; i++ {
			//rs := v.Index(i)
			rs := v.Index(i)
			out += FieldTag(rs)
		}
		out += "\n]"

		//fmt.Println(n)
		//n := t.Len()
		//for i := 0; i < n; i++ {
		//	//out += t.Field(i).Tag.Get("json") + "--" + t.Field(i).Tag.Get("comment")
		//	fmt.Println(FieldTag(t.Field(i)))
		//}
	}else if t.Kind() == reflect.Map {
		ks := v.MapKeys()
		out += "\n{"
		for i := 0; i < len(ks); i++ {
			out += FieldTag(ks[i])
			out += ":"
			key := v.MapIndex(ks[i])
			out += FieldTag(key)
		}
		out += "\n}"

		//fmt.Println(n)
		//n := t.Len()
		//for i := 0; i < n; i++ {
		//	//out += t.Field(i).Tag.Get("json") + "--" + t.Field(i).Tag.Get("comment")
		//	fmt.Println(FieldTag(t.Field(i)))
		//}
	}else{
		out = "\n" + v.String()
	}

	return out
}

func main() {
	//appDir, _:= utils.GetCurrentDir()
	//appDir += "/test.json"
	//
	//cn := config.ConfigNode{}
	//err := config.LoadJsonNode(appDir, "node", &cn)
	//fmt.Println(cn)
	//fmt.Println(err)
	func(){
		fmt.Println("---------------")
		a2 := Args2{}
		a2.A = 1
		a2.B = 2
		b,_ := json.Marshal(a2)
		fmt.Println(string(b))

		//b1,_ := json.MarshalIndent(a2, "comment", "comment")
		//fmt.Println(string(b1))
	}()

	//func(){
	//	fmt.Println("---------------")
	//	var a2 []Args2
	//	a2 = append(a2, Args2{})
	//	b,_ := json.Marshal(a2)
	//	fmt.Println(string(b))
	//	fmt.Println(FieldTag(reflect.ValueOf(a2)))
	//}()

	func(){
		fmt.Println("---------------")
		var a2 map[string]Args2
		a2 = make(map[string]Args2)
		a2["abc"] = Args2{}
		b,_ := json.Marshal(a2)
		fmt.Println(string(b))
		fmt.Println(FieldTag(reflect.ValueOf(a2)))
	}()

	return

/*
	// Start a server and corresponding client.
	////
	LoadRsaKeys2()
	var conn *websocket.Conn

	for ; ; {
		fmt.Println("Please input command: ")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		if argv[0]=="q"{
			break
		} else if argv[0] == "w" {
		}else if argv[0] == "c" {
			conn = StartWsClient()
		}else if argv[0] == "s" {
			if(conn != nil){
				conn.Write([]byte(argv[1]))
			}
		}else if argv[0] == "login" {
			m, err := install.LoginUser()
			if err != nil {
				fmt.Println(err)
				continue
			}

			d, err := json.Marshal(m)
			if err != nil {
				fmt.Println(err)
				continue
			}

			var ud data.UserData
			encryptData(string(d), &ud)

			dispatchData := data.UserRequestData{}
			dispatchData.Method.Version = "v1"
			dispatchData.Method.Srv = "account"
			dispatchData.Method.Function = "login"
			dispatchData.Argv = ud

			d, err = json.Marshal(dispatchData)

			if conn != nil && err == nil{
				conn.Write(d)
			}
		}
	}
*/
	//return
}

