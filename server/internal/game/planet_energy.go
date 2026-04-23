package game

import (
	"math"

	"spacegame/internal/game/building"
)

// getEnergyConsumption returns the energy consumption for a building at the given level.
// Solar buildings return negative values (representing production).
// Energy storage buildings return small positive values (representing consumption).
func (p *Planet) getEnergyConsumption(bt string, level int) float64 {
	return building.EnergyConsumption(bt, level)
}

// calculateEnergy computes total energy production and consumption in a single pass.
// Returns (production, consumption) where production is from solar panels
// and consumption is from all other enabled buildings + fleet.
func (p *Planet) calculateEnergy() (production, consumption float64) {
	for _, b := range p.Buildings {
		if !b.IsWorking() {
			continue
		}
		con := p.getEnergyConsumption(b.Type, b.Level)
		if con < 0 {
			production += -con
		} else {
			consumption += con
		}
	}
	consumption += p.Fleet.TotalEnergyConsumption()
	return
}

// tickEnergy processes energy production, consumption, clamping, and auto-disable logic.
func (p *Planet) tickEnergy() {
	// Update energy buffer max capacity based on energy_storage level
	energyStorageLevel := 0
	for _, b := range p.Buildings {
		if b.Type == "energy_storage" && b.IsWorking() {
			energyStorageLevel += b.Level
		}
	}
	p.EnergyBuffer.UpdateMax(energyStorageLevel)

	// Calculate energy production and consumption in single pass
	energyProduction, energyConsumption := p.calculateEnergy()

	p.EnergyBuffer.Value += energyProduction
	p.EnergyBuffer.Value -= energyConsumption

	// Clamp buffer to max
	if p.EnergyBuffer.Value > p.EnergyBuffer.Max {
		p.EnergyBuffer.Value = p.EnergyBuffer.Max
	}

	// Auto-disable buildings when energy buffer is empty
	if p.EnergyBuffer.Value <= 0 {
		for i := range p.Buildings {
			b := &p.Buildings[i]
			if b.IsBuilding() || b.IsBuildComplete() {
				continue
			}
			if b.Type == "solar" || b.Type == "energy_storage" {
				continue
			}
			if b.Enabled {
				b.Enabled = false
			}
		}
	}

	// Update max energy for display
	p.Resources.MaxEnergy = p.EnergyBuffer.Max

	// Update energy balance for display
	p.EnergyBalance = energyProduction - energyConsumption
}

// tickResources processes resource production and clamping.
func (p *Planet) tickResources() {
	// Calculate resource production (only from enabled buildings)
	totalProduction := p.calculateResourceProduction()

	// Apply resource production to resources (energy is handled in tickEnergy)
	p.Resources.Energy = p.EnergyBuffer.Value
	p.Resources.Food += totalProduction.Food
	p.Resources.Composite += totalProduction.Composite
	p.Resources.Mechanisms += totalProduction.Mechanisms
	p.Resources.Reagents += totalProduction.Reagents
	p.Resources.Money += totalProduction.Money
	p.Resources.AlienTech += totalProduction.AlienTech

	// Clamp resources to storage capacity
	storageCapacity := p.CalculateStorageCapacity()
	p.Resources.Food = math.Min(p.Resources.Food, storageCapacity)
	p.Resources.Composite = math.Min(p.Resources.Composite, storageCapacity)
	p.Resources.Mechanisms = math.Min(p.Resources.Mechanisms, storageCapacity)
	p.Resources.Reagents = math.Min(p.Resources.Reagents, storageCapacity)

	// Clamp resources to non-negative
	p.Resources.Food = math.Max(0, p.Resources.Food)
	p.Resources.Composite = math.Max(0, p.Resources.Composite)
	p.Resources.Mechanisms = math.Max(0, p.Resources.Mechanisms)
	p.Resources.Reagents = math.Max(0, p.Resources.Reagents)
	p.Resources.Money = math.Max(0, p.Resources.Money)
	p.Resources.AlienTech = math.Max(0, p.Resources.AlienTech)
}
