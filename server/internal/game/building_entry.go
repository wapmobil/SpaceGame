package game

import "spacegame/internal/game/building"

// BuildingEntry represents a single building on a planet with all computed data.
type BuildingEntry struct {
	Type          string       `json:"type"`
	Level         int          `json:"level"`
	BuildProgress float64      `json:"build_progress"` // -1 = working, 0..buildTime = under construction, 0 = build complete
	Enabled       bool         `json:"enabled"`
	BuildTime     float64      `json:"build_time"`     // total build time in seconds
	Cost          CostInfo     `json:"cost"`           // cost to build/upgrade to current level (already paid)
	NextCost      CostInfo     `json:"next_cost"`      // cost to build/upgrade to next level
	Production    building.ProductionResult `json:"production"` // per-tick resource production (energy includes consumption sign)
}

// CostInfo represents the cost to build/upgrade a building.
type CostInfo struct {
	Food  float64 `json:"food"`
	Money float64 `json:"money"`
}

// BuildingsOrder defines the display order of building types.
var BuildingsOrder = []string{
	"base", "farm", "solar", "storage", "factory",
	"energy_storage", "shipyard", "comcenter",
	"composite_drone", "mechanism_factory", "reagent_lab", "dynamo",
}

// BuildingResearchRequirements maps building types to their required research tech ID.
// Empty string means no research required.
var BuildingResearchRequirements = map[string]string{
	"energy_storage": "energy_storage",
	"shipyard":       "ships",
	"comcenter":      "expeditions",
}

// RandomUnlockBuildings lists buildings that are randomly unlocked by planet_exploration.
var RandomUnlockBuildings = []string{"composite_drone", "mechanism_factory", "reagent_lab"}

// IsBuildingUnlocked checks if a building type is unlocked by research.
// researchUnlocks is the building type unlocked by planet_exploration (for random unlocks).
func IsBuildingUnlocked(buildingType string, completed map[string]int, researchUnlocks string) bool {
	// Check fixed research requirements
	if req, ok := BuildingResearchRequirements[buildingType]; ok {
		if completed[req] <= 0 {
			return false
		}
	}

	// Check random unlocks (planet_exploration)
	for _, bt := range RandomUnlockBuildings {
		if buildingType == bt {
			if completed["planet_exploration"] <= 0 {
				return false
			}
			return researchUnlocks == bt
		}
	}

	return true
}

// IsBuilding returns true if the building is under construction.
func (b *BuildingEntry) IsBuilding() bool {
	return b.BuildProgress > 0 && b.BuildTime > 0
}

// IsBuildComplete returns true if the building is completed and waiting for confirmation.
func (b *BuildingEntry) IsBuildComplete() bool {
	return b.BuildProgress == 0 && b.BuildTime > 0
}

// IsWorking returns true if the building is operational.
func (b *BuildingEntry) IsWorking() bool {
	return !b.IsBuilding() && !b.IsBuildComplete() && b.Enabled && b.Level > 0
}
