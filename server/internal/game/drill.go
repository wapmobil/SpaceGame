package game

import (
	"fmt"
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

// DrillCommand represents a player command
type DrillCommand struct {
	Direction string // "left", "right", "" (no horizontal move)
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
	Status         string          `json:"status"`
	TotalEarned    int             `json:"total_earned"`
	World          [][]Cell          `json:"world"`
	ViewTop        int               `json:"-"`
	ViewHeight     int               `json:"-"`
	CreatedAt      time.Time         `json:"created_at"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
	LastMoveTime   time.Time         `json:"last_move_time"`
	ExtractedCells map[string]bool `json:"-"` // tracks (x,y) -> extracted
	PendingDirection string           `json:"-"` // last direction from client, reset after apply
	PendingExtract   bool             `json:"-"` // extract flag, persists until explicitly disabled
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

// activeSessions stores all active drill sessions in memory
var (
	activeSessionsMu sync.RWMutex
	activeSessions   = make(map[string]*DrillGame)
)

// Speed intervals for drill auto-descent
const (
	speed1xInterval = 1 * time.Second
	speed2xInterval = 500 * time.Millisecond
)

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

// RemoveSession removes a drill session from the active sessions map
func RemoveSession(sessionID string) {
	activeSessionsMu.Lock()
	defer activeSessionsMu.Unlock()
	delete(activeSessions, sessionID)
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

// Depth zones for geological layer system
type DepthZone struct {
	Name    string
	MinDepth int
	MaxDepth int
	Dirt     float64
	Stone    float64
	Metal    float64
	Mithril  float64
}

var depthZones = []DepthZone{
	{Name: "surface", MinDepth: 0, MaxDepth: 64, Dirt: 0.60, Stone: 0.25, Metal: 0.12, Mithril: 0.03},
	{Name: "shallow", MinDepth: 64, MaxDepth: 128, Dirt: 0.40, Stone: 0.35, Metal: 0.20, Mithril: 0.05},
	{Name: "forest", MinDepth: 128, MaxDepth: 192, Dirt: 0.20, Stone: 0.45, Metal: 0.30, Mithril: 0.05},
	{Name: "mid", MinDepth: 192, MaxDepth: 256, Dirt: 0.10, Stone: 0.35, Metal: 0.40, Mithril: 0.15},
	{Name: "deep", MinDepth: 256, MaxDepth: 320, Dirt: 0.02, Stone: 0.20, Metal: 0.45, Mithril: 0.33},
	{Name: "abyss", MinDepth: 320, MaxDepth: 9999, Dirt: 0.00, Stone: 0.05, Metal: 0.25, Mithril: 0.70},
}

const (
	noiseGridSize = 16
	noiseScale    = float64(noiseGridSize)
	caveThreshold = 0.85
)

type DrillGame struct {
	config       DrillConfig
	session      DrillSession
	mineLevel    int
	tickInterval time.Duration
	broadcastFn  func(*MoveResult)
	cellsCache   map[string]Cell
	done         chan struct{}

	terrainNoiseGrid [noiseGridSize][noiseGridSize]float64
	caveNoiseGrid    [noiseGridSize][noiseGridSize]float64
	veinNoiseGrid    [noiseGridSize][noiseGridSize]float64
}

// NewDrillGame creates a new drill game session
func NewDrillGame(planetID, playerID string, mineLevel, tickInterval int) *DrillGame {
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
	}

	config := DrillConfig{
		Seed: rand.Int63(),
	}

	var interval time.Duration
	switch tickInterval {
	case 2:
		interval = speed2xInterval
	default:
		interval = speed1xInterval
	}

	game := &DrillGame{
		config:       config,
		session:      session,
		mineLevel:    mineLevel,
		tickInterval: interval,
		cellsCache:   make(map[string]Cell),
		done:         make(chan struct{}),
	}

	game.initNoiseGrids()
	game.generateInitialWorld()
	activeSessionsMu.Lock()
	activeSessions[session.SessionID] = game
	activeSessionsMu.Unlock()

	go game.autoDescentTicker()

	return game
}

func (g *DrillGame) autoDescentTicker() {
	ticker := time.NewTicker(g.tickInterval)
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

func (g *DrillGame) initNoiseGrids() {
	r := rand.New(rand.NewSource(g.config.Seed))

	for x := 0; x < noiseGridSize; x++ {
		for y := 0; y < noiseGridSize; y++ {
			g.terrainNoiseGrid[x][y] = r.Float64()
			g.caveNoiseGrid[x][y] = r.Float64()
			g.veinNoiseGrid[x][y] = r.Float64()
		}
	}

	g.applyCaveSmoothing()
}

func (g *DrillGame) applyCaveSmoothing() {
	newGrid := make([][]float64, noiseGridSize)
	for i := range newGrid {
		newGrid[i] = make([]float64, noiseGridSize)
	}

	for x := 0; x < noiseGridSize; x++ {
		for y := 0; y < noiseGridSize; y++ {
			caveCount := 0
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					nx := (x + dx + noiseGridSize) % noiseGridSize
					ny := (y + dy + noiseGridSize) % noiseGridSize
					if g.caveNoiseGrid[nx][ny] > caveThreshold {
						caveCount++
					}
				}
			}
			if g.caveNoiseGrid[x][y] > caveThreshold {
				if caveCount >= 5 {
					newGrid[x][y] = 1.0
				} else {
					newGrid[x][y] = 0.0
				}
			} else {
				if caveCount >= 6 {
					newGrid[x][y] = 1.0
				} else {
					newGrid[x][y] = 0.0
				}
			}
		}
	}

	for x := 0; x < noiseGridSize; x++ {
		for y := 0; y < noiseGridSize; y++ {
			g.caveNoiseGrid[x][y] = newGrid[x][y]
		}
	}
}

func (g *DrillGame) sampleNoise(grid [noiseGridSize][noiseGridSize]float64, fx, fy float64) float64 {
	gx0 := ((int(fx) % noiseGridSize) + noiseGridSize) % noiseGridSize
	gy0 := ((int(fy) % noiseGridSize) + noiseGridSize) % noiseGridSize
	gx1 := (gx0 + 1) % noiseGridSize
	gy1 := (gy0 + 1) % noiseGridSize

	fx0 := fx - float64(int(fx))
	fy0 := fy - float64(int(fy))

	sx := fx0 * fx0 * (3 - 2*fx0)
	sy := fy0 * fy0 * (3 - 2*fy0)

	n00 := grid[gx0][gy0]
	n10 := grid[gx1][gy0]
	n01 := grid[gx0][gy1]
	n11 := grid[gx1][gy1]

	x1 := n00 + (n10-n00)*sx
	x2 := n01 + (n11-n01)*sx

	return x1 + (x2-x1)*sy
}

func (g *DrillGame) getTerrainNoise(x, y int) float64 {
	fx := float64(x) / noiseScale
	fy := float64(y) / noiseScale
	return g.sampleNoise(g.terrainNoiseGrid, fx, fy)
}

func (g *DrillGame) getCaveNoise(x, y int) float64 {
	fx := float64(x) / noiseScale
	fy := float64(y) / noiseScale
	return g.sampleNoise(g.caveNoiseGrid, fx, fy)
}

func (g *DrillGame) getVeinNoise(x, y int) float64 {
	fx := float64(x) / noiseScale
	fy := float64(y) / noiseScale
	return g.sampleNoise(g.veinNoiseGrid, fx, fy)
}

func (g *DrillGame) getDepthZone(depth int) *DepthZone {
	for i := range depthZones {
		z := &depthZones[i]
		if depth >= z.MinDepth && depth < z.MaxDepth {
			return z
		}
	}
	return &depthZones[len(depthZones)-1]
}

func (g *DrillGame) selectCellType(zone *DepthZone, noise float64, depth int) string {
	noiseBias := (noise - 0.5) * 0.35
	dirt := zone.Dirt + noiseBias
	stone := zone.Stone + noiseBias*0.5
	metal := zone.Metal + noiseBias*0.3
	mithril := zone.Mithril - noiseBias*0.5

	if mithril < 0 {
		metal += mithril
		mithril = 0
	}
	if metal < 0 {
		stone += metal
		metal = 0
	}
	if stone < 0 {
		dirt += stone
		stone = 0
	}
	if dirt < 0 {
		dirt = 0
	}

	rVal := noise
	if rVal < dirt {
		return CellDirt
	} else if rVal < dirt+stone {
		return CellStone
	} else if rVal < dirt+stone+metal {
		return CellMetal
	}
	return CellMithril
}

func (g *DrillGame) isCave(x, y int) bool {
	caveNoise := g.getCaveNoise(x, y)
	return caveNoise > caveThreshold
}

func (g *DrillGame) getVeinMultiplier(x, y int) float64 {
	veinNoise := g.getVeinNoise(x, y)
	if veinNoise > 0.7 {
		return 1.8
	} else if veinNoise > 0.5 {
		return 1.3
	} else if veinNoise > 0.3 {
		return 1.0
	}
	return 0.5
}

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
	if extracted, ok := g.session.ExtractedCells[cellKey]; ok && extracted {
		cell.Extracted = true
		cell.ResourceType = ""
		cell.ResourceValue = 0
	}
	return cell
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

	if g.session.ExtractedCells == nil {
		g.session.ExtractedCells = make(map[string]bool)
	}
	g.session.ExtractedCells[cellKey] = true
	g.addResource(def, 1, def.Value)
	result.Extracted = 1
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

// GetTickInterval returns the tick interval duration for this drill game
func (g *DrillGame) GetTickInterval() time.Duration {
	return g.tickInterval
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
func (g *DrillGame) Complete() int {
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
