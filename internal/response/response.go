package response

import (
	"fmt"
	"github/Flarenzy/learn-http-protocol-golang/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func (s StatusCode) String() string {
	switch s {
	case StatusOk:
		return "OK"
	case StatusBadRequest:
		return "Bad Request"
	case StatusInternalServerError:
		return "Internal Server Error"
	default:
		return ""
	}
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, statusCode)
	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	if headers == nil {
		return fmt.Errorf("empty headers")
	}
	for k, v := range headers {
		fieldLine := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.Write([]byte(fieldLine))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil

}
