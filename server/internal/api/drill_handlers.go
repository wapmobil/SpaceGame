package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"spacegame/internal/game"
)

func handleStartDrill(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		playerID := AuthPlayerFromContext(r).ID

		for _, dg := range game.ActiveSessions() {
			s := dg.GetSession()
			if s.PlayerID == playerID && s.PlanetID == planetID {
				if s.Status == "active" {
					Error(w, http.StatusConflict, "Already have an active drill session")
					return
				}
				game.RemoveSession(s.SessionID)
			}
		}

		var mineLevel int
		err := db.QueryRow("SELECT level FROM buildings WHERE planet_id = $1 AND type = 'mine'", planetID).Scan(&mineLevel)
		if err != nil {
			mineLevel = 0
		}

		cost := 100 * mineLevel

		p := game.Instance().GetPlanet(planetID)
		if p == nil {
			Error(w, http.StatusNotFound, "Planet not found")
			return
		}
		if p.Resources.Iron < float64(cost) {
			Error(w, http.StatusConflict, fmt.Sprintf("Insufficient iron. Need %d, have %.0f", cost, p.Resources.Iron))
			return
		}
		p.Resources.Iron -= float64(cost)
		game.Instance().SavePlanet(p)

		dg := game.NewDrillGame(planetID, playerID, mineLevel)
		session := dg.GetSession()

		dg.SetBroadcastFn(func(result *game.MoveResult) {
			updateData := map[string]interface{}{
				"session_id":     session.SessionID,
				"drill_hp":       result.DrillHP,
				"drill_max_hp":   result.DrillMaxHP,
				"depth":          result.Depth,
				"drill_x":        result.DrillX,
				"resources":      result.Resources,
				"total_earned":   result.TotalEarned,
				"status":         session.Status,
				"game_ended":     result.GameEnded,
			}
			if result.GameEnded {
				updateData["end_reason"] = result.EndReason
			}
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

		Created(w, response)
	}
}

func handleDrillCommand(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		playerID := AuthPlayerFromContext(r).ID

		var req DrillCommandRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		dg := game.FindActiveSession(planetID, playerID)
		if dg == nil {
			Error(w, http.StatusNotFound, "No active drill session")
			return
		}

		dg.SetCommand(req.Direction, req.Extract)

		JSON(w, http.StatusOK, DrillCommandResponse{Status: "command_received"})
	}
}

func handleDrillChunk(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		playerID := AuthPlayerFromContext(r).ID

		dg := game.FindActiveSession(planetID, playerID)
		if dg == nil {
			Error(w, http.StatusNotFound, "No active drill session")
			return
		}

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

		JSON(w, http.StatusOK, response)
	}
}

func handleCompleteDrill(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		playerID := AuthPlayerFromContext(r).ID

		dg := game.FindActiveSession(planetID, playerID)
		if dg == nil {
			Error(w, http.StatusNotFound, "No active drill session")
			return
		}

		totalEarned := dg.Complete()
		sess := dg.GetSession()

		if totalEarned > 0 {
			p := game.Instance().GetPlanet(planetID)
			if p != nil {
				p.Resources.Money += totalEarned
				game.Instance().SavePlanet(p)
			}
		}

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

		JSON(w, http.StatusOK, response)
	}
}

func handleDestroyDrill(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		playerID := AuthPlayerFromContext(r).ID

		dg := game.FindActiveSession(planetID, playerID)
		if dg == nil {
			Error(w, http.StatusNotFound, "No active drill session")
			return
		}

		sess := dg.GetSession()
		if sess.Status != "active" {
			JSON(w, http.StatusOK, map[string]interface{}{
				"status":         sess.Status,
				"drill_hp":       sess.DrillHP,
				"drill_max_hp":   sess.DrillMaxHP,
				"depth":          sess.Depth,
				"drill_x":        sess.DrillX,
				"resources":      make([]DrillResourceResponse, 0),
				"total_earned":   sess.TotalEarned,
				"game_ended":     sess.Status != "active",
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

		JSON(w, http.StatusOK, response)
	}
}

func handleCleanupDrill(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		playerID := AuthPlayerFromContext(r).ID

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
				game.RemoveSession(sess.SessionID)
			}
		}

		JSON(w, http.StatusOK, map[string]interface{}{
			"status": "cleaned",
		})
	}
}
