package planet_survey

import (
	"time"
)

type LocationBuilding struct {
	ID            string
	LocationID    string
	BuildingType  string
	Level         int
	Active        bool
	BuildProgress float64
	BuildTime     float64
	CostFood      float64
	CostIron      float64
	CostMoney     float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ResourceMap map[string]float64

func TickLocationBuildings(location *Location, buildings []*LocationBuilding, resources map[string]float64) float64 {
	totalProduction := 0.0

	for _, lb := range buildings {
		if !lb.Active || lb.BuildingType == "" {
			continue
		}

		if location.SourceRemaining <= 0 {
			lb.Active = false
			location.BuildingActive = false
			continue
		}

		def := GetBuildingDef(lb.BuildingType)
		if def == nil {
			continue
		}

		prod := GetProduction(lb.BuildingType, lb.Level)

		resources["food"] += prod.Food
		resources["iron"] += prod.Iron
		resources["composite"] += prod.Composite
		resources["mechanisms"] += prod.Mechanisms
		resources["reagents"] += prod.Reagents
		resources["energy"] += prod.Energy
		resources["money"] += prod.Money
		resources["alien_tech"] += prod.AlienTech

		totalProduction += prod.Food + prod.Iron + prod.Composite + prod.Mechanisms + prod.Reagents + prod.Energy + prod.Money + prod.AlienTech

		sourceConsumption := def.SourceConsumption * float64(lb.Level)
		location.SourceRemaining -= sourceConsumption

		if location.SourceRemaining <= 0 {
			location.SourceRemaining = 0
			lb.Active = false
			location.BuildingActive = false
			location.Active = false
		}
	}

	return totalProduction
}
