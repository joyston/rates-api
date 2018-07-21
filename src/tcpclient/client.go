package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	// connect to this socket
	conn, err := net.Dial("tcp", "127.0.0.1:4000")
	if err != nil {
		fmt.Println("***Error Here")
	}

	for {

		//read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		//read from socket
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			break
		}
		if strings.HasPrefix(strings.TrimSpace(message), "UNAUTHORISED") == true {
			fmt.Print(message)
		} else if strings.HasPrefix(strings.TrimSpace(message), "TIMEOUT") == true {
			fmt.Print(message)
			break
		} else {
			if strings.HasPrefix(strings.TrimSpace(message), "AUTHORISED") == false {
				fmt.Print(message)
			}
		}
	}
}
