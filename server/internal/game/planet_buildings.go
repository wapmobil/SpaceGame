package game

import (
	"fmt"
	"log"

	"spacegame/internal/game/building"
)

// FindBuildingIndex returns the index of a building by type, or -1 if not found.
func (p *Planet) FindBuildingIndex(bt string) int {
	for i, b := range p.Buildings {
		if b.Type == bt {
			return i
		}
	}
	return -1
}

// AddBuilding adds or upgrades a building on the planet.
func (p *Planet) AddBuilding(bt string) (float64, float64, error) {
	idx := p.FindBuildingIndex(bt)
	currentLevel := 0
	if idx >= 0 {
		currentLevel = p.Buildings[idx].Level
	}

	// Check if already under construction
	var alreadyConstructing bool
	if idx >= 0 && p.Buildings[idx].IsBuilding() {
		alreadyConstructing = true
	}

	// Check prerequisites: farm must be built first, solar after farm
	hasFarm := false
	hasSolar := false
	for _, b := range p.Buildings {
		if b.Type == "farm" && b.Level > 0 {
			hasFarm = true
		}
		if b.Type == "solar" && b.Level > 0 {
			hasSolar = true
		}
	}
	if bt != "farm" {
		if !hasFarm {
			return 0, 0, &PlanetError{PlanetID: p.ID, Reason: "prerequisite_missing", Extra: "Build a Farm first."}
		}
	}
	if bt != "farm" && bt != "solar" {
		if !hasSolar {
			return 0, 0, &PlanetError{PlanetID: p.ID, Reason: "prerequisite_missing", Extra: "Build a Solar Panel first."}
		}
	}

	// Check max concurrent construction limit
	if p.ActiveConstruction >= p.GetMaxConcurrentBuildings() {
		return 0, 0, &PlanetError{PlanetID: p.ID, Reason: "max_constructions_reached", Extra: fmt.Sprintf("Max constructions reached (%d/%d). Research Parallel Construction to unlock more.", p.ActiveConstruction, p.GetMaxConcurrentBuildings())}
	}

	// Get cost for next level
	cost := p.GetBuildingCost(bt, currentLevel)

	// Check affordability
	if cost.Food > p.Resources.Food {
		return 0, 0, &PlanetError{PlanetID: p.ID, Reason: "insufficient_food", Extra: fmt.Sprintf("Need %.0f food, have %.0f", cost.Food, p.Resources.Food)}
	}
	if cost.Money > p.Resources.Money {
		return 0, 0, &PlanetError{PlanetID: p.ID, Reason: "insufficient_money", Extra: fmt.Sprintf("Need %.0f money, have %.0f", cost.Money, p.Resources.Money)}
	}

	// Deduct resources
	p.Resources.Food -= cost.Food
	p.Resources.Money -= cost.Money

	// Upgrade or add building
	if idx >= 0 {
		p.Buildings[idx].Level = currentLevel + 1
		p.Buildings[idx].BuildProgress = p.GetBuildTime(bt, currentLevel+1)
		p.Buildings[idx].Enabled = false
		p.PopulateBuildingEntry(idx)
	} else {
		p.Buildings = append(p.Buildings, BuildingEntry{
			Type:          bt,
			Level:         1,
			BuildProgress: p.GetBuildTime(bt, 1),
			Enabled:       false,
		})
		p.PopulateBuildingEntry(len(p.Buildings) - 1)
	}

	// Track active construction
	if !alreadyConstructing {
		p.ActiveConstruction++
	}

	return cost.Food, cost.Money, nil
}

// AddBuildingDirect sets a building level without any checks (for testing).
func (p *Planet) AddBuildingDirect(bt string, level int) {
	idx := p.FindBuildingIndex(bt)
	if idx >= 0 {
		p.Buildings[idx].Level = level
		p.Buildings[idx].BuildProgress = -1
		p.Buildings[idx].Enabled = true
		p.PopulateBuildingEntry(idx)
	} else {
		p.Buildings = append(p.Buildings, BuildingEntry{
			Type:          bt,
			Level:         level,
			BuildProgress: -1,
			Enabled:       true,
		})
		p.PopulateBuildingEntry(len(p.Buildings) - 1)
	}
}

// GetBuildingLevel returns the level of a building.
func (p *Planet) GetBuildingLevel(bt string) int {
	idx := p.FindBuildingIndex(bt)
	if idx < 0 {
		return 0
	}
	return p.Buildings[idx].Level
}

// GetMaxConcurrentBuildings returns the maximum number of simultaneous construction projects.
func (p *Planet) GetMaxConcurrentBuildings() int {
	max := 1
	if lvl, ok := p.Research.GetCompleted()["parallel_construction"]; ok && lvl > 0 {
		max += lvl
	}
	return max
}

// ConfirmBuilding confirms a completed building construction, making it operational.
func (p *Planet) ConfirmBuilding(bt string) error {
	idx := p.FindBuildingIndex(bt)
	if idx < 0 {
		return &PlanetError{PlanetID: p.ID, Reason: "building_not_found"}
	}
	if !p.Buildings[idx].IsBuildComplete() {
		return &PlanetError{PlanetID: p.ID, Reason: "building_not_ready"}
	}
	p.Buildings[idx].BuildProgress = -1
	p.Buildings[idx].Enabled = true
	p.PopulateBuildingEntry(idx)

	if p.game != nil && p.game.db != nil {
		_, err := p.game.db.Exec(`
			UPDATE buildings SET build_progress = -1, enabled = true, updated_at = NOW()
			WHERE planet_id = $1 AND type = $2
		`, p.ID, bt)
		if err != nil {
			log.Printf("Error saving building confirmation for %s on planet %s: %v", bt, p.ID, err)
		}
	}

	return nil
}

// GetPendingBuildings returns all completed-but-unconfirmed building types.
func (p *Planet) GetPendingBuildings() map[string]bool {
	result := make(map[string]bool)
	for _, b := range p.Buildings {
		if b.IsBuildComplete() {
			result[b.Type] = true
		}
	}
	return result
}

// GetBuildTime returns the build time for a building at the given level.
func (p *Planet) GetBuildTime(bt string, level int) float64 {
	return building.BuildTime(bt, level)
}

// GetBuildingCost returns the multi-resource cost to build a building at the given level.
func (p *Planet) GetBuildingCost(bt string, level int) building.CostMulti {
	return building.Cost(bt, level)
}

// stepBuildingEntry steps a building by index.
func (p *Planet) stepBuildingEntry(idx int) {
	if idx < 0 || idx >= len(p.Buildings) {
		return
	}
	b := &p.Buildings[idx]
	if b.IsBuilding() {
		b.BuildProgress -= p.BuildSpeed
		if b.BuildProgress < 0 {
			b.BuildProgress = 0
		}
		if !b.IsBuilding() && b.BuildProgress >= 0 {
			// Construction completed
			p.ActiveConstruction--
			if p.ActiveConstruction < 0 {
				p.ActiveConstruction = 0
			}
		}
	}
}

// PopulateBuildingEntry populates all computed fields for a building entry.
func (p *Planet) PopulateBuildingEntry(idx int) {
	if idx < 0 || idx >= len(p.Buildings) {
		return
	}
	b := &p.Buildings[idx]
	level := b.Level

	b.BuildTime = p.GetBuildTime(b.Type, level)
	cost := p.GetBuildingCost(b.Type, level-1)
	b.Cost = CostInfo{Food: cost.Food, Money: cost.Money}
	nextCost := p.GetBuildingCost(b.Type, level)
	b.NextCost = CostInfo{Food: nextCost.Food, Money: nextCost.Money}

	if b.Enabled {
		b.Production = p.getProduction(b.Type, level)
		b.Production.Energy = -p.getEnergyConsumption(b.Type, level)
	} else {
		b.Production = building.ProductionResult{}
	}
}

// CalculateStorageCapacity returns the total storage capacity.
func (p *Planet) CalculateStorageCapacity() float64 {
	base := 1000.0
	if idx := p.FindBuildingIndex("storage"); idx >= 0 {
		base += float64(p.Buildings[idx].Level) * 1000
	}
	return base
}

// GetTotalBuildingLevels returns the sum of all building levels.
func (p *Planet) GetTotalBuildingLevels() int {
	total := 0
	for _, b := range p.Buildings {
		total += b.Level
	}
	return total
}

// BaseOperational returns true if the planet has food available.
func (p *Planet) BaseOperational() bool {
	return p.Resources.Food > 0
}
