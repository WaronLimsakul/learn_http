package headers

import (
	"bytes"
	"strings"
	"fmt"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

const crlf = "\r\n"

// parse one key-val pair to be in headers
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIdx := bytes.Index(data, []byte(crlf))
	// didn't find any crlf
	if crlfIdx == -1 {
		return 0, false, nil
	}
	// find crlf at the start, so no headers
	if crlfIdx == 0 {
		// only done when we found crlf separator
		return 2, true, nil
	}
	pair := string(data[:crlfIdx])
	key, val, err := getFieldLinePair(pair)
	if err != nil {
		return 0, false, err
	}
	if !validFieldName(key) {
		return 0, false, fmt.Errorf("invalid field-name: %s", key)
	}
	key = strings.ToLower(key)

	if _, ok := h[key]; ok {
		h[key] = h[key] + ", " + val
	} else {
		h[key] = val
	}

	return crlfIdx + 2, false, nil
}

// return trim value o
func getFieldLinePair(s string) (key, val string, err error) {
	s = strings.Trim(s, " ")
	key, val, found := strings.Cut(s, ":")
	if !found {
		return "", "", fmt.Errorf("Invalid field line: %s", s)
	}

	if strings.Contains(key, " ") {
		return "", "", fmt.Errorf("Invalid header name: %s", key)
	}

	val = strings.TrimLeft(val, " ")
	if strings.Contains(val, " ") {
		return "", "", fmt.Errorf("Invalid header value: %s", val)
	}

	return
}

func validFieldName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || strings.Contains("!#$%&'*+-.^_`|~", string(ch))) {
			return false
		}
	}
	return  true
}
