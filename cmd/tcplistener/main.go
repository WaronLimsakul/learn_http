package main

import (
	"fmt"
	"net"

	"github.com/WaronLimsakul/learn_http/internal/request"
)

// TCP need a listener and handshake
// Pros: Packets arrive in order + guarantee no lost
// Cons: Slow, just send "packets" (bytes of data)
func main() {
	tcpListener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	defer tcpListener.Close()
	conn, err := tcpListener.Accept()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println("connection accepted")
	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	req.PrintRequest()
	fmt.Println("connection closed")
}
