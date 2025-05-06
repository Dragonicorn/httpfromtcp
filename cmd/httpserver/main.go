package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dragonicorn/httpfromtcp/internal/request"
	"github.com/dragonicorn/httpfromtcp/internal/response"
	"github.com/dragonicorn/httpfromtcp/internal/server"
)

const port = 42069

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	var (
		err error
		he  server.HandlerError
	)
	if req.RequestLine.RequestTarget == "/yourproblem" {
		he.Status = response.StatusCode400
		he.Message = "Your problem is not my problem\n"
	} else if req.RequestLine.RequestTarget == "/myproblem" {
		he.Status = response.StatusCode500
		he.Message = "Woopsie, my bad\n"
	} else {
		he.Status = response.StatusCode200
		_, err = w.Write([]byte("All good, frfr\n"))
		if err != nil {
			fmt.Printf("Error in handler writing response body: %v\n", err)
		}
	}
	return &he
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
