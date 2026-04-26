package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"wolf/pkg"
	"wolf/pkg/server"
)

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// 创建玩家
	player := new(pkg.Player)
	fmt.Println("请输入昵称")
	_, err = fmt.Scanln(&player.Name)
	if err != nil {
		return
	}
	msg := server.CSMessage{
		Type:    server.MsgCreatePlayer,
		Player:  player,
		Content: player.Name,
	}
	jsonMsg, _ := json.Marshal(msg)
	err = conn.WriteMessage(websocket.TextMessage, jsonMsg)
	if err != nil {
		return
	}

	roomCh := make(chan *server.Room, 100)
	gameStartCh := make(chan server.SCMessage, 1)
	// 启动接收消息的goroutine
	go receiveLoop(conn, roomCh, gameStartCh)
	// 显示 lobby 菜单
	showLobbyMenu(conn, player)
	// 等待房间信息
	gameStartMsg := waitingRoomLoop(conn, player, roomCh, gameStartCh)
	if gameStartMsg == nil {
		return
	}

	handleRoleChoice(conn, player, gameStartMsg)

	ShowGameMenu(conn, player)
	select {}
}

func receiveLoop(conn *websocket.Conn, roomCh chan<- *server.Room, gameStartCh chan<- server.SCMessage) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("连接断开: %v\n", err)
			return
		}

		var msg server.SCMessage
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("消息解析失败:", err)
			continue
		}

		switch msg.Type {
		case server.MsgRoomInfo:
			var room server.Room
			err = json.Unmarshal([]byte(msg.Content), &room)
			if err != nil {
				log.Println("房间信息解析失败:", err)
				continue
			}
			roomCh <- &room
		case server.MsgStartGame:
			gameStartCh <- msg
		case server.MsgSetFirst:
			fmt.Println(msg.Content)
		case server.MsgChat:
			fmt.Println(string(msg.Content))
		default:
			fmt.Println(string(msg.Content))
		}
	}
}

func showLobbyMenu(conn *websocket.Conn, player *pkg.Player) {
	fmt.Println("\033[2J\033[H")
	fmt.Println("欢迎来到狼人杀")
	fmt.Println("=========操作菜单如下=========")
	fmt.Println("1.创建房间")
	fmt.Println("2.加入房间")
	fmt.Println("3.退出游戏")
	fmt.Println(">请输入序号")
	var op int
	var msg server.CSMessage
	_, err := fmt.Scanln(&op)
	if err != nil {
		return
	}
	switch op {
	case 1:
		msg.Type = server.MsgCreatRoom
		msg.RoomID = player.Name
		msg.Player = player
		fmt.Println("创建房间的消息", msg)
	case 2:
		fmt.Println("请输入房间id")
		var roomID string
		_, err := fmt.Scanln(&roomID)
		if err != nil {
			return
		}
		msg.RoomID = roomID
		msg.Type = server.MsgJoinRoom
		msg.Player = player
		msg.Content = fmt.Sprintf("玩家%s加入房间%s", player.Name, roomID)
	case 3:
		return
	}
	jsonStr, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, jsonStr)
}

func waitingRoomLoop(conn *websocket.Conn, player *pkg.Player, roomCh <-chan *server.Room, gameStartCh <-chan server.SCMessage) *server.SCMessage {
	myName := player.Name
	var currentRoom *server.Room

	select {
	case currentRoom = <-roomCh:
	case startMsg := <-gameStartCh:
		setRolesFromContent(player, startMsg.Content)
		return &startMsg
	}

	if currentRoom == nil {
		return nil
	}

	doneCh := make(chan struct{})
	inputCh := make(chan int)
	go func() {
		for {
			select {
			case <-doneCh:
				return
			default:
				var idx int
				_, err := fmt.Scanln(&idx)
				if err != nil {
					return
				}
				select {
				case inputCh <- idx:
				case <-doneCh:
					return
				}
			}
		}
	}()
	defer close(doneCh)

	for {
		displayRoom(currentRoom, player)

		select {
		case room := <-roomCh:
			currentRoom = room
		case startMsg := <-gameStartCh:
			if setRolesFromContent(player, startMsg.Content) {
				fmt.Println("游戏已开始！")
				return &startMsg
			}
			fmt.Println(startMsg.Content)
		case input := <-inputCh:
			switch input {
			case 1:
				mySelf, ok := currentRoom.Players[player.Seat-1]
				if !ok {
					continue
				}
				mySelf.Ready = !mySelf.Ready
				msgType := server.MsgReady
				if !mySelf.Ready {
					msgType = server.MsgUnReady
				}
				msg := server.CSMessage{
					Type:    msgType,
					RoomID:  currentRoom.ID,
					Player:  currentRoom.Players[mySelf.Seat],
					Content: fmt.Sprintf("%d", mySelf.Seat),
				}
				jsonMsg, _ := json.Marshal(msg)
				conn.WriteMessage(websocket.TextMessage, jsonMsg)
			case 2:
				if currentRoom.Owner == myName {
					msg := server.CSMessage{
						Type:    server.MsgStartGame,
						RoomID:  currentRoom.ID,
						Player:  currentRoom.Players[player.Seat],
						Content: fmt.Sprintf("%d", player.Seat),
					}
					jsonMsg, _ := json.Marshal(msg)
					conn.WriteMessage(websocket.TextMessage, jsonMsg)
				}
			case 3:
				fmt.Println("退出房间")
				return nil
			}
		}
	}
}

func displayRoom(room *server.Room, player *pkg.Player) {
	fmt.Println("\033[2J\033[H")
	fmt.Println("=== 狼人杀 - 等待房间 ===")
	fmt.Printf("房间ID: %s\n", room.ID)
	fmt.Println()
	fmt.Println("[玩家列表]")
	for idx, player := range room.Players {
		readyMark := "❌"
		if player.Ready {
			readyMark = "✅"
		}
		fmt.Printf("玩家%d:%s %s\n", idx+1, player.Name, readyMark)
	}

	mySelf, ok := room.Players[player.Seat]
	if !ok {
		return
	}
	if !mySelf.Ready {
		fmt.Print("1.准备  ")
	} else {
		fmt.Print("1.取消准备  ")
	}
	if room.Owner == player.Name {
		fmt.Print("2.开始游戏  ")
	}
	fmt.Println("3.退出房间")
	fmt.Print("请输入选项>")
}

func setRolesFromContent(player *pkg.Player, content string) bool {
	parts := strings.Split(content, "身份1")
	if len(parts) < 2 {
		return false
	}
	rest := parts[1]
	restParts := strings.Split(rest, "身份2")
	if len(restParts) < 2 {
		return false
	}
	role1Name := strings.TrimRight(restParts[0], ",， ")
	role2Name := strings.TrimSpace(restParts[1])
	player.Role1 = pkg.NewRole(role1Name)
	player.Role2 = pkg.NewRole(role2Name)
	return true
}

func handleRoleChoice(conn *websocket.Conn, player *pkg.Player, gameStartMsg *server.SCMessage) {
	if player.Role1 == nil || player.Role2 == nil {
		return
	}
	fmt.Printf("请选择你想先使用的身份 (1.%s, 2.%s): ", player.Role1.Name, player.Role2.Name)
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return
	}

	response := server.CSMessage{
		Type:    server.MsgSetFirst,
		RoomID:  gameStartMsg.RoomID,
		Content: input,
		Player:  player,
	}
	jsonMsg, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, jsonMsg)

	// Sync local player with server: swap if user chose the second role
	if player.Role2.Name == input {
		player.Role1, player.Role2 = player.Role2, player.Role1
	}
}

func ShowGameMenu(conn *websocket.Conn, player *pkg.Player) {
	fmt.Println("\033[2J\033[H")
	fmt.Println("=== 游戏进行中 ===")
	if player.Role1 == nil {
		fmt.Println("角色信息获取失败")
		return
	}
	fmt.Printf("你的身份:1.%s,2.%s\n", player.Role1.Name, player.Role2.Name)
}
