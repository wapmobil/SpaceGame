package building

// Solar produces energy.
type Solar struct {
	Building
}

// NewSolar creates a new Solar building.
func NewSolar(planetID string) *Solar {
	return &Solar{
		Building: Building{
			BuildingType:  TypeSolar,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
	}
}

// Consumption returns negative value (energy production) per level (-15 per level).
func (s *Solar) Consumption() int {
	return -15
}

// BuildTime returns the time to build next level.
func (s *Solar) BuildTime() float64 {
	return float64(s.BuildingLevel*s.BuildingLevel*200 + 80)
}

// Cost returns the food cost to build next level.
func (s *Solar) Cost() float64 {
	return float64(s.BuildingLevel*s.BuildingLevel*200 + 80)
}

// CostMulti returns the multi-resource cost to build next level.
func (s *Solar) CostMulti() CostMulti {
	level := s.BuildingLevel
	return CostMulti{
		Food:  float64(level*level*120 + 48),
		Money: float64(level*level*80 + 32),
	}
}

// Produce returns the energy production for one tick at current level.
func (s *Solar) Produce(level int) ProductionResult {
	if level <= 0 {
		return ProductionResult{}
	}
	return ProductionResult{
		Energy:    float64(level) * 15,
		HasEnergy: true,
	}
}
