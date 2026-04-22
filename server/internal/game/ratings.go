package game

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// RatingCategory defines the categories for leaderboard rankings.
type RatingCategory string

const (
	RatingMoney        RatingCategory = "money"
	RatingFood         RatingCategory = "food"
	RatingShips        RatingCategory = "ships"
	RatingBuildings    RatingCategory = "buildings"
	RatingTotalResources RatingCategory = "total_resources"
)

// AllRatingCategories returns the list of all rating categories.
func AllRatingCategories() []RatingCategory {
	return []RatingCategory{
		RatingMoney,
		RatingFood,
		RatingShips,
		RatingBuildings,
		RatingTotalResources,
	}
}

// RatingEntry represents a single entry in the leaderboard.
type RatingEntry struct {
	Rank       int             `json:"rank"`
	PlanetID   string          `json:"planet_id"`
	PlayerName string          `json:"player_name"`
	Category   string          `json:"category"`
	Value      float64         `json:"value"`
	Updated    time.Time       `json:"updated"`
}

// RatingsResult is the response for a ratings query.
type RatingsResult struct {
	Category string         `json:"category"`
	Entries  []RatingEntry  `json:"entries"`
	Total    int            `json:"total"`
}

// RandomEventType defines the type of random event.
type RandomEventType string

const (
	RandomEventShortCircuit RandomEventType = "short_circuit"
	RandomEventTheft        RandomEventType = "theft"
	RandomEventStorageCollapse RandomEventType = "storage_collapse"
	RandomEventMineCollapse RandomEventType = "mine_collapse"
)

// RandomEvent represents a triggered random event.
type RandomEvent struct {
	Type        RandomEventType `json:"type"`
	Description string          `json:"description"`
	PlanetID    string          `json:"planet_id"`
	PlayerID    string          `json:"player_id"`
	Timestamp   time.Time       `json:"timestamp"`
	Resolved    bool            `json:"resolved"`
	ResolveCost map[string]float64 `json:"resolve_cost,omitempty"`
}

// EventDef defines a random event configuration.
type EventDef struct {
	Type        RandomEventType
	Chance      float64       // probability per tick (e.g., 0.02 = 2%)
	Description string
	ResolveCost map[string]float64
	ApplyFn     func(p *Planet) (string, error)
}

// ComputeRatings computes ratings for all categories across all planets.
func (g *Game) ComputeRatings() {
	g.mu.RLock()
	planets := make([]*Planet, 0, len(g.planets))
	for _, p := range g.planets {
		planets = append(planets, p)
	}
	g.mu.RUnlock()

	if len(planets) == 0 {
		log.Println("No planets to compute ratings for")
		return
	}

	if g.db == nil {
		log.Println("No database connection, skipping ratings computation")
		return
	}

	for _, cat := range AllRatingCategories() {
		log.Printf("Computing ratings for category: %s", cat)
		_, err := g.db.Exec("SELECT compute_ratings_for_category($1)", string(cat))
		if err != nil {
			log.Printf("Error computing ratings for %s: %v", cat, err)
		}
	}

	log.Printf("Ratings computed for %d planets across %d categories", len(planets), len(AllRatingCategories()))
}

// GetRatings retrieves leaderboard entries for a given category.
func (g *Game) GetRatings(category string, limit int, planetID string) (*RatingsResult, error) {
	if category == "" {
		category = string(RatingTotalResources)
	}

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	var rows *sql.Rows
	var err error

	if planetID != "" {
		rows, err = g.db.Query(`
			SELECT r.planet_id, p.name as player_name, r.value, r.updated_at
			FROM ratings r
			JOIN planets p ON r.planet_id = p.id
			WHERE r.category = $1 AND r.planet_id = $2
			ORDER BY r.value DESC
		`, category, planetID)
	} else {
		rows, err = g.db.Query(`
			SELECT r.planet_id, p.name as player_name, r.value, r.updated_at
			FROM ratings r
			JOIN planets p ON r.planet_id = p.id
			WHERE r.category = $1
			ORDER BY r.value DESC
			LIMIT $2
		`, category, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []RatingEntry
	rank := 0
	for rows.Next() {
		var entry RatingEntry
		entry.Category = category
		if err := rows.Scan(&entry.PlanetID, &entry.PlayerName, &entry.Value, &entry.Updated); err != nil {
			log.Printf("Error scanning rating row: %v", err)
			continue
		}
		rank++
		entry.Rank = rank
		entries = append(entries, entry)
	}

	return &RatingsResult{
		Category: category,
		Entries:  entries,
		Total:    len(entries),
	}, nil
}

// GetPlayerRank retrieves a specific player's rank in a category.
func (g *Game) GetPlayerRank(category string, planetID string) (*RatingEntry, error) {
	if category == "" {
		category = string(RatingTotalResources)
	}

	var rank int
	var playerPlanetID string
	var playerName string
	var value float64
	var updated time.Time

	err := g.db.QueryRow(`
		SELECT r.rank, r.planet_id, r.player_name, r.value, r.updated_at
		FROM (
			SELECT planet_id, player_name, value,
			       ROW_NUMBER() OVER (ORDER BY value DESC) AS rank
			FROM ratings
			WHERE category = $1
		) r
		JOIN planets p ON r.planet_id = p.id
		WHERE r.planet_id = $2
	`, category, planetID).Scan(&rank, &playerPlanetID, &playerName, &value, &updated)

	if err != nil {
		return nil, err
	}

	return &RatingEntry{
		Rank:       rank,
		PlanetID:   playerPlanetID,
		PlayerName: playerName,
		Category:   category,
		Value:      value,
		Updated:    updated,
	}, nil
}

// GetRandomEvents returns the list of configured random events.
func GetRandomEvents() []EventDef {
	return []EventDef{
		{
			Type:      RandomEventShortCircuit,
			Chance:    0.005, // 0.5% per tick
			Description: "Short Circuit: Energy production disrupted for 1 tick. Pay resources to fix.",
			ResolveCost: map[string]float64{
				"money": 100,
			},
			ApplyFn: applyShortCircuit,
		},
		{
			Type:      RandomEventTheft,
			Chance:    0.005, // 0.5% per tick
			Description: "Theft: Lost 5-20% of money to space pirates.",
			ResolveCost: map[string]float64{},
			ApplyFn: applyTheft,
		},
		{
			Type:      RandomEventStorageCollapse,
			Chance:    0.005, // 0.5% per tick
			Description: "Storage Roof Collapse: Lost 5-20% of stored resources.",
			ResolveCost: map[string]float64{
				"money": 50,
			},
			ApplyFn: applyStorageCollapse,
		},
		{
			Type:      RandomEventMineCollapse,
			Chance:    0.003, // 0.3% per tick
			Description: "Mine Collapse: Lost a mining mini-game level.",
			ResolveCost: map[string]float64{
				"money": 200,
			},
			ApplyFn: applyMineCollapse,
		},
	}
}

// TriggerRandomEvents checks all planets for random events and applies them.
func (g *Game) TriggerRandomEvents() {
	g.mu.RLock()
	planets := make([]*Planet, 0, len(g.planets))
	for _, p := range g.planets {
		planets = append(planets, p)
	}
	g.mu.RUnlock()

	events := GetRandomEvents()

	for _, p := range planets {
		for _, eventDef := range events {
			if rand.Float64() < eventDef.Chance {
				log.Printf("Random event triggered on planet %s: %s", p.ID, eventDef.Description)

				description, err := eventDef.ApplyFn(p)
				if err != nil {
					log.Printf("Error applying event %s to planet %s: %v", eventDef.Type, p.ID, err)
					continue
				}

				// Log event to database
				if g.db != nil {
					_, err := g.db.Exec(
						"INSERT INTO events (planet_id, player_id, event_type, description) VALUES ($1, $2, $3, $4)",
						p.ID, p.OwnerID, string(eventDef.Type), description,
					)
					if err != nil {
						log.Printf("Error logging event for planet %s: %v", p.ID, err)
					} else {
						log.Printf("Event logged for planet %s: %s", p.ID, string(eventDef.Type))
					}
				}

				// Broadcast event via WebSocket
				if g.broadcastFunc != nil {
					state := p.GetState()
					g.broadcastFunc(p.ID, p.OwnerID, state)
				}
			}
		}
	}
}

func applyShortCircuit(p *Planet) (string, error) {
	// Reset energy production for 1 tick by temporarily setting energy to 0
	p.EnergyBuffer.Value = 0
	p.Resources.MaxEnergy = 0
	return fmt.Sprintf("Short circuit: Energy production reset. Pay 100 money to restore production."), nil
}

func applyTheft(p *Planet) (string, error) {
	// Lose 5-20% of money
	lossPercent := 0.05 + rand.Float64()*0.15 // 5% to 20%
	loss := p.Resources.Money * lossPercent
	p.Resources.Money -= loss
	if p.Resources.Money < 0 {
		p.Resources.Money = 0
	}
	return fmt.Sprintf("Theft: Lost %.0f money (%.0f%%) to space pirates.", loss, lossPercent*100), nil
}

func applyStorageCollapse(p *Planet) (string, error) {
	// Lose 5-20% of a random stored resource
	resources := []struct {
		name  string
		value *float64
	}{
		{"food", &p.Resources.Food},
		{"composite", &p.Resources.Composite},
		{"mechanisms", &p.Resources.Mechanisms},
		{"reagents", &p.Resources.Reagents},
	}

	// Only consider resources that have some amount
	var available []struct {
		name  string
		value *float64
	}
	for _, r := range resources {
		if *r.value > 0 {
			available = append(available, r)
		}
	}

	if len(available) == 0 {
		return "Storage collapse: No resources to lose.", nil
	}

	// Pick a random resource
	selected := available[rand.Intn(len(available))]
	lossPercent := 0.05 + rand.Float64()*0.15
	loss := *selected.value * lossPercent
	*selected.value -= loss
	if *selected.value < 0 {
		*selected.value = 0
	}

	return fmt.Sprintf("Storage collapse: Lost %.0f %s (%.0f%%) due to roof collapse.", loss, selected.name, lossPercent*100), nil
}

func applyMineCollapse(p *Planet) (string, error) {
	// Reduce planet level (representing mining mini-game level loss)
	if p.Level > 1 {
		oldLevel := p.Level
		p.Level--
		return fmt.Sprintf("Mine collapse: Planet level reduced from %d to %d (lost a mining mini-game level).", oldLevel, p.Level), nil
	}
	return "Mine collapse: Could not reduce level further (already at minimum).", nil
}

// ComputePlanetRatingValue calculates the total resource value for a planet.
func ComputePlanetRatingValue(p *Planet) float64 {
	total := p.Resources.Food + p.Resources.Composite + p.Resources.Mechanisms +
		p.Resources.Reagents + p.Resources.Money + p.Resources.AlienTech
	return total
}

// ComputePlanetShips returns the total ship count for a planet.
func ComputePlanetShips(p *Planet) float64 {
	return float64(p.GetTotalShipCount())
}

// ComputePlanetBuildings returns the total building count for a planet.
func ComputePlanetBuildings(p *Planet) float64 {
	return float64(p.GetTotalBuildingLevels())
}

// ComputePlanetMoney returns the money value for a planet.
func ComputePlanetMoney(p *Planet) float64 {
	return p.Resources.Money
}

// ComputePlanetFood returns the food value for a planet.
func ComputePlanetFood(p *Planet) float64 {
	return p.Resources.Food
}

// ResolveEvent attempts to resolve a random event by paying the required cost.
func (g *Game) ResolveEvent(planetID string, eventType string) (string, error) {
	p := g.GetPlanet(planetID)
	if p == nil {
		return "", &PlanetError{PlanetID: planetID, Reason: "planet_not_found"}
	}

	events := GetRandomEvents()
	var eventDef *EventDef
	for i := range events {
		if events[i].Type == RandomEventType(eventType) {
			eventDef = &events[i]
			break
		}
	}

	if eventDef == nil {
		return "", &PlanetError{PlanetID: planetID, Reason: "unknown_event_type"}
	}

	// Check if player can afford the resolve cost
	for resource, cost := range eventDef.ResolveCost {
		switch resource {
		case "money":
			if p.Resources.Money < cost {
				return fmt.Sprintf("Cannot resolve event: need %.0f money, have %.0f", cost, p.Resources.Money), nil
			}
		case "food":
			if p.Resources.Food < cost {
				return fmt.Sprintf("Cannot resolve event: need %.0f food, have %.0f", cost, p.Resources.Food), nil
			}
		case "composite":
			if p.Resources.Composite < cost {
				return fmt.Sprintf("Cannot resolve event: need %.0f composite, have %.0f", cost, p.Resources.Composite), nil
			}
		case "mechanisms":
			if p.Resources.Mechanisms < cost {
				return fmt.Sprintf("Cannot resolve event: need %.0f mechanisms, have %.0f", cost, p.Resources.Mechanisms), nil
			}
		case "reagents":
			if p.Resources.Reagents < cost {
				return fmt.Sprintf("Cannot resolve event: need %.0f reagents, have %.0f", cost, p.Resources.Reagents), nil
			}
		}
	}

	// Pay the cost
	for resource, cost := range eventDef.ResolveCost {
		switch resource {
		case "money":
			p.Resources.Money -= cost
		case "food":
			p.Resources.Food -= cost
		case "composite":
			p.Resources.Composite -= cost
		case "mechanisms":
			p.Resources.Mechanisms -= cost
		case "reagents":
			p.Resources.Reagents -= cost
		}
	}

	// Resolve the event effects
	switch eventDef.Type {
	case RandomEventShortCircuit:
		// Restore energy production
		p.Resources.MaxEnergy = p.calculateMaxEnergy()
		production, consumption := p.calculateEnergyBalance()
		if production >= consumption {
			p.EnergyBuffer.Value = p.Resources.MaxEnergy
		}
	case RandomEventTheft:
		// No recovery needed - money is lost
	case RandomEventStorageCollapse:
		// No recovery needed - resources are lost
	case RandomEventMineCollapse:
		// Level is reduced, player can rebuild
	}

	// Log resolved event in database
	if g.db != nil {
		_, err := g.db.Exec(
			"UPDATE events SET resolved = TRUE, resolved_at = NOW() WHERE planet_id = $1 AND event_type = $2 AND resolved = FALSE ORDER BY created_at DESC LIMIT 1",
			planetID, eventType,
		)
		if err != nil {
			log.Printf("Error updating event as resolved: %v", err)
		}
	}

	return fmt.Sprintf("Event resolved: %s. Paid %.0f money in repair costs.", eventDef.Description, eventDef.ResolveCost["money"]), nil
}

// StatsKey represents a statistics key for tracking.
type StatsKey string

const (
	// Cumulative lifetime stats
	StatDaysPlayed       StatsKey = "days_played"
	StatFirstLogin       StatsKey = "first_login"
	StatLastLogin        StatsKey = "last_login"
	StatTotalFoodProduce StatsKey = "total_food_produced"
	StatTotalCompositeProduce StatsKey = "total_composite_produced"
	StatTotalMechProduce StatsKey = "total_mechanisms_produced"
	StatTotalReagentProduce StatsKey = "total_reagents_produced"
	StatTotalMoneyEarned StatsKey = "total_money_earned"
	StatTotalMoneySpent  StatsKey = "total_money_spent"
	StatTotalBuildings   StatsKey = "total_buildings_constructed"
	StatTotalResearch    StatsKey = "total_research_completed"
	StatTotalBattlesWon  StatsKey = "total_battles_won"
	StatTotalBattlesLost StatsKey = "total_battles_lost"
	StatTotalExpeditions StatsKey = "total_expeditions_completed"
	StatMiningPlayed     StatsKey = "total_mining_sessions_played"
	StatMiningCompleted  StatsKey = "total_mining_sessions_completed"
	StatMiningFailed     StatsKey = "total_mining_sessions_failed"
	StatTotalAlienTech   StatsKey = "total_alien_tech_earned"
	StatTotalEnergyProd  StatsKey = "total_energy_produced"

	// Per-type stats (indexed by building/ship type)
	StatFarmBuilt   StatsKey = "buildings_farm"
	StatSolarBuilt  StatsKey = "buildings_solar"
	StatStorageBuilt StatsKey = "buildings_storage"
	StatBaseBuilt   StatsKey = "buildings_base"
	StatFactoryBuilt StatsKey = "buildings_factory"
	StatEnergyStorageBuilt StatsKey = "buildings_energy_storage"
	StatShipyardBuilt StatsKey = "buildings_shipyard"
	StatComcenterBuilt StatsKey = "buildings_comcenter"

	// Per-ship-type stats
	StatShipScoutBuilt   StatsKey = "ships_scout"
	StatShipFrigateBuilt StatsKey = "ships_frigate"
	StatShipCruiserBuilt StatsKey = "ships_cruiser"
	StatShipDestroyerBuilt StatsKey = "ships_destroyer"
	StatShipCarrierBuilt StatsKey = "ships_carrier"
	StatShipTransportBuilt StatsKey = "ships_transport"
	StatShipBattlecruiserBuilt StatsKey = "ships_battlecruiser"
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
		StatTotalBattlesWon,
		StatTotalBattlesLost,
		StatTotalExpeditions,
		StatMiningPlayed,
		StatMiningCompleted,
		StatMiningFailed,
		StatTotalAlienTech,
		StatTotalEnergyProd,
		StatFarmBuilt,
		StatSolarBuilt,
		StatStorageBuilt,
		StatBaseBuilt,
		StatFactoryBuilt,
		StatEnergyStorageBuilt,
		StatShipyardBuilt,
		StatComcenterBuilt,
		StatShipScoutBuilt,
		StatShipFrigateBuilt,
		StatShipCruiserBuilt,
		StatShipDestroyerBuilt,
		StatShipCarrierBuilt,
		StatShipTransportBuilt,
		StatShipBattlecruiserBuilt,
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
	case "factory":
		statKey = StatFactoryBuilt
	case "energy_storage":
		statKey = StatEnergyStorageBuilt
	case "shipyard":
		statKey = StatShipyardBuilt
	case "comcenter":
		statKey = StatComcenterBuilt
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

// TrackBattleResult records a battle outcome.
func (g *Game) TrackBattleResult(planetID, playerID string, won bool) {
	if g.db == nil {
		return
	}

	key := StatTotalBattlesLost
	if won {
		key = StatTotalBattlesWon
	}

	g.db.Exec(`
		INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (player_id, stat_key)
		DO UPDATE SET stat_value = player_stats.stat_value + 1, updated_at = NOW()
	`, playerID, planetID, string(key))
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

// TrackMiningSession records a mining session result.
func (g *Game) TrackMiningSession(planetID, playerID string, completed bool) {
	if g.db == nil {
		return
	}

	g.db.Exec(`
		INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (player_id, stat_key)
		DO UPDATE SET stat_value = player_stats.stat_value + 1, updated_at = NOW()
	`, playerID, planetID, string(StatMiningPlayed))

	if completed {
		g.db.Exec(`
			INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
			VALUES ($1, $2, $3, 1, NOW())
			ON CONFLICT (player_id, stat_key)
			DO UPDATE SET stat_value = player_stats.stat_value + 1, updated_at = NOW()
		`, playerID, planetID, string(StatMiningCompleted))
	} else {
		g.db.Exec(`
			INSERT INTO player_stats (player_id, planet_id, stat_key, stat_value, updated_at)
			VALUES ($1, $2, $3, 1, NOW())
			ON CONFLICT (player_id, stat_key)
			DO UPDATE SET stat_value = player_stats.stat_value + 1, updated_at = NOW()
		`, playerID, planetID, string(StatMiningFailed))
	}
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
	case "battlecruiser":
		statKey = StatShipBattlecruiserBuilt
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


