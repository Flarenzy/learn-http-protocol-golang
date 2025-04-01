package main

import (
	"github/Flarenzy/learn-http-protocol-golang/internal/request"
	"github/Flarenzy/learn-http-protocol-golang/internal/response"
	"github/Flarenzy/learn-http-protocol-golang/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

const badRequestResponse = `
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`

const internalErrorResponse = `
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`

const okResponse = `
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`

func handleRequest(w *response.Writter, r *request.Request) {
	if r == nil {
		return
	}
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		err := w.WriteStatusLine(400)
		if err != nil {
			log.Print("error writting status line to connection with status 400")
			return
		}
		headers := response.GetDefaultHeaders(len(badRequestResponse))
		headers.Set("Content-Type", "text/html")
		err = w.WriteHeaders(headers)
		if err != nil {
			log.Print("error writting headers to connection with status 400")
		}
		_, err = w.WriteBody([]byte(badRequestResponse))
		if err != nil {
			log.Print("error writting body to connection with status 400")
		}
		return
	case "/myproblem":
		err := w.WriteStatusLine(500)
		if err != nil {
			log.Print("error writting status line to connection with status 500")
			return
		}
		headers := response.GetDefaultHeaders(len(internalErrorResponse))
		headers.Set("Content-Type", "text/html")
		err = w.WriteHeaders(headers)
		if err != nil {
			log.Print("error writting headers to connection with status 500")
		}
		_, err = w.WriteBody([]byte(internalErrorResponse))
		if err != nil {
			log.Print("error writting body to connection with status 500")
		}
	default:
		err := w.WriteStatusLine(200)
		if err != nil {
			log.Print("error writting status line to connection with status 200")
			return
		}
		headers := response.GetDefaultHeaders(len(okResponse))
		headers.Set("Content-Type", "text/html")
		err = w.WriteHeaders(headers)
		if err != nil {
			log.Print("error writting headers to connection with status 200")
		}
		_, err = w.WriteBody([]byte(okResponse))
		if err != nil {
			log.Print("error writting body to connection with status 200")
		}
		return
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
