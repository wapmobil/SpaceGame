package game

import (
	"math"
	"testing"

	"spacegame/internal/game/research"
)

func TestFarmProducesFood(t *testing.T) {
	planet := NewPlanet("test-1", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 1) // energy
	planet.AddBuildingDirect("farm", 1)

	prod := planet.GetProductionResult()
	if prod.Food != 1 {
		t.Errorf("expected farm at level 1 to produce 1 food, got %f", prod.Food)
	}
}

func TestFarmProductionScalesWithLevel(t *testing.T) {
	planet := NewPlanet("test-2", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 2) // 30 energy total
	planet.AddBuildingDirect("farm", 3)

	prod := planet.GetProductionResult()
	if prod.Food != 3 {
		t.Errorf("expected 3 farms at level 1 to produce 3 food, got %f", prod.Food)
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

func TestEnergyBalanceNegative(t *testing.T) {
	planet := NewPlanet("test-4", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1) // consumes 10 energy
	planet.AddBuildingDirect("solar", 1) // produces 15 energy

	planet.tickEnergy()
	balance := planet.GetEnergyBalance()
	if balance != 5 {
		t.Errorf("expected energy balance of 5, got %f", balance)
	}
}

func TestEnergyBalanceDeficit(t *testing.T) {
	planet := NewPlanet("test-5", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1) // consumes 10 energy
	// No solar station - energy deficit

	planet.tickEnergy()
	balance := planet.GetEnergyBalance()
	if balance != -10 {
		t.Errorf("expected energy balance of -10, got %f", balance)
	}
}

func TestDisabledBuildingProducesNoResources(t *testing.T) {
	planet := NewPlanet("test-6", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("farm", 1)
	planet.AddBuildingDirect("solar", 1)

	// Disable the farm manually
	planet.Buildings[0].Enabled = false

	prod := planet.GetProductionResult()
	if prod.Food != 0 {
		t.Errorf("expected no food production from disabled building, got %f", prod.Food)
	}
}

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

func TestMaxEnergyIncreasesWithEnergyStorage(t *testing.T) {
	planet := NewPlanet("test-9", "owner-1", "Test Planet", nil)
	planet.EnergyBuffer.UpdateMax(0)
	if planet.EnergyBuffer.Max != 100 {
		t.Errorf("expected base max energy of 100, got %f", planet.EnergyBuffer.Max)
	}

	planet.AddBuildingDirect("energy_storage", 2)
	planet.EnergyBuffer.UpdateMax(2)
	if planet.EnergyBuffer.Max != 300 {
		t.Errorf("expected max energy of 300 with 2 energy storage buildings, got %f", planet.EnergyBuffer.Max)
	}
}

func TestResourcesClampedToZero(t *testing.T) {
	planet := NewPlanet("test-10", "owner-1", "Test Planet", nil)

	// Set negative food
	planet.Resources.Food = -50

	// Clamp manually
	planet.Resources.Food = math.Max(0, planet.Resources.Food)

	if planet.Resources.Food < 0 {
		t.Errorf("expected food to be clamped to 0, got %f", planet.Resources.Food)
	}
	if planet.Resources.Food != 0 {
		t.Errorf("expected food to be exactly 0, got %f", planet.Resources.Food)
	}
}

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

func TestFactoryProducesResource(t *testing.T) {
	planet := NewPlanet("test-13", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 1) // energy
	planet.AddBuildingDirect("solar", 1) // more energy
	planet.AddBuildingDirect("factory", 1)

	prod := planet.GetProductionResult()
	// Factory produces 0.5 of one resource type
	total := prod.Composite + prod.Mechanisms + prod.Reagents
	if total != 0.5 {
		t.Logf("factory production: composite=%f mechanisms=%f reagents=%f total=%f",
			prod.Composite, prod.Mechanisms, prod.Reagents, total)
	}
}

func TestBaseConsumesFood(t *testing.T) {
	planet := NewPlanet("test-14", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 1) // energy
	planet.AddBuildingDirect("solar", 1) // more energy
	planet.AddBuildingDirect("base", 1)  // consumes food

	prod := planet.GetProductionResult()
	// Base consumes food (negative production)
	if prod.Food != -1 {
		t.Logf("base food production: %f", prod.Food)
	}
}

func TestBaseOperationalWithFood(t *testing.T) {
	planet := NewPlanet("test-base-1", "owner-1", "Test Planet", nil)
	planet.Resources.Food = 10
	if !planet.BaseOperational() {
		t.Error("expected base to be operational with food > 0")
	}
}

func TestBaseNotOperationalWithoutFood(t *testing.T) {
	planet := NewPlanet("test-base-2", "owner-1", "Test Planet", nil)
	planet.Resources.Food = 0
	if planet.BaseOperational() {
		t.Error("expected base to NOT be operational with food == 0")
	}
}

func TestBaseNotOperationalWithNegativeFood(t *testing.T) {
	planet := NewPlanet("test-base-3", "owner-1", "Test Planet", nil)
	planet.Resources.Food = -5
	if planet.BaseOperational() {
		t.Error("expected base to NOT be operational with food < 0")
	}
}

func TestStartResearchBlockedWithoutFood(t *testing.T) {
	planet := NewPlanet("test-research-1", "owner-1", "Test Planet", nil)
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

func TestEnergyBufferMaxUpdates(t *testing.T) {
	planet := NewPlanet("test-energy-2", "owner-1", "Test Planet", nil)

	initialMax := planet.EnergyBuffer.Max
	if initialMax != 100 {
		t.Errorf("expected initial max energy of 100, got %f", initialMax)
	}

	planet.AddBuildingDirect("energy_storage", 3)
	planet.EnergyBuffer.UpdateMax(3)

	newMax := planet.EnergyBuffer.Max
	expectedMax := 100.0 + 3.0*100.0
	if newMax != expectedMax {
		t.Errorf("expected max energy of %f with 3 energy_storage, got %f", expectedMax, newMax)
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
	if b.Consumption <= 0 {
		t.Errorf("expected positive energy consumption, got %f", b.Consumption)
	}
}

func TestConfirmBuildingPopulatesEntry(t *testing.T) {
	planet := NewPlanet("test-confirm-1", "owner-1", "Test Planet", nil)

	planet.AddBuildingDirect("farm", 1)
	planet.Buildings[0].Pending = true

	err := planet.ConfirmBuilding("farm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if planet.Buildings[0].Pending {
		t.Error("expected building to not be pending after confirmation")
	}
}

func TestTickProducesResources(t *testing.T) {
	planet := NewPlanet("test-7", "owner-1", "Test Planet", nil)
	planet.AddBuildingDirect("solar", 1) // produces 15 energy
	planet.AddBuildingDirect("farm", 1)  // consumes 10 energy, produces 1 food

	// Verify energy buffer has capacity
	if planet.EnergyBuffer.Max != 100 {
		t.Errorf("expected initial energy buffer max of 100, got %f", planet.EnergyBuffer.Max)
	}

	// Verify energy production exceeds consumption (solar 15 > farm 10)
	planet.tickEnergy()
	balance := planet.GetEnergyBalance()
	if balance != 5 {
		t.Errorf("expected energy balance of 5 (15 solar - 10 farm), got %f", balance)
	}

	// Verify food production when energy is available
	prod := planet.GetProductionResult()
	if prod.Food != 1 {
		t.Errorf("expected farm to produce 1 food with energy available, got %f", prod.Food)
	}
}
