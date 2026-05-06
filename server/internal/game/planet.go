package game

import (
	"log"
	"math/rand"
	"time"

	"spacegame/internal/game/expedition"
	"spacegame/internal/game/planet_survey"
	"spacegame/internal/game/research"
	"spacegame/internal/game/ship"
)


type PlanetResourceType string

const (
	ResourceComposite  PlanetResourceType = "composite"
	ResourceMechanisms PlanetResourceType = "mechanisms"
	ResourceReagents   PlanetResourceType = "reagents"
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
	GardenBedState *GardenBedState
	ResourceType       PlanetResourceType
	ExpeditionChains []*planet_survey.ExpeditionChain
	Locations        []*planet_survey.Location
	Description      string
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
		Expeditions:      make([]*expedition.Expedition, 0),
		ExplorationMgr:   expedition.NewExplorationManager(),
		GardenBedState:   nil,
		ExpeditionChains: make([]*planet_survey.ExpeditionChain, 0),
		Locations:        make([]*planet_survey.Location, 0),
		game:             g,
	}
	types := []PlanetResourceType{ResourceComposite, ResourceMechanisms, ResourceReagents}
	p.ResourceType = types[rand.Intn(len(types))]
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
	_ = p.Research.GetLastCompleted()
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

// ExpeditionPlanet interface implementation

func (p *Planet) GetID() string { return p.ID }
func (p *Planet) GetOwnerID() string { return p.OwnerID }
func (p *Planet) GetName() string { return p.Name }
func (p *Planet) GetDescription() string { return p.Description }
func (p *Planet) GetLevel() int { return p.Level }
func (p *Planet) GetResourceType() string { return string(p.ResourceType) }

func (p *Planet) GetResources() map[string]*float64 {
	return map[string]*float64{
		"food":      &p.Resources.Food,
		"iron":      &p.Resources.Iron,
		"composite": &p.Resources.Composite,
		"mechanisms": &p.Resources.Mechanisms,
		"reagents":  &p.Resources.Reagents,
	}
}

func (p *Planet) AddResource(res string, amount float64) {
	switch res {
	case "food":
		p.Resources.Food += amount
	case "iron":
		p.Resources.Iron += amount
	case "composite":
		p.Resources.Composite += amount
	case "mechanisms":
		p.Resources.Mechanisms += amount
	case "reagents":
		p.Resources.Reagents += amount
	}
}

// CanStartExpeditionChain checks if player can start a new expedition chain.
func (p *Planet) CanStartExpeditionChain() bool {
	if !p.BaseOperational() {
		return false
	}
	if _, ok := p.Research.GetCompleted()["planet_exploration"]; !ok {
		return false
	}
	return false
}

// HasActiveOrGeneratingChain checks if there's an active or generating expedition chain.
func (p *Planet) HasActiveOrGeneratingChain() bool {
	for _, ch := range p.ExpeditionChains {
		if ch.Status == "active" || ch.Status == "generating" {
			return true
		}
	}
	return false
}

// CreateExpeditionChain creates a new expedition chain without generating an event.
func (p *Planet) CreateExpeditionChain(inventory map[string]float64) (*planet_survey.ExpeditionChain, error) {
	return planet_survey.CreateExpeditionChain(p, inventory)
}

// GenerateExpeditionEvent generates the first event for a chain via LLM.
func (p *Planet) GenerateExpeditionEvent(chain *planet_survey.ExpeditionChain) (*planet_survey.ExpeditionEvent, error) {
	return planet_survey.GenerateEvent(chain, p)
}

// StartExpeditionChain creates a new expedition chain.
func (p *Planet) StartExpeditionChain(inventory map[string]float64) (*planet_survey.ExpeditionChain, *planet_survey.ExpeditionEvent, error) {
	return planet_survey.StartExpeditionChain(p, inventory)
}

// RecordExpeditionChoice records the player's choice on a chain event.
func (p *Planet) RecordExpeditionChoice(chainID string, choiceIndex int) error {
	chain := p.GetActiveExpeditionChain(chainID)
	if chain == nil {
		return &PlanetError{PlanetID: p.ID, Reason: "chain_not_found"}
	}
	return planet_survey.RecordChoice(chain, choiceIndex)
}

// GenerateNextExpeditionEvent generates the next event after a choice has been recorded.
func (p *Planet) GenerateNextExpeditionEvent(chain *planet_survey.ExpeditionChain) (*planet_survey.ExpeditionEvent, error) {
	return planet_survey.GenerateNextEvent(chain, p)
}

// ResolveExpeditionChoice processes a player choice and generates next event.
func (p *Planet) ResolveExpeditionChoice(chainID string, choiceIndex int) (*planet_survey.ExpeditionEvent, error) {
	chain := p.GetActiveExpeditionChain(chainID)
	if chain == nil {
		return nil, &PlanetError{PlanetID: p.ID, Reason: "chain_not_found"}
	}
	return planet_survey.ResolveChoice(chain, p, choiceIndex)
}

// ReturnInventoryToPlanet returns the expedition inventory back to the planet.
func (p *Planet) ReturnInventoryToPlanetDirect(inventory map[string]float64) {
	for res, amount := range inventory {
		if amount <= 0 {
			continue
		}
		switch res {
		case "food":
			p.Resources.Food += amount
		case "iron":
			p.Resources.Iron += amount
		case "composite":
			p.Resources.Composite += amount
		case "mechanisms":
			p.Resources.Mechanisms += amount
		case "reagents":
			p.Resources.Reagents += amount
		}
	}
}

// GetActiveExpeditionChain returns the active chain by ID.
func (p *Planet) GetActiveExpeditionChain(chainID string) *planet_survey.ExpeditionChain {
	for _, ch := range p.ExpeditionChains {
		if ch.ID == chainID && ch.Status == "active" {
			return ch
		}
	}
	return nil
}

// GetExpeditionChains returns all chains (active + recent completed).
func (p *Planet) GetExpeditionChains() []*planet_survey.ExpeditionChain {
	return p.ExpeditionChains
}

// ReturnExpeditionInventory returns inventory to planet and removes chain from memory.
func (p *Planet) ReturnExpeditionInventory(chainID string) {
	for i, ch := range p.ExpeditionChains {
		if ch.ID == chainID {
			planet_survey.ReturnInventoryToPlanet(p, ch.Inventory)
			p.ExpeditionChains = append(p.ExpeditionChains[:i], p.ExpeditionChains[i+1:]...)
			return
		}
	}
}
