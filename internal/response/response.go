package response

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/dragonicorn/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusCode200 = iota
	StatusCode400
	StatusCode500
)

var (
	ReasonPhrases = map[StatusCode]string{
		StatusCode200: "200 OK",
		StatusCode400: "400 Bad Request",
		StatusCode500: "500 Internal Server Error",
	}
)

type WriteState int

const (
	StateStatus = iota
	StateHeader
	StateBody
	StateChunkedBody
	StateChunkedBodyDone
	StateTrailers
)

type Writer struct {
	Writer     io.Writer
	State      WriteState
	StatusCode StatusCode
	Headers    headers.Headers
	Body       bytes.Buffer
}

func writeStatusLine(w io.Writer, statusCode StatusCode) error {
	var (
		err error
		r   string
	)
	r = "HTTP/1.1 "
	if rp, ok := ReasonPhrases[statusCode]; ok {
		r += rp
	}
	r += "\r\n"
	_, err = w.Write([]byte(r))
	return err
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var (
		err error
	)
	if w.State == StateStatus {
		w.StatusCode = statusCode
		err = writeStatusLine(w.Writer, statusCode)
		if err == nil {
			w.State = StateHeader
		}
	} else {
		err = fmt.Errorf("Error: writing response status line out of sequence")
	}
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := make(headers.Headers)
	h["Content-Length"] = strconv.Itoa(contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/html"
	return h
}

func writeHeaders(w io.Writer, headers headers.Headers) error {
	var (
		err  error
		k, v string
	)
	for k, v = range headers {
		_, err = w.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			break
		}
	}
	_, err = w.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	var (
		err error
	)
	if w.State == StateHeader {
		w.Headers = headers
		err = writeHeaders(w.Writer, headers)
		if err == nil {
			w.State = StateBody
		}
	} else {
		err = fmt.Errorf("Error: writing response headers out of sequence")
	}
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	var (
		err error
		n   int
	)
	if w.State == StateBody {
		n, err = w.Body.Write(p)
		if err != nil || n != len(p) {
			err = fmt.Errorf("Error writing response body: %v\n", err)
		}
	} else {
		err = fmt.Errorf("Error: writing body out of sequence")
	}
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	var (
		err     error
		n, crlf int
	)
	if w.State == StateChunkedBody {
		w.Body.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
		n, err = w.Body.Write(p)
		if err != nil || n != len(p) {
			err = fmt.Errorf("Error writing response body chunk: %v\n", err)
		}
		crlf, err = w.Body.Write([]byte("\r\n"))
		if err != nil || crlf != 2 {
			err = fmt.Errorf("Error writing response body chunk: %v\n", err)
		}
		n += 2
	} else {
		err = fmt.Errorf("Error: writing chunked body out of sequence")
	}
	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int64, error) {
	var (
		err error
		n   int
	)
	if w.State == StateChunkedBodyDone {
		n, err = w.Body.Write([]byte("0\r\n"))
		if err == nil {
			w.State = StateTrailers
		}
	} else {
		err = fmt.Errorf("Error: writing chunked body end out of sequence")
	}
	return int64(n), err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	var (
		err error
	)
	if w.State == StateTrailers {
		err = writeHeaders(w.Writer, h)
		if err == nil {
			_, err = w.Body.Write([]byte("\r\n"))
		}
	} else {
		err = fmt.Errorf("Error: writing response trailers out of sequence")
	}
	return err
}
