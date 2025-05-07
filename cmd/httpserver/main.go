package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dragonicorn/httpfromtcp/internal/headers"
	"github.com/dragonicorn/httpfromtcp/internal/request"
	"github.com/dragonicorn/httpfromtcp/internal/response"
	"github.com/dragonicorn/httpfromtcp/internal/server"
)

const port = 42069

func handler(w *response.Writer, req *request.Request) error {
	var (
		err error
		h   headers.Headers
		msg string
		sc  response.StatusCode
	)
	if req.RequestLine.RequestTarget == "/yourproblem" {
		sc = response.StatusCode400
		msg = "<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>\n"
	} else if req.RequestLine.RequestTarget == "/myproblem" {
		sc = response.StatusCode500
		msg = "<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>\n"
	} else {
		sc = response.StatusCode200
		msg = "<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>\n"
	}

	err = w.WriteStatusLine(sc)
	if err == nil {
		h = response.GetDefaultHeaders(len(msg))
		err = w.WriteHeaders(h)
		if err == nil {
			_, err = w.WriteBody([]byte(msg))
		}
	}
	return err
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port ", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
