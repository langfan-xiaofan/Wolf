package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer conn.Close()
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if string(msg) == "开始倒计时五秒" {
			for i := 5; i > 0; i-- {
				conn.WriteMessage(msgType, []byte(fmt.Sprintf(">倒计时:%d", i)))
				// w.Write([]byte(fmt.Sprintf("%d\r", i)))
				fmt.Printf("%d\r", i)
				time.Sleep(time.Second)
			}
			conn.WriteMessage(msgType, []byte("\n>倒计时五秒结束"))
			conn.WriteMessage(msgType, []byte("\n>"))
		} else {
			conn.WriteMessage(msgType, []byte(">你好啊\n"))
			conn.WriteMessage(msgType, []byte(">"))
		}
		// log.Printf("Received message: %s\n", string(msg))
		// if err := conn.WriteMessage(msgType, msg); err != nil {
		// 	http.Error(w, err.Error(), http.StatusBadRequest)
		// 	return
		// }
	}
}

func main() {
	//http.HandleFunc("/", handle)
	//http.ListenAndServe(":8080", nil)
	for _ = range 10000000 {
		roles := []string{"狼人", "狼王", "潜行狼", "村民", "企鹅", "熊", "猎人", "狐狸", "乌鸦", "蝙蝠", "复制人", "村民"}
		// Fisher-Yates shuffle for global uniqueness across all player slots
		for i := len(roles) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			roles[i], roles[j] = roles[j], roles[i]
		}
		//fmt.Println(roles)
		roles_map := make(map[int][]string)
		idx := 0
		for i := range 6 {
			roles_map[i] = []string{roles[idx], roles[idx+1]}
			idx += 2
		}
		//fmt.Println(roles_map)
		check := make(map[string]int)
		for _, v := range roles_map {
			check[v[0]]++
			check[v[1]]++
			if v[0] == "村民" {
				if check[v[0]] > 2 {
					fmt.Println("error!:")
					fmt.Println(roles_map)
				}
			} else {
				if check[v[0]] > 1 {
					fmt.Println("error!:")
					fmt.Println(roles_map)
				}
			}
			if v[1] == "村民" {
				if check[v[1]] > 2 {
					fmt.Println("error!:")
					fmt.Println(roles_map)
				}
			} else {
				if check[v[1]] > 1 {
					fmt.Println("error!:")
					fmt.Println(roles_map)
				}
			}
		}

	}
}
