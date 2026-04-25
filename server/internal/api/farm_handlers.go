package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"spacegame/internal/game"
)

func handleGetFarm(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		farmState, err := game.GetFarmState(p)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to get farm state")
			return
		}

		if farmState == nil {
			JSON(w, http.StatusNotFound, map[string]string{"error": "Farm not built"})
			return
		}

		log.Printf("handleGetFarm: planetID=%s, farmState=%s", planetID, string(farmState))
		w.Header().Set("Content-Type", "application/json")
		w.Write(farmState)
	}
}

func handleFarmAction(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		ownerID := AuthPlayerFromContext(r).ID

		var req FarmActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Action == "" {
			Error(w, http.StatusBadRequest, "Missing action")
			return
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		result, err := game.FarmActionInternal(p, req.Action, req.RowIndex, req.PlantType)
		if err != nil {
			JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if !result.Success {
			JSON(w, http.StatusBadRequest, result)
			return
		}

		wsBroadcast.BroadcastFarmUpdate(ownerID, map[string]interface{}{
			"planet_id": planetID,
			"rows":      result.Rows,
			"last_tick": result.LastTick,
			"food_gain": result.FoodGain,
		})

		JSON(w, http.StatusOK, result)
	}
}
