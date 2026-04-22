package building

// Storage increases resource storage capacity.
type Storage struct {
	Building
}

// NewStorage creates a new Storage building.
func NewStorage(planetID string) *Storage {
	return &Storage{
		Building: Building{
			BuildingType:  TypeStorage,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
	}
}

// Consumption returns 0 (no energy cost).
func (s *Storage) Consumption() int {
	return 0
}

// BuildTime returns the time to build next level.
func (s *Storage) BuildTime() float64 {
	return float64(s.BuildingLevel*s.BuildingLevel+1) * 100
}

// Cost returns the food cost to build next level.
func (s *Storage) Cost() float64 {
	return float64(s.BuildingLevel*s.BuildingLevel+1) * 100
}

// CostMulti returns the multi-resource cost to build next level.
func (s *Storage) CostMulti() CostMulti {
	level := s.BuildingLevel
	return CostMulti{
		Food:  float64(level*level+1) * 60,
		Money: float64(level*level+1) * 40,
	}
}

// Produce returns no production (storage only increases capacity).
func (s *Storage) Produce(level int) ProductionResult {
	return ProductionResult{}
}

// Capacity returns the storage capacity added by this building at the given level.
func (s *Storage) Capacity(level int) float64 {
	if level <= 0 {
		return 1000
	}
	return float64(level) * 1000
}
