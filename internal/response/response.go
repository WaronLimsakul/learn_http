package response

import (
	"fmt"
	"strconv"
	"strings"
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
	writingTrailers
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

func (w *Writer) WriteChunkedBody(p []byte) (n int, err error) {
	if w.state != writingBody {
		return 0, fmt.Errorf("invalid writer state: %d", w.state)
	}
	chunk := []byte{}
	firstLine := []byte(fmt.Sprintf("%X", len(p)) + crlf)
	chunk = append(chunk, firstLine...)
	chunk = append(chunk, p...)
	chunk = append(chunk, []byte(crlf)...)
	n, err = w.conn.Write(chunk)
	return
}

func (w *Writer) WriteChunkedBodyDone() (n int, err error) {
	if w.state != writingBody {
		return 0, fmt.Errorf("invalid writer state: %d", w.state)
	}
	n, err = w.conn.Write([]byte("0\r\n"))
	w.state = writingTrailers
	return
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writingTrailers {
		return fmt.Errorf("cannot writing in state: %v", w.state)
	}
	end := crlf
	trailerField, ok := h.Get("Trailer")
	// no trailer, can end with crlf right away
	if !ok {
		w.conn.Write([]byte(end))
		return nil
	}

	trailer := ""
	keys := strings.Split(trailerField, ", ")
	for _, key := range keys {
		val, ok := h.Get(key)
		if !ok {
			return fmt.Errorf("couldn't find key: %s in headers", key)
		}
		trailer += key + ":" + val + crlf
	}
	_, err := w.conn.Write([]byte(trailer + crlf))
	return err
}
