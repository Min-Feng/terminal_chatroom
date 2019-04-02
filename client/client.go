package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		log.Println("connection close")
		conn.Close()
	}()

	go receiveMessage(conn)
	sendMessage(conn)

}

func receiveMessage(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func sendMessage(conn net.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// Scanner 會去掉\n 若沒有補上去,接收方會不知道斷點在那
		conn.Write(append(scanner.Bytes(), '\n'))
	}
}
