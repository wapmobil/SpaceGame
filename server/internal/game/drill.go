package game

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DrillCommand represents a player command
type DrillCommand struct {
	Direction string // "left", "right", "" (no horizontal move)
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

// MoveDirection represents horizontal movement
type MoveDirection int

const (
	MoveLeft MoveDirection = iota
	MoveRight
	MoveDown
)

// Default drill configuration
const (
	DefaultWorldWidth = 5
	DefaultViewHeight = 5
	DefaultSeed       = 42
)

// Speed intervals for drill auto-descent
const (
	speed1xInterval = 1 * time.Second
	speed2xInterval = 500 * time.Millisecond
)

// activeSessions stores all active drill sessions in memory
var (
	activeSessionsMu sync.RWMutex
	activeSessions   = make(map[string]*DrillGame)
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

// DrillGame is the core drill mini-game engine
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
