package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"spacegame/internal/auth"
	"spacegame/internal/game"
	"spacegame/internal/game/expedition"
	"spacegame/internal/game/research"
	"spacegame/internal/game/ship"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleRegister(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req PlayerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		playerID, err := auth.GeneratePlayerID()
		if err != nil {
			http.Error(w, "Failed to generate player ID", http.StatusInternalServerError)
			return
		}

		authToken, err := auth.GenerateAuthToken()
		if err != nil {
			http.Error(w, "Failed to generate auth token", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(
			"INSERT INTO players (id, auth_token) VALUES ($1, $2)",
			playerID, authToken,
		)
		if err != nil {
			http.Error(w, "Failed to create player", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(PlayerResponse{
			ID:        playerID,
			AuthToken: authToken,
			Name:      req.Name,
		})
	}
}

func handleListPlanets(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		rows, err := db.Query(
			"SELECT id, player_id, name, level, resources FROM planets WHERE player_id = $1 ORDER BY created_at DESC",
			playerID,
		)
		if err != nil {
			http.Error(w, "Failed to list planets", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var planets []PlanetResponse
		for rows.Next() {
			var p PlanetResponse
			var resourcesJSON []byte
			if err := rows.Scan(&p.ID, &p.PlayerID, &p.Name, &p.Level, &resourcesJSON); err != nil {
				http.Error(w, "Failed to scan planet", http.StatusInternalServerError)
				return
			}
			json.Unmarshal(resourcesJSON, &p.Resources)
			planets = append(planets, p)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(planets)
	}
}

func handleCreatePlanet(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		var req CreatePlanetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var planetID string
		err = db.QueryRow(
			`INSERT INTO planets (player_id, name) VALUES ($1, $2) RETURNING id`,
			playerID, req.Name,
		).Scan(&planetID)
		if err != nil {
			http.Error(w, "Failed to create planet", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": planetID})
	}
}

func handleGetResearch(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		planetID := chiURLParam(r, "id")
		if planetID == "" {
			http.Error(w, "Missing planet id", http.StatusBadRequest)
			return
		}

		// Verify planet belongs to player
		var ownerID string
		err = db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
		if err != nil {
			http.Error(w, "Planet not found", http.StatusNotFound)
			return
		}
		if ownerID != playerID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Get planet from game engine
		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			// Planet not loaded yet, load from DB
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		jsonBytes, err := p.GetResearchJSON()
		if err != nil {
			http.Error(w, "Failed to get research state", http.StatusInternalServerError)
			return
		}

		availableBytes, err := p.GetAvailableResearch()
		if err != nil {
			http.Error(w, "Failed to get available research", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"research":  json.RawMessage(jsonBytes),
			"available": json.RawMessage(availableBytes),
		})
	}
}

func handleStartResearch(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		planetID := chiURLParam(r, "id")
		if planetID == "" {
			http.Error(w, "Missing planet id", http.StatusBadRequest)
			return
		}

		// Verify planet belongs to player
		var ownerID string
		err = db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
		if err != nil {
			http.Error(w, "Planet not found", http.StatusNotFound)
			return
		}
		if ownerID != playerID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var req StartResearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.TechID == "" {
			http.Error(w, "Missing tech_id", http.StatusBadRequest)
			return
		}

		tech := research.GetTechByID(req.TechID)
		if tech == nil {
			http.Error(w, "Unknown technology", http.StatusBadRequest)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		err = p.StartResearch(req.TechID)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "insufficient_resources"):
				http.Error(w, "Insufficient resources", http.StatusConflict)
			case strings.Contains(errMsg, "prerequisites_not_met"):
				http.Error(w, "Prerequisites not met", http.StatusConflict)
			case strings.Contains(errMsg, "already_in_progress"):
				http.Error(w, "Research already in progress", http.StatusConflict)
			case strings.Contains(errMsg, "already_completed"):
				http.Error(w, "Research already completed", http.StatusConflict)
			case strings.Contains(errMsg, "max_level"):
				http.Error(w, "Maximum level reached", http.StatusConflict)
			default:
				http.Error(w, "Research error", http.StatusConflict)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "started",
			"tech_id": req.TechID,
		})
	}
}

func handleGetFleet(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		planetID := chiURLParam(r, "id")
		if planetID == "" {
			http.Error(w, "Missing planet id", http.StatusBadRequest)
			return
		}

		var ownerID string
		err = db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
		if err != nil {
			http.Error(w, "Planet not found", http.StatusNotFound)
			return
		}
		if ownerID != playerID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		fleet := p.GetFleet()
		shipyard := p.GetShipyard()
		shipyardLevel := p.Buildings["shipyard"]
		maxSlots := shipyard.MaxSlots(p.Buildings["base"])

		resp := FleetResponse{
			Ships:            fleet.GetShipState(),
			TotalShips:       fleet.TotalShipCount(),
			TotalSlots:       fleet.TotalSlots(),
			MaxSlots:         maxSlots,
			TotalCargo:       fleet.TotalCargoCapacity(),
			TotalEnergy:      fleet.TotalEnergyConsumption(),
			TotalDamage:      fleet.TotalDamage(),
			TotalHP:          fleet.TotalHP(),
			ShipyardLevel:    shipyardLevel,
			ShipyardQueueLen: shipyard.GetQueuedCount(),
			ShipyardProgress: shipyard.GetQueueProgress(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func handleBuildShip(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		planetID := chiURLParam(r, "id")
		if planetID == "" {
			http.Error(w, "Missing planet id", http.StatusBadRequest)
			return
		}

		var ownerID string
		err = db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
		if err != nil {
			http.Error(w, "Planet not found", http.StatusNotFound)
			return
		}
		if ownerID != playerID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var req BuildShipRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.ShipType == "" {
			http.Error(w, "Missing ship_type", http.StatusBadRequest)
			return
		}

		st := ship.GetShipType(ship.TypeID(req.ShipType))
		if st == nil {
			http.Error(w, "Unknown ship type", http.StatusBadRequest)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		err = p.BuildShip(st.TypeID)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "unknown_ship_type"):
				http.Error(w, "Unknown ship type", http.StatusBadRequest)
			case strings.Contains(errMsg, "cannot_build"):
				http.Error(w, "Cannot build ship - check resources, shipyard level, and available slots", http.StatusConflict)
			default:
				http.Error(w, "Build failed", http.StatusConflict)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "queued",
			"ship_id": string(st.TypeID),
		})
	}
}

func handleGetAvailableShips(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		planetID := chiURLParam(r, "id")
		if planetID == "" {
			http.Error(w, "Missing planet id", http.StatusBadRequest)
			return
		}

		var ownerID string
		err = db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
		if err != nil {
			http.Error(w, "Planet not found", http.StatusNotFound)
			return
		}
		if ownerID != playerID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		shipyardLevel := p.Buildings["shipyard"]
		maxSlots := p.Shipyard.MaxSlots(p.Buildings["base"])

		allTypes := ship.AllShipTypes()
		available := make([]ShipTypeResponse, 0, len(allTypes))

		for _, st := range allTypes {
			canBuild := st.MinShipyard <= shipyardLevel &&
				st.Cost.CanAfford(p.Resources.Food, p.Resources.Composite, p.Resources.Mechanisms, p.Resources.Reagents, p.Resources.Money) &&
				p.Fleet.CanAddShip(st, 1, maxSlots)

			available = append(available, ShipTypeResponse{
				TypeID:       string(st.TypeID),
				Name:         st.Name,
				Description:  st.Description,
				Slots:        st.Slots,
				Cargo:        st.Cargo,
				Energy:       st.Energy,
				HP:           st.HP,
				Armor:        st.Armor,
				WeaponMinDmg: st.WeaponMinDmg,
				WeaponMaxDmg: st.WeaponMaxDmg,
				Cost: Cost{
					Food:       st.Cost.Food,
					Composite:  st.Cost.Composite,
					Mechanisms: st.Cost.Mechanisms,
					Reagents:   st.Cost.Reagents,
					Money:      st.Cost.Money,
				},
				BuildTime:     st.BuildTime,
				MinShipyard:   st.MinShipyard,
				CanBuild:      canBuild,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ship_types": available,
			"shipyard_level": shipyardLevel,
			"available_slots": maxSlots - p.Fleet.TotalSlots(),
		})
	}
}

// chiURLParam extracts URL parameters from chi router requests.
func chiURLParam(r *http.Request, key string) string {
	path := r.URL.Path
	prefix := "/api/planets/"
	if !strings.HasPrefix(path, prefix) {
		// Try expedition prefix
		expPrefix := "/api/expeditions/"
		if strings.HasPrefix(path, expPrefix) {
			rest := strings.TrimPrefix(path, expPrefix)
			rest = strings.TrimSuffix(rest, "/action")
			return rest
		}
		return ""
	}
	rest := strings.TrimPrefix(path, prefix)
	rest = strings.TrimSuffix(rest, "/research")
	rest = strings.TrimSuffix(rest, "/research/start")
	rest = strings.TrimSuffix(rest, "/fleet")
	rest = strings.TrimSuffix(rest, "/ship/build")
	rest = strings.TrimSuffix(rest, "/ships/available")
	rest = strings.TrimSuffix(rest, "/battles")
	rest = strings.TrimSuffix(rest, "/expeditions")
	return rest
}

func handleGetBattles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		planetID := chiURLParam(r, "id")
		if planetID == "" {
			http.Error(w, "Missing planet id", http.StatusBadRequest)
			return
		}

		var ownerID string
		err = db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
		if err != nil {
			http.Error(w, "Planet not found", http.StatusNotFound)
			return
		}
		if ownerID != playerID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		battles := p.GetBattleHistory()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"battles": battles,
			"total":   len(battles),
		})
	}
}

func handleCreateExpedition(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		planetID := chiURLParam(r, "id")
		if planetID == "" {
			http.Error(w, "Missing planet id", http.StatusBadRequest)
			return
		}

		var ownerID string
		err = db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
		if err != nil {
			http.Error(w, "Planet not found", http.StatusNotFound)
			return
		}
		if ownerID != playerID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var req StartExpeditionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.ExpeditionType == "" {
			http.Error(w, "Missing expedition_type", http.StatusBadRequest)
			return
		}

		// Validate expedition type
		var expType expedition.Type
		switch req.ExpeditionType {
		case "exploration":
			expType = expedition.TypeExploration
		case "trade":
			expType = expedition.TypeTrade
		case "support":
			expType = expedition.TypeSupport
		default:
			http.Error(w, "Invalid expedition_type", http.StatusBadRequest)
			return
		}

		// Validate duration
		duration := req.Duration
		if duration <= 0 {
			duration = 3600 // default 1 hour
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		// Build expedition fleet from requested ship types
		expFleet := ship.NewFleet()
		for i, shipType := range req.ShipTypes {
			count := 0
			if i < len(req.ShipCounts) {
				count = req.ShipCounts[i]
			}
			if count <= 0 {
				continue
			}
			st := ship.GetShipType(ship.TypeID(shipType))
			if st != nil {
				expFleet.AddShip(st, count)
			}
		}

		if expFleet.TotalShipCount() == 0 {
			// Use entire fleet
			expFleet = p.GetFleet()
		}

		target := req.Target
		if target == "" {
			target = string(expType)
		}

		exp, err := p.StartExpedition(expType, expFleet, target, duration)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "expeditions_not_researched"):
				http.Error(w, "Expeditions not researched yet", http.StatusConflict)
			case strings.Contains(errMsg, "max_expeditions_reached"):
				http.Error(w, "Maximum concurrent expeditions reached", http.StatusConflict)
			case strings.Contains(errMsg, "no_ships_available"):
				http.Error(w, "No ships available for expedition", http.StatusBadRequest)
			case strings.Contains(errMsg, "insufficient_energy"):
				http.Error(w, "Insufficient energy for expedition", http.StatusConflict)
			default:
				http.Error(w, "Failed to start expedition", http.StatusConflict)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":          "started",
			"expedition_id":   exp.ID,
			"expedition_type": exp.ExpeditionType,
			"duration":        exp.Duration,
			"fleet_size":      exp.Fleet.TotalShipCount(),
		})
	}
}

func handleGetExpeditions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		planetID := chiURLParam(r, "id")
		if planetID == "" {
			http.Error(w, "Missing planet id", http.StatusBadRequest)
			return
		}

		var ownerID string
		err = db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
		if err != nil {
			http.Error(w, "Planet not found", http.StatusNotFound)
			return
		}
		if ownerID != playerID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		expeditions := p.GetExpeditions()
		expeditionsUnlocked := false
		if _, ok := p.GetResearchCompleted()["expeditions"]; ok {
			expeditionsUnlocked = true
		}

		resp := ExpeditionsListResponse{
			Expeditions:         make([]ExpeditionResponse, 0, len(expeditions)),
			ActiveCount:         p.GetActiveExpeditionsCount(),
			MaxExpeditions:      p.GetMaxExpeditions(),
			ExpeditionsUnlocked: expeditionsUnlocked,
		}

		resp.CanStartNew = !expeditionsUnlocked || resp.ActiveCount < resp.MaxExpeditions

		for _, exp := range expeditions {
			npcResp := (*NPCPlanetResponse)(nil)
			if exp.DiscoveredNPC != nil {
				npc := exp.DiscoveredNPC
				npcResp = &NPCPlanetResponse{
					ID:             npc.ID,
					Name:           npc.Name,
					Type:           string(npc.Type),
					Resources:      npc.Resources,
					TotalResources: npc.TotalResources(),
					HasCombat:      npc.HasCombatShips(),
					FleetStrength:  npc.TotalFleetStrength(),
				}
				if npc.EnemyFleet != nil {
					npcResp.EnemyFleet = npc.EnemyFleet.GetShipState()
				}
			}

			actionResp := make([]ExpeditionActionResp, 0, len(exp.Actions))
			for _, a := range exp.Actions {
				actionResp = append(actionResp, ExpeditionActionResp{
					ID:       a.ID,
					Type:     a.Type,
					Label:    a.Label,
					Required: a.Required,
				})
			}

			resp.Expeditions = append(resp.Expeditions, ExpeditionResponse{
				ID:             exp.ID,
				PlanetID:       exp.PlanetID,
				Target:         exp.Target,
				Progress:       exp.Progress,
				Status:         string(exp.Status),
				ExpeditionType: string(exp.ExpeditionType),
				Duration:       exp.Duration,
				ElapsedTime:    exp.ElapsedTime,
				FleetShips:     exp.Fleet.GetShipState(),
				FleetTotal:     exp.Fleet.TotalShipCount(),
				FleetCargo:     exp.Fleet.TotalCargoCapacity(),
				FleetEnergy:    exp.Fleet.TotalEnergyConsumption(),
				FleetDamage:    exp.Fleet.TotalDamage(),
				DiscoveredNPC:  npcResp,
				Actions:        actionResp,
				CreatedAt:      exp.CreatedAt.Format(time.RFC3339),
				UpdatedAt:      exp.UpdatedAt.Format(time.RFC3339),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func handleExpeditionAction(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		expeditionID := chiURLParam(r, "id")
		if expeditionID == "" {
			http.Error(w, "Missing expedition id", http.StatusBadRequest)
			return
		}

		var req ExpeditionActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Action == "" {
			http.Error(w, "Missing action", http.StatusBadRequest)
			return
		}

		// Find the planet that owns this expedition
		var planetID string
		err = db.QueryRow(`
			SELECT planet_id FROM expeditions WHERE id = $1
		`, expeditionID).Scan(&planetID)
		if err != nil {
			http.Error(w, "Expedition not found", http.StatusNotFound)
			return
		}

		var ownerID string
		err = db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
		if err != nil {
			http.Error(w, "Planet not found", http.StatusNotFound)
			return
		}
		if ownerID != playerID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		err = p.DoExpeditionAction(expeditionID, req.Action)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "expedition_not_found"):
				http.Error(w, "Expedition not found", http.StatusNotFound)
			case strings.Contains(errMsg, "expedition_not_at_point"):
				http.Error(w, "Expedition not at a point of interest", http.StatusConflict)
			case strings.Contains(errMsg, "no_npc_discovered"):
				http.Error(w, "No NPC discovered yet", http.StatusConflict)
			case strings.Contains(errMsg, "no_combat_ships"):
				http.Error(w, "No combat ships in expedition fleet", http.StatusConflict)
			case strings.Contains(errMsg, "unknown_action"):
				http.Error(w, "Unknown action type", http.StatusBadRequest)
			default:
				http.Error(w, "Action failed", http.StatusConflict)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "action_completed",
			"action":  req.Action,
			"expedition_id": expeditionID,
		})
	}
}
