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

// calculateEnergyProduction returns total energy production from solar panels.
func (p *Planet) calculateEnergyProduction() float64 {
	var total float64
	for _, b := range p.Buildings {
		if b.Type == "solar" && b.BuildProgress <= 0 && !b.Pending && b.Enabled {
			con := p.getEnergyConsumption(b.Type, b.Level)
			if con < 0 {
				total += -con
			}
		}
	}
	return total
}

// calculateEnergyConsumption returns total energy consumption from non-solar buildings and fleet.
func (p *Planet) calculateEnergyConsumption() float64 {
	var total float64
	for _, b := range p.Buildings {
		if b.Type != "solar" && b.BuildProgress <= 0 && !b.Pending && b.Enabled {
			total += p.getEnergyConsumption(b.Type, b.Level)
		}
	}
	total += p.Fleet.TotalEnergyConsumption()
	return total
}

// calculateMaxEnergy returns the max energy based on energy_storage level.
func (p *Planet) calculateMaxEnergy() float64 {
	energyStorageLevel := 0
	for _, b := range p.Buildings {
		if b.Type == "energy_storage" {
			energyStorageLevel += b.Level
		}
	}
	return float64(100 + energyStorageLevel*100)
}

// tickEnergy processes energy production, consumption, clamping, and auto-disable logic.
func (p *Planet) tickEnergy() {
	// Update energy buffer max capacity based on energy_storage level
	energyStorageLevel := 0
	for _, b := range p.Buildings {
		if b.Type == "energy_storage" {
			energyStorageLevel += b.Level
		}
	}
	p.EnergyBuffer.UpdateMax(energyStorageLevel)

	// Calculate and apply energy production
	energyProduction := p.calculateEnergyProduction()
	p.EnergyBuffer.Value += energyProduction

	// Calculate and apply energy consumption
	energyConsumption := p.calculateEnergyConsumption()
	p.EnergyBuffer.Value -= energyConsumption

	// Clamp buffer to max
	if p.EnergyBuffer.Value > p.EnergyBuffer.Max {
		p.EnergyBuffer.Value = p.EnergyBuffer.Max
	}

	// Auto-disable buildings when energy buffer is empty
	if p.EnergyBuffer.Value <= 0 {
		for i := range p.Buildings {
			b := &p.Buildings[i]
			if b.BuildProgress > 0 || b.Pending {
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
