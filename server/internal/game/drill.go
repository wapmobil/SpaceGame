package game

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
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
	ResourceOil:      {Type: ResourceOil, Name: "Нефть", Icon: "🛢️", Value: 1, DigTime: 3.0, DepthStart: 0, DepthEnd: 1000, SpawnChance: 0.08, Damage: 3},
	ResourceGas:      {Type: ResourceGas, Name: "Газ", Icon: "💨", Value: 2, DigTime: 3.5, DepthStart: 0, DepthEnd: 1000, SpawnChance: 0.03, Damage: 3},
	ResourceCopper:   {Type: ResourceCopper, Name: "Медь", Icon: "🟠", Value: 10, DigTime: 4.0, DepthStart: 50, DepthEnd: 100, SpawnChance: 0.07, Damage: 5},
	ResourceCoal:     {Type: ResourceCoal, Name: "Уголь", Icon: "⬛", Value: 5, DigTime: 3.5, DepthStart: 50, DepthEnd: 150, SpawnChance: 0.08, Damage: 5},
	ResourceSilver:   {Type: ResourceSilver, Name: "Серебро", Icon: "⚪", Value: 15, DigTime: 5.0, DepthStart: 100, DepthEnd: 200, SpawnChance: 0.05, Damage: 8},
	ResourceGold:     {Type: ResourceGold, Name: "Золото", Icon: "🟡", Value: 25, DigTime: 6.0, DepthStart: 150, DepthEnd: 300, SpawnChance: 0.04, Damage: 10},
	ResourcePlatinum: {Type: ResourcePlatinum, Name: "Платина", Icon: "🔘", Value: 30, DigTime: 7.0, DepthStart: 200, DepthEnd: 400, SpawnChance: 0.03, Damage: 12},
	ResourceDiamond:  {Type: ResourceDiamond, Name: "Алмазы", Icon: "💎", Value: 60, DigTime: 8.0, DepthStart: 300, DepthEnd: 500, SpawnChance: 0.02, Damage: 15},
	ResourceExotic:   {Type: ResourceExotic, Name: "Экзотика", Icon: "🔮", Value: 200, DigTime: 10.0, DepthStart: 500, DepthEnd: 9999, SpawnChance: 0.01, Damage: 20},
}

// DrillCommand represents a player command
type DrillCommand struct {
	Direction string // "left", "right", "" (no horizontal move)
}

// Cell represents a single cell in the drill world
type Cell struct {
	X              int     `json:"x"`
	Y              int     `json:"y"`
	CellType       string  `json:"cell_type"`
	ResourceType   string  `json:"resource_type,omitempty"`
	ResourceAmount float64 `json:"resource_amount,omitempty"`
	ResourceValue  float64 `json:"resource_value,omitempty"`
	Extracted      bool    `json:"extracted"`
}

// DrillSession represents the current state of a drill session
type DrillSession struct {
	ID             string            `json:"id"`
	SessionID      string            `json:"session_id"`
	PlanetID       string            `json:"planet_id"`
	PlayerID       string            `json:"player_id"`
	DrillHP        int               `json:"drill_hp"`
	DrillMaxHP     int               `json:"drill_max_hp"`
	Depth          int               `json:"depth"`
	DrillX         int               `json:"drill_x"`
	WorldWidth     int               `json:"world_width"`
	Resources      []DrillResource   `json:"resources"`
	Status         string            `json:"status"`
	TotalEarned    float64           `json:"total_earned"`
	World          [][]Cell          `json:"world"`
	ViewTop        int               `json:"-"`
	ViewHeight     int               `json:"-"`
	CreatedAt      time.Time         `json:"created_at"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
	LastMoveTime   time.Time         `json:"last_move_time"`
	ExtractedCells map[string]float64 `json:"-"` // tracks (x,y) -> remaining resource amount
	PendingDirection string           `json:"-"` // last direction from client, reset after apply
	PendingExtract   bool             `json:"-"` // extract flag, persists until explicitly disabled
}

// DrillResource represents a collected resource in the session
type DrillResource struct {
	Type    string  `json:"type"`
	Name    string  `json:"name"`
	Icon    string  `json:"icon"`
	Amount  float64 `json:"amount"`
	Value   float64 `json:"value"`
}

// MoveResult represents the result of a drill move action
type MoveResult struct {
	Success       bool            `json:"success"`
	Message       string          `json:"message,omitempty"`
	DrillHP       int             `json:"drill_hp"`
	DrillMaxHP    int             `json:"drill_max_hp"`
	Depth         int             `json:"depth"`
	DrillX        int             `json:"drill_x"`
	Resources     []DrillResource `json:"resources"`
	TotalEarned   float64         `json:"total_earned"`
	GameEnded     bool            `json:"game_ended"`
	EndReason     string          `json:"end_reason,omitempty"`
	NewResource   *ResourceHit    `json:"new_resource,omitempty"`
	Extracted     float64         `json:"extracted,omitempty"`
	World         [][]Cell        `json:"world,omitempty"`
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
	Seed int64
}

// DrillGame is the main engine for the drill mini-game
type DrillGame struct {
	config      DrillConfig
	session     DrillSession
	mineLevel   int
	broadcastFn func(*MoveResult)
	cellsCache  map[string]Cell
	done        chan struct{}
}

// activeSessions stores all active drill sessions in memory
var (
	activeSessionsMu sync.RWMutex
	activeSessions   = make(map[string]*DrillGame)
)

// autoDescentInterval is how often the drill descends automatically
const autoDescentInterval = 1 * time.Second

// ActiveSessions returns all active drill sessions
func ActiveSessions() map[string]*DrillGame {
	activeSessionsMu.RLock()
	defer activeSessionsMu.RUnlock()
	result := make(map[string]*DrillGame, len(activeSessions))
	for k, v := range activeSessions {
		result[k] = v
	}
	return result
}

// FindActiveSession finds an active or ended drill session by planet and player ID
func FindActiveSession(planetID, playerID string) *DrillGame {
	activeSessionsMu.RLock()
	defer activeSessionsMu.RUnlock()
	for _, dg := range activeSessions {
		s := dg.GetSession()
		if s.PlanetID == planetID && s.PlayerID == playerID && (s.Status == "active" || s.Status == "failed") {
			return dg
		}
	}
	return nil
}

// Default drill configuration
const (
	DefaultWorldWidth = 5
	DefaultViewHeight = 5
	DefaultSeed       = 42
)

// NewDrillGame creates a new drill game session
func NewDrillGame(planetID, playerID string, mineLevel int) *DrillGame {
	maxHP := 10 + 100*mineLevel
	session := DrillSession{
		ID:             uuid.New().String(),
		SessionID:      fmt.Sprintf("drill-%d", time.Now().UnixNano()),
		PlanetID:       planetID,
		PlayerID:       playerID,
		DrillHP:        maxHP,
		DrillMaxHP:     maxHP,
		Depth:          0,
		DrillX:         DefaultWorldWidth / 2,
		WorldWidth:     DefaultWorldWidth,
		Resources:      []DrillResource{},
		Status:         "active",
		TotalEarned:    0,
		ViewHeight:     DefaultViewHeight,
		CreatedAt:      time.Now(),
		ExtractedCells: make(map[string]float64),
	}

	config := DrillConfig{
		Seed: rand.Int63(),
	}

	game := &DrillGame{
		config:     config,
		session:    session,
		mineLevel:  mineLevel,
		cellsCache: make(map[string]Cell),
		done:       make(chan struct{}),
	}

	game.generateInitialWorld()
	activeSessionsMu.Lock()
	activeSessions[session.SessionID] = game
	activeSessionsMu.Unlock()

	go game.autoDescentTicker()

	return game
}

func (g *DrillGame) autoDescentTicker() {
	ticker := time.NewTicker(autoDescentInterval)
	defer ticker.Stop()
	for {
		select {
		case <-g.done:
			return
		case <-ticker.C:
			if g.session.Status != "active" {
				return
			}
			g.ApplyCommandWithBroadcast()
		}
	}
}

// SetBroadcastFn sets the callback for broadcasting drill updates
func (g *DrillGame) SetBroadcastFn(fn func(*MoveResult)) {
	g.broadcastFn = fn
}

// SetCommand memorizes a command from the client without applying it immediately
func (g *DrillGame) SetCommand(direction string, extract *bool) {
	g.session.PendingDirection = direction
	if extract != nil {
		g.session.PendingExtract = *extract
	}
}

// ApplyCommand applies the pending command and returns the result
func (g *DrillGame) ApplyCommand() *MoveResult {
	// Save extract flag before resetting direction
	extract := g.session.PendingExtract
	direction := g.session.PendingDirection
	g.session.PendingDirection = "" // reset direction only

	result := &MoveResult{
		DrillHP:     g.session.DrillHP,
		DrillMaxHP:  g.session.DrillMaxHP,
		Depth:       g.session.Depth,
		DrillX:      g.session.DrillX,
		Resources:   g.session.Resources,
		TotalEarned: g.session.TotalEarned,
	}

	if g.session.Status != "active" {
		result.Success = false
		result.Message = "Drill session is not active"
		result.GameEnded = true
		result.EndReason = "session_ended"
		return result
	}

	// 1. Horizontal movement
	if direction == "left" {
		g.session.DrillX--
		result.DrillX = g.session.DrillX
	} else if direction == "right" {
		g.session.DrillX++
		result.DrillX = g.session.DrillX
	}

	// 2. Always move down on auto-descent
	g.processDrillDown(result, extract)

	// 3. Extract from the NEW cell (after moving down)
	g.processExtraction(result, extract)

	// 4. Regenerate world at new position
	g.regenerateWorld()
	result.World = g.session.World

	// 5. Check if drill is destroyed
	if g.session.DrillHP <= 0 {
		g.session.DrillHP = 0
		g.session.Status = "failed"
		now := time.Now()
		g.session.CompletedAt = &now
		g.convertResourcesToMoney()
		result.TotalEarned = g.session.TotalEarned
		result.GameEnded = true
		result.EndReason = "drill_destroyed"
		return result
	}

	result.Success = true
	return result
}

// ApplyCommandWithBroadcast applies the pending command and broadcasts via callback
func (g *DrillGame) ApplyCommandWithBroadcast() *MoveResult {
	result := g.ApplyCommand()
	if g.broadcastFn != nil {
		g.broadcastFn(result)
	}
	return result
}

// generateInitialWorld creates the initial 5x5 world centered at depth 0
func (g *DrillGame) generateInitialWorld() {
	g.session.World = g.buildWorldAt(g.session.Depth, g.session.DrillX)
}

// buildWorldAt builds a 5x5 world with drill at top center
func (g *DrillGame) buildWorldAt(depth, drillX int) [][]Cell {
	world := make([][]Cell, DefaultWorldWidth)
	for dy := 0; dy < DefaultWorldWidth; dy++ {
		rowIdx := dy
		world[rowIdx] = make([]Cell, DefaultWorldWidth)
		for dx := -2; dx <= 2; dx++ {
			colIdx := dx + 2
			x := drillX + dx
			y := depth + dy
			cell := g.getCellWithState(x, y)
			world[rowIdx][colIdx] = cell
		}
	}
	return world
}

// GetChunk generates a chunk of the world centered at given coordinates
func (g *DrillGame) GetChunk(centerX, centerY, width, height int) [][]Cell {
	chunk := make([][]Cell, height)
	for dy := 0; dy < height; dy++ {
		chunk[dy] = make([]Cell, width)
		for dx := 0; dx < width; dx++ {
			x := centerX - width/2 + dx
			y := centerY - height/2 + dy
			chunk[dy][dx] = g.getCellWithState(x, y)
		}
	}
	return chunk
}

// getCellWithState returns a cell with tracked extraction state
func (g *DrillGame) getCellWithState(x, y int) Cell {
	cell := g.getCellAt(x, y)
	cellKey := fmt.Sprintf("%d,%d", x, y)
	if remaining, ok := g.session.ExtractedCells[cellKey]; ok {
		cell.Extracted = remaining <= 0
		cell.ResourceAmount = remaining
		if cell.Extracted {
			cell.ResourceType = ""
			cell.ResourceValue = 0
		}
	}
	return cell
}

// getCellAt deterministically generates a cell based on seed and coordinates
func (g *DrillGame) getCellAt(x, y int) Cell {
	key := fmt.Sprintf("%d,%d", x, y)
	if cell, ok := g.cellsCache[key]; ok {
		return cell
	}

	cell := Cell{X: x, Y: y, CellType: CellDirt}

	// Create a deterministic RNG for cell type based on coordinates and seed
	typeSeed := g.config.Seed + int64(x)*1000003 + int64(y)*1000033
	r := rand.New(rand.NewSource(typeSeed))

	depthFactor := float64(y) / 800.0
	if depthFactor > 1.0 {
		depthFactor = 1.0
	}

	// Determine cell type based on depth
	rVal := r.Float64()
	if rVal < 0.3+depthFactor*0.15 {
		cell.CellType = CellDirt
	} else if rVal < 0.55+depthFactor*0.15 {
		cell.CellType = CellStone
	} else if rVal < 0.75+depthFactor*0.15 {
		cell.CellType = CellMetal
	} else {
		cell.CellType = CellMithril
	}

	// Create a separate deterministic RNG for resource spawning
	resSeed := typeSeed + 5000000000
	resRng := rand.New(rand.NewSource(resSeed))

	// Try to spawn a resource
	if resRng.Float64() < 0.12 {
		resource := g.selectResourceForDepth(depthFactor, resSeed)
		if resource != nil {
			cell.ResourceType = resource.Type
			amountRng := rand.New(rand.NewSource(typeSeed + 1))
			cell.ResourceAmount = float64(amountRng.Intn(5)+3)
			cell.ResourceValue = resource.Value * cell.ResourceAmount
		}
	}

	g.cellsCache[key] = cell
	return cell
}

// selectResourceForDepth picks a random resource based on depth using a deterministic seed
func (g *DrillGame) selectResourceForDepth(depthFactor float64, seed int64) *ResourceDef {
	mixed := uint64(seed)
	mixed ^= mixed >> 30
	mixed *= 0xbf58476d1ce4e5b9
	mixed ^= mixed >> 27
	mixed *= 0x94d049bb133111eb
	mixed ^= mixed >> 31
	r := rand.New(rand.NewSource(int64(mixed)))
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

// regenerateWorld rebuilds the 5x5 world centered on current drill position
func (g *DrillGame) regenerateWorld() {
	g.session.World = g.buildWorldAt(g.session.Depth, g.session.DrillX)
	g.cleanupCache()
}

// cleanupCache removes cells from the cache that are far above the drill
func (g *DrillGame) cleanupCache() {
	if len(g.cellsCache) == 0 {
		return
	}
	cutoff := g.session.Depth - 20
	for key := range g.cellsCache {
		var x, y int
		fmt.Sscanf(key, "%d,%d", &x, &y)
		if y < cutoff {
			delete(g.cellsCache, key)
		}
	}
}

// processDrillDown moves the drill downward and handles resource damage
func (g *DrillGame) processDrillDown(result *MoveResult, extract bool) {
	newDepth := g.session.Depth + 1

	// Check cell at new depth, same X
	newCell := g.getCellAt(g.session.DrillX, newDepth)

	// Apply damage from cell type
	damage := g.getCellDamage(newCell.CellType)
	g.session.DrillHP -= damage

	// Extra damage: resource cell but extraction not active
	if newCell.ResourceType != "" && newCell.ResourceAmount > 0 && !extract {
		g.session.DrillHP -= 5
	}

	// Notify about resource at new depth (but don't extract)
	if newCell.ResourceType != "" && newCell.ResourceAmount > 0 {
		def, ok := resourceDefinitions[newCell.ResourceType]
		if ok {
			result.NewResource = &ResourceHit{
				Type:   def.Type,
				Name:   def.Name,
				Icon:   def.Icon,
				Amount: newCell.ResourceAmount,
				Value:  newCell.ResourceValue,
			}
		}
	}

	g.session.Depth = newDepth
}

// processExtraction handles resource extraction from current cell
func (g *DrillGame) processExtraction(result *MoveResult, extract bool) {
	cellKey := fmt.Sprintf("%d,%d", g.session.DrillX, g.session.Depth)
	currentCell := g.getCellWithState(g.session.DrillX, g.session.Depth)

	if currentCell.ResourceType == "" || currentCell.Extracted {
		// Extra damage: extraction active but no resource on cell
		if extract {
			g.session.DrillHP -= 3
		}
		return
	}

	// Only extract if extract flag is true
	if !extract {
		return
	}

	def, ok := resourceDefinitions[currentCell.ResourceType]
	if !ok {
		return
	}

	extractRate := currentCell.ResourceAmount / def.DigTime * 0.5
	if extractRate > currentCell.ResourceAmount {
		extractRate = currentCell.ResourceAmount
	}

	remaining := currentCell.ResourceAmount - extractRate
	g.session.ExtractedCells[cellKey] = math.Max(0, remaining)

	g.addResource(def, extractRate, def.Value*extractRate)
	result.Extracted = extractRate
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
	return g.session.World
}

// GetAvailableDirections returns the available movement directions
func (g *DrillGame) GetAvailableDirections() []MoveDirection {
	var dirs []MoveDirection
	dirs = append(dirs, MoveLeft)
	dirs = append(dirs, MoveRight)
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

// GetSeed returns the session seed
func (g *DrillGame) GetSeed() int64 {
	return g.config.Seed
}

// GetDrillCooldown returns the cooldown between drill sessions
func GetDrillCooldown() time.Duration {
	return 30 * time.Second
}

// Destroy sets drill HP to 0 triggering game over
func (g *DrillGame) Destroy() {
	if g.session.Status != "active" {
		return
	}
	g.session.DrillHP = 0
	g.session.Status = "failed"
	now := time.Now()
	g.session.CompletedAt = &now
	g.convertResourcesToMoney()
	activeSessionsMu.Lock()
	delete(activeSessions, g.session.SessionID)
	activeSessionsMu.Unlock()
	close(g.done)
}

// Complete marks the session as completed and converts resources to money
func (g *DrillGame) Complete() float64 {
	if g.session.Status != "active" {
		return 0
	}
	g.session.Status = "completed"
	now := time.Now()
	g.session.CompletedAt = &now
	g.convertResourcesToMoney()
	totalEarned := g.session.TotalEarned
	activeSessionsMu.Lock()
	delete(activeSessions, g.session.SessionID)
	activeSessionsMu.Unlock()
	close(g.done)
	return totalEarned
}
