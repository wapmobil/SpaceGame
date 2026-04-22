package building

// CompositeDrone produces composite resources.
type CompositeDrone struct {
	Building
}

// NewCompositeDrone creates a new CompositeDrone building.
func NewCompositeDrone(planetID string) *CompositeDrone {
	return &CompositeDrone{
		Building: Building{
			BuildingType:  TypeCompositeDrone,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
	}
}

// Consumption returns energy consumed per level (10 per level).
func (c *CompositeDrone) Consumption() int {
	return 10
}

// BuildTime returns the time to build next level.
func (c *CompositeDrone) BuildTime() float64 {
	return float64(c.BuildingLevel*c.BuildingLevel + 1) * 100
}

// Cost returns the food cost to build next level.
func (c *CompositeDrone) Cost() float64 {
	return float64(c.BuildingLevel*c.BuildingLevel + 1) * 100
}

// CostMulti returns the multi-resource cost to build next level.
func (c *CompositeDrone) CostMulti() CostMulti {
	level := c.BuildingLevel
	return CostMulti{
		Food:  float64(level*level+1) * 60,
		Money: float64(level*level+1) * 40,
	}
}

// Produce returns the composite production for one tick at current level.
func (c *CompositeDrone) Produce(level int) ProductionResult {
	if level <= 0 {
		return ProductionResult{}
	}
	return ProductionResult{
		HasEnergy: true,
		Composite: float64(level) * 0.5,
	}
}
