package main

import (
	"log"
	"os"
	"syscall"
	"os/signal"

	"github.com/WaronLimsakul/learn_http/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port)
	if err != nil {
		log.Fatalf("Error start serving: %v\n", err)
	}

	defer server.Close()
	// I just realize that we can put another argument in Println
	log.Println("Server started on port ", port)

	// signal channel has buffer size 1
	sigChan := make(chan os.Signal, 1)
	// relay specified signal to our sigChan
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // wait until something come out of channel (sign to stop)
	log.Println("Server stopped gracefully")
}
