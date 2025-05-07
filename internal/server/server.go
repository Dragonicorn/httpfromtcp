package server

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/dragonicorn/httpfromtcp/internal/headers"
	"github.com/dragonicorn/httpfromtcp/internal/request"
	"github.com/dragonicorn/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request) error

type Server struct {
	Closed   atomic.Bool
	Listener net.Listener
	Handler  Handler
}

func videoHandler(w *response.Writer, req *request.Request) error {
	var (
		buffer []byte
		err    error
		h      headers.Headers
		n      int64
	)

	buffer, err = os.ReadFile("./assets/vim.mp4")
	if err != nil {
		return err
	}
	err = w.WriteStatusLine(response.StatusCode200)
	if err == nil {
		h = response.GetDefaultHeaders(len(buffer))
		h["Content-Type"] = "video/mp4"
		err = w.WriteHeaders(h)
		if err == nil {
			_, err = w.WriteBody(buffer)
			if err == nil {
				n, err = w.Body.WriteTo(w.Writer)
				if err != nil || n != int64(len(buffer)) {
					fmt.Printf("Error writing to connection: %v\n", err)
				}
			}
		}
	}
	return err
}

const BUFFER_SIZE int = 4096

func httpbinHandler(w *response.Writer, req *request.Request) error {
	var (
		body    bytes.Buffer
		buf     []byte
		err     error
		h       headers.Headers
		hash    hash.Hash
		b, l, n int
		// o   int64
		res *http.Response
		url string = "https://httpbin.org" + strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	)
	res, err = http.Get(url)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	if res.StatusCode > 299 {
		err = w.WriteStatusLine(response.StatusCode400)
		return fmt.Errorf("Error: Bad Status Code returned from %s\n", url)
	}
	buf = make([]byte, BUFFER_SIZE)

	err = w.WriteStatusLine(response.StatusCode200)
	if err != nil {
		return err
	}
	h = response.GetDefaultHeaders(0)
	delete(h, "Content-Length")
	h["Transfer-Encoding"] = "chunked"
	h["Content-Type"] = res.Header.Get("Content-Type")
	h["Trailer"] = "X-Content-SHA256, X-Content-Length"
	err = w.WriteHeaders(h)
	if err != nil {
		return err
	}
	w.State = response.StateChunkedBody
	for w.State != response.StateChunkedBodyDone {
		n, err = res.Body.Read(buf)
		// fmt.Printf("%d bytes read from httpbin.org\n", n)
		// io.EOF error will be returned if no more data is available and NO data was read into buffer
		// io.ErrUnexpectedEOF error will be returned if no more data is available but SOME data was read into buffer
		if (err == io.EOF && n == 0) || (err == io.ErrUnexpectedEOF) {
			err = nil
			w.State = response.StateChunkedBodyDone
		} else if err != nil {
			return err
		}
		if n > 0 {
			b, err = body.Write(buf[:n])
			if err != nil {
				return err
			}
			// update length of body received
			l += b
			// write chunk to response
			_, err = w.WriteChunkedBody(buf[:n])
			if err == nil {
				_, err = w.Body.WriteTo(w.Writer)
			}
			if err != nil {
				return err
			}
			// fmt.Printf("%d bytes written to response channel\n", o)
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err == nil {
		_, err = w.Body.WriteTo(w.Writer)
	}
	if err != nil {
		return err
	}

	// add trailers to chunked response
	clear(h)
	hash = sha256.New()
	hash.Write(body.Bytes())
	h["X-Content-SHA256"] = fmt.Sprintf("%x", hash.Sum(nil))
	h["X-Content-Length"] = fmt.Sprintf("%d", l)
	err = w.WriteTrailers(h)
	if err == nil {
		_, err = w.Body.WriteTo(w.Writer)
	}
	return err
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

	// check for proxy request to httpbin.org
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		var handler = s.Handler
		s.Handler = httpbinHandler
		err = s.Handler(&w, req)
		s.Handler = handler
	} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/video") {
		var handler = s.Handler
		s.Handler = videoHandler
		err = s.Handler(&w, req)
		s.Handler = handler
	} else {
		// handle standard request
		err = s.Handler(&w, req)
		if err == nil {
			cl, err = strconv.ParseInt((w.Headers["Content-Length"]), 10, 64)
			if err == nil {
				n, err = w.Body.WriteTo(c)
				if err != nil || n != cl {
					fmt.Printf("Error writing to connection: %v\n", err)
				}
			}
		}
	}
	if err != nil {
		fmt.Printf("Error in handler function: %v\n", err)
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
