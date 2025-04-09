package headers

import (
	"bytes"
	"strings"
	"fmt"
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
		return 0, false, fmt.Errorf("invalid field-name: '%s'", key)
	}
	h.Set(key, val)
	return crlfIdx + 2, false, nil
}

func (h Headers) Set(key, val string) {
	key = strings.ToLower(key)

	if _, ok := h[key]; ok {
		h[key] = h[key] + ", " + val
	} else {
		h[key] = val
	}
}

func (h Headers) Reset(key, val string) {
	key = strings.ToLower(key)
	h[key] = val
}

// return trim value o
func getFieldLinePair(s string) (key, val string, err error) {
	s  = strings.TrimSpace(s)
	key, val, found := strings.Cut(s, ":")
	if !found {
		return "", "", fmt.Errorf("Invalid field line: %s", s)
	}

	if strings.Contains(key, " ") {
		return "", "", fmt.Errorf("Invalid header name: %s", key)
	}

	val = strings.TrimSpace(val)

	return
}

func validFieldName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			strings.Contains("!#$%&'*+-.^_`|~", string(ch)) ) {
			return false
		}
	}
	return  true
}

func (h Headers) Get(s string) (val string, found bool) {
	val, found = h[strings.ToLower(s)]
	return
}

func (h Headers) Delete(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}
