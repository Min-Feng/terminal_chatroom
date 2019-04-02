package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client struct {
	msgChan chan string
	conn    net.Conn
}

// 特別處,將每個 connetion 與 channel 綁在一起,視為同一個
// 由於chan型別,又傳遞一個chan
// 因此字串通過不同的channel 轉發到特定 connection
var enter chan client
var leave chan client
var broadcast chan string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	enter = make(chan client)
	leave = make(chan client)
	broadcast = make(chan string)
}

func main() {
	l, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		log.Println(err)
	}

	go clientManager()
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

func clientManager() {
	// 利用channel 將map型別維持在單一 goroutine
	// 藉此達成concurrency safe
	// 而不需要考慮加鎖開鎖
	clients := make(map[client]bool)

	var msg string

	for {
		select {
		case cli := <-enter:
			msg = fmt.Sprintf("歡迎來%s到聊天室", cli.conn.RemoteAddr().String())
			cli.msgChan <- msg
			clients[cli] = true

			// 原本打算把所有訊息邏輯集中於此
			// 但發現 broadcast 變成發送跟接收都在同一個 routine
			// 於是改變作法
			// msg := fmt.Sprintf("%s進入了聊天室", conn.RemoteAddr().String())
			// broadcast <- msg

		case cli := <-leave:
			delete(clients, cli)
			close(cli.msgChan)

		case msg := <-broadcast:
			for cli := range clients {
				cli.msgChan <- msg
			}
		}
	}
}

func sendMessage(cli client) {
	// 如同每一個channel 綁定一個 connection
	// 往channel 送訊息 ,如同對特定的connection 執行動作
	for msg := range cli.msgChan {
		fmt.Fprintf(cli.conn, "\t%s\n", msg)
	}
}

func receiveMessage(cli client) {
	input := bufio.NewScanner(cli.conn)
	for input.Scan() {
		msg := fmt.Sprintf("%s說:%s", cli.conn.RemoteAddr().String(), input.Text())
		broadcast <- msg
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Println("welcome", conn.RemoteAddr().String())

	// 可以嘗試send receive 都各自開一個channel來維護
	cli := client{make(chan string), conn}
	go sendMessage(cli)

	msg := fmt.Sprintf("%s進入了聊天室", conn.RemoteAddr().String())
	broadcast <- msg

	enter <- cli
	receiveMessage(cli)
	leave <- cli

	msg = fmt.Sprintf("%s離開了聊天室", conn.RemoteAddr().String())
	broadcast <- msg

	log.Println("good bye", conn.RemoteAddr().String())
}
