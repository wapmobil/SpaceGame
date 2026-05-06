package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"spacegame/internal/game"
	"spacegame/internal/game/planet_survey"

	"github.com/go-chi/chi/v5"
)

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
		locationID := chi.URLParam(r, "build")

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

		buildingID := p.ID + "_lb_" + r.Context().Value("time").(string)
		lb := &planet_survey.LocationBuilding{
			ID:            buildingID,
			LocationID:    loc.ID,
			BuildingType:  req.BuildingType,
			Level:         1,
			Active:        true,
			BuildProgress: 1.0,
			BuildTime:     1.0,
			CostFood:      foodCost,
			CostIron:      ironCost,
			CostMoney:     moneyCost,
		}

		loc.Buildings = append(loc.Buildings, lb)
		loc.BuildingType = req.BuildingType
		loc.BuildingLevel = 1
		loc.BuildingActive = true

		game.Instance().SavePlanet(p)

		wsBroadcast.BroadcastNotification(ownerID, "Building constructed: "+req.BuildingType, "location_update")

		JSON(w, http.StatusOK, map[string]interface{}{
			"id":            lb.ID,
			"location_id":   loc.ID,
			"building_type": lb.BuildingType,
			"level":         lb.Level,
			"active":        lb.Active,
			"cost_food":     foodCost,
			"cost_iron":     ironCost,
			"cost_money":    moneyCost,
		})
	}
}

func handleRemoveBuilding(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		ownerID := AuthPlayerFromContext(r).ID
		locationID := chi.URLParam(r, "build")

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
			"status":      "removed",
			"location_id": locationID,
		})
	}
}

func handleAbandonLocation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		ownerID := AuthPlayerFromContext(r).ID
		locationID := chi.URLParam(r, "build")

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
			"status":      "abandoned",
			"location_id": locationID,
		})
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

// --- Expedition chain handlers ---

func handleStartExpedition(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		log.Printf("handleStartExpedition: planetID=%s", planetID)

		var req StartExpeditionChainRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("handleStartExpedition: invalid body: %v", err)
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			log.Printf("handleStartExpedition: planet not found: %s", planetID)
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		log.Printf("handleStartExpedition: planet=%s, baseOperational=%v, hasChain=%v, food=%.0f", p.ID, p.BaseOperational(), p.HasActiveOrGeneratingChain(), p.Resources.Food)

		if !p.BaseOperational() {
			log.Printf("handleStartExpedition: base not operational")
			Error(w, http.StatusBadRequest, "Planet base requires food to operate")
			return
		}

		if _, ok := p.Research.GetCompleted()["planet_exploration"]; !ok {
			log.Printf("handleStartExpedition: planet_exploration research not completed")
			Error(w, http.StatusBadRequest, "Planet exploration research not completed")
			return
		}

		if p.HasActiveOrGeneratingChain() {
			log.Printf("handleStartExpedition: active/generating chain exists, chainCount=%d", len(p.ExpeditionChains))
			Error(w, http.StatusConflict, "An expedition is already active or being generated")
			return
		}

		log.Printf("handleStartExpedition: validating inventory: %v", req.Inventory)
		if err := planet_survey.ValidateInventory(req.Inventory); err != nil {
			log.Printf("handleStartExpedition: validate inventory failed: %v", err)
			Error(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Printf("handleStartExpedition: checking resources, inventory=%v, planetResources: food=%.0f iron=%.0f composite=%.0f mechanisms=%.0f reagents=%.0f", req.Inventory, p.Resources.Food, p.Resources.Iron, p.Resources.Composite, p.Resources.Mechanisms, p.Resources.Reagents)
		if !hasExpeditionResources(p, req.Inventory) {
			log.Printf("handleStartExpedition: insufficient resources")
			Error(w, http.StatusConflict, "Insufficient resources")
			return
		}

		deductExpeditionResources(p, req.Inventory)
		log.Printf("handleStartExpedition: resources deducted")

		chain, err := p.CreateExpeditionChain(req.Inventory)
		if err != nil {
			log.Printf("handleStartExpedition: CreateExpeditionChain failed: %v", err)
			p.ReturnInventoryToPlanetDirect(req.Inventory)
			game.Instance().SavePlanet(p)
			Error(w, http.StatusInternalServerError, "Failed to create expedition: "+err.Error())
			return
		}
		log.Printf("handleStartExpedition: chain created, id=%s", chain.ID)

		p.ExpeditionChains = append(p.ExpeditionChains, chain)

		db := game.Instance().DB()
		log.Printf("handleStartExpedition: got db=%v", db != nil)
		if err := planet_survey.SaveChainToDB(chain, db); err != nil {
			log.Printf("handleStartExpedition: SaveChainToDB failed: %v", err)
			p.ReturnInventoryToPlanetDirect(req.Inventory)
			p.ExpeditionChains = append(p.ExpeditionChains[:len(p.ExpeditionChains)-1])
			game.Instance().SavePlanet(p)
			Error(w, http.StatusInternalServerError, "Failed to save expedition: "+err.Error())
			return
		}

		game.Instance().SavePlanet(p)
		log.Printf("handleStartExpedition: planet saved, responding with chain_id=%s", chain.ID)

		// Start LLM generation asynchronously in a goroutine
		go func() {
			log.Printf("LLM generation goroutine started for chain %s, owner=%s", chain.ID, p.OwnerID)
			defer log.Printf("LLM generation goroutine finished for chain %s", chain.ID)
			event, genErr := p.GenerateExpeditionEvent(chain)

			log.Printf("LLM event generated for chain %s, genErr=%v, event=%v", chain.ID, genErr, event != nil)

			if genErr != nil {
				planet_survey.SaveChainToDB(chain, game.Instance().DB())
				log.Printf("Failed to generate expedition event for chain %s: %v", chain.ID, genErr)

				wsBroadcast.BroadcastExpeditionComplete(p.OwnerID, map[string]interface{}{
					"chain_id": chain.ID,
					"status":   "failed",
					"error":    genErr.Error(),
				})
				return
			}

			// Save chain with updated status
			if err := planet_survey.SaveChainToDB(chain, game.Instance().DB()); err != nil {
				log.Printf("Error saving chain %s: %v", chain.ID, err)
			}

			// Save events to DB
			if err := planet_survey.SaveEventsToDB(chain.ID, planet_survey.GetEventHistory(chain), game.Instance().DB()); err != nil {
				log.Printf("Error saving events for chain %s: %v", chain.ID, err)
			} else {
				log.Printf("Events saved for chain %s, count=%d", chain.ID, len(chain.Events))
			}

			game.Instance().SavePlanet(p)
			log.Printf("Planet saved for chain %s", chain.ID)

			wsBroadcast.BroadcastExpeditionEvent(p.OwnerID, map[string]interface{}{
				"chain_id":    chain.ID,
				"event":       eventToResponse(event),
				"inventory":   chain.Inventory,
				"event_count": chain.EventCount,
			})
			log.Printf("Broadcast expedition_event for chain %s", chain.ID)
		}()

		JSON(w, http.StatusOK, map[string]interface{}{
			"chain_id":  chain.ID,
			"status":    "generating",
			"inventory": chain.Inventory,
		})
	}
}

func handleGetExpeditionChains(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		chains := p.GetExpeditionChains()
		response := make([]ExpeditionChainResponse, 0, len(chains))

		for _, ch := range chains {
			resp := chainToResponse(ch)
			response = append(response, resp)
		}

		JSON(w, http.StatusOK, ExpeditionChainListResponse{Chains: response, Total: len(response)})
	}
}

func handleGetExpeditionEvent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chainID := chi.URLParam(r, "chainID")
		p := ensurePlanetLoaded(PlanetIDFromContext(r))
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		chain := p.GetActiveExpeditionChain(chainID)
		if chain == nil {
			Error(w, http.StatusNotFound, "Chain not found or not active")
			return
		}

		event := planet_survey.GetCurrentEvent(chain)
		if event == nil {
			Error(w, http.StatusNotFound, "No current event")
			return
		}

		JSON(w, http.StatusOK, eventToResponse(event))
	}
}

func handleExpeditionChoice(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chainID := chi.URLParam(r, "chainID")

		var req ExpeditionChoiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		chain := p.GetActiveExpeditionChain(chainID)
		if chain == nil {
			Error(w, http.StatusNotFound, "Chain not found or not active")
			return
		}

		err := p.RecordExpeditionChoice(chainID, req.ChoiceIndex)
		if err != nil {
			Error(w, http.StatusBadRequest, "Failed to record choice: "+err.Error())
			return
		}

		planet_survey.SaveChainToDB(chain, game.Instance().DB())
		game.Instance().SavePlanet(p)

		// Return immediately with "generating" status
		JSON(w, http.StatusOK, ExpeditionChoiceResult{
			Chain:   chainToResponse(chain),
			Inventory: chain.Inventory,
		})

		// Start LLM generation asynchronously in a goroutine
		go func() {
			event, genErr := p.GenerateNextExpeditionEvent(chain)

			if genErr != nil {
				planet_survey.SaveChainToDB(chain, game.Instance().DB())
				log.Printf("Failed to generate next expedition event for chain %s: %v", chain.ID, genErr)

				wsBroadcast.BroadcastExpeditionComplete(p.OwnerID, map[string]interface{}{
					"chain_id": chainID,
					"status":   "failed",
					"error":    genErr.Error(),
				})
				return
			}

			// Save chain with updated status
			if err := planet_survey.SaveChainToDB(chain, game.Instance().DB()); err != nil {
				log.Printf("Error saving chain %s: %v", chain.ID, err)
			}

			// Save events to DB
			if err := planet_survey.SaveEventsToDB(chain.ID, planet_survey.GetEventHistory(chain), game.Instance().DB()); err != nil {
				log.Printf("Error saving events for chain %s: %v", chain.ID, err)
			}

			if event != nil {
				game.Instance().SavePlanet(p)

				wsBroadcast.BroadcastExpeditionEvent(p.OwnerID, map[string]interface{}{
					"chain_id":    chainID,
					"event":       eventToResponse(event),
					"inventory":   chain.Inventory,
					"event_count": chain.EventCount,
				})
			} else {
				// Chain ended
				var locResp *LocationResponse
				if chain.DiscoveredLocation != nil {
					locResp = locationToResponse(chain.DiscoveredLocation)
					p.Locations = append(p.Locations, chain.DiscoveredLocation)
				}

				p.ReturnInventoryToPlanetDirect(chain.Inventory)
				p.ReturnExpeditionInventory(chainID)
				game.Instance().SavePlanet(p)

				wsBroadcast.BroadcastExpeditionComplete(p.OwnerID, map[string]interface{}{
					"chain_id":  chainID,
					"status":    "completed",
					"inventory": chain.Inventory,
					"location":  locResp,
				})
			}
		}()
	}
}

func handleGetExpeditionEvents(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chainID := chi.URLParam(r, "chainID")
		p := ensurePlanetLoaded(PlanetIDFromContext(r))
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		chain := p.GetActiveExpeditionChain(chainID)
		if chain != nil {
			events := planet_survey.GetEventHistory(chain)
			resp := make([]ExpeditionEventResponse, 0, len(events))
			for _, e := range events {
				r := eventToResponse(&e)
				resp = append(resp, *r)
			}
			JSON(w, http.StatusOK, resp)
			return
		}

		events, err := planet_survey.LoadChainEvents(chainID, game.Instance().DB())
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to load events")
			return
		}

		resp := make([]ExpeditionEventResponse, 0, len(events))
		for _, e := range events {
			r := eventToResponse(&e)
			resp = append(resp, *r)
		}
		JSON(w, http.StatusOK, resp)
	}
}

func handleGetExpeditionEventLog(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chainID := chi.URLParam(r, "chainID")

		rows, err := game.Instance().DB().Query(`
			SELECT event_id, description, choices, player_choice,
				rewards_received, created_at
			FROM expedition_events
			WHERE chain_id = $1
			ORDER BY created_at ASC
		`, chainID)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to query event log")
			return
		}
		defer rows.Close()

		entries := make([]ExpeditionEventLogEntry, 0)
		for rows.Next() {
			var entry ExpeditionEventLogEntry
			var choicesJSON, rewardsJSON []byte
			var playerChoice *int

			err := rows.Scan(
				&entry.EventID, &entry.Description,
				&choicesJSON, &playerChoice,
				&rewardsJSON, &entry.CreatedAt,
			)
			if err != nil {
				continue
			}

			entry.RewardsReceived = make(map[string]float64)
			if len(rewardsJSON) > 0 {
				json.Unmarshal(rewardsJSON, &entry.RewardsReceived)
			}

			if playerChoice != nil {
				entry.PlayerChoice = *playerChoice
				var choices []ExpeditionChoiceResp
				if len(choicesJSON) > 0 {
					json.Unmarshal(choicesJSON, &choices)
				}
				if *playerChoice >= 0 && *playerChoice < len(choices) {
					entry.ChoiceLabel = choices[*playerChoice].Label
				}
			}

			entries = append(entries, entry)
		}

		JSON(w, http.StatusOK, entries)
	}
}

// --- Helper functions ---

func hasExpeditionResources(p *game.Planet, inventory map[string]float64) bool {
	for res, amount := range inventory {
		switch res {
		case "food":
			if p.Resources.Food < amount {
				return false
			}
		case "iron":
			if p.Resources.Iron < amount {
				return false
			}
		case "composite":
			if p.Resources.Composite < amount {
				return false
			}
		case "mechanisms":
			if p.Resources.Mechanisms < amount {
				return false
			}
		case "reagents":
			if p.Resources.Reagents < amount {
				return false
			}
		}
	}
	return true
}

func deductExpeditionResources(p *game.Planet, inventory map[string]float64) {
	for res, amount := range inventory {
		switch res {
		case "food":
			p.Resources.Food -= amount
		case "iron":
			p.Resources.Iron -= amount
		case "composite":
			p.Resources.Composite -= amount
		case "mechanisms":
			p.Resources.Mechanisms -= amount
		case "reagents":
			p.Resources.Reagents -= amount
		}
	}
}

func eventToResponse(e *planet_survey.ExpeditionEvent) *ExpeditionEventResponse {
	if e == nil {
		return nil
	}
	choices := make([]ExpeditionChoiceResp, 0, len(e.Choices))
	for _, c := range e.Choices {
		choices = append(choices, ExpeditionChoiceResp{
			Label:       c.Label,
			Description: c.Description,
			Reward:      c.Reward,
			NextEventID: c.NextEventID,
		})
	}
	return &ExpeditionEventResponse{
		EventID:         e.EventID,
		Description:     e.Description,
		ImmediateReward: e.ImmediateReward,
		Choices:         choices,
		IsEnd:           e.IsEnd,
		LocationReward:  e.LocationReward,
	}
}

func chainToResponse(ch *planet_survey.ExpeditionChain) ExpeditionChainResponse {
	resp := ExpeditionChainResponse{
		ID:                ch.ID,
		PlanetID:          ch.PlanetID,
		OwnerID:           ch.OwnerID,
		Status:            ch.Status,
		EventCount:        ch.EventCount,
		CurrentEventIndex: ch.CurrentEventIndex,
		Inventory:         ch.Inventory,
		CreatedAt:         ch.CreatedAt,
		UpdatedAt:         ch.UpdatedAt,
	}
	if ch.DiscoveredLocation != nil {
		resp.DiscoveredLocation = locationToResponse(ch.DiscoveredLocation)
	}
	return resp
}

func locationToResponse(loc *planet_survey.Location) *LocationResponse {
	if loc == nil {
		return nil
	}
	resp := &LocationResponse{
		ID:              loc.ID,
		Type:            loc.Type,
		Name:            loc.Name,
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
	return resp
}
