package game

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
)

// LoadGardenBedFromDB loads garden bed state from the database for a planet
func LoadGardenBedFromDB(planet *Planet) error {
	if planet == nil || planet.game == nil || planet.game.db == nil {
		return nil
	}

	hasGardenBedGrid, err := planet.game.db.ColumnExists(context.Background(), "planets", "garden_bed_grid")
	if err != nil || !hasGardenBedGrid {
		// Garden bed columns don't exist yet, create default garden bed
		farmLevel := planet.GetBuildingLevel("farm")
		if farmLevel > 0 {
			planet.GardenBedState = NewGardenBedState(farmLevel)
		}
		return nil
	}

	var gardenBedGridJSON []byte

	err = planet.game.db.QueryRow(`
		SELECT garden_bed_grid FROM planets WHERE id = $1
	`, planet.ID).Scan(&gardenBedGridJSON)

	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		log.Printf("Error loading garden bed for planet %s: %v", planet.ID, err)
		return err
	}

	// Load farm building level
	farmLevel := planet.GetBuildingLevel("farm")
	if farmLevel == 0 {
		// No farm building, but we still need to parse the grid if it exists
		if len(gardenBedGridJSON) > 0 {
			planet.GardenBedState = NewGardenBedStateFromJSON(gardenBedGridJSON, 0)
		}
		return nil
	}

	// Always create garden bed state with correct building level
	// If JSON has saved data, parse it; otherwise create empty state
	if len(gardenBedGridJSON) > 0 && string(gardenBedGridJSON) != "null" && string(gardenBedGridJSON) != "[]" {
		var rows []GardenBedRow
		if err := json.Unmarshal(gardenBedGridJSON, &rows); err == nil && len(rows) > 0 {
			// Normalize legacy data: empty string status -> empty
			for i := range rows {
				if rows[i].Status == "" {
					rows[i].Status = GardenBedRowEmpty
				}
			}
			// Ensure row count matches building level
			if len(rows) < farmLevel {
				// Extend rows to match building level
				for i := len(rows); i < farmLevel; i++ {
					rows = append(rows, GardenBedRow{Status: GardenBedRowEmpty, Weeds: 0, WaterTimer: 0, Stage: 0})
				}
			}
			planet.GardenBedState = &GardenBedState{
			Rows:     rows,
			RowCount: len(rows),
		}
		log.Printf("Garden bed state loaded from DB for planet %s (rows=%d)", planet.ID, len(rows))
			return nil
		}
	}

	// No saved data or parse failed, create fresh state
	planet.GardenBedState = NewGardenBedState(farmLevel)
	log.Printf("Garden bed state created fresh for planet %s (level=%d, rows=%d)", planet.ID, farmLevel, farmLevel)

	return nil
}

// SaveGardenBedToDB saves garden bed state to the database for a planet
func SaveGardenBedToDB(planet *Planet) {
	if planet == nil || planet.game == nil || planet.game.db == nil {
		return
	}

	hasGardenBedGrid, err := planet.game.db.ColumnExists(context.Background(), "planets", "garden_bed_grid")
	if err != nil || !hasGardenBedGrid {
		return
	}

	if planet.GardenBedState == nil {
		return
	}

	gardenBedGridData, err := json.Marshal(planet.GardenBedState.Rows)
	if err != nil {
		log.Printf("Error marshaling garden bed grid for planet %s: %v", planet.ID, err)
		return
	}

	_, err = planet.game.db.Exec(`
		UPDATE planets 
		SET garden_bed_grid = $1::jsonb, updated_at = NOW()
		WHERE id = $2
	`, string(gardenBedGridData), planet.ID)

	if err != nil {
		log.Printf("Error saving garden bed for planet %s: %v", planet.ID, err)
	}
}

// GetGardenBedState returns the garden bed state as JSON bytes.
// Returns nil when no garden bed exists on the planet.
func GetGardenBedState(planet *Planet) ([]byte, error) {
	if planet == nil {
		return nil, nil
	}

	// Lazy init: create GardenBedState if farm building exists but state is nil
	if planet.GardenBedState == nil {
		farmLevel := planet.GetBuildingLevel("farm")
		log.Printf("GetGardenBedState: GardenBedState is nil for planet %s, farmLevel=%d", planet.ID, farmLevel)
		if farmLevel > 0 {
			planet.GardenBedState = NewGardenBedState(farmLevel)
			log.Printf("GetGardenBedState: created fresh garden bed state for planet %s (rows=%d)", planet.ID, farmLevel)
		}
	}

	if planet.GardenBedState == nil {
		return nil, nil
	}

	result := map[string]interface{}{
		"rows":      planet.GardenBedState.Rows,
		"row_count": planet.GardenBedState.RowCount,
	}

	return json.Marshal(result)
}

// GardenBedAction handles player garden bed actions via the API
func GardenBedAction(planet *Planet, action string, rowIndex int, plantType string) ([]byte, error) {
	result, err := gardenBedAction(planet, action, rowIndex, plantType)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// GardenBedActionInternal is the internal garden bed action handler (used by both API and WS)
func GardenBedActionInternal(planet *Planet, action string, rowIndex int, plantType string) (*GardenBedActionResult, error) {
	return gardenBedAction(planet, action, rowIndex, plantType)
}

// UpdateGardenBedGrid updates the garden bed grid from the API and saves to DB
func UpdateGardenBedGrid(planet *Planet, rows []GardenBedRow) error {
	if planet == nil || planet.GardenBedState == nil {
		return &PlanetError{PlanetID: planet.ID, Reason: "no_garden_bed"}
	}

	if len(rows) != planet.GardenBedState.RowCount {
		return &PlanetError{PlanetID: planet.ID, Reason: "row_count_mismatch"}
	}

	planet.GardenBedState.Rows = rows
	SaveGardenBedToDB(planet)
	return nil
}
