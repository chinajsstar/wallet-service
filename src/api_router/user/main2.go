package main

import (
	"golang.org/x/net/websocket"
	"fmt"
	"net/http"
	"strings"
	"../base/utils"
	"sync"
	"net/rpc/jsonrpc"
	"net/rpc"
	rpc2 "github.com/ethereum/go-ethereum/rpc"
	"net"
)

///
type Args struct {
	A string `json:"a"`
	B string `json:"b"`
}
type Arith int

func (arith *Arith)Add(args *Args, ) string  {
	fmt.Println("args:", args)
	return "ok"
}
///

type WsConn struct{

}

type WsServer struct{
	rwmu sync.RWMutex
	users map[*websocket.Conn]string
	clients map[*websocket.Conn]*WsConn
}

// http 处理
// rpcRequest represents a RPC request.
// rpcRequest implements the io.ReadWriteCloser interface.
type rpcRequest struct {
	data string     // holds the JSON formated RPC request
	conn *websocket.Conn // holds the JSON formated RPC response
	done chan bool     // signals then end of the RPC request
}

// Read implements the io.ReadWriteCloser Read method.
func (r *rpcRequest) Read(p []byte) (n int, err error) {
	p = []byte(r.data)
	l := len(p)
	fmt.Println(r.data)
	return l, nil
}

// Write implements the io.ReadWriteCloser Write method.
func (r *rpcRequest) Write(p []byte) (n int, err error) {
	return r.conn.Write(p)
}

// Close implements the io.ReadWriteCloser Close method.
func (r *rpcRequest) Close() error {
	r.done <- true
	return nil
}

// NewRPCRequest returns a new rpcRequest.
func newRPCRequest(conn *websocket.Conn, d string) *rpcRequest {
	done := make(chan bool)
	return &rpcRequest{d, conn, done}
}
func (ws *WsServer)handleData(conn *websocket.Conn, data string) {
	rpcReq := newRPCRequest(conn, data)

	// go and wait
	go rpc.ServeCodec(jsonrpc.NewServerCodec(rpcReq))
	<-rpcReq.done
}

func (ws *WsServer)closeConn(conn *websocket.Conn) {
	delete(ws.users, conn)
}

func (ws *WsServer)handleWebSocket(conn *websocket.Conn) {

	for {
		//判断是否重复连接
		if _, ok := ws.users[conn]; !ok {
			ws.users[conn] = "user"
			fmt.Println("a user is come in...")
		}

		fmt.Println("开始解析数据...")
		var data string
		err := websocket.Message.Receive(conn, &data)
		fmt.Println("data：", data, ",len", len(data))
		ws.handleData(conn, data)

		if err != nil {
			//移除出错的链接
			ws.closeConn(conn)
			fmt.Println("接收出错...")
			break
		}
	}
}
func (ws *WsServer)Start() error{
	rpc.Register(new(Arith))
	ws.users = make(map[*websocket.Conn]string)
	ws.clients = make(map[*websocket.Conn]*WsConn)

	//绑定socket方法
	http.Handle("/ws", websocket.Handler(ws.handleWebSocket))
	//开始监听
	err := http.ListenAndServe(":8400", nil)
	fmt.Println("err:", err)

	return err
}
func StartWsServer()  {
	wss := &WsServer{}
	wss.Start()
}

func StartWsClient() *websocket.Conn {
	conn, err := websocket.Dial("ws://127.0.0.1:8400/ws", "", "test://wallet/")
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

func main() {
	// Start a server and corresponding client.

	//var srv *rpc2.Server
	//var l net.Listener
	var client *rpc2.Client

	////
	var conn *websocket.Conn

	for ; ; {
		fmt.Println("Please input command: ")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		if argv[0]=="q"{
			break
		} else if argv[0] == "w" {
			//go StartWsServer()
			_, _ = StartWsServer2()
		}else if argv[0] == "c" {
			//conn = StartWsClient()
			client = StartWsClient2()
		}else if argv[0] == "s" {
			if(conn != nil){
				conn.Write([]byte(argv[1]))
			}
		} else if argv[0] == "ss" {
			var res string
			args := &Args{}
			args.A ="a"
			args.B = "b"
			if client != nil{
				client.Call(&res, "arith_add", args)
			}
			fmt.Println(res)
		}
	}

	return
}

