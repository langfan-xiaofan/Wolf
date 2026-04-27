package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"wolf/pkg/server"
)

func RenderLobby() {
	fmt.Println("\u001B[H\u001B[2J")
	fmt.Println("欢迎来到狼人杀")
	fmt.Println("=========操作菜单如下=========")
	fmt.Println("1.创建房间")
	fmt.Println("2.加入房间")
	fmt.Println("3.退出游戏")
	fmt.Print("请输入选项>")
}

func RenderRoom(content string) {
	fmt.Println("\u001B[H\u001B[2J")

	var room server.Room
	err := json.Unmarshal([]byte(content), &room)
	if err != nil {
		fmt.Println("房间信息解析失败:", err)
		return
	}

	fmt.Println("=== 狼人杀 - 等待房间 ===")
	fmt.Printf("房间ID: %s\n", room.ID)
	fmt.Println()
	fmt.Println("[玩家列表]")
	for i := range len(room.Players) {
		readyMark := "❌"
		if room.Players[i].Ready {
			readyMark = "✅"
		}
		fmt.Printf("玩家%d:%s %s\n", i+1, room.Players[i].Name, readyMark)
	}
	fmt.Println()
	isReady := false
	isOwner := room.Owner == GlobalState.MyName
	for _, p := range room.Players {
		if p.Name == GlobalState.MyName {
			isReady = p.Ready
			break
		}
	}

	readyText := "准备"
	if isReady {
		readyText = "取消准备"
	}

	if isOwner {
		fmt.Printf("1.%s  2.开始游戏  3.退出房间\n", readyText)
	} else {
		fmt.Printf("1.%s  2.退出房间\n", readyText)
	}
	fmt.Print("请输入选项>")
}

func RenderGameStart(content string) {
	fmt.Println("\u001B[H\u001B[2J")
	fmt.Println("=== 游戏开始 ===")
	fmt.Println(content)
}

func RenderRoleChoice(role1, role2 string) {
	fmt.Println("\u001B[H\u001B[2J")
	fmt.Println("=== 选择初始身份 ===")
	fmt.Printf("你的身份是: %s 和 %s\n", role1, role2)
	fmt.Printf("请选择你想先使用的身份 (输入 %s 或 %s): ", role1, role2)
}

func RenderChat(content string) {
	fmt.Println(content)
}

func RenderSetFirst(content string) {
	fmt.Println("\033[H\033[2J")
	fmt.Println("=== 设置初始身份 ===")
	fmt.Println(content)
	fmt.Print("请输入选项>")
}

func RenderMessage(content string) {
	fmt.Println(content)
}

func ParseRoomFromContent(content string) (*server.Room, error) {
	var room server.Room
	err := json.Unmarshal([]byte(content), &room)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func ParseRolesFromContent(content string) (role1, role2 string, ok bool) {
	parts := strings.Split(content, "身份1")
	if len(parts) < 2 {
		return "", "", false
	}
	rest := parts[1]
	restParts := strings.Split(rest, "身份2")
	if len(restParts) < 2 {
		return "", "", false
	}
	role1 = strings.TrimRight(restParts[0], ",， ")
	role2 = strings.TrimSpace(restParts[1])
	return role1, role2, true
}

func RenderGame(content string) {
	fmt.Println("\033[2J\033[H]")
	fmt.Println("=== 游戏进行中 ===")
	fmt.Println(content)
	// 持续显示自己的身份牌
	if len(GlobalState.Role1) > 0 && len(GlobalState.Role2) > 0 {
		fmt.Printf("\n你的身份牌: [%s] [%s]\n", GlobalState.Role1, GlobalState.Role2)
	}
	fmt.Println("\n请输入操作>")
}
