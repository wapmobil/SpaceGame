package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"spacegame/internal/game"
)

func handleGetRatings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			Error(w, http.StatusInternalServerError, "Game not initialized")
			return
		}

		var result *game.RatingsResult
		var err error
		if planetID != "" {
			entry, err := g.GetPlayerRank(category, planetID)
			if err != nil {
				Error(w, http.StatusInternalServerError, "Failed to get player rank")
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
				Error(w, http.StatusInternalServerError, "Failed to get ratings")
				return
			}
		}

		JSON(w, http.StatusOK, result)
	}
}

func handleGetStats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)

		g := game.Instance()
		if g == nil {
			Error(w, http.StatusInternalServerError, "Game not initialized")
			return
		}

		statsTracker := game.NewStatsTracker(g)
		var response map[string]interface{}
		var err error

		if planetID != "" {
			response, err = statsTracker.GetStatsForPlanet(planetID)
		} else {
			playerID := AuthPlayerFromContext(r).ID
			response, err = statsTracker.GetStatsSummary(playerID)
		}

		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to get stats")
			return
		}

		JSON(w, http.StatusOK, response)
	}
}

func handleResolveEvent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)

		var req struct {
			EventType string `json:"event_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.EventType == "" {
			Error(w, http.StatusBadRequest, "Missing event_type")
			return
		}

		g := game.Instance()
		if g == nil {
			Error(w, http.StatusInternalServerError, "Game not initialized")
			return
		}

		message, err := g.ResolveEvent(planetID, req.EventType)
		if err != nil {
			errMsg := err.Error()
			switch {
			case errMsg == "not_found":
				Error(w, http.StatusNotFound, "Event not found")
			case errMsg == "already_resolved":
				Error(w, http.StatusConflict, "Event already resolved")
			case errMsg == "insufficient_resources":
				Error(w, http.StatusConflict, "Insufficient resources to resolve event")
			default:
				Error(w, http.StatusConflict, "Failed to resolve event")
			}
			return
		}

		JSON(w, http.StatusOK, map[string]interface{}{
			"status":  "resolved",
			"message": message,
		})
	}
}
