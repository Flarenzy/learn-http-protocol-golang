package server

import (
	"bytes"
	"fmt"
	"github/Flarenzy/learn-http-protocol-golang/internal/request"
	"github/Flarenzy/learn-http-protocol-golang/internal/response"
	"io"
	"log"
	"net"
	"strconv"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

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
			log.Printf("server isn't running")
			break
		}
		conn, err := s.listener.Accept()
		log.Printf("accepted conn at addr %s", conn.RemoteAddr())
		if err != nil {
			log.Printf("error: unable to accept connection. %s", err.Error())
			return
		}
		go func(conn net.Conn) {
			s.handle(conn)
			log.Printf("INFO: handeled conn on addr %s, clossing.", conn.RemoteAddr())
		}(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	if cw, ok := conn.(interface{ CloseWrite() error }); ok {
		defer cw.CloseWrite()
	} else {
		log.Printf("Connection doesn't implement CloseWrite method")
	}
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("ERROR: unable to parse request")
	}
	buf := make([]byte, 0)
	buffer := bytes.NewBuffer(buf)
	handlerError := s.handler(buffer, req)
	if handlerError != nil {
		err = writeHandlerError(conn, *handlerError)
		if err != nil {
			log.Printf("ERROR: unable to write error to conn.")
		}
		return
	}
	err = response.WriteStatusLine(conn, response.StatusOk)
	if err != nil {
		log.Printf("ERROR: error writting status-line to conn")
		return
	}

	defaultHeaders := response.GetDefaultHeaders(buffer.Len())
	err = response.WriteHeaders(conn, defaultHeaders)
	if err != nil {
		log.Printf("ERROR: error writting headers to conn")
		return
	}

	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		log.Printf("ERROR: error writting body to conn")
	}
}

func writeHandlerError(w io.Writer, h HandlerError) error {
	msg := fmt.Sprintf("HTTP/1.1 %s \r\nConnection: close\r\n\r\n%s", strconv.Itoa(h.StatusCode), h.ErrorMessage)
	_, err := w.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

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
