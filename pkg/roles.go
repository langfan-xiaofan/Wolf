package pkg

type RoleInterface interface {
	Death() bool
	CheckIsAction() bool
	IsActionable() bool
	SetActionable(bool)
	SetDeath(bool)
	SetAction(bool)
	SetCurse(bool)
	CheckIsCurse() bool
}

type BaseRole struct {
	Name        string          `json:"name"`
	IsDeath     bool            `json:"death"`
	Actionable  bool            `json:"actionable"`
	Des         string          `json:"des"`
	IsAction    bool            `json:"is_action"`
	IsCursed    bool            `json:"is_curse"`
	IsFrozen    bool            `json:"is_frozen"`
	SkillStatus map[string]bool `json:"skill_status"`
}

func (b *BaseRole) Death() bool {
	return b.IsDeath
}

func (b *BaseRole) IsActionable() bool {
	return b.Actionable
}

func (b *BaseRole) SetActionable(status bool) {
	b.Actionable = status
}

func (b *BaseRole) SetDeath(death bool) {
	b.IsDeath = death
}

func (b *BaseRole) CheckIsAction() bool {
	return b.IsAction
}

func (b *BaseRole) SetAction(status bool) {
	b.IsAction = status
}

func (b *BaseRole) SetCurse(status bool) {
	b.IsCursed = status
}

func (b *BaseRole) CheckIsCurse() bool {
	return b.IsCursed
}

func (b *BaseRole) SetFrozen(status bool) {
	b.IsFrozen = status
}

func (b *BaseRole) GetFrozen() bool {
	return b.IsFrozen
}

type SkillContext struct {
	Self    *Player        `json:"self"`
	Targets []*Player      `json:"targets"`
	Players []*Player      `json:"players"`
	Mid     int            `json:"mid"`
	MP      map[string]int `json:"mp"`
	Extra   interface{}    `json:"extra"`
	Camp    string         `json:"camp"`
}

type Skill interface {
	GetName() string
	Execute(ctx *SkillContext)
}

type Role struct {
	BaseRole
	Skills []Skill `json:"skills"`
}

func (r *Role) UseSkill(skillName string, ctx *SkillContext) {
	for _, skill := range r.Skills {
		if skill.GetName() == skillName {
			skill.Execute(ctx)
		}
	}
}

func (r *Role) HasSkill(name string) bool {
	for _, skill := range r.Skills {
		if skill.GetName() == name {
			return true
		}
	}
	return false
}

type WolfSkill struct{}

func (w *WolfSkill) Execute(ctx *SkillContext) {
	if ctx.Targets == nil {
		return
	}
	for _, target := range ctx.Targets {
		if _, ok := ctx.MP[target.ID]; !ok {
			ctx.MP[target.ID] = 1
		} else {
			ctx.MP[target.ID]++
		}
	}
}

func (w *WolfSkill) GetName() string {
	return "狼人"
}

type WolfKing struct{}

func (w *WolfKing) Execute(ctx *SkillContext) {
	if ctx.Extra == "lose" {
		w.KillAsKing(ctx)
	} else {
		w.KillAsWolf(ctx)
	}
}

func (w *WolfKing) KillAsKing(ctx *SkillContext) {
	for _, target := range ctx.Targets {
		if !target.Role1.Death() {
			target.Role1.SetDeath(true)
		} else {
			target.Role2.SetDeath(true)
		}
	}
}

func (w *WolfKing) KillAsWolf(ctx *SkillContext) {
	for _, target := range ctx.Targets {
		if _, ok := ctx.MP[target.ID]; !ok {
			ctx.MP[target.ID] = 1
		} else {
			ctx.MP[target.ID]++
		}
	}
}

func (w *WolfKing) GetName() string {
	return "狼王"
}

type StealthWolf struct{}

func (w *StealthWolf) Execute(ctx *SkillContext) {
	if ctx.Targets == nil {
		return
	}
	for _, target := range ctx.Targets {
		if _, ok := ctx.MP[target.ID]; !ok {
			ctx.MP[target.ID] = 1
		} else {
			ctx.MP[target.ID]++
		}
	}
}

func (w *StealthWolf) GetName() string {
	return "潜行狼"
}

type PenguinSkill struct {
}

func (p *PenguinSkill) Execute(ctx *SkillContext) {
	if ctx.Targets == nil {
		return
	}
	for _, target := range ctx.Targets {
		if !target.Role1.Death() {
			target.Role1.SetFrozen(true)
		} else {
			target.Role2.SetFrozen(true)
		}
	}
}

func (p *PenguinSkill) GetName() string {
	return "企鹅"
}

func BtoI(a bool) int {
	if a {
		return 1
	}
	return 0
}

type BearSkill struct{}

func (b *BearSkill) Execute(ctx *SkillContext) {
	mid := ctx.Mid
	players := ctx.Players
	indices := []int{(mid - 1 + len(players)) % len(players), mid, (mid + 1 + len(players)) % len(players)}
	for _, index := range indices {
		if players[index].Role1.CheckIsAction() || players[index].Role2.CheckIsAction() {
			ctx.Extra = true
			return
		}
	}
	ctx.Extra = false
}

func (b *BearSkill) GetName() string {
	return "熊"
}

type HunterSkill struct{}

func (h *HunterSkill) Execute(ctx *SkillContext) {
	if ctx.Targets == nil {
		return
	}
	for _, player := range ctx.Targets {
		if !player.Role1.Death() {
			player.Role1.SetDeath(true)
		} else {
			player.Role2.SetDeath(true)
		}
	}
}

func (h *HunterSkill) GetName() string {
	return "猎人"
}

type FoxSkill struct{}

func (f *FoxSkill) Execute(ctx *SkillContext) {
	if ctx.Targets == nil {
		return
	}
	for _, player := range ctx.Targets {
		if !player.Role1.Death() {
			ctx.MP[player.ID] = BtoI(player.Role1.CheckIsAction())
		} else {
			ctx.MP[player.ID] = BtoI(player.Role2.CheckIsAction())
		}
	}
}

func (f *FoxSkill) GetName() string {
	return "狐狸"
}

type BatSkill struct {
}

func (b *BatSkill) Execute(ctx *SkillContext) {
	if ctx.Targets == nil {
		return
	}
	if ctx.Extra == "antidote" {
		b.UseAntidote(ctx)
	} else {
		b.UsePoison(ctx)
	}
}

func (b *BatSkill) UseAntidote(ctx *SkillContext) {
	for _, target := range ctx.Targets {
		if target.Role1.Death() {
			if target.Role2.Death() {
				target.Role2.SetDeath(false)
			} else {
				target.Role1.SetDeath(false)
			}
		}
	}
}

func (b *BatSkill) UsePoison(ctx *SkillContext) {
	for _, target := range ctx.Targets {
		if !target.Role1.Death() {
			target.Role1.SetDeath(true)
		} else {
			target.Role2.SetDeath(true)
		}
	}
}

func (b *BatSkill) GetName() string {
	return "蝙蝠"
}

type ClonePerson struct {
}

func (c *ClonePerson) Execute(ctx *SkillContext) {

}

func (c *ClonePerson) GetName() string {
	return "复制人"
}

type CrowSkill struct {
}

func (c *CrowSkill) Execute(ctx *SkillContext) {
	for _, player := range ctx.Targets {
		if !player.Role1.Death() {
			player.Role1.SetCurse(true)
		} else {
			player.Role2.SetCurse(true)
		}
	}
}

func (c *CrowSkill) GetName() string {
	return "乌鸦"
}

// NewRole 创建带技能的角色
func NewRole(roleType string) *Role {
	role := &Role{BaseRole: BaseRole{Name: roleType}}
	switch roleType {
	case "狼人":
		role.Name = "狼人"
		role.Skills = []Skill{&WolfSkill{}}
	case "狼王":
		role.Name = "狼王"
		role.Skills = []Skill{&WolfKing{}}
	case "潜行狼":
		role.Name = "潜行狼"
		role.Skills = []Skill{&StealthWolf{}}
	case "企鹅":
		role.Name = "企鹅"
		role.Skills = []Skill{&PenguinSkill{}}
	case "熊":
		role.Name = "熊"
		role.Skills = []Skill{&BearSkill{}}
	case "猎人":
		role.Name = "猎人"
		role.Skills = []Skill{&HunterSkill{}}
	case "狐狸":
		role.Name = "狐狸"
		role.Skills = []Skill{&FoxSkill{}}
	case "蝙蝠":
		role.Name = "蝙蝠"
		role.Skills = []Skill{&BatSkill{}}
	case "复制人":
		role.Name = "复制人"
		role.Skills = []Skill{&ClonePerson{}}
	case "乌鸦":
		role.Name = "乌鸦"
		role.Skills = []Skill{&CrowSkill{}}
	default:
		role.Name = "村民"
		role.Skills = []Skill{}
	}
	return role
}

type PlayerSkill interface {
	Spike() string
}
type Player struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Role1   *Role  `json:"role1"`
	Role2   *Role  `json:"role2"`
	Status  string `json:"status"`
	TopRole *Role  `json:"activity"`
	Ready   bool   `json:"ready"`
	Seat    int    `json:"seat"`
}

//func (p *Player) Spike() string {
//	panic("implement me")
//}

func (p *Player) SetReady(status bool) {
	p.Ready = status
}
