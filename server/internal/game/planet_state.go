package game

import (
	"encoding/json"
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

	// Build available research list for WS updates
	availableBytes, _ := p.GetAvailableResearch()
	var availableList []map[string]interface{}
	if err := json.Unmarshal(availableBytes, &availableList); err != nil {
		availableList = make([]map[string]interface{}, 0)
	}

	// Build completed research map for WS updates (tech_id -> level)
	completedResearch := make(map[string]int)
	for techID, level := range p.Research.GetCompleted() {
		completedResearch[techID] = level
	}

	// Build garden bed state for WS updates
	var gardenBedState map[string]interface{}
	if p.GardenBedState == nil {
		farmLevel := p.GetBuildingLevel("farm")
		if farmLevel > 0 {
			p.GardenBedState = NewGardenBedState(farmLevel)
		}
	}
	if p.GardenBedState != nil && p.GardenBedState.RowCount > 0 {
		gardenBedState = map[string]interface{}{
			"rows":      p.GardenBedState.Rows,
			"row_count": p.GardenBedState.RowCount,
		}
	}

	result := map[string]interface{}{
		"id":                        p.ID,
		"owner_id":                  p.OwnerID,
		"name":                      p.Name,
		"level":                     p.Level,
		"resources":                 p.Resources,
		"buildings":                 p.Buildings,
		"build_speed":               p.BuildSpeed,
		"energy_balance":            p.EnergyBalance,
		"energy_buffer":             p.EnergyBuffer,
		"last_tick":                 p.LastTick.Format(time.RFC3339),
		"fleet":                     p.Fleet.GetShipState(),
		"shipyard_level":            shipyardLevel,
		"shipyard_queue":            p.Shipyard.Queue,
		"shipyard_slots":            p.Fleet.TotalSlots(),
		"shipyard_max":              p.Shipyard.MaxSlots(baseLevel),
		"shipyard_progress":         p.Shipyard.GetQueueProgress(),
		"expeditions":               p.GetExpeditionState(),
		"active_expeditions":        p.GetActiveExpeditionsCount(),
		"max_expeditions":           p.GetMaxExpeditions(),
		"active_constructions":      p.ActiveConstruction,
		"max_constructions":         p.GetMaxConcurrentBuildings(),
		"pending_buildings":         p.GetPendingBuildings(),
		"storage_capacity":          p.CalculateStorageCapacity(),
		"research_paused":           !p.HasOperationalBase(),
		"research":                  researchStates,
		"available_research":        availableList,
		"completed_research":        completedResearch,
		"garden_bed_state":          gardenBedState,
		"resource_type":             string(p.ResourceType),
		"base_level":                p.GetBuildingLevel("base"),
		"command_center_level":      p.GetBuildingLevel("command_center"),
		"surface_expeditions":       p.GetSurfaceExpeditionState(),
		"locations":                 p.GetLocationState(),
		
		"range_stats":               p.GetRangeStatsState(),
		"expedition_history":        p.ExpeditionHistory,
		"max_surface_expeditions":   p.GetMaxSurfaceExpeditions(),
		"can_start_planet_survey":   p.CanStartPlanetSurvey(),
		"can_start_space_expedition": p.CanStartSpaceExpedition(),
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
	BaseOperational          bool
	BaseLevel                int
	CommandCenterLevel       int
	MaxSurfaceExpeditions    int
	CanResearch              bool
	CanExpedition            bool
	CanPlanetSurvey          bool
	PlanetSurveyUnlocked     bool
	}

// GetBuildDetails returns a BuildDetails struct for the frontend.
func (p *Planet) GetBuildDetails() BuildDetails {
	var resourceProduction building.ProductionResult
	resourceProduction = p.calculateResourceProduction()

	baseOperational := p.Resources.Food > 0
	expeditionsUnlocked := false
	if p.Research != nil {
		if _, ok := p.Research.GetCompleted()["space_expeditions"]; ok {
			expeditionsUnlocked = true
		}
	}

	planetSurveyUnlocked := false
	if p.Research != nil {
		if _, ok := p.Research.GetCompleted()["planet_exploration"]; ok {
			planetSurveyUnlocked = true
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
		BaseOperational:          baseOperational,
		BaseLevel:                p.GetBuildingLevel("base"),
		CommandCenterLevel:       p.GetBuildingLevel("command_center"),
		MaxSurfaceExpeditions:    p.GetMaxSurfaceExpeditions(),
		CanResearch:              baseOperational,
		CanExpedition:      baseOperational && expeditionsUnlocked,
		CanPlanetSurvey:      baseOperational && planetSurveyUnlocked,
		PlanetSurveyUnlocked: planetSurveyUnlocked,
	}
}

func (p *Planet) GetSurfaceExpeditionState() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(p.SurfaceExpeditions))
	for _, exp := range p.SurfaceExpeditions {
		state := map[string]interface{}{
			"id":            exp.ID,
			"planet_id":     exp.PlanetID,
			"status":        exp.Status,
			"progress":      exp.Progress,
			"duration":      exp.Duration,
			"elapsed_time":  exp.ElapsedTime,
			"range":         exp.Range,
			"created_at":    exp.CreatedAt.Format(time.RFC3339),
			"updated_at":    exp.UpdatedAt.Format(time.RFC3339),
		}
		if exp.Discovered != nil {
			state["discovered"] = map[string]interface{}{
				"id":             exp.Discovered.ID,
				"type":           exp.Discovered.Type,
				"name":           exp.Discovered.Name,
				"source_resource": exp.Discovered.SourceResource,
				"source_remaining": exp.Discovered.SourceRemaining,
				"building_type":  exp.Discovered.BuildingType,
				"building_level": exp.Discovered.BuildingLevel,
				"building_active": exp.Discovered.BuildingActive,
			}
		}
		result = append(result, state)
	}
	return result
}

func (p *Planet) GetLocationState() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(p.Locations))
	for _, loc := range p.Locations {
		state := map[string]interface{}{
			"id":              loc.ID,
			"type":            loc.Type,
			"name":            loc.Name,
			"active":          loc.Active,
			"building_active": loc.BuildingActive,
			"building_type":   loc.BuildingType,
			"building_level":  loc.BuildingLevel,
			"source_resource": loc.SourceResource,
			"source_amount":   loc.SourceAmount,
			"source_remaining": loc.SourceRemaining,
			"discovered_at":   loc.DiscoveredAt.Format(time.RFC3339),
			"created_at":      loc.CreatedAt.Format(time.RFC3339),
			"updated_at":      loc.UpdatedAt.Format(time.RFC3339),
			"buildings":       p.GetLocationBuildingState(loc.ID),
		}
		result = append(result, state)
	}
	return result
}

func (p *Planet) GetLocationBuildingState(locationID string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, loc := range p.Locations {
		if loc.ID == locationID {
			for _, lb := range loc.Buildings {
				state := map[string]interface{}{
					"id":            lb.ID,
					"building_type": lb.BuildingType,
					"level":         lb.Level,
					"active":        lb.Active,
					"build_progress": lb.BuildProgress,
					"build_time":    lb.BuildTime,
					"cost_food":     lb.CostFood,
					"cost_iron":     lb.CostIron,
					"cost_money":    lb.CostMoney,
					"created_at":    lb.CreatedAt.Format(time.RFC3339),
					"updated_at":    lb.UpdatedAt.Format(time.RFC3339),
				}
				result = append(result, state)
			}
		}
	}
	return result
}

func (p *Planet) GetRangeStatsState() map[string]interface{} {
	result := make(map[string]interface{})
	for rangeStr, stats := range p.RangeStats {
		result[rangeStr] = map[string]interface{}{
			"total_expeditions": stats.TotalExpeditions,
			"locations_found":   stats.LocationsFound,
		}
	}
	return result
}

func (p *Planet) CanStartPlanetSurvey() bool {
	if !p.BaseOperational() {
		return false
	}
	if _, ok := p.Research.GetCompleted()["planet_exploration"]; !ok {
		return false
	}
	activeCount := 0
	for _, exp := range p.SurfaceExpeditions {
		if exp.Status == "active" {
			activeCount++
		}
	}
	if activeCount >= p.MaxExpeditions {
		return false
	}
	return true
}

func (p *Planet) GetMaxSurfaceExpeditions() int {
	max := 1
	if p.Research != nil {
		if lvl, ok := p.Research.GetCompleted()["advanced_exploration"]; ok && lvl > 0 {
			max += lvl
		}
	}
	return max
}

func (p *Planet) CanStartSpaceExpedition() bool {
	if !p.BaseOperational() {
		return false
	}
	if _, ok := p.Research.GetCompleted()["space_expeditions"]; !ok {
		return false
	}
	return true
}
