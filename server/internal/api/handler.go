package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
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
		// Try market prefix
		marketPrefix := "/api/market/"
		if strings.HasPrefix(path, marketPrefix) {
			rest := strings.TrimPrefix(path, marketPrefix)
			rest = strings.TrimSuffix(rest, "/orders")
			rest = strings.TrimSuffix(rest, "/match")
			rest = strings.TrimSuffix(rest, "/traders")
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
	rest = strings.TrimSuffix(rest, "/market/orders")
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

// getMarketplace returns the global marketplace instance.
func getMarketplace() *game.Marketplace {
	g := game.Instance()
	if g == nil {
		return nil
	}
	return g.Marketplace
}

func handleCreateMarketOrder(db *sql.DB) http.HandlerFunc {
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

		var req CreateMarketOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Resource == "" {
			http.Error(w, "Missing resource", http.StatusBadRequest)
			return
		}
		if req.OrderType == "" {
			http.Error(w, "Missing order_type", http.StatusBadRequest)
			return
		}
		if req.Amount <= 0 {
			http.Error(w, "Amount must be positive", http.StatusBadRequest)
			return
		}
		if req.Price <= 0 {
			http.Error(w, "Price must be positive", http.StatusBadRequest)
			return
		}

		// Validate order type
		if req.OrderType != "buy" && req.OrderType != "sell" {
			http.Error(w, "Invalid order_type: must be 'buy' or 'sell'", http.StatusBadRequest)
			return
		}

		// Validate resource
		validResources := map[string]bool{
			"food": true, "composite": true, "mechanisms": true, "reagents": true,
		}
		if !validResources[req.Resource] {
			http.Error(w, "Invalid resource", http.StatusBadRequest)
			return
		}

		// Check energy cost
		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		if p.Resources.Energy < game.OrderCreationCost {
			http.Error(w, fmt.Sprintf("Insufficient energy. Need %.0f energy, have %.0f", game.OrderCreationCost, p.Resources.Energy), http.StatusConflict)
			return
		}

		// Check resource availability
		orderType := game.OrderType(req.OrderType)
		if orderType == game.OrderSell {
			switch req.Resource {
			case "food":
				if p.Resources.Food < req.Amount {
					http.Error(w, fmt.Sprintf("Insufficient %s. Need %.0f, have %.0f", req.Resource, req.Amount, p.Resources.Food), http.StatusConflict)
					return
				}
			case "composite":
				if p.Resources.Composite < req.Amount {
					http.Error(w, fmt.Sprintf("Insufficient %s. Need %.0f, have %.0f", req.Resource, req.Amount, p.Resources.Composite), http.StatusConflict)
					return
				}
			case "mechanisms":
				if p.Resources.Mechanisms < req.Amount {
					http.Error(w, fmt.Sprintf("Insufficient %s. Need %.0f, have %.0f", req.Resource, req.Amount, p.Resources.Mechanisms), http.StatusConflict)
					return
				}
			case "reagents":
				if p.Resources.Reagents < req.Amount {
					http.Error(w, fmt.Sprintf("Insufficient %s. Need %.0f, have %.0f", req.Resource, req.Amount, p.Resources.Reagents), http.StatusConflict)
					return
				}
			}
		}

		// Create order
		mp := getMarketplace()
		if mp == nil {
			http.Error(w, "Marketplace not initialized", http.StatusInternalServerError)
			return
		}

		order, err := mp.CreateOrder(planetID, playerID, req.Resource, orderType, req.Amount, req.Price, req.IsPrivate)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "exceeds maximum"):
				http.Error(w, errMsg, http.StatusBadRequest)
			case strings.Contains(errMsg, "price must be"):
				http.Error(w, errMsg, http.StatusBadRequest)
			case strings.Contains(errMsg, "invalid resource"):
				http.Error(w, errMsg, http.StatusBadRequest)
			case strings.Contains(errMsg, "invalid order type"):
				http.Error(w, errMsg, http.StatusBadRequest)
			default:
				http.Error(w, "Failed to create order", http.StatusInternalServerError)
			}
			return
		}

		// Deduct order creation cost
		p.Resources.Energy -= game.OrderCreationCost
		if p.Resources.Energy < 0 {
			p.Resources.Energy = 0
		}

		// Deduct resources for sell orders (reserved)
		if orderType == game.OrderSell {
			switch req.Resource {
			case "food":
				p.Resources.Food -= req.Amount
			case "composite":
				p.Resources.Composite -= req.Amount
			case "mechanisms":
				p.Resources.Mechanisms -= req.Amount
			case "reagents":
				p.Resources.Reagents -= req.Amount
			}
		}

		// Reserve resources for buy orders
		if orderType == game.OrderBuy {
			switch req.Resource {
			case "food":
				p.Resources.Food -= req.Amount * req.Price
			case "composite":
				p.Resources.Composite -= req.Amount * req.Price
			case "mechanisms":
				p.Resources.Mechanisms -= req.Amount * req.Price
			case "reagents":
				p.Resources.Reagents -= req.Amount * req.Price
			}
		}

		resp := MarketOrderResponse{
			ID:               order.ID,
			PlanetID:         order.PlanetID,
			PlayerID:         order.PlayerID,
			Resource:         order.Resource,
			OrderType:        string(order.OrderType),
			Amount:           order.Amount,
			Price:            order.Price,
			IsPrivate:        order.IsPrivate,
			Link:             order.Link,
			Status:           string(order.Status),
			CreatedAt:        order.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        order.UpdatedAt.Format(time.RFC3339),
			ReservedResources: order.ReservedResources,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

func handleGetMyOrders(db *sql.DB) http.HandlerFunc {
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

		mp := getMarketplace()
		if mp == nil {
			http.Error(w, "Marketplace not initialized", http.StatusInternalServerError)
			return
		}

		orders := mp.GetMyOrders(playerID)

		resp := make([]MarketOrderResponse, 0, len(orders))
		for _, order := range orders {
			resp = append(resp, MarketOrderResponse{
				ID:               order.ID,
				PlanetID:         order.PlanetID,
				PlayerID:         order.PlayerID,
				Resource:         order.Resource,
				OrderType:        string(order.OrderType),
				Amount:           order.Amount,
				Price:            order.Price,
				IsPrivate:        order.IsPrivate,
				Link:             order.Link,
				Status:           string(order.Status),
				CreatedAt:        order.CreatedAt.Format(time.RFC3339),
				UpdatedAt:        order.UpdatedAt.Format(time.RFC3339),
				ReservedResources: order.ReservedResources,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func handleGetGlobalMarket(db *sql.DB) http.HandlerFunc {
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

		mp := getMarketplace()
		if mp == nil {
			http.Error(w, "Marketplace not initialized", http.StatusInternalServerError)
			return
		}

		orders := mp.GetVisibleOrders(playerID)

		// Group by resource and type for better display
		buyOrders := make(map[string][]MarketOrderResponse)
		sellOrders := make(map[string][]MarketOrderResponse)

		for _, order := range orders {
			resp := MarketOrderResponse{
				ID:               order.ID,
				PlanetID:         order.PlanetID,
				PlayerID:         order.PlayerID,
				Resource:         order.Resource,
				OrderType:        string(order.OrderType),
				Amount:           order.Amount,
				Price:            order.Price,
				IsPrivate:        order.IsPrivate,
				Link:             order.Link,
				Status:           string(order.Status),
				CreatedAt:        order.CreatedAt.Format(time.RFC3339),
				UpdatedAt:        order.UpdatedAt.Format(time.RFC3339),
				ReservedResources: order.ReservedResources,
			}
			if order.OrderType == game.OrderBuy {
				buyOrders[order.Resource] = append(buyOrders[order.Resource], resp)
			} else {
				sellOrders[order.Resource] = append(sellOrders[order.Resource], resp)
			}
		}

		// Calculate best prices
		bestBuyPrice := 0.0
		bestSellPrice := math.MaxFloat64

		for _, orders := range buyOrders {
			for _, order := range orders {
				if order.Price > bestBuyPrice {
					bestBuyPrice = order.Price
				}
			}
		}
		for _, orders := range sellOrders {
			for _, order := range orders {
				if order.Price < bestSellPrice {
					bestSellPrice = order.Price
				}
			}
		}
		if bestSellPrice == math.MaxFloat64 {
			bestSellPrice = 0
		}

		// Calculate total volume
		totalVolume := 0.0
		for _, orders := range buyOrders {
			for _, order := range orders {
				totalVolume += order.Amount * order.Price
			}
		}
		for _, orders := range sellOrders {
			for _, order := range orders {
				totalVolume += order.Amount * order.Price
			}
		}

		// Count NPC traders
		npcTraderCount := len(mp.GetAllNPCTraders())

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"buy_orders":       buyOrders,
			"sell_orders":      sellOrders,
			"best_buy_price":   bestBuyPrice,
			"best_sell_price":  bestSellPrice,
			"total_volume":     totalVolume,
			"active_orders":    len(orders),
			"npc_trader_count": npcTraderCount,
		})
	}
}

func handleDeleteMarketOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
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

		orderID := chiURLParam(r, "id")
		if orderID == "" {
			http.Error(w, "Missing order id", http.StatusBadRequest)
			return
		}

		mp := getMarketplace()
		if mp == nil {
			http.Error(w, "Marketplace not initialized", http.StatusInternalServerError)
			return
		}

		order := mp.GetOrder(orderID)
		if order == nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		// Check ownership
		if order.PlayerID != playerID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Find the planet to refund resources
		p := game.Instance().GetPlanet(order.PlanetID)
		if p == nil {
			// Try to find any planet owned by this player
			var planetID string
			err = db.QueryRow("SELECT id FROM planets WHERE player_id = $1 LIMIT 1", playerID).Scan(&planetID)
			if err == nil {
				p = game.Instance().GetPlanet(planetID)
				if p == nil {
					p = game.NewPlanet(planetID, playerID, "", game.Instance())
				}
			}
		}

		// Refund reserved resources
		if p != nil {
			for resource, amount := range order.ReservedResources {
				switch resource {
				case "food":
					p.Resources.Food += amount
				case "composite":
					p.Resources.Composite += amount
				case "mechanisms":
					p.Resources.Mechanisms += amount
				case "reagents":
					p.Resources.Reagents += amount
				}
			}

			// Refund energy cost
			p.Resources.Energy += game.OrderCreationCost
		}

		// Delete the order
		err = mp.DeleteOrder(orderID)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "not found"):
				http.Error(w, "Order not found", http.StatusNotFound)
			case strings.Contains(errMsg, "not active"):
				http.Error(w, "Order is not active and cannot be deleted", http.StatusConflict)
			default:
				http.Error(w, "Failed to delete order", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "deleted",
			"order_id": orderID,
			"refunded_energy": game.OrderCreationCost,
		})
	}
}

func handleGetNPCTraders(db *sql.DB) http.HandlerFunc {
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

		_ = playerID // authenticated user

		mp := getMarketplace()
		if mp == nil {
			http.Error(w, "Marketplace not initialized", http.StatusInternalServerError)
			return
		}

		traders := mp.GetAllNPCTraders()

		resp := make([]NPCTraderResponse, 0, len(traders))
		for _, trader := range traders {
			resp = append(resp, NPCTraderResponse{
				ID:        trader.ID,
				Name:      trader.Name,
				PlanetID:  trader.PlanetID,
				OrderID:   trader.OrderID,
				CreatedAt: trader.CreatedAt.Format(time.RFC3339),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func handleMatchOrders(db *sql.DB) http.HandlerFunc {
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

		_ = playerID // authenticated user

		mp := getMarketplace()
		if mp == nil {
			http.Error(w, "Marketplace not initialized", http.StatusInternalServerError)
			return
		}

		result := mp.MatchOrders()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(MarketMatchingResponse{
			MatchedOrders:  result.MatchedOrders,
			ExecutedTrades: result.ExecutedTrades,
			TotalVolume:    result.TotalVolume,
		})
	}
}
