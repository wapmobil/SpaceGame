package planet_survey

import "time"

type ExpeditionHistoryEntry struct {
	ID              string
	PlanetID        string
	ExpeditionType  string
	Status          string
	Result          string
	Discovered      string
	ResourcesGained map[string]float64
	CreatedAt       time.Time
	CompletedAt     time.Time
}
