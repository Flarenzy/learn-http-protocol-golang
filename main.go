package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(f)
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
	var buff [8]byte
	var currentLineContent string
	for {
		read, err := f.Read(buff[:])
		if err != nil {
			if err == io.EOF {
				logger.Info("reached EOF")
				break
			}
			logger.Error("error reading from file", "err", err.Error())
		}
		splitLine := strings.Split(string(buff[:read]), "\n")
		if len(splitLine) == 1 {
			currentLineContent += splitLine[0]
			continue
		} else if len(splitLine) == 2 {
			currentLineContent += splitLine[0]
			fmt.Printf("read: %s\n", currentLineContent)
			currentLineContent = splitLine[1]
		} else {
			for i, line := range splitLine {
				if i == 0 {
					fmt.Printf("read: %s\n", currentLineContent+line)
					continue
				}
				if i == len(splitLine)-1 {
					currentLineContent = line
					break
				}
				fmt.Printf("read: %s\n", line)
			}
		}
	}
	if currentLineContent != "" {
		fmt.Printf("read: %s\n", currentLineContent)
	}

}
