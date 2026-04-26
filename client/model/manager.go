package model

import "wolf/pkg"

type Manager struct {
	Roles []*pkg.Role `json:"roles"`
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) AddRole(role *pkg.Role) {
}
