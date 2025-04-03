package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	const (
		host string = "localhost"
		port int    = 42069
	)
	var (
		err   error
		line  string
		rd    *bufio.Reader
		udpC  *net.UDPConn
		udpEP *net.UDPAddr
	)

	udpEP, err = net.ResolveUDPAddr("udp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		fmt.Printf("error resolving UDP endpoint - %v\n", err)
		os.Exit(1)
	}

	udpC, err = net.DialUDP("udp", nil, udpEP)
	if err != nil {
		fmt.Printf("error opening port %d - %v\n", port, err)
		os.Exit(1)
	}
	defer udpC.Close()

	rd = bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		line, err = rd.ReadString('\n')
		_, err = udpC.Write([]byte(line))
		if err != nil {
			fmt.Printf("error writing to UDP port - %v\n", err)
		}
	}
}
