package building

import "math"

// CostMulti represents the multi-resource cost to build the next level.
type CostMulti struct {
	Food  float64 `json:"food"`
	Money float64 `json:"money"`
}

// BuildTime returns the build time in seconds for a building at the given level.
func BuildTime(bt string, level int) float64 {
	switch bt {
	case "farm":
		return float64(level*level*level*2 + 8)
	case "solar":
		return float64(level*level*3 + 7)
	case "storage":
		return float64(level*level+1) * 5
	case "base":
		return math.Pow(2, float64(level+2)) + 6
	case "factory":
		return float64(level*2+1) * 10 + 10
	case "energy_storage":
		return float64(level*level) * 5 + 8
	case "shipyard":
		return math.Pow(2, float64(level+4)) + 10
	case "comcenter":
		return float64(level+1) * 30
	case "composite_drone":
		return float64(level*level+1) * 5
	case "mechanism_factory":
		return float64(level*level+1) * 5
	case "reagent_lab":
		return float64(level*level+1) * 5
	default:
		return 10
	}
}

// Cost returns the multi-resource cost to build/upgrade a building at the given level.
func Cost(bt string, level int) CostMulti {
	switch bt {
	case "farm":
		return CostMulti{
			Food:  float64(level*level*level*20 + 30),
			Money: float64(level*level*level*10 + 15),
		}
	case "solar":
		return CostMulti{
			Food:  float64(level*level*50 + 15),
			Money: float64(level*level*30 + 10),
		}
	case "storage":
		return CostMulti{
			Food:  float64(level*level+1) * 60,
			Money: float64(level*level+1) * 40,
		}
	case "base":
		return CostMulti{
			Food:  math.Pow(2, float64(level+2)),
			Money: math.Pow(2, float64(level+3)),
		}
	case "factory":
		return CostMulti{
			Food:  float64(level*2+1) * 2500,
			Money: float64(level*2+1) * 1500,
		}
	case "energy_storage":
		return CostMulti{
			Food:  float64(level*level) * 300,
			Money: float64(level*level) * 200,
		}
	case "shipyard":
		val := math.Pow(2, float64(level+5)) * 0.5
		return CostMulti{Food: val, Money: val}
	case "comcenter":
		return CostMulti{
			Food:  float64(level) * 10000,
			Money: float64(level) * 10000,
		}
	case "composite_drone", "mechanism_factory", "reagent_lab":
		return CostMulti{
			Food:  float64(level*level+1) * 60,
			Money: float64(level*level+1) * 40,
		}
	default:
		return CostMulti{}
	}
}

// EnergyConsumption returns the energy consumption for a building at the given level.
// Returns negative values for energy-producing buildings (e.g., solar).
func EnergyConsumption(bt string, level int) float64 {
	switch bt {
	case "farm":
		return float64(level) * 10
	case "solar":
		return float64(level) * -15
	case "storage":
		return 0
	case "base":
		return float64(level) * 20
	case "factory":
		return float64(level) * 25
	case "energy_storage":
		return float64(level) * 2
	case "shipyard":
		return float64(level) * 16
	case "comcenter":
		return float64(level) * 100
	case "composite_drone":
		return float64(level) * 10
	case "mechanism_factory":
		return float64(level) * 10
	case "reagent_lab":
		return float64(level) * 10
	default:
		return 0
	}
}

// Production returns the per-tick resource production for a building at the given level.
func Production(bt string, level int) ProductionResult {
	if level <= 0 {
		return ProductionResult{}
	}
	var prod ProductionResult
	switch bt {
	case "farm":
		prod.Food = float64(level)
	case "solar":
		prod.Energy = float64(level) * 15
	case "base":
		prod.Food = -float64(level)
	case "factory":
		prod.Composite = float64(level) * 0.5
	case "composite_drone":
		prod.Composite = float64(level) * 0.5
	case "mechanism_factory":
		prod.Mechanisms = float64(level) * 0.5
	case "reagent_lab":
		prod.Reagents = float64(level) * 0.5
	}
	return prod
}
