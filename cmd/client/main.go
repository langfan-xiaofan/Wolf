package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"wolf/pkg"
	"wolf/pkg/client"
	"wolf/pkg/server"
)

var conn *websocket.Conn

func main() {
	var err error
	conn, _, err = websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 创建玩家
	fmt.Print("请输入昵称>")
	name := readLine()
	player := &pkg.Player{Name: name}

	client.GlobalState.MyName = name

	// 发送创建玩家消息
	sendMsg(&server.CSMessage{
		Type:    server.MsgCreatePlayer,
		Player:  player,
		Content: name,
	})

	// 初始化大厅
	client.RenderLobby()

	// 事件通道：服务器消息都走这里
	eventCh := make(chan server.SCMessage, 100)

	// goroutine: 读取服务器消息
	go wsReceiveLoop(eventCh)

	// goroutine: 读取用户输入并直接发送到服务器
	go inputLoop()

	// 主循环: 只负责渲染服务器消息
	for msg := range eventCh {
		handleServerMessage(&msg)
	}
}

func wsReceiveLoop(eventCh chan<- server.SCMessage) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("连接断开:", err)
			close(eventCh)
			return
		}
		var msg server.SCMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			fmt.Println("消息解析失败:", err)
			fmt.Println("消息内容", message)
			continue
		}
		eventCh <- msg
	}
}

func inputLoop() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		handleUserInput(input)
	}
}

func handleUserInput(input string) {
	switch client.GlobalState.Screen {
	case client.ScreenLobby:
		handleLobbyInput(input)
	case client.ScreenRoom:
		handleRoomInput(input)
	case client.ScreenRoleChoice:
		handleRoleChoiceInput(input)
	case client.ScreenGame:
		handleGameInput(input)
	}
}

func handleLobbyInput(input string) {
	switch input {
	case "1":
		sendMsg(&server.CSMessage{
			Type:    server.MsgCreatRoom,
			RoomID:  client.GlobalState.MyName,
			Player:  &pkg.Player{Name: client.GlobalState.MyName},
			Content: client.GlobalState.MyName,
		})
	case "2":
		fmt.Print("请输入房间ID>")
		roomID := readLine()
		sendMsg(&server.CSMessage{
			Type:   server.MsgJoinRoom,
			RoomID: roomID,
			Player: &pkg.Player{Name: client.GlobalState.MyName},
			Content: MustJson(struct {
				PlayerName string `json:"playerName"`
				RoomID     string `json:"roomID"`
			}{PlayerName: client.GlobalState.MyName, RoomID: roomID}),
		})
	case "3":
		fmt.Println("退出游戏")
		os.Exit(0)
	}
}

func handleRoomInput(input string) {
	switch input {
	case "1":
		// 查找自己的准备状态
		isReady := false
		if client.GlobalState.Room != nil {
			for _, p := range client.GlobalState.Room.Players {
				if p.Name == client.GlobalState.MyName {
					isReady = p.Ready
					break
				}
			}
		}
		msgType := server.MsgReady
		if isReady {
			msgType = server.MsgUnReady
		}
		content := struct {
			PlayerName string `json:"playerName"`
			RoomID     string `json:"roomID"`
		}{PlayerName: client.GlobalState.MyName, RoomID: client.GlobalState.RoomID}
		jsonContent, _ := json.Marshal(content)
		sendMsg(&server.CSMessage{
			Type:    msgType,
			RoomID:  client.GlobalState.RoomID,
			Player:  &pkg.Player{Name: client.GlobalState.MyName},
			Content: string(jsonContent),
		})
	case "2":
		if client.GlobalState.Room != nil && client.GlobalState.Room.Owner == client.GlobalState.MyName {
			sendMsg(&server.CSMessage{
				Type:    server.MsgStartGame,
				RoomID:  client.GlobalState.RoomID,
				Player:  &pkg.Player{Name: client.GlobalState.MyName},
				Content: client.GlobalState.MyName,
			})
		} else {
			client.GlobalState.Screen = client.ScreenLobby
			client.RenderLobby()
		}
	case "3":
		client.GlobalState.Screen = client.ScreenLobby
		client.RenderLobby()
	}
}

func handleRoleChoiceInput(input string) {
	sendMsg(&server.CSMessage{
		Type:   server.MsgSetFirst,
		RoomID: client.GlobalState.RoomID,
		Player: &pkg.Player{Name: client.GlobalState.MyName},
		Content: MustJson(struct {
			PlayerName string `json:"playerName"`
			First      string `json:"first"`
		}{PlayerName: client.GlobalState.MyName, First: input}),
	})
}

func handleGameInput(input string) {
	sendMsg(&server.CSMessage{
		Type:    server.MsgNightAction,
		RoomID:  client.GlobalState.RoomID,
		Player:  &pkg.Player{Name: client.GlobalState.MyName},
		Content: input,
	})
}

func handleServerMessage(msg *server.SCMessage) {
	switch msg.Type {
	case server.MsgRoomInfo:
		room, err := client.ParseRoomFromContent(msg.Content)
		if err != nil {
			fmt.Println("房间信息解析失败:", err)
			fmt.Println("房间信息", room)
			return
		}
		client.GlobalState.Room = room
		client.GlobalState.RoomID = room.ID
		client.GlobalState.Screen = client.ScreenRoom
		client.RenderRoom(msg.Content)

	case server.MsgStartGame:
		client.GlobalState.GameInfo = msg.Content
		client.GlobalState.Screen = client.ScreenGame
		client.RenderGame(msg.Content)

	case server.MsgSetFirst:
		// 解析角色信息 "你的身份是:村民,乌鸦\n请选择..."
		if msg.Player != nil && msg.Player.Role1 != nil && msg.Player.Role2 != nil {
			if len(msg.Player.Role1.Name) > 0 && len(msg.Player.Role2.Name) > 0 {
				client.GlobalState.Role1 = msg.Player.Role1.Name
				client.GlobalState.Role2 = msg.Player.Role2.Name
			}
		}
		client.RenderSetFirst(msg.Content)
		client.GlobalState.Screen = client.ScreenRoleChoice

	case server.MsgChat:
		client.RenderChat(msg.Content)

	default:
		if msg.Content != "" {
			fmt.Println(msg.Content)
		}
	}
}

func sendMsg(msg *server.CSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("消息序列化失败:", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		fmt.Println("发送消息失败:", err)
	}
}

func readLine() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}
func MustJson(obj interface{}) string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}
