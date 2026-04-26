package server

import (
	"github.com/gorilla/websocket"
	"wolf/pkg"
)

type Handler struct {
}

func (handler *Handler) CreateRoom(Name string) {
	roomID := Name
	GlobalRoomManager.CreateRoom(roomID, GlobalConns.GetPlayer(Name))
	conn := GlobalConns.GetConn(Name)
	//msg:=SCMessage{
	//	Type:    MsgRoomInfo,
	//}
	conn.WriteMessage(websocket.TextMessage, []byte(roomID))
}

func (handler *Handler) JoinRoom(Name string, roomID string) {
	room := GlobalRoomManager.GetRoom(roomID)
	if room != nil {
		room.Players[len(room.Players)-1] = GlobalConns.GetPlayer(Name)
	}
}

func (handler *Handler) LeaveRoom(Name string, roomID string) {
	room := GlobalRoomManager.GetRoom(roomID)
	delete(room.Players, GlobalConns.GetPlayer(Name).Seat)
}

func (handler *Handler) CreatePlayer(Name string, conn *websocket.Conn) *pkg.Player {
	GlobalConns.Add(Name, conn)
	return &pkg.Player{Name: Name}
}

func (handler *Handler) Ready(Name string) {
	player := GlobalConns.GetPlayer(Name)
	player.Ready = true
}

func (handler *Handler) UnReady(Name string) {
	player := GlobalConns.GetPlayer(Name)
	player.Ready = false
}

func (handler *Handler) SetFirst(Name string, First string) {
	player := GlobalConns.GetPlayer(Name)
	if player.Role1.Name != First {
		player.Role1, player.Role2 = player.Role2, player.Role1
	}
}

func (handler *Handler) Leave(Name string) {
	GlobalConns.RemovePlayerAndConn(Name)
}
