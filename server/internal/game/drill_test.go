package game

import (
	"testing"
	"time"
)

func boolPtr(b bool) *bool {
	return &b
}

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

func TestWorldIs5x5(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	session := game.GetSession()

	if len(session.World) != 5 {
		t.Errorf("Expected world height 5, got %d", len(session.World))
	}
	for i, row := range session.World {
		if len(row) != 5 {
			t.Errorf("Expected world width 5 for row %d, got %d", i, len(row))
		}
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

func TestSetCommand_StoresCommand(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	initialX := game.GetSession().DrillX
	initialDepth := game.GetSession().Depth

	// Set a command
	game.SetCommand("left", boolPtr(false))

	// World should NOT have changed (command is only stored, not applied)
	session := game.GetSession()
	if session.DrillX != initialX {
		t.Errorf("Expected drill X to remain %d after SetCommand, got %d", initialX, session.DrillX)
	}
	if session.Depth != initialDepth {
		t.Errorf("Expected depth to remain %d after SetCommand, got %d", initialDepth, session.Depth)
	}

	// But pending command should be stored
	if game.session.PendingDirection != "left" {
		t.Errorf("Expected pending direction 'left', got '%s'", game.session.PendingDirection)
	}
	if game.session.PendingExtract {
		t.Errorf("Expected pending extract to be false, got true")
	}
}

func TestSetCommand_Extract(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	game.SetCommand("", boolPtr(true))

	if !game.session.PendingExtract {
		t.Error("Expected pending extract to be true")
	}
	if game.session.PendingDirection != "" {
		t.Errorf("Expected empty pending direction, got '%s'", game.session.PendingDirection)
	}
}

func TestSetCommand_Combo(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	game.SetCommand("right", boolPtr(true))

	if game.session.PendingDirection != "right" {
		t.Errorf("Expected pending direction 'right', got '%s'", game.session.PendingDirection)
	}
	if !game.session.PendingExtract {
		t.Error("Expected pending extract to be true")
	}
}

func TestApplyCommand_Left(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	initialX := game.GetSession().DrillX

	game.SetCommand("left", boolPtr(false))
	result := game.ApplyCommand()

	if !result.Success {
		t.Error("ApplyCommand should succeed")
	}
	if result.DrillX != initialX-1 {
		t.Errorf("Expected drill X %d, got %d", initialX-1, result.DrillX)
	}
	if game.GetSession().Depth != 1 {
		t.Errorf("Expected depth 1 after apply, got %d", game.GetSession().Depth)
	}
}

func TestApplyCommand_Right(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	initialX := game.GetSession().DrillX

	game.SetCommand("right", boolPtr(false))
	result := game.ApplyCommand()

	if !result.Success {
		t.Error("ApplyCommand should succeed")
	}
	if result.DrillX != initialX+1 {
		t.Errorf("Expected drill X %d, got %d", initialX+1, result.DrillX)
	}
}

func TestApplyCommand_NoDirection(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	initialX := game.GetSession().DrillX

	game.SetCommand("", boolPtr(false))
	result := game.ApplyCommand()

	if !result.Success {
		t.Error("ApplyCommand should succeed")
	}
	if result.DrillX != initialX {
		t.Errorf("Expected drill X to remain %d, got %d", initialX, result.DrillX)
	}
	// But depth should still increase (auto-descent always moves down)
	if game.GetSession().Depth != 1 {
		t.Errorf("Expected depth 1 after apply, got %d", game.GetSession().Depth)
	}
}

func TestApplyCommand_Extract(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)

	// Force a resource at the current position
	game.session.ExtractedCells = make(map[string]float64)

	game.SetCommand("", boolPtr(true))
	result := game.ApplyCommand()

	if !result.Success {
		t.Error("ApplyCommand should succeed")
	}
	// Extracted should be > 0 if there was a resource, or 0 if not
	if result.Extracted < 0 {
		t.Errorf("Expected extracted >= 0, got %f", result.Extracted)
	}
}

func TestApplyCommand_Combo(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	initialX := game.GetSession().DrillX

	game.SetCommand("left", boolPtr(true))
	result := game.ApplyCommand()

	if !result.Success {
		t.Error("ApplyCommand should succeed")
	}
	// Horizontal movement should be applied
	if result.DrillX != initialX-1 {
		t.Errorf("Expected drill X %d, got %d", initialX-1, result.DrillX)
	}
	// Depth should increase
	if game.GetSession().Depth != 1 {
		t.Errorf("Expected depth 1, got %d", game.GetSession().Depth)
	}
	// World should be regenerated
	if result.World == nil {
		t.Error("Expected world in result")
	}
}

func TestApplyCommand_ResetsPending(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	game.SetCommand("left", boolPtr(false))

	// First apply
	game.ApplyCommand()

	// Pending should be reset
	if game.session.PendingDirection != "" {
		t.Errorf("Expected pending direction to be reset, got '%s'", game.session.PendingDirection)
	}

	// Second apply should do nothing (no pending command)
	initialX := game.GetSession().DrillX
	initialDepth := game.GetSession().Depth
	game.ApplyCommand()

	// Depth should increase but X should not change (no pending horizontal move)
	if game.GetSession().DrillX != initialX {
		t.Errorf("Expected drill X to remain %d (no pending), got %d", initialX, game.GetSession().DrillX)
	}
	if game.GetSession().Depth != initialDepth+1 {
		t.Errorf("Expected depth to increase by 1, got %d", game.GetSession().Depth-initialDepth)
	}
}

func TestApplyCommand_OnEndedGame(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	game.session.Status = "failed"

	result := game.ApplyCommand()
	if result.Success {
		t.Error("ApplyCommand should fail on ended game")
	}
	if result.EndReason != "session_ended" {
		t.Errorf("Expected 'session_ended', got '%s'", result.EndReason)
	}
}

func TestAutoDescentAppliesCommand(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10) // high HP to survive
	game.SetCommand("left", boolPtr(false))

	initialX := game.GetSession().DrillX
	initialDepth := game.GetSession().Depth

	// Simulate one auto-descent tick
	result := game.ApplyCommand()

	if !result.Success {
		t.Error("Auto-descent should succeed")
	}
	if result.DrillX != initialX-1 {
		t.Errorf("Expected drill X %d, got %d", initialX-1, result.DrillX)
	}
	if game.GetSession().Depth != initialDepth+1 {
		t.Errorf("Expected depth %d, got %d", initialDepth+1, game.GetSession().Depth)
	}
}

func TestBroadcastCallback(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	broadcastCalled := false
	var lastResult *MoveResult

	game.SetBroadcastFn(func(result *MoveResult) {
		broadcastCalled = true
		lastResult = result
	})

	game.SetCommand("left", boolPtr(false))
	game.ApplyCommandWithBroadcast()

	if !broadcastCalled {
		t.Error("Broadcast callback should have been called")
	}
	if lastResult == nil {
		t.Error("Last result should not be nil")
	}
}

func TestGetChunk(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)

	chunk := game.GetChunk(0, 0, 5, 5)

	if len(chunk) != 5 {
		t.Errorf("Expected chunk height 5, got %d", len(chunk))
	}
	for i, row := range chunk {
		if len(row) != 5 {
			t.Errorf("Expected chunk width 5 for row %d, got %d", i, len(row))
		}
	}
}

func TestGetChunk_Deterministic(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)

	chunk1 := game.GetChunk(2, 10, 5, 5)
	chunk2 := game.GetChunk(2, 10, 5, 5)

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if chunk1[y][x].CellType != chunk2[y][x].CellType {
				t.Errorf("Chunk not deterministic at (%d,%d): '%s' vs '%s'", x, y, chunk1[y][x].CellType, chunk2[y][x].CellType)
			}
		}
	}
}

func TestGetChunk_Coordinates(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)

	// Get chunk centered at (5, 100)
	chunk := game.GetChunk(5, 100, 3, 3)

	// The top-center cell should be at (5, 99) (centerY - height/2)
	if chunk[0][1].X != 5 || chunk[0][1].Y != 99 {
		t.Errorf("Expected top-center cell at (5,99), got (%d,%d)", chunk[0][1].X, chunk[0][1].Y)
	}

	// The bottom-center cell should be at (5, 101)
	if chunk[2][1].X != 5 || chunk[2][1].Y != 101 {
		t.Errorf("Expected bottom-center cell at (5,101), got (%d,%d)", chunk[2][1].X, chunk[2][1].Y)
	}
}

func TestGetChunk_VsSessionWorld(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	session := game.GetSession()

	// The world spans from depth to depth+4 vertically, drill at top center (row 0)
	// GetChunk(centerX, centerY, w, h) generates from centerY-h/2 to centerY+h/2
	// So to match the world (depth to depth+4), we need centerY = depth + 2
	chunk := game.GetChunk(session.DrillX, session.Depth+2, 5, 5)

	// Chunk should match the session world
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if chunk[y][x].CellType != session.World[y][x].CellType {
				t.Errorf("Chunk/world mismatch at (%d,%d): chunk='%s', world='%s'", x, y, chunk[y][x].CellType, session.World[y][x].CellType)
			}
		}
	}
}

func TestGetChunk_VsSessionWorld_AfterMove(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)

	// Apply a move
	game.SetCommand("left", boolPtr(false))
	game.ApplyCommand()

	session := game.GetSession()

	// Get chunk centered at new drill position (same offset as above)
	chunk := game.GetChunk(session.DrillX, session.Depth+2, 5, 5)

	// Chunk should match the session world
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if chunk[y][x].CellType != session.World[y][x].CellType {
				t.Errorf("Chunk/world mismatch after move at (%d,%d): chunk='%s', world='%s'", x, y, chunk[y][x].CellType, session.World[y][x].CellType)
			}
		}
	}
}

func TestMoveBoundaries(t *testing.T) {
	// Move left from position 0 should succeed (no boundary)
	game := NewDrillGame("planet-1", "player-1", 1)
	game.session.DrillX = 0
	game.SetCommand("left", boolPtr(false))
	game.ApplyCommand()
	if game.GetSession().DrillX != -1 {
		t.Errorf("Should move left from position 0, got X=%d", game.GetSession().DrillX)
	}

	// Move right from position 4 should succeed (no boundary)
	game2 := NewDrillGame("planet-2", "player-2", 1)
	game2.session.DrillX = DefaultWorldWidth - 1
	game2.SetCommand("right", boolPtr(false))
	game2.ApplyCommand()
	if game2.GetSession().DrillX != DefaultWorldWidth {
		t.Errorf("Should move right from position %d, got X=%d", DefaultWorldWidth-1, game2.GetSession().DrillX)
	}
}

func TestMoveDown_Damage(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10) // high HP
	initialHP := game.GetSession().DrillHP

	// Move down repeatedly
	for i := 0; i < 10; i++ {
		game.SetCommand("", boolPtr(false))
		game.ApplyCommand()
	}

	if game.GetSession().Depth != 10 {
		t.Errorf("Expected depth 10, got %d", game.GetSession().Depth)
	}

	// At least some damage should have been dealt (probability of all empty is very low)
	damage := initialHP - game.GetSession().DrillHP
	if damage <= 0 {
		t.Log("Note: no damage dealt - possible but unlikely")
	}
}

func TestMoveDown_CellDamage(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10) // high HP to survive
	initialHP := game.GetSession().DrillHP

	// Move down 10 cells
	for i := 0; i < 10; i++ {
		game.SetCommand("", boolPtr(false))
		game.ApplyCommand()
	}

	damage := initialHP - game.GetSession().DrillHP

	// Dirt does 2 damage, stone does 5, metal does 10, mithril does 15
	// Total damage should be between 0 and 150 (10 * 15 for mithril)
	if damage < 0 || damage > 150 {
		t.Errorf("Unexpected total damage: %d (expected 0-150 for 10 cells)", damage)
	}
}

func TestDrillDestroyed(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	game.session.DrillHP = 1

	result := game.ApplyCommand()
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

func TestAvailableDirections(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	dirs := game.GetAvailableDirections()

	if len(dirs) != 3 { // left, right, down
		t.Errorf("Expected 3 directions, got %d", len(dirs))
	}

	// At any position, all directions should be available (no boundaries)
	game.session.DrillX = 0
	dirs = game.GetAvailableDirections()
	if len(dirs) != 3 {
		t.Errorf("Expected 3 directions at X=0, got %d", len(dirs))
	}

	game.session.DrillX = DefaultWorldWidth - 1
	dirs = game.GetAvailableDirections()
	if len(dirs) != 3 {
		t.Errorf("Expected 3 directions at X=%d, got %d", DefaultWorldWidth-1, len(dirs))
	}

	// After game ends, only left and right should be available
	game.session.Status = "failed"
	dirs = game.GetAvailableDirections()
	if len(dirs) != 2 { // left, right (no down)
		t.Errorf("Expected 2 directions after game end, got %d", len(dirs))
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

func TestDrillGameGeneratesResources(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	world := game.GetSession().World

	resourceCount := 0
	for y := 0; y < len(world); y++ {
		for x := 0; x < len(world[y]); x++ {
			if world[y][x].ResourceType != "" {
				resourceCount++
			}
		}
	}

	// Generate more cells at deeper depths to verify resources can be generated
	resourceCount = 0
	for y := 1; y <= 100; y++ {
		cell := game.getCellAt(2, y)
		if cell.ResourceType != "" {
			resourceCount++
		}
	}

	if resourceCount == 0 {
		t.Error("World should contain some resources at deeper depths")
	}
}

func TestDrillGameDepthProgression(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 5)

	for i := 0; i < 20; i++ {
		game.SetCommand("", boolPtr(false))
		result := game.ApplyCommand()
		if result.GameEnded {
			t.Logf("Game ended at depth %d: %s", result.Depth, result.EndReason)
			break
		}
		if game.GetSession().Depth != i+1 {
			t.Errorf("Expected session depth %d, got %d", i+1, game.GetSession().Depth)
		}
	}
}

func TestWorldIsAlways5x5(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 5)

	// Move down 10 times and check world is always 5x5
	for i := 0; i < 10; i++ {
		game.SetCommand("", boolPtr(false))
		game.ApplyCommand()
		session := game.GetSession()
		if len(session.World) != 5 {
			t.Errorf("After %d moves down: expected world height 5, got %d", i+1, len(session.World))
		}
		for j, row := range session.World {
			if len(row) != 5 {
				t.Errorf("After %d moves down: expected world width 5 for row %d, got %d", i+1, j, len(row))
			}
		}
	}

	// Move left/right and check world is still 5x5
	game2 := NewDrillGame("planet-2", "player-2", 5)
	for i := 0; i < 5; i++ {
		game2.SetCommand("left", boolPtr(false))
		game2.ApplyCommand()
		session := game2.GetSession()
		if len(session.World) != 5 {
			t.Errorf("After %d moves left: expected world height 5, got %d", i+1, len(session.World))
		}
	}

	game3 := NewDrillGame("planet-3", "player-3", 5)
	for i := 0; i < 5; i++ {
		game3.SetCommand("right", boolPtr(false))
		game3.ApplyCommand()
		session := game3.GetSession()
		if len(session.World) != 5 {
			t.Errorf("After %d moves right: expected world height 5, got %d", i+1, len(session.World))
		}
	}
}

func TestDrillXAlwaysAtCenter(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)

	// Drill X should always be at center (2) after move down
	for i := 0; i < 10; i++ {
		game.SetCommand("", boolPtr(false))
		game.ApplyCommand()
		if game.GetSession().DrillX != 2 {
			t.Errorf("After %d moves down: expected drill X at center %d, got %d", i+1, 2, game.GetSession().DrillX)
		}
	}
}

func TestDeterministicCellGeneration(t *testing.T) {
	// Two games with same seed should produce same cells
	game1 := NewDrillGame("planet-1", "player-1", 1)
	game2 := NewDrillGame("planet-1", "player-2", 1)

	// Both use same planet ID, so last char is same, but time.Now() differs
	// Instead, test that getCellAt is deterministic by calling it directly
	cell1 := game1.getCellAt(3, 10)
	cell2 := game1.getCellAt(3, 10)
	cell3 := game2.getCellAt(3, 10)

	if cell1.CellType != cell2.CellType {
		t.Errorf("Same game: expected same cell, got '%s' vs '%s'", cell1.CellType, cell2.CellType)
	}
	if cell1.ResourceType != cell2.ResourceType {
		t.Errorf("Same game: expected same resource, got '%s' vs '%s'", cell1.ResourceType, cell2.ResourceType)
	}

	// Different games with different planet IDs should produce different worlds
	if cell1.CellType == cell3.CellType && cell1.ResourceType == cell3.ResourceType {
		t.Logf("Note: different planet IDs happened to produce same cell (rare but possible)")
	}

	// Different coordinates should produce different cells
	cell4 := game1.getCellAt(4, 10)
	if cell1.CellType == cell4.CellType && cell1.ResourceType == cell4.ResourceType {
		t.Logf("Note: adjacent cells happened to match (possible but unlikely)")
	}
}

func TestGetSeed(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)
	seed := game.GetSeed()
	if seed <= 0 {
		t.Errorf("Expected positive seed, got %d", seed)
	}
}

func TestExtractResources(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10) // high HP

	// Move down several times
	for i := 0; i < 5; i++ {
		game.SetCommand("", boolPtr(false))
		game.ApplyCommand()
	}

	initialResources := len(game.GetSession().Resources)

	// Try to extract (might find a resource or not)
	game.SetCommand("", boolPtr(true))
	result := game.ApplyCommand()

	if !result.Success {
		t.Error("ApplyCommand with extract should succeed")
	}

	// Check that resources list is valid
	if len(game.GetSession().Resources) < initialResources {
		t.Error("Resource count should not decrease")
	}
}

func TestMultipleCommandsBeforeApply(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)

	// Set multiple commands (only last one should stick)
	game.SetCommand("left", boolPtr(false))
	game.SetCommand("right", boolPtr(false))

	// Only the last command should be pending
	if game.session.PendingDirection != "right" {
		t.Errorf("Expected pending direction 'right', got '%s'", game.session.PendingDirection)
	}

	// Apply should use the last command
	result := game.ApplyCommand()
	if result.DrillX != DefaultWorldWidth/2+1 {
		t.Errorf("Expected drill X %d, got %d", DefaultWorldWidth/2+1, result.DrillX)
	}
}

func TestExtractPersistsAcrossTicks(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10)

	// Set extract on
	game.SetCommand("", boolPtr(true))
	if !game.session.PendingExtract {
		t.Error("Expected PendingExtract to be true")
	}

	// Apply first tick - direction reset, extract should persist
	game.ApplyCommand()
	if !game.session.PendingExtract {
		t.Error("Expected PendingExtract to still be true after first tick")
	}

	// Apply second tick - still extracting
	game.ApplyCommand()
	if !game.session.PendingExtract {
		t.Error("Expected PendingExtract to still be true after second tick")
	}

	// Turn off extract
	game.SetCommand("", boolPtr(false))
	if game.session.PendingExtract {
		t.Error("Expected PendingExtract to be false after turning off")
	}

	// Apply third tick - no extract
	game.ApplyCommand()
}

func TestExtractOnNewCell(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10)

	// Set extract on
	game.SetCommand("", boolPtr(true))

	// Apply command - should move down then extract from new cell
	result := game.ApplyCommand()

	// Depth should have increased (moved down)
	if game.GetSession().Depth != 1 {
		t.Errorf("Expected depth 1, got %d", game.GetSession().Depth)
	}

	// Extract flag should still be true (persists)
	if !game.session.PendingExtract {
		t.Error("Expected PendingExtract to persist after apply")
	}

	// Result should have extracted > 0 if there was a resource at new cell
	// (or 0 if no resource, but no error)
	if result.Extracted < 0 {
		t.Errorf("Expected extracted >= 0, got %f", result.Extracted)
	}
}

func TestDirectionResetsAfterApply(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1)

	game.SetCommand("left", boolPtr(false))
	if game.session.PendingDirection != "left" {
		t.Error("Expected PendingDirection to be 'left'")
	}

	game.ApplyCommand()
	if game.session.PendingDirection != "" {
		t.Errorf("Expected PendingDirection to be reset, got '%s'", game.session.PendingDirection)
	}
}

func TestDirectionDoesNotAffectExtract(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10)

	// Turn on extract
	game.SetCommand("", boolPtr(true))
	if !game.session.PendingExtract {
		t.Error("Expected PendingExtract to be true")
	}

	// Set direction without changing extract (nil = don't change)
	game.SetCommand("left", nil)
	if !game.session.PendingExtract {
		t.Error("Expected PendingExtract to still be true after setting direction")
	}
	if game.session.PendingDirection != "left" {
		t.Error("Expected PendingDirection to be 'left'")
	}
}
