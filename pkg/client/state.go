package client

import "wolf/pkg/server"

type ScreenType int

const (
	ScreenLobby ScreenType = iota
	ScreenRoom
	ScreenGame
	ScreenRoleChoice
)

type State struct {
	Screen   ScreenType
	RoomID   string
	MyName   string
	Room     *server.Room
	Role1    string
	Role2    string
	GameInfo string
}

var GlobalState = &State{Screen: ScreenLobby}
