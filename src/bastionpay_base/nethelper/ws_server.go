package nethelper

import (
	"sync"
	"net/http"
	"golang.org/x/net/websocket"
	l4g "github.com/alecthomas/log4go"
)

type WsServer struct{
	rwmu sync.RWMutex

	// valid clients
	clients map[*websocket.Conn]struct{}
}

func NewWsServer() *WsServer {
	ws := &WsServer{}
	ws.clients = make(map[*websocket.Conn]struct{})

	return ws
}

func (ws *WsServer)Start(addr string) error{
	//bind socket
	http.Handle("/ws", websocket.Handler(ws.handleWebSocket))
	//listen
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		l4g.Error("%s", err.Error())
	}

	return err
}

func (ws *WsServer)add(conn *websocket.Conn) error{
	var err error

	ws.rwmu.Lock()
	defer ws.rwmu.Unlock()

	ws.clients[conn] = struct{}{}

	return err
}

func (ws *WsServer)remove(conn *websocket.Conn) error{
	var err error

	conn.Close()

	ws.rwmu.Lock()
	defer ws.rwmu.Unlock()

	delete(ws.clients, conn)

	return err
}

func (ws *WsServer)handleData(conn *websocket.Conn, data string) error{
	l4g.Debug("data：%s, len: %d", data, len(data))
	if data == "hi" {
		// add client
		ws.add(conn)
	}

	return nil
}

func (ws *WsServer)handleWebSocket(conn *websocket.Conn) {
	for {
		l4g.Debug("handle a conn...")
		var err error
		var data string
		err = websocket.Message.Receive(conn, &data)
		if err == nil{
			err = ws.handleData(conn, data)
		}

		if err != nil {
			// remove client
			ws.remove(conn)
			l4g.Error("read filed: %s", err.Error())
			break
		}
	}
}