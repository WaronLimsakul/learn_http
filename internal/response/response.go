package response

import (
	"io"
	"fmt"
	"strconv"

	"github.com/WaronLimsakul/learn_http/internal/headers"
)

type StatusCode int
const (
	StatusOK StatusCode = 200
	StatusBadRequest = 400
	StatusServerError = 500
)

const crlf = "\r\n"


// not sure if we need to write crlf
func WriteStatusLine(w io.Writer, code StatusCode) error {
	statusLine := fmt.Sprintf("HTTP/1.1 %d", code)
	switch code {
		case 200:
			statusLine += " OK"
		case 400:
			statusLine += " Bad Request"
		case 500:
			statusLine += " Internal Server Error"
	}
	statusLine += crlf
	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "closed")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	resHeaders := ""
	for key, val := range headers {
		resHeaders += key + ":"
		resHeaders += " " + val
		resHeaders += crlf
	}
	resHeaders += crlf
	_, err := w.Write([]byte(resHeaders))
	return err
}
