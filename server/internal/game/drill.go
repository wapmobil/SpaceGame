package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Cell types for the drill world
const (
	CellEmpty = "empty"
	CellDirt  = "dirt"
	CellStone = "stone"
	CellMetal = "metal"
	CellMithril = "mithril"
)

// Resource types that can be found underground
const (
	ResourceOil     = "oil"
	ResourceGas     = "gas"
	ResourceCopper  = "copper"
	ResourceCoal    = "coal"
	ResourceSilver  = "silver"
	ResourceGold    = "gold"
	ResourcePlatinum = "platinum"
	ResourceDiamond = "diamond"
	ResourceExotic  = "exotic"
)

// Resource definition
type ResourceDef struct {
	Type        string
	Name        string
	Icon        string
	Value       float64 // base value in money
	DigTime     float64 // seconds to fully extract
	DepthStart  int     // minimum depth to appear
	DepthEnd    int     // maximum depth to appear
	SpawnChance float64 // base chance to spawn per cell (0-1)
	Damage      int     // damage to drill when passing through
}

var resourceDefinitions = map[string]*ResourceDef{
	ResourceOil:      {Type: ResourceOil, Name: "Нефть", Icon: "🛢️", Value: 15, DigTime: 3.0, DepthStart: 0, DepthEnd: 50, SpawnChance: 0.08, Damage: 3},
	ResourceGas:      {Type: ResourceGas, Name: "Газ", Icon: "💨", Value: 20, DigTime: 3.5, DepthStart: 0, DepthEnd: 50, SpawnChance: 0.06, Damage: 3},
	ResourceCopper:   {Type: ResourceCopper, Name: "Медь", Icon: "🟠", Value: 30, DigTime: 4.0, DepthStart: 50, DepthEnd: 100, SpawnChance: 0.07, Damage: 5},
	ResourceCoal:     {Type: ResourceCoal, Name: "Уголь", Icon: "⬛", Value: 25, DigTime: 3.5, DepthStart: 50, DepthEnd: 150, SpawnChance: 0.08, Damage: 5},
	ResourceSilver:   {Type: ResourceSilver, Name: "Серебро", Icon: "⚪", Value: 50, DigTime: 5.0, DepthStart: 100, DepthEnd: 200, SpawnChance: 0.05, Damage: 8},
	ResourceGold:     {Type: ResourceGold, Name: "Золото", Icon: "🟡", Value: 100, DigTime: 6.0, DepthStart: 150, DepthEnd: 300, SpawnChance: 0.04, Damage: 10},
	ResourcePlatinum: {Type: ResourcePlatinum, Name: "Платина", Icon: "🔘", Value: 150, DigTime: 7.0, DepthStart: 200, DepthEnd: 400, SpawnChance: 0.03, Damage: 12},
	ResourceDiamond:  {Type: ResourceDiamond, Name: "Алмазы", Icon: "💎", Value: 250, DigTime: 8.0, DepthStart: 300, DepthEnd: 500, SpawnChance: 0.02, Damage: 15},
	ResourceExotic:   {Type: ResourceExotic, Name: "Экзотика", Icon: "🔮", Value: 500, DigTime: 10.0, DepthStart: 500, DepthEnd: 9999, SpawnChance: 0.01, Damage: 20},
}

// Cell represents a single cell in the drill world
type Cell struct {
	X            int     `json:"x"`
	Y            int     `json:"y"`
	CellType     string  `json:"cell_type"`
	ResourceType string  `json:"resource_type,omitempty"`
	ResourceAmount float64 `json:"resource_amount,omitempty"`
	ResourceValue float64 `json:"resource_value,omitempty"`
	Extracted    bool    `json:"extracted"`
}

// DrillSession represents the current state of a drill session
type DrillSession struct {
	ID             string    `json:"id"`
	SessionID      string    `json:"session_id"`
	PlanetID       string    `json:"planet_id"`
	PlayerID       string    `json:"player_id"`
	DrillHP        int       `json:"drill_hp"`
	DrillMaxHP     int       `json:"drill_max_hp"`
	Depth          int       `json:"depth"`
	DrillX         int       `json:"drill_x"`
	WorldWidth     int       `json:"world_width"`
	Resources      []DrillResource `json:"resources"`
	Status         string    `json:"status"`
	TotalEarned    float64   `json:"total_earned"`
	World          [][]Cell  `json:"world"`
	ViewTop        int       `json:"-"`
	ViewHeight     int       `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	LastMoveTime   time.Time `json:"last_move_time"`
}

// DrillResource represents a collected resource in the session
type DrillResource struct {
	Type     string  `json:"type"`
	Name     string  `json:"name"`
	Icon     string  `json:"icon"`
	Amount   float64 `json:"amount"`
	Value    float64 `json:"value"`
}

// MoveResult represents the result of a drill move action
type MoveResult struct {
	Success      bool       `json:"success"`
	Message      string     `json:"message,omitempty"`
	DrillHP      int        `json:"drill_hp"`
	DrillMaxHP   int        `json:"drill_max_hp"`
	Depth        int        `json:"depth"`
	DrillX       int        `json:"drill_x"`
	Resources    []DrillResource `json:"resources"`
	TotalEarned  float64    `json:"total_earned"`
	GameEnded    bool       `json:"game_ended"`
	EndReason    string     `json:"end_reason,omitempty"`
	NewResource  *ResourceHit `json:"new_resource,omitempty"`
	Extracted    float64    `json:"extracted,omitempty"`
}

// ResourceHit represents hitting a new resource
type ResourceHit struct {
	Type   string  `json:"type"`
	Name   string  `json:"name"`
	Icon   string  `json:"icon"`
	Amount float64 `json:"amount"`
	Value  float64 `json:"value"`
}

// MoveDirection represents horizontal movement
type MoveDirection int

const (
	MoveLeft MoveDirection = iota
	MoveRight
	MoveDown
)

// DrillConfig holds configuration for drill world generation
type DrillConfig struct {
	WorldWidth int
	ViewHeight int
	Seed       int64
}

// DrillGame is the main engine for the drill mini-game
type DrillGame struct {
	config   DrillConfig
	session  DrillSession
	rng      *rand.Rand
	baseLevel int
}

// Default drill configuration
const (
	DefaultWorldWidth = 20
	DefaultViewHeight = 15
	DefaultSeed       = 42
)

// NewDrillGame creates a new drill game session
func NewDrillGame(planetID, playerID string, baseLevel int) *DrillGame {
	maxHP := 80 + baseLevel*40
	session := DrillSession{
		ID:         uuid.New().String(),
		SessionID:  fmt.Sprintf("drill-%d", time.Now().UnixNano()),
		PlanetID:   planetID,
		PlayerID:   playerID,
		DrillHP:    maxHP,
		DrillMaxHP: maxHP,
		Depth:      0,
		DrillX:     DefaultWorldWidth / 2,
		WorldWidth: DefaultWorldWidth,
		Resources:  []DrillResource{},
		Status:     "active",
		TotalEarned: 0,
		ViewHeight: DefaultViewHeight,
		CreatedAt:  time.Now(),
	}

	config := DrillConfig{
		WorldWidth: DefaultWorldWidth,
		ViewHeight: DefaultViewHeight,
		Seed:       time.Now().UnixNano() + int64(planetID[len(planetID)-1]),
	}

	game := &DrillGame{
		config:    config,
		session:   session,
		rng:       rand.New(rand.NewSource(config.Seed)),
		baseLevel: baseLevel,
	}

	game.generateWorld()
	return game
}

// generateWorld generates the drill world
func (g *DrillGame) generateWorld() {
	worldHeight := 800 // max depth
	world := make([][]Cell, worldHeight)

	for y := 0; y < worldHeight; y++ {
		world[y] = make([]Cell, g.config.WorldWidth)
		for x := 0; x < g.config.WorldWidth; x++ {
			cell := Cell{X: x, Y: y, CellType: CellEmpty}

			depthFactor := float64(y) / 800.0

			// Determine cell type based on depth
			r := g.rng.Float64()
			if r < 0.3+depthFactor*0.2 {
				cell.CellType = CellDirt
				if cell.CellType == CellDirt {
					cell.ResourceType = ""
				}
			} else if r < 0.55+depthFactor*0.15 {
				cell.CellType = CellStone
			} else if r < 0.7+depthFactor*0.1 {
				cell.CellType = CellMetal
			} else if depthFactor > 0.6 && r < 0.75+depthFactor*0.05 {
				cell.CellType = CellMithril
			}

			// Try to spawn a resource
			if cell.CellType != CellEmpty && g.rng.Float64() < 0.12 {
				resource := g.selectResource(depthFactor)
				if resource != nil {
					cell.ResourceType = resource.Type
					cell.ResourceAmount = float64(g.rng.Intn(5)+3)
					cell.ResourceValue = resource.Value * cell.ResourceAmount
				}
			}

			world[y][x] = cell
		}
	}

	g.session.World = world
}

// selectResource picks a random resource based on depth
func (g *DrillGame) selectResource(depthFactor float64) *ResourceDef {
	var candidates []*ResourceDef
	for _, def := range resourceDefinitions {
		factor := float64(def.DepthStart) / 800.0
		endFactor := float64(def.DepthEnd) / 800.0
		if depthFactor >= factor && depthFactor <= endFactor {
			candidates = append(candidates, def)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Weight by spawn chance
	totalChance := 0.0
	for _, c := range candidates {
		totalChance += c.SpawnChance
	}

	r := g.rng.Float64() * totalChance
	accumulated := 0.0
	for _, c := range candidates {
		accumulated += c.SpawnChance
		if r <= accumulated {
			return c
		}
	}

	return candidates[0]
}

// Move processes a drill move action
func (g *DrillGame) Move(direction MoveDirection, extract bool) *MoveResult {
	if g.session.Status != "active" {
		return &MoveResult{
			Success:   false,
			Message:   "Drill session is not active",
			GameEnded: true,
			EndReason: "session_ended",
		}
	}

	result := &MoveResult{
		DrillHP:     g.session.DrillHP,
		DrillMaxHP:  g.session.DrillMaxHP,
		Depth:       g.session.Depth,
		DrillX:      g.session.DrillX,
		Resources:   g.session.Resources,
		TotalEarned: g.session.TotalEarned,
	}

	// Handle horizontal movement
	if direction == MoveLeft {
		if g.session.DrillX > 0 {
			g.session.DrillX--
			result.DrillX = g.session.DrillX
			result.Success = true
		}
	} else if direction == MoveRight {
		if g.session.DrillX < g.config.WorldWidth-1 {
			g.session.DrillX++
			result.DrillX = g.session.DrillX
			result.Success = true
		}
	} else if direction == MoveDown {
		result.Success = true
		g.processDrillDown(result)
		result.DrillX = g.session.DrillX
	} else {
		result.Success = true
	}

	// Handle extraction
	if extract && direction != MoveDown {
		g.processExtraction(result)
	}

	// Check if drill is destroyed
	if g.session.DrillHP <= 0 {
		g.session.DrillHP = 0
		g.session.Status = "failed"
		now := time.Now()
		g.session.CompletedAt = &now
		result.GameEnded = true
		result.EndReason = "drill_destroyed"
		g.saveSession()
		return result
	}

	g.session.LastMoveTime = time.Now()
	g.saveSession()
	return result
}

// processDrillDown moves the drill downward
func (g *DrillGame) processDrillDown(result *MoveResult) {
	world := g.session.World
	newDepth := g.session.Depth + 1

	if newDepth >= len(world) {
		// Reached the bottom
		g.session.Status = "completed"
		now := time.Now()
		g.session.CompletedAt = &now
		result.GameEnded = true
		result.EndReason = "reached_bottom"
		g.convertResourcesToMoney()
		g.saveSession()
		return
	}

	// Get cell at new depth, same X
	cell := world[newDepth][g.session.DrillX]

	// Apply damage from cell type
	damage := g.getCellDamage(cell.CellType)
	g.session.DrillHP -= damage

	// Check if there's a resource at this position
	if cell.ResourceType != "" && !cell.Extracted {
		def, ok := resourceDefinitions[cell.ResourceType]
		if ok {
			result.NewResource = &ResourceHit{
				Type:   def.Type,
				Name:   def.Name,
				Icon:   def.Icon,
				Amount: cell.ResourceAmount,
				Value:  cell.ResourceValue,
			}
			// Mark resource as passively extracted (small amount)
			extracted := cell.ResourceAmount * 0.1
			cell.ResourceAmount -= extracted
			if cell.ResourceAmount <= 0 {
				cell.ResourceAmount = 0
				cell.Extracted = true
			}
			// Update world
			world[newDepth][g.session.DrillX] = cell

			// Add to collected resources
			g.addResource(def, extracted, def.Value*extracted)
			result.Extracted = extracted
		}
	}

	g.session.Depth = newDepth
}

// processExtraction handles resource extraction when holding extract button
func (g *DrillGame) processExtraction(result *MoveResult) {
	world := g.session.World
	currentCell := world[g.session.Depth][g.session.DrillX]

	if currentCell.ResourceType != "" && !currentCell.Extracted {
		def, ok := resourceDefinitions[currentCell.ResourceType]
		if ok {
			// Extract amount per move
			extractRate := currentCell.ResourceAmount / def.DigTime * 0.5
			extracted := extractRate
			currentCell.ResourceAmount -= extracted
			if currentCell.ResourceAmount <= 0 {
				currentCell.ResourceAmount = 0
				currentCell.Extracted = true
			}
			// Update world
			world[g.session.Depth][g.session.DrillX] = currentCell

			g.addResource(def, extracted, def.Value*extracted)
			result.Extracted = extracted

			// Check if fully extracted
			if currentCell.ResourceAmount <= 0 {
				result.NewResource = nil
			}
		}
	}
}

// getCellDamage returns the damage a cell type deals to the drill
func (g *DrillGame) getCellDamage(cellType string) int {
	switch cellType {
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
func (g *DrillGame) addResource(def *ResourceDef, amount float64, value float64) {
	// Check if resource already exists
	for i, r := range g.session.Resources {
		if r.Type == def.Type {
			g.session.Resources[i].Amount += amount
			g.session.Resources[i].Value += value
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
}

// convertResourcesToMoney converts all resources to money value
func (g *DrillGame) convertResourcesToMoney() {
	total := 0.0
	for _, r := range g.session.Resources {
		total += r.Value
	}
	g.session.TotalEarned = total
}

// GetSession returns the current drill session state
func (g *DrillGame) GetSession() DrillSession {
	return g.session
}

// GetDisplayWorld returns the visible portion of the world for rendering
func (g *DrillGame) GetDisplayWorld() [][]Cell {
	world := g.session.World
	viewTop := g.session.Depth - g.config.ViewHeight/2
	if viewTop < 0 {
		viewTop = 0
	}
	viewBottom := viewTop + g.config.ViewHeight
	if viewBottom > len(world) {
		viewBottom = len(world)
	}

	display := make([][]Cell, viewBottom-viewTop)
	for i := viewTop; i < viewBottom; i++ {
		display[i-viewTop] = make([]Cell, len(world[i]))
		copy(display[i-viewTop], world[i])
	}

	return display
}

// GetAvailableDirections returns the available movement directions
func (g *DrillGame) GetAvailableDirections() []MoveDirection {
	var dirs []MoveDirection
	if g.session.DrillX > 0 {
		dirs = append(dirs, MoveLeft)
	}
	if g.session.DrillX < g.config.WorldWidth-1 {
		dirs = append(dirs, MoveRight)
	}
	if g.session.Status == "active" {
		dirs = append(dirs, MoveDown)
	}
	return dirs
}

// GetDrillDirectionString converts a MoveDirection to a string
func GetDrillDirectionString(dir MoveDirection) string {
	switch dir {
	case MoveLeft:
		return "left"
	case MoveRight:
		return "right"
	case MoveDown:
		return "down"
	default:
		return "unknown"
	}
}

// ParseDrillDirection parses a string to a MoveDirection
func ParseDrillDirection(s string) (MoveDirection, error) {
	switch s {
	case "left", "l", "west":
		return MoveLeft, nil
	case "right", "r", "east":
		return MoveRight, nil
	case "down", "d", "south", "":
		return MoveDown, nil
	default:
		return MoveDown, fmt.Errorf("invalid direction: %s", s)
	}
}

// GetDrillCooldown returns the cooldown between drill sessions
func GetDrillCooldown() time.Duration {
	return 30 * time.Second
}

// saveSession saves the current session state to JSONB in the world data
func (g *DrillGame) saveSession() {
	// World is already stored in the session struct
	// The actual DB save is handled by the API handler
}

// LoadGameFromState creates a DrillGame from existing state
func LoadGameFromState(planetID, playerID string, baseLevel int, world [][]Cell, drillHP, drillMaxHP, depth, drillX int, resources []DrillResource, totalEarned float64, status string) *DrillGame {
	if status == "" {
		status = "active"
	}
	session := DrillSession{
		ID:         uuid.New().String(),
		SessionID:  fmt.Sprintf("drill-%d", time.Now().UnixNano()),
		PlanetID:   planetID,
		PlayerID:   playerID,
		DrillHP:    drillHP,
		DrillMaxHP: drillMaxHP,
		Depth:      depth,
		DrillX:     drillX,
		WorldWidth: DefaultWorldWidth,
		Resources:  resources,
		Status:     status,
		TotalEarned: totalEarned,
		ViewHeight: DefaultViewHeight,
		CreatedAt:  time.Now(),
		World:      world,
	}

	config := DrillConfig{
		WorldWidth: DefaultWorldWidth,
		ViewHeight: DefaultViewHeight,
		Seed:       time.Now().UnixNano(),
	}

	return &DrillGame{
		config:    config,
		session:   session,
		rng:       rand.New(rand.NewSource(config.Seed)),
		baseLevel: baseLevel,
	}
}

// GetResourcesAsJSON returns the collected resources as JSON
func (g *DrillGame) GetResourcesAsJSON() string {
	data, _ := json.Marshal(g.session.Resources)
	return string(data)
}

// ParseResourcesFromJSON parses collected resources from JSON
func ParseResourcesFromJSON(jsonStr string) ([]DrillResource, error) {
	var resources []DrillResource
	if err := json.Unmarshal([]byte(jsonStr), &resources); err != nil {
		return nil, err
	}
	return resources, nil
}
