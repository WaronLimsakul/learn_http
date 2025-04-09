package main

import (
	"log"
	"os"
	"syscall"
	"os/signal"

	"github.com/WaronLimsakul/learn_http/internal/server"
	"github.com/WaronLimsakul/learn_http/internal/request"
	"github.com/WaronLimsakul/learn_http/internal/response"
)

const port = 42069

func main() {
	server, err := server.Serve(port, defaultHandler)
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

func defaultHandler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			w.WriteStatusLine(400)
			msg := []byte(
`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`)
			headers := response.GetDefaultHeaders(len(msg))
			headers.Reset("Content-Type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody(msg)
			return
		case "/myproblem":
			w.WriteStatusLine(500)
			msg := []byte(
`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`)
			headers := response.GetDefaultHeaders(len(msg))
			headers.Reset("Content-Type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody(msg)
			return
	}

	w.WriteStatusLine(200)
	msg := []byte(
`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`)
	headers := response.GetDefaultHeaders(len(msg))
	headers.Reset("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(msg)
}
