package game

import (
	"time"

	"spacegame/internal/game/building"
)

// GetState returns the planet's state as a JSON-serializable map.
func (p *Planet) GetState() map[string]interface{} {
	shipyardLevel := p.GetBuildingLevel("shipyard")
	baseLevel := p.GetBuildingLevel("base")

	// Populate building production/consumption for WS state updates
	for i := range p.Buildings {
		p.PopulateBuildingEntry(i)
	}

	// Build research state for WS updates
	researchStates := make([]map[string]interface{}, 0)
	for techID, state := range p.Research.States {
		researchStates = append(researchStates, map[string]interface{}{
			"tech_id":      techID,
			"completed":    state.Completed,
			"in_progress":  state.InProgress,
			"progress":     state.TotalTime - state.Progress,
			"total_time":   state.TotalTime,
			"progress_pct": p.Research.GetResearchProgress(techID),
		})
	}

	// Build farm state for WS updates
	var farmState map[string]interface{}
	if p.FarmState == nil {
		farmLevel := p.GetBuildingLevel("farm")
		if farmLevel > 0 {
			p.FarmState = NewFarmState(farmLevel)
		}
	}
	if p.FarmState != nil && p.FarmState.RowCount > 0 {
		farmState = map[string]interface{}{
			"rows":      p.FarmState.Rows,
			"last_tick": p.FarmState.LastTick,
			"row_count": p.FarmState.RowCount,
		}
	}

	result := map[string]interface{}{
		"id":               p.ID,
		"owner_id":         p.OwnerID,
		"name":             p.Name,
		"level":            p.Level,
		"resources":        p.Resources,
		"buildings":        p.Buildings,
		"build_speed":      p.BuildSpeed,
		"energy_balance":   p.EnergyBalance,
		"energy_buffer":    p.EnergyBuffer,
		"last_tick":        p.LastTick.Format(time.RFC3339),
		"fleet":            p.Fleet.GetShipState(),
		"shipyard_level":   shipyardLevel,
		"shipyard_queue":   p.Shipyard.Queue,
		"shipyard_slots":   p.Fleet.TotalSlots(),
		"shipyard_max":     p.Shipyard.MaxSlots(baseLevel),
		"shipyard_progress": p.Shipyard.GetQueueProgress(),
		"expeditions":       p.GetExpeditionState(),
		"active_expeditions":   p.GetActiveExpeditionsCount(),
		"max_expeditions":      p.GetMaxExpeditions(),
		"active_constructions": p.ActiveConstruction,
		"max_constructions":    p.GetMaxConcurrentBuildings(),
		"pending_buildings":    p.GetPendingBuildings(),
		"storage_capacity":   p.CalculateStorageCapacity(),
		"research_paused":    !p.HasOperationalBase(),
		"research":           researchStates,
		"farm_state":         farmState,
	}
	return result
}

// GetEnergyBalance returns the current energy balance.
func (p *Planet) GetEnergyBalance() float64 {
	return p.EnergyBalance
}

// BuildDetails contains the data needed for the build-details API response.
type BuildDetails struct {
	Resources          PlanetResources
	EnergyBuffer       EnergyBuffer
	Buildings          []BuildingEntry
	EnergyBalance      float64
	ResourceProduction building.ProductionResult
	ActiveConstruction int
	MaxConstruction    int
	BaseOperational    bool
	CanResearch        bool
	CanExpedition      bool
	}

// GetBuildDetails returns a BuildDetails struct for the frontend.
func (p *Planet) GetBuildDetails() BuildDetails {
	var resourceProduction building.ProductionResult
	resourceProduction = p.calculateResourceProduction()

	baseOperational := p.Resources.Food > 0
	expeditionsUnlocked := false
	if p.Research != nil {
		if _, ok := p.Research.GetCompleted()["expeditions"]; ok {
			expeditionsUnlocked = true
		}
	}

	return BuildDetails{
		Resources:          p.Resources,
		EnergyBuffer:       p.EnergyBuffer,
		Buildings:          p.Buildings,
		EnergyBalance:      p.EnergyBalance,
		ResourceProduction: resourceProduction,
		ActiveConstruction: p.ActiveConstruction,
		MaxConstruction:    p.GetMaxConcurrentBuildings(),
		BaseOperational:    baseOperational,
		CanResearch:        baseOperational,
		CanExpedition:      baseOperational && expeditionsUnlocked,
	}
}
