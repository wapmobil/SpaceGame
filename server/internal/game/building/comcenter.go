package building

// ComCenter is the command center that unlocks advanced tech.
type ComCenter struct {
	Building
}

// NewComCenter creates a new ComCenter building.
func NewComCenter(planetID string) *ComCenter {
	return &ComCenter{
		Building: Building{
			BuildingType:  TypeComCenter,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
	}
}

// Consumption returns energy consumed per level (100 per level).
func (c *ComCenter) Consumption() int {
	return 100
}

// BuildTime returns the time to build next level.
func (c *ComCenter) BuildTime() float64 {
	return float64(c.BuildingLevel) + 10000
}

// Cost returns the food cost to build next level.
func (c *ComCenter) Cost() float64 {
	if c.BuildingLevel == 0 {
		return 10000000
	}
	return 10000000 * float64(c.BuildingLevel)
}

// CostMulti returns the multi-resource cost to build next level.
func (c *ComCenter) CostMulti() CostMulti {
	level := c.BuildingLevel
	return CostMulti{
		Food:  float64(level) * 10000,
		Money: float64(level) * 10000,
	}
}

// Produce returns no production (comcenter only unlocks features).
func (c *ComCenter) Produce(level int) ProductionResult {
	return ProductionResult{}
}
