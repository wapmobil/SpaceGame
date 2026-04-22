package building

// BuildingType represents the type of building.
type BuildingType string

const (
	TypeFarm           BuildingType = "farm"
	TypeSolar          BuildingType = "solar"
	TypeStorage        BuildingType = "storage"
	TypeBase           BuildingType = "base"
	TypeFactory        BuildingType = "factory"
	TypeEnergyStorage  BuildingType = "energy_storage"
	TypeShipyard       BuildingType = "shipyard"
	TypeComCenter      BuildingType = "comcenter"
	TypeCompositeDrone BuildingType = "composite_drone"
	TypeMechanismFactory BuildingType = "mechanism_factory"
	TypeReagentLab     BuildingType = "reagent_lab"
)

// Building represents a game building with level and construction progress.
type Building struct {
	BuildingType  BuildingType `json:"type"`
	BuildingLevel int          `json:"level"`
	BuildProgress float64      `json:"build_progress"`
	PlanetID      string       `json:"planet_id"`
}

// New creates a new building with the given type.
func New(bt BuildingType, planetID string) *Building {
	return &Building{
		BuildingType:  bt,
		BuildingLevel: 0,
		BuildProgress: 0,
		PlanetID:      planetID,
	}
}

// GetType returns the building type.
func (b *Building) GetType() BuildingType {
	return b.BuildingType
}

// GetLevel returns the current building level.
func (b *Building) GetLevel() int {
	return b.BuildingLevel
}

// SetLevel sets the building level.
func (b *Building) SetLevel(l int) {
	b.BuildingLevel = l
}

// IsBuilding returns true if the building is currently under construction.
func (b *Building) IsBuilding() bool {
	return b.BuildProgress > 0
}

// Step advances construction progress. Called each game tick.
func (b *Building) Step(buildSpeed float64) {
	if b.BuildProgress > 0 {
		b.BuildProgress -= buildSpeed
		if b.BuildProgress <= 0 {
			b.BuildingLevel++
			b.BuildProgress = 0
		}
	}
}

// Consumption returns the energy consumption per level (positive = consumes, negative = produces).
func (b *Building) Consumption() int {
	return 0
}

// BuildTime returns the time needed to build the next level.
func (b *Building) BuildTime() float64 {
	return float64(b.BuildingLevel+2*(b.BuildingLevel*b.BuildingLevel))/1.0 + 10
}

// CostMulti represents the multi-resource cost to build the next level.
type CostMulti struct {
	Food  float64 `json:"food"`
	Money float64 `json:"money"`
}

// Cost returns the food cost to build the next level.
func (b *Building) Cost() float64 {
	return 0
}

// CostMulti returns the multi-resource cost to build the next level.
func (b *Building) CostMulti() CostMulti {
	return CostMulti{}
}

// Producer is the interface that all buildings implement.
type Producer interface {
	Consumption() int
	BuildTime() float64
	Cost() float64
	CostMulti() CostMulti
	Produce(level int) ProductionResult
}
