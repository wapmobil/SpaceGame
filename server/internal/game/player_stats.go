package game

import (
	"log"
	"time"
)

// StatsKey represents a statistics key for tracking.
type StatsKey string

const (
	// Cumulative lifetime stats
	StatDaysPlayed         StatsKey = "days_played"
	StatFirstLogin         StatsKey = "first_login"
	StatLastLogin          StatsKey = "last_login"
	StatTotalFoodProduce   StatsKey = "total_food_produced"
	StatTotalCompositeProduce StatsKey = "total_composite_produced"
	StatTotalMechProduce   StatsKey = "total_mechanisms_produced"
	StatTotalReagentProduce StatsKey = "total_reagents_produced"
	StatTotalMoneyEarned   StatsKey = "total_money_earned"
	StatTotalMoneySpent    StatsKey = "total_money_spent"
	StatTotalBuildings     StatsKey = "total_buildings_constructed"
	StatTotalResearch      StatsKey = "total_research_completed"
	StatTotalExpeditions   StatsKey = "total_expeditions_completed"
	StatTotalAlienTech     StatsKey = "total_alien_tech_earned"
	StatTotalEnergyProd    StatsKey = "total_energy_produced"

	// Per-type stats (indexed by building/ship type)
	StatFarmBuilt          StatsKey = "buildings_farm"
	StatSolarBuilt         StatsKey = "buildings_solar"
	StatStorageBuilt       StatsKey = "buildings_storage"
	StatBaseBuilt          StatsKey = "buildings_base"
	StatEnergyStorageBuilt StatsKey = "buildings_energy_storage"
	StatShipyardBuilt      StatsKey = "buildings_shipyard"

	// Per-ship-type stats
	StatShipScoutBuilt     StatsKey = "ships_scout"
	StatShipFrigateBuilt   StatsKey = "ships_frigate"
	StatShipCruiserBuilt   StatsKey = "ships_cruiser"
	StatShipDestroyerBuilt StatsKey = "ships_destroyer"
	StatShipCarrierBuilt   StatsKey = "ships_carrier"
	StatShipTransportBuilt StatsKey = "ships_transport"
	StatShipDestroyer2Built  StatsKey = "ships_destroyer_2"
	StatShipFrigate2Built    StatsKey = "ships_frigate_2"
	StatShipScout2Built      StatsKey = "ships_scout_2"
)

// AllStatsKeys returns all defined stats keys.
func AllStatsKeys() []StatsKey {
	return []StatsKey{
		StatDaysPlayed,
		StatFirstLogin,
		StatLastLogin,
		StatTotalFoodProduce,
		StatTotalCompositeProduce,
		StatTotalMechProduce,
		StatTotalReagentProduce,
		StatTotalMoneyEarned,
		StatTotalMoneySpent,
		StatTotalBuildings,
		StatTotalResearch,
		StatTotalExpeditions,
		StatTotalAlienTech,
		StatTotalEnergyProd,
		StatFarmBuilt,
		StatSolarBuilt,
		StatStorageBuilt,
		StatBaseBuilt,
		StatEnergyStorageBuilt,
		StatShipyardBuilt,
		StatShipScoutBuilt,
		StatShipFrigateBuilt,
		StatShipCruiserBuilt,
		StatShipDestroyerBuilt,
		StatShipCarrierBuilt,
		StatShipTransportBuilt,
		StatShipDestroyer2Built,
		StatShipFrigate2Built,
		StatShipScout2Built,
	}
}

// TrackBuildingCompleted records that a building was completed.
func (g *Game) TrackBuildingCompleted(planetID, playerID, buildingType string) {
	if g.db == nil {
		return
	}

	var statKey StatsKey
	switch buildingType {
	case "farm":
		statKey = StatFarmBuilt
	case "solar":
		statKey = StatSolarBuilt
	case "storage":
		statKey = StatStorageBuilt
	case "base":
		statKey = StatBaseBuilt
	case "energy_storage":
		statKey = StatEnergyStorageBuilt
	case "shipyard":
		statKey = StatShipyardBuilt
	default:
		return
	}

	_, err := g.db.Exec(`
		INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (player_id, stat_key)
		DO UPDATE SET stat_value = player_stats.stat_value + 1, updated_at = NOW()
	`, playerID, planetID, string(statKey))

	if err != nil {
		log.Printf("Error tracking building %s: %v", buildingType, err)
	}

	// Also update total buildings count
	g.db.Exec(`
		INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (player_id, stat_key)
		DO UPDATE SET stat_value = player_stats.stat_value + 1, updated_at = NOW()
	`, playerID, planetID, string(StatTotalBuildings))
}

// TrackResourceProduced records resource production.
func (g *Game) TrackResourceProduced(planetID, playerID string, prod ProductionResult) {
	if g.db == nil {
		return
	}

	updates := []struct {
		key   StatsKey
		value float64
	}{
		{StatTotalFoodProduce, prod.Food},
		{StatTotalCompositeProduce, prod.Composite},
		{StatTotalMechProduce, prod.Mechanisms},
		{StatTotalReagentProduce, prod.Reagents},
		{StatTotalEnergyProd, prod.Energy},
	}

	for _, u := range updates {
		if u.value > 0 {
			g.db.Exec(`
				INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
				VALUES ($1, $2, $3, $4, NOW())
				ON CONFLICT (player_id, stat_key)
				DO UPDATE SET stat_value = player_stats.stat_value + $4, updated_at = NOW()
			`, playerID, planetID, string(u.key), u.value)
		}
	}

	if prod.Money > 0 {
		g.db.Exec(`
			INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (player_id, stat_key)
			DO UPDATE SET stat_value = player_stats.stat_value + $4, updated_at = NOW()
		`, playerID, planetID, string(StatTotalMoneyEarned), prod.Money)
	}

	if prod.AlienTech > 0 {
		g.db.Exec(`
			INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (player_id, stat_key)
			DO UPDATE SET stat_value = player_stats.stat_value + $4, updated_at = NOW()
		`, playerID, planetID, string(StatTotalAlienTech), prod.AlienTech)
	}
}

// TrackMoneySpent records money spent.
func (g *Game) TrackMoneySpent(planetID, playerID string, amount float64) {
	if g.db == nil || amount <= 0 {
		return
	}

	g.db.Exec(`
		INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (player_id, stat_key)
		DO UPDATE SET stat_value = player_stats.stat_value + $4, updated_at = NOW()
	`, playerID, planetID, string(StatTotalMoneySpent), amount)
}

// TrackExpeditionCompleted records a completed expedition.
func (g *Game) TrackExpeditionCompleted(planetID, playerID string) {
	if g.db == nil {
		return
	}

	g.db.Exec(`
		INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (player_id, stat_key)
		DO UPDATE SET stat_value = player_stats.stat_value + 1, updated_at = NOW()
	`, playerID, planetID, string(StatTotalExpeditions))
}

// TrackResearchCompleted records a completed research.
func (g *Game) TrackResearchCompleted(planetID, playerID string) {
	if g.db == nil {
		return
	}

	g.db.Exec(`
		INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (player_id, stat_key)
		DO UPDATE SET stat_value = player_stats.stat_value + 1, updated_at = NOW()
	`, playerID, planetID, string(StatTotalResearch))
}

// TrackShipBuilt records a ship being built.
func (g *Game) TrackShipBuilt(planetID, playerID, shipType string) {
	if g.db == nil {
		return
	}

	var statKey StatsKey
	switch shipType {
	case "scout":
		statKey = StatShipScoutBuilt
	case "frigate":
		statKey = StatShipFrigateBuilt
	case "cruiser":
		statKey = StatShipCruiserBuilt
	case "destroyer":
		statKey = StatShipDestroyerBuilt
	case "carrier":
		statKey = StatShipCarrierBuilt
	case "transport":
		statKey = StatShipTransportBuilt
	case "destroyer_2":
		statKey = StatShipDestroyer2Built
	case "frigate_2":
		statKey = StatShipFrigate2Built
	case "scout_2":
		statKey = StatShipScout2Built
	default:
		return
	}

	g.db.Exec(`
		INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (player_id, stat_key)
		DO UPDATE SET stat_value = player_stats.stat_value + 1, updated_at = NOW()
	`, playerID, planetID, string(statKey))
}

// UpdateLoginStats updates login-related statistics.
func (g *Game) UpdateLoginStats(planetID, playerID string) {
	if g.db == nil {
		return
	}

	// Get current stats
	var firstLogin string
	err := g.db.QueryRow(`
		SELECT stat_value::TEXT FROM player_stats 
		WHERE player_id = $1 AND stat_key = $2
	`, playerID, string(StatFirstLogin)).Scan(&firstLogin)

	if firstLogin == "" || err != nil {
		// Set first login
		now := time.Now().Format(time.RFC3339)
		g.db.Exec(`
			INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (player_id, stat_key)
			DO UPDATE SET stat_value = EXCLUDED.stat_value, updated_at = NOW()
		`, playerID, planetID, string(StatFirstLogin), now)
	}

	// Always update last login
	now := time.Now().Format(time.RFC3339)
	g.db.Exec(`
		INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (player_id, stat_key)
		DO UPDATE SET stat_value = EXCLUDED.stat_value, updated_at = NOW()
	`, playerID, planetID, string(StatLastLogin), now)

	// Update days played
	var daysPlayed float64
	err = g.db.QueryRow(`
		SELECT stat_value FROM player_stats 
		WHERE player_id = $1 AND stat_key = $2
	`, playerID, string(StatDaysPlayed)).Scan(&daysPlayed)

	if err != nil || daysPlayed == 0 {
		daysPlayed = 1
	} else {
		daysPlayed += 1
	}

	g.db.Exec(`
		INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (player_id, stat_key)
		DO UPDATE SET stat_value = EXCLUDED.stat_value, updated_at = NOW()
	`, playerID, planetID, string(StatDaysPlayed), daysPlayed)
}

// GetPlayerStats retrieves all stats for a player.
func (g *Game) GetPlayerStats(playerID string) (map[string]float64, error) {
	if g.db == nil {
		return nil, nil
	}

	rows, err := g.db.Query(`
		SELECT stat_key, stat_value FROM player_stats
		WHERE player_id = $1
		ORDER BY stat_key
	`, playerID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]float64)
	for rows.Next() {
		var key string
		var value float64
		if err := rows.Scan(&key, &value); err != nil {
			log.Printf("Error scanning stat row: %v", err)
			continue
		}
		stats[key] = value
	}

	return stats, nil
}

// GetDailyStats retrieves daily stats for a player.
func (g *Game) GetDailyStats(playerID string, startDate, endDate time.Time) (map[string]map[string]float64, error) {
	if g.db == nil {
		return nil, nil
	}

	rows, err := g.db.Query(`
		SELECT date, stat_key, stat_value FROM daily_stats
		WHERE player_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date DESC, stat_key
	`, playerID, startDate, endDate)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]map[string]float64)
	for rows.Next() {
		var dateStr string
		var key string
		var value float64
		if err := rows.Scan(&dateStr, &key, &value); err != nil {
			log.Printf("Error scanning daily stat row: %v", err)
			continue
		}
		if result[dateStr] == nil {
			result[dateStr] = make(map[string]float64)
		}
		result[dateStr][key] = value
	}

	return result, nil
}

// ResetDailyStats resets all daily counters and archives them.
func (g *Game) ResetDailyStats() {
	if g.db == nil {
		return
	}

	today := time.Now().Format("2006-01-02")
	_, err := g.db.Exec("SELECT reset_daily_stats($1)", today)
	if err != nil {
		log.Printf("Error resetting daily stats: %v", err)
	} else {
		log.Printf("Daily stats reset for %s", today)
	}
}

// GetEventHistory retrieves event history for a player.
func (g *Game) GetEventHistory(playerID string, limit int) ([]map[string]interface{}, error) {
	if g.db == nil {
		return nil, nil
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := g.db.Query(`
		SELECT id, planet_id, event_type, description, resolved, created_at, resolved_at
		FROM events
		WHERE player_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, playerID, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id int
		var planetID, eventType, description string
		var resolved bool
		var createdAt, resolvedAt interface{}

		if err := rows.Scan(&id, &planetID, &eventType, &description, &resolved, &createdAt, &resolvedAt); err != nil {
			log.Printf("Error scanning event row: %v", err)
			continue
		}

		event := map[string]interface{}{
			"id":          id,
			"planet_id":   planetID,
			"type":        eventType,
			"description": description,
			"resolved":    resolved,
			"created_at":  createdAt,
		}
		if resolvedAt != nil {
			event["resolved_at"] = resolvedAt
		}
		events = append(events, event)
	}

	return events, nil
}
