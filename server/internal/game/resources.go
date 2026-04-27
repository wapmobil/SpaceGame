package game

import "spacegame/internal/game/building"

// ProductionResult is an alias for building.ProductionResult for backwards compatibility.
type ProductionResult = building.ProductionResult

// ResourceName represents the name of a game resource.
type ResourceName string

const (
	Food        ResourceName = "food"
	Iron        ResourceName = "iron"
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
		{Iron, "iron", "🪨"},
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

// PlanetResources holds the current resource amounts.
type PlanetResources struct {
	Food            float64 `json:"food"`
	Iron            float64 `json:"iron"`
	Composite       float64 `json:"composite"`
	Mechanisms      float64 `json:"mechanisms"`
	Reagents        float64 `json:"reagents"`
	Energy          float64 `json:"energy"`
	MaxEnergy       float64 `json:"max_energy"`
	Money           float64 `json:"money"`
	AlienTech       float64 `json:"alien_tech"`
	StorageCapacity float64 `json:"storage_capacity"`
	ResearchUnlocks string  `json:"research_unlocks"`
}
