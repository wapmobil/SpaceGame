package game

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// Plant types
const (
	PlantWheat     = "wheat"
	PlantBerries   = "berries"
	PlantRaspberry = "raspberry"
	PlantRose      = "rose"
	PlantSunflower = "sunflower"
	PlantMelon     = "melon"
	PlantBanana    = "banana"
	PlantBlueberry = "blueberry"
)

// Garden bed row statuses
const (
	GardenBedRowEmpty     = "empty"
	GardenBedRowPlanted   = "planted"
	GardenBedRowMature    = "mature"
	GardenBedRowWithered  = "withered"
)

// GardenBedPlant defines a plant type
type GardenBedPlant struct {
	Type        string
	Name        string
	Icon        string
	SeedCost    float64
	MoneyReward float64
	FoodReward  float64
	UnlockLevel int
	Stages      int
	StageNames  []string
	WeedCost    float64
	WaterCost   float64
	GrowthTicks int64
}

var gardenBedPlants = map[string]*GardenBedPlant{
	PlantWheat:     {Type: PlantWheat, Name: "Пшеница", Icon: "🌾", SeedCost: 5, MoneyReward: 15, FoodReward: 5, UnlockLevel: 1, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 2, WaterCost: 1, GrowthTicks: 60},
	PlantBerries:   {Type: PlantBerries, Name: "Ягоды", Icon: "🫐", SeedCost: 15, MoneyReward: 45, FoodReward: 15, UnlockLevel: 2, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 5, WaterCost: 3, GrowthTicks: 120},
	PlantRaspberry: {Type: PlantRaspberry, Name: "Малина", Icon: "🪴", SeedCost: 25, MoneyReward: 80, FoodReward: 25, UnlockLevel: 3, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 15, WaterCost: 5, GrowthTicks: 180},
	PlantRose:      {Type: PlantRose, Name: "Космическая роза", Icon: "🌷", SeedCost: 60, MoneyReward: 200, FoodReward: 50, UnlockLevel: 5, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 25, WaterCost: 10, GrowthTicks: 300},
	PlantSunflower: {Type: PlantSunflower, Name: "Космический подсолнух", Icon: "🌻", SeedCost: 120, MoneyReward: 400, FoodReward: 80, UnlockLevel: 7, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 20, WaterCost: 30, GrowthTicks: 450},
	PlantMelon:     {Type: PlantMelon, Name: "Космическая дыня", Icon: "🍈", SeedCost: 250, MoneyReward: 800, FoodReward: 120, UnlockLevel: 9, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 30, WaterCost: 20, GrowthTicks: 600},
	PlantBanana:    {Type: PlantBanana, Name: "Лунный банан", Icon: "🌙", SeedCost: 500, MoneyReward: 1700, FoodReward: 150, UnlockLevel: 11, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 50, WaterCost: 50, GrowthTicks: 900},
	PlantBlueberry: {Type: PlantBlueberry, Name: "Звёздная голубика", Icon: "🫐", SeedCost: 1000, MoneyReward: 3500, FoodReward: 300, UnlockLevel: 13, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 80, WaterCost: 50, GrowthTicks: 1500},
}

// GetGardenBedPlant returns a garden bed plant by type, or nil if unknown
func GetGardenBedPlant(typeID string) *GardenBedPlant {
	return gardenBedPlants[typeID]
}

// GetAllGardenBedPlants returns all garden bed plant definitions
func GetAllGardenBedPlants() []*GardenBedPlant {
	result := make([]*GardenBedPlant, 0, len(gardenBedPlants))
	for _, p := range gardenBedPlants {
		result = append(result, p)
	}
	return result
}

// GardenBedRow represents a single row in the garden bed grid
type GardenBedRow struct {
	Status           string  `json:"status"`
	PlantType        string  `json:"plant_type,omitempty"`
	Stage            int     `json:"stage,omitempty"`
	Weeds            int     `json:"weeds"`
	WaterTimer       int     `json:"water_timer"`
	LastTick         int64   // internal: dedup within a tick call
	GardenBedTicksSinceLast int   `json:"-"` // internal: garden bed ticks since last growth check
	WitherTimer      int     `json:"wither_timer,omitempty"`
	StageProgress    int     `json:"stage_progress"` // internal: progress within current stage
	TicksToMature    int     `json:"ticks_to_mature,omitempty"`
}

// GardenBedState represents the complete garden bed state for a planet
type GardenBedState struct {
	Rows     []GardenBedRow `json:"rows"`
	RowCount int            `json:"row_count"`
}

// NewGardenBedState creates a new empty garden bed with the given row count
func NewGardenBedState(rowCount int) *GardenBedState {
	rows := make([]GardenBedRow, rowCount)
	for i := range rows {
		rows[i] = GardenBedRow{
			Status:        GardenBedRowEmpty,
			Weeds:         0,
			WaterTimer:    0,
			Stage:         0,
			WitherTimer:   0,
			StageProgress: 0,
		}
	}
	return &GardenBedState{
		Rows:     rows,
		RowCount: rowCount,
	}
}

// NewGardenBedStateFromJSON deserializes garden bed state from JSONB data
func NewGardenBedStateFromJSON(data []byte, rowCount int) *GardenBedState {
	if len(data) == 0 || string(data) == "null" || string(data) == "[]" {
		return NewGardenBedState(rowCount)
	}

	// Try parsing as GardenBedState first
	var state GardenBedState
	if err := json.Unmarshal(data, &state); err != nil {
		// Try parsing as []GardenBedRow
		var rows []GardenBedRow
		if err2 := json.Unmarshal(data, &rows); err2 != nil {
			return NewGardenBedState(rowCount)
		}
		state.Rows = rows
		state.RowCount = len(rows)
	}

	// Normalize legacy data: empty string status -> empty
	for i := range state.Rows {
		if state.Rows[i].Status == "" {
			state.Rows[i].Status = GardenBedRowEmpty
		}
		if state.Rows[i].WitherTimer == 0 {
			state.Rows[i].WitherTimer = 0
		}
	}

	// Ensure row count matches
	if len(state.Rows) != state.RowCount {
		state.RowCount = len(state.Rows)
	}

	return &state
}

// GardenBedTick processes one garden bed tick (every gardenBedTickInterval game ticks)
// gardenBedTickNum is the garden bed tick number (gameTick / gardenBedTickInterval)
// Returns true if the state changed
func GardenBedTick(gb *GardenBedState, gardenBedTickNum int64) bool {
	if gb == nil || len(gb.Rows) == 0 {
		return false
	}

	changed := false

	for i := range gb.Rows {
		row := &gb.Rows[i]

		// --- Empty rows: weed growth ---
		if row.Status == GardenBedRowEmpty {
			weedChance := 0.02 // 2% for empty rows
			if rand.Float64() < weedChance && row.Weeds < 3 {
				row.Weeds++
				changed = true
			}
			row.LastTick = gardenBedTickNum
			continue
		} else if row.Status == GardenBedRowWithered {
			// --- Withered rows: no processing, just tick ---
			row.LastTick = gardenBedTickNum
			continue
		} else if row.Status == GardenBedRowPlanted {
			plant := gardenBedPlants[row.PlantType]
			if plant == nil {
				row.Status = GardenBedRowEmpty
				row.PlantType = ""
				row.Stage = 0
				row.Weeds = 0
				row.WaterTimer = 0
				row.LastTick = gardenBedTickNum
				row.GardenBedTicksSinceLast = 0
				row.StageProgress = 0
				changed = true
				continue
			}

			// Skip if already processed this tick
			if row.LastTick == gardenBedTickNum {
				continue
			}

			// Check if watered BEFORE decrementing
			isWatered := row.WaterTimer > 0

			// Weed spawn: 3% without water, 5% with water, up to 3 weeds
			weedChance := 0.03
			if isWatered {
				weedChance = 0.05
			}
			if rand.Float64() < weedChance && row.Weeds < 3 {
				row.Weeds++
				changed = true
			}

			// Water timer decrements
			if row.WaterTimer > 0 {
				row.WaterTimer--
				changed = true
			}

			// Growth: only if weeds < 3 (not fully blocked)
			if row.Weeds < 3 {
				advanceStage := false

				if isWatered {
					// Watered: advance 1 stage every 10 ticks
					row.StageProgress++
					if row.StageProgress >= 10 {
						advanceStage = true
						row.StageProgress = 0
					}
				} else {
					// Normal: advance 1 stage every 20 ticks
					row.StageProgress++
					if row.StageProgress >= 20 {
						advanceStage = true
						row.StageProgress = 0
					}
				}

				if advanceStage {
					newStage := row.Stage + 1
					if newStage >= plant.Stages-1 {
						newStage = plant.Stages - 1
						row.Status = GardenBedRowMature
						row.WitherTimer = 0
						row.StageProgress = 0
					}
					row.Stage = newStage
					changed = true
				}
			}

			// Calculate ticks to mature
			if row.Weeds >= 3 {
				row.TicksToMature = -1 // blocked
			} else {
				remainingStages := (plant.Stages - 1) - row.Stage
				if remainingStages <= 0 {
					row.TicksToMature = 0
				} else if isWatered {
					row.TicksToMature = remainingStages*10 - row.StageProgress
				} else {
					row.TicksToMature = remainingStages*20 - row.StageProgress
				}
			}

			row.LastTick = gardenBedTickNum
			continue
		} else if row.Status == GardenBedRowMature {
			// --- Mature rows: wither check ---
			plant := gardenBedPlants[row.PlantType]
			if plant == nil {
				row.Status = GardenBedRowEmpty
				row.PlantType = ""
				row.Stage = 0
				row.Weeds = 0
				row.WaterTimer = 0
				row.LastTick = gardenBedTickNum
				row.GardenBedTicksSinceLast = 0
				row.WitherTimer = 0
				row.StageProgress = 0
				row.TicksToMature = 0
				changed = true
				continue
			}

			// Calculate wither time based on water status
			witherTicks := int64(300) // default 300 ticks (50 min)
			if row.WaterTimer > 0 {
				witherTicks = 500 // watered: 500 ticks (83 min)
			}
			if row.Weeds >= 1 {
				if witherTicks > 150 {
					witherTicks = 150
				}
			}
			if row.Weeds >= 2 {
				if witherTicks > 100 {
					witherTicks = 100
				}
			}
			if row.Weeds >= 3 {
				if witherTicks > 50 {
					witherTicks = 50
				}
			}

			// Increment wither timer
			row.WitherTimer++
			if row.WitherTimer >= int(witherTicks) {
				row.Status = GardenBedRowWithered
				changed = true
			}

			// TicksToMature = remaining wither time
			remainingWither := int(witherTicks) - row.WitherTimer
			if remainingWither < 0 {
				remainingWither = 0
			}
			row.TicksToMature = remainingWither

			row.LastTick = gardenBedTickNum
			continue
		}
	}

	return changed
}

// ProcessGardenBedTick is the main entry point called from planet_tick.go
// Only processes when gameTick % gardenBedTickInterval == 0 (once every N game ticks)
func ProcessGardenBedTick(planet *Planet, gameTick int64) {
	if planet == nil || planet.GardenBedState == nil {
		return
	}

	const gardenBedTickInterval = 10

	// Only process every N ticks
	if gameTick%gardenBedTickInterval != 0 {
		return
	}

	gardenBedTickNum := gameTick / gardenBedTickInterval
	GardenBedTick(planet.GardenBedState, gardenBedTickNum)

	// Save garden bed state to DB
	SaveGardenBedToDB(planet)
}

// GardenBedActionResult is the result of a garden bed action
type GardenBedActionResult struct {
	Success      bool           `json:"success"`
	Error        string         `json:"error,omitempty"`
	Rows         []GardenBedRow `json:"rows"`
	FoodGain     float64        `json:"food_gain,omitempty"`
	MoneyGain    float64        `json:"money_gain,omitempty"`
	FoodCost     float64        `json:"food_cost,omitempty"`
	SeedCost     float64        `json:"seed_cost,omitempty"`
	UnlockLevel  int            `json:"unlock_level,omitempty"`
	WitherTimer  int            `json:"wither_timer,omitempty"`
	CooldownEnd  int64          `json:"cooldown_end"`
}

// GardenBedActionCooldown is the cooldown between garden bed actions in seconds
const GardenBedActionCooldown = 1 * time.Second

// gardenBedLastActionTime stores the last action time per planet
var gardenBedLastActionTime = make(map[string]time.Time)
var gardenBedActionMu = &sync.Mutex{}

// getGardenBedActionKey returns the cache key for a planet's garden bed action cooldown
func getGardenBedActionKey(planetID string) string {
	return "garden_bed:" + planetID
}

// GetGardenBedCooldownEnd returns the Unix timestamp (seconds) when the cooldown expires,
// or 0 if no cooldown is active.
func GetGardenBedCooldownEnd(planetID string) int64 {
	gardenBedActionMu.Lock()
	defer gardenBedActionMu.Unlock()
	lastAction, exists := gardenBedLastActionTime[getGardenBedActionKey(planetID)]
	if !exists {
		return 0
	}
	remaining := GardenBedActionCooldown - time.Since(lastAction)
	if remaining <= 0 {
		return 0
	}
	return time.Now().Add(remaining).Unix()
}

// ClearGardenBedCooldown clears the cooldown for a planet (for testing)
func ClearGardenBedCooldown(planetID string) {
	gardenBedActionMu.Lock()
	delete(gardenBedLastActionTime, getGardenBedActionKey(planetID))
	gardenBedActionMu.Unlock()
}

// gardenBedAction handles player garden bed actions with cooldown enforcement
// Returns the result as JSON bytes
func gardenBedAction(planet *Planet, action string, rowIndex int, plantType string) (*GardenBedActionResult, error) {
	if planet == nil || planet.GardenBedState == nil {
		return nil, &PlanetError{PlanetID: planet.ID, Reason: "no_garden_bed", Extra: "Garden beds not available"}
	}

	gb := planet.GardenBedState
	if rowIndex < 0 || rowIndex >= len(gb.Rows) {
		return &GardenBedActionResult{
			Success: false,
			Error:   "Invalid row index",
			Rows:    gb.Rows,
		}, nil
	}

	// Check cooldown
	gardenBedActionMu.Lock()
	lastAction, exists := gardenBedLastActionTime[getGardenBedActionKey(planet.ID)]
	gardenBedActionMu.Unlock()

	if exists && time.Since(lastAction) < GardenBedActionCooldown {
		remaining := GardenBedActionCooldown - time.Since(lastAction)
		cooldownEnd := time.Now().Add(remaining).Unix()
		return &GardenBedActionResult{
			Success:     false,
			Error:       "Cooldown active. Try again in " + remaining.Round(time.Second).String(),
			Rows:        gb.Rows,
			CooldownEnd: cooldownEnd,
		}, nil
	}

	row := &gb.Rows[rowIndex]
	result := &GardenBedActionResult{
		Rows: gb.Rows,
	}

	switch action {
	case "plant":
		if row.Status != GardenBedRowEmpty {
			result.Error = "Row is not empty"
			return result, nil
		}
		plant := gardenBedPlants[plantType]
		if plant == nil {
			result.Error = "Unknown plant type"
			return result, nil
		}
		// Check unlock level
		farmLevel := planet.GetBuildingLevel("farm")
		if farmLevel < plant.UnlockLevel {
			result.Error = "Requires farm level " + itoa(plant.UnlockLevel)
			result.UnlockLevel = plant.UnlockLevel
			return result, nil
		}
		// Check money
		if planet.Resources.Money < plant.SeedCost {
			result.Error = "Not enough money. Need " + itoaF(plant.SeedCost) + "💰, have " + itoaF(planet.Resources.Money) + "💰"
			result.SeedCost = plant.SeedCost
			return result, nil
		}
		// Deduct seed cost
		planet.Resources.Money -= plant.SeedCost
		result.SeedCost = plant.SeedCost
		row.Status = GardenBedRowPlanted
		row.PlantType = plantType
		row.Stage = 0
		row.Weeds = 0
		row.WaterTimer = 0
		row.WitherTimer = 0
		row.LastTick = 0
		row.GardenBedTicksSinceLast = 0
		row.StageProgress = 0
		row.TicksToMature = 0
		result.Success = true

	case "weed":
		if row.Weeds <= 0 {
			result.Error = "No weeds to remove"
			return result, nil
		}
		farmLevel := planet.GetBuildingLevel("farm")
		// Get weed cost
		weedCost := float64(farmLevel) * 10.0
		if row.Status == GardenBedRowPlanted || row.Status == GardenBedRowMature || row.Status == GardenBedRowWithered {
			plant := gardenBedPlants[row.PlantType]
			if plant != nil {
				weedCost = plant.WeedCost * float64(farmLevel) * 10
			}
		}
		// Check food
		if planet.Resources.Food < weedCost {
			result.Error = "Not enough food. Need " + itoaF(weedCost) + "🍍, have " + itoaF(planet.Resources.Food) + "🍍"
			result.FoodCost = weedCost
			return result, nil
		}
		// Deduct food and remove weed
		planet.Resources.Food -= weedCost
		result.FoodCost = weedCost
		row.Weeds--
		result.Success = true

	case "water":
		if row.Status == GardenBedRowEmpty {
			result.Error = "Row is empty"
			return result, nil
		}
		farmLevel := planet.GetBuildingLevel("farm")
		// Get water cost
		waterCost := float64(farmLevel) * 10.0
		if row.Status == GardenBedRowPlanted || row.Status == GardenBedRowMature || row.Status == GardenBedRowWithered {
			plant := gardenBedPlants[row.PlantType]
			if plant != nil {
				waterCost = plant.WaterCost * float64(farmLevel) * 10
			}
		}
		// Check food
		if planet.Resources.Food < waterCost {
			result.Error = "Not enough food. Need " + itoaF(waterCost) + "🍍, have " + itoaF(planet.Resources.Food) + "🍍"
			result.FoodCost = waterCost
			return result, nil
		}
		// Deduct food and set water timer
		planet.Resources.Food -= waterCost
		result.FoodCost = waterCost
		row.WaterTimer = 100
		result.Success = true

	case "harvest":
		if row.Status != GardenBedRowMature {
			result.Error = "Plant is not mature"
			return result, nil
		}
		plant := gardenBedPlants[row.PlantType]
		if plant != nil {
			planet.Resources.Food += plant.FoodReward
			planet.Resources.Money += plant.MoneyReward
			result.FoodGain = plant.FoodReward
			result.MoneyGain = plant.MoneyReward
		}
		// Reset row to empty
		row.Status = GardenBedRowEmpty
		row.PlantType = ""
		row.Stage = 0
		row.Weeds = 0
		row.WaterTimer = 0
		row.WitherTimer = 0
		row.LastTick = 0
		row.GardenBedTicksSinceLast = 0
		row.StageProgress = 0
		row.TicksToMature = 0
		result.Success = true

	case "clear":
		if row.Status != GardenBedRowWithered {
			result.Error = "Only withered plants can be cleared"
			return result, nil
		}
		farmLevel := planet.GetBuildingLevel("farm")
		clearCost := float64(farmLevel) * 10.0
		if planet.Resources.Food < clearCost {
			result.Error = "Not enough food. Need " + itoaF(clearCost) + "🍍, have " + itoaF(planet.Resources.Food) + "🍍"
			result.FoodCost = clearCost
			return result, nil
		}
		planet.Resources.Food -= clearCost
		result.FoodCost = clearCost
		row.Status = GardenBedRowEmpty
		row.PlantType = ""
		row.Stage = 0
		row.Weeds = 0
		row.WaterTimer = 0
		row.WitherTimer = 0
		row.LastTick = 0
		row.GardenBedTicksSinceLast = 0
		row.StageProgress = 0
		row.TicksToMature = 0
		result.Success = true

	default:
		result.Error = "Unknown action: " + action
		return result, nil
	}

	// Update cooldown
	gardenBedActionMu.Lock()
	gardenBedLastActionTime[getGardenBedActionKey(planet.ID)] = time.Now()
	gardenBedActionMu.Unlock()

	result.CooldownEnd = time.Now().Add(GardenBedActionCooldown).Unix()

	// Save to DB
	SaveGardenBedToDB(planet)

	return result, nil
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func itoaF(f float64) string {
	if f == float64(int(f)) {
		return strconv.Itoa(int(f))
	}
	return strconv.FormatFloat(f, 'f', 1, 64)
}
