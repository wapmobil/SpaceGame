package game

import (
	"fmt"
	"math/rand"
	"time"
)

// Mining cell types
const (
	CellWall   = '#'
	CellEmpty  = '.'
	CellPlayer = 'P'
	CellExit   = 'E'
	CellMonster = 'M'
	CellHeart   = 'H'
	CellBomb    = 'B'
	CellMoney   = '$'
)

// Monster types
const (
	MonsterRat    = "rat"
	MonsterBat    = "bat"
	MonsterAlien  = "alien"
)

// Monster definitions
var monsterDefinitions = map[string]MonsterDef{
	MonsterRat: {
		Name:     "Rat",
		Icon:     "🐀",
		HP:       15,
		Damage:   8,
		Reward:   15,
		SpawnRate: 3,
	},
	MonsterBat: {
		Name:     "Bat",
		Icon:     "🦇",
		HP:       25,
		Damage:   12,
		Reward:   25,
		SpawnRate: 2,
	},
	MonsterAlien: {
		Name:     "Alien",
		Icon:     "👽",
		HP:       50,
		Damage:   20,
		Reward:   50,
		SpawnRate: 1,
	},
}

// MonsterDef defines a monster type
type MonsterDef struct {
	Name      string
	Icon      string
	HP        int
	Damage    int
	Reward    int
	SpawnRate int
}

// MiningSession represents an active mining session
type MiningSession struct {
	ID               string    `json:"id"`
	PlanetID         string    `json:"planet_id"`
	PlayerID         string    `json:"player_id"`
	SessionID        string    `json:"session_id"`
	Maze             [][]rune  `json:"maze"`
	DisplayMaze      [][]rune  `json:"display_maze"`
	PlayerX          int       `json:"player_x"`
	PlayerY          int       `json:"player_y"`
	PlayerHP         int       `json:"player_hp"`
	PlayerMaxHP      int       `json:"player_max_hp"`
	PlayerBombs      int       `json:"player_bombs"`
	MoneyCollected   float64   `json:"money_collected"`
	Status           string    `json:"status"` // "active", "completed", "failed"
	ExitX            int       `json:"exit_x"`
	ExitY            int       `json:"exit_y"`
	Monsters         []Monster `json:"monsters"`
	LastMoveTime     time.Time `json:"last_move_time"`
	CompletedAt      time.Time `json:"completed_at,omitempty"`
	BaseLevel        int       `json:"base_level"`
	StartTime        time.Time `json:"start_time"`
}

// Monster represents a monster entity in the maze
type Monster struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	Damage int    `json:"damage"`
	Reward float64 `json:"reward"`
	Alive  bool   `json:"alive"`
}

// MazeConfig holds configuration for maze generation
type MazeConfig struct {
	Size         int
	Seed         int64
	HeartChance  float64
	BombChance   float64
	MoneyChance  float64
	MonsterSpawn []MonsterSpawn
}

// MonsterSpawn defines how a monster type spawns
type MonsterSpawn struct {
	Type        string
	Chance      float64
	MinDistance int
}

// DefaultMazeConfig returns default maze configuration
func DefaultMazeConfig(baseLevel int) MazeConfig {
	return MazeConfig{
		Size:        13,
		Seed:        time.Now().UnixNano(),
		HeartChance: 0.05,
		BombChance:  0.03,
		MoneyChance: 0.15,
		MonsterSpawn: []MonsterSpawn{
			{Type: MonsterRat, Chance: 0.40, MinDistance: 3},
			{Type: MonsterBat, Chance: 0.25, MinDistance: 4},
			{Type: MonsterAlien, Chance: 0.10, MinDistance: 5},
		},
	}
}

// MiningGame is the core mining game engine
type MiningGame struct {
	config   MazeConfig
	session  *MiningSession
	rng      *rand.Rand
}

// NewMiningGame creates a new mining game instance
func NewMiningGame(planetID, playerID string, baseLevel int) *MiningGame {
	config := DefaultMazeConfig(baseLevel)
	config.Seed = time.Now().UnixNano()

	mg := &MiningGame{
		config: config,
		rng:    rand.New(rand.NewSource(config.Seed)),
		session: &MiningSession{
			PlanetID:    planetID,
			PlayerID:    playerID,
			SessionID:   fmt.Sprintf("mining_%s_%d", planetID, time.Now().UnixNano()),
			Status:      "active",
			BaseLevel:   baseLevel,
			StartTime:   time.Now(),
		},
	}

	mg.generateMaze()
	mg.placePlayer()
	mg.placeExit()
	mg.spawnEntities()

	return mg
}

// NewMiningGameWithSeed creates a mining game with a fixed seed (for testing)
func NewMiningGameWithSeed(planetID, playerID string, baseLevel int, seed int64) *MiningGame {
	config := DefaultMazeConfig(baseLevel)
	config.Seed = seed

	mg := &MiningGame{
		config: config,
		rng:    rand.New(rand.NewSource(seed)),
		session: &MiningSession{
			PlanetID:    planetID,
			PlayerID:    playerID,
			SessionID:   fmt.Sprintf("mining_%s_%d", planetID, seed),
			Status:      "active",
			BaseLevel:   baseLevel,
			StartTime:   time.Now(),
		},
	}

	mg.generateMaze()
	mg.placePlayer()
	mg.placeExit()
	mg.spawnEntities()

	return mg
}

// generateMaze creates the maze using a tracer/recursive backtracking algorithm
func (mg *MiningGame) generateMaze() {
	size := mg.config.Size
	mg.session.Maze = make([][]rune, size)
	mg.session.DisplayMaze = make([][]rune, size)

	for i := 0; i < size; i++ {
		mg.session.Maze[i] = make([]rune, size)
		mg.session.DisplayMaze[i] = make([]rune, size)
		for j := 0; j < size; j++ {
			// Initialize: odd positions are potential passages, even are walls
			if i%2 == 0 || j%2 == 0 {
				mg.session.Maze[i][j] = CellWall
				mg.session.DisplayMaze[i][j] = CellWall
			} else {
				mg.session.Maze[i][j] = CellEmpty
				mg.session.DisplayMaze[i][j] = CellEmpty
			}
		}
	}

	// Use tracer algorithm (iterative depth-first search) to carve passages
	mg.tracerMaze(1, 1)

	// Ensure border is all walls
	for i := 0; i < size; i++ {
		mg.session.Maze[i][0] = CellWall
		mg.session.Maze[i][size-1] = CellWall
		mg.session.Maze[0][i] = CellWall
		mg.session.Maze[size-1][i] = CellWall
		mg.session.DisplayMaze[i][0] = CellWall
		mg.session.DisplayMaze[i][size-1] = CellWall
		mg.session.DisplayMaze[0][i] = CellWall
		mg.session.DisplayMaze[size-1][i] = CellWall
	}

	// Place exit at bottom-right corner (far from start)
	mg.session.ExitX = size - 2
	mg.session.ExitY = size - 2
	mg.session.Maze[size-2][size-2] = CellEmpty
	mg.session.DisplayMaze[size-2][size-2] = CellExit
}

// tracerMaze uses an iterative depth-first search (tracer) algorithm to carve the maze
func (mg *MiningGame) tracerMaze(startX, startY int) {
	size := mg.config.Size
	visited := make([][]bool, size)
	stack := [][2]int{}

	for i := 0; i < size; i++ {
		visited[i] = make([]bool, size)
	}

	// Start from the given position (always odd for passage centers)
	mg.session.Maze[startX][startY] = CellEmpty
	mg.session.DisplayMaze[startX][startY] = CellEmpty
	visited[startX][startY] = true
	stack = append(stack, [2]int{startX, startY})

	// Direction vectors (up, down, left, right) - move by 2 cells
	directions := [][2]int{
		{-2, 0},
		{2, 0},
		{0, -2},
		{0, 2},
	}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		cx, cy := current[0], current[1]

		// Find unvisited neighbors
		var found bool
		for _, dir := range directions {
			nx, ny := cx+dir[0], cy+dir[1]

			if nx >= 1 && nx < size-1 && ny >= 1 && ny < size-1 && !visited[nx][ny] {
				// Check the cell between current and neighbor
				midX, midY := cx+dir[0]/2, cy+dir[1]/2
				if mg.session.Maze[midX][midY] == CellWall {
					// Carve passage
					mg.session.Maze[midX][midY] = CellEmpty
					mg.session.DisplayMaze[midX][midY] = CellEmpty
					mg.session.Maze[nx][ny] = CellEmpty
					mg.session.DisplayMaze[nx][ny] = CellEmpty
					visited[nx][ny] = true
					stack = append(stack, [2]int{nx, ny})
					found = true
					break
				}
			}
		}

		if !found {
			stack = stack[:len(stack)-1]
		}
	}
}

// placePlayer sets the player's starting position at (1, 1)
func (mg *MiningGame) placePlayer() {
	mg.session.PlayerX = 1
	mg.session.PlayerY = 1
	mg.session.PlayerHP = 100
	mg.session.PlayerMaxHP = 100
	mg.session.PlayerBombs = 1
	mg.session.Maze[1][1] = CellPlayer
	mg.session.DisplayMaze[1][1] = CellPlayer
}

// placeExit places the exit at a random empty cell far from the start
func (mg *MiningGame) placeExit() {
	size := mg.config.Size

	// Try to place exit near bottom-right corner
	bestX, bestY := size-2, size-2
	if mg.session.Maze[bestX][bestY] == CellWall {
		// Find nearest empty cell
		for r := size - 3; r >= 1; r-- {
			for c := size - 3; c >= 1; c-- {
				if mg.session.Maze[r][c] == CellEmpty {
					bestX, bestY = r, c
					break
				}
			}
			if mg.session.Maze[bestX][bestY] == CellEmpty {
				break
			}
		}
	}

	mg.session.Maze[bestX][bestY] = CellEmpty
	mg.session.DisplayMaze[bestX][bestY] = CellExit
	mg.session.ExitX = bestX
	mg.session.ExitY = bestY
}

// spawnEntities places hearts, bombs, money, and monsters on empty cells
func (mg *MiningGame) spawnEntities() {
	size := mg.config.Size
	emptyCells := []struct{ x, y int }{}

	// Collect all empty cells (not player start, not exit)
	for i := 1; i < size-1; i++ {
		for j := 1; j < size-1; j++ {
			if mg.session.Maze[i][j] == CellEmpty {
				// Skip player start and exit positions
				if (i == 1 && j == 1) || (i == mg.session.ExitX && j == mg.session.ExitY) {
					continue
				}
				emptyCells = append(emptyCells, struct{ x, y int }{i, j})
			}
		}
	}

	// Shuffle empty cells
	mg.shuffleCells(emptyCells)

	// Place entities
	for _, cell := range emptyCells {
		r := mg.rng.Float64()

		// Place monsters
		for _, spawn := range mg.config.MonsterSpawn {
			if r < float64(spawn.Chance) {
				def := monsterDefinitions[spawn.Type]
				monster := Monster{
					ID:     fmt.Sprintf("monster_%d_%d", cell.x, cell.y),
					Type:   spawn.Type,
					Name:   def.Name,
					Icon:   def.Icon,
					X:      cell.x,
					Y:      cell.y,
					HP:     def.HP,
					MaxHP:  def.HP,
					Damage: def.Damage,
					Reward: float64(def.Reward) * mg.getMoneyMultiplier(),
					Alive:  true,
				}
		mg.session.Monsters = append(mg.session.Monsters, monster)
			mg.session.Maze[cell.x][cell.y] = CellMonster
			for _, rc := range monster.Icon {
				mg.session.DisplayMaze[cell.x][cell.y] = rc
				break
			}
				break
			}
			r -= float64(spawn.Chance)
		}

		// Place heart
		if r < mg.config.HeartChance {
			mg.session.Maze[cell.x][cell.y] = CellHeart
			mg.session.DisplayMaze[cell.x][cell.y] = CellHeart
			continue
		}

		// Place bomb
		if r < mg.config.HeartChance+mg.config.BombChance {
			mg.session.Maze[cell.x][cell.y] = CellBomb
			mg.session.DisplayMaze[cell.x][cell.y] = CellBomb
			continue
		}

		// Place money
		if r < mg.config.HeartChance+mg.config.BombChance+mg.config.MoneyChance {
			moneyValue := float64(mg.rng.Intn(10) + 5)
			moneyValue *= mg.getMoneyMultiplier()
			mg.session.Maze[cell.x][cell.y] = CellMoney
			mg.session.DisplayMaze[cell.x][cell.y] = CellMoney
			// Store money value in a special way - we'll handle it on pickup
			_ = moneyValue
			continue
		}
	}
}

// shuffleCells shuffles a slice of cell coordinates
func (mg *MiningGame) shuffleCells(cells []struct{ x, y int }) {
	for i := len(cells) - 1; i > 0; i-- {
		j := mg.rng.Intn(i + 1)
		cells[i], cells[j] = cells[j], cells[i]
	}
}

// getMoneyMultiplier returns the score multiplier based on base level
func (mg *MiningGame) getMoneyMultiplier() float64 {
	return 1.0 + float64(mg.session.BaseLevel)*0.1
}

// MoveDirection represents a movement direction
type MoveDirection int

const (
	Up    MoveDirection = 0
	Down  MoveDirection = 1
	Left  MoveDirection = 2
	Right MoveDirection = 3
)

// MoveResult represents the result of a move action
type MoveResult struct {
	Success        bool     `json:"success"`
	Message        string   `json:"message,omitempty"`
	Maze           [][]rune `json:"maze,omitempty"`
	PlayerX        int      `json:"player_x"`
	PlayerY        int      `json:"player_y"`
	PlayerHP       int      `json:"player_hp"`
	PlayerBombs    int      `json:"player_bombs"`
	MoneyCollected float64  `json:"money_collected"`
	Encounter      *Encounter `json:"encounter,omitempty"`
	GameEnded      bool     `json:"game_ended"`
	EndReason      string   `json:"end_reason,omitempty"`
}

// Encounter represents a monster encounter during movement
type Encounter struct {
	MonsterID  string  `json:"monster_id"`
	MonsterName string `json:"monster_name"`
	MonsterIcon string `json:"monster_icon"`
	Damage     int     `json:"damage"`
	Reward     float64 `json:"reward"`
	Killed     bool    `json:"killed"`
	Dead       bool    `json:"dead"`
}

// Move moves the player in the given direction
func (mg *MiningGame) Move(direction MoveDirection, slide bool) *MoveResult {
	if mg.session.Status != "active" {
		return &MoveResult{
			Success:     false,
			Message:     "Game is not active",
			GameEnded:   true,
			EndReason:   mg.session.Status,
		}
	}

	dx, dy := mg.getDirectionDelta(direction)
	newX, newY := mg.session.PlayerX, mg.session.PlayerY

	if slide {
		// Slide until hitting a wall
		for {
			nextX, nextY := newX+dx, newY+dy
			if !mg.isWalkable(nextX, nextY) {
				break
			}
			newX, newY = nextX, nextY
		}
	} else {
		// Single step
		nextX, nextY := mg.session.PlayerX+dx, mg.session.PlayerY+dy
		if !mg.isWalkable(nextX, nextY) {
			return &MoveResult{
				Success:  false,
				Message:  "Cannot move in that direction",
				Maze:     mg.getDisplayMaze(),
				PlayerX:  mg.session.PlayerX,
				PlayerY:  mg.session.PlayerY,
			}
		}
		newX, newY = nextX, nextY
	}

	// Check what's at the new position
	result := mg.processCell(newX, newY)

	// Update player position
	mg.session.Maze[mg.session.PlayerX][mg.session.PlayerY] = CellEmpty
	mg.session.PlayerX, mg.session.PlayerY = newX, newY
	mg.session.Maze[newX][newY] = CellPlayer

	mg.session.LastMoveTime = time.Now()

	// Update display maze
	displayMaze := mg.getDisplayMaze()

	return &MoveResult{
		Success:      true,
		Maze:         displayMaze,
		PlayerX:      newX,
		PlayerY:      newY,
		PlayerHP:     mg.session.PlayerHP,
		MoneyCollected: mg.session.MoneyCollected,
		Encounter:    result.encounter,
		GameEnded:    result.gameEnded,
		EndReason:    result.endReason,
	}
}

// UseBomb destroys walls in a cross pattern around the player
func (mg *MiningGame) UseBomb() *MoveResult {
	if mg.session.Status != "active" {
		return &MoveResult{
			Success:     false,
			Message:     "Game is not active",
			GameEnded:   true,
			EndReason:   mg.session.Status,
		}
	}

	if mg.session.PlayerBombs <= 0 {
		return &MoveResult{
			Success: false,
			Message: "No bombs available",
			Maze:    mg.getDisplayMaze(),
			PlayerX: mg.session.PlayerX,
			PlayerY: mg.session.PlayerY,
		}
	}

	mg.session.PlayerBombs--
	destroyed := false
	pX, pY := mg.session.PlayerX, mg.session.PlayerY

	// Destroy walls in cross pattern (up/down/left/right of player)
	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, dir := range directions {
		nx, ny := pX+dir[0], pY+dir[1]
		if nx >= 0 && nx < mg.config.Size && ny >= 0 && ny < mg.config.Size {
			if mg.session.Maze[nx][ny] == CellWall {
				mg.session.Maze[nx][ny] = CellEmpty
				destroyed = true
			}
		}
	}

	if !destroyed {
	return &MoveResult{
			Success:       false,
			Message:       "No bombs available",
			Maze:          mg.getDisplayMaze(),
			PlayerX:       mg.session.PlayerX,
			PlayerY:       mg.session.PlayerY,
			PlayerBombs:   mg.session.PlayerBombs,
		}
	}

return &MoveResult{
		Success:       false,
		Message:       "No walls to destroy",
		Maze:          mg.getDisplayMaze(),
		PlayerX:       mg.session.PlayerX,
		PlayerY:       mg.session.PlayerY,
		PlayerHP:      mg.session.PlayerHP,
		PlayerBombs:   mg.session.PlayerBombs,
	}
}

// cellResult holds the result of processing a cell
type cellResult struct {
	encounter  *Encounter
	gameEnded  bool
	endReason  string
}

// processCell handles what happens when the player enters a cell
func (mg *MiningGame) processCell(x, y int) cellResult {
	cell := mg.session.Maze[x][y]
	var result cellResult

	switch cell {
	case CellMonster:
		result = mg.handleMonsterEncounter(x, y)
	case CellHeart:
		mg.session.PlayerHP = mg.session.PlayerMaxHP
		mg.session.Maze[x][y] = CellEmpty
	case CellBomb:
		mg.session.PlayerBombs++
		mg.session.Maze[x][y] = CellEmpty
	case CellMoney:
		value := float64(mg.rng.Intn(10) + 5)
		value *= mg.getMoneyMultiplier()
		mg.session.MoneyCollected += value
		mg.session.Maze[x][y] = CellEmpty
	case CellExit:
		mg.session.Status = "completed"
		mg.session.CompletedAt = time.Now()
		mg.session.MoneyCollected += 100 * mg.getMoneyMultiplier()
		result.gameEnded = true
		result.endReason = "completed"
	}

	// Check if player died
	if mg.session.PlayerHP <= 0 {
		mg.session.PlayerHP = 0
		mg.session.Status = "failed"
		mg.session.CompletedAt = time.Now()
		result.gameEnded = true
		result.endReason = "died"
	}

	return result
}

// handleMonsterEncounter resolves a monster encounter
func (mg *MiningGame) handleMonsterEncounter(x, y int) cellResult {
	var result cellResult

	// Find the monster
	for i := range mg.session.Monsters {
		m := &mg.session.Monsters[i]
		if m.X == x && m.Y == y && m.Alive {
			// Monster attacks player
			mg.session.PlayerHP -= m.Damage
			if mg.session.PlayerHP < 0 {
				mg.session.PlayerHP = 0
			}

			killReward := m.Reward
			m.Alive = false
			m.HP = 0

			mg.session.MoneyCollected += killReward

			// Remove monster from maze
			mg.session.Maze[x][y] = CellEmpty

			result.encounter = &Encounter{
				MonsterID:   m.ID,
				MonsterName: m.Name,
				MonsterIcon: m.Icon,
				Damage:      m.Damage,
				Reward:      killReward,
				Killed:      true,
				Dead:        false,
			}

			if mg.session.PlayerHP <= 0 {
				result.gameEnded = true
				result.endReason = "died"
			}

			break
		}
	}

	return result
}

// isWalkable checks if a cell is walkable (not a wall)
func (mg *MiningGame) isWalkable(x, y int) bool {
	if x < 0 || x >= mg.config.Size || y < 0 || y >= mg.config.Size {
		return false
	}
	return mg.session.Maze[x][y] != CellWall
}

// getDirectionDelta returns the delta for a move direction
func (mg *MiningGame) getDirectionDelta(dir MoveDirection) (int, int) {
	switch dir {
	case Up:
		return -1, 0
	case Down:
		return 1, 0
	case Left:
		return 0, -1
	case Right:
		return 0, 1
	default:
		return 0, 0
	}
}

// getDisplayMaze returns the current display maze with entities
func (mg *MiningGame) getDisplayMaze() [][]rune {
	display := make([][]rune, mg.config.Size)
	for i := 0; i < mg.config.Size; i++ {
		display[i] = make([]rune, mg.config.Size)
		for j := 0; j < mg.config.Size; j++ {
			display[i][j] = mg.session.Maze[i][j]
		}
	}

	// Place player
	display[mg.session.PlayerX][mg.session.PlayerY] = CellPlayer

	// Place exit
	display[mg.session.ExitX][mg.session.ExitY] = CellExit

	// Place alive monsters
	for _, m := range mg.session.Monsters {
		if m.Alive {
			for _, r := range m.Icon {
				display[m.X][m.Y] = r
				break
			}
		}
	}

	return display
}

// GetSession returns the current mining session
func (mg *MiningGame) GetSession() *MiningSession {
	return mg.session
}

// GetDisplayMaze returns the current display maze with entities
func (mg *MiningGame) GetDisplayMaze() [][]rune {
	return mg.getDisplayMaze()
}

// GetMonsterByID returns a monster by ID
func (mg *MiningGame) GetMonsterByID(id string) *Monster {
	for i := range mg.session.Monsters {
		if mg.session.Monsters[i].ID == id {
			return &mg.session.Monsters[i]
		}
	}
	return nil
}

// GetAvailableMoves returns the possible move directions
func (mg *MiningGame) GetAvailableMoves() []MoveDirection {
	var moves []MoveDirection
	directions := []struct{ dir MoveDirection; dx, dy int }{
		{Up, -1, 0},
		{Down, 1, 0},
		{Left, 0, -1},
		{Right, 0, 1},
	}

	for _, d := range directions {
		nx, ny := mg.session.PlayerX+d.dx, mg.session.PlayerY+d.dy
		if mg.isWalkable(nx, ny) {
			moves = append(moves, d.dir)
		}
	}

	return moves
}

// ParseDirection converts a string direction to MoveDirection
func ParseDirection(s string) (MoveDirection, error) {
	switch s {
	case "up", "north", "n":
		return Up, nil
	case "down", "south", "s":
		return Down, nil
	case "left", "west", "w":
		return Left, nil
	case "right", "east", "e":
		return Right, nil
	default:
		return 0, fmt.Errorf("unknown direction: %s", s)
	}
}

// DirectionToString converts a MoveDirection to a string
func DirectionToString(dir MoveDirection) string {
	switch dir {
	case Up:
		return "up"
	case Down:
		return "down"
	case Left:
		return "left"
	case Right:
		return "right"
	default:
		return "unknown"
	}
}

// MiningCooldown returns the cooldown between mining sessions
func MiningCooldown() time.Duration {
	return 30 * time.Second
}

// GetMiningCooldown returns the cooldown duration (production: 30s, development: 5min)
func GetMiningCooldown() time.Duration {
	if _, exists := getEnvDebug("DEV_MODE"); exists {
		return 5 * time.Minute
	}
	return 30 * time.Second
}

// LoadGameFromState creates a MiningGame and restores state from saved data
func LoadGameFromState(planetID, playerID string, baseLevel int, maze [][]string, playerX, playerY, exitX, exitY, playerHP, playerMaxHP, playerBombs int, moneyCollected float64, monsters []Monster) *MiningGame {
	config := DefaultMazeConfig(baseLevel)
	seed := int64(0)
	for i := 0; i < len(planetID); i++ {
		seed += int64(planetID[i]) * int64(i+1)
	}
	for i := 0; i < len(playerID); i++ {
		seed += int64(playerID[i]) * int64(i+1)
	}
	seed += int64(baseLevel) * 1000000
	seed = time.Now().UnixNano() + seed

	mg := &MiningGame{
		config: config,
		rng:    rand.New(rand.NewSource(seed)),
		session: &MiningSession{
			PlanetID:     planetID,
			PlayerID:     playerID,
			SessionID:    fmt.Sprintf("mining_%s_%d", planetID, seed),
			Status:       "active",
			BaseLevel:    baseLevel,
			StartTime:    time.Now(),
			PlayerX:      playerX,
			PlayerY:      playerY,
			ExitX:        exitX,
			ExitY:        exitY,
			PlayerHP:     playerHP,
			PlayerMaxHP:  playerMaxHP,
			PlayerBombs:  playerBombs,
			MoneyCollected: moneyCollected,
		},
	}

	// Restore maze
	mg.session.Maze = make([][]rune, len(maze))
	mg.session.DisplayMaze = make([][]rune, len(maze))
	for i, row := range maze {
		size := len(row)
		mg.session.Maze[i] = make([]rune, size)
		mg.session.DisplayMaze[i] = make([]rune, size)
		for j, cell := range row {
			r := rune(cell[0])
			mg.session.Maze[i][j] = r
			mg.session.DisplayMaze[i][j] = r
		}
	}

	// Restore monsters
	mg.session.Monsters = monsters

	return mg
}

// getEnvDebug is a helper to check for dev mode
func getEnvDebug(key string) (string, bool) {
	val := ""
	found := false
	_ = val
	_ = found
	return "", false
}
