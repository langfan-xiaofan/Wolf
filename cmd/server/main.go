package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"wolf/pkg/server"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	msg := new(server.CSMessage)
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("玩家断开连接:%v\n", err)
			return
		}
		err = json.Unmarshal(message, msg)
		if err != nil {
			log.Printf("消息解析失败: %v\n", err)
			continue
		}
		switch msg.Type {
		case server.MsgCreatRoom:
			roomID := msg.RoomID
			msg.Player.Seat = 0
			server.GlobalRoomManager.CreateRoom(roomID, msg.Player)
			log.Printf("玩家%s创建了房间%s", msg.Player.Name, roomID)
			room := server.GlobalRoomManager.GetRoom(roomID)
			jsonRoom, _ := json.Marshal(room)
			resp := &server.SCMessage{
				Type:    server.MsgRoomInfo,
				RoomID:  roomID,
				Content: string(jsonRoom),
			}
			jsonMsg, _ := json.Marshal(resp)
			err = conn.WriteMessage(websocket.TextMessage, jsonMsg)
		case server.MsgJoinRoom:
			log.Println("接收到的加入房间的消息", msg)
			room := server.GlobalRoomManager.GetRoom(msg.RoomID)
			if room == nil {
				resp := &server.SCMessage{Type: server.MsgJoinRoom, Content: "房间不存在"}
				jsonMsg, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, jsonMsg)
				break
			}
			log.Println("玩家信息", msg.Player)
			msg.Player.Seat = len(room.Players)
			err := server.GlobalRoomManager.JoinRoom(msg.RoomID, msg.Player)
			if err != nil {
				resp := &server.SCMessage{Type: server.MsgJoinRoom, Content: err.Error()}
				jsonMsg, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, jsonMsg)
				break
			}
			server.GlobalRoomManager.BroadcastRoom(room.ID)
		case server.MsgCreatePlayer:
			server.GlobalConns.Add(msg.Player.Name, conn)
			log.Printf("玩家 %s 连接成功", msg.Player.Name)
			resp := &server.SCMessage{Type: server.MsgCreatePlayer, Content: fmt.Sprintf("用户%s加入成功", msg.Player.Name)}
			server.GlobalConns.SendToPlayer(msg.Player.Name, resp)
		case server.MsgStartGame:
			room := server.GlobalRoomManager.GetRoom(msg.RoomID)
			if room == nil {
				break
			}
			if room.Owner != msg.Player.Name {
				resp := &server.SCMessage{Type: server.MsgChat, Content: "只有房主才可以开始游戏"}
				jsonMsg, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, jsonMsg)
				break
			}
			err = room.StartGame()
			if err != nil {
				log.Println(err)
				resp := &server.SCMessage{Type: server.MsgChat, Content: err.Error()}
				jsonMsg, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, jsonMsg)
				break
			}
			for _, v := range room.Players {
				resp := &server.SCMessage{
					Type:    server.MsgStartGame,
					RoomID:  msg.RoomID,
					Content: fmt.Sprintf("游戏开始！你的身份是:身份1%s,身份2%s", v.Role1.Name, v.Role2.Name),
				}
				err = server.GlobalConns.SendToPlayer(v.Name, resp)
				if err != nil {
					log.Printf("发送角色信息给 %s 失败: %v\n", v.Name, err)
				}
			}
		case server.MsgReady:
			room := server.GlobalRoomManager.GetRoom(msg.RoomID)
			if room == nil {
				break
			}
			player := server.GlobalRoomManager.GetPlayer(msg.Player.Name, room.ID)
			if player == nil {
				log.Printf("玩家 %s 不在房间 %s 中\n", msg.Player.Name, room.ID)
				break
			}
			player.SetReady(true)
			log.Printf("玩家 %s 已准备\n", msg.Player.Name)
			server.GlobalRoomManager.BroadcastRoom(room.ID)
		case server.MsgUnReady:
			room := server.GlobalRoomManager.GetRoom(msg.RoomID)
			if room == nil {
				break
			}
			player := server.GlobalRoomManager.GetPlayer(msg.Player.Name, room.ID)
			if player == nil {
				log.Printf("玩家 %s 不在房间 %s 中\n", msg.Player.Name, room.ID)
				break
			}
			player.SetReady(false)
			log.Printf("玩家 %s 取消准备\n", msg.Player.Name)
			server.GlobalRoomManager.BroadcastRoom(room.ID)
		case server.MsgRoomInfo:
			room := server.GlobalRoomManager.GetRoom(msg.RoomID)
			jsonRoom, _ := json.Marshal(room)
			resp := &server.SCMessage{
				Type:    server.MsgRoomInfo,
				RoomID:  msg.RoomID,
				Content: string(jsonRoom),
			}
			jsonMsg, _ := json.Marshal(resp)
			err = conn.WriteMessage(websocket.TextMessage, jsonMsg)
		case server.MsgSetFirst:
			first := msg.Content
			room := server.GlobalRoomManager.GetRoom(msg.RoomID)
			if msg.Player.Role1.Name == first {

			} else {
				msg.Player.Role1, msg.Player.Role2 = msg.Player.Role2, msg.Player.Role1
			}
			room.Players[msg.Player.Seat] = msg.Player
		}
	}
}

func main() {
	http.HandleFunc("/", wsHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
