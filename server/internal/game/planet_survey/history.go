package planet_survey

import "time"

type ExpeditionHistoryEntry struct {
	ID              string
	PlanetID        string
	ExpeditionType  string
	Status          string
	Result          string
	Discovered      string
	LocationType    string
	ResourcesGained map[string]float64
	CreatedAt       time.Time
	CompletedAt     time.Time
}
