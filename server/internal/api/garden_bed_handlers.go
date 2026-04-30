package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"spacegame/internal/game"
)

func handleGetGardenBed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		gardenBedState, err := game.GetGardenBedState(p)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to get garden bed state")
			return
		}

		if gardenBedState == nil {
			JSON(w, http.StatusNotFound, map[string]string{"error": "Garden beds not built"})
			return
		}

		log.Printf("handleGetGardenBed: planetID=%s, gardenBedState=%s", planetID, string(gardenBedState))
		w.Header().Set("Content-Type", "application/json")
		w.Write(gardenBedState)
	}
}

func handleGardenBedAction(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		ownerID := AuthPlayerFromContext(r).ID

		var req GardenBedActionRequest
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

		result, err := game.GardenBedActionInternal(p, req.Action, req.RowIndex, req.PlantType)
		if err != nil {
			JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if !result.Success {
			JSON(w, http.StatusBadRequest, result)
			return
		}

wsBroadcast.BroadcastGardenBedUpdate(ownerID, map[string]interface{}{
		"planet_id":    planetID,
		"rows":         result.Rows,
		"food_gain":    result.FoodGain,
		"money_gain":   result.MoneyGain,
		"food_cost":    result.FoodCost,
		"seed_cost":    result.SeedCost,
		"unlock_level": result.UnlockLevel,
		"wither_timer": result.WitherTimer,
	})

		JSON(w, http.StatusOK, result)
	}
}
