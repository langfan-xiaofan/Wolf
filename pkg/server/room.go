package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"wolf/pkg"
)

type Room struct {
	ID      string           `json:"id"`
	Owner   string           `json:"owner"`
	Players []*pkg.Player    `json:"players"`
	Status  int              `json:"status"`
	Game    *Game            `json:"game"`
	Addr    map[string]string `json:"addr"`
}

type RoomManager struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

var GlobalRoomManager = NewRoomManager()

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) CreateRoom(roomID string, owner *pkg.Player) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	room := &Room{
		ID:      roomID,
		Owner:   owner.Name,
		Players: []*pkg.Player{owner},
		Status:  int(RoomWaiting),
	}
	owner.Seat = 0
	rm.rooms[roomID] = room
	return room
}

func (rm *RoomManager) GetRoom(roomID string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	var room *Room
	var ok bool
	if room, ok = rm.rooms[roomID]; !ok {
		fmt.Println("获取不到房间" + roomID)
	}
	return room
}

func (rm *RoomManager) JoinRoom(roomID string, player *pkg.Player) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	room, ok := rm.rooms[roomID]
	if !ok {
		return fmt.Errorf("房间%s不存在", roomID)
	}
	player.Seat = len(room.Players)
	room.Players = append(room.Players, player)
	fmt.Println("房间信息", room)
	return nil
}

func (rm *RoomManager) BroadcastRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	room, ok := rm.rooms[roomID]
	if !ok {
		return
	}
	jsonRoom, _ := json.Marshal(room)
	resp := &SCMessage{
		Type:    MsgRoomInfo,
		RoomID:  roomID,
		Content: string(jsonRoom),
	}
	jsonMsg, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	for _, player := range room.Players {
		conn := GlobalConns.GetConn(player.Name)
		if conn == nil {
			continue
		}
		conn.WriteMessage(websocket.TextMessage, jsonMsg)
		fmt.Printf("用户的昵称%s\n", player.Name)
	}
}

func (rm *RoomManager) GetPlayer(playerName string, roomID string) *pkg.Player {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	room, ok := rm.rooms[roomID]
	if !ok {
		return nil
	}
	for _, player := range room.Players {
		if player.Name == playerName {
			return player
		}
	}
	return nil
}

func (rm *RoomManager) GetPlayerBySeat(roomID string, seat int) *pkg.Player {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	room, ok := rm.rooms[roomID]
	if !ok {
		return nil
	}
	for _, player := range room.Players {
		if player.Seat == seat {
			return player
		}
	}
	return nil
}
