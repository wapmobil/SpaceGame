package game

import (
	"database/sql"
	"log"
	"time"
)

// RatingCategory defines the categories for leaderboard rankings.
type RatingCategory string

const (
	RatingMoney          RatingCategory = "money"
	RatingFood           RatingCategory = "food"
	RatingShips          RatingCategory = "ships"
	RatingBuildings      RatingCategory = "buildings"
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
