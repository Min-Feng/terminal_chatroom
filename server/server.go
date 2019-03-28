package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

var enter chan net.Conn
var leave chan net.Conn
var message chan string

type msg

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	enter = make(chan net.Conn)
	leave = make(chan net.Conn)
	message = make(chan string)
}

func main() {
	l, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		log.Println(err)
	}

	go connectionManager()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
		}

		defer func() {
			log.Println("connection close")
			conn.Close()
		}()

		go handleConn(conn)
	}
}

func connectionManager() {
	clients := make(map[net.Conn]bool)
	for {
		select {
		case conn := <-enter:
			clients[conn] = true
			log.Println("welcome", conn.RemoteAddr())
			fmt.Fprintln(conn, "歡迎來到聊天室", conn.RemoteAddr())

			for c := range clients {
				if c != conn {
					go fmt.Fprintln(c, c.RemoteAddr(), "進入了聊天室")
				}
			}
		case conn := <-leave:
			delete(clients, conn)
			log.Println("good bye", conn.RemoteAddr())

			for c := range clients {
				go fmt.Fprintln(c, c.RemoteAddr(), "離開了聊天室")
			}
		case msg := <-message:
			for conn := range clients {
				go sendMessage(conn, msg)
			}
		}
	}
}

func sendMessage(conn net.Conn, msg string) {
	fmt.Fprintf(conn, "\t %s 說:%s\n", conn.RemoteAddr(), msg)
}

func handleConn(c net.Conn) {
	enter <- c
	input := bufio.NewScanner(c)
	for input.Scan() {
		message <- input.Text()
	}
	leave <- c
}
