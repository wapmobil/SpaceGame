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

	// 2. Energy tick
	p.tickEnergy()

	// 3. Resource production and clamping
	p.tickResources()

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
	if p.game.shouldSave(p.ID) {
		p.game.savePlanet(p)
	}

	// 8. Broadcast state update
	p.broadcastPlanetUpdate()
}
