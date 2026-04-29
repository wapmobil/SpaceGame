package research

// TreeID identifies which research tree a technology belongs to.
type TreeID int

const (
	TreeStandard TreeID = 1
	TreeAlien    TreeID = 2
)

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
}

// AllTechs returns every defined technology node.
func AllTechs() []*Tech {
	return []*Tech{
		// Tree 1 - Standard Research
		{
			ID:          "planet_exploration",
			Name:        "Planet Exploration",
			Description: "Открывает систему планетарной разведки",
			CostFood:    100,
			CostMoney:   100,
			BuildTime:   60,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   nil,
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
	DependsOn:   nil,
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
		},
		{
			ID:          "space_expeditions",
			Name:        "Space Expeditions",
			Description: "Открывает космические экспедиции",
			CostFood:    1500,
			CostMoney:   1000,
			BuildTime:   300,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"trade"},
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
		},
		{
			ID:          "location_buildings",
			Name:        "Location Buildings",
			Description: "Позволяет строить здания на локациях",
			CostFood:    800,
			CostMoney:   600,
			BuildTime:   250,
			Tree:        TreeStandard,
			MaxLevel:    1,
			DependsOn:   []string{"planet_exploration"},
		},
		{
			ID:          "advanced_exploration",
			Name:        "Advanced Exploration",
			Description: "+1 слот локаций за уровень (макс. 4)",
			CostFood:    1000,
			CostMoney:   800,
			BuildTime:   300,
			Tree:        TreeStandard,
			MaxLevel:    3,
			DependsOn:   []string{"location_buildings"},
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

