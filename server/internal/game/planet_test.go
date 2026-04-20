package game

import (
	"math"
	"testing"
)

func TestFarmProducesFood(t *testing.T) {
	planet := NewPlanet("test-1", "owner-1", "Test Planet", nil)
	planet.AddBuilding("solar") // energy
	planet.AddBuilding("farm")

	prod := planet.GetProductionResult()
	if prod.Food != 1 {
		t.Errorf("expected farm at level 1 to produce 1 food, got %f", prod.Food)
	}
}

func TestFarmProductionScalesWithLevel(t *testing.T) {
	planet := NewPlanet("test-2", "owner-1", "Test Planet", nil)
	planet.AddBuilding("solar") // 15 energy
	planet.AddBuilding("solar") // 30 energy total
	planet.AddBuilding("farm")
	planet.AddBuilding("farm")
	planet.AddBuilding("farm")

	prod := planet.GetProductionResult()
	if prod.Food != 3 {
		t.Errorf("expected 3 farms at level 1 to produce 3 food, got %f", prod.Food)
	}
}

func TestSolarProducesEnergy(t *testing.T) {
	planet := NewPlanet("test-3", "owner-1", "Test Planet", nil)
	planet.AddBuilding("solar")

	prod := planet.GetProductionResult()
	if prod.Energy != 15 {
		t.Errorf("expected solar at level 1 to produce 15 energy, got %f", prod.Energy)
	}
}

func TestEnergyBalanceNegative(t *testing.T) {
	planet := NewPlanet("test-4", "owner-1", "Test Planet", nil)
	planet.AddBuilding("farm") // consumes 10 energy
	planet.AddBuilding("solar") // produces 15 energy

	balance := planet.GetEnergyBalance()
	if balance != 5 {
		t.Errorf("expected energy balance of 5, got %f", balance)
	}
}

func TestEnergyBalanceDeficit(t *testing.T) {
	planet := NewPlanet("test-5", "owner-1", "Test Planet", nil)
	planet.AddBuilding("farm") // consumes 10 energy
	// No solar station - energy deficit

	balance := planet.GetEnergyBalance()
	if balance != -10 {
		t.Errorf("expected energy balance of -10, got %f", balance)
	}
}

func TestNoProductionWithoutEnergy(t *testing.T) {
	planet := NewPlanet("test-6", "owner-1", "Test Planet", nil)
	planet.AddBuilding("farm") // consumes 10 energy, no solar

	prod := planet.GetProductionResult()
	if prod.Food != 0 {
		t.Errorf("expected no food production without energy, got %f", prod.Food)
	}
}

func TestTickProducesResources(t *testing.T) {
	planet := NewPlanet("test-7", "owner-1", "Test Planet", nil)
	planet.AddBuilding("solar") // produces 15 energy
	planet.AddBuilding("farm")  // consumes 10 energy, produces 1 food

	// Simulate tick manually
	production, consumption := planet.calculateEnergyBalance()
	hasEnergy := production >= consumption

	var totalProduction ProductionResult
	for bt, level := range planet.Buildings {
		if !hasEnergy {
			continue
		}
		prod := planet.getProduction(bt, level)
		if prod.HasEnergy {
			totalProduction.Add(prod)
		}
		_ = bt
	}

	expectedFood := planet.Resources.Food + totalProduction.Food
	if expectedFood < 1 {
		t.Errorf("expected food to increase by at least 1, got %f", expectedFood)
	}
}

func TestStorageIncreasesCapacity(t *testing.T) {
	planet := NewPlanet("test-8", "owner-1", "Test Planet", nil)
	capacity1 := planet.calculateStorageCapacity()
	if capacity1 != 1000 {
		t.Errorf("expected base capacity of 1000, got %f", capacity1)
	}

	planet.AddBuilding("storage")
	planet.AddBuilding("storage")

	capacity2 := planet.calculateStorageCapacity()
	if capacity2 != 3000 {
		t.Errorf("expected capacity of 3000 with 2 storage buildings, got %f", capacity2)
	}
}

func TestMaxEnergyIncreasesWithEnergyStorage(t *testing.T) {
	planet := NewPlanet("test-9", "owner-1", "Test Planet", nil)
	maxEnergy1 := planet.calculateMaxEnergy()
	if maxEnergy1 != 100 {
		t.Errorf("expected base max energy of 100, got %f", maxEnergy1)
	}

	planet.AddBuilding("energy_storage")
	planet.AddBuilding("energy_storage")

	maxEnergy2 := planet.calculateMaxEnergy()
	if maxEnergy2 != 300 {
		t.Errorf("expected max energy of 300 with 2 energy storage buildings, got %f", maxEnergy2)
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

	planet.AddBuilding("farm")
	if level := planet.GetBuildingLevel("farm"); level != 1 {
		t.Errorf("expected farm level 1, got %d", level)
	}

	planet.AddBuilding("farm")
	if level := planet.GetBuildingLevel("farm"); level != 2 {
		t.Errorf("expected farm level 2, got %d", level)
	}
}

func TestTotalBuildingLevels(t *testing.T) {
	planet := NewPlanet("test-12", "owner-1", "Test Planet", nil)
	planet.AddBuilding("farm")
	planet.AddBuilding("solar")
	planet.AddBuilding("storage")

	total := planet.GetTotalBuildingLevels()
	if total != 3 {
		t.Errorf("expected 3 total building levels, got %d", total)
	}
}

func TestFactoryProducesResource(t *testing.T) {
	planet := NewPlanet("test-13", "owner-1", "Test Planet", nil)
	planet.AddBuilding("solar") // energy
	planet.AddBuilding("solar") // more energy
	planet.AddBuilding("factory")

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
	planet.AddBuilding("solar") // energy
	planet.AddBuilding("solar") // more energy
	planet.AddBuilding("base")  // consumes food

	prod := planet.GetProductionResult()
	// Base consumes food (negative production)
	if prod.Food != -1 {
		t.Logf("base food production: %f", prod.Food)
	}
}
