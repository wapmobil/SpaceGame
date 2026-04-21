package building

// ReagentLab produces reagents resources.
type ReagentLab struct {
	Building
}

// NewReagentLab creates a new ReagentLab building.
func NewReagentLab(planetID string) *ReagentLab {
	return &ReagentLab{
		Building: Building{
			BuildingType:  TypeReagentLab,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
	}
}

// Consumption returns energy consumed per level (10 per level).
func (r *ReagentLab) Consumption() int {
	return 10
}

// BuildTime returns the time to build next level.
func (r *ReagentLab) BuildTime() float64 {
	return float64(r.BuildingLevel*r.BuildingLevel + 1) * 100
}

// Cost returns the food cost to build next level.
func (r *ReagentLab) Cost() float64 {
	return float64(r.BuildingLevel*r.BuildingLevel + 1) * 100
}

// Produce returns the reagents production for one tick at current level.
func (r *ReagentLab) Produce(level int) ProductionResult {
	if level <= 0 {
		return ProductionResult{}
	}
	return ProductionResult{
		HasEnergy: true,
		Reagents: float64(level) * 0.5,
	}
}
