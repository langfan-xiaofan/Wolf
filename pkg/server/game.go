package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"wolf/pkg"
)

type Phase int

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
