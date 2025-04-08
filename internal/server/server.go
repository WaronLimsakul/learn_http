package server

import (
	"net"
	"fmt"
	"log"
	"io"
	"bytes"
	"sync/atomic"
	"github.com/WaronLimsakul/learn_http/internal/response"
	"github.com/WaronLimsakul/learn_http/internal/request"
)



// We don't need even a field to be public to be exported.
type Server struct {
	listener net.Listener
	handler Handler
	isClosed atomic.Bool // use this type because it is thread-safe + sync
}

// For Reporting error + write to body.
// The writer here is just buffer. Not a real connection.
type Handler func(w io.Writer, req request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message string
}

func localHost(port int) string {
	return fmt.Sprintf(":%d", port)
}

func Serve(port int, handler Handler) (*Server, error) {
	// get a listener at the port they want
	listener, err := net.Listen("tcp", localHost(port))
	if err != nil {
		return nil, err
	}

	server := Server{
		listener: listener,
		handler: handler,
		isClosed: atomic.Bool{}, // zero value is false
	}

	// send it to wait for request in the background
	go server.listen()
	return &server, nil
}

func (s *Server) Close() error {
	err := s.listener.Close()
	if err != nil {
		return fmt.Errorf("error closing server: %w", err)
	}
	s.isClosed.Store(true)
	return nil
}

func (s *Server) listen() {
	// don't have to check if it's closed here. Read the next line
	for {
		// Accept() stops and waits for the next request to come.
		// However, if the server (+ listener) get closed. It will return err right away,
		// so we catch that by check if server is closed inside and return if it is
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			log.Println("bad connection: ", err)
			continue
		}

		// after connecting with current request, handle it in background
		go s.handle(conn, s.handler)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		writeError(conn, &HandlerError{ StatusCode: 400 })
		return
	}

	// bytes.Buffer is a []byte but will be treated like a buffer.
	// It has Read, Write, ETC. So easy to work with.
	resBuff := bytes.Buffer{} // Buffer for handler to write as a reponse writer.

	// MUST return pointer of buffer because .Write implemented by *Buffer, not Buffer
	handlerError := s.handler(&resBuff, *req)
	if handlerError != nil {
		writeError(conn, handlerError)
		return
	}

	headers := response.GetDefaultHeaders(resBuff.Len())
	response.WriteStatusLine(conn, response.StatusOK)

	response.WriteHeaders(conn, headers)

	conn.Write(resBuff.Bytes())

	return
}

// intend to write it back to the connection directly
func writeError(conn io.Writer, hErr *HandlerError) error {
	err := response.WriteStatusLine(conn, hErr.StatusCode)
	if err != nil {
		return err
	}
	h := response.GetDefaultHeaders(len(hErr.Message))
	err = response.WriteHeaders(conn, h)
	if err != nil {
		return err
	}
	conn.Write([]byte(hErr.Message))
	return nil
}
