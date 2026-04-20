package game

import (
	"encoding/json"
	"log"
	"math"
	"time"

	"spacegame/internal/game/research"
	"spacegame/internal/game/ship"
)



// PlanetResources holds the current resource amounts.
type PlanetResources struct {
	Food      float64 `json:"food"`
	Composite float64 `json:"composite"`
	Mechanisms float64 `json:"mechanisms"`
	Reagents  float64 `json:"reagents"`
	Energy    float64 `json:"energy"`
	MaxEnergy float64 `json:"max_energy"`
	Money     float64 `json:"money"`
	AlienTech float64 `json:"alien_tech"`
}

// Planet manages a single planet's game state.
type Planet struct {
	ID            string
	OwnerID       string
	Name          string
	Level         int
	Buildings     map[string]int
	BuildProgress map[string]float64
	Resources     PlanetResources
	BuildSpeed    float64
	LastTick      time.Time
	EnergyBalance float64
	Research      *research.ResearchSystem
	Fleet         *ship.Fleet
	Shipyard      *ship.Shipyard
	game          *Game
}

// NewPlanet creates a new planet with default resources.
func NewPlanet(id, ownerID, name string, g *Game) *Planet {
	p := &Planet{
		ID:        id,
		OwnerID:   ownerID,
		Name:      name,
		Level:     1,
		Buildings: make(map[string]int),
		BuildProgress: make(map[string]float64),
		Resources: PlanetResources{
			Food:      100,
			Composite: 0,
			Mechanisms: 0,
			Reagents:  0,
			Energy:    100,
			MaxEnergy: 100,
			Money:     500,
			AlienTech: 0,
		},
		BuildSpeed: 1.0,
		LastTick:   time.Now(),
		Fleet:      ship.NewFleet(),
		Shipyard:   ship.NewShipyard(),
		game:       g,
	}
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

// AddBuilding adds or upgrades a building on the planet.
func (p *Planet) AddBuilding(bt string) {
	currentLevel := p.Buildings[bt]
	p.Buildings[bt] = currentLevel + 1
	p.BuildProgress[bt] = p.getBuildTime(bt, currentLevel+1)
}

// GetBuildingLevel returns the level of a building.
func (p *Planet) GetBuildingLevel(bt string) int {
	return p.Buildings[bt]
}

// StepBuilding advances construction progress for a building.
func (p *Planet) StepBuilding(bt string) {
	if progress, ok := p.BuildProgress[bt]; ok && progress > 0 {
		p.BuildProgress[bt] -= p.BuildSpeed
		if p.BuildProgress[bt] <= 0 {
			delete(p.BuildProgress, bt)
		}
	}
}

// getBuildTime returns the build time for a building at the given level.
func (p *Planet) getBuildTime(bt string, level int) float64 {
	switch bt {
	case "farm":
		return float64(level*level*level*20 + 100)
	case "solar":
		return float64(level*level*200 + 80)
	case "storage":
		return float64(level*level+1) * 100
	case "base":
		return math.Pow(2, float64(level+3)) + 100
	case "factory":
		return float64(level*2+1) * 100000
	case "energy_storage":
		return float64(level*level) + 1000
	case "shipyard":
		return math.Pow(2, float64(level+7)) + 3000
	case "comcenter":
		if level == 0 {
			return 10000000
		}
		return 10000000 * float64(level)
	default:
		return 100
	}
}

// getEnergyConsumption returns the energy consumption for a building at the given level.
func (p *Planet) getEnergyConsumption(bt string, level int) float64 {
	switch bt {
	case "farm":
		return float64(level) * 10
	case "solar":
		return float64(level) * -15 // negative = production
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
	default:
		return 0
	}
}

// getProduction returns the production result for a building at the given level.
func (p *Planet) getProduction(bt string, level int) ProductionResult {
	if level <= 0 {
		return ProductionResult{}
	}
	prod := ProductionResult{HasEnergy: true}
	switch bt {
	case "farm":
		prod.Food = float64(level)
	case "solar":
		prod.Energy = float64(level) * 15
	case "base":
		prod.Food = -float64(level) // base consumes food
	case "factory":
		// Factory produces random resource
		prod.Composite = float64(level) * 0.5
	case "energy_storage":
		// No direct production
	case "shipyard":
		// No direct production
	case "comcenter":
		// No direct production
	case "storage":
		// No direct production
	}
	return prod
}

// calculateEnergyBalance computes total energy production and consumption.
func (p *Planet) calculateEnergyBalance() (production float64, consumption float64) {
	for bt, level := range p.Buildings {
		con := p.getEnergyConsumption(bt, level)
		if con < 0 {
			production += float64(-con)
		} else if con > 0 {
			consumption += con
		}
	}
	return production, consumption
}

// calculateMaxEnergy returns the maximum energy capacity.
func (p *Planet) calculateMaxEnergy() float64 {
	base := 100.0
	for bt, level := range p.Buildings {
		if bt == "energy_storage" {
			base += float64(level) * 100
		}
	}
	return base
}

// calculateStorageCapacity returns the total storage capacity.
func (p *Planet) calculateStorageCapacity() float64 {
	base := 1000.0
	if level, ok := p.Buildings["storage"]; ok {
		base += float64(level) * 1000
	}
	return base
}

// Tick processes one game tick (1 second).
func (p *Planet) Tick() {
	now := time.Now()
	_ = now.Sub(p.LastTick).Seconds()
	p.LastTick = now

	// Update building construction progress
	for bt := range p.Buildings {
		p.StepBuilding(bt)
	}

	// Advance research progress
	p.Research.Tick()

	// Advance ship construction
	if completed := p.Shipyard.Tick(); completed != nil {
		st := ship.GetShipType(*completed)
		if st != nil {
			p.Fleet.AddShip(st, 1)
		}
	}

	// Calculate ship energy consumption
	shipEnergy := p.Fleet.TotalEnergyConsumption()

	// Calculate energy balance
	production, consumption := p.calculateEnergyBalance()
	consumption += shipEnergy
	hasEnergy := production >= consumption

	// Update max energy capacity
	maxEnergy := p.calculateMaxEnergy()
	if maxEnergy > p.Resources.MaxEnergy {
		p.Resources.MaxEnergy = maxEnergy
	}

	// Calculate production
	var totalProduction ProductionResult
	for bt, level := range p.Buildings {
		if !hasEnergy {
			continue
		}
		prod := p.getProduction(bt, level)
		if prod.HasEnergy {
			totalProduction.Add(prod)
		}
		_ = bt // suppress unused warning
	}

	// Apply production to resources
	if hasEnergy {
		p.Resources.Energy += totalProduction.Energy
		if p.Resources.Energy > p.Resources.MaxEnergy {
			p.Resources.Energy = p.Resources.MaxEnergy
		}
		p.Resources.Food += totalProduction.Food
		p.Resources.Composite += totalProduction.Composite
		p.Resources.Mechanisms += totalProduction.Mechanisms
		p.Resources.Reagents += totalProduction.Reagents
		p.Resources.Money += totalProduction.Money
		p.Resources.AlienTech += totalProduction.AlienTech

		// Clamp resources to storage capacity
		storageCapacity := p.calculateStorageCapacity()
		p.Resources.Food = math.Min(p.Resources.Food, storageCapacity)
		p.Resources.Composite = math.Min(p.Resources.Composite, storageCapacity)
		p.Resources.Mechanisms = math.Min(p.Resources.Mechanisms, storageCapacity)
		p.Resources.Reagents = math.Min(p.Resources.Reagents, storageCapacity)
		p.Resources.Energy = math.Min(p.Resources.Energy, p.Resources.MaxEnergy)
	}

	// Clamp resources to non-negative
	p.Resources.Food = math.Max(0, p.Resources.Food)
	p.Resources.Composite = math.Max(0, p.Resources.Composite)
	p.Resources.Mechanisms = math.Max(0, p.Resources.Mechanisms)
	p.Resources.Reagents = math.Max(0, p.Resources.Reagents)
	p.Resources.Energy = math.Max(0, p.Resources.Energy)
	p.Resources.Money = math.Max(0, p.Resources.Money)
	p.Resources.AlienTech = math.Max(0, p.Resources.AlienTech)

	p.EnergyBalance = production - consumption

	// Save to DB (throttled)
	if p.game.shouldSave(p.ID) {
		p.game.savePlanet(p)
	}

	// Broadcast state update
	p.game.broadcastPlanetUpdate(p)
}

// GetState returns the planet's state as a JSON-serializable map.
func (p *Planet) GetState() map[string]interface{} {
	shipyardLevel := p.Buildings["shipyard"]
	maxSlots := p.Shipyard.MaxSlots(p.Buildings["base"])

	return map[string]interface{}{
		"id":               p.ID,
		"owner_id":         p.OwnerID,
		"name":             p.Name,
		"level":            p.Level,
		"resources":        p.Resources,
		"buildings":        p.Buildings,
		"build_progress":   p.BuildProgress,
		"build_speed":      p.BuildSpeed,
		"energy_balance":   p.EnergyBalance,
		"last_tick":        p.LastTick.Format(time.RFC3339),
		"fleet":            p.Fleet.GetShipState(),
		"shipyard_level":   shipyardLevel,
		"shipyard_queue":   p.Shipyard.Queue,
		"shipyard_slots":   p.Fleet.TotalSlots(),
		"shipyard_max":     maxSlots,
		"shipyard_progress": p.Shipyard.GetQueueProgress(),
	}
}

// GetResourcesJSON returns a JSON representation of resources.
func (p *Planet) GetResourcesJSON() ([]byte, error) {
	return json.Marshal(p.Resources)
}

// GetEnergyBalance returns the current energy balance.
func (p *Planet) GetEnergyBalance() float64 {
	production, consumption := p.calculateEnergyBalance()
	return production - consumption
}

// GetProductionResult returns the production result for this tick.
func (p *Planet) GetProductionResult() ProductionResult {
	production, consumption := p.calculateEnergyBalance()
	hasEnergy := production >= consumption

	var totalProduction ProductionResult
	for bt, level := range p.Buildings {
		if !hasEnergy {
			continue
		}
		prod := p.getProduction(bt, level)
		if prod.HasEnergy {
			totalProduction.Add(prod)
		}
		_ = bt
	}
	return totalProduction
}

// GetTotalBuildingLevels returns the sum of all building levels.
func (p *Planet) GetTotalBuildingLevels() int {
	total := 0
	for _, level := range p.Buildings {
		total += level
	}
	return total
}

// LogPlanetState logs the current planet state for debugging.
func (p *Planet) LogPlanetState() {
	log.Printf("Planet %s (%s): Food=%.0f Energy=%.0f/%.0f Money=%.0f Buildings=%v",
		p.Name, p.ID,
		p.Resources.Food,
		p.Resources.Energy, p.Resources.MaxEnergy,
		p.Resources.Money,
		p.Buildings)
}

// StartResearch begins researching a technology on this planet.
func (p *Planet) StartResearch(techID string) error {
	tech := research.GetTechByID(techID)
	if tech == nil {
		return &PlanetError{planetID: p.ID, reason: "tech_not_found"}
	}
	return p.Research.StartResearch(tech, p.Resources.Food, p.Resources.Money, p.Resources.AlienTech)
}

// GetAvailableResearch returns technologies that can be researched on this planet.
func (p *Planet) GetAvailableResearch() ([]byte, error) {
	return p.Research.GetAvailableForAPI()
}

// GetResearchProgress returns the progress percentage (0-100) for a tech.
func (p *Planet) GetResearchProgress(techID string) float64 {
	return p.Research.GetResearchProgress(techID)
}

// GetResearchState returns the full research state for a tech.
func (p *Planet) GetResearchState(techID string) *research.ResearchState {
	return p.Research.GetResearchState(techID)
}

// GetResearchJSON returns all research state as JSON.
func (p *Planet) GetResearchJSON() ([]byte, error) {
	return p.Research.GetResearchJSON()
}

// GetResearchErrors returns the completed tech map.
func (p *Planet) GetResearchCompleted() map[string]int {
	return p.Research.GetCompleted()
}

// CanBuildShip checks if the planet can build a ship of the given type.
func (p *Planet) CanBuildShip(typeID ship.TypeID) bool {
	st := ship.GetShipType(typeID)
	if st == nil {
		return false
	}

	shipyardLevel := p.Buildings["shipyard"]
	if st.MinShipyard > shipyardLevel {
		return false
	}

	if !st.Cost.CanAfford(p.Resources.Food, p.Resources.Composite, p.Resources.Mechanisms, p.Resources.Reagents, p.Resources.Money) {
		return false
	}

	maxSlots := p.Shipyard.MaxSlots(p.Buildings["base"])
	return p.Fleet.CanAddShip(st, 1, maxSlots)
}

// BuildShip queues a ship for construction. Returns error if can't build.
func (p *Planet) BuildShip(typeID ship.TypeID) error {
	st := ship.GetShipType(typeID)
	if st == nil {
		return &PlanetError{planetID: p.ID, reason: "unknown_ship_type"}
	}

	if !p.CanBuildShip(typeID) {
		return &PlanetError{planetID: p.ID, reason: "cannot_build"}
	}

	if err := p.Shipyard.QueueShip(st); err != nil {
		return &PlanetError{planetID: p.ID, reason: "queue_failed"}
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

// GetShipCount returns the number of ships of a given type.
func (p *Planet) GetShipCount(typeID ship.TypeID) int {
	return p.Fleet.GetShipCount(typeID)
}

// GetTotalShipCount returns the total number of ships in the fleet.
func (p *Planet) GetTotalShipCount() int {
	return p.Fleet.TotalShipCount()
}

// GetShipEnergyConsumption returns the energy consumed by the fleet.
func (p *Planet) GetShipEnergyConsumption() float64 {
	return p.Fleet.TotalEnergyConsumption()
}

// PlanetError represents an error that occurred during a planet operation.
type PlanetError struct {
	planetID string
	reason   string
}

func (e *PlanetError) Error() string {
	return "planet error: " + e.reason + " (planet: " + e.planetID + ")"
}
