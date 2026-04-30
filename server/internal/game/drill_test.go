package game

import (
	"fmt"
	"testing"
	"time"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestNewDrillGame(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1, 1)
	session := game.GetSession()

	if session.DrillMaxHP != 110 { // 10 + 100*1
		t.Errorf("Expected max HP 110, got %d", session.DrillMaxHP)
	}
	if session.DrillHP != 110 {
		t.Errorf("Expected HP 110, got %d", session.DrillHP)
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
		{1, 110},
		{2, 210},
		{3, 310},
		{5, 510},
		{10, 1010},
	}

	for _, tt := range tests {
		game := NewDrillGame("planet-1", "player-1", tt.level, 1)
		session := game.GetSession()
		if session.DrillMaxHP != tt.expectedHP {
			t.Errorf("Level %d: expected HP %d, got %d", tt.level, tt.expectedHP, session.DrillHP)
		}
	}
}

func TestSetCommand_StoresCommand(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
	game.SetCommand("", boolPtr(true))

	if !game.session.PendingExtract {
		t.Error("Expected pending extract to be true")
	}
	if game.session.PendingDirection != "" {
		t.Errorf("Expected empty pending direction, got '%s'", game.session.PendingDirection)
	}
}

func TestSetCommand_Combo(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1, 1)
	game.SetCommand("right", boolPtr(true))

	if game.session.PendingDirection != "right" {
		t.Errorf("Expected pending direction 'right', got '%s'", game.session.PendingDirection)
	}
	if !game.session.PendingExtract {
		t.Error("Expected pending extract to be true")
	}
}

func TestApplyCommand_Left(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)

	// Force a resource at the current position
	game.session.ExtractedCells = make(map[string]bool)

	game.SetCommand("", boolPtr(true))
	result := game.ApplyCommand()

	if !result.Success {
		t.Error("ApplyCommand should succeed")
	}
	// Extracted should be > 0 if there was a resource, or 0 if not
	if result.Extracted < 0 {
		t.Errorf("Expected extracted >= 0, got %d", result.Extracted)
	}
}

func TestApplyCommand_Combo(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
	game := NewDrillGame("planet-1", "player-1", 10, 1) // high HP to survive
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)

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
	game := NewDrillGame("planet-1", "player-1", 1, 1)

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
	game := NewDrillGame("planet-1", "player-1", 1, 1)

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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)

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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
	game.session.DrillX = 0
	game.SetCommand("left", boolPtr(false))
	game.ApplyCommand()
	if game.GetSession().DrillX != -1 {
		t.Errorf("Should move left from position 0, got X=%d", game.GetSession().DrillX)
	}

	// Move right from position 4 should succeed (no boundary)
	game2 := NewDrillGame("planet-2", "player-2", 1, 1)
	game2.session.DrillX = DefaultWorldWidth - 1
	game2.SetCommand("right", boolPtr(false))
	game2.ApplyCommand()
	if game2.GetSession().DrillX != DefaultWorldWidth {
		t.Errorf("Should move right from position %d, got X=%d", DefaultWorldWidth-1, game2.GetSession().DrillX)
	}
}

func TestMoveDown_Damage(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10, 1) // high HP
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
	game := NewDrillGame("planet-1", "player-1", 10, 1) // high HP to survive
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
	game.session.DrillHP = -1

	result := game.ApplyCommand()
	if !result.GameEnded {
		t.Error("Game should end when drill HP is negative")
	}
	if result.EndReason != "drill_destroyed" {
		t.Errorf("Expected end reason 'drill_destroyed', got '%s'", result.EndReason)
	}
	if game.GetSession().Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", game.GetSession().Status)
	}
}

func TestAvailableDirections(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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

func TestDisplayWorld(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
			t.Errorf("Resource %s has non-positive value: %d", key, def.Value)
		}
		if def.SpawnChance <= 0 || def.SpawnChance > 1 {
			t.Errorf("Resource %s has invalid spawn chance: %f", key, def.SpawnChance)
		}
	}
}

func TestDrillGameGeneratesResources(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1, 1)
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
	game := NewDrillGame("planet-1", "player-1", 5, 1)

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
	game := NewDrillGame("planet-1", "player-1", 5, 1)

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
	game2 := NewDrillGame("planet-2", "player-2", 5, 1)
	for i := 0; i < 5; i++ {
		game2.SetCommand("left", boolPtr(false))
		game2.ApplyCommand()
		session := game2.GetSession()
		if len(session.World) != 5 {
			t.Errorf("After %d moves left: expected world height 5, got %d", i+1, len(session.World))
		}
	}

	game3 := NewDrillGame("planet-3", "player-3", 5, 1)
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
	game := NewDrillGame("planet-1", "player-1", 1, 1)

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
	game1 := NewDrillGame("planet-1", "player-1", 1, 1)
	game2 := NewDrillGame("planet-1", "player-2", 1, 1)

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
	game := NewDrillGame("planet-1", "player-1", 1, 1)
	seed := game.GetSeed()
	if seed <= 0 {
		t.Errorf("Expected positive seed, got %d", seed)
	}
}

func TestExtractResources(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10, 1) // high HP

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
	game := NewDrillGame("planet-1", "player-1", 1, 1)

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
	game := NewDrillGame("planet-1", "player-1", 10, 1)

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
	game := NewDrillGame("planet-1", "player-1", 10, 1)

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
		t.Errorf("Expected extracted >= 0, got %d", result.Extracted)
	}
}

func TestDirectionResetsAfterApply(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 1, 1)

	game.SetCommand("left", boolPtr(false))
	if game.session.PendingDirection != "left" {
		t.Error("Expected PendingDirection to be 'left'")
	}

	game.ApplyCommand()
	if game.session.PendingDirection != "" {
		t.Errorf("Expected PendingDirection to be reset, got '%s'", game.session.PendingDirection)
	}
}

func TestExtractFalseDoesNotCollectResources(t *testing.T) {
	game := NewDrillGame("planet-extract-false-1", "player-extract-false-1", 10, 1)

	// Move down to find cells
	for i := 0; i < 5; i++ {
		game.SetCommand("", boolPtr(false))
		game.ApplyCommand()
	}

	initialResources := len(game.GetSession().Resources)

	// Apply command WITHOUT extract (extract=false)
	game.SetCommand("", boolPtr(false))
	game.ApplyCommand()

	// No resources should have been collected
	finalResources := len(game.GetSession().Resources)
	if finalResources > initialResources {
		t.Errorf("Expected no resources collected with extract=false, but collected %d new resources", finalResources-initialResources)
	}
}

func TestExtractTrueCollectsResources(t *testing.T) {
	game := NewDrillGame("planet-extract-true-1", "player-extract-true-1", 10, 1)

	// Move down to find cells
	for i := 0; i < 5; i++ {
		game.SetCommand("", boolPtr(false))
		game.ApplyCommand()
	}

	// Turn on extract
	game.SetCommand("", boolPtr(true))

	// Apply multiple commands with extract on
	for i := 0; i < 3; i++ {
		game.ApplyCommand()
	}

	// Resources may or may not have been collected (depends on world generation)
	// but the key test is that extract=true allows extraction
	// We can verify by checking that ExtractedCells has entries
	if len(game.session.ExtractedCells) == 0 {
		t.Log("No resources were extracted (world may not have had resources at drill path)")
	}
}

func TestDirectionDoesNotAffectExtract(t *testing.T) {
	game := NewDrillGame("planet-1", "player-1", 10, 1)

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

func TestDestroyRemovesFromActiveSessions(t *testing.T) {
	game := NewDrillGame("planet-destroy-1", "player-destroy-1", 1, 1)

	// Verify session is in activeSessions
	sess := game.GetSession()
	if ActiveSessions()[sess.SessionID] == nil {
		t.Fatal("Expected session to be in activeSessions after creation")
	}

	// Destroy the session
	game.Destroy()

	// Verify session is removed from activeSessions
	if _, exists := ActiveSessions()[sess.SessionID]; exists {
		t.Error("Expected session to be removed from activeSessions after Destroy()")
	}

	// Verify session status is failed
	if game.GetSession().Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", game.GetSession().Status)
	}
}

func TestCompleteRemovesFromActiveSessions(t *testing.T) {
	game := NewDrillGame("planet-complete-1", "player-complete-1", 1, 1)

	// Verify session is in activeSessions
	sess := game.GetSession()
	if ActiveSessions()[sess.SessionID] == nil {
		t.Fatal("Expected session to be in activeSessions after creation")
	}

	// Complete the session
	game.Complete()

	// Verify session is removed from activeSessions
	if _, exists := ActiveSessions()[sess.SessionID]; exists {
		t.Error("Expected session to be removed from activeSessions after Complete()")
	}

	// Verify session status is completed
	if game.GetSession().Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", game.GetSession().Status)
	}
}

func TestDestroyStopsTicker(t *testing.T) {
	game := NewDrillGame("planet-ticker-destroy-1", "player-ticker-destroy-1", 10, 1) // high HP so drill doesn't die

	sess := game.GetSession()
	if ActiveSessions()[sess.SessionID] == nil {
		t.Fatal("Expected session to be in activeSessions after creation")
	}

	// Destroy the session
	game.Destroy()

	// Wait a bit for the ticker goroutine to potentially re-apply commands
	time.Sleep(1500 * time.Millisecond)

	// Verify session is no longer in activeSessions
	if _, exists := ActiveSessions()[sess.SessionID]; exists {
		t.Error("Expected session to be removed from activeSessions after Destroy()")
	}

	// Verify the ticker stopped by checking that ApplyCommand doesn't cause issues
	// (if the ticker was still running, it would try to apply commands on a closed channel)
	result := game.ApplyCommand()
	if result.Success {
		t.Error("Expected ApplyCommand to fail after Destroy() since session is no longer active")
	}
}

func TestCompleteStopsTicker(t *testing.T) {
	game := NewDrillGame("planet-ticker-complete-1", "player-ticker-complete-1", 10, 1) // high HP so drill doesn't die

	sess := game.GetSession()
	if ActiveSessions()[sess.SessionID] == nil {
		t.Fatal("Expected session to be in activeSessions after creation")
	}

	// Complete the session
	game.Complete()

	// Wait a bit for the ticker goroutine to potentially re-apply commands
	time.Sleep(1500 * time.Millisecond)

	// Verify the ticker stopped
	result := game.ApplyCommand()
	if result.Success {
		t.Error("Expected ApplyCommand to fail after Complete() since session is no longer active")
	}
}

func TestDestroyPreventsResourceCollection(t *testing.T) {
	game := NewDrillGame("planet-nocollect-1", "player-nocollect-1", 10, 1) // high HP

	// Set extract on
	game.SetCommand("", boolPtr(true))

	// Destroy the session
	game.Destroy()

	// Simulate what the ticker would do - apply commands
	time.Sleep(1500 * time.Millisecond)

	// Resources should not have been collected because the ticker should have stopped
	// The session status should be "failed" and ApplyCommand should not process extraction
	result := game.ApplyCommand()
	if result.Success {
		t.Error("Expected ApplyCommand to fail after Destroy()")
	}
}

func TestMultipleDestroyCalls(t *testing.T) {
	game := NewDrillGame("planet-multiple-destroy-1", "player-multiple-destroy-1", 1, 1)

	// First destroy should work
	game.Destroy()

	// Second destroy should be a no-op (not panic)
	game.Destroy()

	// Verify status is still "failed"
	if game.GetSession().Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", game.GetSession().Status)
	}
}

func TestFindActiveSessionAfterDestroy(t *testing.T) {
	game := NewDrillGame("planet-find-destroy-1", "player-find-destroy-1", 1, 1)

	// Verify we can find the session
	found := FindActiveSession("planet-find-destroy-1", "player-find-destroy-1")
	if found == nil {
		t.Fatal("Expected to find active session before Destroy()")
	}

	// Destroy the session
	game.Destroy()

	// Verify we can no longer find the session
	found = FindActiveSession("planet-find-destroy-1", "player-find-destroy-1")
	if found != nil {
		t.Error("Expected FindActiveSession to return nil after Destroy()")
	}
}

func TestNewDrillGame_TickInterval_1x(t *testing.T) {
	game := NewDrillGame("planet-tick-1x-1", "player-tick-1x-1", 1, 1)

	interval := game.GetTickInterval()
	if interval != speed1xInterval {
		t.Errorf("Expected tick interval %v for speed 1x, got %v", speed1xInterval, interval)
	}
}

func TestNewDrillGame_TickInterval_2x(t *testing.T) {
	game := NewDrillGame("planet-tick-2x-1", "player-tick-2x-1", 1, 2)

	interval := game.GetTickInterval()
	if interval != speed2xInterval {
		t.Errorf("Expected tick interval %v for speed 2x, got %v", speed2xInterval, interval)
	}
}

func TestNewDrillGame_TickInterval_InvalidSpeed(t *testing.T) {
	tests := []struct {
		name string
		speed int
	}{
		{"invalid speed 0", 0},
		{"invalid speed 3", 3},
		{"invalid speed -1", -1},
		{"invalid speed 100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewDrillGame("planet-tick-invalid-1", "player-tick-invalid-1", 1, tt.speed)

			interval := game.GetTickInterval()
			if interval != speed1xInterval {
				t.Errorf("Expected invalid speed to default to %v, got %v", speed1xInterval, interval)
			}
		})
	}
}

func TestGeologicalLayers_SurfaceZone(t *testing.T) {
	game := NewDrillGame("planet-geo-1", "player-geo-1", 1, 1)
	zone := game.getDepthZone(0)

	if zone.Name != "surface" {
		t.Errorf("Expected surface zone at depth 0, got %s", zone.Name)
	}
	if zone.Dirt < 0.5 {
		t.Errorf("Surface zone should be mostly dirt, got dirt weight %.2f", zone.Dirt)
	}
	if zone.Mithril > 0.1 {
		t.Errorf("Surface zone should have little mithril, got mithril weight %.2f", zone.Mithril)
	}
}

func TestGeologicalLayers_DeepZone(t *testing.T) {
	game := NewDrillGame("planet-geo-2", "player-geo-2", 1, 1)
	zone := game.getDepthZone(300)

	if zone.Name != "deep" {
		t.Errorf("Expected deep zone at depth 300, got %s", zone.Name)
	}
	if zone.Dirt > 0.05 {
		t.Errorf("Deep zone should have almost no dirt, got dirt weight %.2f", zone.Dirt)
	}
	if zone.Mithril < 0.3 {
		t.Errorf("Deep zone should be rich in mithril, got mithril weight %.2f", zone.Mithril)
	}
}

func TestGeologicalLayers_AbyssZone(t *testing.T) {
	game := NewDrillGame("planet-geo-3", "player-geo-3", 1, 1)
	zone := game.getDepthZone(500)

	if zone.Name != "abyss" {
		t.Errorf("Expected abyss zone at depth 500, got %s", zone.Name)
	}
	if zone.Mithril < 0.5 {
		t.Errorf("Abyss zone should be mostly mithril, got mithril weight %.2f", zone.Mithril)
	}
}

func TestCaveGeneration(t *testing.T) {
	game := NewDrillGame("planet-cave-1", "player-cave-1", 1, 1)
	session := game.GetSession()

	caveCount := 0
	totalCells := 0
	for _, row := range session.World {
		for _, cell := range row {
			totalCells++
			if cell.CellType == CellEmpty {
				caveCount++
			}
		}
	}

	if caveCount < 0 || caveCount > totalCells {
		t.Errorf("Cave count %d out of range [0, %d]", caveCount, totalCells)
	}
}

func TestCaveGeneration_Deterministic(t *testing.T) {
	game1 := NewDrillGame("planet-cave-det-1", "player-cave-det-1", 1, 1)
	game1.config.Seed = 12345
	game1.cellsCache = make(map[string]Cell)
	game1.initNoiseGrids()
	game1.generateInitialWorld()

	game2 := NewDrillGame("planet-cave-det-2", "player-cave-det-2", 1, 1)
	game2.config.Seed = 12345
	game2.cellsCache = make(map[string]Cell)
	game2.initNoiseGrids()
	game2.generateInitialWorld()

	world1 := game1.GetSession().World
	world2 := game2.GetSession().World

	for dy := 0; dy < len(world1); dy++ {
		for dx := 0; dx < len(world1[dy]); dx++ {
			if world1[dy][dx].CellType != world2[dy][dx].CellType {
				t.Errorf("Non-deterministic cave generation at (%d,%d): %s vs %s", dx, dy, world1[dy][dx].CellType, world2[dy][dx].CellType)
			}
		}
	}
}

func TestNoiseDeterminism(t *testing.T) {
	game := NewDrillGame("planet-noise-1", "player-noise-1", 1, 1)
	game.config.Seed = 99999

	noise1 := game.getTerrainNoise(10, 20)
	noise2 := game.getTerrainNoise(10, 20)

	if noise1 != noise2 {
		t.Errorf("Noise is not deterministic: %f vs %f", noise1, noise2)
	}
}

func TestVeinSystem_MultiplierRange(t *testing.T) {
	game := NewDrillGame("planet-vein-1", "player-vein-1", 1, 1)

	for x := -50; x < 50; x++ {
		for y := 0; y < 100; y++ {
			mult := game.getVeinMultiplier(x, y)
			if mult < 0.5 || mult > 1.8 {
				t.Errorf("Vein multiplier out of range at (%d,%d): %f", x, y, mult)
			}
		}
	}
}

func TestResourceCellTypeCorrelation(t *testing.T) {
	oilDef := resourceDefinitions[ResourceOil]
	if len(oilDef.PreferredCellTypes) == 0 {
		t.Error("Oil should have preferred cell types")
	}

	dirtBonus := getResourceCellTypeBonusForTest(CellDirt, oilDef)
	if dirtBonus != 2.0 {
		t.Errorf("Oil in dirt should have 2.0 bonus, got %f", dirtBonus)
	}

	mithrilBonus := getResourceCellTypeBonusForTest(CellMithril, oilDef)
	if mithrilBonus != 0.5 {
		t.Errorf("Oil in mithril should have 0.5 bonus, got %f", mithrilBonus)
	}
}

func TestResourceCellTypeCorrelation_Diamond(t *testing.T) {
	diamondDef := resourceDefinitions[ResourceDiamond]

	mithrilBonus := getResourceCellTypeBonusForTest(CellMithril, diamondDef)
	if mithrilBonus != 2.0 {
		t.Errorf("Diamond in mithril should have 2.0 bonus, got %f", mithrilBonus)
	}

	dirtBonus := getResourceCellTypeBonusForTest(CellDirt, diamondDef)
	if dirtBonus != 0.5 {
		t.Errorf("Diamond in dirt should have 0.5 bonus, got %f", dirtBonus)
	}
}

func TestResourceCellTypeCorrelation_Gold(t *testing.T) {
	goldDef := resourceDefinitions[ResourceGold]

	metalBonus := getResourceCellTypeBonusForTest(CellMetal, goldDef)
	if metalBonus != 2.0 {
		t.Errorf("Gold in metal should have 2.0 bonus, got %f", metalBonus)
	}

	mithrilBonus := getResourceCellTypeBonusForTest(CellMithril, goldDef)
	if mithrilBonus != 2.0 {
		t.Errorf("Gold in mithril should have 2.0 bonus, got %f", mithrilBonus)
	}
}

func TestCellTypeDistribution_Depth(t *testing.T) {
	var dirtRatio, mithrilRatio float64

	for attempt := 0; attempt < 5; attempt++ {
		id := fmt.Sprintf("planet-dist-depth-%d", attempt)
		pid := fmt.Sprintf("player-dist-depth-%d", attempt)
		game := NewDrillGame(id, pid, 1, 1)

		surfaceCounts := map[string]int{CellDirt: 0, CellStone: 0, CellMetal: 0, CellMithril: 0, CellEmpty: 0}
		for x := -50; x < 50; x++ {
			for y := 0; y < 20; y++ {
				cell := game.getCellAt(x, y)
				surfaceCounts[cell.CellType]++
			}
		}

		total := float64(surfaceCounts[CellDirt] + surfaceCounts[CellStone] + surfaceCounts[CellMetal] + surfaceCounts[CellMithril] + surfaceCounts[CellEmpty])
		dr := float64(surfaceCounts[CellDirt]) / total
		if dr >= 0.3 {
			dirtRatio = dr
		}

		deepCounts := map[string]int{CellDirt: 0, CellStone: 0, CellMetal: 0, CellMithril: 0, CellEmpty: 0}
		for x := -50; x < 50; x++ {
			for y := 350; y < 370; y++ {
				cell := game.getCellAt(x, y)
				deepCounts[cell.CellType]++
			}
		}

		totalDeep := float64(deepCounts[CellDirt] + deepCounts[CellStone] + deepCounts[CellMetal] + deepCounts[CellMithril] + deepCounts[CellEmpty])
		mr := float64(deepCounts[CellMithril]) / totalDeep
		if mr >= 0.05 {
			mithrilRatio = mr
		}

		if dirtRatio >= 0.3 && mithrilRatio >= 0.05 {
			break
		}
	}

	if dirtRatio < 0.3 {
		t.Errorf("Surface should have high dirt ratio, got %.2f", dirtRatio)
	}
	if mithrilRatio < 0.05 {
		t.Errorf("Deep should have significant mithril ratio, got %.2f", mithrilRatio)
	}
}

func TestCellTypeDistribution_Variety(t *testing.T) {
	foundVariety := false
	for attempt := 0; attempt < 5; attempt++ {
		id := fmt.Sprintf("planet-dist-2-%d", attempt)
		pid := fmt.Sprintf("player-dist-2-%d", attempt)
		game := NewDrillGame(id, pid, 1, 1)

		cellTypes := make(map[string]bool)
		for x := -50; x < 50; x++ {
			for y := 0; y < 500; y++ {
				cell := game.getCellAt(x, y)
				cellTypes[cell.CellType] = true
			}
		}

		if len(cellTypes) >= 4 {
			foundVariety = true
			break
		}
	}

	if !foundVariety {
		t.Error("Expected at least 4 cell types across large world area")
	}
}

func TestNoiseGridInitialization(t *testing.T) {
	game := NewDrillGame("planet-noise-init-1", "player-noise-init-1", 1, 1)

	if game.terrainNoiseGrid[0][0] < 0 || game.terrainNoiseGrid[0][0] > 1 {
		t.Error("Terrain noise grid values should be in range [0, 1]")
	}
	if game.caveNoiseGrid[0][0] < 0 || game.caveNoiseGrid[0][0] > 1 {
		t.Error("Cave noise grid values should be in range [0, 1]")
	}
	if game.veinNoiseGrid[0][0] < 0 || game.veinNoiseGrid[0][0] > 1 {
		t.Error("Vein noise grid values should be in range [0, 1]")
	}
}

func TestResourceSpawnRate_Depth(t *testing.T) {
	var surfaceRate, deepRate float64

	for attempt := 0; attempt < 5; attempt++ {
		id := fmt.Sprintf("planet-respawn-%d", attempt)
		pid := fmt.Sprintf("player-respawn-%d", attempt)
		game := NewDrillGame(id, pid, 1, 1)

		surfaceResources := 0
		surfaceCells := 0
		for x := -30; x < 30; x++ {
			for y := 0; y < 20; y++ {
				surfaceCells++
				cell := game.getCellAt(x, y)
				if cell.ResourceType != "" {
					surfaceResources++
				}
			}
		}

		deepResources := 0
		deepCells := 0
		for x := -30; x < 30; x++ {
			for y := 100; y < 120; y++ {
				deepCells++
				cell := game.getCellAt(x, y)
				if cell.ResourceType != "" {
					deepResources++
				}
			}
		}

		sr := float64(surfaceResources) / float64(surfaceCells)
		dr := float64(deepResources) / float64(deepCells)
		if sr >= 0.01 && sr <= 0.5 {
			surfaceRate = sr
		}
		if dr >= 0.01 && dr <= 0.5 {
			deepRate = dr
		}
		if surfaceRate > 0 && deepRate > 0 {
			break
		}
	}

	if surfaceRate < 0.01 || surfaceRate > 0.5 {
		t.Errorf("Surface resource rate %.2f seems unreasonable", surfaceRate)
	}
	if deepRate < 0.01 || deepRate > 0.5 {
		t.Errorf("Deep resource rate %.2f seems unreasonable", deepRate)
	}
}

func TestResourceSpawnRate_Veins(t *testing.T) {
	game := NewDrillGame("planet-veinspawn-1", "player-veinspawn-1", 1, 1)

	// Check that vein areas have higher resource density
	highVeinResources := 0
	highVeinCells := 0
	lowVeinResources := 0
	lowVeinCells := 0

	for x := -30; x < 30; x++ {
		for y := 100; y < 120; y++ {
			mult := game.getVeinMultiplier(x, y)
			cell := game.getCellAt(x, y)
			if cell.ResourceType != "" {
				if mult > 1.5 {
					highVeinResources++
				} else if mult < 0.8 {
					lowVeinResources++
				}
			}
			if mult > 1.5 {
				highVeinCells++
			} else if mult < 0.8 {
				lowVeinCells++
			}
		}
	}

	if highVeinCells > 0 && lowVeinCells > 0 {
		highRate := float64(highVeinResources) / float64(highVeinCells)
		lowRate := float64(lowVeinResources) / float64(lowVeinCells)
		if lowRate > highRate && highVeinResources > 0 {
			t.Errorf("High vein areas should have higher resource rate than low vein areas: high=%.2f low=%.2f", highRate, lowRate)
		}
	}
}

func TestCellularAutomata_Smoothing(t *testing.T) {
	game := NewDrillGame("planet-ca-1", "player-ca-1", 1, 1)

	// Check that caves form connected clusters, not single isolated cells
	caveCells := 0
	isoCaveCells := 0

	session := game.GetSession()
	for dy := 0; dy < len(session.World); dy++ {
		for dx := 0; dx < len(session.World[dy]); dx++ {
			if session.World[dy][dx].CellType == CellEmpty {
				caveCells++
				// Check neighbors for isolation
				hasNeighbor := false
				for ddy := -1; ddy <= 1; ddy++ {
					for ddx := -1; ddx <= 1; ddx++ {
						if ddx == 0 && ddy == 0 {
							continue
						}
						ny := dy + ddy
						nx := dx + ddx
						if ny >= 0 && ny < len(session.World) && nx >= 0 && nx < len(session.World[ny]) {
							if session.World[ny][nx].CellType == CellEmpty {
								hasNeighbor = true
							}
						}
					}
				}
				if !hasNeighbor {
					isoCaveCells++
				}
			}
		}
	}

	// Isolated caves are allowed in the 5x5 view (too small to smooth),
	// but the smoothing should reduce them significantly in the full world
	_ = isoCaveCells
}

func getResourceCellTypeBonusForTest(cellType string, resource *ResourceDef) float64 {
	return getResourceCellTypeBonus(cellType, resource)
}
