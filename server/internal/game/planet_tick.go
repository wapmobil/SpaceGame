package game

import (
	"time"

	"spacegame/internal/game/planet_survey"
	"spacegame/internal/game/ship"
)

// Tick processes one game tick (1 second).
func (p *Planet) Tick(gameTick int64) {
	p.LastTick = time.Now()

	// 0. Recalculate build speed from research
	p.RecalculateBuildSpeed()

	// 1. Step building construction progress
	for i := range p.Buildings {
		p.stepBuildingEntry(i)
	}
	p.RecalculateActiveConstruction()
	// Disable buildings that are under construction or ready for confirmation
	for i := range p.Buildings {
		b := &p.Buildings[i]
		if (b.IsBuilding() || b.IsBuildComplete()) && b.Enabled {
			b.Enabled = false
		}
	}

	// 2. Energy tick
	p.tickEnergy()

	// 3. Resource production and clamping
	p.tickResources()
	// 3.5 Auto-disable dynamo when food is depleted
	p.autoDisableDynamo()
	// 3.6 Auto-disable base when food is depleted
	p.autoDisableBase()

	// 3.7 Garden bed tick (every 10 ticks, only if farm is working)
	if p.GardenBedState != nil && p.GetBuildingLevel("farm") > 0 {
		farmIdx := p.FindBuildingIndex("farm")
		if farmIdx >= 0 && farmIdx < len(p.Buildings) && p.Buildings[farmIdx].IsWorking() {
			ProcessGardenBedTick(p, gameTick)
		}
	}

	// 4. Advance research progress (pause if base is not operational)
	if p.HasOperationalBase() {
		p.Research.Tick()
		p.applyResearchEffects()
	}

	// 5. Advance ship construction
	if completed := p.Shipyard.Tick(); completed != nil {
		st := ship.GetShipType(*completed)
		if st != nil {
			p.Fleet.AddShip(st, 1)
		}
	}

	// 6. Advance expeditions
	p.TickExpeditions()

	// 6.5 Tick location buildings
	p.TickLocationBuildings()

	// 7. Save to DB (throttled)
	if p.game != nil && p.game.shouldSave(p.ID) {
		p.game.savePlanet(p)
	}

	// 8. Broadcast state update
	p.broadcastPlanetUpdate()
}

func (p *Planet) TickLocationBuildings() {
	resources := map[string]float64{
		"food":       p.Resources.Food,
		"iron":       p.Resources.Iron,
		"composite":  p.Resources.Composite,
		"mechanisms": p.Resources.Mechanisms,
		"reagents":   p.Resources.Reagents,
		"energy":     p.Resources.Energy,
		"money":      p.Resources.Money,
		"alien_tech": p.Resources.AlienTech,
	}

	for _, loc := range p.Locations {
		if !loc.Active || !loc.BuildingActive {
			continue
		}

		locBuildings := make([]*planet_survey.LocationBuilding, 0)
		for _, lb := range loc.Buildings {
			locBuildings = append(locBuildings, lb)
		}

		planet_survey.TickLocationBuildings(loc, locBuildings, resources)
	}

	p.Resources.Food = resources["food"]
	p.Resources.Iron = resources["iron"]
	p.Resources.Composite = resources["composite"]
	p.Resources.Mechanisms = resources["mechanisms"]
	p.Resources.Reagents = resources["reagents"]
	p.Resources.Energy = resources["energy"]
	p.Resources.Money = resources["money"]
	p.Resources.AlienTech = resources["alien_tech"]
}
