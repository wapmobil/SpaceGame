package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"spacegame/internal/game/expedition"
	"spacegame/internal/game/ship"

	"github.com/go-chi/chi/v5"
)

func handleCreateExpedition(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)

		var req StartExpeditionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.ExpeditionType == "" {
			Error(w, http.StatusBadRequest, "Missing expedition_type")
			return
		}

		var expType expedition.Type
		switch req.ExpeditionType {
		case "exploration":
			expType = expedition.TypeExploration
		case "trade":
			expType = expedition.TypeTrade
		case "support":
			expType = expedition.TypeSupport
		default:
			Error(w, http.StatusBadRequest, "Invalid expedition_type")
			return
		}

		duration := req.Duration
		if duration <= 0 {
			duration = 3600
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

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
				Error(w, http.StatusBadRequest, "Planet base requires food to operate")
			case strings.Contains(errMsg, "expeditions_not_researched"):
				Error(w, http.StatusConflict, "Expeditions not researched yet")
			case strings.Contains(errMsg, "max_expeditions_reached"):
				Error(w, http.StatusConflict, "Maximum concurrent expeditions reached")
			case strings.Contains(errMsg, "no_ships_available"):
				Error(w, http.StatusBadRequest, "No ships available for expedition")
			case strings.Contains(errMsg, "insufficient_energy"):
				Error(w, http.StatusConflict, "Insufficient energy for expedition")
			default:
				Error(w, http.StatusConflict, "Failed to start expedition")
			}
			return
		}

		Created(w, map[string]interface{}{
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
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
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

		JSON(w, http.StatusOK, resp)
	}
}

func handleExpeditionAction(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID := AuthPlayerFromContext(r).ID
		expeditionID := chi.URLParam(r, "id")

		var req ExpeditionActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Action == "" {
			Error(w, http.StatusBadRequest, "Missing action")
			return
		}

		var planetID string
		err := db.QueryRow(`
			SELECT planet_id FROM expeditions WHERE id = $1
		`, expeditionID).Scan(&planetID)
		if err != nil {
			Error(w, http.StatusNotFound, "Expedition not found")
			return
		}

		var ownerID string
		err = db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
		if err != nil {
			Error(w, http.StatusNotFound, "Planet not found")
			return
		}
		if ownerID != playerID {
			Error(w, http.StatusForbidden, "Forbidden")
			return
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		err = p.DoExpeditionAction(expeditionID, req.Action)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "expedition_not_found"):
				Error(w, http.StatusNotFound, "Expedition not found")
			case strings.Contains(errMsg, "expedition_not_at_point"):
				Error(w, http.StatusConflict, "Expedition not at a point of interest")
			case strings.Contains(errMsg, "no_npc_discovered"):
				Error(w, http.StatusConflict, "No NPC discovered yet")
			case strings.Contains(errMsg, "no_combat_ships"):
				Error(w, http.StatusConflict, "No combat ships in expedition fleet")
			case strings.Contains(errMsg, "unknown_action"):
				Error(w, http.StatusBadRequest, "Unknown action type")
			default:
				Error(w, http.StatusConflict, "Action failed")
			}
			return
		}

		JSON(w, http.StatusOK, map[string]interface{}{
			"status":        "action_completed",
			"action":        req.Action,
			"expedition_id": expeditionID,
		})
	}
}
