package game

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
)

// LoadFarmFromDB loads farm state from the database for a planet
func LoadFarmFromDB(planet *Planet) error {
	if planet == nil || planet.game == nil || planet.game.db == nil {
		return nil
	}

	hasFarmGrid, err := planet.game.db.ColumnExists(context.Background(), "planets", "farm_grid")
	if err != nil || !hasFarmGrid {
		// Farm columns don't exist yet, create default farm
		farmLevel := planet.GetBuildingLevel("farm")
		if farmLevel > 0 {
			planet.FarmState = NewFarmState(farmLevel)
		}
		return nil
	}

	var farmGridJSON []byte
	var farmLastTick int64

	err = planet.game.db.QueryRow(`
		SELECT farm_grid, farm_last_tick FROM planets WHERE id = $1
	`, planet.ID).Scan(&farmGridJSON, &farmLastTick)

	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		log.Printf("Error loading farm for planet %s: %v", planet.ID, err)
		return err
	}

	// Load farm building level
	farmLevel := planet.GetBuildingLevel("farm")
	if farmLevel == 0 {
		// No farm building, but we still need to parse the grid if it exists
		if len(farmGridJSON) > 0 {
			planet.FarmState = NewFarmStateFromJSON(farmGridJSON, 0)
		}
		return nil
	}

	// Always create farm state with correct building level
	// If JSON has saved data, parse it; otherwise create empty state
	if len(farmGridJSON) > 0 && string(farmGridJSON) != "null" && string(farmGridJSON) != "[]" {
		var rows []FarmRow
		if err := json.Unmarshal(farmGridJSON, &rows); err == nil && len(rows) > 0 {
			// Ensure row count matches building level
			if len(rows) < farmLevel {
				// Extend rows to match building level
				for i := len(rows); i < farmLevel; i++ {
					rows = append(rows, FarmRow{Status: FarmRowEmpty, Weeds: 0, WaterTimer: 0, Stage: 0})
				}
			}
			planet.FarmState = &FarmState{
				Rows:     rows,
				LastTick: farmLastTick,
				RowCount: len(rows),
			}
			planet.FarmLastTick = farmLastTick
			return nil
		}
	}

	// No saved data or parse failed, create fresh state
	planet.FarmState = NewFarmState(farmLevel)
	planet.FarmLastTick = farmLastTick

	return nil
}

// SaveFarmToDB saves farm state to the database for a planet
func SaveFarmToDB(planet *Planet) {
	if planet == nil || planet.game == nil || planet.game.db == nil {
		return
	}

	hasFarmGrid, err := planet.game.db.ColumnExists(context.Background(), "planets", "farm_grid")
	if err != nil || !hasFarmGrid {
		return
	}

	if planet.FarmState == nil {
		return
	}

	farmGridData, err := json.Marshal(planet.FarmState.Rows)
	if err != nil {
		log.Printf("Error marshaling farm grid for planet %s: %v", planet.ID, err)
		return
	}

	_, err = planet.game.db.Exec(`
		UPDATE planets 
		SET farm_grid = $1::jsonb, farm_last_tick = $2, updated_at = NOW()
		WHERE id = $3
	`, string(farmGridData), planet.FarmState.LastTick, planet.ID)

	if err != nil {
		log.Printf("Error saving farm for planet %s: %v", planet.ID, err)
	}
}

// GetFarmState returns the farm state as JSON bytes
func GetFarmState(planet *Planet) ([]byte, error) {
	if planet == nil {
		return json.Marshal(map[string]interface{}{
			"rows":      []interface{}{},
			"last_tick": 0,
			"row_count": 0,
		})
	}

	// Lazy init: create FarmState if farm building exists but state is nil
	if planet.FarmState == nil {
		farmLevel := planet.GetBuildingLevel("farm")
		if farmLevel > 0 {
			planet.FarmState = NewFarmState(farmLevel)
		}
	}

	if planet.FarmState == nil {
		return json.Marshal(map[string]interface{}{
			"rows":      []interface{}{},
			"last_tick": 0,
			"row_count": 0,
		})
	}

	result := map[string]interface{}{
		"rows":      planet.FarmState.Rows,
		"last_tick": planet.FarmState.LastTick,
		"row_count": planet.FarmState.RowCount,
	}

	return json.Marshal(result)
}

// FarmAction handles player farm actions via the API
func FarmAction(planet *Planet, action string, rowIndex int, plantType string) ([]byte, error) {
	result, err := farmAction(planet, action, rowIndex, plantType)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// FarmActionInternal is the internal farm action handler (used by both API and WS)
func FarmActionInternal(planet *Planet, action string, rowIndex int, plantType string) (*FarmActionResult, error) {
	return farmAction(planet, action, rowIndex, plantType)
}

// UpdateFarmGrid updates the farm grid from the API and saves to DB
func UpdateFarmGrid(planet *Planet, rows []FarmRow) error {
	if planet == nil || planet.FarmState == nil {
		return &PlanetError{PlanetID: planet.ID, Reason: "no_farm"}
	}

	if len(rows) != planet.FarmState.RowCount {
		return &PlanetError{PlanetID: planet.ID, Reason: "row_count_mismatch"}
	}

	planet.FarmState.Rows = rows
	SaveFarmToDB(planet)
	return nil
}
