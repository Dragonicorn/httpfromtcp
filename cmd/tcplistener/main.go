package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func buildLine(ch chan string, line string, lines []string) string {
	line += lines[0]
	if len(lines) > 1 {
		// fmt.Printf("sending line '%s' to channel...\n", line)
		ch <- line
		line = buildLine(ch, "", lines[1:])
	}
	return line
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	go func() {
		var (
			err  error
			line string
			n    int
		)
		b := make([]byte, 8)
		for {
			n, err = f.Read(b)
			if n == 0 && err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("error reading data - %v\n", err)
				close(ch)
				f.Close()
				os.Exit(1)
			}
			line = buildLine(ch, line, strings.Split(string(b[:n]), "\n"))
		}
		if len(line) > 0 {
			// fmt.Printf("sending last line '%s' to channel...\n", line)
			ch <- line
			close(ch)
			fmt.Printf("TCP connection closed\n")
			f.Close()
		}
	}()
	return ch
}

func main() {
	const (
		// fn string = "messages.txt"
		// host string = "127.0.0.1"
		port int = 42069
	)
	// var (
	// 	err   error
	// 	line  string
	// 	tcpC  *net.TCPConn
	// 	tcpEP *net.TCPAddr
	// 	tcpL  *net.TCPListener
	// )

	// f, err := os.Open(fn)
	// if err != nil {
	// 	fmt.Printf("error opening file '%s' - %v\n", f.Name(), err)
	// 	os.Exit(1)
	// }

	// tcpEP, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	// if err != nil {
	// 	fmt.Printf("error resolving TCP endpoint - %v\n", err)
	// 	os.Exit(1)
	// }

	// tcpL, err = net.ListenTCP("tcp", tcpEP)
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("error opening port %d - %v\n", port, err)
		os.Exit(1)
	}
	// defer tcpL.Close()
	defer l.Close()

	// tcpC, err = tcpL.AcceptTCP()
	c, err := l.Accept()
	if err != nil {
		fmt.Printf("error accepting connection on port %d - %v\n", port, err)
		os.Exit(1)
	}
	fmt.Printf("TCP connection accepted on port %d\n", port)
	for line := range getLinesChannel(c) {
		// fmt.Printf("read: %s\n", line)
		fmt.Println(line)
	}
}
