package building

import "testing"

func TestEnergyConsumptionCommandCenter(t *testing.T) {
	tests := []struct {
		level    int
		expected float64
	}{
		{1, 50},
		{3, 150},
	}
	for _, tc := range tests {
		result := EnergyConsumption("command_center", tc.level)
		if result != tc.expected {
			t.Errorf("EnergyConsumption(command_center, %d) = %v, want %v", tc.level, result, tc.expected)
		}
	}
}

func TestProductionCommandCenter(t *testing.T) {
	tests := []struct {
		level    int
		expected float64
	}{
		{1, -10},
		{3, -30},
	}
	for _, tc := range tests {
		result := Production("command_center", tc.level)
		if result.Food != tc.expected {
			t.Errorf("Production(command_center, %d).Food = %v, want %v", tc.level, result.Food, tc.expected)
		}
	}
}
