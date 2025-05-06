package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/dragonicorn/httpfromtcp/internal/request"
	"github.com/dragonicorn/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	Status  response.StatusCode
	Message string
}

type Server struct {
	Closed   atomic.Bool
	Listener net.Listener
	Handler  Handler
}

func WriteHandlerError(w io.Writer, he *HandlerError) {
	w.Write([]byte(fmt.Sprintf("Error: %s\r\n%s\r\n", response.ReasonPhrases[he.Status], he.Message)))
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

// Primeagen's handle function
// func (s *Server) handle(conn net.Conn) {
// 	defer conn.Close()
// 	req, err := request.RequestFromReader(conn)
// 	if err != nil {
// 		hErr := &HandlerError{
// 			StatusCode: response.StatusCodeBadRequest,
// 			Message:    err.Error(),
// 		}
// 		hErr.Write(conn)
// 		return
// 	}
// 	buf := bytes.NewBuffer([]byte{})
// 	hErr := s.handler(buf, req)
// 	if hErr != nil {
// 		hErr.Write(conn)
// 		return
// 	}
// 	b := buf.Bytes()
// 	response.WriteStatusLine(conn, response.StatusCodeSuccess)
// 	headers := response.GetDefaultHeaders(len(b))
// 	response.WriteHeaders(conn, headers)
// 	conn.Write(b)
// 	return
// }

func (s *Server) handle(c net.Conn) {
	var (
		body bytes.Buffer
		he   *HandlerError
		err  error
		n    int64
		req  *request.Request
	)
	// parse request from connection
	req, err = request.RequestFromReader(c)
	// handle request
	// fmt.Printf("handle - req.RequestLine.RequestTarget: %s\n", req.RequestLine.RequestTarget)
	he = s.Handler(&body, req)
	// fmt.Printf("handle - he.Status: %v\n", he.Status)
	if he.Status != response.StatusCode200 {
		WriteHandlerError(&body, he)
	}
	err = response.WriteStatusLine(c, he.Status)
	if err == nil {
		err = response.WriteHeaders(c, response.GetDefaultHeaders(body.Len()))
		if err == nil {
			n, err = body.WriteTo(c)
		}
	}
	if err != nil || n != int64(body.Len()) {
		fmt.Printf("Error writing to connection: %v\n", err)
	}
	c.Close()
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
