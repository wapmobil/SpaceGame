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

	"github.com/google/uuid"

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
			"INSERT INTO players (id, auth_token, name) VALUES ($1, $2, $3)",
			playerID, authToken, req.Name,
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

func handleLogin(db *sql.DB) http.HandlerFunc {
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

		if req.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		var playerID, authToken string
		err := db.QueryRow("SELECT id, auth_token FROM players WHERE name = $1", req.Name).Scan(&playerID, &authToken)
		if err == sql.ErrNoRows {
			playerID, err = auth.GeneratePlayerID()
			if err != nil {
				http.Error(w, "Failed to generate player ID", http.StatusInternalServerError)
				return
			}
			authToken, err = auth.GenerateAuthToken()
			if err != nil {
				http.Error(w, "Failed to generate auth token", http.StatusInternalServerError)
				return
			}
			_, err = db.Exec("INSERT INTO players (id, auth_token, name) VALUES ($1, $2, $3)", playerID, authToken, req.Name)
			if err != nil {
				http.Error(w, "Failed to create player", http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			http.Error(w, "Failed to query player", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
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

		// Add planet to game engine
		g := game.Instance()
		if g != nil {
			planet := game.NewPlanet(planetID, playerID, req.Name, g)
			g.AddPlanet(planet)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": planetID})
	}
}

func handleGetPlanet(db *sql.DB) http.HandlerFunc {
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

		resp := map[string]interface{}{
			"id":           p.ID,
			"player_id":    p.OwnerID,
			"name":         p.Name,
			"level":        p.Level,
			"resources":    p.Resources,
			"buildings":    p.Buildings,
			"build_progress": p.BuildProgress,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func handleGetBuildings(db *sql.DB) http.HandlerFunc {
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

		type buildingInfo struct {
			Type         string  `json:"type"`
			Level        int     `json:"level"`
			BuildProgress float64 `json:"build_progress"`
		}

		var buildings []buildingInfo
		for bType, level := range p.Buildings {
			progress := p.BuildProgress[bType]
			buildings = append(buildings, buildingInfo{
				Type:         bType,
				Level:        level,
				BuildProgress: progress,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buildings)
	}
}

func handleBuildBuilding(db *sql.DB) http.HandlerFunc {
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

		var req BuildBuildingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Type == "" {
			http.Error(w, "Missing building type", http.StatusBadRequest)
			return
		}

		validBuildings := map[string]bool{
			"farm": true, "solar": true, "storage": true, "base": true,
			"factory": true, "energy_storage": true, "shipyard": true,
			"comcenter": true, "composite_drone": true, "mechanism_factory": true,
			"reagent_lab": true,
		}
		if !validBuildings[req.Type] {
			http.Error(w, "Unknown building type", http.StatusBadRequest)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			p = game.NewPlanet(planetID, ownerID, "", game.Instance())
		}

		p.AddBuilding(req.Type)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "started",
			"type":     req.Type,
			"progress": p.BuildProgress[req.Type],
		})
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
	rest = strings.TrimSuffix(rest, "/mining")
	rest = strings.TrimSuffix(rest, "/mining/start")
	rest = strings.TrimSuffix(rest, "/market/orders")
	rest = strings.TrimSuffix(rest, "/buildings")
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

// Helper functions for mining API

// monsterDefinitions is a local copy for the API handlers
var monsterDefinitions = map[string]struct {
	Name      string
	Icon      string
	HP        int
	Damage    int
	Reward    int
	SpawnRate int
}{
	"rat":  {"Rat", "🐀", 15, 8, 15, 3},
	"bat":  {"Bat", "🦇", 25, 12, 25, 2},
	"alien": {"Alien", "👽", 50, 20, 50, 1},
}

func runeMazeToStringMaze(maze [][]rune) [][]string {
	result := make([][]string, len(maze))
	for i, row := range maze {
		result[i] = make([]string, len(row))
		for j, cell := range row {
			result[i][j] = string(cell)
		}
	}
	return result
}

func monstersToResponses(monsters []game.Monster) []MonsterResponse {
	resp := make([]MonsterResponse, 0, len(monsters))
	for _, m := range monsters {
		resp = append(resp, MonsterResponse{
			ID:      m.ID,
			Type:    m.Type,
			Name:    m.Name,
			Icon:    m.Icon,
			X:       m.X,
			Y:       m.Y,
			HP:      m.HP,
			MaxHP:   m.MaxHP,
			Damage:  m.Damage,
			Reward:  m.Reward,
			Alive:   m.Alive,
		})
	}
	return resp
}

func movesToStrings(moves []game.MoveDirection) []string {
	result := make([]string, len(moves))
	for i, m := range moves {
		result[i] = game.DirectionToString(m)
	}
	return result
}

// miningGames stores active mining game instances for testing
var miningGames = make(map[string]*game.MiningGame)

func handleStartMining(db *sql.DB) http.HandlerFunc {
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

		// Check for existing active mining session
		var activeCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM mining_sessions 
			WHERE player_id = $1 AND status = 'active'
		`, playerID).Scan(&activeCount)
		if err != nil {
			http.Error(w, "Failed to check active sessions", http.StatusInternalServerError)
			return
		}
		if activeCount > 0 {
			http.Error(w, "Already have an active mining session", http.StatusConflict)
			return
		}

		// Get base level from planet
		var baseLevel int
		err = db.QueryRow("SELECT level FROM planets WHERE id = $1", planetID).Scan(&baseLevel)
		if err != nil {
			http.Error(w, "Failed to get planet level", http.StatusInternalServerError)
			return
		}

		// Check cooldown
		var lastCompleted time.Time
		err = db.QueryRow(`
			SELECT completed_at FROM mining_sessions 
			WHERE player_id = $1 AND status = 'completed' 
			ORDER BY completed_at DESC LIMIT 1
		`, playerID).Scan(&lastCompleted)
		if err == nil && time.Since(lastCompleted) < game.GetMiningCooldown() {
			remaining := game.GetMiningCooldown() - time.Since(lastCompleted)
			http.Error(w, fmt.Sprintf("Mining cooldown active. Try again in %v", remaining.Round(time.Second)), http.StatusConflict)
			return
		}

		// Create mining game
		mg := game.NewMiningGame(planetID, playerID, baseLevel)
		session := mg.GetSession()
		session.ID = uuid.New().String()

		// Save session to database
		mazeJSON, _ := json.Marshal(session.Maze)
		displayMazeJSON, _ := json.Marshal(session.DisplayMaze)

		_, err = db.Exec(`
			INSERT INTO mining_sessions 
			(id, planet_id, player_id, session_id, player_hp, player_max_hp, player_bombs, money_collected, 
			 maze, display_maze, player_x, player_y, exit_x, exit_y, base_level, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW())
		`, session.ID, planetID, playerID, session.SessionID, session.PlayerHP, session.PlayerMaxHP,
			session.PlayerBombs, session.MoneyCollected, mazeJSON, displayMazeJSON,
			session.PlayerX, session.PlayerY, session.ExitX, session.ExitY,
			session.BaseLevel, session.Status)
		if err != nil {
			http.Error(w, "Failed to save mining session", http.StatusInternalServerError)
			return
		}

		// Save entities to database
		for _, m := range mg.GetSession().Monsters {
			_, err = db.Exec(`
				INSERT INTO mining_entities 
				(session_id, entity_type, x, y, hp, damage, reward, alive)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, session.ID, m.Type, m.X, m.Y, m.HP, m.Damage, m.Reward, m.Alive)
			if err != nil {
				http.Error(w, "Failed to save mining entity", http.StatusInternalServerError)
				return
			}
		}

		// Store game instance for move operations
		miningGames[session.SessionID] = mg

		// Build response
		response := MiningStartResponse{
			Status:         "active",
			SessionID:      session.SessionID,
			Maze:           runeMazeToStringMaze(mg.GetDisplayMaze()),
			PlayerX:        session.PlayerX,
			PlayerY:        session.PlayerY,
			PlayerHP:       session.PlayerHP,
			PlayerMaxHP:    session.PlayerMaxHP,
			PlayerBombs:    session.PlayerBombs,
			MoneyCollected: session.MoneyCollected,
			ExitX:          session.ExitX,
			ExitY:          session.ExitY,
			BaseLevel:      session.BaseLevel,
			Monsters:       monstersToResponses(mg.GetSession().Monsters),
			AvailableMoves: movesToStrings(mg.GetAvailableMoves()),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

func handleMiningMove(db *sql.DB) http.HandlerFunc {
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

		// Verify planet ownership
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

		var req MiningMoveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Direction == "" {
			http.Error(w, "Missing direction", http.StatusBadRequest)
			return
		}

		// Find the active mining session for this player and planet
		var sessionID string
		err = db.QueryRow(`
			SELECT session_id FROM mining_sessions 
			WHERE planet_id = $1 AND player_id = $2 AND status = 'active'
			ORDER BY created_at DESC LIMIT 1
		`, planetID, playerID).Scan(&sessionID)
		if err != nil {
			http.Error(w, "No active mining session", http.StatusNotFound)
			return
		}

		mg, exists := miningGames[sessionID]
		if !exists {
			http.Error(w, "Mining game not found", http.StatusNotFound)
			return
		}

		direction, err := game.ParseDirection(req.Direction)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid direction: %s", req.Direction), http.StatusBadRequest)
			return
		}

		result := mg.Move(direction, req.Slide)
		if !result.Success {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": result.Message,
			})
			return
		}

		// Update session in database
		mazeJSON, _ := json.Marshal(result.Maze)
		_, err = db.Exec(`
			UPDATE mining_sessions 
			SET maze = $1, player_x = $2, player_y = $3, player_hp = $4, 
			    money_collected = $5, last_move_time = NOW()
			WHERE session_id = $6
		`, mazeJSON, result.PlayerX, result.PlayerY, result.PlayerHP,
			result.MoneyCollected, sessionID)
		if err != nil {
			http.Error(w, "Failed to update mining session", http.StatusInternalServerError)
			return
		}

		// Update game ended status
		if result.GameEnded {
			_, err = db.Exec(`
				UPDATE mining_sessions 
				SET status = $1, completed_at = NOW()
				WHERE session_id = $2
			`, result.EndReason, sessionID)
			if err != nil {
				http.Error(w, "Failed to update mining session status", http.StatusInternalServerError)
				return
			}
		}

		// Update monster alive status in database
		for _, m := range mg.GetSession().Monsters {
			_, err = db.Exec(`
				UPDATE mining_entities SET alive = $1 WHERE session_id = $2
			`, m.Alive, sessionID)
			if err != nil {
				http.Error(w, "Failed to update mining entity", http.StatusInternalServerError)
				return
			}
		}

		// Build response
		encounterResp := (*EncounterResponse)(nil)
		if result.Encounter != nil {
			encounterResp = &EncounterResponse{
				MonsterID:   result.Encounter.MonsterID,
				MonsterName: result.Encounter.MonsterName,
				MonsterIcon: result.Encounter.MonsterIcon,
				Damage:      result.Encounter.Damage,
				Reward:      result.Encounter.Reward,
				Killed:      result.Encounter.Killed,
			}
		}

		response := MiningMoveResponse{
			Success:        true,
			Message:        result.Message,
			Maze:           runeMazeToStringMaze(result.Maze),
			PlayerX:        result.PlayerX,
			PlayerY:        result.PlayerY,
			PlayerHP:       result.PlayerHP,
			PlayerBombs:    mg.GetSession().PlayerBombs,
			MoneyCollected: result.MoneyCollected,
			Encounter:      encounterResp,
			GameEnded:      result.GameEnded,
			EndReason:      result.EndReason,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func handleGetMining(db *sql.DB) http.HandlerFunc {
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

		// Get active session
		var sessionID string
		var playerHP, playerMaxHP, playerBombs int
		var moneyCollected float64
		var status string
		var playerX, playerY, exitX, exitY, baseLevel int
		var mazeJSON []byte
		var startTime, completedAt time.Time

		err = db.QueryRow(`
			SELECT session_id, player_hp, player_max_hp, player_bombs, money_collected,
			       status, player_x, player_y, exit_x, exit_y, base_level, maze, created_at, completed_at
			FROM mining_sessions 
			WHERE planet_id = $1 AND player_id = $2 
			ORDER BY created_at DESC LIMIT 1
		`, planetID, playerID).Scan(&sessionID, &playerHP, &playerMaxHP, &playerBombs,
			&moneyCollected, &status, &playerX, &playerY, &exitX, &exitY,
			&baseLevel, &mazeJSON, &startTime, &completedAt)

		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "no_session",
			})
			return
		}
		if err != nil {
			http.Error(w, "Failed to get mining session", http.StatusInternalServerError)
			return
		}

		// Try to get live game state if session is active
		var displayMaze [][]string
		var monsters []MonsterResponse
		var availableMoves []string

		if mg, exists := miningGames[sessionID]; exists {
			session := mg.GetSession()
			displayMaze = runeMazeToStringMaze(mg.GetDisplayMaze())
			monsters = monstersToResponses(session.Monsters)
			availableMoves = movesToStrings(mg.GetAvailableMoves())
		} else {
			// Parse maze from database
			var parsedMaze [][]rune
			if err := json.Unmarshal(mazeJSON, &parsedMaze); err == nil {
				displayMaze = runeMazeToStringMaze(parsedMaze)
			}

			// Get monsters from database
			rows, err := db.Query(`
				SELECT id, entity_type, x, y, hp, damage, reward, alive
				FROM mining_entities 
				WHERE session_id = $1 AND entity_type IN ('rat','bat','alien')
			`, sessionID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var m MonsterResponse
					var monsterID string
					var hp, damage, x, y int
					var reward float64
					var alive bool
					if err := rows.Scan(&monsterID, &m.Type, &x, &y, &hp, &damage, &reward, &alive); err == nil {
						if def, ok := monsterDefinitions[m.Type]; ok {
							m.Name = def.Name
							m.Icon = def.Icon
						}
						m.X = x
						m.Y = y
						m.HP = hp
						m.MaxHP = hp
						m.Damage = damage
						m.Reward = reward
						m.Alive = alive
						monsters = append(monsters, m)
					}
				}
			}

			_ = startTime
			_ = completedAt
		}

		response := MiningStateResponse{
			SessionID:      sessionID,
			PlanetID:       planetID,
			PlayerID:       playerID,
			Maze:           displayMaze,
			PlayerX:        playerX,
			PlayerY:        playerY,
			PlayerHP:       playerHP,
			PlayerMaxHP:    playerMaxHP,
			PlayerBombs:    playerBombs,
			MoneyCollected: moneyCollected,
			Status:         status,
			ExitX:          exitX,
			ExitY:          exitY,
			BaseLevel:      baseLevel,
			Monsters:       monsters,
			AvailableMoves: availableMoves,
			StartTime:      startTime.Format(time.RFC3339),
			CompletedAt:    completedAt.Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func handleGetRatings(db *sql.DB) http.HandlerFunc {
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

		category := r.URL.Query().Get("category")
		planetID := r.URL.Query().Get("planet_id")
		limitStr := r.URL.Query().Get("limit")

		limit := 100
		if limitStr != "" {
			fmt.Sscanf(limitStr, "%d", &limit)
			if limit <= 0 || limit > 1000 {
				limit = 100
			}
		}

		g := game.Instance()
		if g == nil {
			http.Error(w, "Game not initialized", http.StatusInternalServerError)
			return
		}

		var result *game.RatingsResult
		if planetID != "" {
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

			entry, err := g.GetPlayerRank(category, planetID)
			if err != nil {
				http.Error(w, "Failed to get player rank", http.StatusInternalServerError)
				return
			}

			result = &game.RatingsResult{
				Category: category,
				Entries:  []game.RatingEntry{*entry},
				Total:    1,
			}
		} else {
			result, err = g.GetRatings(category, limit, "")
			if err != nil {
				http.Error(w, "Failed to get ratings", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func handleGetStats(db *sql.DB) http.HandlerFunc {
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
		if planetID != "" {
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
		}

		g := game.Instance()
		if g == nil {
			http.Error(w, "Game not initialized", http.StatusInternalServerError)
			return
		}

		statsTracker := game.NewStatsTracker(g)
		var response map[string]interface{}

		if planetID != "" {
			response, err = statsTracker.GetStatsForPlanet(planetID)
		} else {
			response, err = statsTracker.GetStatsSummary(playerID)
		}

		if err != nil {
			http.Error(w, "Failed to get stats", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func handleGetEventHistory(db *sql.DB) http.HandlerFunc {
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

		limitStr := r.URL.Query().Get("limit")
		limit := 50
		if limitStr != "" {
			fmt.Sscanf(limitStr, "%d", &limit)
			if limit <= 0 || limit > 100 {
				limit = 50
			}
		}

		g := game.Instance()
		if g == nil {
			http.Error(w, "Game not initialized", http.StatusInternalServerError)
			return
		}

		events, err := g.GetEventHistory(playerID, limit)
		if err != nil {
			http.Error(w, "Failed to get event history", http.StatusInternalServerError)
			return
		}

		if events == nil {
			events = []map[string]interface{}{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"events": events,
			"total":  len(events),
		})
	}
}

func handleResolveEvent(db *sql.DB) http.HandlerFunc {
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

		var req struct {
			EventType string `json:"event_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.EventType == "" {
			http.Error(w, "Missing event_type", http.StatusBadRequest)
			return
		}

		g := game.Instance()
		if g == nil {
			http.Error(w, "Game not initialized", http.StatusInternalServerError)
			return
		}

		message, err := g.ResolveEvent(planetID, req.EventType)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "planet_not_found"):
				http.Error(w, "Planet not found", http.StatusNotFound)
			case strings.Contains(errMsg, "unknown_event_type"):
				http.Error(w, "Unknown event type", http.StatusBadRequest)
			default:
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": errMsg,
				})
				return
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": message,
		})
	}
}
