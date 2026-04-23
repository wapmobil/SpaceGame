package research

import "log"

// TreeID identifies which research tree a technology belongs to.
type TreeID int

const (
	TreeStandard TreeID = 1
	TreeAlien    TreeID = 2
)

// EffectFunc is called when a technology is researched.
type EffectFunc func(planetID string, level int)

// Tech defines a single technology node in a research tree.
type Tech struct {
	ID          string
	Name        string
	Description string
	CostFood    float64
	CostMoney   float64
	CostAlien   float64
	BuildTime   float64 // seconds
	Tree        TreeID
	MaxLevel    int
	DependsOn   []string // tech IDs that must be completed first
	Effect      EffectFunc
}

// AllTechs returns every defined technology node.
func AllTechs() []*Tech {
	return []*Tech{
		// Tree 1 - Standard Research
		{
			ID:          "planet_exploration",
			Name:        "Planet Exploration",
			Description: "Unlocks Factory building",
			CostFood:    100,
			CostMoney:   100,
			BuildTime:   60,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   nil,
			Effect:      applyPlanetExploration,
		},
		{
			ID:          "energy_storage",
			Name:        "Energy Storage",
			Description: "Unlocks Energy Storage building",
			CostFood:    200,
			CostMoney:   150,
			BuildTime:   90,
			Tree:        TreeStandard,
			MaxLevel:    5,
			DependsOn:   []string{"planet_exploration"},
			Effect:      applyEnergyStorage,
		},
		{
			ID:          "energy_saving",
			Name:        "Energy Saving",
			Description: "-10% energy consumption per level (up to 4 levels)",
			CostFood:    300,
			CostMoney:   200,
			BuildTime:   120,
			Tree:        TreeStandard,
			MaxLevel:    4,
			DependsOn:   []string{"energy_storage"},
			Effect:      applyEnergySaving,
		},
		{
			ID:          "trade",
			Name:        "Trade",
			Description: "Unlocks Marketplace",
			CostFood:    400,
			CostMoney:   300,
			BuildTime:   120,
			Tree:        TreeStandard,
			MaxLevel:    2,
			DependsOn:   []string{"planet_exploration"},
			Effect:      applyTrade,
		},
		{
			ID:          "ships",
			Name:        "Ships",
			Description: "Unlocks Shipyard",
			CostFood:    500,
			CostMoney:   400,
			BuildTime:   150,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"planet_exploration"},
			Effect:      applyShips,
		},
		{
			ID:          "upgraded_energy_storage",
			Name:        "Upgraded Energy Storage",
			Description: "+20% energy capacity per level (up to 3 levels)",
			CostFood:    600,
			CostMoney:   500,
			BuildTime:   180,
			Tree:        TreeStandard,
			MaxLevel:    3,
			DependsOn:   []string{"energy_saving"},
			Effect:      applyUpgradedEnergyStorage,
		},
		{
			ID:          "fast_construction",
			Name:        "Fast Construction",
			Description: "Building speed bonus per level (up to 3 levels)",
			CostFood:    800,
			CostMoney:   600,
			BuildTime:   200,
			Tree:        TreeStandard,
			MaxLevel:    3,
			DependsOn:   []string{"ships"},
			Effect:      applyFastConstruction,
		},
		{
			ID:          "compact_storage",
			Name:        "Compact Storage",
			Description: "2x storage capacity per level (up to 3 levels)",
			CostFood:    1000,
			CostMoney:   800,
			BuildTime:   240,
			Tree:        TreeStandard,
			MaxLevel:    3,
			DependsOn:   []string{"fast_construction"},
			Effect:      applyCompactStorage,
		},
		{
			ID:          "parallel_construction",
			Name:        "Parallel Construction",
			Description: "+1 simultaneous construction project per level (up to 3 levels)",
			CostFood:    2000,
			CostMoney:   1500,
			BuildTime:   300,
			Tree:        TreeStandard,
			MaxLevel:    3,
			DependsOn:   []string{"fast_construction", "compact_storage"},
			Effect:      applyParallelConstruction,
		},
		{
			ID:          "expeditions",
			Name:        "Expeditions",
			Description: "Unlocks expedition system",
			CostFood:    1500,
			CostMoney:   1000,
			BuildTime:   300,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"trade"},
			Effect:      applyExpeditions,
		},
		{
			ID:          "command_center",
			Name:        "Command Center",
			Description: "Unlocks second research tree (Alien Technology)",
			CostFood:    5000,
			CostMoney:   3000,
			BuildTime:   600,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"expeditions"},
			Effect:      applyCommandCenter,
		},
		{
			ID:          "trade_connections",
			Name:        "Trade Connections",
			Description: "Unlocks advanced trading options",
			CostFood:    600,
			CostMoney:   450,
			BuildTime:   150,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"trade"},
			Effect:      applyTradeConnections,
		},
		{
			ID:          "fast_construction_2",
			Name:        "Fast Construction 2",
			Description: "Further building speed bonus",
			CostFood:    1200,
			CostMoney:   900,
			BuildTime:   250,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"fast_construction"},
			Effect:      applyFastConstruction2,
		},
		{
			ID:          "compact_storage_2",
			Name:        "Compact Storage 2",
			Description: "4x storage capacity",
			CostFood:    1500,
			CostMoney:   1200,
			BuildTime:   300,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"compact_storage", "fast_construction_2"},
			Effect:      applyCompactStorage2,
		},
		{
			ID:          "fast_construction_3",
			Name:        "Fast Construction 3",
			Description: "Maximum building speed bonus",
			CostFood:    2000,
			CostMoney:   1500,
			BuildTime:   350,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"fast_construction_2"},
			Effect:      applyFastConstruction3,
		},
		{
			ID:          "compact_storage_3",
			Name:        "Compact Storage 3",
			Description: "8x storage capacity",
			CostFood:    2500,
			CostMoney:   2000,
			BuildTime:   400,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"compact_storage_2", "fast_construction_3"},
			Effect:      applyCompactStorage3,
		},
		{
			ID:          "upgraded_energy_storage_2",
			Name:        "Upgraded Energy Storage 2",
			Description: "Maximum energy capacity boost",
			CostFood:    800,
			CostMoney:   700,
			BuildTime:   200,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"upgraded_energy_storage"},
			Effect:      applyUpgradedEnergyStorage2,
		},
		// Tree 2 - Alien Technology
		{
			ID:          "alien_technologies",
			Name:        "Alien Technologies",
			Description: "Unlocks alien technology tree",
			CostAlien:   10,
			BuildTime:   300,
			Tree:        TreeAlien,
			MaxLevel:    1,
			DependsOn:   []string{"command_center"},
			Effect:      applyAlienTechnologies,
		},
		{
			ID:          "additional_expedition",
			Name:        "Additional Expedition",
			Description: "+1 concurrent expedition",
			CostAlien:   15,
			BuildTime:   200,
			Tree:        TreeAlien,
			MaxLevel:    1,
			DependsOn:   []string{"alien_technologies"},
			Effect:      applyAdditionalExpedition,
		},
		{
			ID:          "super_energy_storage",
			Name:        "Super Energy Storage",
			Description: "+20% energy capacity per level (up to 5 levels)",
			CostAlien:   20,
			BuildTime:   300,
			Tree:        TreeAlien,
			MaxLevel:    5,
			DependsOn:   []string{"alien_technologies"},
			Effect:      applySuperEnergyStorage,
		},
	}
}

// GetTechByID returns a tech by its ID, or nil if not found.
func GetTechByID(id string) *Tech {
	for _, t := range AllTechs() {
		if t.ID == id {
			return t
		}
	}
	return nil
}

// TechsByTree returns technologies for a specific tree.
func TechsByTree(tree TreeID) []*Tech {
	var result []*Tech
	for _, t := range AllTechs() {
		if t.Tree == tree {
			result = append(result, t)
		}
	}
	return result
}

// CanStart checks if a player has the resources to start researching this tech.
func (t *Tech) CanStart(food, money, alienTech float64) bool {
	if t.CostFood > food {
		return false
	}
	if t.CostMoney > money {
		return false
	}
	if t.CostAlien > alienTech {
		return false
	}
	return true
}

// DeductCost removes the cost from the given resources.
func (t *Tech) DeductCost(food, money, alienTech *float64) {
	*food -= t.CostFood
	*money -= t.CostMoney
	*alienTech -= t.CostAlien
}

// Effect names for each tech.
func (t *Tech) EffectName() string {
	switch t.ID {
	case "planet_exploration":
		return "enable_factory"
	case "energy_storage":
		return "enable_accum"
	case "energy_saving":
		return "eco_power"
	case "trade":
		return "enable_trading"
	case "ships":
		return "enable_ships"
	case "upgraded_energy_storage":
		return "upgrade_accum"
	case "fast_construction":
		return "fastbuild"
	case "parallel_construction":
		return "parallel_construction"
	case "compact_storage":
		return "upgrade_capacity"
	case "expeditions":
		return "enable_expeditions"
	case "command_center":
		return "enable_commandcenter"
	case "alien_technologies":
		return "upgrage_inotech"
	case "additional_expedition":
		return "upgrage_max_expeditions"
	case "super_energy_storage":
		return "upgrade_super_accum"
	case "trade_connections":
		return "trade_connections"
	case "fast_construction_2":
		return "fast_construction_2"
	case "compact_storage_2":
		return "compact_storage_2"
	case "fast_construction_3":
		return "fast_construction_3"
	case "compact_storage_3":
		return "compact_storage_3"
	case "upgraded_energy_storage_2":
		return "upgraded_energy_storage_2"
	default:
		return ""
	}
}

// Effect functions that modify planet behavior when research completes.

func applyTradeConnections(planetID string, level int) {
	log.Printf("[Research] Planet %s unlocked Trade Connections (level %d)", planetID, level)
}

func applyFastConstruction2(planetID string, level int) {
	log.Printf("[Research] Planet %s Fast Construction 2 (level %d)", planetID, level)
}

func applyCompactStorage2(planetID string, level int) {
	log.Printf("[Research] Planet %s Compact Storage 2 (level %d)", planetID, level)
}

func applyFastConstruction3(planetID string, level int) {
	log.Printf("[Research] Planet %s Fast Construction 3 (level %d)", planetID, level)
}

func applyCompactStorage3(planetID string, level int) {
	log.Printf("[Research] Planet %s Compact Storage 3 (level %d)", planetID, level)
}

func applyUpgradedEnergyStorage2(planetID string, level int) {
	log.Printf("[Research] Planet %s Upgraded Energy Storage 2 (level %d)", planetID, level)
}
