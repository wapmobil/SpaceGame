package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"spacegame/internal/auth"
	"spacegame/internal/game"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleRegister(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PlayerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		playerID, err := auth.GeneratePlayerID()
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to generate player ID")
			return
		}

		authToken, err := auth.GenerateAuthToken()
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to generate auth token")
			return
		}

		_, err = db.Exec(
			"INSERT INTO players (id, auth_token, name) VALUES ($1, $2, $3)",
			playerID, authToken, req.Name,
		)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to create player")
			return
		}

		Created(w, PlayerResponse{
			ID:        playerID,
			AuthToken: authToken,
			Name:      req.Name,
		})
	}
}

func handleLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PlayerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Name == "" {
			Error(w, http.StatusBadRequest, "Name is required")
			return
		}

		var playerID, authToken string
		err := db.QueryRow("SELECT id, auth_token FROM players WHERE name = $1", req.Name).Scan(&playerID, &authToken)
		if err == sql.ErrNoRows {
			playerID, err = auth.GeneratePlayerID()
			if err != nil {
				Error(w, http.StatusInternalServerError, "Failed to generate player ID")
				return
			}
			authToken, err = auth.GenerateAuthToken()
			if err != nil {
				Error(w, http.StatusInternalServerError, "Failed to generate auth token")
				return
			}
			_, err = db.Exec("INSERT INTO players (id, auth_token, name) VALUES ($1, $2, $3)", playerID, authToken, req.Name)
			if err != nil {
				Error(w, http.StatusInternalServerError, "Failed to create player")
				return
			}
		} else if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to query player")
			return
		}

		JSON(w, http.StatusOK, PlayerResponse{
			ID:        playerID,
			AuthToken: authToken,
			Name:      req.Name,
		})
	}
}

func handleListPlanets(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			Error(w, http.StatusUnauthorized, "Missing auth token")
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			Error(w, http.StatusUnauthorized, "Invalid auth token")
			return
		}

		rows, err := db.Query(
			"SELECT id, player_id, name, level, resources FROM planets WHERE player_id = $1 ORDER BY created_at DESC",
			playerID,
		)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to list planets")
			return
		}
		defer rows.Close()

		planets := make([]PlanetResponse, 0)
		for rows.Next() {
			var p PlanetResponse
			var resourcesJSON []byte
			if err := rows.Scan(&p.ID, &p.PlayerID, &p.Name, &p.Level, &resourcesJSON); err != nil {
				Error(w, http.StatusInternalServerError, "Failed to scan planet")
				return
			}
			json.Unmarshal(resourcesJSON, &p.Resources)
			planets = append(planets, p)
		}

		// Load descriptions for all planets
		var hasDesc bool
		descCheckErr := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'planets' AND column_name = 'description'
			)
		`).Scan(&hasDesc)
		if descCheckErr == nil && hasDesc {
			descRows, err := db.Query("SELECT id, description FROM planets WHERE player_id = $1", playerID)
			if err == nil {
				defer descRows.Close()
				for descRows.Next() {
					var planetID string
					var description string
					if err := descRows.Scan(&planetID, &description); err == nil {
						for i := range planets {
							if planets[i].ID == planetID {
								planets[i].Description = description
								break
							}
						}
					}
				}
			}
		}

		JSON(w, http.StatusOK, planets)
	}
}

func handleCreatePlanet(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("X-Auth-Token")
		if authToken == "" {
			Error(w, http.StatusUnauthorized, "Missing auth token")
			return
		}

		var playerID string
		err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			Error(w, http.StatusUnauthorized, "Invalid auth token")
			return
		}

		var req CreatePlanetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		var planetID string
		err = db.QueryRow(
			`INSERT INTO planets (player_id, name) VALUES ($1, $2) RETURNING id`,
			playerID, req.Name,
		).Scan(&planetID)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to create planet")
			return
		}

		g := game.Instance()
		if g != nil {
			planet := game.NewPlanet(planetID, playerID, req.Name, g)
			g.AddPlanet(planet)

			// Async description generation
			resourceType := string(planet.ResourceType)
			go func(planetID, name, resourceType string) {
				desc := game.GenerateDescription(name, resourceType)
				if desc == "" {
					return
				}
				_, err := db.Exec(`UPDATE planets SET description = $1 WHERE id = $2`, desc, planetID)
				if err != nil {
					log.Printf("Error saving description for planet %s: %v", planetID, err)
					return
				}
				// Update in-memory planet
				if g != nil {
					p := g.GetPlanet(planetID)
					if p != nil {
						p.Description = desc
					}
				}
			}(planetID, req.Name, resourceType)
		}

		Created(w, map[string]string{"id": planetID})
	}
}

func handleGetPlanet(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		p.SortBuildings()

		resp := map[string]interface{}{
			"id":          p.ID,
			"player_id":   p.OwnerID,
			"name":        p.Name,
			"level":       p.Level,
			"description": p.Description,
			"resources":   p.Resources,
			"buildings":   p.Buildings,
		}

		JSON(w, http.StatusOK, resp)
	}
}
