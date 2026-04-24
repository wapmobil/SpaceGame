package building

import "math"

// CostMulti represents the multi-resource cost to build the next level.
type CostMulti struct {
	Food  float64 `json:"food"`
	Iron  float64 `json:"iron"`
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
	case "dynamo":
		return float64(level*level+1) * 5 + 10
	case "mine":
		return float64(level*level*2 + 10)
	default:
		return 10
	}
}

// Cost returns the multi-resource cost to build/upgrade a building at the given level.
// level = current building level; returns cost to build/upgrade TO the next level.
// level=0 = first build (to level 1) → food only.
// level=1-2 (upgrade to 2-3) → food + iron.
// level>=3 (upgrade to 4+) → food + iron + money.
func Cost(bt string, level int) CostMulti {
	switch bt {
	case "farm":
		if level == 0 {
			return CostMulti{Food: float64(level*level*level*20 + 30)}
		}
		if level <= 2 {
			return CostMulti{
				Food: float64(level*level*level*20 + 30),
				Iron: float64(level*level*5 + 5),
			}
		}
		return CostMulti{
			Food:  float64(level*level*level*20 + 30),
			Iron:  float64(level*level*5 + 5),
			Money: float64(level * 20),
		}
	case "solar":
		if level == 0 {
			return CostMulti{Food: float64(level*level*50 + 15), Iron: float64(level*level*3 + 3)}
		}
		if level <= 2 {
			return CostMulti{
				Food: float64(level*level*50 + 15),
				Iron: float64(level*level*3 + 3),
			}
		}
		return CostMulti{
			Food:  float64(level*level*50 + 15),
			Iron:  float64(level*level*3 + 3),
			Money: float64(level * 15),
		}
	case "storage":
		if level == 0 {
			return CostMulti{Food: float64(level*level+1) * 60}
		}
		if level <= 2 {
			return CostMulti{
				Food: float64(level*level+1) * 60,
				Iron: float64(level*level+1) * 4,
			}
		}
		return CostMulti{
			Food:  float64(level*level+1) * 60,
			Iron:  float64(level*level+1) * 4,
			Money: float64(level*level+1) * 2,
		}
	case "base":
		if level == 0 {
			return CostMulti{Food: 20, Iron: 40}
		}
		if level == 1 {
			return CostMulti{Food: 200, Iron: 400}
		}
		if level <= 2 {
			return CostMulti{
				Food:  float64(200 * int(math.Pow(2, float64(level-1)))),
				Iron:  float64(400 * int(math.Pow(2, float64(level-1)))),
			}
		}
		return CostMulti{
			Food:  float64(200 * int(math.Pow(2, float64(level-1)))),
			Iron:  float64(400 * int(math.Pow(2, float64(level-1)))),
			Money: math.Pow(2, float64(level-1)),
		}
	case "factory":
		if level == 0 {
			return CostMulti{Food: float64(level*2+1) * 2500}
		}
		if level <= 2 {
			return CostMulti{
				Food: float64(level*2+1) * 2500,
				Iron: float64(level*2+1) * 40,
			}
		}
		return CostMulti{
			Food:  float64(level*2+1) * 2500,
			Iron:  float64(level*2+1) * 40,
			Money: float64(level*2+1) * 20,
		}
	case "energy_storage":
		if level == 0 {
			return CostMulti{Food: float64(level*level) * 300}
		}
		if level <= 2 {
			return CostMulti{
				Food: float64(level*level) * 300,
				Iron: float64(level*level) * 20,
			}
		}
		return CostMulti{
			Food:  float64(level*level) * 300,
			Iron:  float64(level*level) * 20,
			Money: float64(level*level) * 10,
		}
	case "shipyard":
		if level == 0 {
			return CostMulti{Food: math.Pow(2, float64(level+5)) * 0.5}
		}
		if level <= 2 {
			return CostMulti{
				Food: math.Pow(2, float64(level+5)) * 0.5,
				Iron: math.Pow(2, float64(level)),
			}
		}
		return CostMulti{
			Food:  math.Pow(2, float64(level+5)) * 0.5,
			Iron:  math.Pow(2, float64(level)),
			Money: math.Pow(2, float64(level-1)),
		}
	case "comcenter":
		if level == 0 {
			return CostMulti{Food: float64(level) * 10000}
		}
		if level <= 2 {
			return CostMulti{
				Food: float64(level) * 10000,
				Iron: float64(level) * 200,
			}
		}
		return CostMulti{
			Food:  float64(level) * 10000,
			Iron:  float64(level) * 200,
			Money: float64(level) * 100,
		}
	case "composite_drone", "mechanism_factory", "reagent_lab":
		if level == 0 {
			return CostMulti{Food: float64(level*level+1) * 60}
		}
		if level <= 2 {
			return CostMulti{
				Food: float64(level*level+1) * 60,
				Iron: float64(level*level+1) * 4,
			}
		}
		return CostMulti{
			Food:  float64(level*level+1) * 60,
			Iron:  float64(level*level+1) * 4,
			Money: float64(level*level+1) * 2,
		}
	case "dynamo":
		if level == 0 {
			return CostMulti{Food: float64(level*level*30 + 10)}
		}
		if level <= 2 {
			return CostMulti{
				Food: float64(level*level*30 + 10),
				Iron: float64(level*level*3 + 3),
			}
		}
		return CostMulti{
			Food:  float64(level*level*30 + 10),
			Iron:  float64(level*level*3 + 3),
			Money: float64(level * 10),
		}
	case "mine":
		if level == 0 {
			return CostMulti{Food: float64(level*level*30 + 10)}
		}
		if level <= 2 {
			return CostMulti{
				Food: float64(level*level*30 + 10),
				Iron: float64(level*level*5 + 5),
			}
		}
		return CostMulti{
			Food:  float64(level*level*30 + 10),
			Iron:  float64(level*level*5 + 5),
			Money: float64(level * 15),
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
	case "dynamo":
		return float64(level) * -12
	case "mine":
		return float64(level) * 8
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
	case "dynamo":
		prod.Food = -float64(level)
	case "mine":
		prod.Iron = float64(level)
	}
	return prod
}

// NextLevelDeltas returns the delta in production/consumption when upgrading from current level to next level.
// Energy is returned as positive for production, negative for consumption (matching frontend convention).
func NextLevelDeltas(bt string, currentLevel int) ProductionResult {
	nextLevel := currentLevel + 1
	current := Production(bt, currentLevel)
	next := Production(bt, nextLevel)

	currentEnergy := EnergyConsumption(bt, currentLevel)
	nextEnergy := EnergyConsumption(bt, nextLevel)

	return ProductionResult{
		Food:          next.Food - current.Food,
		Iron:          next.Iron - current.Iron,
		Composite:     next.Composite - current.Composite,
		Mechanisms:    next.Mechanisms - current.Mechanisms,
		Reagents:      next.Reagents - current.Reagents,
		Energy:        -nextEnergy - (-currentEnergy), // negate to match frontend convention
		Money:         next.Money - current.Money,
		AlienTech:     next.AlienTech - current.AlienTech,
	}
}
