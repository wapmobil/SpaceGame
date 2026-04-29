package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"spacegame/internal/game"
	"spacegame/internal/game/planet_survey"

	"github.com/go-chi/chi/v5"
)

func handleStartPlanetSurvey(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)

		var req StartPlanetSurveyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Duration <= 0 {
			Error(w, http.StatusBadRequest, "Missing or invalid duration")
			return
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		if !p.CanStartPlanetSurvey() {
			if !p.BaseOperational() {
				Error(w, http.StatusBadRequest, "Planet base requires food to operate")
				return
			}
			if _, ok := p.Research.GetCompleted()["planet_exploration"]; !ok {
				Error(w, http.StatusBadRequest, "Planet exploration research not completed")
				return
			}
			if time.Now().Unix() < p.ExpeditionCooldown {
				Error(w, http.StatusBadRequest, "Survey cooldown active")
				return
			}
			Error(w, http.StatusBadRequest, "Cannot start planet survey")
			return
		}

		maxDuration := getMaxDurationForBaseLevel(p.Level)
		if req.Duration > maxDuration {
			Error(w, http.StatusBadRequest, "Duration exceeds maximum for base level")
			return
		}

		food, iron, money := planet_survey.CalculateCost(p.Level, float64(req.Duration))
		if p.Resources.Food < food || p.Resources.Iron < iron || p.Resources.Money < money {
			Error(w, http.StatusConflict, "Insufficient resources")
			return
		}

		p.Resources.Food -= food
		p.Resources.Iron -= iron
		p.Resources.Money -= money

		rangeStr := getRangeForDuration(req.Duration)
		exp := planet_survey.NewSurfaceExpedition(
			p.ID+"_psurvey_"+time.Now().Format("20060102150405"),
			p.ID,
			rangeStr,
			p.Level,
		)

		p.SurfaceExpeditions = append(p.SurfaceExpeditions, exp)
		p.ExpeditionCooldown = time.Now().Unix() + 30

		game.Instance().SavePlanet(p)

		resp := map[string]interface{}{
			"id":                  exp.ID,
			"planet_id":           exp.PlanetID,
			"status":              exp.Status,
			"progress":            exp.Progress,
			"duration":            exp.Duration,
			"elapsed_time":        exp.ElapsedTime,
			"range":               exp.Range,
			"food_cost":           food,
			"iron_cost":           iron,
			"money_cost":          money,
			"cooldown_end":        p.ExpeditionCooldown,
			"created_at":          exp.CreatedAt.Format(time.RFC3339),
			"updated_at":          exp.UpdatedAt.Format(time.RFC3339),
		}

		JSON(w, http.StatusOK, resp)
	}
}

func handleGetPlanetSurvey(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		expeditions := make([]map[string]interface{}, 0, len(p.SurfaceExpeditions))
		for _, exp := range p.SurfaceExpeditions {
			expState := map[string]interface{}{
				"id":           exp.ID,
				"planet_id":    exp.PlanetID,
				"status":       exp.Status,
				"progress":     exp.Progress,
				"duration":     exp.Duration,
				"elapsed_time": exp.ElapsedTime,
				"range":        exp.Range,
				"created_at":   exp.CreatedAt.Format(time.RFC3339),
				"updated_at":   exp.UpdatedAt.Format(time.RFC3339),
			}
			if exp.Discovered != nil {
				expState["discovered"] = map[string]interface{}{
					"id":              exp.Discovered.ID,
					"type":            exp.Discovered.Type,
					"name":            exp.Discovered.Name,
					"source_resource": exp.Discovered.SourceResource,
					"source_remaining": exp.Discovered.SourceRemaining,
					"building_type":   exp.Discovered.BuildingType,
					"building_level":  exp.Discovered.BuildingLevel,
					"building_active": exp.Discovered.BuildingActive,
				}
			}
			expeditions = append(expeditions, expState)
		}

		rangeStats := make(map[string]ExpeditionRangeStatsResponse)
		for rangeStr, stats := range p.RangeStats {
			rangeStats[rangeStr] = ExpeditionRangeStatsResponse{
				TotalExpeditions: stats.TotalExpeditions,
				LocationsFound:   stats.LocationsFound,
			}
		}

		maxDuration := getMaxDurationForBaseLevel(p.Level)
		costPerMin := planet_survey.GetCostPerMin(p.Level)

		resp := map[string]interface{}{
			"expeditions":      expeditions,
			"range_stats":      rangeStats,
			"max_duration":     maxDuration,
			"cost_per_min":     map[string]float64{"food": costPerMin.Food, "iron": costPerMin.Iron, "money": costPerMin.Money},
			"cooldown_end":     p.ExpeditionCooldown,
			"can_start_survey": p.CanStartPlanetSurvey(),
		}

		JSON(w, http.StatusOK, resp)
	}
}

func handleGetLocations(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		locations := make([]LocationResponse, 0, len(p.Locations))
		for _, loc := range p.Locations {
			resp := LocationResponse{
				ID:              loc.ID,
				Type:            loc.Type,
				Name:            loc.Name,
				BuildingType:    nil,
				BuildingLevel:   loc.BuildingLevel,
				BuildingActive:  loc.BuildingActive,
				SourceResource:  loc.SourceResource,
				SourceAmount:    loc.SourceAmount,
				SourceRemaining: loc.SourceRemaining,
				Active:          loc.Active,
				DiscoveredAt:    loc.DiscoveredAt,
			}
			if loc.BuildingType != "" {
				resp.BuildingType = &loc.BuildingType
			}
			locations = append(locations, resp)
		}

		JSON(w, http.StatusOK, locations)
	}
}

func handleBuildOnLocation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		authPlayer := AuthPlayerFromContext(r)
		ownerID := ""
		if authPlayer != nil {
			ownerID = authPlayer.ID
		}
		locationID := chi.URLParam(r, "id")

		var req BuildOnLocationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.BuildingType == "" {
			Error(w, http.StatusBadRequest, "Missing building_type")
			return
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		loc := findLocation(p, locationID)
		if loc == nil {
			Error(w, http.StatusNotFound, "Location not found")
			return
		}

		if loc.OwnerID != ownerID {
			Error(w, http.StatusForbidden, "Location not owned by player")
			return
		}

		if _, ok := p.Research.GetCompleted()["location_buildings"]; !ok {
			Error(w, http.StatusBadRequest, "Location buildings research not completed")
			return
		}

	if loc.BuildingType != "" {
			Error(w, http.StatusBadRequest, "Building already exists on this location")
			return
		}

		def := planet_survey.GetBuildingDef(req.BuildingType)
		if def == nil {
			Error(w, http.StatusBadRequest, "Unknown building type")
			return
		}

		allowedBuildings := getLocationBuildingTypes(loc.Type)
		validBuilding := false
		for _, bt := range allowedBuildings {
			if bt == req.BuildingType {
				validBuilding = true
				break
			}
		}
		if !validBuilding {
			Error(w, http.StatusBadRequest, "Building type not valid for this location")
			return
		}

		baseFood, baseIron, baseMoney := planet_survey.GetBuildingCostByRarity(planet_survey.GetLocationRarityWeight(loc.Type))
		buildLevelMultiplier := float64(loc.BuildingLevel + 1)
		foodCost := baseFood * buildLevelMultiplier
		ironCost := baseIron * buildLevelMultiplier
		moneyCost := baseMoney * buildLevelMultiplier

		if p.Resources.Food < foodCost || p.Resources.Iron < ironCost || p.Resources.Money < moneyCost {
			Error(w, http.StatusConflict, "Insufficient resources")
			return
		}

		p.Resources.Food -= foodCost
		p.Resources.Iron -= ironCost
		p.Resources.Money -= moneyCost

		buildingID := p.ID + "_lb_" + time.Now().Format("20060102150405")
		lb := &planet_survey.LocationBuilding{
			ID:           buildingID,
			LocationID:   loc.ID,
			BuildingType: req.BuildingType,
			Level:        1,
			Active:       true,
			BuildProgress: 1.0,
			BuildTime:    1.0,
			CostFood:     foodCost,
			CostIron:     ironCost,
			CostMoney:    moneyCost,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		loc.Buildings = append(loc.Buildings, lb)
		loc.BuildingType = req.BuildingType
		loc.BuildingLevel = 1
		loc.BuildingActive = true

		game.Instance().SavePlanet(p)

		wsBroadcast.BroadcastNotification(ownerID, "Building constructed: "+req.BuildingType, "location_update")

		JSON(w, http.StatusOK, map[string]interface{}{
			"id":              lb.ID,
			"location_id":     loc.ID,
			"building_type":   lb.BuildingType,
			"level":           lb.Level,
			"active":          lb.Active,
			"cost_food":       foodCost,
			"cost_iron":       ironCost,
			"cost_money":      moneyCost,
			"created_at":      lb.CreatedAt.Format(time.RFC3339),
			"updated_at":      lb.UpdatedAt.Format(time.RFC3339),
		})
	}
}

func handleRemoveBuilding(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		ownerID := AuthPlayerFromContext(r).ID
		locationID := chi.URLParam(r, "id")

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		loc := findLocation(p, locationID)
		if loc == nil {
			Error(w, http.StatusNotFound, "Location not found")
			return
		}

		if loc.OwnerID != ownerID {
			Error(w, http.StatusForbidden, "Location not owned by player")
			return
		}

		if len(loc.Buildings) == 0 {
			Error(w, http.StatusBadRequest, "No building to remove")
			return
		}

		loc.Buildings = nil
		loc.BuildingType = ""
		loc.BuildingLevel = 0
		loc.BuildingActive = false

		game.Instance().SavePlanet(p)

		wsBroadcast.BroadcastNotification(ownerID, "Building removed from location", "location_update")

		JSON(w, http.StatusOK, map[string]interface{}{
			"status": "removed",
			"location_id": locationID,
		})
	}
}

func handleAbandonLocation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		ownerID := AuthPlayerFromContext(r).ID
		locationID := chi.URLParam(r, "id")

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		loc := findLocation(p, locationID)
		if loc == nil {
			Error(w, http.StatusNotFound, "Location not found")
			return
		}

		if loc.OwnerID != ownerID {
			Error(w, http.StatusForbidden, "Location not owned by player")
			return
		}

		for i, l := range p.Locations {
			if l.ID == locationID {
				p.Locations = append(p.Locations[:i], p.Locations[i+1:]...)
				break
			}
		}

		game.Instance().SavePlanet(p)

		wsBroadcast.BroadcastNotification(ownerID, "Location abandoned", "location_update")

		JSON(w, http.StatusOK, map[string]interface{}{
			"status": "abandoned",
			"location_id": locationID,
		})
	}
}

func handleGetExpeditionHistory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		history := make([]ExpeditionHistoryResponse, 0, len(p.ExpeditionHistory))
		for _, entry := range p.ExpeditionHistory {
			resp := ExpeditionHistoryResponse{
				ID:              entry.ID,
				Status:          entry.Status,
				Result:          entry.Result,
				Discovered:      entry.Discovered,
				ResourcesGained: entry.ResourcesGained,
				CreatedAt:       entry.CreatedAt,
				CompletedAt:     entry.CompletedAt,
			}
			history = append(history, resp)
		}

		JSON(w, http.StatusOK, history)
	}
}

func findLocation(p *game.Planet, locationID string) *planet_survey.Location {
	for _, loc := range p.Locations {
		if loc.ID == locationID {
			return loc
		}
	}
	return nil
}

func getMaxDurationForBaseLevel(baseLevel int) int {
	switch baseLevel {
	case 1:
		return 300
	case 2:
		return 600
	case 3:
		return 1200
	default:
		return 300
	}
}

func getRangeForDuration(duration int) string {
	switch {
	case duration <= 300:
		return "300s"
	case duration <= 600:
		return "600s"
	default:
		return "1200s"
	}
}

func getLocationBuildingTypes(locationType string) []string {
	for _, lt := range planet_survey.GetLocationTypes() {
		if lt.Type == locationType {
			buildings := make([]string, len(lt.Buildings))
			for i, b := range lt.Buildings {
				buildings[i] = b.BuildingType
			}
			return buildings
		}
	}
	return nil
}
