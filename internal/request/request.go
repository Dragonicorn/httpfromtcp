package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dragonicorn/httpfromtcp/internal/headers"
)

const BUFFER_SIZE int = 8

type parserState int

const (
	requestStateInitialized parserState = iota
	requestStateParsingHeaders
	requestStateParsedHeader
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	ParserState parserState
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func fixCRLF(str string) string {
	return strings.Replace(strings.Replace(str, "\r", "<CR>", -1), "\n", "<LF>", -1)
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
	var (
		done bool
		n    int
		err  error
	)
	fmt.Printf("Data to parse: \"%s\" (%d bytes)\n", fixCRLF(string(data)), len(data))
	if req.ParserState == requestStateInitialized {
		n, err = req.parseRequestLine(string(data))
		if n == 0 {
			return 0, nil
		}
		if err != nil {
			return n, err
		}
		req.ParserState = requestStateParsingHeaders
		req.Headers = make(headers.Headers)
		fmt.Printf("\trequest parse done - returning n: %d\n", n)
		return n, nil
	}
	if req.ParserState == requestStateParsingHeaders {
		n, done, err = req.Headers.Parse(data)
		if n == 0 {
			return 0, err
		}
		if err != nil {
			return n, err
		}
		if done {
			req.ParserState = requestStateParsingBody
			fmt.Printf("\theaders parse done - returning n: %d\n", n)
		} else {
			req.ParserState = requestStateParsedHeader
			fmt.Printf("\theader parsed - returning n: %d\n", n)
		}
		return n, nil
	}
	if req.ParserState == requestStateDone {
		return 0, fmt.Errorf("error trying to read data when already done")
	}
	return 0, fmt.Errorf("unknown state error")
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	var (
		buf, add            []byte
		err                 error
		cl, n, parsed, read int
		req                 Request
		state               parserState // used to detect completion of parsing one entity
	)
	buf = make([]byte, BUFFER_SIZE)
	req.ParserState = requestStateInitialized
	state = req.ParserState
	for req.ParserState != requestStateDone {
		n, err = io.ReadFull(reader, buf[read:])
		// fmt.Printf("Previous Buffer Data:    \"%s\" (Read: %d bytes)\n", string(buf[:read]), read)
		// fmt.Printf("Newly Added Buffer Data: \"%s\" (Read: %d bytes)\n", string(buf[read:]), n)
		// fmt.Printf("Entire Buffer Contents:  \"%s\" (Read: %d bytes)\n", string(buf), read+n)
		// update number of bytes read from the reader
		// fmt.Printf("%d bytes read from reader...\n", n)
		read += n
		fmt.Printf("\t%d total bytes read...\n", read)
		// if err != nil {
		if ((err == io.EOF) && (n == 0)) || (err == io.ErrUnexpectedEOF) {
			// io.EOF error will be returned if no more data is available and NO data was read into buffer
			// if err == io.EOF {
			// 	if n == 0 {
			//fmt.Printf("\tnumber of bytes read n=%d, bytes in buffer read=%d\n", n, read)
			if req.ParserState == requestStateParsingBody {
				fmt.Printf("\tBody Length: %d / Content-Length: %d\n", read, cl)
				n = read
				err = nil
				if read > cl {
					n = cl
					err = fmt.Errorf("error - body length is greater than Content-Length indicated in header")
					fmt.Printf("%v\n", err)
				}
				if read < cl {
					err = fmt.Errorf("error - body length is less than Content-Length indicated in header")
					fmt.Printf("%v\n", err)
				}
				// copy body content from buffer to request body
				req.Body = make([]byte, n)
				copy(req.Body, buf[0:n])
				fmt.Printf("Data consumed: (%d bytes)\n", n)
				req.ParserState = requestStateDone
				break
			} else if read == 0 {
				err = nil
				req.ParserState = requestStateDone
				break
			}
			// }
			// io.ErrUnexpectedEOF error will be returned if no more data is available but SOME data was read into buffer
			// } else if err != io.ErrUnexpectedEOF {
		} else if err != nil {
			fmt.Errorf("error reading request - %v", err)
			return nil, err
		}
		// }
		// // update number of bytes read from the reader
		// // fmt.Printf("%d bytes read from reader...\n", n)
		// read += n
		// fmt.Printf("\t%d total bytes read...\n", read)
		if req.ParserState != requestStateParsingBody {
			n, err = req.parse(buf[:read])
			if err != nil {
				fmt.Errorf("error parsing request - %v", err)
				return nil, err
			}
			// update number of bytes parsed from the buffer
			parsed += n
		}
		// if no data parsed, increase buffer size keeping existing data
		if parsed == 0 {
			fmt.Printf("\tincreasing buffer size to ")
			add = make([]byte, len(buf)*2)
			copy(add, buf)
			buf = add
			fmt.Printf("%d/%d bytes\n", len(buf), cap(buf))
		} else {
			fmt.Printf("Contents of buffer parsed so far: \"%s\" (%d bytes)\n", fixCRLF(string(buf[:parsed])), parsed)
		}
		if req.ParserState != state {
			if req.ParserState == requestStateParsedHeader {
				req.ParserState = requestStateParsingHeaders
			}
			if req.ParserState == requestStateParsingBody {
				fmt.Println()
				if req.Headers.Get("Content-Length") == "" {
					fmt.Printf("\tNo Content-Length Header in request\n")
					cl = 0
				} else {
					cl, err = strconv.Atoi(req.Headers.Get("Content-Length"))
					if err != nil {
						fmt.Printf("error retrieving Content-Length from header - %v", err)
						cl = 0
					}
				}
				fmt.Printf("\tContent-Length (from header): %d\n", cl)
				if cl == 0 {
					req.ParserState = requestStateDone
				}
			}
			state = req.ParserState

			// remove parsed data from buffer
			if (0 < parsed) && (parsed < len(buf)) {
				copy(buf, buf[parsed:read])
			}
			read -= parsed
			parsed = 0
			fmt.Printf("Contents of cleaned buffer: \"%s\" (%d bytes)\n\n", fixCRLF(string(buf[:read])), read)
		}
	}
	// fmt.Println("----------------")
	fmt.Printf("\tBody: %s (%d bytes)\n", fixCRLF(string(req.Body)), len(req.Body))
	return &req, err
}

func main() {

}
