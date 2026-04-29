package planet_survey

import (
	"math"
	"math/rand"
	"time"
)

type SurfaceExpedition struct {
	ID            string
	PlanetID      string
	Status        string
	Progress      float64
	Duration      float64
	ElapsedTime   float64
	Discovered    *Location
	Range         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ExpeditionRangeStats struct {
	TotalExpeditions int
	LocationsFound   int
}

type CostPerMin struct {
	Food  float64
	Iron  float64
	Money float64
}

var rangeDurations = map[string]float64{
	"300s":  300,
	"600s":  600,
	"1200s": 1200,
}

var baseLevelDurations = map[int]float64{
	1: 300,
	2: 600,
	3: 1200,
}

var baseLevelCosts = map[int]CostPerMin{
	1: {Food: 100, Iron: 100, Money: 10},
	2: {Food: 200, Iron: 200, Money: 20},
	3: {Food: 400, Iron: 400, Money: 40},
}

func NewSurfaceExpedition(id, planetID, rangeStr string, baseLevel int) *SurfaceExpedition {
	duration, ok := rangeDurations[rangeStr]
	if !ok {
		duration = 300
	}
	if lvl, ok := baseLevelDurations[baseLevel]; ok {
		if lvl > duration {
			duration = lvl
		}
	}

	now := time.Now()
	return &SurfaceExpedition{
		ID:          id,
		PlanetID:    planetID,
		Status:      "active",
		Progress:    0,
		Duration:    duration,
		ElapsedTime: 0,
		Range:       rangeStr,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func Tick(exp *SurfaceExpedition, dt float64) {
	exp.ElapsedTime += dt
	exp.Progress = exp.ElapsedTime / exp.Duration
	exp.UpdatedAt = time.Now()
}

func IsExpired(exp *SurfaceExpedition) bool {
	return exp.ElapsedTime >= exp.Duration
}

func GetCostPerMin(baseLevel int) CostPerMin {
	if cost, ok := baseLevelCosts[baseLevel]; ok {
		return cost
	}
	return baseLevelCosts[1]
}

func CalculateCost(baseLevel int, duration float64) (food, iron, money float64) {
	costPerMin := GetCostPerMin(baseLevel)
	minutes := duration / 60.0
	food = costPerMin.Food * minutes
	iron = costPerMin.Iron * minutes
	money = costPerMin.Money * minutes
	return
}

func CalculateDiscoveryChance(count int) float64 {
	if count <= 0 {
		count = 0
	}
	ratio := float64(count) / 3.0
	squared := ratio * ratio
	denominator := 1.0 + squared
	return 0.45 * (1.0 / denominator)
}

func GetResourceChance(count int) float64 {
	if count > 0 {
		return 1.0 / float64(count)
	}
	return 1.0
}

func CalculateResourceRecovery(baseLevel int, count int, isSuccess bool) map[string]float64 {
	resourceChance := GetResourceChance(count)
	if isSuccess {
		resourceChance /= 5.0
	}

	result := make(map[string]float64)
	maxAmounts := map[string]float64{
		"food":       1000,
		"iron":       1000,
		"money":      250,
		"reagents":   100,
		"composite":  100,
		"mechanisms": 100,
	}

	for resName, maxAmt := range maxAmounts {
		if rand.Float64() < resourceChance {
			amount := rand.Float64() * maxAmt * float64(baseLevel)
			result[resName] = math.Round(amount*100) / 100
		} else {
			result[resName] = 0
		}
	}

	return result
}
