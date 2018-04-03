package service

import (
	"fmt"
	"io"
	"sync"
)

type Connection struct {
	Cg *ConnectionGroup // connection group
	Conn io.ReadWriteCloser 	// holds the JSON formated RPC response
	Done chan bool 		// signals then end of the RPC request
}

// Read implements the io.ReadWriteCloser Read method.
func (r *Connection) Read(p []byte) (n int, err error) {
	n,err = r.Conn.Read(p)
	if err == io.EOF {
		r.Close()
	}
	return n, err
}

// Write implements the io.ReadWriteCloser Write method.
func (r *Connection) Write(p []byte) (n int, err error) {
	n, err = r.Conn.Write(p)
	return n, err
}

// Close implements the io.ReadWriteCloser Close method.
func (r *Connection) Close() error {
	r.Conn.Close()
	r.Done <- true
	if r.Cg != nil {
		r.Cg.remove(r)
	}
	return nil
}

type ConnectionGroup struct{
	rwmu sync.RWMutex
	connections map[*Connection]struct{}
}

func NewConnectionGroup() *ConnectionGroup {
	cg := &ConnectionGroup{}
	cg.connections = make(map[*Connection]struct{})

	return cg
}

func (cg *ConnectionGroup)Register(conn io.ReadWriteCloser) *Connection {
	cn := &Connection{}
	cn.Cg = cg
	cn.Conn = conn
	cn.Done = make(chan bool)

	cg.rwmu.Lock()
	defer cg.rwmu.Unlock()

	cg.connections[cn] = struct {}{}

	fmt.Println("connection count = ", len(cg.connections))

	return cn
}

func (cg *ConnectionGroup)remove(cn *Connection)  {
	cg.rwmu.Lock()
	defer cg.rwmu.Unlock()

	delete(cg.connections, cn)

	fmt.Println("connection count = ", len(cg.connections))
}