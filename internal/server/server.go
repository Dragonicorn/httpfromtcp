package server

import (
	"fmt"
	// "io"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/dragonicorn/httpfromtcp/internal/request"
	"github.com/dragonicorn/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request) error

type Server struct {
	Closed   atomic.Bool
	Listener net.Listener
	Handler  Handler
}

func (s *Server) Close() error {
	s.Closed.Store(true)
	err := s.Listener.Close()
	return err
}

func (s *Server) listen() {
	for {
		c, err := s.Listener.Accept()
		if err != nil {
			if s.Closed.Load() {
				return
			}
			fmt.Printf("Error accepting connection: %v\n", err)
		} else {
			go s.handle(c)
		}
	}
}

func (s *Server) handle(c net.Conn) {
	var (
		err   error
		cl, n int64
		req   *request.Request
		w     response.Writer = response.Writer{
			Writer: c,
			State:  response.StateStatus,
			// StatusCode: response.StatusCode,
			// Headers:    headers.Headers,
			// Body:       bytes.Buffer,
		}
	)
	defer c.Close()

	// parse request from connection
	req, err = request.RequestFromReader(c)
	if err != nil {
		fmt.Printf("Error parsing request: %v\n", err)
		return
	}

	// handle request
	err = s.Handler(&w, req)
	if err != nil {
		fmt.Printf("Error in handler function: %v\n", err)
		return
	}
	cl, err = strconv.ParseInt((w.Headers["Content-Length"]), 10, 64)
	if err == nil {
		n, err = w.Body.WriteTo(c)
	}
	if err != nil || n != cl {
		fmt.Printf("Error writing to connection: %v\n", err)
	}
}

func Serve(port int, handler Handler) (*Server, error) {
	var (
		l      net.Listener
		server Server
		err    error
	)
	l, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Error opening port %d: %v\n", port, err)
		return nil, err
	}
	server.Handler = handler
	server.Listener = l
	server.Closed.Store(false)
	go server.listen()
	return &server, nil
}
