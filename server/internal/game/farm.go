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

// Farm row statuses
const (
	FarmRowEmpty     = "empty"
	FarmRowPlanted   = "planted"
	FarmRowMature    = "mature"
	FarmRowWithered  = "withered"
)

// FarmPlant defines a plant type
type FarmPlant struct {
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

var farmPlants = map[string]*FarmPlant{
	PlantWheat:     {Type: PlantWheat, Name: "Пшеница", Icon: "🌾", SeedCost: 5, MoneyReward: 15, FoodReward: 5, UnlockLevel: 1, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 2, WaterCost: 1, GrowthTicks: 60},
	PlantBerries:   {Type: PlantBerries, Name: "Ягоды", Icon: "🫐", SeedCost: 15, MoneyReward: 45, FoodReward: 15, UnlockLevel: 2, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 5, WaterCost: 3, GrowthTicks: 120},
	PlantRaspberry: {Type: PlantRaspberry, Name: "Малина", Icon: "🪴", SeedCost: 25, MoneyReward: 80, FoodReward: 25, UnlockLevel: 3, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 15, WaterCost: 5, GrowthTicks: 180},
	PlantRose:      {Type: PlantRose, Name: "Космическая роза", Icon: "🌷", SeedCost: 60, MoneyReward: 200, FoodReward: 50, UnlockLevel: 5, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 25, WaterCost: 10, GrowthTicks: 300},
	PlantSunflower: {Type: PlantSunflower, Name: "Космический подсолнух", Icon: "🌻", SeedCost: 120, MoneyReward: 400, FoodReward: 80, UnlockLevel: 7, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 20, WaterCost: 30, GrowthTicks: 450},
	PlantMelon:     {Type: PlantMelon, Name: "Космическая дыня", Icon: "🍈", SeedCost: 250, MoneyReward: 800, FoodReward: 120, UnlockLevel: 9, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 30, WaterCost: 20, GrowthTicks: 600},
	PlantBanana:    {Type: PlantBanana, Name: "Лунный банан", Icon: "🌙", SeedCost: 500, MoneyReward: 1700, FoodReward: 150, UnlockLevel: 11, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 50, WaterCost: 50, GrowthTicks: 900},
	PlantBlueberry: {Type: PlantBlueberry, Name: "Звёздная голубика", Icon: "🫐", SeedCost: 1000, MoneyReward: 3500, FoodReward: 300, UnlockLevel: 13, Stages: 3, StageNames: []string{"Семя", "Росток", "Созрело"}, WeedCost: 80, WaterCost: 50, GrowthTicks: 1500},
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
	WitherTimer      int     `json:"wither_timer,omitempty"`
	StageProgress    int     `json:"-"` // internal: progress within current stage (0 or 1)
	TicksToMature    int     `json:"ticks_to_mature,omitempty"`
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
			Status:        FarmRowEmpty,
			Weeds:         0,
			WaterTimer:    0,
			Stage:         0,
			WitherTimer:   0,
			StageProgress: 0,
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

		// --- Empty rows: weed growth ---
		if row.Status == FarmRowEmpty {
			weedChance := 0.05 // 5% for empty rows
			if row.WaterTimer > 0 {
				weedChance = 0.06 // 6% for watered empty rows
			}
			if rand.Float64() < weedChance && row.Weeds < 3 {
				row.Weeds++
				changed = true
			}
			row.LastTick = farmTickNum
			continue
		} else if row.Status == FarmRowWithered {
			// --- Withered rows: no processing, just tick ---
			row.LastTick = farmTickNum
			continue
		} else if row.Status == FarmRowPlanted {
			plant := farmPlants[row.PlantType]
			if plant == nil {
				row.Status = FarmRowEmpty
				row.PlantType = ""
				row.Stage = 0
				row.Weeds = 0
				row.WaterTimer = 0
				row.LastTick = farmTickNum
				row.FarmTicksSinceLast = 0
				row.StageProgress = 0
				changed = true
				continue
			}

			// Skip if already processed this tick
			if row.LastTick == farmTickNum {
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
					row.StageProgress++
					if row.StageProgress >= 2 {
						advanceStage = true
						row.StageProgress = 0
					}
				}

				if advanceStage {
					newStage := row.Stage + 1
					if newStage >= plant.Stages-1 {
						newStage = plant.Stages - 1
						row.Status = FarmRowMature
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
					row.TicksToMature = remainingStages
				} else {
					row.TicksToMature = remainingStages*2 - row.StageProgress
				}
			}

			row.LastTick = farmTickNum
			continue
		} else if row.Status == FarmRowMature {
			// --- Mature rows: wither check ---
			plant := farmPlants[row.PlantType]
			if plant == nil {
				row.Status = FarmRowEmpty
				row.PlantType = ""
				row.Stage = 0
				row.Weeds = 0
				row.WaterTimer = 0
				row.LastTick = farmTickNum
				row.FarmTicksSinceLast = 0
				row.WitherTimer = 0
				row.StageProgress = 0
				row.TicksToMature = 0
				changed = true
				continue
			}

			// Calculate wither time based on water status
			witherTicks := int64(30) // default 30 ticks (5 min)
			if row.WaterTimer > 0 {
				witherTicks = 50 // watered: 50 ticks (8.3 min)
			}
			if row.Weeds >= 1 {
				if witherTicks > 15 {
					witherTicks = 15
				}
			}
			if row.Weeds >= 2 {
				if witherTicks > 10 {
					witherTicks = 10
				}
			}
			if row.Weeds >= 3 {
				if witherTicks > 5 {
					witherTicks = 5
				}
			}

			// Increment wither timer
			row.WitherTimer++
			if row.WitherTimer >= int(witherTicks) {
				row.Status = FarmRowWithered
				changed = true
			}

			// TicksToMature = remaining wither time
			remainingWither := int(witherTicks) - row.WitherTimer
			if remainingWither < 0 {
				remainingWither = 0
			}
			row.TicksToMature = remainingWither

			row.LastTick = farmTickNum
			continue
		}
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
	Success      bool      `json:"success"`
	Error        string    `json:"error,omitempty"`
	Rows         []FarmRow `json:"rows"`
	LastTick     int64     `json:"last_tick"`
	FoodGain     float64   `json:"food_gain,omitempty"`
	MoneyGain    float64   `json:"money_gain,omitempty"`
	FoodCost     float64   `json:"food_cost,omitempty"`
	SeedCost     float64   `json:"seed_cost,omitempty"`
	UnlockLevel  int       `json:"unlock_level,omitempty"`
	WitherTimer  int       `json:"wither_timer,omitempty"`
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
		row.Status = FarmRowPlanted
		row.PlantType = plantType
		row.Stage = 0
		row.Weeds = 0
		row.WaterTimer = 0
		row.WitherTimer = 0
		row.LastTick = 0
		row.FarmTicksSinceLast = 0
		row.StageProgress = 0
		row.TicksToMature = 0
		result.Success = true

	case "weed":
		if row.Status == FarmRowEmpty {
			result.Error = "Row is empty"
			return result, nil
		}
		if row.Weeds <= 0 {
			result.Error = "No weeds to remove"
			return result, nil
		}
		// Get weed cost
		weedCost := 2.0
		if row.Status == FarmRowPlanted || row.Status == FarmRowMature || row.Status == FarmRowWithered {
			plant := farmPlants[row.PlantType]
			if plant != nil {
				weedCost = plant.WeedCost
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
		if row.Status == FarmRowEmpty {
			result.Error = "Row is empty"
			return result, nil
		}
		// Get water cost
		waterCost := 1.0
		if row.Status == FarmRowPlanted || row.Status == FarmRowMature || row.Status == FarmRowWithered {
			plant := farmPlants[row.PlantType]
			if plant != nil {
				waterCost = plant.WaterCost
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
			planet.Resources.Money += plant.MoneyReward
			result.FoodGain = plant.FoodReward
			result.MoneyGain = plant.MoneyReward
		}
		// Reset row to empty
		row.Status = FarmRowEmpty
		row.PlantType = ""
		row.Stage = 0
		row.Weeds = 0
		row.WaterTimer = 0
		row.WitherTimer = 0
		row.LastTick = 0
		row.FarmTicksSinceLast = 0
		row.StageProgress = 0
		row.TicksToMature = 0
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

func itoa(n int) string {
	return strconv.Itoa(n)
}

func itoaF(f float64) string {
	if f == float64(int(f)) {
		return strconv.Itoa(int(f))
	}
	return strconv.FormatFloat(f, 'f', 1, 64)
}
