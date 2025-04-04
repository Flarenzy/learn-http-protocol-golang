package server

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github/Flarenzy/learn-http-protocol-golang/internal/headers"
	"github/Flarenzy/learn-http-protocol-golang/internal/request"
	"github/Flarenzy/learn-http-protocol-golang/internal/response"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
)

type Server struct {
	port     int
	handler  Handler
	listener *net.TCPListener
	running  *atomic.Bool
}

type HandlerError struct {
	StatusCode   int
	ErrorMessage string
}

type Handler func(w *response.Writter, req *request.Request)

func newServer(port int, handler Handler, listener *net.TCPListener) *Server {
	r := atomic.Bool{}
	r.Store(true)
	return &Server{
		port:     port,
		handler:  handler,
		listener: listener,
		running:  &r,
	}
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		return nil, fmt.Errorf("internal error unable to make TCPListener")
	}
	if handler == nil {
		return nil, fmt.Errorf("error: need handler func")
	}
	s := newServer(
		port,
		handler,
		tcpListener,
	)
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	if s.listener == nil {
		s.running.Store(false)
		return nil
	}
	s.running.Store(false)
	err := s.listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) listen() {
	for {
		if !s.running.Load() {
			log.Printf("server isn't running\n")
			break
		}
		conn, err := s.listener.Accept()
		log.Printf("accepted conn at addr %s", conn.RemoteAddr())
		if err != nil {
			log.Printf("error: unable to accept connection. %s\n", err.Error())
			return
		}
		go func(conn net.Conn) {
			s.handle(conn)
			log.Printf("INFO: handeled conn on addr %s, clossing.\n", conn.RemoteAddr())
		}(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	if cw, ok := conn.(interface{ CloseWrite() error }); ok {
		defer cw.CloseWrite()
	} else {
		log.Printf("Connection doesn't implement CloseWrite method\n")
	}
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("ERROR: unable to parse request\n")
		return
	}
	w := response.NewWritter(conn)
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		proxyHanlder(w, req)
		return
	}
	s.handler(w, req)
}

func proxyHanlder(w *response.Writter, r *request.Request) {
	proxedResp, err := proxyToHttpbin(strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/"))
	if err != nil {
		log.Printf("ERROR: unable to proxy to httpbin\n")
		return
	}
	buf := make([]byte, 1024)
	err = w.WriteStatusLine(response.StatusCode(proxedResp.StatusCode))
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return
	}
	h := proxedResp.Header
	if v := h.Get("Content-Length"); v != "" {
		h.Del("Content-Length")
	}
	respHeaders := headers.NewHeaders()
	for k := range h {
		respHeaders.Set(k, h.Get(k))
	}
	respHeaders.Set("Transfer-Encoding", "chunked")
	respHeaders.Set("Trailer", "X-Content-Sha256, X-Content-Length")
	w.WriteHeaders(respHeaders)
	defer proxedResp.Body.Close()
	sumOfWrittenBytes := 0
	var fullBody []byte
	for {
		n, err := proxedResp.Body.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("DEBUG: reached EOF finished processing body.")
				_, err := w.WriteChunkedBodyDone()
				if err != nil {
					log.Printf("ERROR: unable to write chunked body done.\n")
					return
				}
				break
			}
			log.Printf("ERROR: unable to read body of response.\n")
			return
		}
		fullBody = append(fullBody, buf[:n]...)
		log.Printf("DEBUG: read %d number of bytes from proxy body.\n", n)
		numOfWrittenBytes, err := w.WriteChunkedBody(buf[:n])
		if err != nil {
			log.Printf("ERROR: unable to write chuned body.")
			return
		}
		log.Printf("DEBUG: wrote %d number of bytes as chunked body", numOfWrittenBytes)
		sumOfWrittenBytes += numOfWrittenBytes
	}

	sha := sha256.Sum256(fullBody)
	fullBodyLen := len(fullBody)
	fullBodyLenStr := strconv.Itoa(fullBodyLen)
	log.Printf("Content sha256: %s\n Len: %s", hex.EncodeToString(sha[:]), strconv.Itoa(fullBodyLen))
	trailers := headers.NewHeaders()
	trailers.Set("X-Content-Sha256", hex.EncodeToString(sha[:])) // Correct capitalization
	trailers.Set("X-Content-Length", fullBodyLenStr)

	err = w.WriteTrailers(trailers)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return
	}

}

func proxyToHttpbin(target string) (*http.Response, error) {
	url := fmt.Sprintf("https://httpbin.org/%s", target)
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// func writeHandlerError(w io.Writer, h HandlerError) error {
// 	msg := fmt.Sprintf("HTTP/1.1 %s \r\nConnection: close\r\n\r\n%s", strconv.Itoa(h.StatusCode), h.ErrorMessage)
// 	_, err := w.Write([]byte(msg))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func encode[T any](w http.ResponseWriter, _ *http.Request, status int, v T) error {
// 	w.Header().Set("Content-Type", "text/plain")
// 	w.WriteHeader(status)
// 	if err := json.NewEncoder(w).Encode(v); err != nil {
// 		return fmt.Errorf("encode json: %w", err)
// 	}
// 	return nil
// }

// func decode[T any](r *http.Request) (T, error) {
// 	var v T
// 	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
// 		return v, fmt.Errorf("decode json: %w", err)
// 	}
// 	return v, nil
// }
