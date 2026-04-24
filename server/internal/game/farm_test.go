package game

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewFarmState(t *testing.T) {
	farm := NewFarmState(3)

	if farm.RowCount != 3 {
		t.Errorf("Expected row count 3, got %d", farm.RowCount)
	}
	if len(farm.Rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(farm.Rows))
	}
	for i, row := range farm.Rows {
		if row.Status != FarmRowEmpty {
			t.Errorf("Row %d: expected status '%s', got '%s'", i, FarmRowEmpty, row.Status)
		}
		if row.Weeds != 0 {
			t.Errorf("Row %d: expected 0 weeds, got %d", i, row.Weeds)
		}
	}
}

func TestNewFarmStateFromJSON(t *testing.T) {
	// Test with valid JSON
	rows := []FarmRow{
		{Status: FarmRowPlanted, PlantType: PlantWheat, Stage: 1, Weeds: 0, WaterTimer: 0},
		{Status: FarmRowEmpty, Weeds: 0, WaterTimer: 0},
	}
	data, _ := json.Marshal(rows)

	farm := NewFarmStateFromJSON(data, 2)
	if farm.RowCount != 2 {
		t.Errorf("Expected row count 2, got %d", farm.RowCount)
	}
	if len(farm.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(farm.Rows))
	}
	if farm.Rows[0].Status != FarmRowPlanted {
		t.Errorf("Expected row 0 status '%s', got '%s'", FarmRowPlanted, farm.Rows[0].Status)
	}
	if farm.Rows[0].PlantType != PlantWheat {
		t.Errorf("Expected row 0 plant type '%s', got '%s'", PlantWheat, farm.Rows[0].PlantType)
	}

	// Test with empty JSON
	emptyFarm := NewFarmStateFromJSON([]byte("[]"), 3)
	if emptyFarm.RowCount != 3 {
		t.Errorf("Expected row count 3 for empty farm, got %d", emptyFarm.RowCount)
	}

	// Test with null JSON
	nullFarm := NewFarmStateFromJSON([]byte("null"), 2)
	if nullFarm.RowCount != 2 {
		t.Errorf("Expected row count 2 for null farm, got %d", nullFarm.RowCount)
	}
}

func TestFarmTick_Growth(t *testing.T) {
	farm := NewFarmState(1)
	farm.Rows[0] = FarmRow{
		Status:     FarmRowPlanted,
		PlantType:  PlantWheat,
		Stage:      0,
		Weeds:      0,
		WaterTimer: 0,
		LastTick:   0,
	}

	// First tick at 100: FarmTicksSinceLast becomes 1, no advance
	FarmTick(farm, 100)
	if farm.Rows[0].Stage != 0 {
		t.Errorf("After tick 100: expected stage 0, got %d", farm.Rows[0].Stage)
	}

	// Tick at 200: FarmTicksSinceLast becomes 2, advance to stage 1
	changed := FarmTick(farm, 200)
	if !changed {
		t.Error("Expected state change on tick 200")
	}
	if farm.Rows[0].Stage != 1 {
		t.Errorf("After tick 200: expected stage 1, got %d", farm.Rows[0].Stage)
	}

	// Tick at 300: FarmTicksSinceLast becomes 1, no advance
	FarmTick(farm, 300)
	if farm.Rows[0].Stage != 1 {
		t.Errorf("After tick 300: expected stage 1, got %d", farm.Rows[0].Stage)
	}

	// Tick at 400: FarmTicksSinceLast becomes 2, advance to stage 2 (mature)
	changed = FarmTick(farm, 400)
	if !changed {
		t.Error("Expected state change on tick 400")
	}
	if farm.Rows[0].Stage != 2 {
		t.Errorf("After tick 400: expected stage 2, got %d", farm.Rows[0].Stage)
	}
	if farm.Rows[0].Status != FarmRowMature {
		t.Errorf("After tick 400: expected status '%s', got '%s'", FarmRowMature, farm.Rows[0].Status)
	}
}

func TestFarmTick_WaterAcceleration(t *testing.T) {
	farm := NewFarmState(1)
	farm.Rows[0] = FarmRow{
		Status:     FarmRowPlanted,
		PlantType:  PlantWheat,
		Stage:      0,
		Weeds:      0,
		WaterTimer: 2,
		LastTick:   0,
	}

	// Tick 100: water_timer > 0, advance stage
	changed := FarmTick(farm, 100)
	if !changed {
		t.Error("Expected state change on tick 100 with water")
	}
	if farm.Rows[0].Stage != 1 {
		t.Errorf("After tick 100 with water: expected stage 1, got %d", farm.Rows[0].Stage)
	}
	if farm.Rows[0].WaterTimer != 1 {
		t.Errorf("After tick 100: expected water_timer 1, got %d", farm.Rows[0].WaterTimer)
	}

	// Tick 200: water_timer > 0, advance stage again
	changed = FarmTick(farm, 200)
	if !changed {
		t.Error("Expected state change on tick 200 with water")
	}
	if farm.Rows[0].Stage != 2 {
		t.Errorf("After tick 200 with water: expected stage 2, got %d", farm.Rows[0].Stage)
	}
	if farm.Rows[0].WaterTimer != 0 {
		t.Errorf("After tick 200: expected water_timer 0, got %d", farm.Rows[0].WaterTimer)
	}
}

func TestFarmTick_NoChangeWhenAlreadyProcessed(t *testing.T) {
	farm := NewFarmState(1)
	farm.Rows[0] = FarmRow{
		Status:   FarmRowPlanted,
		PlantType: PlantWheat,
		Stage:    0,
		Weeds:    0,
		WaterTimer: 0,
		LastTick: 100,
	}

	// Calling FarmTick again with same tick should not change anything
	changed := FarmTick(farm, 100)
	if changed {
		t.Error("Expected no change when tick already processed")
	}
	if farm.Rows[0].Stage != 0 {
		t.Errorf("Expected stage still 0, got %d", farm.Rows[0].Stage)
	}
}

func TestFarmTick_EmptyRow(t *testing.T) {
	farm := NewFarmState(1)
	// Row is empty by default

	changed := FarmTick(farm, 100)
	if changed {
		t.Error("Expected no change for empty row")
	}
}

func TestFarmTick_MatureRow(t *testing.T) {
	farm := NewFarmState(1)
	farm.Rows[0] = FarmRow{
		Status:   FarmRowMature,
		PlantType: PlantWheat,
		Stage:    2,
		Weeds:    0,
		WaterTimer: 0,
		LastTick: 0,
	}

	changed := FarmTick(farm, 100)
	if changed {
		t.Error("Expected no change for mature row")
	}
	if farm.Rows[0].Status != FarmRowMature {
		t.Errorf("Expected status still '%s', got '%s'", FarmRowMature, farm.Rows[0].Status)
	}
}

func TestFarmAction_Plant(t *testing.T) {
	// We need a planet for the action to work
	// Create a mock planet with farm state
	planet := &Planet{
		ID:      "test-planet-plant",
		FarmState: NewFarmState(1),
		Resources: PlanetResources{Food: 100},
	}

	result, err := FarmActionInternal(planet, "plant", 0, PlantWheat)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].Status != FarmRowPlanted {
		t.Errorf("Expected status '%s', got '%s'", FarmRowPlanted, result.Rows[0].Status)
	}
	if result.Rows[0].PlantType != PlantWheat {
		t.Errorf("Expected plant type '%s', got '%s'", PlantWheat, result.Rows[0].PlantType)
	}
	if result.Rows[0].Stage != 0 {
		t.Errorf("Expected stage 0, got %d", result.Rows[0].Stage)
	}
}

func TestFarmAction_PlantOnNonEmptyRow(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "plant", 0, PlantBerries)
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

func TestFarmAction_Weed(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 2, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "weed", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].Weeds != 1 {
		t.Errorf("Expected 1 weed after weeding, got %d", result.Rows[0].Weeds)
	}
}

func TestFarmAction_WeedMax(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 3, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "weed", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].Weeds != 2 {
		t.Errorf("Expected 2 weeds after weeding from 3, got %d", result.Rows[0].Weeds)
	}
}

func TestFarmAction_Water(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "water", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].WaterTimer != 10 {
		t.Errorf("Expected water_timer 10, got %d", result.Rows[0].WaterTimer)
	}
}

func TestFarmAction_Harvest(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowMature, PlantType: PlantWheat, Stage: 2, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "harvest", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].Status != FarmRowEmpty {
		t.Errorf("Expected status '%s' after harvest, got '%s'", FarmRowEmpty, result.Rows[0].Status)
	}
	if result.FoodGain != 5 {
		t.Errorf("Expected food gain 5, got %.0f", result.FoodGain)
	}
	if planet.Resources.Food != 105 {
		t.Errorf("Expected 105 food after harvest, got %.0f", planet.Resources.Food)
	}
}

func TestFarmAction_HarvestBerries(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowMature, PlantType: PlantBerries, Stage: 2, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "harvest", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.FoodGain != 15 {
		t.Errorf("Expected food gain 15 for berries, got %.0f", result.FoodGain)
	}
}

func TestFarmAction_HarvestMelon(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowMature, PlantType: PlantMelon, Stage: 2, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "harvest", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.FoodGain != 30 {
		t.Errorf("Expected food gain 30 for melon, got %.0f", result.FoodGain)
	}
}

func TestFarmAction_HarvestNonMature(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "harvest", 0, "")
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

func TestFarmAction_InvalidRowIndex(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "plant", 5, PlantWheat)
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

func TestFarmAction_UnknownAction(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "dig", 0, "")
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

func TestFarmAction_Cooldown(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// First action should succeed
	result, _ := FarmActionInternal(planet, "plant", 0, PlantWheat)
	if !result.Success {
		t.Errorf("First action should succeed, got error: %s", result.Error)
	}

	// Second action immediately should fail due to cooldown
	result, _ = FarmActionInternal(planet, "water", 0, "")
	if result.Success {
		t.Error("Second action should fail due to cooldown")
	}
	if result.Error == "" {
		t.Error("Expected cooldown error message")
	}
}

func TestFarmPlantDefinitions(t *testing.T) {
	wheat := farmPlants[PlantWheat]
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

	berries := farmPlants[PlantBerries]
	if berries == nil {
		t.Fatal("Berries plant not found")
	}
	if berries.FoodReward != 15 {
		t.Errorf("Expected berries food reward 15, got %.0f", berries.FoodReward)
	}

	melon := farmPlants[PlantMelon]
	if melon == nil {
		t.Fatal("Melon plant not found")
	}
	if melon.FoodReward != 30 {
		t.Errorf("Expected melon food reward 30, got %.0f", melon.FoodReward)
	}
}

func TestGetAllFarmPlants(t *testing.T) {
	plants := GetAllFarmPlants()
	if len(plants) != 3 {
		t.Errorf("Expected 3 plant types, got %d", len(plants))
	}
}

func TestGetFarmPlant_Unknown(t *testing.T) {
	plant := GetFarmPlant("unknown")
	if plant != nil {
		t.Error("Expected nil for unknown plant type")
	}
}

func TestFarmTick_MultipleRows(t *testing.T) {
	farm := NewFarmState(3)

	// Plant different types in different rows
	farm.Rows[0] = FarmRow{Status: FarmRowPlanted, PlantType: PlantWheat, Stage: 0, Weeds: 0, WaterTimer: 0, LastTick: 0}
	farm.Rows[1] = FarmRow{Status: FarmRowPlanted, PlantType: PlantBerries, Stage: 0, Weeds: 0, WaterTimer: 0, LastTick: 0}
	farm.Rows[2] = FarmRow{Status: FarmRowEmpty, Weeds: 0, WaterTimer: 0}

	// Process ticks
	FarmTick(farm, 100)
	FarmTick(farm, 200)
	FarmTick(farm, 300)
	FarmTick(farm, 400)
	FarmTick(farm, 500)
	FarmTick(farm, 600)

	// Row 0 (wheat) should be mature at stage 2
	if farm.Rows[0].Status != FarmRowMature {
		t.Errorf("Row 0: expected status '%s', got '%s'", FarmRowMature, farm.Rows[0].Status)
	}
	if farm.Rows[0].Stage != 2 {
		t.Errorf("Row 0: expected stage 2, got %d", farm.Rows[0].Stage)
	}

	// Row 1 (berries) should also be mature
	if farm.Rows[1].Status != FarmRowMature {
		t.Errorf("Row 1: expected status '%s', got '%s'", FarmRowMature, farm.Rows[1].Status)
	}
	if farm.Rows[1].Stage != 2 {
		t.Errorf("Row 1: expected stage 2, got %d", farm.Rows[1].Stage)
	}

	// Row 2 should still be empty
	if farm.Rows[2].Status != FarmRowEmpty {
		t.Errorf("Row 2: expected status '%s', got '%s'", FarmRowEmpty, farm.Rows[2].Status)
	}
}

func TestFarmTick_WaterTimerExpiry(t *testing.T) {
	farm := NewFarmState(1)
	farm.Rows[0] = FarmRow{
		Status:     FarmRowPlanted,
		PlantType:  PlantWheat,
		Stage:      0,
		Weeds:      0,
		WaterTimer: 1,
		LastTick:   0,
	}

	// Tick 100: water_timer > 0, advance stage, water_timer becomes 0
	changed := FarmTick(farm, 100)
	if !changed {
		t.Error("Expected state change")
	}
	if farm.Rows[0].Stage != 1 {
		t.Errorf("After tick 100: expected stage 1, got %d", farm.Rows[0].Stage)
	}
	if farm.Rows[0].WaterTimer != 0 {
		t.Errorf("After tick 100: expected water_timer 0, got %d", farm.Rows[0].WaterTimer)
	}

	// Tick 200: no water, FarmTicksSinceLast becomes 1, no advance
	changed = FarmTick(farm, 200)
	// changed may be false since water_timer is already 0 and no stage advance
	if farm.Rows[0].Stage != 1 {
		t.Errorf("After tick 200: expected stage 1 (no water), got %d", farm.Rows[0].Stage)
	}

	// Tick 300: FarmTicksSinceLast becomes 2, advance to stage 2 (mature)
	changed = FarmTick(farm, 300)
	if !changed {
		t.Error("Expected state change")
	}
	if farm.Rows[0].Stage != 2 {
		t.Errorf("After tick 300: expected stage 2, got %d", farm.Rows[0].Stage)
	}
	if farm.Rows[0].Status != FarmRowMature {
		t.Errorf("After tick 300: expected status '%s', got '%s'", FarmRowMature, farm.Rows[0].Status)
	}
}

func TestFarmAction_PlantInvalidType(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "plant", 0, "tomato")
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

func TestFarmAction_WeedEmptyRow(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "weed", 0, "")
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

func TestFarmAction_WaterEmptyRow(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-base",
		FarmState: &FarmState{
			Rows: []FarmRow{
				{Status: FarmRowEmpty, Weeds: 0, WaterTimer: 0},
			},
			RowCount: 1,
		},
		Resources: PlanetResources{Food: 100},
	}

	// Clear cooldown between tests
	ClearFarmCooldown("test-planet-base")
	result, err := FarmActionInternal(planet, "water", 0, "")
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

func TestFarmStateJSONSerialization(t *testing.T) {
	farm := NewFarmState(2)
	farm.Rows[0] = FarmRow{Status: FarmRowPlanted, PlantType: PlantWheat, Stage: 1, Weeds: 1, WaterTimer: 5, LastTick: 100}
	farm.Rows[1] = FarmRow{Status: FarmRowEmpty, Weeds: 0, WaterTimer: 0, LastTick: 0}

	data, err := json.Marshal(farm)
	if err != nil {
		t.Fatalf("Failed to marshal farm state: %v", err)
	}

	var restored FarmState
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal farm state: %v", err)
	}

	if restored.RowCount != farm.RowCount {
		t.Errorf("Expected row count %d, got %d", farm.RowCount, restored.RowCount)
	}
	if restored.Rows[0].Status != farm.Rows[0].Status {
		t.Errorf("Expected row 0 status '%s', got '%s'", farm.Rows[0].Status, restored.Rows[0].Status)
	}
	if restored.Rows[0].PlantType != farm.Rows[0].PlantType {
		t.Errorf("Expected row 0 plant type '%s', got '%s'", farm.Rows[0].PlantType, restored.Rows[0].PlantType)
	}
	if restored.Rows[0].Stage != farm.Rows[0].Stage {
		t.Errorf("Expected row 0 stage %d, got %d", farm.Rows[0].Stage, restored.Rows[0].Stage)
	}
}

func TestFarmTick_NoTickForEmptyRows(t *testing.T) {
	farm := NewFarmState(2)

	// All rows empty, no ticks should change anything
	startTick := time.Now()
	changed := FarmTick(farm, 100)
	elapsed := time.Since(startTick)

	if elapsed > 100*time.Millisecond {
		t.Errorf("FarmTick took too long (%v) on empty farm", elapsed)
	}
	if changed {
		t.Error("Expected no change for all-empty farm")
	}
}
