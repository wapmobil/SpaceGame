package game

import "spacegame/internal/game/building"

// getProduction returns the production result for a building at the given level.
func (p *Planet) getProduction(bt string, level int) building.ProductionResult {
	return building.Production(bt, level)
}

// calculateResourceProduction returns per-tick resource production from all operational buildings.
func (p *Planet) calculateResourceProduction() building.ProductionResult {
	var prod building.ProductionResult
	for _, b := range p.Buildings {
		if !b.IsWorking() {
			continue
		}
		prod.Add(p.getProduction(b.Type, b.Level))
	}
	return prod
}

// GetProductionResult returns the production result for this tick.
func (p *Planet) GetProductionResult() ProductionResult {
	return p.calculateResourceProduction()
}

// autoDisableDynamo disables dynamo buildings when food deficit is unsustainable.
func (p *Planet) autoDisableDynamo() {
	prod := p.calculateResourceProduction()
	if prod.Food >= 0 || p.Resources.Food > -prod.Food {
		return
	}
	for i := range p.Buildings {
		b := &p.Buildings[i]
		if b.Type == "dynamo" && b.IsWorking() {
			b.Enabled = false
		}
	}
}

// autoDisableBase disables the base building when food is depleted.
func (p *Planet) autoDisableBase() {
	if p.Resources.Food <= 0 {
		for i := range p.Buildings {
			b := &p.Buildings[i]
			if b.Type == "base" && b.IsWorking() {
				b.Enabled = false
			}
		}
	}
}
