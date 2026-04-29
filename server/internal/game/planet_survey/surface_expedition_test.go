package planet_survey

import (
	"testing"
	"time"
)

func TestNewSurfaceExpedition_Duration(t *testing.T) {
	exp := NewSurfaceExpedition("exp-1", "planet-1", "300s", 1)
	if exp.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", exp.Status)
	}
	if exp.Duration != 300 {
		t.Errorf("expected duration 300, got %f", exp.Duration)
	}
	if exp.ElapsedTime != 0 {
		t.Errorf("expected elapsed time 0, got %f", exp.ElapsedTime)
	}
	if exp.Progress != 0 {
		t.Errorf("expected progress 0, got %f", exp.Progress)
	}
}

func TestNewSurfaceExpedition_BaseLevelDuration(t *testing.T) {
	for _, tc := range []struct {
		rangeStr   string
		baseLevel  int
		expectedMin float64
	}{
		{"300s", 1, 300},
		{"300s", 2, 600},
		{"300s", 3, 1200},
		{"600s", 1, 600},
		{"600s", 2, 600},
		{"600s", 3, 1200},
		{"1200s", 1, 1200},
		{"1200s", 2, 1200},
		{"1200s", 3, 1200},
	} {
		exp := NewSurfaceExpedition("exp-test", "planet-1", tc.rangeStr, tc.baseLevel)
		if exp.Duration < tc.expectedMin {
			t.Errorf("range %s baseLevel %d: expected duration >= %f, got %f",
				tc.rangeStr, tc.baseLevel, tc.expectedMin, exp.Duration)
		}
	}
}

func TestNewSurfaceExpedition_CreatedAt(t *testing.T) {
	before := time.Now()
	exp := NewSurfaceExpedition("exp-2", "planet-1", "300s", 1)
	after := time.Now()
	if exp.CreatedAt.Before(before) || exp.CreatedAt.After(after) {
		t.Errorf("expected CreatedAt between %v and %v, got %v", before, after, exp.CreatedAt)
	}
}

func TestTick_ElapsedTimeIncreases(t *testing.T) {
	exp := NewSurfaceExpedition("exp-3", "planet-1", "600s", 1)
	Tick(exp, 10)
	if exp.ElapsedTime != 10 {
		t.Errorf("expected elapsed time 10, got %f", exp.ElapsedTime)
	}
	if exp.Progress != 10.0/600.0 {
		t.Errorf("expected progress %f, got %f", 10.0/600.0, exp.Progress)
	}
}

func TestTick_MultipleTicks(t *testing.T) {
	exp := NewSurfaceExpedition("exp-4", "planet-1", "600s", 1)
	Tick(exp, 100)
	Tick(exp, 200)
	if exp.ElapsedTime != 300 {
		t.Errorf("expected elapsed time 300, got %f", exp.ElapsedTime)
	}
}

func TestIsExpired_NotExpired(t *testing.T) {
	exp := NewSurfaceExpedition("exp-5", "planet-1", "600s", 1)
	Tick(exp, 100)
	if IsExpired(exp) {
		t.Error("expected expedition not to be expired")
	}
}

func TestIsExpired_Expired(t *testing.T) {
	exp := NewSurfaceExpedition("exp-6", "planet-1", "600s", 1)
	Tick(exp, 600)
	if !IsExpired(exp) {
		t.Error("expected expedition to be expired")
	}
}

func TestIsExpired_JustExpired(t *testing.T) {
	exp := NewSurfaceExpedition("exp-7", "planet-1", "300s", 1)
	Tick(exp, 300)
	if !IsExpired(exp) {
		t.Error("expected expedition to be expired at exactly duration")
	}
}

func TestGetCostPerMin(t *testing.T) {
	tests := []struct {
		baseLevel int
		food      float64
		iron      float64
		money     float64
	}{
		{1, 100, 100, 10},
		{2, 200, 200, 20},
		{3, 400, 400, 40},
	}

	for _, tc := range tests {
		cost := GetCostPerMin(tc.baseLevel)
		if cost.Food != tc.food {
			t.Errorf("baseLevel %d: expected food %f, got %f", tc.baseLevel, tc.food, cost.Food)
		}
		if cost.Iron != tc.iron {
			t.Errorf("baseLevel %d: expected iron %f, got %f", tc.baseLevel, tc.iron, cost.Iron)
		}
		if cost.Money != tc.money {
			t.Errorf("baseLevel %d: expected money %f, got %f", tc.baseLevel, tc.money, cost.Money)
		}
	}
}

func TestGetCostPerMin_Default(t *testing.T) {
	cost := GetCostPerMin(5)
	if cost.Food != 100 {
		t.Errorf("expected default cost food 100, got %f", cost.Food)
	}
}

func TestCalculateCost(t *testing.T) {
	food, iron, money := CalculateCost(1, 300)
	expectedFood := 100.0 * 5.0
	expectedIron := 100.0 * 5.0
	expectedMoney := 10.0 * 5.0

	if food != expectedFood {
		t.Errorf("expected food %f, got %f", expectedFood, food)
	}
	if iron != expectedIron {
		t.Errorf("expected iron %f, got %f", expectedIron, iron)
	}
	if money != expectedMoney {
		t.Errorf("expected money %f, got %f", expectedMoney, money)
	}
}

func TestCalculateCost_Level2(t *testing.T) {
	food, iron, money := CalculateCost(2, 600)
	expectedFood := 200.0 * 10.0
	expectedIron := 200.0 * 10.0
	expectedMoney := 20.0 * 10.0

	if food != expectedFood {
		t.Errorf("expected food %f, got %f", expectedFood, food)
	}
	if iron != expectedIron {
		t.Errorf("expected iron %f, got %f", expectedIron, iron)
	}
	if money != expectedMoney {
		t.Errorf("expected money %f, got %f", expectedMoney, money)
	}
}

func TestCalculateDiscoveryChance(t *testing.T) {
	chance0 := CalculateDiscoveryChance(0)
	expected0 := 0.45
	if chance0 != expected0 {
		t.Errorf("expected chance at count=0 to be %f, got %f", expected0, chance0)
	}

	chance3 := CalculateDiscoveryChance(3)
	expected3 := 0.45 * (1.0 / (1.0 + 1.0))
	diff3 := chance3 - expected3
	if diff3 < 0 {
		diff3 = -diff3
	}
	if diff3 > 0.0001 {
		t.Errorf("expected chance at count=3 to be ~%f, got %f (diff=%f)", expected3, chance3, diff3)
	}

	chance6 := CalculateDiscoveryChance(6)
	expected6 := 0.45 * (1.0 / (1.0 + 4.0))
	diff6 := chance6 - expected6
	if diff6 < 0 {
		diff6 = -diff6
	}
	if diff6 > 0.0001 {
		t.Errorf("expected chance at count=6 to be ~%f, got %f (diff=%f)", expected6, chance6, diff6)
	}
}

func TestCalculateDiscoveryChance_Decreasing(t *testing.T) {
	prev := CalculateDiscoveryChance(0)
	for count := 1; count <= 20; count++ {
		curr := CalculateDiscoveryChance(count)
		if curr > prev {
			t.Errorf("expected discovery chance to decrease, count=%d: prev=%f curr=%f", count, prev, curr)
		}
		prev = curr
	}
}

func TestGetResourceChance(t *testing.T) {
	chance0 := GetResourceChance(0)
	if chance0 != 1.0 {
		t.Errorf("expected chance at count=0 to be 1.0, got %f", chance0)
	}

	chance1 := GetResourceChance(1)
	if chance1 != 1.0 {
		t.Errorf("expected chance at count=1 to be 1.0, got %f", chance1)
	}

	chance3 := GetResourceChance(3)
	if chance3 != 1.0/3.0 {
		t.Errorf("expected chance at count=3 to be %f, got %f", 1.0/3.0, chance3)
	}

	chance10 := GetResourceChance(10)
	if chance10 != 0.1 {
		t.Errorf("expected chance at count=10 to be 0.1, got %f", chance10)
	}
}

func TestCalculateResourceRecovery_Success(t *testing.T) {
	recovery := CalculateResourceRecovery(1, 1, true)
	totalResources := 0.0
	for _, amount := range recovery {
		totalResources += amount
	}
	if totalResources > 0 {
		for resName, amount := range recovery {
			maxAmt := 1000.0
			if resName == "money" {
				maxAmt = 250.0
			}
			if amount > maxAmt {
				t.Errorf("resource %s: expected max %f, got %f", resName, maxAmt, amount)
			}
		}
	}
}

func TestCalculateResourceRecovery_Failure(t *testing.T) {
	recovery := CalculateResourceRecovery(1, 1, false)
	totalResources := 0.0
	for _, amount := range recovery {
		totalResources += amount
	}
	if totalResources > 0 {
		for resName, amount := range recovery {
			maxAmt := 1000.0
			if resName == "money" {
				maxAmt = 250.0
			}
			if amount > maxAmt {
				t.Errorf("resource %s: expected max %f, got %f", resName, maxAmt, amount)
			}
		}
	}
}

func TestCalculateResourceRecovery_BaseLevelScaling(t *testing.T) {
	total1 := 0.0
	for range 50 {
		recovery := CalculateResourceRecovery(1, 1, true)
		for _, amount := range recovery {
			total1 += amount
		}
	}
	avg1 := total1 / 50.0

	total3 := 0.0
	for range 50 {
		recovery := CalculateResourceRecovery(3, 1, true)
		for _, amount := range recovery {
			total3 += amount
		}
	}
	avg3 := total3 / 50.0

	if avg3 < avg1 {
		t.Errorf("expected level 3 avg total recovery (%f) >= level 1 avg (%f)", avg3, avg1)
	}
}
