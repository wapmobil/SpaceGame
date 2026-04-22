package building

import (
	"math/rand"
)

// ResourceType represents a resource type that a factory can produce.
type ResourceType string

const (
	ResourceComposite ResourceType = "composite"
	ResourceMechanisms ResourceType = "mechanisms"
	ResourceReagents  ResourceType = "reagents"
)

var factoryResourceTypes = []ResourceType{
	ResourceComposite,
	ResourceMechanisms,
	ResourceReagents,
}

// Factory produces random construction resources.
type Factory struct {
	Building
	ResourceType ResourceType
	ProdCounter  float64
}

// NewFactory creates a new Factory building with a random resource type.
func NewFactory(planetID string) *Factory {
	r := factoryResourceTypes[rand.Intn(len(factoryResourceTypes))]
	return &Factory{
		Building: Building{
			BuildingType:  TypeFactory,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
		ResourceType: r,
		ProdCounter:  0,
	}
}

// Consumption returns energy consumed per level (25 per level).
func (f *Factory) Consumption() int {
	return 25
}

// BuildTime returns the time to build next level.
func (f *Factory) BuildTime() float64 {
	return float64(f.BuildingLevel*2+1) * 100000
}

// Cost returns the food cost to build next level.
func (f *Factory) Cost() float64 {
	return float64(f.BuildingLevel*2+1) * 100000
}

// CostMulti returns the multi-resource cost to build next level.
func (f *Factory) CostMulti() CostMulti {
	level := f.BuildingLevel
	return CostMulti{
		Food:  float64(level*2+1) * 2500,
		Money: float64(level*2+1) * 1500,
	}
}

// Produce returns the resource production for one tick at current level.
func (f *Factory) Produce(level int) ProductionResult {
	if level <= 0 {
		return ProductionResult{}
	}
	// Production rate: 0.5 per level per second
	result := ProductionResult{HasEnergy: true}
	switch f.ResourceType {
	case ResourceComposite:
		result.Composite = float64(level) * 0.5
	case ResourceMechanisms:
		result.Mechanisms = float64(level) * 0.5
	case ResourceReagents:
		result.Reagents = float64(level) * 0.5
	}
	return result
}

// GetResourceType returns the resource this factory produces.
func (f *Factory) GetResourceType() ResourceType {
	return f.ResourceType
}

// SetResourceType sets the resource this factory produces.
func (f *Factory) SetResourceType(r ResourceType) {
	f.ResourceType = r
}
