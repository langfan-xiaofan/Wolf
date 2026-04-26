package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"wolf/pkg"
)

type PlayerManager struct {
	mu      sync.Mutex
	players map[string]*pkg.Player
	conns   map[string]*websocket.Conn
}

var GlobalConns = NewPlayerManager()

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		players: make(map[string]*pkg.Player),
		conns:   make(map[string]*websocket.Conn),
	}
}

func (c *PlayerManager) Add(playerName string, conn *websocket.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.players[playerName] = &pkg.Player{Name: playerName}
	c.conns[playerName] = conn
}

func (c *PlayerManager) GetConn(playerName string) *websocket.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	if conn, ok := c.conns[playerName]; ok {
		return conn
	} else {
		return nil
	}
}

func (c *PlayerManager) GetPlayer(playerName string) *pkg.Player {
	c.mu.Lock()
	defer c.mu.Unlock()
	if player, ok := c.players[playerName]; ok {
		return player
	} else {
		return nil
	}
}

func (c *PlayerManager) RemovePlayerAndConn(playerName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.conns, playerName)
	delete(c.players, playerName)
}

func (c *PlayerManager) SendToPlayer(playerName string, msg *SCMessage) error {
	conn := c.GetConn(playerName)
	if conn == nil {
		return fmt.Errorf("玩家%s不在线", playerName)
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	err = conn.WriteMessage(websocket.TextMessage, data)
	return err
}
