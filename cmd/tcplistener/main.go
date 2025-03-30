package main

import (
	"fmt"
	"github/Flarenzy/learn-http-protocol-golang/internal/request"
	"log/slog"
	"net"
	"os"
)

func main() {
	logFile, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	defer func(logFile *os.File) {
		err = logFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(logFile)
	listen, err := net.Listen("tcp", ":42069")
	if err != nil {
		logger.Error("error binding tcp port", "err", err.Error())
		os.Exit(1)
	}
	defer func(listen net.Listener) {
		err = listen.Close()
		if err != nil {
			fmt.Printf("error closing TCP listener %s", err.Error())
		}
	}(listen)
	logger.Info("Listening on :42069")
	for {
		conn, err := listen.Accept()
		if err != nil {
			logger.Error("error getting a new connection", "err", err.Error())
			os.Exit(1)
		}
		logger.Info("Connection accepted", "addr", conn.RemoteAddr())
		logger.Info("Starting reading request")
		req, err := request.RequestFromReader(conn)
		if err != nil {
			logger.Error("error parsing request from connection", "err", err.Error())
			os.Exit(1)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for h := range req.Headers {
			fmt.Printf("- %s: %s\n", h, req.Headers[h])
		}

	}

}
