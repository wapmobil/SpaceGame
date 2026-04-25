package game

import (
	"log"
	"math/rand"
	"time"

	"spacegame/internal/game/expedition"
	"spacegame/internal/game/research"
	"spacegame/internal/game/ship"
)


// Planet manages a single planet's game state.
type Planet struct {
	ID               string
	OwnerID          string
	Name             string
	Level            int
	Buildings        []BuildingEntry
	ActiveConstruction int
	Resources          PlanetResources
	BuildSpeed         float64
	LastTick         time.Time
	EnergyBalance    float64
	EnergyBuffer     EnergyBuffer
	Research         *research.ResearchSystem
	Fleet            *ship.Fleet
	Shipyard         *ship.Shipyard
	Expeditions      []*expedition.Expedition
	ExplorationMgr   *expedition.ExplorationManager
	GardenBedState   *GardenBedState
	GardenBedLastTick int64
	game             *Game
}

// NewPlanet creates a new planet with default resources.
func NewPlanet(id, ownerID, name string, g *Game) *Planet {
	p := &Planet{
		ID:               id,
		OwnerID:          ownerID,
		Name:             name,
		Level:            1,
		Buildings:        make([]BuildingEntry, 0),
		Resources: PlanetResources{
			Food:      80,
			Iron:      5,
			Composite: 0,
			Mechanisms: 0,
			Reagents:  0,
			Energy:    20,
			MaxEnergy: 20,
			Money:     500,
			AlienTech: 0,
		},
		BuildSpeed:     1.0,
		LastTick:       time.Now(),
		Fleet:          ship.NewFleet(),
		Shipyard:       ship.NewShipyard(),
		Expeditions:    make([]*expedition.Expedition, 0),
		ExplorationMgr: expedition.NewExplorationManager(),
		GardenBedState:   nil,
		GardenBedLastTick: 0,
		game:           g,
	}
	p.EnergyBuffer = NewEnergyBuffer()
	if g != nil {
		p.Research = research.NewResearchSystem(id, g.db)
		if err := p.Research.LoadFromDB(); err != nil {
			log.Printf("Error loading research for planet %s: %v", id, err)
		}
	} else {
		p.Research = research.NewResearchSystem(id, nil)
	}
	return p
}

// StartResearch begins researching a technology on this planet.
func (p *Planet) StartResearch(techID string) error {
	if !p.BaseOperational() {
		return &PlanetError{PlanetID: p.ID, Reason: "base_not_operational", Extra: "Planet base requires food to operate. Produce food to unlock research."}
	}
	tech := research.GetTechByID(techID)
	if tech == nil {
		return &PlanetError{PlanetID: p.ID, Reason: "tech_not_found"}
	}
	return p.Research.StartResearch(tech, &p.Resources.Food, &p.Resources.Money, &p.Resources.AlienTech)
}

// GetAvailableResearch returns technologies that can be researched on this planet.
func (p *Planet) GetAvailableResearch() ([]byte, error) {
	return p.Research.GetAvailableForAPI()
}

// GetResearchJSON returns all research state as JSON.
func (p *Planet) GetResearchJSON() ([]byte, error) {
	return p.Research.GetResearchJSON()
}

// GetResearchErrors returns the completed tech map.
func (p *Planet) GetResearchCompleted() map[string]int {
	return p.Research.GetCompleted()
}

// RecalculateBuildSpeed updates BuildSpeed based on fast_construction research level.
func (p *Planet) RecalculateBuildSpeed() {
	speed := 1.0
	if lvl, ok := p.Research.GetCompleted()["fast_construction"]; ok && lvl > 0 {
		speed += float64(lvl) * 0.2
	}
	p.BuildSpeed = speed
}

// applyResearchEffects applies effects for newly completed research.
func (p *Planet) applyResearchEffects() {
	for techID := range p.Research.GetLastCompleted() {
		switch techID {
		case "planet_exploration":
			buildings := []string{"composite_drone", "mechanism_factory", "reagent_lab"}
			idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(buildings))
			p.Resources.ResearchUnlocks = buildings[idx]
		}
	}
}

// CanBuildShip checks if the planet can build a ship of the given type.
func (p *Planet) CanBuildShip(typeID ship.TypeID) bool {
	st := ship.GetShipType(typeID)
	if st == nil {
		return false
	}

	shipyardLevel := p.GetBuildingLevel("shipyard")
	if st.MinShipyard > shipyardLevel {
		return false
	}

	if !st.Cost.CanAfford(p.Resources.Food, p.Resources.Composite, p.Resources.Mechanisms, p.Resources.Reagents, p.Resources.Money) {
		return false
	}

	maxSlots := p.Shipyard.MaxSlots(p.GetBuildingLevel("base"))
	return p.Fleet.CanAddShip(st, 1, maxSlots)
}

// BuildShip queues a ship for construction. Returns error if can't build.
func (p *Planet) BuildShip(typeID ship.TypeID) error {
	st := ship.GetShipType(typeID)
	if st == nil {
		return &PlanetError{PlanetID: p.ID, Reason: "unknown_ship_type"}
	}

	if !p.CanBuildShip(typeID) {
		return &PlanetError{PlanetID: p.ID, Reason: "cannot_build"}
	}

	if err := p.Shipyard.QueueShip(st); err != nil {
		return &PlanetError{PlanetID: p.ID, Reason: "queue_failed"}
	}

	p.Shipyard.DeductCost(st, &p.Resources.Food, &p.Resources.Composite, &p.Resources.Mechanisms, &p.Resources.Reagents, &p.Resources.Money)

	return nil
}

// GetFleet returns the planet's fleet.
func (p *Planet) GetFleet() *ship.Fleet {
	return p.Fleet
}

// GetShipyard returns the planet's shipyard.
func (p *Planet) GetShipyard() *ship.Shipyard {
	return p.Shipyard
}

// GetTotalShipCount returns the total number of ships in the fleet.
func (p *Planet) GetTotalShipCount() int {
	return p.Fleet.TotalShipCount()
}

// PlanetError represents an error that occurred during a planet operation.
type PlanetError struct {
	PlanetID string
	Reason   string
	Extra    string
}

func (e *PlanetError) Error() string {
	if e.Extra != "" {
		return "planet error: " + e.Reason + " - " + e.Extra + " (planet: " + e.PlanetID + ")"
	}
	return "planet error: " + e.Reason + " (planet: " + e.PlanetID + ")"
}
