package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		// the empty line
		// headers are done, consume the CRLF
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	if len(parts) != 2 {
		return 0, false, fmt.Errorf("header must contain a colon")
	}
	key := string(parts[0])

	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)
	if !isValidFieldName(key) {
		return 0, false, fmt.Errorf("invalid char in field-name %s", key)
	}
	v, ok := h[key]
	if !ok {
		h.Set(strings.ToLower(key), string(value))
	} else {
		h.Set(strings.ToLower(key), fmt.Sprintf("%s, %s", v, string(value)))
	}

	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	h[key] = value
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func NewHeaders() Headers {
	h := make(Headers)
	return h
}

func isValidFieldName(s string) bool {
	specialCharsMap := specialCharSet()
	for _, c := range s {
		_, ok := specialCharsMap[c]
		if !(c >= 'A' && c <= 'Z') && !(c >= 'a' && c <= 'z') && !(c >= '0' && c <= '9') && !ok {
			return false
		}
	}
	return true
}

func specialCharSet() map[rune]bool {
	specialChars := []rune{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}
	specialCharsMap := make(map[rune]bool)
	for _, c := range specialChars {
		specialCharsMap[c] = true
	}
	return specialCharsMap
}
