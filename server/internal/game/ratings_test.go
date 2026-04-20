package game

import (
	"math/rand"
	"testing"
	"time"

	"spacegame/internal/game/building"
	"spacegame/internal/game/ship"
)

func TestComputePlanetRatingValue(t *testing.T) {
	p := &Planet{
		ID:   "test-planet",
		Name: "Test Planet",
		Resources: PlanetResources{
			Food:      100,
			Composite: 200,
			Mechanisms: 150,
			Reagents:  50,
			Money:     500,
			AlienTech: 25,
		},
	}

	expected := float64(1025)
	actual := ComputePlanetRatingValue(p)

	if actual != expected {
		t.Errorf("Expected rating value %f, got %f", expected, actual)
	}
}

func TestComputePlanetShips(t *testing.T) {
	p := &Planet{
		ID:   "test-planet",
		Name: "Test Planet",
		Fleet: ship.NewFleet(),
	}

	// Fleet with ships
	p.Fleet.AddShip(ship.AllShipTypes()[0], 5)

	actual := ComputePlanetShips(p)
	if actual != 5 {
		t.Errorf("Expected 5 ships, got %f", actual)
	}
}

func TestComputePlanetBuildings(t *testing.T) {
	p := &Planet{
		ID:        "test-planet",
		Name:      "Test Planet",
		Buildings: map[string]int{"farm": 3, "solar": 2, "base": 1},
	}

	actual := ComputePlanetBuildings(p)
	if actual != 6 {
		t.Errorf("Expected 6 buildings, got %f", actual)
	}
}

func TestComputePlanetMoney(t *testing.T) {
	p := &Planet{
		ID: "test-planet",
		Resources: PlanetResources{
			Money: 500,
		},
	}

	actual := ComputePlanetMoney(p)
	if actual != 500 {
		t.Errorf("Expected 500 money, got %f", actual)
	}
}

func TestComputePlanetFood(t *testing.T) {
	p := &Planet{
		ID: "test-planet",
		Resources: PlanetResources{
			Food: 200,
		},
	}

	actual := ComputePlanetFood(p)
	if actual != 200 {
		t.Errorf("Expected 200 food, got %f", actual)
	}
}

func TestGetRandomEvents(t *testing.T) {
	events := GetRandomEvents()

	if len(events) != 4 {
		t.Errorf("Expected 4 random events, got %d", len(events))
	}

	expectedTypes := []RandomEventType{
		RandomEventShortCircuit,
		RandomEventTheft,
		RandomEventStorageCollapse,
		RandomEventMineCollapse,
	}

	for i, event := range events {
		if event.Type != expectedTypes[i] {
			t.Errorf("Event %d: expected type %s, got %s", i, expectedTypes[i], event.Type)
		}
	}
}

func TestApplyShortCircuit(t *testing.T) {
	p := &Planet{
		ID:   "test-planet",
		Name: "Test Planet",
		Resources: PlanetResources{
			Energy:    100,
			MaxEnergy: 100,
		},
	}

	desc, err := applyShortCircuit(p)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if p.Resources.Energy != 0 {
		t.Errorf("Expected energy to be reset to 0, got %f", p.Resources.Energy)
	}

	if p.Resources.MaxEnergy != 0 {
		t.Errorf("Expected max energy to be reset to 0, got %f", p.Resources.MaxEnergy)
	}

	if desc == "" {
		t.Error("Expected non-empty description")
	}
}

func TestApplyTheft(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	p := &Planet{
		ID:   "test-planet",
		Name: "Test Planet",
		Resources: PlanetResources{
			Money: 1000,
		},
	}

	oldMoney := p.Resources.Money
	desc, err := applyTheft(p)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Money should be reduced by 5-20%
	lost := oldMoney - p.Resources.Money
	if lost < 50 || lost > 200 {
		t.Errorf("Expected to lose 50-200 money, lost %f", lost)
	}

	if p.Resources.Money < 0 {
		t.Errorf("Money should not be negative, got %f", p.Resources.Money)
	}

	if desc == "" {
		t.Error("Expected non-empty description")
	}
}

func TestApplyStorageCollapse(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	p := &Planet{
		ID:   "test-planet",
		Name: "Test Planet",
		Resources: PlanetResources{
			Food:      500,
			Composite: 300,
			Mechanisms: 200,
			Reagents:  100,
		},
	}

	oldFood := p.Resources.Food
	desc, err := applyStorageCollapse(p)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// At least one resource should be reduced
	if p.Resources.Food >= oldFood && p.Resources.Composite >= 300 &&
		p.Resources.Mechanisms >= 200 && p.Resources.Reagents >= 100 {
		t.Error("Expected at least one resource to be reduced")
	}

	if desc == "" {
		t.Error("Expected non-empty description")
	}
}

func TestApplyStorageCollapseNoResources(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	p := &Planet{
		ID:   "test-planet",
		Name: "Test Planet",
		Resources: PlanetResources{
			Food:      0,
			Composite: 0,
			Mechanisms: 0,
			Reagents:  0,
		},
	}

	desc, err := applyStorageCollapse(p)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if desc == "" {
		t.Error("Expected non-empty description")
	}
}

func TestApplyMineCollapse(t *testing.T) {
	p := &Planet{
		ID:    "test-planet",
		Name:  "Test Planet",
		Level: 5,
	}

	desc, err := applyMineCollapse(p)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if p.Level != 4 {
		t.Errorf("Expected level to be reduced to 4, got %d", p.Level)
	}

	if desc == "" {
		t.Error("Expected non-empty description")
	}
}

func TestApplyMineCollapseMinimumLevel(t *testing.T) {
	p := &Planet{
		ID:    "test-planet",
		Name:  "Test Planet",
		Level: 1,
	}

	desc, err := applyMineCollapse(p)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if p.Level != 1 {
		t.Errorf("Expected level to stay at 1, got %d", p.Level)
	}

	if desc == "" {
		t.Error("Expected non-empty description")
	}
}

func TestAllRatingCategories(t *testing.T) {
	categories := AllRatingCategories()

	if len(categories) != 5 {
		t.Errorf("Expected 5 rating categories, got %d", len(categories))
	}

	expectedCategories := []RatingCategory{
		RatingMoney,
		RatingFood,
		RatingShips,
		RatingBuildings,
		RatingTotalResources,
	}

	for i, cat := range categories {
		if cat != expectedCategories[i] {
			t.Errorf("Category %d: expected %s, got %s", i, expectedCategories[i], cat)
		}
	}
}

func TestAllStatsKeys(t *testing.T) {
	keys := AllStatsKeys()

	if len(keys) < 30 {
		t.Errorf("Expected at least 30 stats keys, got %d", len(keys))
	}

	// Check for expected keys
	expectedKeys := []StatsKey{
		StatDaysPlayed,
		StatFirstLogin,
		StatLastLogin,
		StatTotalFoodProduce,
		StatTotalBuildings,
		StatTotalResearch,
		StatTotalBattlesWon,
		StatTotalBattlesLost,
		StatTotalExpeditions,
		StatMiningPlayed,
	}

	for _, expected := range expectedKeys {
		found := false
		for _, key := range keys {
			if key == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected stats key %s not found", expected)
		}
	}
}

func TestStatsTrackerCreation(t *testing.T) {
	g := New()
	tracker := NewStatsTracker(g)

	if tracker == nil {
		t.Fatal("Expected non-nil StatsTracker")
	}

	if tracker.game != g {
		t.Error("StatsTracker should reference the Game instance")
	}
}

func TestProductionResult(t *testing.T) {
	prod := ProductionResult{
		Food:      10,
		Composite: 5,
		Mechanisms: 3,
		Reagents:  2,
		Energy:    15,
		Money:     50,
		AlienTech: 1,
		HasEnergy: true,
	}

	if prod.Food != 10 || prod.Composite != 5 || prod.Mechanisms != 3 {
		t.Error("ProductionResult fields not set correctly")
	}

	if !prod.HasEnergy {
		t.Error("HasEnergy should be true")
	}

	if prod.IsZero() {
		t.Error("ProductionResult should not be zero")
	}

	zeroProd := ProductionResult{}
	if !zeroProd.IsZero() {
		t.Error("Empty ProductionResult should be zero")
	}
}

func TestProductionResultAdd(t *testing.T) {
	p1 := ProductionResult{Food: 10, Energy: 5}
	p2 := ProductionResult{Food: 5, Money: 100}

	p1.Add(p2)

	if p1.Food != 15 {
		t.Errorf("Expected Food to be 15, got %f", p1.Food)
	}

	if p1.Energy != 5 {
		t.Errorf("Expected Energy to be 5, got %f", p1.Energy)
	}

	if p1.Money != 100 {
		t.Errorf("Expected Money to be 100, got %f", p1.Money)
	}
}

func TestPlanetResources(t *testing.T) {
	resources := PlanetResources{
		Food:      100,
		Composite: 50,
		Mechanisms: 25,
		Reagents:  75,
		Energy:    200,
		MaxEnergy: 200,
		Money:     500,
		AlienTech: 10,
	}

	if resources.Food != 100 || resources.Money != 500 {
		t.Error("PlanetResources fields not set correctly")
	}
}

func TestRandomEventChanceConfig(t *testing.T) {
	events := GetRandomEvents()

	// Verify that event chances are reasonable (between 0 and 1)
	for _, event := range events {
		if event.Chance < 0 || event.Chance > 1 {
			t.Errorf("Event %s has invalid chance: %f", event.Type, event.Chance)
		}
	}

	// Short circuit should have the lowest chance
	scEvent := events[0]
	if scEvent.Chance != 0.005 {
		t.Errorf("Short circuit chance should be 0.005, got %f", scEvent.Chance)
	}
}

func TestEventDefResolveCosts(t *testing.T) {
	events := GetRandomEvents()

	// Short circuit should have resolve cost
	scEvent := events[0]
	if scEvent.ResolveCost == nil {
		t.Error("Short circuit should have resolve costs")
	}
	if _, ok := scEvent.ResolveCost["money"]; !ok {
		t.Error("Short circuit should require money to resolve")
	}

	// Storage collapse should have resolve cost
	sc2Event := events[2]
	if sc2Event.ResolveCost == nil {
		t.Error("Storage collapse should have resolve costs")
	}

	// Mine collapse should have resolve cost
	mcEvent := events[3]
	if mcEvent.ResolveCost == nil {
		t.Error("Mine collapse should have resolve costs")
	}
}

func TestBuildingProductionResult(t *testing.T) {
	prod := building.ProductionResult{
		Food:      10,
		Composite: 5,
		HasEnergy: true,
	}

	if prod.Food != 10 || !prod.HasEnergy {
		t.Error("building.ProductionResult not set correctly")
	}

	if prod.IsZero() {
		t.Error("ProductionResult with values should not be zero")
	}
}

func TestResourceNames(t *testing.T) {
	resources := AllResources()

	if len(resources) != 7 {
		t.Errorf("Expected 7 resource types, got %d", len(resources))
	}

	expectedNames := []ResourceName{Food, Composite, Mechanisms, Reagents, Energy, Money, AlienTech}
	for i, res := range resources {
		if res.Name != expectedNames[i] {
			t.Errorf("Resource %d: expected %s, got %s", i, expectedNames[i], res.Name)
		}
	}
}

func TestResourceIcons(t *testing.T) {
	icons := ResourceIcons()

	if icons[Food] != "food" {
		t.Errorf("Food icon should be 'food', got '%s'", icons[Food])
	}

	if icons[Money] != "money" {
		t.Errorf("Money icon should be 'money', got '%s'", icons[Money])
	}
}

func TestResourceEmojis(t *testing.T) {
	emojis := ResourceEmojis()

	if emojis[Food] != "🍍" {
		t.Errorf("Food emoji should be '🍍', got '%s'", emojis[Food])
	}

	if emojis[Money] != "💰" {
		t.Errorf("Money emoji should be '💰', got '%s'", emojis[Money])
	}
}

func TestStorageResourceTypes(t *testing.T) {
	types := AllStorageResources()

	if len(types) != 4 {
		t.Errorf("Expected 4 storage resource types, got %d", len(types))
	}
}

func TestStorageResourceEmojis(t *testing.T) {
	emojis := StorageResourceEmojis()

	if emojis[StorageFood] != "🍍" {
		t.Errorf("StorageFood emoji should be '🍍', got '%s'", emojis[StorageFood])
	}
}

func TestResourceTypes(t *testing.T) {
	types := ResourceTypes()

	if len(types) != 4 {
		t.Errorf("Expected 4 production resource types, got %d", len(types))
	}

	expectedTypes := []ResourceName{Food, Composite, Mechanisms, Reagents}
	for i, rt := range types {
		if rt != expectedTypes[i] {
			t.Errorf("Resource type %d: expected %s, got %s", i, expectedTypes[i], rt)
		}
	}
}
