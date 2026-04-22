package building

// EnergyStorage stores excess energy.
type EnergyStorage struct {
	Building
	EnergyStored float64
	UpgradeLevel float64
}

// NewEnergyStorage creates a new EnergyStorage building.
func NewEnergyStorage(planetID string) *EnergyStorage {
	return &EnergyStorage{
		Building: Building{
			BuildingType:  TypeEnergyStorage,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
		EnergyStored: 0,
		UpgradeLevel: 1.0,
	}
}

// Consumption returns energy consumed per level (2 per level).
func (e *EnergyStorage) Consumption() int {
	return 2
}

// BuildTime returns the time to build next level.
func (e *EnergyStorage) BuildTime() float64 {
	return float64(e.BuildingLevel*e.BuildingLevel) + 1000
}

// Cost returns the food cost to build next level.
func (e *EnergyStorage) Cost() float64 {
	return float64(e.BuildingLevel*e.BuildingLevel) * 1000
}

// CostMulti returns the multi-resource cost to build next level.
func (e *EnergyStorage) CostMulti() CostMulti {
	level := e.BuildingLevel
	return CostMulti{
		Food:  float64(level*level) * 300,
		Money: float64(level*level) * 200,
	}
}

// Produce returns no direct production (energy storage only stores).
func (e *EnergyStorage) Produce(level int) ProductionResult {
	return ProductionResult{}
}

// Capacity returns the energy storage capacity at the given level.
func (e *EnergyStorage) Capacity(level int) float64 {
	if level <= 0 {
		return 0
	}
	return float64(level) * 100 * e.UpgradeLevel
}

// AddEnergy adds energy to storage, clamping to capacity.
func (e *EnergyStorage) AddEnergy(amount float64) {
	if e.BuildingLevel <= 0 {
		return
	}
	e.EnergyStored += amount
	if e.EnergyStored < 0 {
		e.EnergyStored = 0
	}
	if e.EnergyStored > e.Capacity(e.BuildingLevel) {
		e.EnergyStored = e.Capacity(e.BuildingLevel)
	}
}
