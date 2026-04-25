package game

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewGardenBedState(t *testing.T) {
	gardenBed := NewGardenBedState(3)

	if gardenBed.RowCount != 3 {
		t.Errorf("Expected row count 3, got %d", gardenBed.RowCount)
	}
	if len(gardenBed.Rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(gardenBed.Rows))
	}
	for i, row := range gardenBed.Rows {
		if row.Status != GardenBedRowEmpty {
			t.Errorf("Row %d: expected status '%s', got '%s'", i, GardenBedRowEmpty, row.Status)
		}
		if row.Weeds != 0 {
			t.Errorf("Row %d: expected 0 weeds, got %d", i, row.Weeds)
		}
	}
}

func TestNewGardenBedStateFromJSON(t *testing.T) {
	// Test with valid JSON
	rows := []GardenBedRow{
		{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 1, Weeds: 0, WaterTimer: 0},
		{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0},
	}
	data, _ := json.Marshal(rows)

	gardenBed := NewGardenBedStateFromJSON(data, 2)
	if gardenBed.RowCount != 2 {
		t.Errorf("Expected row count 2, got %d", gardenBed.RowCount)
	}
	if len(gardenBed.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(gardenBed.Rows))
	}
	if gardenBed.Rows[0].Status != GardenBedRowPlanted {
		t.Errorf("Expected row 0 status '%s', got '%s'", GardenBedRowPlanted, gardenBed.Rows[0].Status)
	}
	if gardenBed.Rows[0].PlantType != PlantWheat {
		t.Errorf("Expected row 0 plant type '%s', got '%s'", PlantWheat, gardenBed.Rows[0].PlantType)
	}

	// Test with empty JSON
	gardenBedState := NewGardenBedStateFromJSON([]byte("[]"), 3)
	if gardenBedState.RowCount != 3 {
		t.Errorf("Expected row count 3 for empty gardenBed, got %d", gardenBedState.RowCount)
	}

	// Test with null JSON
	gardenBedStateNull := NewGardenBedStateFromJSON([]byte("null"), 2)
	if gardenBedStateNull.RowCount != 2 {
		t.Errorf("Expected row count 2 for null gardenBed, got %d", gardenBedStateNull.RowCount)
	}
}

func TestGardenBedTick_Growth(t *testing.T) {
	gardenBed := NewGardenBedState(1)
	gardenBed.Rows[0] = GardenBedRow{
		Status:     GardenBedRowPlanted,
		PlantType:  PlantWheat,
		Stage:      0,
		Weeds:      0,
		WaterTimer: 0,
		LastTick:   0,
	}

	// First tick at 100: GardenBedTicksSinceLast becomes 1, no advance
	GardenBedTick(gardenBed, 100)
	if gardenBed.Rows[0].Stage != 0 {
		t.Errorf("After tick 100: expected stage 0, got %d", gardenBed.Rows[0].Stage)
	}

	// Tick at 200: GardenBedTicksSinceLast becomes 2, advance to stage 1
	changed := GardenBedTick(gardenBed, 200)
	if !changed {
		t.Error("Expected state change on tick 200")
	}
	if gardenBed.Rows[0].Stage != 1 {
		t.Errorf("After tick 200: expected stage 1, got %d", gardenBed.Rows[0].Stage)
	}

	// Tick at 300: GardenBedTicksSinceLast becomes 1, no advance
	GardenBedTick(gardenBed, 300)
	if gardenBed.Rows[0].Stage != 1 {
		t.Errorf("After tick 300: expected stage 1, got %d", gardenBed.Rows[0].Stage)
	}

	// Tick at 400: GardenBedTicksSinceLast becomes 2, advance to stage 2 (mature)
	changed = GardenBedTick(gardenBed, 400)
	if !changed {
		t.Error("Expected state change on tick 400")
	}
	if gardenBed.Rows[0].Stage != 2 {
		t.Errorf("After tick 400: expected stage 2, got %d", gardenBed.Rows[0].Stage)
	}
	if gardenBed.Rows[0].Status != GardenBedRowMature {
		t.Errorf("After tick 400: expected status '%s', got '%s'", GardenBedRowMature, gardenBed.Rows[0].Status)
	}
}

func TestGardenBedTick_WaterAcceleration(t *testing.T) {
	gardenBed := NewGardenBedState(1)
	gardenBed.Rows[0] = GardenBedRow{
		Status:     GardenBedRowPlanted,
		PlantType:  PlantWheat,
		Stage:      0,
		Weeds:      0,
		WaterTimer: 2,
		LastTick:   0,
	}

	// Tick 100: water_timer > 0, advance stage
	changed := GardenBedTick(gardenBed, 100)
	if !changed {
		t.Error("Expected state change on tick 100 with water")
	}
	if gardenBed.Rows[0].Stage != 1 {
		t.Errorf("After tick 100 with water: expected stage 1, got %d", gardenBed.Rows[0].Stage)
	}
	if gardenBed.Rows[0].WaterTimer != 1 {
		t.Errorf("After tick 100: expected water_timer 1, got %d", gardenBed.Rows[0].WaterTimer)
	}

	// Tick 200: water_timer > 0, advance stage again
	changed = GardenBedTick(gardenBed, 200)
	if !changed {
		t.Error("Expected state change on tick 200 with water")
	}
	if gardenBed.Rows[0].Stage != 2 {
		t.Errorf("After tick 200 with water: expected stage 2, got %d", gardenBed.Rows[0].Stage)
	}
	if gardenBed.Rows[0].WaterTimer != 0 {
		t.Errorf("After tick 200: expected water_timer 0, got %d", gardenBed.Rows[0].WaterTimer)
	}
}

func TestGardenBedTick_NoChangeWhenAlreadyProcessed(t *testing.T) {
	gardenBed := NewGardenBedState(1)
	gardenBed.Rows[0] = GardenBedRow{
		Status:   GardenBedRowPlanted,
		PlantType: PlantWheat,
		Stage:    0,
		Weeds:    0,
		WaterTimer: 0,
		LastTick: 100,
	}

	// Calling GardenBedTick again with same tick should not change anything
	changed := GardenBedTick(gardenBed, 100)
	if changed {
		t.Error("Expected no change when tick already processed")
	}
	if gardenBed.Rows[0].Stage != 0 {
		t.Errorf("Expected stage still 0, got %d", gardenBed.Rows[0].Stage)
	}
}

func TestGardenBedTick_EmptyRow(t *testing.T) {
	gardenBed := NewGardenBedState(1)
	// Row is empty by default

	GardenBedTick(gardenBed, 100)
	// Empty rows only update LastTick, weeds are random
	if gardenBed.Rows[0].LastTick != 100 {
		t.Errorf("Expected LastTick 100, got %d", gardenBed.Rows[0].LastTick)
	}
}

func TestGardenBedTick_MatureRow(t *testing.T) {
	gardenBed := NewGardenBedState(1)
	gardenBed.Rows[0] = GardenBedRow{
		Status:   GardenBedRowMature,
		PlantType: PlantWheat,
		Stage:    2,
		Weeds:    0,
		WaterTimer: 0,
		LastTick: 0,
	}

	changed := GardenBedTick(gardenBed, 100)
	if changed {
		t.Error("Expected no change for mature row")
	}
	if gardenBed.Rows[0].Status != GardenBedRowMature {
		t.Errorf("Expected status still '%s', got '%s'", GardenBedRowMature, gardenBed.Rows[0].Status)
	}
}

func TestGardenBedAction_Plant(t *testing.T) {
	planet := &Planet{
		ID:      "test-planet-plant",
		GardenBedState: NewGardenBedState(1),
		Resources: PlanetResources{Food: 100, Money: 100},
		Buildings: []BuildingEntry{{Type: "farm", Level: 1, Enabled: true}},
	}

	result, err := GardenBedActionInternal(planet, "plant", 0, PlantWheat)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].Status != GardenBedRowPlanted {
		t.Errorf("Expected status '%s', got '%s'", GardenBedRowPlanted, result.Rows[0].Status)
	}
	if result.Rows[0].PlantType != PlantWheat {
		t.Errorf("Expected plant type '%s', got '%s'", PlantWheat, result.Rows[0].PlantType)
	}
	if result.Rows[0].Stage != 0 {
		t.Errorf("Expected stage 0, got %d", result.Rows[0].Stage)
	}
	if result.SeedCost != 5 {
		t.Errorf("Expected seed cost 5, got %.0f", result.SeedCost)
	}
	if planet.Resources.Money != 95 {
		t.Errorf("Expected money 95 after planting, got %.0f", planet.Resources.Money)
	}
}

func TestGardenBedAction_PlantOnNonEmptyRow(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "plant", 0, PlantBerries)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure when planting on non-empty row")
	}
	if result.Error != "Row is not empty" {
		t.Errorf("Expected error 'Row is not empty', got '%s'", result.Error)
	}
}

func TestGardenBedAction_Weed(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 2, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
		Buildings: []BuildingEntry{{Type: "farm", Level: 1, Enabled: true}},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "weed", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].Weeds != 1 {
		t.Errorf("Expected 1 weed after weeding, got %d", result.Rows[0].Weeds)
	}
	if result.FoodCost != 2 {
		t.Errorf("Expected food cost 2 for wheat weed, got %.0f", result.FoodCost)
	}
	if planet.Resources.Food != 98 {
		t.Errorf("Expected 98 food after weeding, got %.0f", planet.Resources.Food)
	}
}

func TestGardenBedAction_WeedMax(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 3, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
		Buildings: []BuildingEntry{{Type: "farm", Level: 1, Enabled: true}},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "weed", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].Weeds != 2 {
		t.Errorf("Expected 2 weeds after weeding from 3, got %d", result.Rows[0].Weeds)
	}
	if result.FoodCost != 2 {
		t.Errorf("Expected food cost 2 for wheat weed, got %.0f", result.FoodCost)
	}
}

func TestGardenBedAction_Water(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
		Buildings: []BuildingEntry{{Type: "farm", Level: 1, Enabled: true}},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "water", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].WaterTimer != 10 {
		t.Errorf("Expected water_timer 10, got %d", result.Rows[0].WaterTimer)
	}
	if result.FoodCost != 1 {
		t.Errorf("Expected food cost 1 for wheat water, got %.0f", result.FoodCost)
	}
	if planet.Resources.Food != 99 {
		t.Errorf("Expected 99 food after watering, got %.0f", planet.Resources.Food)
	}
}

func TestGardenBedAction_Harvest(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowMature, PlantType: PlantWheat, Stage: 2, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100, Money: 50},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "harvest", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].Status != GardenBedRowEmpty {
		t.Errorf("Expected status '%s' after harvest, got '%s'", GardenBedRowEmpty, result.Rows[0].Status)
	}
	if result.FoodGain != 5 {
		t.Errorf("Expected food gain 5, got %.0f", result.FoodGain)
	}
	if result.MoneyGain != 15 {
		t.Errorf("Expected money gain 15, got %.0f", result.MoneyGain)
	}
	if planet.Resources.Food != 105 {
		t.Errorf("Expected 105 food after harvest, got %.0f", planet.Resources.Food)
	}
	if planet.Resources.Money != 65 {
		t.Errorf("Expected 65 money after harvest, got %.0f", planet.Resources.Money)
	}
}

func TestGardenBedAction_HarvestBerries(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowMature, PlantType: PlantBerries, Stage: 2, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100, Money: 50},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "harvest", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.FoodGain != 15 {
		t.Errorf("Expected food gain 15 for berries, got %.0f", result.FoodGain)
	}
	if result.MoneyGain != 45 {
		t.Errorf("Expected money gain 45 for berries, got %.0f", result.MoneyGain)
	}
}

func TestGardenBedAction_HarvestMelon(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowMature, PlantType: PlantMelon, Stage: 2, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100, Money: 50},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "harvest", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.FoodGain != 120 {
		t.Errorf("Expected food gain 120 for melon, got %.0f", result.FoodGain)
	}
	if result.MoneyGain != 800 {
		t.Errorf("Expected money gain 800 for melon, got %.0f", result.MoneyGain)
	}
}

func TestGardenBedAction_HarvestNonMature(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "harvest", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure when harvesting non-mature plant")
	}
	if result.Error != "Plant is not mature" {
		t.Errorf("Expected error 'Plant is not mature', got '%s'", result.Error)
	}
}

func TestGardenBedAction_InvalidRowIndex(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "plant", 5, PlantWheat)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for invalid row index")
	}
	if result.Error != "Invalid row index" {
		t.Errorf("Expected error 'Invalid row index', got '%s'", result.Error)
	}
}

func TestGardenBedAction_UnknownAction(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "dig", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for unknown action")
	}
	if result.Error != "Unknown action: dig" {
		t.Errorf("Expected error 'Unknown action: dig', got '%s'", result.Error)
	}
}

func TestGardenBedAction_Cooldown(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100, Money: 100},
		Buildings: []BuildingEntry{{Type: "farm", Level: 1, Enabled: true}},
	}

	// First action should succeed
	result, _ := GardenBedActionInternal(planet, "plant", 0, PlantWheat)
	if !result.Success {
		t.Errorf("First action should succeed, got error: %s", result.Error)
	}

	// Second action immediately should fail due to cooldown
	result, _ = GardenBedActionInternal(planet, "water", 0, "")
	if result.Success {
		t.Error("Second action should fail due to cooldown")
	}
	if result.Error == "" {
		t.Error("Expected cooldown error message")
	}
}

func TestGardenBedPlantDefinitions(t *testing.T) {
	wheat := gardenBedPlants[PlantWheat]
	if wheat == nil {
		t.Fatal("Wheat plant not found")
	}
	if wheat.FoodReward != 5 {
		t.Errorf("Expected wheat food reward 5, got %.0f", wheat.FoodReward)
	}
	if wheat.Name != "Пшеница" {
		t.Errorf("Expected wheat name 'Пшеница', got '%s'", wheat.Name)
	}
	if wheat.Icon != "🌾" {
		t.Errorf("Expected wheat icon '🌾', got '%s'", wheat.Icon)
	}
	if wheat.SeedCost != 5 {
		t.Errorf("Expected wheat seed cost 5, got %.0f", wheat.SeedCost)
	}
	if wheat.MoneyReward != 15 {
		t.Errorf("Expected wheat money reward 15, got %.0f", wheat.MoneyReward)
	}
	if wheat.UnlockLevel != 1 {
		t.Errorf("Expected wheat unlock level 1, got %d", wheat.UnlockLevel)
	}

	berries := gardenBedPlants[PlantBerries]
	if berries == nil {
		t.Fatal("Berries plant not found")
	}
	if berries.FoodReward != 15 {
		t.Errorf("Expected berries food reward 15, got %.0f", berries.FoodReward)
	}

	melon := gardenBedPlants[PlantMelon]
	if melon == nil {
		t.Fatal("Melon plant not found")
	}
	if melon.FoodReward != 120 {
		t.Errorf("Expected melon food reward 120, got %.0f", melon.FoodReward)
	}
	if melon.MoneyReward != 800 {
		t.Errorf("Expected melon money reward 800, got %.0f", melon.MoneyReward)
	}
	if melon.UnlockLevel != 9 {
		t.Errorf("Expected melon unlock level 9, got %d", melon.UnlockLevel)
	}

	// Check new plants exist
	rose := gardenBedPlants[PlantRose]
	if rose == nil {
		t.Fatal("Rose plant not found")
	}
	if rose.Name != "Космическая роза" {
		t.Errorf("Expected rose name 'Космическая роза', got '%s'", rose.Name)
	}
	if rose.UnlockLevel != 5 {
		t.Errorf("Expected rose unlock level 5, got %d", rose.UnlockLevel)
	}

	banana := gardenBedPlants[PlantBanana]
	if banana == nil {
		t.Fatal("Banana plant not found")
	}
	if banana.Name != "Лунный банан" {
		t.Errorf("Expected banana name 'Лунный банан', got '%s'", banana.Name)
	}
	if banana.UnlockLevel != 11 {
		t.Errorf("Expected banana unlock level 11, got %d", banana.UnlockLevel)
	}
}

func TestGetAllGardenBedPlants(t *testing.T) {
	plants := GetAllGardenBedPlants()
	if len(plants) != 8 {
		t.Errorf("Expected 8 plant types, got %d", len(plants))
	}
}

func TestGetGardenBedPlant_Unknown(t *testing.T) {
	plant := GetGardenBedPlant("unknown")
	if plant != nil {
		t.Error("Expected nil for unknown plant type")
	}
}

func TestGardenBedTick_MultipleRows(t *testing.T) {
	gardenBed := NewGardenBedState(3)

	// Plant different types in different rows
	gardenBed.Rows[0] = GardenBedRow{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 0, WaterTimer: 0, LastTick: 0}
	gardenBed.Rows[1] = GardenBedRow{Status: GardenBedRowPlanted, PlantType: PlantBerries, Stage: 0, Weeds: 0, WaterTimer: 0, LastTick: 0}
	gardenBed.Rows[2] = GardenBedRow{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0}

	// Process ticks
	GardenBedTick(gardenBed, 100)
	GardenBedTick(gardenBed, 200)
	GardenBedTick(gardenBed, 300)
	GardenBedTick(gardenBed, 400)
	GardenBedTick(gardenBed, 500)
	GardenBedTick(gardenBed, 600)

	// Row 0 (wheat) should be mature at stage 2
	if gardenBed.Rows[0].Status != GardenBedRowMature {
		t.Errorf("Row 0: expected status '%s', got '%s'", GardenBedRowMature, gardenBed.Rows[0].Status)
	}
	if gardenBed.Rows[0].Stage != 2 {
		t.Errorf("Row 0: expected stage 2, got %d", gardenBed.Rows[0].Stage)
	}

	// Row 1 (berries) should also be mature
	if gardenBed.Rows[1].Status != GardenBedRowMature {
		t.Errorf("Row 1: expected status '%s', got '%s'", GardenBedRowMature, gardenBed.Rows[1].Status)
	}
	if gardenBed.Rows[1].Stage != 2 {
		t.Errorf("Row 1: expected stage 2, got %d", gardenBed.Rows[1].Stage)
	}

	// Row 2 should still be empty
	if gardenBed.Rows[2].Status != GardenBedRowEmpty {
		t.Errorf("Row 2: expected status '%s', got '%s'", GardenBedRowEmpty, gardenBed.Rows[2].Status)
	}
}

func TestGardenBedTick_WaterTimerExpiry(t *testing.T) {
	gardenBed := NewGardenBedState(1)
	gardenBed.Rows[0] = GardenBedRow{
		Status:     GardenBedRowPlanted,
		PlantType:  PlantWheat,
		Stage:      0,
		Weeds:      0,
		WaterTimer: 1,
		LastTick:   0,
	}

	// Tick 100: water_timer > 0, advance stage, water_timer becomes 0
	changed := GardenBedTick(gardenBed, 100)
	if !changed {
		t.Error("Expected state change")
	}
	if gardenBed.Rows[0].Stage != 1 {
		t.Errorf("After tick 100: expected stage 1, got %d", gardenBed.Rows[0].Stage)
	}
	if gardenBed.Rows[0].WaterTimer != 0 {
		t.Errorf("After tick 100: expected water_timer 0, got %d", gardenBed.Rows[0].WaterTimer)
	}

	// Tick 200: no water, GardenBedTicksSinceLast becomes 1, no advance
	changed = GardenBedTick(gardenBed, 200)
	// changed may be false since water_timer is already 0 and no stage advance
	if gardenBed.Rows[0].Stage != 1 {
		t.Errorf("After tick 200: expected stage 1 (no water), got %d", gardenBed.Rows[0].Stage)
	}

	// Tick 300: GardenBedTicksSinceLast becomes 2, advance to stage 2 (mature)
	changed = GardenBedTick(gardenBed, 300)
	if !changed {
		t.Error("Expected state change")
	}
	if gardenBed.Rows[0].Stage != 2 {
		t.Errorf("After tick 300: expected stage 2, got %d", gardenBed.Rows[0].Stage)
	}
	if gardenBed.Rows[0].Status != GardenBedRowMature {
		t.Errorf("After tick 300: expected status '%s', got '%s'", GardenBedRowMature, gardenBed.Rows[0].Status)
	}
}

func TestGardenBedAction_PlantInvalidType(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "plant", 0, "tomato")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for invalid plant type")
	}
	if result.Error != "Unknown plant type" {
		t.Errorf("Expected error 'Unknown plant type', got '%s'", result.Error)
	}
}

func TestGardenBedAction_WeedEmptyRow(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "weed", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure when weeding empty row")
	}
	if result.Error != "Row is empty" {
		t.Errorf("Expected error 'Row is empty', got '%s'", result.Error)
	}
}

func TestGardenBedAction_WaterEmptyRow(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		GardenBedState: &GardenBedState{
			Rows: []GardenBedRow{
				{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearGardenBedCooldown("test-planet-base")
	result, err := GardenBedActionInternal(planet, "water", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure when watering empty row")
	}
	if result.Error != "Row is empty" {
		t.Errorf("Expected error 'Row is empty', got '%s'", result.Error)
	}
}

func TestGardenBedStateJSONSerialization(t *testing.T) {
	gardenBed := NewGardenBedState(2)
	gardenBed.Rows[0] = GardenBedRow{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 1, Weeds: 1, WaterTimer: 5, LastTick: 100}
	gardenBed.Rows[1] = GardenBedRow{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0, LastTick: 0}

	data, err := json.Marshal(gardenBed)
	if err != nil {
		t.Fatalf("Failed to marshal gardenBed state: %v", err)
	}

	var restored GardenBedState
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal gardenBed state: %v", err)
	}

	if restored.RowCount != gardenBed.RowCount {
		t.Errorf("Expected row count %d, got %d", gardenBed.RowCount, restored.RowCount)
	}
	if restored.Rows[0].Status != gardenBed.Rows[0].Status {
		t.Errorf("Expected row 0 status '%s', got '%s'", gardenBed.Rows[0].Status, restored.Rows[0].Status)
	}
	if restored.Rows[0].PlantType != gardenBed.Rows[0].PlantType {
		t.Errorf("Expected row 0 plant type '%s', got '%s'", gardenBed.Rows[0].PlantType, restored.Rows[0].PlantType)
	}
	if restored.Rows[0].Stage != gardenBed.Rows[0].Stage {
		t.Errorf("Expected row 0 stage %d, got %d", gardenBed.Rows[0].Stage, restored.Rows[0].Stage)
	}
}

func TestGardenBedTick_NoTickForEmptyRows(t *testing.T) {
	gardenBed := NewGardenBedState(2)

	// All rows empty - they should just update LastTick
	startTick := time.Now()
	GardenBedTick(gardenBed, 100)
	elapsed := time.Since(startTick)

	if elapsed > 100*time.Millisecond {
		t.Errorf("GardenBedTick took too long (%v) on empty gardenBed", elapsed)
	}
	// Empty rows may get weeds (random), so we only check that LastTick was updated
	for i := range gardenBed.Rows {
		if gardenBed.Rows[i].LastTick != 100 {
			t.Errorf("Row %d: expected LastTick 100, got %d", i, gardenBed.Rows[i].LastTick)
		}
	}
}
