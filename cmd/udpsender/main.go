package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("INFO: Starting udp sender")
	addr, err := net.ResolveUDPAddr("udp4", "localhost:42069")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	udp, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return
	}
	defer func(udp *net.UDPConn) {
		err := udp.Close()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}(udp)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		readStr, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
		_, err = udp.Write([]byte(readStr))
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
