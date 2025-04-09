package response

import (
	"fmt"
	"strconv"
	"net"

	"github.com/WaronLimsakul/learn_http/internal/headers"
)

type StatusCode int
const (
	StatusOK StatusCode = 200
	StatusBadRequest = 400
	StatusServerError = 500
)

type writerState int

const (
	initialized writerState = iota
	writingHeaders
	writingBody
	done
)

type Writer struct {
	conn net.Conn
	state writerState
}

const crlf = "\r\n"

func NewResponseWriter(conn net.Conn) *Writer {
	return &Writer{
		conn: conn,
		state: initialized,
	}
}

// not sure if we need to write crlf
func (w *Writer) WriteStatusLine(code StatusCode) error {
	if w.state != initialized {
		return fmt.Errorf("invalid writer state: %d", w.state)
	}
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
	_, err := w.conn.Write([]byte(statusLine))
	w.state = writingHeaders
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writingHeaders {
		return fmt.Errorf("invalid writer state: %d", w.state)
	}
	resHeaders := ""
	for key, val := range headers {
		resHeaders += key + ":"
		resHeaders += " " + val
		resHeaders += crlf
	}
	resHeaders += crlf
	_, err := w.conn.Write([]byte(resHeaders))
	w.state = writingBody
	return err
}

func (w *Writer) WriteBody(p []byte) (n int, err error) {
	if w.state != writingBody {
		return 0, fmt.Errorf("invalid writer state: %d", w.state)
	}
	n, err = w.conn.Write(p)
	w.state = done
	return
}
