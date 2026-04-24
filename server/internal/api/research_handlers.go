package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"spacegame/internal/game/research"
)

func handleGetResearch(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		jsonBytes, err := p.GetResearchJSON()
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to get research state")
			return
		}

		availableBytes, err := p.GetAvailableResearch()
		if err != nil {
			Error(w, http.StatusInternalServerError, "Failed to get available research")
			return
		}

		JSON(w, http.StatusOK, map[string]interface{}{
			"research":        json.RawMessage(jsonBytes),
			"available":       json.RawMessage(availableBytes),
			"research_paused": !p.HasOperationalBase(),
		})
	}
}

func handleStartResearch(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)

		var req StartResearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.TechID == "" {
			Error(w, http.StatusBadRequest, "Missing tech_id")
			return
		}

		tech := research.GetTechByID(req.TechID)
		if tech == nil {
			Error(w, http.StatusBadRequest, "Unknown technology")
			return
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		err := p.StartResearch(req.TechID)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "base_not_operational"):
				Error(w, http.StatusBadRequest, "Planet base requires food to operate")
			case strings.Contains(errMsg, "insufficient_resources"):
				Error(w, http.StatusConflict, "Insufficient resources")
			case strings.Contains(errMsg, "prerequisites_not_met"):
				Error(w, http.StatusConflict, "Prerequisites not met")
			case strings.Contains(errMsg, "already_in_progress"):
				Error(w, http.StatusConflict, "Research already in progress")
			case strings.Contains(errMsg, "already_completed"):
				Error(w, http.StatusConflict, "Research already completed")
			case strings.Contains(errMsg, "max_level"):
				Error(w, http.StatusConflict, "Maximum level reached")
			default:
				Error(w, http.StatusConflict, "Research error")
			}
			return
		}

		Created(w, map[string]string{
			"status":  "started",
			"tech_id": req.TechID,
		})
	}
}
