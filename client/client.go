package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

func main() {
	var user string
	flag.StringVar(&user, "user", "", "The username you wish to use")
	flag.Parse()
	if user == "" {
		fmt.Println("You must specify a username.")
		return
	}
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()
	if _, err := conn.Write([]byte(fmt.Sprintf("User=%s\n", user))); err != nil {
		fmt.Printf("Failed to establish connection: %v", err)
		return
	}

	go readMessages(conn)

	fmt.Println("Connected to chat server. Type your messages:")
	for {
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		fmt.Fprintf(conn, message+"\n")
	}
}

func readMessages(conn net.Conn) {
	for {
		message, err := bufio.NewReader(conn).ReadSlice('\n')
		if err != nil {
			fmt.Printf("Error: %v", err)
			conn.Close()
			return
		}
		fmt.Print("Message from server: " + string(message))
	}
}
