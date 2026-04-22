package building

import "math"

// Base is the headquarters building.
type Base struct {
	Building
	Taxes int
}

// NewBase creates a new Base building.
func NewBase(planetID string) *Base {
	return &Base{
		Building: Building{
			BuildingType:  TypeBase,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
		Taxes: 1,
	}
}

// Consumption returns energy consumed per level (20 per level).
func (b *Base) Consumption() int {
	return 20
}

// BuildTime returns the time to build next level.
func (b *Base) BuildTime() float64 {
	return math.Pow(2, float64(b.Building.BuildingLevel+3)) + 100
}

// Cost returns the food cost to build next level.
func (b *Base) Cost() float64 {
	return math.Pow(2, float64(b.Building.BuildingLevel+3))
}

// CostMulti returns the multi-resource cost to build next level.
func (b *Base) CostMulti() CostMulti {
	level := b.Building.BuildingLevel
	return CostMulti{
		Food:  math.Pow(2, float64(level+2)),
		Money: math.Pow(2, float64(level+3)),
	}
}

// Produce returns the food consumption (negative production) for one tick at current level.
func (b *Base) Produce(level int) ProductionResult {
	if level <= 0 {
		return ProductionResult{}
	}
	eatFood := float64(b.Taxes * level)
	return ProductionResult{
		Food:      -eatFood,
		HasEnergy: true,
	}
}

// EatFood returns the food consumed by the base at the given level.
func (b *Base) EatFood(level int) float64 {
	return float64(b.Taxes * level)
}
