package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"wolf/pkg"
)

type Handler struct {
}

var GlobalHandler = NewHandler()

func NewHandler() *Handler {
	return &Handler{}
}

type HandlerFunc func(ctx *context.Context)

func (handler *Handler) CreateRoom(playerName string) {
	roomID := playerName
	GlobalRoomManager.CreateRoom(roomID, GlobalConns.GetPlayer(playerName))
	conn := GlobalConns.GetConn(playerName)
	room := GlobalRoomManager.GetRoom(roomID)
	if room == nil {
		return
	}
	jsonRoom, _ := json.Marshal(room)
	resp := &SCMessage{
		Type:    MsgRoomInfo,
		Content: string(jsonRoom),
	}
	jsonResp, _ := json.Marshal(resp)
	//fmt.Println(jsonResp)
	conn.WriteMessage(websocket.TextMessage, jsonResp)
}

func (handler *Handler) JoinRoom(playerName string, roomID string) {
	player := GlobalConns.GetPlayer(playerName)
	if player == nil {
		return
	}
	err := GlobalRoomManager.JoinRoom(roomID, player)
	if err != nil {
		return
	}
	GlobalRoomManager.BroadcastRoom(roomID)
}

func (handler *Handler) LeaveRoom(playerName string, roomID string) {
	room := GlobalRoomManager.GetRoom(roomID)
	if room == nil {
		return
	}
	player := GlobalConns.GetPlayer(playerName)
	if player == nil {
		return
	}
	seat := player.Seat
	newPlayers := make([]*pkg.Player, 0, len(room.Players))
	for _, p := range room.Players {
		if p.Seat != seat {
			newPlayers = append(newPlayers, p)
		}
	}
	room.Players = newPlayers
	for i := range room.Players {
		room.Players[i].Seat = i
	}
	GlobalRoomManager.BroadcastRoom(roomID)
}

func (handler *Handler) CreatePlayer(playerName string, conn *websocket.Conn) *pkg.Player {
	GlobalConns.Add(playerName, conn)
	resp := &SCMessage{Type: MsgCreatePlayer, Content: playerName}
	jsonResp, _ := json.Marshal(resp)
	conn.WriteMessage(websocket.TextMessage, jsonResp)
	return &pkg.Player{Name: playerName}
}

func (handler *Handler) Ready(playerName string, roomID string) {
	player := GlobalRoomManager.GetPlayer(playerName, roomID)
	if player == nil {
		return
	}
	player.Ready = true
	GlobalRoomManager.BroadcastRoom(roomID)
}

func (handler *Handler) UnReady(playerName string, roomID string) {
	player := GlobalRoomManager.GetPlayer(playerName, roomID)
	if player == nil {
		return
	}
	player.Ready = false
	GlobalRoomManager.BroadcastRoom(roomID)
}

func (handler *Handler) SetFirst(playerName string, roomID string, first string) {
	player := GlobalRoomManager.GetPlayer(playerName, roomID)
	if player == nil {
		return
	}
	if player.Role1.Name != first {
		player.Role1, player.Role2 = player.Role2, player.Role1
	}
	fmt.Println("SetFirst:role1", player.Role1.Name, "role2", player.Role2.Name)
}

func (handler *Handler) Leave(playerName string) {
	GlobalConns.RemovePlayerAndConn(playerName)
}

func (handler *Handler) StartGame(roomID string) {
	room := GlobalRoomManager.GetRoom(roomID)
	err := room.StartGame()
	if err != nil {
		return
	}
	// room.StartGame() 已经通过 SetFirst 发送了 MsgSetFirst
	// 玩家选择完身份顺序后，再由游戏逻辑发送 MsgStartGame
}

func UnmarshalContent(content string) (map[string]interface{}, error) {
	var contentMap map[string]interface{}
	err := json.Unmarshal([]byte(content), &contentMap)
	if err != nil {
		return nil, err
	}
	return contentMap, nil
}
