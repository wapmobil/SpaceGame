package game

import (
	"time"

	"spacegame/internal/game/ship"
)

// Tick processes one game tick (1 second).
func (p *Planet) Tick() {
	p.LastTick = time.Now()

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

	// 4. Advance research progress
	p.Research.Tick()

	// 5. Advance ship construction
	if completed := p.Shipyard.Tick(); completed != nil {
		st := ship.GetShipType(*completed)
		if st != nil {
			p.Fleet.AddShip(st, 1)
		}
	}

	// 6. Advance expeditions
	p.TickExpeditions()

	// 7. Save to DB (throttled)
	if p.game != nil && p.game.shouldSave(p.ID) {
		p.game.savePlanet(p)
	}

	// 8. Broadcast state update
	p.broadcastPlanetUpdate()
}
