package game

import (
	"spacegame/internal/game/building"
)

// ProductionResult is an alias for building.ProductionResult for backwards compatibility.
type ProductionResult = building.ProductionResult

// ResourceName represents the name of a game resource.
type ResourceName string

const (
	Food        ResourceName = "food"
	Composite   ResourceName = "composite"
	Mechanisms  ResourceName = "mechanisms"
	Reagents    ResourceName = "reagents"
	Energy      ResourceName = "energy"
	Money       ResourceName = "money"
	AlienTech   ResourceName = "alien_tech"
)

// ResourceInfo holds display and production info for a resource.
type ResourceInfo struct {
	Name  ResourceName
	Icon  string
	Emoji string
}

// AllResources returns the ordered list of all resource definitions.
func AllResources() []ResourceInfo {
	return []ResourceInfo{
		{Food, "food", "🍍"},
		{Composite, "composite", "🧱"},
		{Mechanisms, "mechanisms", "⚙️"},
		{Reagents, "reagents", "🛢"},
		{Energy, "energy", "⚡"},
		{Money, "money", "💰"},
		{AlienTech, "alien_tech", "📟"},
	}
}

// ResourceIcons maps resource names to their display icons.
func ResourceIcons() map[ResourceName]string {
	m := make(map[ResourceName]string)
	for _, r := range AllResources() {
		m[r.Name] = r.Icon
	}
	return m
}

// ResourceEmojis maps resource names to their display emojis.
func ResourceEmojis() map[ResourceName]string {
	m := make(map[ResourceName]string)
	for _, r := range AllResources() {
		m[r.Name] = r.Emoji
	}
	return m
}

// ResourceTypes returns the names of production resources (non-energy, non-currency).
func ResourceTypes() []ResourceName {
	return []ResourceName{Food, Composite, Mechanisms, Reagents}
}



// StorageResourceType represents a resource that can be stored.
type StorageResourceType string

const (
	StorageFood     StorageResourceType = "food"
	StorageComposite StorageResourceType = "composite"
	StorageMechanisms StorageResourceType = "mechanisms"
	StorageReagents  StorageResourceType = "reagents"
)

// AllStorageResources returns the list of storable resource types.
func AllStorageResources() []StorageResourceType {
	return []StorageResourceType{StorageFood, StorageComposite, StorageMechanisms, StorageReagents}
}

// StorageResourceEmojis maps storage resource types to their emojis.
func StorageResourceEmojis() map[StorageResourceType]string {
	return map[StorageResourceType]string{
		StorageFood:     "🍍",
		StorageComposite: "🧱",
		StorageMechanisms: "⚙️",
		StorageReagents:  "🛢",
	}
}



// CalculateEnergyBalance computes the net energy for a planet given its buildings.
func CalculateEnergyBalance(buildings map[building.BuildingType]*building.Building) (production float64, consumption float64) {
	for _, b := range buildings {
		con := b.Consumption()
		if con < 0 {
			production += float64(b.GetLevel()) * float64(-con)
		} else if con > 0 {
			consumption += float64(b.GetLevel()) * float64(con)
		}
	}
	return production, consumption
}

// CalculateStorageCapacity returns the total storage capacity for a resource type.
func CalculateStorageCapacity(buildings map[building.BuildingType]*building.Building, resource StorageResourceType) float64 {
	base := 1000.0
	for _, b := range buildings {
		switch b.GetType() {
		case building.TypeStorage:
			base += float64(b.GetLevel()) * 1000
		case building.TypeEnergyStorage:
			if resource == StorageFood {
				base += float64(b.GetLevel()) * 100
			}
		}
	}
	return base
}

// CalculateMaxEnergy returns the maximum energy capacity.
func CalculateMaxEnergy(buildings map[building.BuildingType]*building.Building) float64 {
	base := 100.0
	for _, b := range buildings {
		if b.GetType() == building.TypeEnergyStorage {
			base += float64(b.GetLevel()) * 100
		}
	}
	return base
}
