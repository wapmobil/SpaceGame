package game

import (
	"math/rand"
	"testing"
	"time"
)

func TestNewDrillGame(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	session := game.GetSession()

	if session.DrillMaxHP != 120 { // 80 + 1*40
		t.Errorf("Expected max HP 120, got %d", session.DrillMaxHP)
	}
	if session.DrillHP != 120 {
		t.Errorf("Expected HP 120, got %d", session.DrillHP)
	}
	if session.Depth != 0 {
		t.Errorf("Expected depth 0, got %d", session.Depth)
	}
	if session.DrillX != DefaultWorldWidth/2 {
		t.Errorf("Expected drill X %d, got %d", DefaultWorldWidth/2, session.DrillX)
	}
	if session.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", session.Status)
	}
	if len(session.World) == 0 {
		t.Error("World should not be empty")
	}
	if len(session.World[0]) != DefaultWorldWidth {
		t.Errorf("Expected world width %d, got %d", DefaultWorldWidth, len(session.World[0]))
	}
}

func TestDrillGameWithDifferentLevels(t *testing.T) {
	tests := []struct {
		level      int
		expectedHP int
	}{
		{1, 120},
		{2, 160},
		{3, 200},
		{5, 280},
		{10, 480},
	}

	for _, tt := range tests {
		game := NewDrillGame("planet-1", "player-1", tt.level)
		session := game.GetSession()
		if session.DrillMaxHP != tt.expectedHP {
			t.Errorf("Level %d: expected HP %d, got %d", tt.level, tt.expectedHP, session.DrillHP)
		}
	}
}

func TestMoveLeft(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	result := game.Move(MoveLeft, false)

	if !result.Success {
		t.Error("Move should succeed")
	}
	if result.DrillX != DefaultWorldWidth/2-1 {
		t.Errorf("Expected drill X %d, got %d", DefaultWorldWidth/2-1, result.DrillX)
	}
}

func TestMoveRight(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	result := game.Move(MoveRight, false)

	if !result.Success {
		t.Error("Move should succeed")
	}
	if result.DrillX != DefaultWorldWidth/2+1 {
		t.Errorf("Expected drill X %d, got %d", DefaultWorldWidth/2+1, result.DrillX)
	}
}

func TestMoveBoundaries(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)

	// Try to move left from position 0
	game.session.DrillX = 0
	result := game.Move(MoveLeft, false)
	if result.DrillX != 0 {
		t.Errorf("Should not move left from boundary, got X=%d", result.DrillX)
	}

	// Try to move right from max position
	game.session.DrillX = DefaultWorldWidth - 1
	result = game.Move(MoveRight, false)
	if result.DrillX != DefaultWorldWidth-1 {
		t.Errorf("Should not move right from boundary, got X=%d", result.DrillX)
	}
}

func TestMoveDown_Damage(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	initialHP := game.GetSession().DrillHP

	// Move down repeatedly
	for i := 0; i < 10; i++ {
		result := game.Move(MoveDown, false)
		if !result.Success {
			t.Errorf("Move down failed at depth %d: %s", i, result.Message)
		}
		if game.GetSession().DrillHP >= initialHP {
			t.Errorf("HP should decrease after moving down, got %d", game.GetSession().DrillHP)
		}
	}

	if game.GetSession().Depth != 10 {
		t.Errorf("Expected depth 10, got %d", game.GetSession().Depth)
	}
}

func TestMoveDown_CellDamage(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10) // high HP to survive
	initialHP := game.GetSession().DrillHP

	// Move down 10 cells
	for i := 0; i < 10; i++ {
		game.Move(MoveDown, false)
	}

	damage := initialHP - game.GetSession().DrillHP
	if damage <= 0 {
		t.Error("Should take damage from cells")
	}

	// Dirt does 2 damage, stone does 5, metal does 10, mithril does 15
	// Average should be around 4-5 per cell
	if damage < 10 || damage > 80 {
		t.Errorf("Unexpected total damage: %d (expected 10-80 for 10 cells)", damage)
	}
}

func TestDrillDestroyed(t *testing.T) {
	// Create a game with very low HP
	session := DrillSession{
		ID:         "test",
		SessionID:  "test",
		PlanetID:   "planet-1",
		PlayerID:   "player-1",
		DrillHP:    5,
		DrillMaxHP: 5,
		Depth:      0,
		DrillX:     10,
		WorldWidth: DefaultWorldWidth,
		Status:     "active",
		World:      make([][]Cell, 100),
	}
	for y := 0; y < 100; y++ {
		session.World[y] = make([]Cell, DefaultWorldWidth)
		for x := 0; x < DefaultWorldWidth; x++ {
			session.World[y][x] = Cell{X: x, Y: y, CellType: CellStone}
		}
	}

	game := &DrillGame{
		config: DrillConfig{
			WorldWidth: DefaultWorldWidth,
			ViewHeight: DefaultViewHeight,
			Seed:       42,
		},
		session: session,
		rng:     rand.New(rand.NewSource(42)),
	}

	result := game.Move(MoveDown, false)
	if !result.GameEnded {
		t.Error("Game should end when drill is destroyed")
	}
	if result.EndReason != "drill_destroyed" {
		t.Errorf("Expected end reason 'drill_destroyed', got '%s'", result.EndReason)
	}
	if game.GetSession().Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", game.GetSession().Status)
	}
}

func TestMoveDown_ActiveGame(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)

	// Move down should succeed
	result := game.Move(MoveDown, false)
	if !result.Success {
		t.Errorf("Move down should succeed: %s", result.Message)
	}
	if result.GameEnded {
		t.Error("Game should not end on first move down")
	}
}

func TestAvailableDirections(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	dirs := game.GetAvailableDirections()

	if len(dirs) != 3 { // left, right, down
		t.Errorf("Expected 3 directions, got %d", len(dirs))
	}

	// Move to left boundary
	game.session.DrillX = 0
	dirs = game.GetAvailableDirections()
	if len(dirs) != 2 { // right, down
		t.Errorf("Expected 2 directions at left boundary, got %d", len(dirs))
	}

	// Move to right boundary
	game.session.DrillX = DefaultWorldWidth - 1
	dirs = game.GetAvailableDirections()
	if len(dirs) != 2 { // left, down
		t.Errorf("Expected 2 directions at right boundary, got %d", len(dirs))
	}
}

func TestParseDrillDirection(t *testing.T) {
	tests := []struct {
		input    string
		expected MoveDirection
		hasError bool
	}{
		{"left", MoveLeft, false},
		{"l", MoveLeft, false},
		{"west", MoveLeft, false},
		{"right", MoveRight, false},
		{"r", MoveRight, false},
		{"east", MoveRight, false},
		{"down", MoveDown, false},
		{"d", MoveDown, false},
		{"south", MoveDown, false},
		{"", MoveDown, false},
		{"invalid", MoveDown, true},
	}

	for _, tt := range tests {
		dir, err := ParseDrillDirection(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("Expected error for input '%s'", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
			}
			if dir != tt.expected {
				t.Errorf("Input '%s': expected %d, got %d", tt.input, tt.expected, dir)
			}
		}
	}
}

func TestDrillDirectionString(t *testing.T) {
	tests := []struct {
		dir      MoveDirection
		expected string
	}{
		{MoveLeft, "left"},
		{MoveRight, "right"},
		{MoveDown, "down"},
	}

	for _, tt := range tests {
		s := GetDrillDirectionString(tt.dir)
		if s != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, s)
		}
	}
}

func TestGetDrillCooldown(t *testing.T) {
	cooldown := GetDrillCooldown()
	if cooldown != 30*time.Second {
		t.Errorf("Expected cooldown 30s, got %v", cooldown)
	}
}

func TestDisplayWorld(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	display := game.GetDisplayWorld()

	if len(display) != DefaultViewHeight {
		t.Errorf("Expected display height %d, got %d", DefaultViewHeight, len(display))
	}

	for i, row := range display {
		if len(row) != DefaultWorldWidth {
			t.Errorf("Row %d: expected width %d, got %d", i, DefaultWorldWidth, len(row))
		}
	}
}

func TestMoveOnEndedGame(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	game.session.Status = "failed"

	result := game.Move(MoveLeft, false)
	if result.Success {
		t.Error("Move should fail on ended game")
	}
	if result.Message != "Drill session is not active" {
		t.Errorf("Expected 'Drill session is not active', got '%s'", result.Message)
	}
}

func TestResourceDefinitions(t *testing.T) {
	if len(resourceDefinitions) == 0 {
		t.Error("Should have resource definitions")
	}

	for key, def := range resourceDefinitions {
		if def.Type == "" {
			t.Errorf("Resource %s has empty type", key)
		}
		if def.Name == "" {
			t.Errorf("Resource %s has empty name", key)
		}
		if def.Value <= 0 {
			t.Errorf("Resource %s has non-positive value: %f", key, def.Value)
		}
		if def.DigTime <= 0 {
			t.Errorf("Resource %s has non-positive dig time: %f", key, def.DigTime)
		}
		if def.SpawnChance <= 0 || def.SpawnChance > 1 {
			t.Errorf("Resource %s has invalid spawn chance: %f", key, def.SpawnChance)
		}
	}
}

func TestLoadGameFromState(t *testing.T) {
	world := make([][]Cell, 100)
	for y := 0; y < 100; y++ {
		world[y] = make([]Cell, DefaultWorldWidth)
		for x := 0; x < DefaultWorldWidth; x++ {
			world[y][x] = Cell{X: x, Y: y, CellType: CellDirt}
		}
	}

	game := LoadGameFromState("planet-1", "player-1", 3, world, 150, 200, 10, 5, []DrillResource{}, 0, "active")
	session := game.GetSession()

	if session.DrillHP != 150 {
		t.Errorf("Expected HP 150, got %d", session.DrillHP)
	}
	if session.DrillMaxHP != 200 {
		t.Errorf("Expected max HP 200, got %d", session.DrillMaxHP)
	}
	if session.Depth != 10 {
		t.Errorf("Expected depth 10, got %d", session.Depth)
	}
	if session.DrillX != 5 {
		t.Errorf("Expected drill X 5, got %d", session.DrillX)
	}
	if session.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", session.Status)
	}
}

func TestGetResourcesAsJSON(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	game.session.Resources = []DrillResource{
		{Type: ResourceOil, Name: "Нефть", Icon: "🛢️", Amount: 10, Value: 150},
		{Type: ResourceGold, Name: "Золото", Icon: "🟡", Amount: 5, Value: 500},
	}

	jsonStr := game.GetResourcesAsJSON()
	if jsonStr == "[]" {
		t.Error("JSON should not be empty")
	}

	parsed, err := ParseResourcesFromJSON(jsonStr)
	if err != nil {
		t.Errorf("Failed to parse resources: %v", err)
	}
	if len(parsed) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(parsed))
	}
}

func TestDrillGameGeneratesResources(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	world := game.GetSession().World

	resourceCount := 0
	for y := 0; y < 100; y++ {
		for x := 0; x < DefaultWorldWidth; x++ {
			if world[y][x].ResourceType != "" {
				resourceCount++
			}
		}
	}

	if resourceCount == 0 {
		t.Error("World should contain some resources")
	}
}

func TestDrillGameDepthProgression(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 5)

	for i := 0; i < 20; i++ {
		result := game.Move(MoveDown, false)
		if result.GameEnded {
			t.Logf("Game ended at depth %d: %s", result.Depth, result.EndReason)
			break
		}
		if game.GetSession().Depth != i+1 {
			t.Errorf("Expected session depth %d, got %d", i+1, game.GetSession().Depth)
		}
	}
}
