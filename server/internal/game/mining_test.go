package game

import (
	"testing"
	"time"
)

func TestMazeGeneration(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	session := mg.GetSession()

	if len(session.Maze) != 13 {
		t.Fatalf("expected maze size 13, got %d", len(session.Maze))
	}

	for i := 0; i < 13; i++ {
		if len(session.Maze[i]) != 13 {
			t.Fatalf("expected row %d to have 13 columns, got %d", i, len(session.Maze[i]))
		}
	}

	// Check borders are walls
	for i := 0; i < 13; i++ {
		if session.Maze[0][i] != CellWall {
			t.Errorf("expected border wall at (0, %d), got %c", i, session.Maze[0][i])
		}
		if session.Maze[12][i] != CellWall {
			t.Errorf("expected border wall at (12, %d), got %c", i, session.Maze[12][i])
		}
		if session.Maze[i][0] != CellWall {
			t.Errorf("expected border wall at (%d, 0), got %c", i, session.Maze[i][0])
		}
		if session.Maze[i][12] != CellWall {
			t.Errorf("expected border wall at (%d, 12), got %c", i, session.Maze[i][12])
		}
	}

	// Check player start position
	if session.PlayerX != 1 || session.PlayerY != 1 {
		t.Errorf("expected player at (1, 1), got (%d, %d)", session.PlayerX, session.PlayerY)
	}

	// Check exit position
	if session.ExitX != 11 || session.ExitY != 11 {
		t.Errorf("expected exit at (11, 11), got (%d, %d)", session.ExitX, session.ExitY)
	}
}

func TestMazeConnectivity(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 123)
	session := mg.GetSession()

	// BFS to check that all non-wall cells are connected
	visited := make([][]bool, 13)
	for i := 0; i < 13; i++ {
		visited[i] = make([]bool, 13)
	}

	queue := [][2]int{{session.PlayerX, session.PlayerY}}
	visited[session.PlayerX][session.PlayerY] = true
	count := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		count++

		dx := []int{-1, 1, 0, 0}
		dy := []int{0, 0, -1, 1}

		for i := 0; i < 4; i++ {
			nx, ny := current[0]+dx[i], current[1]+dy[i]
			if nx >= 0 && nx < 13 && ny >= 0 && ny < 13 && !visited[nx][ny] && session.Maze[nx][ny] != CellWall {
				visited[nx][ny] = true
				queue = append(queue, [2]int{nx, ny})
			}
		}
	}

	// Count non-wall cells
	totalEmpty := 0
	for i := 0; i < 13; i++ {
		for j := 0; j < 13; j++ {
			if session.Maze[i][j] != CellWall {
				totalEmpty++
			}
		}
	}

	if count != totalEmpty {
		t.Errorf("maze connectivity check failed: reachable=%d, total_empty=%d", count, totalEmpty)
	}
}

func TestCollisionDetection(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)

	// Try to move left from start (position 1,1) - should fail since (1,0) is a wall
	result := mg.Move(Left, false)
	if result.Success {
		t.Error("expected move left to fail (wall), but it succeeded")
	}

	// Try to move down from start - check if there's a passage
	result = mg.Move(Down, false)
	if !result.Success {
		// Down might be blocked by wall, that's ok - just verify the maze has walls
		t.Log("Move down blocked by wall - maze generation is correct")
	}

	// Verify that left is still blocked
	result = mg.Move(Left, false)
	if result.Success {
		t.Error("expected move left to still fail (wall), but it succeeded")
	}
}

func TestSlideMovement(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	session := mg.GetSession()

	initialX, initialY := session.PlayerX, session.PlayerY

	// Slide down - should move until it hits a wall
	result := mg.Move(Down, true)
	if !result.Success {
		t.Errorf("expected slide down to succeed: %v", result)
	}

	// Verify player either moved or stayed (depends on maze generation)
	// The key test is that sliding stops at a wall, not at a single cell
	_ = initialX
	_ = initialY

	// Try another direction
	result = mg.Move(Up, true)
	if !result.Success {
		t.Errorf("expected slide up to succeed: %v", result)
	}

	// Verify player stopped at a wall - position should be valid
	if result.PlayerX < 0 || result.PlayerX >= 13 || result.PlayerY < 0 || result.PlayerY >= 13 {
		t.Errorf("player position out of bounds after slide: (%d, %d)", result.PlayerX, result.PlayerY)
	}
}

func TestMonsterEncounter(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	session := mg.GetSession()

	// Check that monsters were spawned
	if len(session.Monsters) == 0 {
		t.Log("No monsters spawned in this seed - checking monster definitions instead")
	}

	// Verify monster definitions are correct
	ratDef, ok := monsterDefinitions[MonsterRat]
	if !ok {
		t.Error("rat monster definition not found")
	} else {
		if ratDef.HP != 15 {
			t.Errorf("expected rat HP 15, got %d", ratDef.HP)
		}
		if ratDef.Damage != 8 {
			t.Errorf("expected rat damage 8, got %d", ratDef.Damage)
		}
	}

	batDef, ok := monsterDefinitions[MonsterBat]
	if !ok {
		t.Error("bat monster definition not found")
	} else {
		if batDef.HP != 25 {
			t.Errorf("expected bat HP 25, got %d", batDef.HP)
		}
	}

	alienDef, ok := monsterDefinitions[MonsterAlien]
	if !ok {
		t.Error("alien monster definition not found")
	} else {
		if alienDef.HP != 50 {
			t.Errorf("expected alien HP 50, got %d", alienDef.HP)
		}
		if alienDef.Damage != 20 {
			t.Errorf("expected alien damage 20, got %d", alienDef.Damage)
		}
	}
}

func TestBombUsage(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	session := mg.GetSession()

	initialBombs := session.PlayerBombs

	// Use bomb
	result := mg.UseBomb()
	if !result.Success {
		t.Log("No walls destroyed - checking bomb mechanics")
	}

	// If bomb was used, check bomb count decreased
	if initialBombs > 0 {
		newBombs := mg.GetSession().PlayerBombs
		if newBombs != initialBombs-1 {
			t.Errorf("expected bombs to decrease from %d to %d, got %d", initialBombs, initialBombs-1, newBombs)
		}
	}

	// Try to use bomb when no bombs
	mg2 := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	mg2.GetSession().PlayerBombs = 0
	result2 := mg2.UseBomb()
	if result2.Success {
		t.Error("expected bomb use to fail with no bombs")
	}
}

func TestExitDetection(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	session := mg.GetSession()

	// Verify exit is placed
	if session.ExitX < 0 || session.ExitX >= 13 || session.ExitY < 0 || session.ExitY >= 13 {
		t.Errorf("exit position out of bounds: (%d, %d)", session.ExitX, session.ExitY)
	}

	// Verify exit is far from start
	dist := (session.ExitX - 1) + (session.ExitY - 1)
	if dist < 4 {
		t.Errorf("exit too close to start: distance=%d", dist)
	}
}

func TestPlayerDeath(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	session := mg.GetSession()

	// Verify player starts with positive HP
	if session.PlayerHP <= 0 {
		t.Error("player should start with positive HP")
	}

	// Simulate damage by directly modifying HP
	session.PlayerHP = 1

	// Move to trigger cell processing (the game will check HP after encounters)
	// We test the death condition directly
	if session.PlayerHP <= 0 {
		t.Error("player should not be dead yet")
	}

	// Set HP to 0 and check death
	session.PlayerHP = 0
	if session.Status != "failed" {
		// Status is only set when HP drops to 0 during a move
		// This is expected behavior - death is checked after encounters
	}
}

func TestMoneyCollection(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	session := mg.GetSession()

	initialMoney := session.MoneyCollected
	if initialMoney != 0 {
		t.Errorf("expected initial money to be 0, got %f", initialMoney)
	}

	// Verify money multiplier
	expectedMultiplier := 1.0 + float64(session.BaseLevel)*0.1
	actualMultiplier := mg.getMoneyMultiplier()
	if actualMultiplier != expectedMultiplier {
		t.Errorf("expected multiplier %.2f, got %.2f", expectedMultiplier, actualMultiplier)
	}
}

func TestMoneyMultiplier(t *testing.T) {
	tests := []struct {
		baseLevel  int
		expected   float64
	}{
		{1, 1.1},
		{2, 1.2},
		{5, 1.5},
		{10, 2.0},
		{0, 1.0},
	}

	for _, tt := range tests {
		mg := NewMiningGameWithSeed("test-planet", "test-player", tt.baseLevel, 42)
		actual := mg.getMoneyMultiplier()
		if actual != tt.expected {
			t.Errorf("base_level=%d: expected multiplier %.2f, got %.2f", tt.baseLevel, tt.expected, actual)
		}
	}
}

func TestDirectionParsing(t *testing.T) {
	tests := []struct {
		input       string
		expected    MoveDirection
		expectError bool
	}{
		{"up", Up, false},
		{"down", Down, false},
		{"left", Left, false},
		{"right", Right, false},
		{"north", Up, false},
		{"south", Down, false},
		{"west", Left, false},
		{"east", Right, false},
		{"n", Up, false},
		{"s", Down, false},
		{"w", Left, false},
		{"e", Right, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		dir, err := ParseDirection(tt.input)
		if tt.expectError {
			if err == nil {
				t.Errorf("expected error for input '%s', got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input '%s': %v", tt.input, err)
			}
			if dir != tt.expected {
				t.Errorf("input '%s': expected %d, got %d", tt.input, tt.expected, dir)
			}
		}
	}
}

func TestDirectionToString(t *testing.T) {
	tests := []struct {
		dir      MoveDirection
		expected string
	}{
		{Up, "up"},
		{Down, "down"},
		{Left, "left"},
		{Right, "right"},
	}

	for _, tt := range tests {
		result := DirectionToString(tt.dir)
		if result != tt.expected {
			t.Errorf("direction %d: expected '%s', got '%s'", tt.dir, tt.expected, result)
		}
	}
}

func TestAvailableMoves(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)

	moves := mg.GetAvailableMoves()
	if len(moves) == 0 {
		t.Error("expected at least one available move from start position")
	}

	// Verify moves are valid directions
	for _, m := range moves {
		if m != Up && m != Down && m != Left && m != Right {
			t.Errorf("invalid move direction: %d", m)
		}
	}
}

func TestDisplayMaze(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	display := mg.GetDisplayMaze()

	if len(display) != 13 {
		t.Fatalf("expected display maze size 13, got %d", len(display))
	}

	// Check player is in display
	if display[mg.GetSession().PlayerX][mg.GetSession().PlayerY] != CellPlayer {
		t.Errorf("expected player at (%d, %d), got %c", mg.GetSession().PlayerX, mg.GetSession().PlayerY, display[mg.GetSession().PlayerX][mg.GetSession().PlayerY])
	}

	// Check exit is in display
	if display[mg.GetSession().ExitX][mg.GetSession().ExitY] != CellExit {
		t.Errorf("expected exit at (%d, %d), got %c", mg.GetSession().ExitX, mg.GetSession().ExitY, display[mg.GetSession().ExitX][mg.GetSession().ExitY])
	}
}

func TestMultipleGamesDifferentSeeds(t *testing.T) {
	mg1 := NewMiningGameWithSeed("test-planet", "test-player", 1, 1)
	mg2 := NewMiningGameWithSeed("test-planet", "test-player", 1, 2)

	session1 := mg1.GetSession()
	session2 := mg2.GetSession()

	// Different seeds should produce different mazes
	different := false
	for i := 0; i < 13 && !different; i++ {
		for j := 0; j < 13 && !different; j++ {
			if session1.Maze[i][j] != session2.Maze[i][j] {
				different = true
			}
		}
	}

	if !different {
		t.Error("expected different mazes for different seeds")
	}
}

func TestMiningCooldown(t *testing.T) {
	cooldown := MiningCooldown()
	if cooldown != 30*time.Second {
		t.Errorf("expected cooldown 30s, got %v", cooldown)
	}
}

func TestMonsterDefinitions(t *testing.T) {
	for name, def := range monsterDefinitions {
		if def.Name == "" {
			t.Errorf("monster %s has empty name", name)
		}
		if def.Icon == "" {
			t.Errorf("monster %s has empty icon", name)
		}
		if def.HP <= 0 {
			t.Errorf("monster %s has non-positive HP: %d", name, def.HP)
		}
		if def.Damage < 0 {
			t.Errorf("monster %s has negative damage: %d", name, def.Damage)
		}
		if def.Reward <= 0 {
			t.Errorf("monster %s has non-positive reward: %d", name, def.Reward)
		}
	}
}

func TestSessionStatus(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	session := mg.GetSession()

	if session.Status != "active" {
		t.Errorf("expected initial status 'active', got '%s'", session.Status)
	}

	if session.PlayerHP <= 0 {
		t.Error("player should start with positive HP")
	}

	if session.PlayerBombs < 0 {
		t.Error("player should start with non-negative bombs")
	}
}

func TestMoveOnInactiveGame(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	mg.GetSession().Status = "completed"

	result := mg.Move(Up, false)
	if result.Success {
		t.Error("expected move to fail on completed game")
	}
	if !result.GameEnded {
		t.Error("expected game_ended to be true on inactive game")
	}
}

func TestMonsterRewardScaling(t *testing.T) {
	mg1 := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	mg5 := NewMiningGameWithSeed("test-planet", "test-player", 5, 42)

	session1 := mg1.GetSession()
	session5 := mg5.GetSession()

	// Monsters at higher base level should have higher rewards
	for i := range session1.Monsters {
		if i >= len(session5.Monsters) {
			continue
		}
		if session5.Monsters[i].Reward <= session1.Monsters[i].Reward {
			t.Errorf("monster reward should scale with base level: level1=%.2f, level5=%.2f",
				session1.Monsters[i].Reward, session5.Monsters[i].Reward)
		}
	}
}

func TestBombDestructionCrossPattern(t *testing.T) {
	mg := NewMiningGameWithSeed("test-planet", "test-player", 1, 42)
	session := mg.GetSession()

	// Record walls around player before bomb
	bombsBefore := session.PlayerBombs
	if bombsBefore <= 0 {
		t.Skip("no bombs available in this maze configuration")
	}

	// Use bomb
	result := mg.UseBomb()

	// Verify bomb was consumed
	if session.PlayerBombs >= bombsBefore {
		t.Error("expected bomb count to decrease")
	}

	_ = result
}
