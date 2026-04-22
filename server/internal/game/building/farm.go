package building

// Farm produces food and consumes energy.
type Farm struct {
	Building
}

// NewFarm creates a new Farm building.
func NewFarm(planetID string) *Farm {
	return &Farm{
		Building: Building{
			BuildingType:  TypeFarm,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
	}
}

// Consumption returns energy consumed per level (10 per level).
func (f *Farm) Consumption() int {
	return 10
}

// BuildTime returns the time to build next level.
func (f *Farm) BuildTime() float64 {
	return float64(f.BuildingLevel*f.BuildingLevel*f.BuildingLevel*20 + 100)
}

// Cost returns the food cost to build next level.
func (f *Farm) Cost() float64 {
	return float64(f.BuildingLevel*f.BuildingLevel*f.BuildingLevel*20 + 100)
}

// CostMulti returns the multi-resource cost to build next level.
func (f *Farm) CostMulti() CostMulti {
	level := f.BuildingLevel
	return CostMulti{
		Food:  float64(level*level*level*20 + 100),
		Money: float64(level*level*level*10 + 50),
	}
}

// Produce returns the food production for one tick at current level.
func (f *Farm) Produce(level int) ProductionResult {
	if level <= 0 {
		return ProductionResult{}
	}
	return ProductionResult{
		Food:      float64(level),
		HasEnergy: true,
	}
}
