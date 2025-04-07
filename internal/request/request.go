package request

import (
	"bytes"
	"io"
	"fmt"
	"strings"
)

type requestState int

const (
	initialized requestState = iota
	done
)
type Request struct {
	RequestLine RequestLine
	state requestState
}

type RequestLine struct {
	HttpVersion string
	RequestTarget string
	Method string
}


const crlf = "\r\n"


func RequestFromReader(reader io.Reader) (*Request, error) {
	const bufferSize = 8
	buffer, req, readIdx := make([]byte, bufferSize), Request{}, 0
	req.state = initialized
	for req.state != done {
		if readIdx >= len(buffer) {
			newBuff := make([]byte, len(buffer) * 2)
			copy(newBuff, buffer)
			buffer = newBuff
		}
		// read will read until it can't. So don't be scared of lost chunk
		read, err := reader.Read(buffer[readIdx:])
		if err == io.EOF {
			req.state = done
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
		}
	}
	return &req, nil
}

// should manage data that stream to request and update the states
func (r *Request) parse(data []byte) (int, error) {
	if r.state == done {
		return 0, fmt.Errorf("error trying to parse data with done request")
	}

	if r.state != initialized {
		return 0, fmt.Errorf("error unknown state")
	}

	consumed, requestLine, err := parseRequestLine(data)
	// if no error, means need more data
	if consumed == 0 {
		return 0, err
	}

	r.RequestLine = *requestLine
	r.state = done
	return consumed, nil
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

func PrintRequest(r *Request) {
	if r == nil {
		return
	}
	// request line
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", r.RequestLine.Method)
	fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
}
