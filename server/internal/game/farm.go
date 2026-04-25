package game

import (
	"encoding/json"
	"math/rand"
	"sync"
	"time"
)

// Plant types
const (
	PlantWheat   = "wheat"
	PlantBerries = "berries"
	PlantMelon   = "melon"
)

// Farm row statuses
const (
	FarmRowEmpty   = "empty"
	FarmRowPlanted = "planted"
	FarmRowMature  = "mature"
)

// FarmPlant defines a plant type
type FarmPlant struct {
	Type       string
	Name       string
	Icon       string
	FoodReward float64
	Stages     int
	StageNames []string
}

var farmPlants = map[string]*FarmPlant{
	PlantWheat: {Type: PlantWheat, Name: "Пшеница", Icon: "🌾", FoodReward: 5, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}},
	PlantBerries: {Type: PlantBerries, Name: "Ягоды", Icon: "🫐", FoodReward: 15, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}},
	PlantMelon:   {Type: PlantMelon, Name: "Космическая дыня", Icon: "🍈", FoodReward: 30, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}},
}

// GetFarmPlant returns a farm plant by type, or nil if unknown
func GetFarmPlant(typeID string) *FarmPlant {
	return farmPlants[typeID]
}

// GetAllFarmPlants returns all farm plant definitions
func GetAllFarmPlants() []*FarmPlant {
	result := make([]*FarmPlant, 0, len(farmPlants))
	for _, p := range farmPlants {
		result = append(result, p)
	}
	return result
}

// FarmRow represents a single row in the farm grid
type FarmRow struct {
	Status           string  `json:"status"`
	PlantType        string  `json:"plant_type,omitempty"`
	Stage            int     `json:"stage,omitempty"`
	Weeds            int     `json:"weeds"`
	WaterTimer       int     `json:"water_timer"`
	LastTick         int64   `json:"last_tick"`
	FarmTicksSinceLast int   `json:"-"` // internal: farm ticks since last growth check
}

// FarmState represents the complete farm state for a planet
type FarmState struct {
	Rows      []FarmRow `json:"rows"`
	LastTick  int64     `json:"last_tick"`
	RowCount  int       `json:"row_count"`
}

// NewFarmState creates a new empty farm with the given row count
func NewFarmState(rowCount int) *FarmState {
	rows := make([]FarmRow, rowCount)
	for i := range rows {
		rows[i] = FarmRow{
			Status:   FarmRowEmpty,
			Weeds:    0,
			WaterTimer: 0,
			Stage:    0,
		}
	}
	return &FarmState{
		Rows:     rows,
		LastTick: 0,
		RowCount: rowCount,
	}
}

// NewFarmStateFromJSON deserializes farm state from JSONB data
func NewFarmStateFromJSON(data []byte, rowCount int) *FarmState {
	if len(data) == 0 || string(data) == "null" || string(data) == "[]" {
		return NewFarmState(rowCount)
	}

	// Try parsing as FarmState first
	var state FarmState
	if err := json.Unmarshal(data, &state); err != nil {
		// Try parsing as []FarmRow
		var rows []FarmRow
		if err2 := json.Unmarshal(data, &rows); err2 != nil {
			return NewFarmState(rowCount)
		}
		state.Rows = rows
		state.RowCount = len(rows)
	}

	// Normalize legacy data: empty string status -> empty
	for i := range state.Rows {
		if state.Rows[i].Status == "" {
			state.Rows[i].Status = FarmRowEmpty
		}
	}

	// Ensure row count matches
	if len(state.Rows) != state.RowCount {
		state.RowCount = len(state.Rows)
	}

	return &state
}

// FarmTick processes one farm tick (every farmTickInterval game ticks)
// farmTickNum is the farm tick number (gameTick / farmTickInterval)
// Returns true if the state changed
func FarmTick(farm *FarmState, farmTickNum int64) bool {
	if farm == nil || len(farm.Rows) == 0 {
		return false
	}

	changed := false

	for i := range farm.Rows {
		row := &farm.Rows[i]

		// Only process rows with planted (non-mature) plants
		if row.Status != FarmRowPlanted {
			continue
		}

		// Skip if we've already processed this tick
		if row.LastTick == farmTickNum {
			continue
		}

		plant := farmPlants[row.PlantType]
		if plant == nil {
			row.Status = FarmRowEmpty
			row.PlantType = ""
			row.Stage = 0
			row.Weeds = 0
			row.WaterTimer = 0
			row.LastTick = farmTickNum
			row.FarmTicksSinceLast = 0
			changed = true
			continue
		}

		// Weed spawn: 10% chance per tick, up to 3 weeds
		if rand.Float64() < 0.10 && row.Weeds < 3 {
			row.Weeds++
			changed = true
		}

		// Check if watered BEFORE decrementing
		isWatered := row.WaterTimer > 0

		// Water timer decrements
		if row.WaterTimer > 0 {
			row.WaterTimer--
			changed = true
		}

		// Growth: only if weeds < 3 (not fully blocked)
		if row.Weeds < 3 {
			advanceStage := false

			if isWatered {
				// Watered: advance 1 stage per tick
				advanceStage = true
			} else {
				// Normal: advance 1 stage every 2 ticks
				row.FarmTicksSinceLast++
				if row.FarmTicksSinceLast >= 2 {
					advanceStage = true
					row.FarmTicksSinceLast = 0
				}
			}

			if advanceStage {
				newStage := row.Stage + 1
				if newStage >= plant.Stages-1 {
					newStage = plant.Stages - 1
					row.Status = FarmRowMature
				}
				row.Stage = newStage
				changed = true
			}
		}

		row.LastTick = farmTickNum
	}

	farm.LastTick = farmTickNum
	return changed
}

// ProcessFarmTick is the main entry point called from planet_tick.go
// Only processes when gameTick % farmTickInterval == 0 (once every N game ticks)
func ProcessFarmTick(planet *Planet, gameTick int64) {
	if planet == nil || planet.FarmState == nil {
		return
	}

	const farmTickInterval = 10

	// Only process every N ticks
	if gameTick%farmTickInterval != 0 {
		return
	}

	farmTickNum := gameTick / farmTickInterval

	// Only process if we haven't already processed up to this farm tick
	if farmTickNum <= planet.FarmState.LastTick {
		return
	}

	FarmTick(planet.FarmState, farmTickNum)

	// Save farm state to DB
	SaveFarmToDB(planet)
}

// FarmActionResult is the result of a farm action
type FarmActionResult struct {
	Success  bool     `json:"success"`
	Error    string   `json:"error,omitempty"`
	Rows     []FarmRow `json:"rows"`
	LastTick int64    `json:"last_tick"`
	FoodGain float64  `json:"food_gain,omitempty"`
}

// FarmActionCooldown is the cooldown between farm actions in seconds
const FarmActionCooldown = 5 * time.Second

// farmLastActionTime stores the last action time per planet
var farmLastActionTime = make(map[string]time.Time)
var farmActionMu = &sync.Mutex{}

// getFarmActionKey returns the cache key for a planet's farm action cooldown
func getFarmActionKey(planetID string) string {
	return "farm:" + planetID
}

// ClearFarmCooldown clears the cooldown for a planet (for testing)
func ClearFarmCooldown(planetID string) {
	farmActionMu.Lock()
	delete(farmLastActionTime, getFarmActionKey(planetID))
	farmActionMu.Unlock()
}

// farmAction handles player farm actions with cooldown enforcement
// Returns the result as JSON bytes
func farmAction(planet *Planet, action string, rowIndex int, plantType string) (*FarmActionResult, error) {
	if planet == nil || planet.FarmState == nil {
		return nil, &PlanetError{PlanetID: planet.ID, Reason: "no_farm", Extra: "Farm not available"}
	}

	farm := planet.FarmState
	if rowIndex < 0 || rowIndex >= len(farm.Rows) {
		return &FarmActionResult{
			Success:  false,
			Error:    "Invalid row index",
			Rows:     farm.Rows,
			LastTick: farm.LastTick,
		}, nil
	}

	// Check cooldown
	farmActionMu.Lock()
	lastAction, exists := farmLastActionTime[getFarmActionKey(planet.ID)]
	farmActionMu.Unlock()

	if exists && time.Since(lastAction) < FarmActionCooldown {
		remaining := FarmActionCooldown - time.Since(lastAction)
		return &FarmActionResult{
			Success:  false,
			Error:    "Cooldown active. Try again in " + remaining.Round(time.Second).String(),
			Rows:     farm.Rows,
			LastTick: farm.LastTick,
		}, nil
	}

	row := &farm.Rows[rowIndex]
	result := &FarmActionResult{
		Rows:     farm.Rows,
		LastTick: farm.LastTick,
	}

	switch action {
	case "plant":
		if row.Status != FarmRowEmpty {
			result.Error = "Row is not empty"
			return result, nil
		}
		plant := farmPlants[plantType]
		if plant == nil {
			result.Error = "Unknown plant type"
			return result, nil
		}
		row.Status = FarmRowPlanted
		row.PlantType = plantType
		row.Stage = 0
		row.Weeds = 0
		row.WaterTimer = 0
		row.LastTick = 0
		result.Success = true

	case "weed":
		if row.Status == FarmRowEmpty {
			result.Error = "Row is empty"
			return result, nil
		}
		if row.Weeds > 0 {
			row.Weeds--
		}
		result.Success = true

	case "water":
		if row.Status == FarmRowEmpty {
			result.Error = "Row is empty"
			return result, nil
		}
		row.WaterTimer = 10
		result.Success = true

	case "harvest":
		if row.Status != FarmRowMature {
			result.Error = "Plant is not mature"
			return result, nil
		}
		plant := farmPlants[row.PlantType]
		if plant != nil {
			planet.Resources.Food += plant.FoodReward
			result.FoodGain = plant.FoodReward
		}
		// Reset row to empty
		row.Status = FarmRowEmpty
		row.PlantType = ""
		row.Stage = 0
		row.Weeds = 0
		row.WaterTimer = 0
		row.LastTick = 0
		result.Success = true

	default:
		result.Error = "Unknown action: " + action
		return result, nil
	}

	// Update cooldown
	farmActionMu.Lock()
	farmLastActionTime[getFarmActionKey(planet.ID)] = time.Now()
	farmActionMu.Unlock()

	// Save to DB
	SaveFarmToDB(planet)

	return result, nil
}
