package game

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewGardenBedState(t *testing.T) {
	gb := NewGardenBedState(3)
	if gb.RowCount != 3 {
		t.Errorf("Expected row count 3, got %d", gb.RowCount)
	}
	if len(gb.Rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(gb.Rows))
	}
	for i, row := range gb.Rows {
		if row.Status != GardenBedRowEmpty {
			t.Errorf("Row %d: expected status '%s', got '%s'", i, GardenBedRowEmpty, row.Status)
		}
		if row.Weeds != 0 {
			t.Errorf("Row %d: expected 0 weeds, got %d", i, row.Weeds)
		}
	}
}

func TestNewGardenBedStateFromJSON(t *testing.T) {
	rows := []GardenBedRow{
		{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 1, Weeds: 0, WaterTimer: 0},
		{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0},
	}
	data, _ := json.Marshal(rows)

	gb := NewGardenBedStateFromJSON(data, 2)
	if gb.RowCount != 2 {
		t.Errorf("Expected row count 2, got %d", gb.RowCount)
	}
	if gb.Rows[0].Status != GardenBedRowPlanted {
		t.Errorf("Expected row 0 status '%s', got '%s'", GardenBedRowPlanted, gb.Rows[0].Status)
	}
	if gb.Rows[0].PlantType != PlantWheat {
		t.Errorf("Expected row 0 plant type '%s', got '%s'", PlantWheat, gb.Rows[0].PlantType)
	}

	gbEmpty := NewGardenBedStateFromJSON([]byte("[]"), 3)
	if gbEmpty.RowCount != 3 {
		t.Errorf("Expected row count 3 for empty, got %d", gbEmpty.RowCount)
	}

	gbNull := NewGardenBedStateFromJSON([]byte("null"), 2)
	if gbNull.RowCount != 2 {
		t.Errorf("Expected row count 2 for null, got %d", gbNull.RowCount)
	}
}

func TestGardenBedTick_NoChangeWhenAlreadyProcessed(t *testing.T) {
	gb := NewGardenBedState(1)
	gb.Rows[0] = GardenBedRow{
		Status:     GardenBedRowPlanted,
		PlantType:  PlantWheat,
		Stage:      0,
		Weeds:      0,
		WaterTimer: 0,
		LastTick:   100,
	}

	changed := GardenBedTick(gb, 100)
	if changed {
		t.Error("Expected no change when tick already processed")
	}
	if gb.Rows[0].Stage != 0 {
		t.Errorf("Expected stage still 0, got %d", gb.Rows[0].Stage)
	}
}

func TestGardenBedTick_MatureRow(t *testing.T) {
	gb := NewGardenBedState(1)
	gb.Rows[0] = GardenBedRow{
		Status:     GardenBedRowMature,
		PlantType:  PlantWheat,
		Stage:      2,
		Weeds:      0,
		WaterTimer: 0,
		LastTick:   0,
	}

	changed := GardenBedTick(gb, 100)
	if changed {
		t.Error("Expected no change for mature row")
	}
	if gb.Rows[0].Status != GardenBedRowMature {
		t.Errorf("Expected status still '%s', got '%s'", GardenBedRowMature, gb.Rows[0].Status)
	}
}

func TestGardenBedAction_Plant(t *testing.T) {
	planet := &Planet{
		ID: "test-planet-plant",
		GardenBedState: NewGardenBedState(1),
		Resources:    PlanetResources{Food: 100, Money: 100},
		Buildings:    []BuildingEntry{{Type: "farm", Level: 1, Enabled: true}},
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
	if result.FoodCost != 20 {
		t.Errorf("Expected food cost 20 for wheat weed, got %.0f", result.FoodCost)
	}
	if planet.Resources.Food != 80 {
		t.Errorf("Expected 80 food after weeding, got %.0f", planet.Resources.Food)
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

	result, err := GardenBedActionInternal(planet, "water", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	if result.Rows[0].WaterTimer != 100 {
		t.Errorf("Expected water_timer 100, got %d", result.Rows[0].WaterTimer)
	}
	if result.FoodCost != 10 {
		t.Errorf("Expected food cost 10 for wheat water, got %.0f", result.FoodCost)
	}
	if planet.Resources.Food != 90 {
		t.Errorf("Expected 90 food after watering, got %.0f", planet.Resources.Food)
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

func TestGardenBedAction_MultipleActions(t *testing.T) {
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

	result, _ := GardenBedActionInternal(planet, "plant", 0, PlantWheat)
	if !result.Success {
		t.Errorf("First action should succeed, got error: %s", result.Error)
	}

	result, _ = GardenBedActionInternal(planet, "water", 0, "")
	if !result.Success {
		t.Errorf("Second action should succeed without cooldown, got error: %s", result.Error)
	}
	if result.Rows[0].WaterTimer != 100 {
		t.Errorf("Expected water_timer 100, got %d", result.Rows[0].WaterTimer)
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

	result, err := GardenBedActionInternal(planet, "weed", 0, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure when weeding empty row")
	}
	if result.Error != "No weeds to remove" {
		t.Errorf("Expected error 'No weeds to remove', got '%s'", result.Error)
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
	gb := NewGardenBedState(2)
	gb.Rows[0] = GardenBedRow{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 1, Weeds: 1, WaterTimer: 5, LastTick: 100}
	gb.Rows[1] = GardenBedRow{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0, LastTick: 0}

	data, err := json.Marshal(gb)
	if err != nil {
		t.Fatalf("Failed to marshal gardenBed state: %v", err)
	}

	var restored GardenBedState
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal gardenBed state: %v", err)
	}

	if restored.RowCount != gb.RowCount {
		t.Errorf("Expected row count %d, got %d", gb.RowCount, restored.RowCount)
	}
	if restored.Rows[0].Status != gb.Rows[0].Status {
		t.Errorf("Expected row 0 status '%s', got '%s'", gb.Rows[0].Status, restored.Rows[0].Status)
	}
	if restored.Rows[0].PlantType != gb.Rows[0].PlantType {
		t.Errorf("Expected row 0 plant type '%s', got '%s'", gb.Rows[0].PlantType, restored.Rows[0].PlantType)
	}
	if restored.Rows[0].Stage != gb.Rows[0].Stage {
		t.Errorf("Expected row 0 stage %d, got %d", gb.Rows[0].Stage, restored.Rows[0].Stage)
	}
}

func TestGardenBedTick_NoTickForEmptyRows(t *testing.T) {
	gb := NewGardenBedState(2)

	startTick := time.Now()
	GardenBedTick(gb, 100)
	elapsed := time.Since(startTick)

	if elapsed > 100*time.Millisecond {
		t.Errorf("GardenBedTick took too long (%v) on empty gardenBed", elapsed)
	}
}
