package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	var buff [8]byte
	stringChan := make(chan string)
	var currentLineContent string
	go func() {
		defer close(stringChan)
		for {
			read, err := f.Read(buff[:])
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
			}
			splitLine := strings.Split(string(buff[:read]), "\n")
			if len(splitLine) == 1 {
				currentLineContent += splitLine[0]
				continue
			} else if len(splitLine) == 2 {
				currentLineContent += splitLine[0]
				stringChan <- currentLineContent
				currentLineContent = splitLine[1]
			} else {
				for i, line := range splitLine {
					if i == 0 {
						stringChan <- currentLineContent + line
						continue
					}
					if i == len(splitLine)-1 {
						currentLineContent = line
						break
					}
					stringChan <- line
				}
			}
		}
		if currentLineContent != "" {
			stringChan <- currentLineContent
		}
		err := f.Close()
		if err != nil {
			fmt.Printf("Error closing file: %s\n", err)
			return
		}
	}()
	return stringChan
}

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	defer func(logFile *os.File) {
		err = logFile.Close()
		if err != nil {
			fmt.Printf(err.Error())
		}
	}(logFile)

	outChan := getLinesChannel(f)
	logger.Info("Starting reading")
	for line := range outChan {
		fmt.Printf("read: %s\n", line)
	}

}
