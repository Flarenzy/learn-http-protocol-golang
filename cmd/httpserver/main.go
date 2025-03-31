package main

import (
	"github/Flarenzy/learn-http-protocol-golang/internal/request"
	"github/Flarenzy/learn-http-protocol-golang/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func handleRequest(w io.Writer, r *request.Request) *server.HandlerError {
	if r == nil {
		return &server.HandlerError{
			StatusCode:   400,
			ErrorMessage: "Bad Request",
		}
	}
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode:   400,
			ErrorMessage: "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode:   500,
			ErrorMessage: "Woopsie, my bad\n",
		}
	default:
		_, err := w.Write([]byte("All good, frfr\n"))
		if err != nil {
			return &server.HandlerError{
				StatusCode:   500,
				ErrorMessage: err.Error(),
			}
		}
		log.Println("Wrote to io.Writer")
		return nil
	}
}

func main() {
	server, err := server.Serve(port, handleRequest)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
