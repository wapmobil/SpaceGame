package game

import (
	"testing"

	"spacegame/internal/game/research"
)

// --- Production tests ---

func TestFarmProducesFood(t *testing.T) {
	planet := NewPlanet("test-1", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 1)
	planet.AddBuildingDirect("farm", 1)

	prod := planet.GetProductionResult()
	if prod.Food != 1 {
		t.Errorf("expected farm at level 1 to produce 1 food, got %f", prod.Food)
	}
}

func TestFarmProductionScalesWithLevel(t *testing.T) {
	planet := NewPlanet("test-2", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 3) // 45 energy total
	planet.AddBuildingDirect("farm", 3)  // level 3 = 3 food per tick

	prod := planet.GetProductionResult()
	if prod.Food != 3 {
		t.Errorf("expected farm at level 3 to produce 3 food, got %f", prod.Food)
	}
}

func TestSolarProducesEnergy(t *testing.T) {
	planet := NewPlanet("test-3", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 1)

	prod := planet.GetProductionResult()
	if prod.Energy != 15 {
		t.Errorf("expected solar at level 1 to produce 15 energy, got %f", prod.Energy)
	}
}

func TestSolarProductionScalesWithLevel(t *testing.T) {
	planet := NewPlanet("test-solar-scale", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 2) // 30 energy

	prod := planet.GetProductionResult()
	if prod.Energy != 30 {
		t.Errorf("expected solar at level 2 to produce 30 energy, got %f", prod.Energy)
	}
}

func TestBaseConsumesFood(t *testing.T) {
	planet := NewPlanet("test-14", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 1)
	planet.AddBuildingDirect("solar", 1)
	planet.AddBuildingDirect("base", 1)

	prod := planet.GetProductionResult()
	if prod.Food != -1 {
		t.Errorf("expected base at level 1 to consume 1 food (-1), got %f", prod.Food)
	}
}

// --- Energy tests ---

func TestEnergyBalancePositive(t *testing.T) {
	planet := NewPlanet("test-4", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)  // consumes 10 energy
	planet.AddBuildingDirect("solar", 1) // produces 15 energy

	planet.tickEnergy()
	balance := planet.GetEnergyBalance()
	if balance != 5 {
		t.Errorf("expected energy balance of 5 (15 solar - 10 farm), got %f", balance)
	}
}

func TestEnergyBalanceDeficit(t *testing.T) {
	planet := NewPlanet("test-5", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1) // consumes 10 energy

	planet.tickEnergy()
	balance := planet.GetEnergyBalance()
	if balance != -10 {
		t.Errorf("expected energy balance of -10, got %f", balance)
	}
}

func TestEnergyBufferCapsAtMax(t *testing.T) {
	planet := NewPlanet("test-overflow", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 2) // 30 energy production

	planet.tickEnergy()

	if planet.EnergyBuffer.Value > planet.EnergyBuffer.Max {
		t.Errorf("expected energy buffer value %f to be clamped at max %f",
			planet.EnergyBuffer.Value, planet.EnergyBuffer.Max)
	}
}

func TestEnergyBalanceUsesCachedValue(t *testing.T) {
	planet := NewPlanet("test-cached", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)
	planet.AddBuildingDirect("solar", 1)

	planet.tickEnergy()
	balance1 := planet.GetEnergyBalance()
	balance2 := planet.GetEnergyBalance()
	if balance1 != balance2 {
		t.Errorf("expected GetEnergyBalance to return cached value, got %f and %f", balance1, balance2)
	}
}

func TestEnergyBufferMaxUpdates(t *testing.T) {
	planet := NewPlanet("test-energy-2", "owner-1", "Test Planet", nil)

	if planet.EnergyBuffer.Max != 100 {
		t.Errorf("expected initial max energy of 100, got %f", planet.EnergyBuffer.Max)
	}

	planet.AddBuildingDirect("energy_storage", 3)
	planet.EnergyBuffer.UpdateMax(3)

	expectedMax := 100.0 + 3.0*1000.0
	if planet.EnergyBuffer.Max != expectedMax {
		t.Errorf("expected max energy of %f with 3 energy_storage, got %f", expectedMax, planet.EnergyBuffer.Max)
	}
}

func TestEnergyBufferMaxClampsValue(t *testing.T) {
	planet := NewPlanet("test-clamp", "owner-1", "Test Planet", nil)
	planet.EnergyBuffer.Value = 150
	planet.EnergyBuffer.UpdateMax(0)

	if planet.EnergyBuffer.Value != 100 {
		t.Errorf("expected energy buffer value clamped to 100, got %f", planet.EnergyBuffer.Value)
	}
}

// --- Disabled buildings ---

func TestDisabledBuildingProducesNoResources(t *testing.T) {
	planet := NewPlanet("test-6", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)
	planet.AddBuildingDirect("solar", 1)

	farmIdx := planet.FindBuildingIndex("farm")
	planet.Buildings[farmIdx].Enabled = false

	prod := planet.GetProductionResult()
	if prod.Food != 0 {
		t.Errorf("expected no food production from disabled building, got %f", prod.Food)
	}
}

func TestDisabledBuildingDoesNotConsumeEnergy(t *testing.T) {
	planet := NewPlanet("test-disabled-energy", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 1) // 15 energy
	planet.AddBuildingDirect("farm", 1)  // 10 energy consumption

	// Disable the farm
	farmIdx := planet.FindBuildingIndex("farm")
	planet.Buildings[farmIdx].Enabled = false

	planet.tickEnergy()
	balance := planet.GetEnergyBalance()
	if balance != 15 {
		t.Errorf("expected energy balance of 15 (farm disabled, no consumption), got %f", balance)
	}
}

// --- Storage ---

func TestStorageIncreasesCapacity(t *testing.T) {
	planet := NewPlanet("test-8", "owner-1", "Test Planet", nil)
	capacity1 := planet.CalculateStorageCapacity()
	if capacity1 != 1000 {
		t.Errorf("expected base capacity of 1000, got %f", capacity1)
	}

	planet.AddBuildingDirect("storage", 2)
	capacity2 := planet.CalculateStorageCapacity()
	if capacity2 != 3000 {
		t.Errorf("expected capacity of 3000 with 2 storage buildings, got %f", capacity2)
	}
}

func TestNoStorageCapacityIsBase(t *testing.T) {
	planet := NewPlanet("test-no-storage", "owner-1", "Test Planet", nil)
	capacity := planet.CalculateStorageCapacity()
	if capacity != 1000 {
		t.Errorf("expected base capacity of 1000 with no storage buildings, got %f", capacity)
	}
}

// --- Building management ---

func TestGetBuildingLevel(t *testing.T) {
	planet := NewPlanet("test-11", "owner-1", "Test Planet", nil)

	if level := planet.GetBuildingLevel("farm"); level != 0 {
		t.Errorf("expected farm level 0, got %d", level)
	}

	planet.AddBuildingDirect("farm", 1)
	if level := planet.GetBuildingLevel("farm"); level != 1 {
		t.Errorf("expected farm level 1, got %d", level)
	}

	planet.AddBuildingDirect("farm", 2)
	if level := planet.GetBuildingLevel("farm"); level != 2 {
		t.Errorf("expected farm level 2, got %d", level)
	}
}

func TestGetBuildingLevelUnknown(t *testing.T) {
	planet := NewPlanet("test-unknown", "owner-1", "Test Planet", nil)
	if level := planet.GetBuildingLevel("nonexistent"); level != 0 {
		t.Errorf("expected level 0 for unknown building, got %d", level)
	}
}

func TestTotalBuildingLevels(t *testing.T) {
	planet := NewPlanet("test-12", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)
	planet.AddBuildingDirect("solar", 1)
	planet.AddBuildingDirect("storage", 1)

	total := planet.GetTotalBuildingLevels()
	if total != 3 {
		t.Errorf("expected 3 total building levels, got %d", total)
	}
}

func TestFindBuildingIndex(t *testing.T) {
	planet := NewPlanet("test-find", "owner-1", "Test Planet", nil)

	if idx := planet.FindBuildingIndex("farm"); idx != -1 {
		t.Errorf("expected -1 for missing building, got %d", idx)
	}

	planet.AddBuildingDirect("farm", 1)
	if idx := planet.FindBuildingIndex("farm"); idx != 0 {
		t.Errorf("expected 0 for first building, got %d", idx)
	}

	planet.AddBuildingDirect("solar", 1)
	if idx := planet.FindBuildingIndex("solar"); idx != 1 {
		t.Errorf("expected 1 for second building, got %d", idx)
	}
}

func TestBuildingSliceOrder(t *testing.T) {
	planet := NewPlanet("test-order-1", "owner-1", "Test Planet", nil)

	planet.AddBuildingDirect("farm", 1)
	planet.AddBuildingDirect("solar", 2)
	planet.AddBuildingDirect("base", 1)

	if len(planet.Buildings) != 3 {
		t.Errorf("expected 3 buildings, got %d", len(planet.Buildings))
	}

	if idx := planet.FindBuildingIndex("farm"); idx < 0 {
		t.Error("expected to find farm building")
	}
	if idx := planet.FindBuildingIndex("shipyard"); idx >= 0 {
		t.Error("expected NOT to find shipyard building")
	}
}

func TestAddBuildingPopulatesEntry(t *testing.T) {
	planet := NewPlanet("test-populate-1", "owner-1", "Test Planet", nil)

	planet.AddBuildingDirect("farm", 1)

	if len(planet.Buildings) != 1 {
		t.Fatalf("expected 1 building, got %d", len(planet.Buildings))
	}

	b := planet.Buildings[0]
	if b.Type != "farm" {
		t.Errorf("expected type 'farm', got '%s'", b.Type)
	}
	if b.Level != 1 {
		t.Errorf("expected level 1, got %d", b.Level)
	}
	if b.BuildTime <= 0 {
		t.Errorf("expected positive build time, got %f", b.BuildTime)
	}
	if b.Production.Food <= 0 {
		t.Errorf("expected positive food production, got %f", b.Production.Food)
	}
	if b.Production.Energy >= 0 {
		t.Errorf("expected negative energy (consumption), got %f", b.Production.Energy)
	}
}

func TestConfirmBuildingPopulatesEntry(t *testing.T) {
	planet := NewPlanet("test-confirm-1", "owner-1", "Test Planet", nil)

	planet.AddBuildingDirect("farm", 1)
	planet.Buildings[0].BuildProgress = 0

	err := planet.ConfirmBuilding("farm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if planet.Buildings[0].BuildProgress != -1 {
		t.Errorf("expected build_progress to be -1 after confirmation, got %f", planet.Buildings[0].BuildProgress)
	}
	if !planet.Buildings[0].Enabled {
		t.Error("expected building to be enabled after confirmation")
	}
}

func TestConfirmBuildingNotFound(t *testing.T) {
	planet := NewPlanet("test-confirm-nf", "owner-1", "Test Planet", nil)

	err := planet.ConfirmBuilding("nonexistent")
	if err == nil {
		t.Fatal("expected error when confirming non-existent building")
	}
}

func TestConfirmBuildingNotReady(t *testing.T) {
	planet := NewPlanet("test-confirm-notpending", "owner-1", "Test Planet", nil)

	planet.AddBuildingDirect("farm", 1)

	err := planet.ConfirmBuilding("farm")
	if err == nil {
		t.Fatal("expected error when confirming non-ready building")
	}
}

func TestGetPendingBuildings(t *testing.T) {
	planet := NewPlanet("test-pending", "owner-1", "Test Planet", nil)

	planet.AddBuildingDirect("farm", 1)
	planet.Buildings[0].BuildProgress = 0

	pending := planet.GetPendingBuildings()
	if !pending["farm"] {
		t.Error("expected farm to be in pending buildings")
	}
}

func TestGetPendingBuildingsEmpty(t *testing.T) {
	planet := NewPlanet("test-pending-empty", "owner-1", "Test Planet", nil)

	pending := planet.GetPendingBuildings()
	if len(pending) != 0 {
		t.Errorf("expected no pending buildings, got %v", pending)
	}
}

// --- Base operational ---

func TestBaseOperational(t *testing.T) {
	tests := []struct {
		name     string
		food     float64
		operational bool
	}{
		{"with food", 10, true},
		{"zero food", 0, false},
		{"negative food", -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planet := NewPlanet(tt.name, "owner-1", "Test Planet", nil)
			planet.AddBuildingDirect("base", 1)
			planet.Resources.Food = tt.food
			if got := planet.BaseOperational(); got != tt.operational {
				t.Errorf("BaseOperational() = %v, want %v (food=%f)", got, tt.operational, tt.food)
			}
		})
	}
}

// --- Research ---

func TestStartResearchBlockedWithoutFood(t *testing.T) {
	planet := NewPlanet("test-research-1", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("base", 1)
	planet.Resources.Food = 0

	tech := research.GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("expected planet_exploration tech to exist")
	}

	err := planet.StartResearch(tech.ID)
	if err == nil {
		t.Fatal("expected error when starting research without food")
	}
}

func TestStartResearchAllowedWithFood(t *testing.T) {
	planet := NewPlanet("test-research-2", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("base", 1)
	planet.Resources.Food = 1000
	planet.Resources.Money = 1000

	tech := research.GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("expected planet_exploration tech to exist")
	}

	err := planet.StartResearch(tech.ID)
	if err != nil {
		t.Fatalf("unexpected error starting research with food: %v", err)
	}
}

func TestStartResearchTechNotFound(t *testing.T) {
	planet := NewPlanet("test-research-nf", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("base", 1)
	planet.Resources.Food = 1000

	err := planet.StartResearch("nonexistent_tech")
	if err == nil {
		t.Fatal("expected error for non-existent tech")
	}
}

// --- Tick integration ---

func TestTickProducesResources(t *testing.T) {
	planet := NewPlanet("test-tick", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 1) // produces 15 energy
	planet.AddBuildingDirect("farm", 1)  // consumes 10 energy, produces 1 food

	if planet.EnergyBuffer.Max != 100 {
		t.Errorf("expected initial energy buffer max of 100, got %f", planet.EnergyBuffer.Max)
	}

	planet.tickEnergy()
	balance := planet.GetEnergyBalance()
	if balance != 5 {
		t.Errorf("expected energy balance of 5 (15 solar - 10 farm), got %f", balance)
	}

	prod := planet.GetProductionResult()
	if prod.Food != 1 {
		t.Errorf("expected farm to produce 1 food with energy available, got %f", prod.Food)
	}
}

func TestTickMultiBuildingProduction(t *testing.T) {
	planet := NewPlanet("test-multi-prod", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 2) // 30 energy
	planet.AddBuildingDirect("farm", 2)  // 2 food

	prod := planet.GetProductionResult()
	if prod.Food != 2 {
		t.Errorf("expected 2 food from 2 farms, got %f", prod.Food)
	}
}

func TestMaxConcurrentBuildingsBase(t *testing.T) {
	planet := NewPlanet("test-max-const", "owner-1", "Test Planet", nil)
	max := planet.GetMaxConcurrentBuildings()
	if max != 1 {
		t.Errorf("expected 1 max concurrent building, got %d", max)
	}
}

func TestMaxConcurrentBuildingsWithResearch(t *testing.T) {
	planet := NewPlanet("test-max-const-res", "owner-1", "Test Planet", nil)
	planet.Resources.Food = 1000
	planet.Resources.Money = 1000

	// Mark parallel_construction as completed (level 1)
	planet.Research.Completed["parallel_construction"] = 1

	max := planet.GetMaxConcurrentBuildings()
	if max != 2 {
		t.Errorf("expected 2 max concurrent buildings after parallel_construction, got %d", max)
	}
}

// --- Energy balance getter ---

func TestGetEnergyBalanceReturnsCached(t *testing.T) {
	planet := NewPlanet("test-balance-cached", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)
	planet.AddBuildingDirect("solar", 1)

	planet.tickEnergy()
	expected := planet.EnergyBalance

	got := planet.GetEnergyBalance()
	if got != expected {
		t.Errorf("expected GetEnergyBalance to return cached %f, got %f", expected, got)
	}
}

// --- GetState ---

func TestGetStateReturnsAllFields(t *testing.T) {
	planet := NewPlanet("test-state", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 1)
	planet.AddBuildingDirect("farm", 1)

	state := planet.GetState()

	if state["id"] != "test-state" {
		t.Errorf("expected id 'test-state', got %v", state["id"])
	}
	if state["owner_id"] != "owner-1" {
		t.Errorf("expected owner_id 'owner-1', got %v", state["owner_id"])
	}
	if state["name"] != "Test Planet" {
		t.Errorf("expected name 'Test Planet', got %v", state["name"])
	}
	if _, ok := state["buildings"]; !ok {
		t.Error("expected 'buildings' key in state")
	}
	if _, ok := state["resources"]; !ok {
		t.Error("expected 'resources' key in state")
	}
	if _, ok := state["expeditions"]; !ok {
		t.Error("expected 'expeditions' key in state")
	}
}

func TestReadyBuildingIsDisabled(t *testing.T) {
	planet := NewPlanet("test-pending-1", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)
	planet.Buildings[0].BuildProgress = 0
	planet.Buildings[0].Enabled = true

	planet.Tick(1)

	if planet.Buildings[0].Enabled {
		t.Error("expected ready building to be disabled after tick")
	}
}

func TestBuildingUnderConstructionIsDisabled(t *testing.T) {
	planet := NewPlanet("test-construction-1", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)
	planet.Buildings[0].BuildProgress = 5
	planet.Buildings[0].Enabled = true

	planet.Tick(1)

	if planet.Buildings[0].Enabled {
		t.Error("expected building under construction to be disabled after tick")
	}
}

func TestCompletedBuildingIsEnabled(t *testing.T) {
	planet := NewPlanet("test-completed-1", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)
	planet.Buildings[0].BuildProgress = -1
	planet.Buildings[0].Enabled = true

	planet.Tick(1)

	if !planet.Buildings[0].IsWorking() {
		t.Error("expected completed building to remain working after tick")
	}
}

func TestDisabledBuildingProductionIsZero(t *testing.T) {
	planet := NewPlanet("test-disabled-prod-1", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)
	planet.Buildings[0].Enabled = true
	planet.PopulateBuildingEntry(0)

	prodEnabled := planet.Buildings[0].Production.Food

	planet.Buildings[0].Enabled = false
	planet.PopulateBuildingEntry(0)

	prodDisabled := planet.Buildings[0].Production.Food

	if prodEnabled != 1 {
		t.Errorf("expected enabled farm production to be 1, got %f", prodEnabled)
	}
	if prodDisabled != 0 {
		t.Errorf("expected disabled farm production to be 0, got %f", prodDisabled)
	}
}

func TestGetStatePopulatesProduction(t *testing.T) {
	planet := NewPlanet("test-state-prod-1", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)
	planet.AddBuildingDirect("solar", 1)

	state := planet.GetState()
	buildings := state["buildings"].([]BuildingEntry)

	farm := buildings[0]
	if farm.Production.Food != 1 {
		t.Errorf("expected farm production in GetState to be 1, got %f", farm.Production.Food)
	}

	solar := buildings[1]
	if solar.Production.Energy != 15 {
		t.Errorf("expected solar production in GetState to be 15, got %f", solar.Production.Energy)
	}
}

func TestFarmUpgrade_InitializesNewGardenBedRows(t *testing.T) {
	planet := NewPlanet("test-farm-upgrade", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 2)

	if planet.GardenBedState == nil {
		t.Fatal("expected garden bed state to be initialized")
	}
	if planet.GardenBedState.RowCount != 2 {
		t.Errorf("expected 2 rows, got %d", planet.GardenBedState.RowCount)
	}

	planet.GardenBedState.Rows[0] = GardenBedRow{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 1, Weeds: 1, WaterTimer: 50}

	planet.ensureGardenBedState()

	planet.AddBuildingDirect("farm", 4)
	planet.ensureGardenBedState()

	if planet.GardenBedState.RowCount != 4 {
		t.Errorf("expected 4 rows after upgrade, got %d", planet.GardenBedState.RowCount)
	}
	if len(planet.GardenBedState.Rows) != 4 {
		t.Errorf("expected 4 row entries, got %d", len(planet.GardenBedState.Rows))
	}
	if planet.GardenBedState.Rows[0].Status != GardenBedRowPlanted {
		t.Error("expected row 0 to be preserved as planted")
	}
	if planet.GardenBedState.Rows[0].PlantType != PlantWheat {
		t.Error("expected row 0 plant type to be wheat")
	}
	if planet.GardenBedState.Rows[0].Weeds != 1 {
		t.Errorf("expected row 0 weeds to be 1, got %d", planet.GardenBedState.Rows[0].Weeds)
	}
	if planet.GardenBedState.Rows[1].Status != GardenBedRowEmpty {
		t.Errorf("expected row 1 to be empty, got '%s'", planet.GardenBedState.Rows[1].Status)
	}
	if planet.GardenBedState.Rows[2].Status != GardenBedRowEmpty {
		t.Errorf("expected row 2 (new) to be empty, got '%s'", planet.GardenBedState.Rows[2].Status)
	}
	if planet.GardenBedState.Rows[3].Status != GardenBedRowEmpty {
		t.Errorf("expected row 3 (new) to be empty, got '%s'", planet.GardenBedState.Rows[3].Status)
	}
}

func TestFarmUpgrade_NewRowUsable(t *testing.T) {
	planet := NewPlanet("test-farm-upgrade-usable", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)

	planet.GardenBedState.Rows[0] = GardenBedRow{Status: GardenBedRowPlanted, PlantType: PlantWheat, Stage: 1, Weeds: 0, WaterTimer: 0}

	planet.ensureGardenBedState()

	planet.AddBuildingDirect("farm", 2)
	planet.ensureGardenBedState()

	result, err := GardenBedActionInternal(planet, "plant", 1, PlantWheat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success planting in new row, got error: %s", result.Error)
	}
	if result.Rows[1].Status != GardenBedRowPlanted {
		t.Errorf("expected row 1 to be planted, got '%s'", result.Rows[1].Status)
	}
	if result.Rows[1].PlantType != PlantWheat {
		t.Errorf("expected row 1 plant type to be wheat, got '%s'", result.Rows[1].PlantType)
	}
}
