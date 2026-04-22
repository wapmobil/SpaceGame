package game

// BuildingEntry represents a single building on a planet with all computed data.
type BuildingEntry struct {
	Type          string     `json:"type"`
	Level         int        `json:"level"`
	BuildProgress float64    `json:"build_progress"` // remaining seconds until completion
	Pending       bool       `json:"pending"`
	Enabled       bool       `json:"enabled"`
	BuildTime     float64    `json:"build_time"`     // total build time in seconds
	Cost          CostInfo   `json:"cost"`           // cost to build/upgrade to current level (already paid)
	NextCost      CostInfo   `json:"next_cost"`      // cost to build/upgrade to next level
	Production    ProdInfo   `json:"production"`     // per-tick resource production
	Consumption   float64    `json:"consumption"`    // per-tick energy consumption
}

// CostInfo represents the cost to build/upgrade a building.
type CostInfo struct {
	Food  float64 `json:"food"`
	Money float64 `json:"money"`
}

// ProdInfo represents per-tick resource production.
type ProdInfo struct {
	Food       float64 `json:"food"`
	Composite  float64 `json:"composite"`
	Mechanisms float64 `json:"mechanisms"`
	Reagents   float64 `json:"reagents"`
	Energy     float64 `json:"energy"`
	Money      float64 `json:"money"`
	AlienTech  float64 `json:"alien_tech"`
}

// Add adds production from another ProdInfo.
func (p *ProdInfo) Add(o ProdInfo) {
	p.Food += o.Food
	p.Composite += o.Composite
	p.Mechanisms += o.Mechanisms
	p.Reagents += o.Reagents
	p.Energy += o.Energy
	p.Money += o.Money
	p.AlienTech += o.AlienTech
}

// BuildingsOrder defines the display order of building types.
var BuildingsOrder = []string{
	"base", "farm", "solar", "storage", "factory",
	"energy_storage", "shipyard", "comcenter",
	"composite_drone", "mechanism_factory", "reagent_lab",
}
