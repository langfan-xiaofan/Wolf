package server

import (
	"errors"
	"math/rand"
	"wolf/pkg"
)

type Server struct {
	Handler       *Handler
	PlayerManager *PlayerManager
	RoomManager   *RoomManager
}

type MsgType string

const (
	MsgCreatRoom    MsgType = "create_room"
	MsgJoinRoom     MsgType = "join_room"
	MsgLeaveRoom    MsgType = "leave_room"
	MsgStartGame    MsgType = "start_game"
	MsgUseSkill     MsgType = "use_skill"
	MsgNightAction  MsgType = "night_action"
	MsgChat         MsgType = "chat"
	MsgCreatePlayer MsgType = "create_player"
	MsgReady        MsgType = "ready"
	MsgUnReady      MsgType = "unready"
	MsgLeaveGame    MsgType = "leave_game"
	MsgRoomInfo     MsgType = "room_info"
	MsgSetFirst     MsgType = "set_first"
)

type CSMessage struct {
	Type    MsgType     `json:"type"`
	RoomID  string      `json:"room_id"`
	Content string      `json:"content"`
	Player  *pkg.Player `json:"player"`
}

type SCMessage struct {
	Type    MsgType     `json:"type"`
	RoomID  string      `json:"room_id"`
	Content string      `json:"content"`
	Player  *pkg.Player `json:"player"`
}

type RoomStatus int

const (
	RoomWaiting RoomStatus = iota
	RoomPlaying
	RoomEnd
)

func (r *Room) StartGame() error {
	if r.Status != int(RoomWaiting) {
		return errors.New("游戏已经开始或已结束")
	}
	if len(r.Players) < 6 {
		return errors.New("房间人数不足，最少需要6人")
	}
	if !r.AllPlayersReady() {
		return errors.New("还有玩家未准备")
	}
	roles := []string{"狼人", "狼王", "潜行狼", "村民", "企鹅", "熊", "猎人", "狐狸", "乌鸦", "蝙蝠", "复制人", "村民"}
	// Fisher-Yates shuffle for global uniqueness across all player slots
	for i := len(roles) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		roles[i], roles[j] = roles[j], roles[i]
	}
	roles_map := make(map[string][]string)
	idx := 0
	for _, player := range r.Players {
		roles_map[player.Name] = []string{roles[idx], roles[idx+1]}
		idx += 2
	}
	r.Game = &Game{
		RoomID:  r.ID,
		Phase:   PhaseNight,
		Players: r.GetAllPlayers(),
	}
	for _, player := range r.Players {
		r.Game.SetFirst(player, pkg.NewRole(roles_map[player.Name][0]), pkg.NewRole(roles_map[player.Name][1]))
	}
	return nil
}

func (r *Room) GetAllPlayers() []*pkg.Player {
	var players []*pkg.Player
	for _, player := range r.Players {
		players = append(players, player)
	}
	return players
}

func (r *Room) AllPlayersReady() bool {
	for _, p := range r.Players {
		if !p.Ready {
			return false
		}
	}
	return true
}
