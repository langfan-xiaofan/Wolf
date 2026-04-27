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
	room := GlobalRoomManager.GetRoom(roomID)
	if room == nil {
		return
	}
	player := GlobalRoomManager.GetPlayer(playerName, roomID)
	if player == nil {
		return
	}
	if player.Role1.Name != first {
		player.Role1, player.Role2 = player.Role2, player.Role1
	}
	player.Activity = player.Role1
	room.SetFirstReady[playerName] = true

	// 检查所有玩家是否都完成了选择
	allReady := true
	for _, p := range room.Players {
		if !room.SetFirstReady[p.Name] {
			allReady = false
			break
		}
	}

	// 所有玩家都选完了，发送 MsgStartGame
	if allReady {
		room.Game.Phase = PhaseNight
		GlobalEventBus.Publish(string(Night), room.Game)
		for _, p := range room.Players {
			room.Game.NotifyPlayer(p, SCMessage{
				Type:    MsgStartGame,
				Content: fmt.Sprintf("游戏开始！你的身份牌是:%s,%s", p.Role1.Name, p.Role2.Name),
			})
		}
	}
}

func (handler *Handler) Leave(playerName string) {
	GlobalConns.RemovePlayerAndConn(playerName)
}

func (handler *Handler) StartGame(roomID string) {
	room := GlobalRoomManager.GetRoom(roomID)
	if room == nil {
		handler.sendError(roomID, "房间不存在")
		return
	}
	err := room.StartGame()
	if err != nil {
		handler.sendError(roomID, err.Error())
		return
	}
}

func (handler *Handler) sendError(roomID string, msg string) {
	resp := &SCMessage{
		Type:    MsgChat,
		RoomID:  roomID,
		Content: msg,
	}
	jsonResp, _ := json.Marshal(resp)
	for _, player := range GlobalRoomManager.GetRoom(roomID).Players {
		conn := GlobalConns.GetConn(player.Name)
		if conn != nil {
			conn.WriteMessage(websocket.TextMessage, jsonResp)
		}
	}
}

func UnmarshalContent(content string) (map[string]interface{}, error) {
	var contentMap map[string]interface{}
	err := json.Unmarshal([]byte(content), &contentMap)
	if err != nil {
		return nil, err
	}
	return contentMap, nil
}
