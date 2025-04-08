package server

import (
	"net"
	"fmt"
	"log"
	"sync/atomic"
	"github.com/WaronLimsakul/learn_http/internal/response"
)


// We don't need even a field to be public to be exported.
type Server struct {
	listener net.Listener
	isClosed atomic.Bool // use this type because it is thread-safe + sync
}

func localHost(port int) string {
	return fmt.Sprintf(":%d", port)
}

func Serve(port int) (*Server, error) {
	// get a listener at the port they want
	listener, err := net.Listen("tcp", localHost(port))
	if err != nil {
		return nil, err
	}

	server := Server{
		listener: listener,
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
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		log.Println("point 1: ", err)
		return
	}
	h := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, h)
	if err != nil {
		log.Println("point 2: ", err)
		return
	}
	return
}
