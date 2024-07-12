package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	DEFAULT_PORT            = "8080"
	DEFAULT_MAX_CONNECTIONS = 5
)

type ChatServer struct {
	clients       map[net.Conn]string
	clientBuffer  chan net.Conn
	broadcastChan chan string
	errChan       chan error
}

func NewServer(port string, maxConns int) ChatServer {
	return ChatServer{
		clients:       make(map[net.Conn]string, maxConns),
		clientBuffer:  make(chan net.Conn, maxConns),
		broadcastChan: make(chan string),
		errChan:       make(chan error),
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", DEFAULT_PORT, "Port on which to run the server, defaults to 8080")
	var maxConnections int
	flag.IntVar(&maxConnections, "max-connections", DEFAULT_MAX_CONNECTIONS, "Number of maximum connections to accept.")
	flag.Parse()
	s := NewServer(port, maxConnections)
	s.StartServer(port, maxConnections)
}

func (c *ChatServer) listen(conn net.Conn) {
	defer func() {
		// on exit, close the connection, delete the connection from the map, and clear the channel
		conn.Close()
		delete(c.clients, conn)
		<-c.clientBuffer
	}()
	sender := c.clients[conn]
	for {
		reader := bufio.NewReader(conn)
		line, _, err := reader.ReadLine()
		if err != nil {
			fmt.Print(err)
			break
		}
		message := string(line)
		fmt.Println(message)
		c.broadcastChan <- fmt.Sprintf("[%s] [%s] - %s\n", sender, time.Now().Format(time.RFC1123), message)
	}
}

func (c *ChatServer) broadcastMessage() {
	for message := range c.broadcastChan {
		fmt.Print(message)
		for client := range c.clients {
			_, err := fmt.Fprint(client, message)
			if err != nil {
				fmt.Printf("failed to write: %v", err)
			}
		}
	}
}

func (c *ChatServer) establishConnection(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	line, _, err := reader.ReadLine()
	if err != nil {
		return fmt.Errorf("failed to read initial message: %v", err)
	}
	headerPieces := strings.Split(string(line), "=")
	headers := map[string]string{}
	currentHeader := ""
	for i, value := range headerPieces {
		if i%2 == 0 {
			currentHeader = value
		} else {
			headers[currentHeader] = value
		}
	}
	user := headers["User"]
	c.clients[conn] = user
	go c.listen(conn)
	return nil
}

func (c *ChatServer) StartServer(port string, maxConnections int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		fmt.Printf("failed to start server on port: %s", port)
		fmt.Printf("%v", err)
		return
	}
	defer listener.Close()
	go c.broadcastMessage()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed accepting connection: %v", err)
			return
		}
		select {
		case c.clientBuffer <- conn:
			if err := c.establishConnection(conn); err != nil {
				fmt.Print(err)
				<-c.clientBuffer
			}
		case err := <-c.errChan:
			fmt.Print(err)
		default:
			fmt.Print("Connection rejected")
			conn.Close()
		}
	}
}
