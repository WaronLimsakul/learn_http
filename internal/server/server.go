package server

import (
	"net"
	"fmt"
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

// We can write response inside handler.
type Handler func(w *response.Writer, req *request.Request)

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
			continue
		}

		// after connecting with current request, handle it in background
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		resWriter := response.NewResponseWriter(conn)
		msg := []byte("couldn't parse request")
		resWriter.WriteStatusLine(400)
		headers := response.GetDefaultHeaders(len(msg))
		resWriter.WriteHeaders(headers)
		resWriter.WriteBody(msg)
		return
	}

	// bytes.Buffer is a []byte but will be treated like a buffer.
	// It has Read, Write, ETC. So easy to work with.
	// resBuff := bytes.Buffer{} // Buffer for handler to write as a reponse writer.

	resWriter := response.NewResponseWriter(conn)

	s.handler(resWriter, req)
}

// intend to write it back to the connection directly
// func writeError(conn io.Writer, hErr *HandlerError) error {
// 	err := response.WriteStatusLine(conn, hErr.StatusCode)
// 	if err != nil {
// 		return err
// 	}
// 	h := response.GetDefaultHeaders(len(hErr.Message))
// 	err = response.WriteHeaders(conn, h)
// 	if err != nil {
// 		return err
// 	}
// 	conn.Write([]byte(hErr.Message))
// 	return nil
// }
