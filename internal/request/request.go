package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func parseRequestLine(msg string) (*Request, error) {
	var (
		line  string
		parts []string
		req   Request
	)
	line = strings.Split(msg, "\r\n")[0]
	if len(line) == 0 {
		return nil, fmt.Errorf("error parsing request")
	}
	parts = strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed request - %s", line)
	}
	if parts[0] != strings.ToUpper(parts[0]) {
		return nil, fmt.Errorf("illegal method in request - %s", parts[0])
	}
	if !strings.Contains(parts[1], "/") {
		return nil, fmt.Errorf("illegal URL in request - %s", parts[1])
	}
	if parts[2] != "HTTP/1.1" {
		return nil, fmt.Errorf("unsupported version - %s", parts[2])
	}
	req.RequestLine.Method = parts[0]
	req.RequestLine.RequestTarget = parts[1]
	req.RequestLine.HttpVersion = strings.Split(parts[2], "/")[1]
	return &req, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	msg, err := io.ReadAll(reader)
	if err != nil {
		fmt.Errorf("error reading request - %v", err)
		return nil, err
	}
	return parseRequestLine(string(msg))
}

func main() {

}
