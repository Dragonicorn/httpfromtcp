package request

import (
	"fmt"
	"io"
	"strings"
)

const BUFFER_SIZE int = 8

type parserState int

const (
	initialized parserState = iota
	done
)

type Request struct {
	ParserState parserState
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (req *Request) parseRequestLine(msg string) (int, error) {
	var (
		crlf  bool
		line  string
		n     int
		parts []string
	)
	line, _, crlf = strings.Cut(msg, "\r\n")
	// fmt.Printf("parseRequestLine msg: \"%s\"\n", msg)
	// fmt.Printf("parseRequestLine line: \"%s\"\n", line)
	// return number of bytes consumed - including CRLF
	n = len(line) + 2
	// return zero bytes consumed if no end-of-line in message
	if !crlf {
		return 0, nil
	}
	parts = strings.Split(line, " ")
	if len(parts) != 3 {
		return n, fmt.Errorf("malformed request - %s", line)
	}
	if parts[0] != strings.ToUpper(parts[0]) {
		return n, fmt.Errorf("illegal method in request - %s", parts[0])
	}
	if !strings.Contains(parts[1], "/") {
		return n, fmt.Errorf("illegal URL in request - %s", parts[1])
	}
	if parts[2] != "HTTP/1.1" {
		return n, fmt.Errorf("unsupported version - %s", parts[2])
	}
	// fill request structure with valid data
	req.RequestLine.Method = parts[0]
	req.RequestLine.RequestTarget = parts[1]
	req.RequestLine.HttpVersion = strings.Split(parts[2], "/")[1]
	return n, nil
}

func (req *Request) parse(data []byte) (int, error) {
	// fmt.Printf("Data to parse: \"%s\"\n", string(data))
	if req.ParserState == initialized {
		n, err := req.parseRequestLine(string(data))
		if n == 0 {
			return 0, nil
		}
		if err != nil {
			return n, err
		}
		req.ParserState = done
		// fmt.Printf("parse done - returning n: %d\n", n)
		return n, nil
	}
	if req.ParserState == done {
		return 0, fmt.Errorf("error trying to read data when already done")
	}
	return 0, fmt.Errorf("unknown state error")
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	var (
		buf, add        []byte
		err             error
		n, parsed, read int
		req             Request
	)
	buf = make([]byte, BUFFER_SIZE)
	req.ParserState = initialized
	for {
		n, err = io.ReadFull(reader, buf[read:])
		// fmt.Printf("Previous Buffer Data:    \"%s\" (Read: %d)\n", string(buf[:read]), read)
		// fmt.Printf("Newly Added Buffer Data: \"%s\" (Read: %d)\n", string(buf[read:]), n)
		// fmt.Printf("Entire Buffer Contents:  \"%s\" (Read: %d)\n", string(buf), read+n)
		// io.EOF error will be returned if no more data is available and NO data was read into buffer
		if err == io.EOF {
			req.ParserState = done
			break
		}
		// io.ErrUnexpectedEOF error will be returned if no more data is available but SOME data was read into buffer
		if err != nil && err != io.ErrUnexpectedEOF {
			fmt.Errorf("error reading request - %v", err)
			return nil, err
		}
		// update number of bytes read from the reader
		// fmt.Printf("%d bytes read from reader...\n", n)
		read += n
		// fmt.Printf("%d total bytes read...\n", read)
		n, err = req.parse(buf)
		if err != nil {
			fmt.Errorf("error parsing request - %v", err)
			return nil, err
		}
		// update number of bytes parsed from the buffer
		parsed += n
		// fmt.Printf("contents of buffer parsed so far: \"%s\" (Parsed: %d)\n", string(buf[:parsed]), parsed)
		if req.ParserState == done {
			// remove parsed data from buffer
			if parsed < len(buf) {
				copy(buf, buf[parsed:read])
			}
			read -= parsed
			parsed = 0
			fmt.Printf("contents of cleaned buffer: \"%s\" (Read: %d)\n", string(buf[:read]), read)
			break
		}
		// increase buffer size to keep existing data and accomodate more data from reader
		// fmt.Printf("%d bytes parsed from reader...\n", parsed)
		// fmt.Printf("buffer limit = %d, capacity = %d\n", len(buf), cap(buf))
		// fmt.Println("increasing buffer capacity...")
		add = make([]byte, len(buf)*2)
		copy(add, buf)
		buf = add
		// fmt.Printf("buffer limit = %d, capacity = %d\n\n", len(buf), cap(buf))
	}
	// fmt.Println("----------------")
	return &req, nil
}

func main() {

}
