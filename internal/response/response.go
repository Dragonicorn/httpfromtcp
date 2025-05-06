package response

import (
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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
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

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := make(headers.Headers)
	h["Content-Length"] = strconv.Itoa(contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
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
