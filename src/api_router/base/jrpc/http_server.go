package jrpc

import (
	"net/http"
	"log"
	"context"
	"sync"
	"net"
	"fmt"
)

// Start Http Server
// @parameter: port string, like ":8080"
// @return: error
func StartHttpServer(port string) error {
	/*http.HandleFunc("/walletrpc", func(w http.ResponseWriter, req *http.Request) {
		log.Println("Http server Accept a client: ", req.RemoteAddr)

		defer req.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		res := NewRPCRequest(req.Body).Call()
		io.Copy(w, res)
	})*/

	log.Println("Start JRPC Http server on ", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Println("#StartJRPCHttpServer Error: ", err.Error())
		return err
	}

	return nil
}

// Start restful Http Server
// @parameter: port string, like ":8080"
// @return: error
func StartHttpServer2(ctx context.Context, wg *sync.WaitGroup, port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("#StartRestFulHttpServer Error:", err.Error())
		return err
	}

	srv := http.Server{Handler:nil}

	log.Println("Start Restful Http server on ", port)
	go srv.Serve(listener)

	fmt.Println("restful wait to quit")
	<-ctx.Done()
	listener.Close()
	fmt.Println("restful I do quit")

	return nil
}