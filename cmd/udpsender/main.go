package main

import (
	"net"
	"fmt"
	"bufio"
	"os"
)

// UDP just connect to the address and send data. It doesn't care if there is a listener or not
// Pros: Fast (need no ack)
// Cons: Packets not in order (1% packet lost in avg)
func main() {
	// the instance of address
	addr, err := net.ResolveUDPAddr("udp", ":42069") // omit host (localhost:42069)
	if err != nil {
		fmt.Printf("point 1: ")
		fmt.Printf("error: %v\n", err)
		return
	}
	// get a udp connection from udp address
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("point 2: ")
		fmt.Printf("error: %v\n", err)
		return
	}

	defer conn.Close()
	stdinReader := bufio.NewReader(os.Stdin)
	for true {
		fmt.Printf("> ")
		line, err := stdinReader.ReadString(byte('\n')) // put the delimitor to stop and return string
		if err != nil {
			fmt.Printf("point 3: ")
			fmt.Printf("error: %v\n", err)
			return
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Printf("point 4: ")
			fmt.Printf("error: %v\n", err)
			return
		}
	}
}
