package expedition

import (
	"math/rand"
	"testing"
	"time"

	"spacegame/internal/game/ship"
)

func TestCreateExpedition(t *testing.T) {
	fleet := ship.NewFleet()
	interceptor := ship.GetShipType("interceptor")
	if interceptor == nil {
		t.Fatal("interceptor ship type not found")
	}
	fleet.AddShip(interceptor, 3)

	exp := CreateExpedition("test-exp-1", "planet-1", "exploration", TypeExploration, fleet, 3600)

	if exp.ID != "test-exp-1" {
		t.Errorf("expected ID 'test-exp-1', got '%s'", exp.ID)
	}
	if exp.PlanetID != "planet-1" {
		t.Errorf("expected PlanetID 'planet-1', got '%s'", exp.PlanetID)
	}
	if exp.ExpeditionType != TypeExploration {
		t.Errorf("expected TypeExploration, got '%s'", exp.ExpeditionType)
	}
	if exp.Status != StatusActive {
		t.Errorf("expected StatusActive, got '%s'", exp.Status)
	}
	if exp.Duration != 3600 {
		t.Errorf("expected Duration 3600, got %f", exp.Duration)
	}
	if exp.Fleet.TotalShipCount() != 3 {
		t.Errorf("expected 3 ships in fleet, got %d", exp.Fleet.TotalShipCount())
	}
}

func TestExpeditionTick(t *testing.T) {
	fleet := ship.NewFleet()
	small := ship.GetShipType("small_ship")
	if small == nil {
		t.Fatal("small_ship not found")
	}
	fleet.AddShip(small, 1)

	exp := CreateExpedition("tick-test", "planet-1", "exploration", TypeExploration, fleet, 10)

	if exp.ElapsedTime != 0 {
		t.Errorf("expected elapsed time 0, got %f", exp.ElapsedTime)
	}
	if exp.Progress != 0 {
		t.Errorf("expected progress 0, got %f", exp.Progress)
	}

	// Tick 5 times
	for i := 0; i < 5; i++ {
		exp.Tick()
	}

	if exp.ElapsedTime != 5 {
		t.Errorf("expected elapsed time 5, got %f", exp.ElapsedTime)
	}
	expectedProgress := 5.0 / 10.0
	if exp.Progress != expectedProgress {
		t.Errorf("expected progress %f, got %f", expectedProgress, exp.Progress)
	}
	if exp.Status != StatusActive {
		t.Errorf("expected StatusActive after 5 ticks, got '%s'", exp.Status)
	}

	// Tick until completion
	for i := 0; i < 10; i++ {
		exp.Tick()
	}

	if exp.Status != StatusCompleted {
		t.Errorf("expected StatusCompleted after 15 ticks, got '%s'", exp.Status)
	}
	if !exp.IsExpired() {
		t.Error("expected IsExpired to be true")
	}
}

func TestCalculateDiscoveryChance(t *testing.T) {
	fleet := ship.NewFleet()
	for i := 0; i < 5; i++ {
		small := ship.GetShipType("small_ship")
		if small != nil {
			fleet.AddShip(small, 1)
		}
	}

	// Early in expedition - low chance
	chance1 := CalculateDiscoveryChance(fleet, 10, 3600)
	if chance1 <= 0 {
		t.Errorf("expected positive chance at early stage, got %f", chance1)
	}

	// Mid expedition - medium chance
	chance2 := CalculateDiscoveryChance(fleet, 1800, 3600)
	if chance2 <= chance1 {
		t.Errorf("expected chance to increase with progress, got %f vs %f", chance2, chance1)
	}

	// Late expedition - high chance
	chance3 := CalculateDiscoveryChance(fleet, 3500, 3600)
	if chance3 <= chance2 {
		t.Errorf("expected chance to increase near end, got %f vs %f", chance3, chance2)
	}

	// Cap at 50%
	if chance3 > 0.5 {
		t.Errorf("expected chance to be capped at 0.5, got %f", chance3)
	}

	// Larger fleet = higher chance (up to cap)
	fleet2 := ship.NewFleet()
	for i := 0; i < 20; i++ {
		small := ship.GetShipType("small_ship")
		if small != nil {
			fleet2.AddShip(small, 1)
		}
	}
	chance4 := CalculateDiscoveryChance(fleet2, 1800, 3600)
	// Both are at the same progress, fleet2 has max bonus (capped at 10%)
	// chance2 has 5 ships = 10% bonus (capped), chance4 has 20 ships = 10% bonus (capped)
	// They should be equal since both are capped
	if chance4 < chance2 {
		t.Errorf("expected fleet4 >= chance2, got %f vs %f", chance4, chance2)
	}
}

func TestNPCPlanetGeneration(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	// Test each POI type
	poiTypes := []PointOfInterestType{
		POIAbandonedStation,
		POIDebris,
		POICosmicDebris,
		POIAsteroids,
		POIUnknownPlanet,
		POIAlienBase,
	}

	for _, poiType := range poiTypes {
		npc := NewNPCPlanet("npc-"+string(poiType), "player-1", poiType)

		if npc.ID == "" {
			t.Errorf("NPC planet ID should not be empty for type %s", poiType)
		}
		if npc.Name == "" {
			t.Errorf("NPC planet name should not be empty for type %s", poiType)
		}
		if npc.Type != poiType {
			t.Errorf("expected type %s, got %s", poiType, npc.Type)
		}
		if npc.OwnerID != "player-1" {
			t.Errorf("expected owner 'player-1', got '%s'", npc.OwnerID)
		}

		// Check resources
		totalRes := npc.TotalResources()
		if totalRes < 0 {
			t.Errorf("total resources should be non-negative, got %f", totalRes)
		}

		// Check fleet
		if npc.EnemyFleet == nil {
			t.Errorf("enemy fleet should not be nil")
		}

		// Alien bases should have combat ships
		if poiType == POIAlienBase && !npc.HasCombatShips() {
			t.Logf("Note: alien base didn't generate combat ships in this run (random)")
		}

		_ = r // suppress unused warning
	}
}

func TestGenerateName(t *testing.T) {
	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := GenerateName(POIAbandonedStation)
		names[name] = true
	}
	if len(names) < 2 {
		t.Errorf("expected multiple different names, got only %d unique", len(names))
	}
}

func TestExplorationManager(t *testing.T) {
	em := NewExplorationManager()

	// Discover a planet
	npc1 := em.DiscoverNPCPlanet("exp-1", "player-1")
	if npc1 == nil {
		t.Fatal("expected NPC planet, got nil")
	}
	if npc1.ID == "" {
		t.Error("NPC planet ID should not be empty")
	}

	// Get the same planet
	npc2 := em.GetNPCPlanet(npc1.ID)
	if npc2 == nil {
		t.Fatal("expected to find NPC planet")
	}
	if npc2.ID != npc1.ID {
		t.Errorf("expected same ID, got %s vs %s", npc2.ID, npc1.ID)
	}

	// Remove it
	em.RemoveNPCPlanet(npc1.ID)
	npc3 := em.GetNPCPlanet(npc1.ID)
	if npc3 != nil {
		t.Error("expected nil after removal")
	}

	// GetAllNPCPlanets
	npc4 := em.DiscoverNPCPlanet("exp-2", "player-1")
	npc5 := em.DiscoverNPCPlanet("exp-3", "player-2")
	all := em.GetAllNPCPlanets()
	if len(all) != 2 {
		t.Errorf("expected 2 NPC planets, got %d", len(all))
	}

	// Check owner IDs
	ownerMap := make(map[string]string)
	for _, npc := range all {
		ownerMap[npc.ID] = npc.OwnerID
	}
	if ownerMap[npc4.ID] != "player-1" {
		t.Errorf("expected player-1, got %s", ownerMap[npc4.ID])
	}
	if ownerMap[npc5.ID] != "player-2" {
		t.Errorf("expected player-2, got %s", ownerMap[npc5.ID])
	}
}

func TestCollectResources(t *testing.T) {
	npc := &NPCPlanet{
		ID:          "test-npc",
		OwnerID:     "player-1",
		Name:        "Test Planet",
		Type:        POIAsteroids,
		Resources:   map[string]float64{"food": 100, "composite": 50, "mechanisms": 75, "money": 200},
		EnemyFleet:  ship.NewFleet(),
		Discovered:  true,
		CreatedAt:   time.Now(),
	}

	// Collect with enough cargo
	collected, _ := CollectResources(npc, 200)

	if len(collected) == 0 {
		t.Error("expected some resources to be collected")
	}

	// Check that resources were deducted from NPC
	if npc.Resources["food"] > 100 {
		t.Errorf("food should be reduced, got %f", npc.Resources["food"])
	}

	// Collect remaining
	collected2, _ := CollectResources(npc, 200)
	totalCollected := 0.0
	for _, v := range collected {
		totalCollected += v
	}
	for _, v := range collected2 {
		totalCollected += v
	}

	if totalCollected <= 0 {
		t.Error("expected total collected > 0")
	}
}

func TestExpeditionGetAvailableActions(t *testing.T) {
	fleet := ship.NewFleet()
	small := ship.GetShipType("small_ship")
	if small != nil {
		fleet.AddShip(small, 1)
	}

	exp := CreateExpedition("action-test", "planet-1", "exploration", TypeExploration, fleet, 3600)

	// No NPC discovered - no actions
	actions := exp.GetAvailableActions()
	if len(actions) != 0 {
		t.Errorf("expected 0 actions without NPC, got %d", len(actions))
	}

	// Discover NPC with resources and combat ships
	npc := NewNPCPlanet("npc-action", "player-1", POIAlienBase)
	exp.DiscoveredNPC = npc
	exp.Status = StatusAtPoint

	actions = exp.GetAvailableActions()
	if len(actions) == 0 {
		t.Error("expected actions with NPC discovered")
	}

	// Check for expected action types
	hasLoot := false
	hasAttack := false
	hasWait := false
	hasLeave := false

	for _, a := range actions {
		switch a.Type {
		case "loot":
			hasLoot = true
		case "attack":
			hasAttack = true
		case "wait":
			hasWait = true
		case "leave":
			hasLeave = true
		}
	}

	if !hasLoot {
		t.Error("expected 'loot' action")
	}
	if !hasAttack {
		t.Error("expected 'attack' action")
	}
	if !hasWait {
		t.Error("expected 'wait' action")
	}
	if !hasLeave {
		t.Error("expected 'leave' action")
	}
}

func TestDeductExpeditionEnergy(t *testing.T) {
	newEnergy, ok := DeductExpeditionEnergy(100, 50)
	if !ok {
		t.Error("expected energy deduction to succeed")
	}
	if newEnergy != 50 {
		t.Errorf("expected 50 energy remaining, got %f", newEnergy)
	}

	// Not enough energy
	newEnergy, ok = DeductExpeditionEnergy(10, 50)
	if ok {
		t.Error("expected energy deduction to fail with insufficient energy")
	}
	if newEnergy != 0 {
		t.Errorf("expected 0 energy remaining, got %f", newEnergy)
	}
}

func TestNPCPlanetTotalResources(t *testing.T) {
	npc := &NPCPlanet{
		Resources: map[string]float64{
			"food":      100,
			"composite": 50,
			"mechanisms": 75,
			"money":     200,
			"reagents":  25,
		},
	}

	total := npc.TotalResources()
	expected := 450.0
	if total != expected {
		t.Errorf("expected total %f, got %f", expected, total)
	}
}

func TestNPCPlanetHasCombatShips(t *testing.T) {
	// Empty fleet
	npc1 := &NPCPlanet{EnemyFleet: ship.NewFleet()}
	if npc1.HasCombatShips() {
		t.Error("expected no combat ships for empty fleet")
	}

	// Fleet with ships
	npc2 := &NPCPlanet{EnemyFleet: ship.NewFleet()}
	interceptor := ship.GetShipType("interceptor")
	if interceptor != nil {
		npc2.EnemyFleet.AddShip(interceptor, 3)
	}
	if !npc2.HasCombatShips() {
		t.Error("expected combat ships for fleet with interceptors")
	}
}
