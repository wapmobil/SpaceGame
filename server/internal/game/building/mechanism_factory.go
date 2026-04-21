package building

// MechanismFactory produces mechanisms resources.
type MechanismFactory struct {
	Building
}

// NewMechanismFactory creates a new MechanismFactory building.
func NewMechanismFactory(planetID string) *MechanismFactory {
	return &MechanismFactory{
		Building: Building{
			BuildingType:  TypeMechanismFactory,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
	}
}

// Consumption returns energy consumed per level (10 per level).
func (m *MechanismFactory) Consumption() int {
	return 10
}

// BuildTime returns the time to build next level.
func (m *MechanismFactory) BuildTime() float64 {
	return float64(m.BuildingLevel*m.BuildingLevel + 1) * 100
}

// Cost returns the food cost to build next level.
func (m *MechanismFactory) Cost() float64 {
	return float64(m.BuildingLevel*m.BuildingLevel + 1) * 100
}

// Produce returns the mechanisms production for one tick at current level.
func (m *MechanismFactory) Produce(level int) ProductionResult {
	if level <= 0 {
		return ProductionResult{}
	}
	return ProductionResult{
		HasEnergy: true,
		Mechanisms: float64(level) * 0.5,
	}
}
