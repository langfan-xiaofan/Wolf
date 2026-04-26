package main

type ClientState struct {
	PlayerID string   `json:"player_id"`
	RoomID   string   `json:"room_id"`
	Phase    string   `json:"phase"`
	MyTurn   bool     `json:"my_turn"`
	CanInput bool     `json:"can_input"`
	Targets  []string `json:"targets"`
}
