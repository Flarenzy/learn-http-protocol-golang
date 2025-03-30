package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync/atomic"
)

type Server struct {
	port     int
	listener *net.TCPListener
	running  *atomic.Bool
}

func newServer(port int, listener *net.TCPListener) *Server {
	r := atomic.Bool{}
	r.Store(true)
	return &Server{
		port:     port,
		listener: listener,
		running:  &r,
	}
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		return nil, fmt.Errorf("internal error unable to make TCPListener")
	}
	s := newServer(
		port,
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
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!"))
	if err != nil {
		log.Printf("ERROR: error writting to conn")
	}
	if cw, ok := conn.(interface{ CloseWrite() error }); ok {
		defer cw.CloseWrite()
	} else {
		log.Printf("Connection doesn't implement CloseWrite method")
	}
}

func encode[T any](w http.ResponseWriter, _ *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
