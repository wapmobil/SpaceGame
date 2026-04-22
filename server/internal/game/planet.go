package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"spacegame/internal/game/battle"
	"spacegame/internal/game/building"
	"spacegame/internal/game/expedition"
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

// BattleRecord stores the result of a completed battle.
type BattleRecord struct {
	ID         string            `json:"id"`
	Opponent   string            `json:"opponent"`
	Result     string            `json:"result"`
	Loot       map[string]float64 `json:"loot"`
	LostShips  map[string]int    `json:"lost_ships"`
	Refund     map[string]float64 `json:"refund"`
	Rounds     int               `json:"rounds"`
	Timestamp  time.Time         `json:"timestamp"`
}

// Planet manages a single planet's game state.
type Planet struct {
	ID               string
	OwnerID          string
	Name             string
	Level            int
	Buildings        map[string]int
	BuildProgress      map[string]float64
	BuildPending      map[string]bool
	ActiveConstruction int
	Resources          PlanetResources
	BuildSpeed         float64
	LastTick         time.Time
	EnergyBalance    float64
	Research         *research.ResearchSystem
	Fleet            *ship.Fleet
	Shipyard         *ship.Shipyard
	Battles          []BattleRecord
	Expeditions      []*expedition.Expedition
	ExplorationMgr   *expedition.ExplorationManager
	game             *Game
}

// NewPlanet creates a new planet with default resources.
func NewPlanet(id, ownerID, name string, g *Game) *Planet {
	p := &Planet{
		ID:             id,
		OwnerID:        ownerID,
		Name:           name,
		Level:          1,
		Buildings:      make(map[string]int),
		BuildProgress:  make(map[string]float64),
		BuildPending:   make(map[string]bool),
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
		BuildSpeed:     1.0,
		LastTick:       time.Now(),
		Fleet:          ship.NewFleet(),
		Shipyard:       ship.NewShipyard(),
		Expeditions:    make([]*expedition.Expedition, 0),
		ExplorationMgr: expedition.NewExplorationManager(),
		game:           g,
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
func (p *Planet) AddBuilding(bt string) (float64, float64, error) {
	currentLevel := p.Buildings[bt]

	// Check if already under construction (don't count as active if already in BuildProgress)
	_, alreadyConstructing := p.BuildProgress[bt]

	// Check max concurrent construction limit
	if p.ActiveConstruction >= p.GetMaxConcurrentBuildings() {
		return 0, 0, &PlanetError{planetID: p.ID, reason: "max_constructions_reached", extra: fmt.Sprintf("Max constructions reached (%d/%d). Research Parallel Construction to unlock more.", p.ActiveConstruction, p.GetMaxConcurrentBuildings())}
	}

	// Get cost for next level (currentLevel)
	cost := p.GetBuildingCost(bt, currentLevel)

	// Check affordability
	if cost.Food > p.Resources.Food {
		return 0, 0, &PlanetError{planetID: p.ID, reason: "insufficient_food", extra: fmt.Sprintf("Need %.0f food, have %.0f", cost.Food, p.Resources.Food)}
	}
	if cost.Money > p.Resources.Money {
		return 0, 0, &PlanetError{planetID: p.ID, reason: "insufficient_money", extra: fmt.Sprintf("Need %.0f money, have %.0f", cost.Money, p.Resources.Money)}
	}

	// Deduct resources
	p.Resources.Food -= cost.Food
	p.Resources.Money -= cost.Money

	// Start construction
	p.Buildings[bt] = currentLevel + 1
	p.BuildProgress[bt] = p.GetBuildTime(bt, currentLevel+1)

	// Track active construction (only if not already counting)
	if !alreadyConstructing {
		p.ActiveConstruction++
	}

	return cost.Food, cost.Money, nil
}

// AddBuildingDirect sets a building level without any checks (for testing).
func (p *Planet) AddBuildingDirect(bt string, level int) {
	p.Buildings[bt] = level
}

// GetBuildingLevel returns the level of a building.
func (p *Planet) GetBuildingLevel(bt string) int {
	return p.Buildings[bt]
}

// GetMaxConcurrentBuildings returns the maximum number of simultaneous construction projects.
func (p *Planet) GetMaxConcurrentBuildings() int {
	max := 1
	if lvl, ok := p.Research.GetCompleted()["parallel_construction"]; ok && lvl > 0 {
		max += lvl
	}
	return max
}

// StepBuilding advances construction progress for a building.
func (p *Planet) StepBuilding(bt string) {
	if progress, ok := p.BuildProgress[bt]; ok && progress > 0 {
		p.BuildProgress[bt] -= p.BuildSpeed
		if p.BuildProgress[bt] <= 0 {
			delete(p.BuildProgress, bt)
			p.ActiveConstruction--
			if p.ActiveConstruction < 0 {
				p.ActiveConstruction = 0
			}
			p.BuildPending[bt] = true
		}
	}
}

// ConfirmBuilding confirms a pending building construction, making it operational.
func (p *Planet) ConfirmBuilding(bt string) error {
	if !p.BuildPending[bt] {
		return &PlanetError{planetID: p.ID, reason: "building_not_pending"}
	}
	delete(p.BuildPending, bt)
	return nil
}

// IsBuildingPending returns true if a building is pending confirmation.
func (p *Planet) IsBuildingPending(bt string) bool {
	return p.BuildPending[bt]
}

// GetPendingBuildings returns all pending building types.
func (p *Planet) GetPendingBuildings() map[string]bool {
	return p.BuildPending
}

// GetBuildTime returns the build time for a building at the given level.
func (p *Planet) GetBuildTime(bt string, level int) float64 {
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
	case "composite_drone":
		return float64(level*level+1) * 100
	case "mechanism_factory":
		return float64(level*level+1) * 100
	case "reagent_lab":
		return float64(level*level+1) * 100
	default:
		return 100
	}
}

// GetBuildingCost returns the multi-resource cost to build a building at the given level.
func (p *Planet) GetBuildingCost(bt string, level int) building.CostMulti {
	switch bt {
	case "farm":
		return building.CostMulti{
			Food:  float64(level*level*level*20 + 100),
			Money: float64(level*level*level*10 + 50),
		}
	case "solar":
		return building.CostMulti{
			Food:  float64(level*level*120 + 48),
			Money: float64(level*level*80 + 32),
		}
	case "storage":
		return building.CostMulti{
			Food:  float64(level*level+1) * 60,
			Money: float64(level*level+1) * 40,
		}
	case "base":
		return building.CostMulti{
			Food:  math.Pow(2, float64(level+2)),
			Money: math.Pow(2, float64(level+3)),
		}
	case "factory":
		return building.CostMulti{
			Food:  float64(level*2+1) * 2500,
			Money: float64(level*2+1) * 1500,
		}
	case "energy_storage":
		return building.CostMulti{
			Food:  float64(level*level) * 300,
			Money: float64(level*level) * 200,
		}
	case "shipyard":
		val := math.Pow(2, float64(level+5)) * 0.5
		return building.CostMulti{
			Food:  val,
			Money: val,
		}
	case "comcenter":
		return building.CostMulti{
			Food:  float64(level) * 10000,
			Money: float64(level) * 10000,
		}
	case "composite_drone":
		fallthrough
	case "mechanism_factory":
		fallthrough
	case "reagent_lab":
		return building.CostMulti{
			Food:  float64(level*level+1) * 60,
			Money: float64(level*level+1) * 40,
		}
	default:
		return building.CostMulti{}
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
	case "composite_drone":
		prod.Composite = float64(level) * 0.5
	case "mechanism_factory":
		prod.Mechanisms = float64(level) * 0.5
	case "reagent_lab":
		prod.Reagents = float64(level) * 0.5
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

	// Advance expeditions
	p.TickExpeditions()

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
		if p.BuildPending[bt] {
			continue
		}
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
	p.broadcastPlanetUpdate()
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
		"expeditions":       p.GetExpeditionState(),
		"active_expeditions":   p.GetActiveExpeditionsCount(),
		"max_expeditions":      p.GetMaxExpeditions(),
		"active_constructions": p.ActiveConstruction,
		"max_constructions":    p.GetMaxConcurrentBuildings(),
		"pending_buildings":    p.BuildPending,
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

// AutoBattle simulates an auto-battle against an NPC planet fleet.
func (p *Planet) AutoBattle(npcFleet *ship.Fleet) *BattleRecord {
	attackerSnapshot := battle.NewFleetSnapshot(p.Fleet)
	defenderSnapshot := battle.NewFleetSnapshot(npcFleet)

	result := battle.CalculateBattle(attackerSnapshot, defenderSnapshot)

	battleRecord := BattleRecord{
		Opponent:  "npc",
		Result:    result.Winner,
		Loot:      result.AttackerLoot,
		LostShips: result.AttackerLost,
		Refund:    result.AttackerRefund,
		Rounds:    result.Rounds,
		Timestamp: time.Now(),
	}

	if result.Winner == "attacker" {
		// Apply loot to planet resources
		p.Resources.Money += result.AttackerLoot["money"]
		p.Resources.AlienTech += result.AttackerLoot["alien_tech"]
		p.Resources.Food += result.AttackerLoot["food"]
		p.Resources.Composite += result.AttackerLoot["composite"]
		p.Resources.Mechanisms += result.AttackerLoot["mechanisms"]
		p.Resources.Reagents += result.AttackerLoot["reagents"]

		// Apply refund for lost ships
		p.Resources.Money += result.AttackerRefund["money"]
		p.Resources.Food += result.AttackerRefund["food"]
		p.Resources.Composite += result.AttackerRefund["composite"]
		p.Resources.Mechanisms += result.AttackerRefund["mechanisms"]
		p.Resources.Reagents += result.AttackerRefund["reagents"]

		// Remove destroyed ships from fleet
		for typeID, count := range result.AttackerLost {
			p.Fleet.RemoveShips(typeID, count)
		}

		log.Printf("Planet %s won battle in %d rounds, loot: %v", p.ID, result.Rounds, result.AttackerLoot)
	} else if result.Winner == "defender" {
		// Remove destroyed ships from fleet
		for typeID, count := range result.AttackerLost {
			p.Fleet.RemoveShips(typeID, count)
		}

		log.Printf("Planet %s lost battle in %d rounds", p.ID, result.Rounds)
	} else {
		// Draw - both sides lost ships
		for typeID, count := range result.AttackerLost {
			p.Fleet.RemoveShips(typeID, count)
		}

		log.Printf("Planet %s drew battle in %d rounds", p.ID, result.Rounds)
	}

	p.Battles = append(p.Battles, battleRecord)

	// Keep only last 50 battles
	if len(p.Battles) > 50 {
		p.Battles = p.Battles[len(p.Battles)-50:]
	}

	return &battleRecord
}

// GetBattleHistory returns the planet's battle history.
func (p *Planet) GetBattleHistory() []BattleRecord {
	return p.Battles
}

// GetFleetSnapshot creates a battle snapshot of the current fleet.
func (p *Planet) GetFleetSnapshot() *battle.FleetSnapshot {
	return battle.NewFleetSnapshot(p.Fleet)
}

// GetFleetStrength returns the total combat strength of the fleet.
func (p *Planet) GetFleetStrength() float64 {
	snapshot := p.GetFleetSnapshot()
	return snapshot.TotalDPS() + snapshot.TotalHP()*0.1
}

// HasFleet returns true if the planet has any ships.
func (p *Planet) HasFleet() bool {
	return p.Fleet.TotalShipCount() > 0
}

// HasCombatFleet returns true if the planet has ships with weapons.
func (p *Planet) HasCombatFleet() bool {
	snapshot := p.GetFleetSnapshot()
	return snapshot.HasCombatShips()
}

// CanStartExpedition checks if the planet can start a new expedition.
func (p *Planet) CanStartExpedition(expType expedition.Type, fleet *ship.Fleet) error {
	// Check if expeditions research is completed
	if _, ok := p.Research.GetCompleted()["expeditions"]; !ok {
		return &PlanetError{planetID: p.ID, reason: "expeditions_not_researched"}
	}

	// Check max concurrent expeditions
	maxExpeditions := 1
	if _, ok := p.Research.GetCompleted()["additional_expedition"]; ok {
		maxExpeditions = 2
	}

	activeCount := 0
	for _, exp := range p.Expeditions {
		if exp.Status == expedition.StatusActive || exp.Status == expedition.StatusAtPoint {
			activeCount++
		}
	}
	if activeCount >= maxExpeditions {
		return &PlanetError{planetID: p.ID, reason: "max_expeditions_reached"}
	}

	// Check fleet has ships
	if fleet.TotalShipCount() == 0 {
		return &PlanetError{planetID: p.ID, reason: "no_ships_available"}
	}

	// Check energy
	energyCost := fleet.TotalEnergyConsumption()
	if p.Resources.Energy < energyCost {
		return &PlanetError{planetID: p.ID, reason: "insufficient_energy"}
	}

	return nil
}

// StartExpedition creates and starts a new expedition.
func (p *Planet) StartExpedition(expType expedition.Type, fleet *ship.Fleet, target string, duration float64) (*expedition.Expedition, error) {
	if err := p.CanStartExpedition(expType, fleet); err != nil {
		return nil, err
	}

	// Create a copy of the fleet for the expedition
	expFleet := ship.NewFleet()
	for key, fs := range fleet.Ships {
		expFleet.Ships[key] = &ship.FleetShip{
			TypeID: fs.TypeID,
			Count:  fs.Count,
			HP:     fs.HP,
		}
	}

	// Remove expedition ships from main fleet
	for key, ef := range expFleet.Ships {
		fleet.RemoveShips(key, ef.Count)
	}

	// Deduct energy for expedition
	expFleetEnergy := battle.NewFleetSnapshot(expFleet).TotalDPS() * 0.5
	p.Resources.Energy -= expFleetEnergy
	if p.Resources.Energy < 0 {
		p.Resources.Energy = 0
	}

	// Create expedition
	id := p.ID + "_exp_" + time.Now().Format("20060102150405")
	exp := expedition.CreateExpedition(id, p.ID, target, expType, expFleet, duration)

	p.Expeditions = append(p.Expeditions, exp)

	log.Printf("Planet %s started %s expedition with %d ships, duration: %.0fs",
		p.Name, expType, expFleet.TotalShipCount(), duration)

	return exp, nil
}

// TickExpeditions processes all active expeditions for one tick.
func (p *Planet) TickExpeditions() {
	for i := len(p.Expeditions) - 1; i >= 0; i-- {
		exp := p.Expeditions[i]

		if exp.IsExpired() {
			if exp.Status == expedition.StatusCompleted {
				p.returnExpedition(i)
			}
			continue
		}

		exp.Tick()

		// Exploration: check for NPC planet discovery
		if exp.ExpeditionType == expedition.TypeExploration && exp.DiscoveredNPC == nil {
			chance := expedition.CalculateDiscoveryChance(exp.Fleet, exp.ElapsedTime, exp.Duration)
			if rand.Float64() < chance {
				npc := p.ExplorationMgr.DiscoverNPCPlanet(exp.ID, p.OwnerID)
				exp.DiscoveredNPC = npc
				exp.Status = expedition.StatusAtPoint
				exp.Actions = exp.GetAvailableActions()
				log.Printf("Planet %s expedition discovered NPC planet: %s (type: %s)",
					p.Name, npc.Name, npc.Type)
			}
		}
	}
}

// DoExpeditionAction performs an action at a point of interest.
func (p *Planet) DoExpeditionAction(expID, actionType string) error {
	exp, idx := p.findExpedition(expID)
	if exp == nil {
		return &PlanetError{planetID: p.ID, reason: "expedition_not_found"}
	}

	if exp.Status != expedition.StatusAtPoint {
		return &PlanetError{planetID: p.ID, reason: "expedition_not_at_point"}
	}

	if exp.DiscoveredNPC == nil {
		return &PlanetError{planetID: p.ID, reason: "no_npc_discovered"}
	}

	npc := exp.DiscoveredNPC

	switch actionType {
	case "loot":
		p.lootNPCPlanet(exp, npc)
	case "attack":
		return p.attackNPCPlanet(exp, npc, idx)
	case "wait":
		exp.Actions = []expedition.ExpeditionAction{
			{ID: "wait", Type: "wait", Label: "Waiting for reinforcements..."},
		}
	case "leave":
		p.leaveNPCPlanet(exp)
	default:
		return &PlanetError{planetID: p.ID, reason: "unknown_action"}
	}

	return nil
}

// lootNPCPlanet collects resources from an NPC planet.
func (p *Planet) lootNPCPlanet(exp *expedition.Expedition, npc *expedition.NPCPlanet) {
	cargoCapacity := exp.Fleet.TotalCargoCapacity()
	collected, _ := expedition.CollectResources(npc, cargoCapacity)

	for resName, amount := range collected {
		log.Printf("Collected %f %s from %s", amount, resName, npc.Name)
		switch resName {
		case "food":
			p.Resources.Food += amount
		case "composite":
			p.Resources.Composite += amount
		case "mechanisms":
			p.Resources.Mechanisms += amount
		case "reagents":
			p.Resources.Reagents += amount
		case "money":
			p.Resources.Money += amount
		case "alien_tech":
			p.Resources.AlienTech += amount
		}
	}

	// Check if all resources collected
	if npc.TotalResources() <= 0 {
		p.ExplorationMgr.RemoveNPCPlanet(npc.ID)
		exp.DiscoveredNPC = nil
		exp.Status = expedition.StatusActive
		exp.Actions = []expedition.ExpeditionAction{}
	} else {
		exp.Actions = exp.GetAvailableActions()
	}
}

// attackNPCPlanet initiates a battle with NPC planet's fleet.
func (p *Planet) attackNPCPlanet(exp *expedition.Expedition, npc *expedition.NPCPlanet, expIdx int) error {
	expSnapshot := battle.NewFleetSnapshot(exp.Fleet)
	if !expSnapshot.HasCombatShips() {
		return &PlanetError{planetID: p.ID, reason: "no_combat_ships"}
	}

	attackerSnapshot := expSnapshot
	defenderSnapshot := battle.NewFleetSnapshot(npc.EnemyFleet)

	result := battle.CalculateBattle(attackerSnapshot, defenderSnapshot)

	if result.Winner == "attacker" {
		// Apply loot
		for resName, amount := range result.AttackerLoot {
			log.Printf("Battle loot: %f %s", amount, resName)
		}

		// Apply refunds
		for typeID, count := range result.AttackerLost {
			exp.Fleet.RemoveShips(typeID, count)
		}

		// Remove NPC planet
		p.ExplorationMgr.RemoveNPCPlanet(npc.ID)
		exp.DiscoveredNPC = nil

		if exp.Fleet.TotalShipCount() == 0 {
			exp.Status = expedition.StatusReturning
			exp.Duration = exp.Duration - exp.ElapsedTime
			exp.ElapsedTime = 0
			go func() {
				time.Sleep(time.Duration((exp.Duration-exp.ElapsedTime)*1000) * time.Millisecond)
				p.returnExpedition(expIdx)
			}()
		} else {
			exp.Status = expedition.StatusActive
			exp.Actions = []expedition.ExpeditionAction{}
		}

		log.Printf("Expedition %s won battle against %s", exp.ID, npc.Name)
	} else {
		// Expedition lost
		for typeID, count := range result.AttackerLost {
			exp.Fleet.RemoveShips(typeID, count)
		}

		if exp.Fleet.TotalShipCount() == 0 {
			exp.Status = expedition.StatusFailed
			log.Printf("Expedition %s failed - all ships lost", exp.ID)
		} else {
			exp.Status = expedition.StatusReturning
			exp.Duration = 60 // 60 seconds to return
			exp.ElapsedTime = 0
			go func() {
				time.Sleep(60 * time.Second)
				p.returnExpedition(expIdx)
			}()
		}
	}

	return nil
}

// leaveNPCPlanet continues the expedition without collecting resources.
func (p *Planet) leaveNPCPlanet(exp *expedition.Expedition) {
	if exp.DiscoveredNPC != nil {
		p.ExplorationMgr.RemoveNPCPlanet(exp.DiscoveredNPC.ID)
	}
	exp.DiscoveredNPC = nil
	exp.Status = expedition.StatusActive
	exp.Actions = []expedition.ExpeditionAction{}
}

// returnExpedition returns an expedition to the home planet.
func (p *Planet) returnExpedition(idx int) {
	if idx < 0 || idx >= len(p.Expeditions) {
		return
	}

	exp := p.Expeditions[idx]

	// Return ships to fleet
	for key, fs := range exp.Fleet.Ships {
		p.Fleet.AddShip(ship.GetShipType(fs.TypeID), fs.Count)
		_ = key
	}

	exp.Status = expedition.StatusCompleted
	p.Expeditions = append(p.Expeditions[:idx], p.Expeditions[idx+1:]...)

	log.Printf("Expedition %s returned to planet %s", exp.ID, p.Name)
}

// findExpedition finds an expedition by ID.
func (p *Planet) findExpedition(id string) (*expedition.Expedition, int) {
	for i, exp := range p.Expeditions {
		if exp.ID == id {
			return exp, i
		}
	}
	return nil, -1
}

// GetExpeditions returns all expeditions for a planet.
func (p *Planet) GetExpeditions() []*expedition.Expedition {
	return p.Expeditions
}

// GetActiveExpeditionsCount returns the number of active expeditions.
func (p *Planet) GetActiveExpeditionsCount() int {
	count := 0
	for _, exp := range p.Expeditions {
		if exp.Status == expedition.StatusActive || exp.Status == expedition.StatusAtPoint {
			count++
		}
	}
	return count
}

// GetMaxExpeditions returns the maximum number of concurrent expeditions.
func (p *Planet) GetMaxExpeditions() int {
	max := 1
	if _, ok := p.Research.GetCompleted()["additional_expedition"]; ok {
		max = 2
	}
	return max
}

// GetExpeditionState returns the expedition state as a JSON-serializable map.
func (p *Planet) GetExpeditionState() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(p.Expeditions))
	for _, exp := range p.Expeditions {
		state := map[string]interface{}{
			"id":               exp.ID,
			"planet_id":        exp.PlanetID,
			"target":           exp.Target,
			"progress":         exp.Progress,
			"status":           exp.Status,
			"expedition_type":  exp.ExpeditionType,
			"duration":         exp.Duration,
			"elapsed_time":     exp.ElapsedTime,
			"fleet_ships":      exp.Fleet.GetShipState(),
			"fleet_total":      exp.Fleet.TotalShipCount(),
			"fleet_cargo":      exp.Fleet.TotalCargoCapacity(),
			"fleet_energy":     exp.Fleet.TotalEnergyConsumption(),
			"fleet_damage":     exp.Fleet.TotalDamage(),
			"discovered_npc":   nil,
			"actions":          exp.Actions,
			"created_at":       exp.CreatedAt.Format(time.RFC3339),
			"updated_at":       exp.UpdatedAt.Format(time.RFC3339),
		}

		if exp.DiscoveredNPC != nil {
			npcState := map[string]interface{}{
				"id":             exp.DiscoveredNPC.ID,
				"name":           exp.DiscoveredNPC.Name,
				"type":           exp.DiscoveredNPC.Type,
				"resources":      exp.DiscoveredNPC.Resources,
				"total_resources": exp.DiscoveredNPC.TotalResources(),
				"has_combat":     exp.DiscoveredNPC.HasCombatShips(),
				"fleet_strength": exp.DiscoveredNPC.TotalFleetStrength(),
			}
			if exp.DiscoveredNPC.EnemyFleet != nil {
				npcState["enemy_fleet"] = exp.DiscoveredNPC.EnemyFleet.GetShipState()
			}
			state["discovered_npc"] = npcState
		}

		result = append(result, state)
	}
	return result
}

// PlanetError represents an error that occurred during a planet operation.
type PlanetError struct {
	planetID string
	reason   string
	extra    string
}

func (e *PlanetError) Error() string {
	if e.extra != "" {
		return "planet error: " + e.reason + " - " + e.extra + " (planet: " + e.planetID + ")"
	}
	return "planet error: " + e.reason + " (planet: " + e.planetID + ")"
}
