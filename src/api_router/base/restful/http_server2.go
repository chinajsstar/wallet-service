package restful

import (
	"net/http"
	"log"
	"context"
	"net"
	"fmt"
	"sync"
)

type HttpHandler interface {
	HandleRequest(w http.ResponseWriter, req *http.Request)error
}

type Handler struct {
	wg *sync.WaitGroup

	handler HttpHandler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.wg.Add(1)
	defer h.wg.Done()

	defer req.Body.Close()

	err := h.handler.HandleRequest(w, req)
	if err != nil {
		 w.WriteHeader(503)
		 w.Write([]byte(err.Error()))
	}
}

// Start restful Http Server
// @parameter: port string, like ":8080"
// @return: error
func StartRestfulHttpServer(ctx context.Context, wg *sync.WaitGroup, h HttpHandler, port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("#StartRestFulHttpServer Error:", err.Error())
		return err
	}

	handler := &Handler{}
	handler.wg = wg
	handler.handler = h
	srv := http.Server{Handler:handler}

	log.Println("Start Restful Http server on ", port)
	go srv.Serve(listener)

	fmt.Println("restful wait to quit")
	<-ctx.Done()
	listener.Close()
	fmt.Println("restful I do quit")

	return nil
}
