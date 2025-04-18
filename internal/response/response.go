package response

import (
	"fmt"
	"github/Flarenzy/learn-http-protocol-golang/internal/headers"
	"net"
	"strconv"
)

type writterState int

type Writter struct {
	conn  net.Conn
	state writterState
}

const (
	writeStatusLine writterState = iota
	writeHeaders
	writeBody
	writeDone
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

func (w *Writter) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writeStatusLine {
		return fmt.Errorf("error status line already written")
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, statusCode)
	_, err := w.conn.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	w.state = writeHeaders
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func (w *Writter) WriteHeaders(headers headers.Headers) error {
	if w.state != writeHeaders {
		return fmt.Errorf("error headers already written")
	}
	if headers == nil {
		return fmt.Errorf("empty headers")
	}
	for k, v := range headers {
		fieldLine := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.conn.Write([]byte(fieldLine))
		if err != nil {
			return err
		}
	}
	_, err := w.conn.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.state = writeBody
	return nil

}

func (w *Writter) WriteBody(p []byte) (int, error) {
	if w.state != writeBody {
		return 0, fmt.Errorf("error, writting body after close or before headers")
	}
	n, err := w.conn.Write(p)
	if err != nil {
		return 0, err
	}
	w.state = writeDone
	return n, nil
}

func NewWritter(conn net.Conn) *Writter {
	return &Writter{
		conn:  conn,
		state: writeStatusLine,
	}
}

func (w *Writter) WriteChunkedBody(p []byte) (int, error) {
	if w.state != writeBody {
		return 0, fmt.Errorf("error, writting body after close or before headers")
	}
	chunLen := len(p)
	chunkLenHex := fmt.Sprintf("%X\r\n", chunLen)
	var buf []byte
	buf = append(buf, []byte(chunkLenHex)...)
	buf = append(buf, p...)
	buf = append(buf, []byte("\r\n")...)
	n, err := w.conn.Write(buf)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (w *Writter) WriteChunkedBodyDone() (int, error) {
	w.state = writeDone
	w.conn.Write([]byte("0\r\n"))
	return 0, nil
}

func (w *Writter) WriteTrailers(h headers.Headers) error {
	if h == nil {
		return fmt.Errorf("no trailers to write")
	}
	for k, v := range h {
		line := []byte(fmt.Sprintf("%s: %s\r\n", k, v))
		_, err := w.conn.Write(line)
		if err != nil {
			return err
		}
	}
	w.conn.Write([]byte("\r\n"))
	return nil
}
