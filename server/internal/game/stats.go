package game

import (
	"fmt"
	"time"
)

// StatsTracker handles statistics tracking for the game.
type StatsTracker struct {
	game *Game
}

// NewStatsTracker creates a new statistics tracker.
func NewStatsTracker(g *Game) *StatsTracker {
	return &StatsTracker{game: g}
}

// TrackProduction records production for all resources in a tick.
func (st *StatsTracker) TrackProduction(planetID, playerID string, prod ProductionResult) {
	st.game.TrackResourceProduced(planetID, playerID, prod)
}

// TrackShipConstruction records ship construction.
func (st *StatsTracker) TrackShipConstruction(planetID, playerID, shipType string, count int) {
	for i := 0; i < count; i++ {
		st.game.TrackShipBuilt(planetID, playerID, shipType)
	}
}

// TrackBuildingConstruction records building construction.
func (st *StatsTracker) TrackBuildingConstruction(planetID, playerID, buildingType string) {
	st.game.TrackBuildingCompleted(planetID, playerID, buildingType)
}

// TrackExpeditionOutcome records expedition results.
func (st *StatsTracker) TrackExpeditionOutcome(planetID, playerID string, completed bool) {
	if completed {
		st.game.TrackExpeditionCompleted(planetID, playerID)
	}
}

// TrackResearchOutcome records research completion.
func (st *StatsTracker) TrackResearchOutcome(planetID, playerID string, completed bool) {
	if completed {
		st.game.TrackResearchCompleted(planetID, playerID)
	}
}

// TrackMoneyTransaction records money earned or spent.
func (st *StatsTracker) TrackMoneyTransaction(planetID, playerID string, amount float64, earned bool) {
	if earned {
		st.game.TrackResourceProduced(planetID, playerID, ProductionResult{Money: amount})
	} else {
		st.game.TrackMoneySpent(planetID, playerID, amount)
	}
}

// RecordLogin records a player login.
func (st *StatsTracker) RecordLogin(planetID, playerID string) {
	st.game.UpdateLoginStats(planetID, playerID)
}

// GetStatsSummary returns a summary of player statistics.
func (st *StatsTracker) GetStatsSummary(playerID string) (map[string]interface{}, error) {
	stats, err := st.game.GetPlayerStats(playerID)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"lifetime": stats,
		"category_totals": map[string]interface{}{
			"resources_produced": map[string]interface{}{
				"food":          stats[string(StatTotalFoodProduce)],
				"composite":     stats[string(StatTotalCompositeProduce)],
				"mechanisms":    stats[string(StatTotalMechProduce)],
				"reagents":      stats[string(StatTotalReagentProduce)],
				"energy":        stats[string(StatTotalEnergyProd)],
			},
			"buildings_constructed": map[string]interface{}{
				"total":          stats[string(StatTotalBuildings)],
				"farms":          stats[string(StatFarmBuilt)],
				"solar_panels":   stats[string(StatSolarBuilt)],
				"storage":        stats[string(StatStorageBuilt)],
				"bases":          stats[string(StatBaseBuilt)],
				"energy_storages": stats[string(StatEnergyStorageBuilt)],
				"shipyards":      stats[string(StatShipyardBuilt)],
			},
			"ships_built": map[string]interface{}{
				"total":        stats[string(StatShipScoutBuilt)] + stats[string(StatShipFrigateBuilt)] +
					stats[string(StatShipCruiserBuilt)] + stats[string(StatShipDestroyerBuilt)] +
					stats[string(StatShipCarrierBuilt)] + stats[string(StatShipTransportBuilt)] +
					stats[string(StatShipDestroyer2Built)] +
					stats[string(StatShipFrigate2Built)] + stats[string(StatShipScout2Built)],
				"scouts":   stats[string(StatShipScoutBuilt)],
				"frigates": stats[string(StatShipFrigateBuilt)],
				"cruisers": stats[string(StatShipCruiserBuilt)],
			},
			"expeditions": stats[string(StatTotalExpeditions)],
			"research":    stats[string(StatTotalResearch)],
			
		},
	}

	return summary, nil
}

// DailyStatsReset runs the daily statistics reset.
func (st *StatsTracker) DailyStatsReset() {
	st.game.ResetDailyStats()
}

// GetDailyStats returns daily statistics for a player.
func (st *StatsTracker) GetDailyStats(playerID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	dailyStats, err := st.game.GetDailyStats(playerID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"daily_stats": dailyStats,
		"start_date":  startDate.Format("2006-01-02"),
		"end_date":    endDate.Format("2006-01-02"),
	}, nil
}

// GetEventHistory returns the event history for a player.
func (st *StatsTracker) GetEventHistory(playerID string, limit int) ([]map[string]interface{}, error) {
	return st.game.GetEventHistory(playerID, limit)
}

// ProcessTickStats processes statistics for a single tick.
func (st *StatsTracker) ProcessTickStats(planet *Planet) {
	if planet == nil {
		return
	}

	prod := planet.GetProductionResult()
	st.TrackProduction(planet.ID, planet.OwnerID, prod)
}

// GetStatsForPlanet returns all stats for a specific planet's player.
func (st *StatsTracker) GetStatsForPlanet(planetID string) (map[string]interface{}, error) {
	p := st.game.GetPlanet(planetID)
	if p == nil {
		return nil, fmt.Errorf("planet not found: %s", planetID)
	}

	summary, err := st.GetStatsSummary(p.OwnerID)
	if err != nil {
		return nil, err
	}

	summary["planet_id"] = planetID
	summary["planet_name"] = p.Name

	return summary, nil
}
