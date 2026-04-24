package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"spacegame/internal/auth"
	"spacegame/internal/game"
	"spacegame/internal/game/building"
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

		planets := make([]PlanetResponse, 0)
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}

		resp := map[string]interface{}{
			"id":          p.ID,
			"player_id":   p.OwnerID,
			"name":        p.Name,
			"level":       p.Level,
			"resources":   p.Resources,
			"buildings":   p.Buildings,
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}

		for i := range p.Buildings {
			p.PopulateBuildingEntry(i)
		}

		details := p.GetBuildDetails()

		buildings := make([]BuildingDetail, len(details.Buildings))
		for i, b := range details.Buildings {
			deltas := building.NextLevelDeltas(b.Type, b.Level)
			buildings[i] = BuildingDetail{
				Type:          b.Type,
				Level:         b.Level,
				BuildProgress: b.BuildProgress,
				Enabled:       b.Enabled,
				BuildTime:     b.BuildTime,
				Cost: CostDetail{
					Food:  b.Cost.Food,
					Iron:  b.Cost.Iron,
					Money: b.Cost.Money,
				},
				NextCost: CostDetail{
					Food:  b.NextCost.Food,
					Iron:  b.NextCost.Iron,
					Money: b.NextCost.Money,
				},
				Production: ProdDetail{
					Food:       b.Production.Food,
					Iron:       b.Production.Iron,
					Composite:  b.Production.Composite,
					Mechanisms: b.Production.Mechanisms,
					Reagents:   b.Production.Reagents,
					Energy:     b.Production.Energy,
					Money:      b.Production.Money,
					AlienTech:  b.Production.AlienTech,
				},
				NextProduction: ProdDetail{
					Food:       b.Production.Food + deltas.Food,
					Iron:       b.Production.Iron + deltas.Iron,
					Composite:  b.Production.Composite + deltas.Composite,
					Mechanisms: b.Production.Mechanisms + deltas.Mechanisms,
					Reagents:   b.Production.Reagents + deltas.Reagents,
					Energy:     b.Production.Energy + deltas.Energy,
					Money:      b.Production.Money + deltas.Money,
					AlienTech:  b.Production.AlienTech + deltas.AlienTech,
				},
				Deltas: ProdDetail{
					Food:       deltas.Food,
					Iron:       deltas.Iron,
					Composite:  deltas.Composite,
					Mechanisms: deltas.Mechanisms,
					Reagents:   deltas.Reagents,
					Energy:     deltas.Energy,
					Money:      deltas.Money,
					AlienTech:  deltas.AlienTech,
				},
			}
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
			"comcenter": true, "market": true, "composite_drone": true, "mechanism_factory": true,
			"reagent_lab": true, "dynamo": true, "mine": true,
		}
		if !validBuildings[req.Type] {
			http.Error(w, "Unknown building type", http.StatusBadRequest)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}

		foodCost, moneyCost, err := p.AddBuilding(req.Type)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]interface{}{
				"error":               err.Error(),
				"active_constructions": p.ActiveConstruction,
				"max_constructions":   p.GetMaxConcurrentBuildings(),
			}
			if pe, ok := err.(*game.PlanetError); ok && pe.Extra != "" {
				resp["extra"] = pe.Extra
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		idx := p.FindBuildingIndex(req.Type)
		if idx >= 0 {
			p.PopulateBuildingEntry(idx)
		}

		level := p.GetBuildingLevel(req.Type)
		wsBroadcast.BroadcastBuildingUpdate(ownerID, planetID, req.Type, level)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":              "started",
			"type":                req.Type,
			"progress":            p.Buildings[idx].BuildProgress,
			"food_cost":           foodCost,
			"money_cost":          moneyCost,
			"active_constructions": p.ActiveConstruction,
			"max_constructions":   p.GetMaxConcurrentBuildings(),
		})
	}
}

func handleConfirmBuilding(db *sql.DB) http.HandlerFunc {
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

		buildingType := chiBuildingTypeParam(r)
		if buildingType == "" {
			http.Error(w, "Missing building type", http.StatusBadRequest)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}

		if p == nil {
			http.Error(w, "Planet not found in game instance", http.StatusNotFound)
			return
		}

		if err := p.ConfirmBuilding(buildingType); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		idx := p.FindBuildingIndex(buildingType)
		if idx >= 0 {
			p.PopulateBuildingEntry(idx)
		}

		level := p.GetBuildingLevel(buildingType)
		wsBroadcast.BroadcastBuildingUpdate(ownerID, planetID, buildingType, level)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "confirmed",
			"type":   buildingType,
		})
	}
}

func handleToggleBuilding(db *sql.DB) http.HandlerFunc {
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

		buildingType := chiBuildingTypeParam(r)
		if buildingType == "" {
			http.Error(w, "Missing building type", http.StatusBadRequest)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}

		if p == nil {
			http.Error(w, "Planet not found in game instance", http.StatusNotFound)
			return
		}

		idx := p.FindBuildingIndex(buildingType)
		if idx < 0 {
			http.Error(w, "Building not found", http.StatusNotFound)
			return
		}

		b := &p.Buildings[idx]
		if b.IsBuilding() || b.IsBuildComplete() {
			http.Error(w, "Building not ready", http.StatusBadRequest)
			return
		}

		b.Enabled = !b.Enabled

		level := p.GetBuildingLevel(buildingType)
		wsBroadcast.BroadcastBuildingUpdate(ownerID, planetID, buildingType, level)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "toggled",
			"type":    buildingType,
			"enabled": b.Enabled,
		})
	}
}

func handleGetBuildDetails(db *sql.DB) http.HandlerFunc {
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}

		for i := range p.Buildings {
			p.PopulateBuildingEntry(i)
		}

		details := p.GetBuildDetails()

		buildings := make([]BuildingDetail, len(details.Buildings))
		for i, b := range details.Buildings {
			deltas := building.NextLevelDeltas(b.Type, b.Level)
			buildings[i] = BuildingDetail{
				Type:          b.Type,
				Level:         b.Level,
				BuildProgress: b.BuildProgress,
				Enabled:       b.Enabled,
				BuildTime:     b.BuildTime,
				Cost: CostDetail{
					Food:  b.Cost.Food,
					Iron:  b.Cost.Iron,
					Money: b.Cost.Money,
				},
				NextCost: CostDetail{
					Food:  b.NextCost.Food,
					Iron:  b.NextCost.Iron,
					Money: b.NextCost.Money,
				},
				Production: ProdDetail{
					Food:       b.Production.Food,
					Iron:       b.Production.Iron,
					Composite:  b.Production.Composite,
					Mechanisms: b.Production.Mechanisms,
					Reagents:   b.Production.Reagents,
					Energy:     b.Production.Energy,
					Money:      b.Production.Money,
					AlienTech:  b.Production.AlienTech,
				},
				NextProduction: ProdDetail{
					Food:       b.Production.Food + deltas.Food,
					Iron:       b.Production.Iron + deltas.Iron,
					Composite:  b.Production.Composite + deltas.Composite,
					Mechanisms: b.Production.Mechanisms + deltas.Mechanisms,
					Reagents:   b.Production.Reagents + deltas.Reagents,
					Energy:     b.Production.Energy + deltas.Energy,
					Money:      b.Production.Money + deltas.Money,
					AlienTech:  b.Production.AlienTech + deltas.AlienTech,
				},
				Deltas: ProdDetail{
					Food:       deltas.Food,
					Iron:       deltas.Iron,
					Composite:  deltas.Composite,
					Mechanisms: deltas.Mechanisms,
					Reagents:   deltas.Reagents,
					Energy:     deltas.Energy,
					Money:      deltas.Money,
					AlienTech:  deltas.AlienTech,
				},
			}
		}

		production := ProdDetail{
			Food:       details.ResourceProduction.Food,
			Iron:       details.ResourceProduction.Iron,
			Composite:  details.ResourceProduction.Composite,
			Mechanisms: details.ResourceProduction.Mechanisms,
			Reagents:   details.ResourceProduction.Reagents,
			Energy:     details.ResourceProduction.Energy,
			Money:      details.ResourceProduction.Money,
			AlienTech:  details.ResourceProduction.AlienTech,
			EnergyNet:  details.EnergyBalance,
		}

		// Calculate costs for buildings not yet built
		buildingCosts := make(map[string]BuildingCostDetail)
		existingTypes := make(map[string]bool)
		for _, b := range buildings {
			existingTypes[b.Type] = true
		}
		for _, bt := range game.BuildingsOrder {
			if !existingTypes[bt] && game.IsBuildingUnlocked(bt, p.Research.GetCompleted(), p.Resources.ResearchUnlocks) {
				cost := p.GetBuildingCost(bt, 0)
				p1 := building.Production(bt, 1)
				e1 := -building.EnergyConsumption(bt, 1)
				deltas := building.NextLevelDeltas(bt, 0)
				nextP := building.Production(bt, 2)
				nextE := -building.EnergyConsumption(bt, 2)
				buildingCosts[bt] = BuildingCostDetail{
					Cost: CostDetail{
						Food:  cost.Food,
						Iron:  cost.Iron,
						Money: cost.Money,
					},
					Production: ProdDetail{
						Food:       p1.Food,
						Iron:       p1.Iron,
						Composite:  p1.Composite,
						Mechanisms: p1.Mechanisms,
						Reagents:   p1.Reagents,
						Energy:     e1,
						Money:      p1.Money,
						AlienTech:  p1.AlienTech,
					},
					NextProduction: ProdDetail{
						Food:       nextP.Food,
						Iron:       nextP.Iron,
						Composite:  nextP.Composite,
						Mechanisms: nextP.Mechanisms,
						Reagents:   nextP.Reagents,
						Energy:     nextE,
						Money:      nextP.Money,
						AlienTech:  nextP.AlienTech,
					},
					Deltas: ProdDetail{
						Food:       deltas.Food,
						Iron:       deltas.Iron,
						Composite:  deltas.Composite,
						Mechanisms: deltas.Mechanisms,
						Reagents:   deltas.Reagents,
						Energy:     deltas.Energy,
						Money:      deltas.Money,
						AlienTech:  deltas.AlienTech,
					},
				}
			}
		}

		resp := BuildDetailsResponse{
			Resources: PlanetResources{
				Food:            details.Resources.Food,
				Iron:            details.Resources.Iron,
				Composite:       details.Resources.Composite,
				Mechanisms:      details.Resources.Mechanisms,
				Reagents:        details.Resources.Reagents,
				Energy:          details.EnergyBuffer.Value,
				MaxEnergy:       details.Resources.MaxEnergy,
				Money:           details.Resources.Money,
				AlienTech:       details.Resources.AlienTech,
				StorageCapacity: p.CalculateStorageCapacity(),
			},
			EnergyBuffer: EnergyBufferDetail{
				Value:   details.EnergyBuffer.Value,
				Max:     details.EnergyBuffer.Max,
				Deficit: details.EnergyBuffer.Deficit,
			},
			Buildings: buildings,
			EnergyBalance: EnergyBalanceDetail{
				Production:  details.ResourceProduction.Energy,
				Consumption: 0,
				Net:         details.EnergyBalance,
			},
			ResourceProduction: production,
			ActiveConstruction: details.ActiveConstruction,
			MaxConstruction:    details.MaxConstruction,
			BaseOperational:    details.BaseOperational,
			CanResearch:        details.CanResearch,
			CanExpedition:      details.CanExpedition,
						BuildingCosts:      buildingCosts,
			ResearchUnlocks:    p.Resources.ResearchUnlocks,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
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
		"research":       json.RawMessage(jsonBytes),
		"available":      json.RawMessage(availableBytes),
		"research_paused": !p.HasOperationalBase(),
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}

		err = p.StartResearch(req.TechID)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "base_not_operational"):
				http.Error(w, "Planet base requires food to operate", http.StatusBadRequest)
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}

		fleet := p.GetFleet()
		shipyard := p.GetShipyard()
		shipyardLevel := p.GetBuildingLevel("shipyard")
		maxSlots := shipyard.MaxSlots(p.GetBuildingLevel("base"))

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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}

		shipyardLevel := p.GetBuildingLevel("shipyard")
		maxSlots := p.Shipyard.MaxSlots(p.GetBuildingLevel("base"))

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
	// Extract first path segment as id (handles nested paths like id/buildings/type/confirm)
	parts := strings.SplitN(rest, "/", 2)
	id := parts[0]
	// Trim known suffixes from the remaining path
	rest = parts[1]
	rest = strings.TrimSuffix(rest, "/research")
	rest = strings.TrimSuffix(rest, "/research/start")
	rest = strings.TrimSuffix(rest, "/fleet")
	rest = strings.TrimSuffix(rest, "/ship/build")
	rest = strings.TrimSuffix(rest, "/ships/available")
	rest = strings.TrimSuffix(rest, "/expeditions")
	rest = strings.TrimSuffix(rest, "/mining")
	rest = strings.TrimSuffix(rest, "/mining/start")
	rest = strings.TrimSuffix(rest, "/market/orders")
	rest = strings.TrimSuffix(rest, "/buildings")
	rest = strings.TrimSuffix(rest, "/confirm")
	rest = strings.TrimSuffix(rest, "/build-details")
	// If there's no remaining path, return just the id
	if parts[1] == "" || rest == "" {
		return id
	}
	return id
}

// chiBuildingTypeParam extracts the building type from the URL path for confirm endpoints.
func chiBuildingTypeParam(r *http.Request) string {
	path := r.URL.Path
	prefix := "/api/planets/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(path, prefix)
	// rest is like {planetId}/buildings/{buildingType}/confirm
	parts := strings.Split(rest, "/")
	if len(parts) < 4 {
		return ""
	}
	// parts[0] = planetId, parts[1] = buildings, parts[2] = buildingType
	return parts[2]
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
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
			case strings.Contains(errMsg, "base_not_operational"):
				http.Error(w, "Planet base requires food to operate", http.StatusBadRequest)
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
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
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}

		if p.EnergyBuffer.Value < game.OrderCreationCost {
			http.Error(w, fmt.Sprintf("Insufficient energy. Need %.0f energy, have %.0f", game.OrderCreationCost, p.EnergyBuffer.Value), http.StatusConflict)
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
		p.EnergyBuffer.Value -= game.OrderCreationCost
		if p.EnergyBuffer.Value < 0 {
			p.EnergyBuffer.Value = 0
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
			p.EnergyBuffer.Value += game.OrderCreationCost
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

func handleSellFood(db *sql.DB) http.HandlerFunc {
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
			Amount float64 `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.Amount <= 0 {
			http.Error(w, "Amount must be positive", http.StatusBadRequest)
			return
		}

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
				log.Printf("Error loading planet from DB: %v", err)
			}
			p = game.Instance().GetPlanet(planetID)
		}
		if p == nil {
			http.Error(w, "Planet not found", http.StatusNotFound)
			return
		}

		if p.Resources.Food < req.Amount {
			http.Error(w, fmt.Sprintf("Insufficient food. Need %.0f, have %.0f", req.Amount, p.Resources.Food), http.StatusConflict)
			return
		}

		// Rate: 10 food = 1 money
		if req.Amount < 10 {
			http.Error(w, "Minimum sell amount is 10 food", http.StatusBadRequest)
			return
		}
		if int(req.Amount)%10 != 0 {
			http.Error(w, "Amount must be a multiple of 10", http.StatusBadRequest)
			return
		}
		moneyEarned := req.Amount / 10.0
		p.Resources.Food -= req.Amount
		p.Resources.Money += moneyEarned

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"food_sold":    req.Amount,
			"money_earned": moneyEarned,
		})
	}
}

// handleStartDrill handles POST /api/planets/{id}/drill/start
func handleStartDrill(db *sql.DB) http.HandlerFunc {
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

		// Check for existing active drill session in memory
		for _, dg := range game.ActiveSessions() {
			s := dg.GetSession()
			if s.PlayerID == playerID && s.PlanetID == planetID {
				if s.Status == "active" {
					http.Error(w, "Already have an active drill session", http.StatusConflict)
					return
				}
				// Clean up old failed/completed sessions
				delete(game.ActiveSessions(), s.SessionID)
			}
		}

		// Get mine building level
		var mineLevel int
		err = db.QueryRow("SELECT level FROM buildings WHERE planet_id = $1 AND type = 'mine'", planetID).Scan(&mineLevel)
		if err != nil {
			mineLevel = 0
		}

		// Check cooldown from players table
		var lastCompleted *time.Time
		err = db.QueryRow("SELECT drill_last_completed FROM players WHERE id = $1", playerID).Scan(&lastCompleted)
		if err == nil && lastCompleted != nil && time.Since(*lastCompleted) < game.GetDrillCooldown() {
			remaining := game.GetDrillCooldown() - time.Since(*lastCompleted)
			http.Error(w, fmt.Sprintf("Drill cooldown active. Try again in %v", remaining.Round(time.Second)), http.StatusConflict)
			return
		}

		// Create drill game
		dg := game.NewDrillGame(planetID, playerID, mineLevel)
		session := dg.GetSession()

		// Wire up broadcast callback for WebSocket updates
		dg.SetBroadcastFn(func(result *game.MoveResult) {
			updateData := map[string]interface{}{
				"session_id": session.SessionID,
				"drill_hp":   result.DrillHP,
				"drill_max_hp": result.DrillMaxHP,
				"depth":      result.Depth,
				"drill_x":    result.DrillX,
				"resources":  result.Resources,
				"total_earned": result.TotalEarned,
				"status":     session.Status,
				"game_ended": result.GameEnded,
			}
			if result.GameEnded {
				updateData["end_reason"] = result.EndReason
			}
			// Build world for broadcast
			if result.World != nil {
				worldResp := make([][]DrillCellResponse, len(result.World))
				for i, row := range result.World {
					worldResp[i] = make([]DrillCellResponse, len(row))
					for j, cell := range row {
						worldResp[i][j] = DrillCellResponse{
							X:              cell.X,
							Y:              cell.Y,
							CellType:       cell.CellType,
							ResourceType:   cell.ResourceType,
							ResourceAmount: cell.ResourceAmount,
							ResourceValue:  cell.ResourceValue,
							Extracted:      cell.Extracted,
						}
					}
				}
				updateData["world"] = worldResp
			}
			wsBroadcast.BroadcastDrillUpdate(playerID, updateData)
		})

		response := DrillStartResponse{
			SessionID:  session.SessionID,
			Seed:       dg.GetSeed(),
			DrillHP:    session.DrillHP,
			DrillMaxHP: session.DrillMaxHP,
			Depth:      session.Depth,
			DrillX:     session.DrillX,
			Status:     session.Status,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// handleDrillCommand handles POST /api/planets/{id}/drill
func handleDrillCommand(db *sql.DB) http.HandlerFunc {
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

		var req DrillCommandRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Find the active drill session
		dg := game.FindActiveSession(planetID, playerID)
		if dg == nil {
			http.Error(w, "No active drill session", http.StatusNotFound)
			return
		}

		// Set the command (applied on next auto-descent tick)
		dg.SetCommand(req.Direction, req.Extract)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DrillCommandResponse{Status: "command_received"})
	}
}

// handleDrillChunk handles GET /api/planets/{id}/drill/chunk
func handleDrillChunk(db *sql.DB) http.HandlerFunc {
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

		// Find the session
		dg := game.FindActiveSession(planetID, playerID)
		if dg == nil {
			http.Error(w, "No active drill session", http.StatusNotFound)
			return
		}

		// Parse chunk parameters
		xStr := r.URL.Query().Get("x")
		yStr := r.URL.Query().Get("y")
		wStr := r.URL.Query().Get("w")
		hStr := r.URL.Query().Get("h")

		var centerX, centerY, width, height int
		fmt.Sscanf(xStr, "%d", &centerX)
		fmt.Sscanf(yStr, "%d", &centerY)
		fmt.Sscanf(wStr, "%d", &width)
		fmt.Sscanf(hStr, "%d", &height)

		if width <= 0 || height <= 0 {
			width = 5
			height = 5
		}

		chunk := dg.GetChunk(centerX, centerY, width, height)

		worldResp := make([][]DrillCellResponse, len(chunk))
		for i, row := range chunk {
			worldResp[i] = make([]DrillCellResponse, len(row))
			for j, cell := range row {
				worldResp[i][j] = DrillCellResponse{
					X:              cell.X,
					Y:              cell.Y,
					CellType:       cell.CellType,
					ResourceType:   cell.ResourceType,
					ResourceAmount: cell.ResourceAmount,
					ResourceValue:  cell.ResourceValue,
					Extracted:      cell.Extracted,
				}
			}
		}

		response := DrillChunkResponse{
			SessionID: dg.GetSession().SessionID,
			Seed:      dg.GetSeed(),
			World:     worldResp,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleCompleteDrill handles POST /api/planets/{id}/drill/complete
func handleCompleteDrill(db *sql.DB) http.HandlerFunc {
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

		// Get active session from memory
		dg := game.FindActiveSession(planetID, playerID)
		if dg == nil {
			http.Error(w, "No active drill session", http.StatusNotFound)
			return
		}

		// Complete the session and get total earned
		totalEarned := dg.Complete()
		sess := dg.GetSession()

		// Add money to planet resources
		if totalEarned > 0 {
			p := game.Instance().GetPlanet(planetID)
			if p != nil {
				p.Resources.Money += totalEarned
				game.Instance().SavePlanet(p)
			}
		}

		// Build response
		worldResp := make([][]DrillCellResponse, len(sess.World))
		for i, row := range sess.World {
			worldResp[i] = make([]DrillCellResponse, len(row))
			for j, cell := range row {
				worldResp[i][j] = DrillCellResponse{
					X:              cell.X,
					Y:              cell.Y,
					CellType:       cell.CellType,
					ResourceType:   cell.ResourceType,
					ResourceAmount: cell.ResourceAmount,
					ResourceValue:  cell.ResourceValue,
					Extracted:      cell.Extracted,
				}
			}
		}

		resourceResp := make([]DrillResourceResponse, len(sess.Resources))
		for i, res := range sess.Resources {
			resourceResp[i] = DrillResourceResponse{
				Type:   res.Type,
				Name:   res.Name,
				Icon:   res.Icon,
				Amount: res.Amount,
				Value:  res.Value,
			}
		}

		response := DrillCompleteResponse{
			SessionID:   sess.SessionID,
			PlanetID:    planetID,
			DrillHP:     sess.DrillHP,
			DrillMaxHP:  sess.DrillMaxHP,
			Depth:       sess.Depth,
			DrillX:      sess.DrillX,
			WorldWidth:  sess.WorldWidth,
			World:       worldResp,
			Resources:   resourceResp,
			Status:      sess.Status,
			TotalEarned: totalEarned,
			CreatedAt:   sess.CreatedAt.Format(time.RFC3339),
			CompletedAt: time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleDestroyDrill handles POST /api/planets/{id}/drill/destroy
// Sets drill HP to 0 to trigger game over naturally
func handleDestroyDrill(db *sql.DB) http.HandlerFunc {
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

		dg := game.FindActiveSession(planetID, playerID)
		if dg == nil {
			http.Error(w, "No active drill session", http.StatusNotFound)
			return
		}

		sess := dg.GetSession()
		if sess.Status != "active" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":     sess.Status,
				"drill_hp":   sess.DrillHP,
				"drill_max_hp": sess.DrillMaxHP,
				"depth":      sess.Depth,
				"drill_x":    sess.DrillX,
				"resources":  make([]DrillResourceResponse, 0),
				"total_earned": sess.TotalEarned,
				"game_ended": sess.Status != "active",
			})
			return
		}

		dg.Destroy()

		totalEarned := dg.GetSession().TotalEarned
		if totalEarned > 0 {
			p := game.Instance().GetPlanet(planetID)
			if p != nil {
				p.Resources.Money += totalEarned
				game.Instance().SavePlanet(p)
			}
		}

		result := dg.GetSession()
		resourceResp := make([]DrillResourceResponse, len(result.Resources))
		for i, res := range result.Resources {
			resourceResp[i] = DrillResourceResponse{
				Type:   res.Type,
				Name:   res.Name,
				Icon:   res.Icon,
				Amount: res.Amount,
				Value:  res.Value,
			}
		}

		worldResp := make([][]DrillCellResponse, len(result.World))
		for i, row := range result.World {
			worldResp[i] = make([]DrillCellResponse, len(row))
			for j, cell := range row {
				worldResp[i][j] = DrillCellResponse{
					X:              cell.X,
					Y:              cell.Y,
					CellType:       cell.CellType,
					ResourceType:   cell.ResourceType,
					ResourceAmount: cell.ResourceAmount,
					ResourceValue:  cell.ResourceValue,
					Extracted:      cell.Extracted,
				}
			}
		}

		response := DrillMoveResponse{
			Success:     true,
			DrillHP:     result.DrillHP,
			DrillMaxHP:  result.DrillMaxHP,
			Depth:       result.Depth,
			DrillX:      result.DrillX,
			World:       worldResp,
			Resources:   resourceResp,
			TotalEarned: result.TotalEarned,
			GameEnded:   true,
			EndReason:   "player_cancelled",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleCleanupDrill handles POST /api/planets/{id}/drill/cleanup
// Removes failed/completed session from memory
func handleCleanupDrill(db *sql.DB) http.HandlerFunc {
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

		dg := game.FindActiveSession(planetID, playerID)
		if dg != nil {
			sess := dg.GetSession()
			if sess.Status == "failed" || sess.Status == "completed" {
				var totalEarned float64
				for _, r := range sess.Resources {
					totalEarned += r.Value
				}
				if totalEarned > 0 {
					p := game.Instance().GetPlanet(planetID)
					if p != nil {
						p.Resources.Money += totalEarned
						game.Instance().SavePlanet(p)
					}
				}
				delete(game.ActiveSessions(), sess.SessionID)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "cleaned",
		})
	}
}
