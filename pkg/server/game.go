package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"os"
	"sync"
	"wolf/pkg"
)

type Phase int

type EventType string

const (
	PenguinActionStart   EventType = "penguin_action_start"
	PenguinActionEnd     EventType = "penguin_action_end"
	WolfActionStart      EventType = "wolf_action_start"
	WolfActionEnd        EventType = "wolf_action_end"
	CrowActionStart      EventType = "crow_action_start"
	CrowActionEnd        EventType = "crow_action_end"
	BearActionEnd        EventType = "bear_action_end"
	WolfKingAction       EventType = "wolf_king_action_end"
	StealthWolfActionEnd EventType = "stealth_wolf_action_end"
	HunterActionEnd      EventType = "hunter_action_end"
	FoxActionEnd         EventType = "fox_action_end"
	BatActionEnd         EventType = "bat_action_end"
	CloneActionEnd       EventType = "clone_action_end"
	SelfDestruct         EventType = "self_destruct"
	Vote                 EventType = "vote"
	WolfVote             EventType = "wolf_vote"
	Night                EventType = "night"
)

func GameFlow(game *Game) {
	switch game.Phase {
	case PhaseNight:
		GlobalEventBus.Publish(string(Night), game)
	}
}

func init() {
	GlobalEventBus.Subscribe(string(Night), func(event interface{}) {
		game := event.(*Game)
		for _, player := range game.Players {
			conn := GlobalConns.GetConn(player.Name)
			msg, _ := json.Marshal(SCMessage{Type: MsgChat, Content: "天黑请闭眼"})
			conn.WriteMessage(websocket.TextMessage, msg)
		}
		GlobalEventBus.Publish(string(PenguinActionStart), game)
	})
	GlobalEventBus.Subscribe(string(Vote), func(event interface{}) {
		game := event.(*Game)
		for _, player := range game.Players {
			conn := GlobalConns.GetConn(player.Name)
			msg1, _ := json.Marshal(SCMessage{Type: MsgChat, Content: "现在开始投票"})
			msg2, _ := json.Marshal(SCMessage{Type: MsgChat, Content: "请选择你要投票的玩家,输入玩家的座位号"})
			conn.WriteMessage(websocket.TextMessage, msg1)
			conn.WriteMessage(websocket.TextMessage, msg2)
		}
	})
	GlobalEventBus.Subscribe(string(SelfDestruct), func(event interface{}) {
		game := event.(*Game)
		game.Phase = PhaseNight
	})
	GlobalEventBus.Subscribe(string(PenguinActionStart), func(event interface{}) {
		game := event.(*Game)
		var penguin *pkg.Player
		for _, player := range game.Players {
			if player.Role1.Name == "penguin" || player.Role2.Name == "penguin" {
				penguin = player
				break
			}
		}
		conn := GlobalConns.GetConn(penguin.Name)
		msg, _ := json.Marshal(SCMessage{Type: MsgChat, Content: "请选择你要冰冻的人，输入玩家的座位号"})
		conn.WriteMessage(websocket.TextMessage, msg)
	})
	GlobalEventBus.Subscribe(string(PenguinActionEnd), func(event interface {
	}) {
		target := event.(string)
		player := GlobalConns.GetPlayer(target)
		player.TopRole.IsFrozen = true
		for _, handler := range GlobalEventBus.Subscribers[string(PenguinActionEnd)] {
			handler(event)
		}
	})
	GlobalEventBus.Subscribe(string(WolfVote), func(event interface{}) {
		room := GlobalRoomManager.GetRoom(event.(string))
		for _, player := range room.Players {
			conn := GlobalConns.GetConn(player.Name)
			msg, _ := json.Marshal(SCMessage{Type: MsgChat, Content: "狼人开始行动"})
			conn.WriteMessage(websocket.TextMessage, msg)
		}
		for _, player := range room.Players {
			if player.TopRole.HasSkill("狼人") || player.Role2.HasSkill("潜行狼") {
				conn := GlobalConns.GetConn(player.Name)
				msg, _ := json.Marshal(SCMessage{Type: MsgChat, Content: "请选择你要击杀的目标,输入玩家的座位号"})
				conn.WriteMessage(websocket.TextMessage, msg)
			}
		}
	})
	GlobalEventBus.Subscribe(string(CrowActionStart), func(event interface {
	}) {
		target := event.(string)
		player := GlobalConns.GetPlayer(target)
		player.TopRole.IsCursed = true
	})
	GlobalEventBus.Subscribe(string(WolfActionEnd), func(event interface{}) {

	})
}

const (
	PhaseNight Phase = iota
	PhaseDay
	PhaseVote
	PhaseEnd
)

var NightOrder = []string{
	"penguin",
	"wolf",
}

var GlobalEventBus = NewEventBus()

type EventBus struct {
	Subscribers map[string][]func(event interface{})
	Mu          sync.Mutex
	Players     []*pkg.Player
}

func NewEventBus() *EventBus {
	return &EventBus{
		Subscribers: make(map[string][]func(event interface{})),
	}
}

func (eb *EventBus) Subscribe(eventType string, handler func(event interface{})) {
	eb.Mu.Lock()
	defer eb.Mu.Unlock()
	eb.Subscribers[eventType] = append(eb.Subscribers[eventType], handler)
}

func (eb *EventBus) Publish(eventType string, event interface{}) {
	eb.Mu.Lock()
	defer eb.Mu.Unlock()
	for _, handler := range eb.Subscribers[eventType] {
		handler(event)
	}
}

type Game struct {
	RoomID          string
	Phase           Phase
	TurnIndex       int
	Players         []*pkg.Player
	VoteMap         map[string]int
	NightActions    map[string][]int
	DeadQueue       []int
	Winner          string
	WolfVotes       map[string]int
	WolfVotesStatus map[string]bool
	Done            chan bool
}

func (g *Game) NotifyPlayer(player *pkg.Player, msg SCMessage) {
	msga := SCMessage{
		Type:    msg.Type,
		RoomID:  g.RoomID,
		Content: msg.Content,
	}
	err := GlobalConns.SendToPlayer(player.Name, &msga)
	if err != nil {
		return
	}
}

func (g *Game) NotifyWolfToVote(player *pkg.Player) *CSMessage {
	msg := SCMessage{
		Type:    MsgNightAction,
		RoomID:  g.RoomID,
		Content: `请选择击杀目标(输入数字)`,
	}
	err := GlobalConns.SendToPlayer(player.ID, &msg)
	if err != nil {
		return nil
	}
	_, i, err := GlobalConns.conns[player.ID].ReadMessage()
	var getMsg CSMessage
	err = json.Unmarshal(i, &getMsg)
	if err != nil {
		return nil
	}
	return &getMsg
}

func (g *Game) NotifyAllPlayers(content string) {
	for _, player := range g.Players {
		g.NotifyPlayer(player, SCMessage{Type: MsgChat, RoomID: g.RoomID, Content: content})
	}
}

func (g *Game) GetAliveWolfs() []*pkg.Player {
	var wolfs []*pkg.Player
	for _, player := range g.Players {
		if player.Status != "dead" {
			if (player.Role1.HasSkill("wolf") || player.Role1.HasSkill("wolf_king")) && !player.Role1.Death() {
				wolfs = append(wolfs, player)
			} else if player.Role2.HasSkill("stealth_wolf") && !player.Role2.Death() {
				wolfs = append(wolfs, player)
			} else if (player.Role2.Name == "wolf" || player.Role2.Name == "wolf_king") && !player.Role2.Death() {
				wolfs = append(wolfs, player)
			}
		}
	}
	return wolfs
}

func (g *Game) DoWolfPhase() {
	var canActive = true
	wolfPlayers := g.GetAliveWolfs()
	for _, player := range wolfPlayers {
		if !g.CanWolCampAct(player) {
			canActive = false
		}
	}
	if !canActive {
		for _, player := range wolfPlayers {
			g.NotifyPlayer(player, SCMessage{Type: MsgChat, RoomID: g.RoomID, Content: "由于你/你的队友被冰冻，所以你们无法行动"})
		}
		return
	}
	g.WolfVotes = make(map[string]int)
	g.WolfVotesStatus = make(map[string]bool)
	scanner := bufio.NewScanner(os.Stdin)
	for _, player := range wolfPlayers {
		g.NotifyWolfToVote(player)
		target := scanner.Text()
		g.WolfVotes[target]++
	}
}

func (g *Game) CanWolCampAct(p *pkg.Player) bool {
	if !p.Role1.Death() && !p.Role1.IsFrozen {
		return true
	} else if !p.Role2.IsFrozen && p.Role2.Name == "stealth_wolf" {
		return true
	} else {
		return false
	}
}

func (g *Game) IsReady() bool {
	for _, player := range g.Players {
		if !player.Ready {
			return false
		}
	}
	return true
}

func (g *Game) SetFirst(player *pkg.Player, role1 *pkg.Role, role2 *pkg.Role) {
	fmt.Printf("[DEBUG] SetFirst: player=%s, sending MsgSetFirst\n", player.Name)
	var msg SCMessage
	msg.Type = MsgSetFirst
	msg.RoomID = g.RoomID
	player.Role1 = role1
	player.Role2 = role2
	msg.Content = `你的身份是` + role1.Name + `和` + role2.Name + "\n" + "请选择你的身份顺序(只需要输入你想先使用的身份)"
	g.NotifyPlayer(player, msg)
}
