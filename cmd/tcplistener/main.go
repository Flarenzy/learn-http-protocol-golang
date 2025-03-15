package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	var buff [8]byte
	stringChan := make(chan string)
	go func() {
		defer close(stringChan)
		currentLineContent := ""
		for {
			read, err := f.Read(buff[:])
			if err != nil {
				if currentLineContent != "" {
					stringChan <- currentLineContent
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("ERROR: %s\n", err)
				return
			}
			splitLine := strings.Split(string(buff[:read]), "\n")
			for i := 0; i < len(splitLine)-1; i++ {
				stringChan <- fmt.Sprintf("%s%s", currentLineContent, splitLine[i])
				currentLineContent = ""
			}
			currentLineContent += splitLine[len(splitLine)-1]
		}

		err := f.Close()
		if err != nil {
			fmt.Printf("Error closing file: %s\n", err)
			return
		}
		fmt.Print("Connection closed.\n")
	}()
	return stringChan
}

func main() {
	fmt.Println("Listening on :42069")
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
	for {
		conn, err := listen.Accept()
		if err != nil {
			logger.Error("error getting a new connection", "err", err.Error())
			os.Exit(1)
		}
		fmt.Println("Connection accepted", conn.RemoteAddr())
		linesChan := getLinesChannel(conn)
		logger.Info("Starting reading")
		for line := range linesChan {
			fmt.Println(line)
		}

	}

}
