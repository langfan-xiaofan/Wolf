package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

type GameClient struct {
	Conn *websocket.Conn
	State *ClientState
	Input chan string
	Message chan string
}

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	conn.WriteMessage(websocket.TextMessage, []byte("开始倒计时五秒"))
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Fatal(err)
			}
			// log.Printf("Received message: %s\r", string(msg))
			fmt.Printf("\r%s", string(msg))
		}
	}()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "exit" {
			break
		}
		conn.WriteMessage(websocket.TextMessage, []byte(line))
	}
}
