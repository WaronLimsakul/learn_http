package request

import (
	"bytes"
	"io"
	"fmt"
	"strings"
	"strconv"

	"github.com/WaronLimsakul/learn_http/internal/headers"
)

type requestState int

const (
	initialized requestState = iota
	parsingHeaders
	parsingBody
	done
)
type Request struct {
	RequestLine RequestLine
	Headers headers.Headers
	Body []byte
	state requestState
}

type RequestLine struct {
	HttpVersion string
	RequestTarget string
	Method string
}


const crlf = "\r\n"


// Loop reading + parsing until the request is done or there are any error
func RequestFromReader(reader io.Reader) (*Request, error) {
	const bufferSize = 8
	buffer := make([]byte, bufferSize)
	req := Request {
		Headers: headers.NewHeaders(),
		Body: make([]byte, 0),
		state: initialized,
	}
	readIdx := 0
	for req.state != done {
		if readIdx >= len(buffer) {
			newBuff := make([]byte, len(buffer) * 2)
			copy(newBuff, buffer)
			buffer = newBuff
		}
		// read will read until it can't. So don't be scared of lost chunk
		read, err := reader.Read(buffer[readIdx:])
		// io.EOF in my implementation means we already read EVERYTHING
		// and there is NOT EVEN a BYTE to read from.
		if err == io.EOF {
			if req.state != done {
				return nil, fmt.Errorf("incomplete request")
			}
			break
		}
		readIdx += read
		consumed, err := req.parse(buffer[:readIdx])
		if err != nil {
			return nil, err
		}

		if consumed > 0 {
			// shift parsed data out
			newBuff := make([]byte, len(buffer))
			copy(newBuff, buffer[consumed:])
			buffer = newBuff
			// Don't forget to shif the index back, or there will
			// be a gap of nil in the buffer
			readIdx -= consumed
		}
	}
	return &req, nil
}

// - Manage data that stream to request and update the states
// - Try to parse as much as possible
func (r *Request) parse(data []byte) (int, error) {
	if r.state == done {
		return 0, fmt.Errorf("error trying to parse data with done request")
	}

	totalParsed := 0
	for r.state != done {
		bytesParsed, err := r.parseSingle(data[totalParsed:])
		totalParsed += bytesParsed
		if err != nil {
			return totalParsed, err
		} else if bytesParsed == 0 {
			break
		}
	}

	return totalParsed, nil
}

func (r *Request) parseSingle(data []byte) (bytesParsed int, err error) {
	switch r.state {
	case initialized:
		var requestLine *RequestLine
		bytesParsed, requestLine, err = parseRequestLine(data)
		if bytesParsed > 0 {
			r.RequestLine = *requestLine
			r.state = parsingHeaders
		}
	case parsingHeaders:
		var parsingDone bool
		bytesParsed, parsingDone, err = r.Headers.Parse(data)
		if parsingDone {
			r.state = parsingBody
		}
	case parsingBody:
		reportedLen, found := r.Headers.Get("Content-Length")
		if !found {
			r.state = done
			return 0, nil
		}
		r.Body = append(r.Body, data...)
		// I to A is "Int to ASCII".
		contentLen, err := strconv.Atoi(reportedLen)
		if err != nil {
			return 0, fmt.Errorf("invalid content-lenght field: %s", reportedLen)
		}
		if len(r.Body) > contentLen {
			return 0, fmt.Errorf(
				"bad content-lenght: reported: %d | actual: %d",
				contentLen, len(r.Body),
			)
		} else if len(r.Body) == contentLen {
			r.state = done
		}
		// If body still less than reported length, then it's ok
		// because we still not finish parsing
		return len(data), nil
	case done:
		return 0, nil
	default:
		return 0, fmt.Errorf("invalid request state: %v", r.state)
	}
	return
}

// return the bytes it consumed, request line ptr, err
func parseRequestLine(data []byte) (int, *RequestLine, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		// means need more data before parsing
		return 0, nil, nil
	}

	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return 0, nil, err
	}
	// return length it consumed (upto + crlf)
	return idx + 2, requestLine, nil
}

func requestLineFromString(s string) (*RequestLine, error) {
	parts := strings.Split(s, " ")
		// request line should have 3 parts
		if len(parts) != 3 {
			return nil, fmt.Errorf("couldn't parse request line: %s", s)
		}

		// first part: method
		method := parts[0]
		for _, ch := range method {
			if ch < 'A' || ch > 'Z' {
				return nil, fmt.Errorf("method not upper case: %s", parts[0])
			}
		}
		// second part: target (nothing to check)
		requestTarget := parts[1]

		// last part: version
		versionParts := strings.Split(parts[2], "/")
		if len(versionParts) != 2 {
			return nil, fmt.Errorf("malformed start-line: %s", s)
		}

		httpPart := versionParts[0]
		if httpPart != "HTTP" {
			return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
		}
		version := versionParts[1]
		if version != "1.1" {
			return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
		}

		return &RequestLine{
			Method: method,
			RequestTarget: requestTarget,
			HttpVersion: version,
		}, nil
}

func (r *Request) PrintRequest() {
	if r == nil {
		return
	}
	// request line
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", r.RequestLine.Method)
	fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
	if len(r.Headers) == 0 {
		return
	}
	fmt.Println("Headers:")
	for key, val := range r.Headers {
		fmt.Printf("- %s: %v\n", key, val)
	}
	if len(r.Body) == 0 {
		return
	}
	fmt.Println("Body:")
	fmt.Printf("%s\n", string(r.Body))
}
