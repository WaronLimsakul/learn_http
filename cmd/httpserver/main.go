package main

import (
	"log"
	"os"
	"io"
	"os/signal"
	"net/http"
	"syscall"
	"strings"
	"fmt"
	"crypto/sha256"
	"strconv"
	"encoding/hex"

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
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		proxyHandler(w, req)
		return
	}

	switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			handle400(w, req)
			return
		case "/myproblem":
			handle500(w, req)
			return
	}

	handle200(w, req)

}

func proxyHandler(w *response.Writer, req *request.Request) {
	httpbinTarget := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	urlTarget := "https://httpbin.org" + httpbinTarget


	headers := response.GetDefaultHeaders(0)
	binResp, err := http.Get(urlTarget)
	if err != nil {
		handle500(w, req)
		return
	}

	headers.Set("Trailer", "X-Content-SHA256")
	headers.Set("Trailer", "X-Content-Length")

	defer binResp.Body.Close()
	w.WriteStatusLine(200)
	headers.Delete("Content-Length")
	headers.Set("Transfer-Encoding", "chunked")
	w.WriteHeaders(headers)

	buffer := make([]byte, 1024)
	fullBody := []byte{}
	for  {
		n, err := binResp.Body.Read(buffer)
		fmt.Println("n: ", n)
		if n > 0 {
			_, err := w.WriteChunkedBody(buffer[:n]) // when .Read(), it fill from start to n-1 bytes
			if err != nil {
				fmt.Println("error writing chunked body:", err)
				break
			}
			fullBody = append(fullBody, buffer[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("error reading from response body:", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("error writing last chunked body:", err)
	}

	hashedBody := sha256.Sum256(fullBody)
	headers.Set("X-Content-SHA256", hex.EncodeToString(hashedBody[:]))
	headers.Set("X-Content-Length", strconv.Itoa(len(fullBody)))

	err = w.WriteTrailers(headers)
	if err != nil {
		fmt.Println("error writing trailers:", err)
	}
}

func handle200(w *response.Writer, req *request.Request) {
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

func handle500(w *response.Writer, req *request.Request) {
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
}

func handle400(w *response.Writer, req *request.Request) {
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
}
