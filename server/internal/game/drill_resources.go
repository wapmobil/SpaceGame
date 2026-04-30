package game

import (
	"fmt"
	"math/rand"
	"sort"
)

// Cell types for the drill world
const (
	CellEmpty   = "empty"
	CellDirt    = "dirt"
	CellStone   = "stone"
	CellMetal   = "metal"
	CellMithril = "mithril"
)

// Resource types that can be found underground
const (
	ResourceOil      = "oil"
	ResourceGas      = "gas"
	ResourceCopper   = "copper"
	ResourceCoal     = "coal"
	ResourceSilver   = "silver"
	ResourceGold     = "gold"
	ResourcePlatinum = "platinum"
	ResourceDiamond  = "diamond"
	ResourceExotic   = "exotic"
)

// Resource definition
type ResourceDef struct {
	Type               string
	Name               string
	Icon               string
	Value              int     // base value in money
	DepthStart         int     // minimum depth to appear
	DepthEnd           int     // maximum depth to appear
	SpawnChance        float64 // base chance to spawn per cell (0-1)
	Damage             int     // damage to drill when passing through
	PreferredCellTypes []string
}

var resourceDefinitions = map[string]*ResourceDef{
	ResourceOil:      {Type: ResourceOil, Name: "Нефть", Icon: "🛢️", Value: 1, DepthStart: 0, DepthEnd: 1000, SpawnChance: 0.08, Damage: 3, PreferredCellTypes: []string{CellDirt}},
	ResourceGas:      {Type: ResourceGas, Name: "Газ", Icon: "💨", Value: 2, DepthStart: 0, DepthEnd: 1000, SpawnChance: 0.03, Damage: 3, PreferredCellTypes: []string{CellDirt}},
	ResourceCopper:   {Type: ResourceCopper, Name: "Медь", Icon: "🟠", Value: 10, DepthStart: 50, DepthEnd: 100, SpawnChance: 0.07, Damage: 5, PreferredCellTypes: []string{CellStone, CellMetal}},
	ResourceCoal:     {Type: ResourceCoal, Name: "Уголь", Icon: "⬛", Value: 5, DepthStart: 50, DepthEnd: 150, SpawnChance: 0.08, Damage: 5, PreferredCellTypes: []string{CellDirt, CellStone}},
	ResourceSilver:   {Type: ResourceSilver, Name: "Серебро", Icon: "⚪", Value: 15, DepthStart: 100, DepthEnd: 200, SpawnChance: 0.05, Damage: 8, PreferredCellTypes: []string{CellStone, CellMetal}},
	ResourceGold:     {Type: ResourceGold, Name: "Золото", Icon: "🟡", Value: 25, DepthStart: 150, DepthEnd: 300, SpawnChance: 0.04, Damage: 10, PreferredCellTypes: []string{CellMetal, CellMithril}},
	ResourcePlatinum: {Type: ResourcePlatinum, Name: "Платина", Icon: "🔘", Value: 30, DepthStart: 200, DepthEnd: 400, SpawnChance: 0.03, Damage: 12, PreferredCellTypes: []string{CellMetal, CellMithril}},
	ResourceDiamond:  {Type: ResourceDiamond, Name: "Алмазы", Icon: "💎", Value: 60, DepthStart: 300, DepthEnd: 500, SpawnChance: 0.02, Damage: 15, PreferredCellTypes: []string{CellMithril}},
	ResourceExotic:   {Type: ResourceExotic, Name: "Экзотика", Icon: "🔮", Value: 200, DepthStart: 500, DepthEnd: 9999, SpawnChance: 0.01, Damage: 20, PreferredCellTypes: []string{CellMithril, CellMetal}},
}

// DrillResource represents a collected resource in the session
type DrillResource struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Amount int    `json:"amount"`
	Value  int    `json:"value"`
}

// MoveResult represents the result of a drill move action
type MoveResult struct {
	Success     bool            `json:"success"`
	Message     string          `json:"message,omitempty"`
	DrillHP     int             `json:"drill_hp"`
	DrillMaxHP  int             `json:"drill_max_hp"`
	Depth       int             `json:"depth"`
	DrillX      int             `json:"drill_x"`
	Resources   []DrillResource `json:"resources"`
	TotalEarned int             `json:"total_earned"`
	GameEnded   bool            `json:"game_ended"`
	EndReason   string          `json:"end_reason,omitempty"`
	NewResource *ResourceHit    `json:"new_resource,omitempty"`
	Extracted   int             `json:"extracted,omitempty"`
	World       [][]Cell        `json:"world,omitempty"`
}

// ResourceHit represents hitting a new resource
type ResourceHit struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Amount int    `json:"amount"`
	Value  int    `json:"value"`
}

// Cell represents a single cell in the drill world
type Cell struct {
	X              int    `json:"x"`
	Y              int    `json:"y"`
	CellType       string `json:"cell_type"`
	ResourceType   string `json:"resource_type,omitempty"`
	ResourceAmount int    `json:"resource_amount,omitempty"`
	ResourceValue  int    `json:"resource_value,omitempty"`
	Extracted      bool   `json:"extracted"`
}

// getCellDamage returns the damage a cell type deals to the drill
func (g *DrillGame) getCellDamage(cellType string) int {
	switch cellType {
	case CellEmpty:
		return 0
	case CellDirt:
		return 2
	case CellStone:
		return 5
	case CellMetal:
		return 10
	case CellMithril:
		return 15
	default:
		return 0
	}
}

// addResource adds a resource to the session's collected resources
func (g *DrillGame) addResource(def *ResourceDef, amount int, value int) {
	// Check if resource already exists
	for i, r := range g.session.Resources {
		if r.Type == def.Type {
			g.session.Resources[i].Amount += amount
			g.session.Resources[i].Value += value
			g.session.TotalEarned += value
			return
		}
	}

	g.session.Resources = append(g.session.Resources, DrillResource{
		Type:   def.Type,
		Name:   def.Name,
		Icon:   def.Icon,
		Amount: amount,
		Value:  value,
	})
	g.session.TotalEarned += value
}

// convertResourcesToMoney converts all resources to money value
func (g *DrillGame) convertResourcesToMoney() {
	total := 0
	for _, r := range g.session.Resources {
		total += r.Value
	}
	g.session.TotalEarned = total
}

// selectResourceForDepth picks a random resource based on depth using a deterministic seed
func (g *DrillGame) selectResourceForDepth(depth int, seed int64) *ResourceDef {
	mixed := uint64(seed)
	mixed ^= mixed >> 30
	mixed *= 0xbf58476d1ce4e5b9
	mixed ^= mixed >> 27
	mixed *= 0x94d049bb133111eb
	mixed ^= mixed >> 31
	r := rand.New(rand.NewSource(int64(mixed)))
	var candidates []*ResourceDef
	for _, def := range resourceDefinitions {
		if depth >= def.DepthStart && depth <= def.DepthEnd {
			candidates = append(candidates, def)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort candidates by type for deterministic iteration
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Type < candidates[j].Type
	})

	// Weight by spawn chance (use integer weights: SpawnChance * 100)
	totalWeight := 0
	for _, c := range candidates {
		totalWeight += int(c.SpawnChance * 100)
	}

	rVal := r.Int63n(int64(totalWeight))
	for _, c := range candidates {
		weight := int(c.SpawnChance * 100)
		if rVal < int64(weight) {
			return c
		}
		rVal -= int64(weight)
	}

	return candidates[0]
}

// getResourceCellTypeBonus returns a multiplier based on cell type preference for a resource
func getResourceCellTypeBonus(cellType string, resource *ResourceDef) float64 {
	if resource == nil {
		return 0.5
	}
	for _, ct := range resource.PreferredCellTypes {
		if cellType == ct {
			return 2.0
		}
	}
	return 0.5
}

// getCellDamageFromType returns the damage a cell type deals to the drill (standalone)
func getCellDamageFromType(cellType string) int {
	switch cellType {
	case CellEmpty:
		return 0
	case CellDirt:
		return 2
	case CellStone:
		return 5
	case CellMetal:
		return 10
	case CellMithril:
		return 15
	default:
		return 0
	}
}

// getCellAt deterministically generates a cell based on seed and coordinates
func (g *DrillGame) getCellAt(x, y int) Cell {
	key := fmt.Sprintf("%d,%d", x, y)
	if cell, ok := g.cellsCache[key]; ok {
		return cell
	}

	cell := Cell{X: x, Y: y}

	// Check for cave first
	if g.isCave(x, y) {
		cell.CellType = CellEmpty
		g.cellsCache[key] = cell
		return cell
	}

	// Determine cell type using geological zones + noise
	zone := g.getDepthZone(y)
	terrainNoise := g.getTerrainNoise(x, y)
	cell.CellType = g.selectCellType(zone, terrainNoise, y)

	// Resource spawning with veins + cell type correlation
	veinMultiplier := g.getVeinMultiplier(x, y)

	typeSeed := g.config.Seed + int64(x)*1000003 + int64(y)*1000033
	r := rand.New(rand.NewSource(typeSeed))

	baseSpawnChance := 0.15
	if r.Float64() < baseSpawnChance*veinMultiplier {
		resource := g.selectResourceForDepth(y, typeSeed)
		if resource != nil {
			cellTypeBonus := getResourceCellTypeBonus(cell.CellType, resource)
			if r.Float64() < cellTypeBonus {
				cell.ResourceType = resource.Type
				cell.ResourceAmount = 1
				cell.ResourceValue = resource.Value
			}
		}
	}

	g.cellsCache[key] = cell
	return cell
}
