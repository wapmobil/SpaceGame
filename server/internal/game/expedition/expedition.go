package expedition

import (
	"math"
	"math/rand"
	"time"

	"spacegame/internal/game/battle"
	"spacegame/internal/game/ship"
)

// Type represents the type of expedition.
type Type string

const (
	TypeExploration Type = "space_exploration" // Discovers NPC planets
	TypeTrade       Type = "space_trade"       // Trade via marketplace
	TypeSupport     Type = "space_support"      // Help other players
)

// Status represents the current status of an expedition.
type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusAtPoint   Status = "at_point" // At a point of interest, waiting for action
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusReturning Status = "returning"
)

// PointOfInterestType represents what kind of point was discovered.
type PointOfInterestType string

const (
	POIAbandonedStation  PointOfInterestType = "abandoned_station"
	POIDebris            PointOfInterestType = "debris"
	POIAsteroids         PointOfInterestType = "asteroids"
	POIUnknownPlanet     PointOfInterestType = "unknown_planet"
	POIAlienBase         PointOfInterestType = "alien_base"
	POICosmicDebris      PointOfInterestType = "cosmic_debris"
)

// Expedition represents an expedition sent from a planet.
type Expedition struct {
	ID            string        `json:"id"`
	PlanetID      string        `json:"planet_id"`
	Fleet         *ship.Fleet   `json:"fleet"`
	Target        string        `json:"target"` // "space_exploration", "space_trade", "space_support", or NPC planet ID
	Progress      float64       `json:"progress"`
	Status        Status        `json:"status"`
	ExpeditionType Type        `json:"expedition_type"`
	Duration      float64       `json:"duration"` // seconds until return
	ElapsedTime   float64       `json:"elapsed_time"`
	DiscoveredNPC *NPCPlanet    `json:"discovered_npc,omitempty"`
	Actions       []ExpeditionAction `json:"actions"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// ExpeditionAction represents an available action at a point of interest.
type ExpeditionAction struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "loot", "attack", "wait", "leave"
	Label     string `json:"label"`
	Required  string `json:"required,omitempty"` // required ship type or condition
}

// CreateExpedition creates a new expedition.
func CreateExpedition(id, planetID, target string, expType Type, fleet *ship.Fleet, duration float64) *Expedition {
	now := time.Now()
	return &Expedition{
		ID:            id,
		PlanetID:      planetID,
		Fleet:         fleet,
		Target:        target,
		Progress:      0,
		Status:        StatusActive,
		ExpeditionType: expType,
		Duration:      duration,
		ElapsedTime:   0,
		DiscoveredNPC: nil,
		Actions:       []ExpeditionAction{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Tick advances the expedition by one second.
func (e *Expedition) Tick() {
	e.ElapsedTime++
	e.Progress = e.ElapsedTime / e.Duration

	if e.Progress >= 1.0 {
		e.Status = StatusCompleted
	}
}

// IsExpired returns true if the expedition has completed.
func (e *Expedition) IsExpired() bool {
	return e.Status == StatusCompleted || e.Status == StatusFailed
}

// GetAvailableActions returns the actions available at the current state.
func (e *Expedition) GetAvailableActions() []ExpeditionAction {
	if e.DiscoveredNPC == nil {
		return []ExpeditionAction{}
	}

	var actions []ExpeditionAction

	actions = append(actions, ExpeditionAction{
		ID:    "leave",
		Type:  "leave",
		Label: "Continue expedition",
	})

	npc := e.DiscoveredNPC
	if npc.TotalResources() > 0 {
		actions = append(actions, ExpeditionAction{
			ID:    "loot",
			Type:  "loot",
			Label: "Collect resources",
		})
	}

	if npc.HasCombatShips() {
		actions = append(actions, ExpeditionAction{
			ID:    "attack",
			Type:  "attack",
			Label: "Attack",
		})
	}

	actions = append(actions, ExpeditionAction{
		ID:    "wait",
		Type:  "wait",
		Label: "Wait for reinforcements",
	})

	return actions
}

// NPCPlanet represents an NPC-controlled planet or point of interest.
type NPCPlanet struct {
	ID          string               `json:"id"`
	OwnerID     string               `json:"owner_id"` // player who discovered it
	Name        string               `json:"name"`
	Type        PointOfInterestType  `json:"type"`
	Resources   map[string]float64   `json:"resources"`
	EnemyFleet  *ship.Fleet          `json:"enemy_fleet"`
	Discovered  bool                 `json:"discovered"`
	CreatedAt   time.Time            `json:"created_at"`
}

// PointNames maps POI types to generated names.
var PointNames = map[PointOfInterestType][]string{
	POIAbandonedStation: {
		"Abandoned Station Alpha", "Derelict Outpost Beta", "Lost Waypoint Gamma",
		"Ruined Station Delta", "Old Relay Epsilon",
	},
	POIDebris: {
		"Shipwreck Cluster", "Debris Field Alpha", "Scrap Zone Beta",
		"Remnant Field Gamma", "Wreckage Delta",
	},
	POIAsteroids: {
		"Asteroid Belt Theta", "Rocky Cluster Iota", "Mining Site Kappa",
		"Stone Ring Lambda", "Asteroid Hub Mu",
	},
	POIUnknownPlanet: {
		"Unknown Planet X-1", "Uncharted World Y-7", "Mystery Planet Z-3",
		"Lost World Alpha", "Hidden Planet Beta",
	},
	POIAlienBase: {
		"Alien Fortress", "Xenon Outpost", "Alien Stronghold",
		"Extra-Terrestrial Base", "Alien Command Center",
	},
	POICosmicDebris: {
		"Cosmic Dust Cloud", "Nebula Remnants", "Space Debris Field",
		"Ethereal Wreckage", "Void Debris",
	},
}

// GenerateName returns a random name for the given POI type.
func GenerateName(poiType PointOfInterestType) string {
	names := PointNames[poiType]
	if len(names) == 0 {
		return "Unknown Point"
	}
	return names[rand.Intn(len(names))]
}

// NewNPCPlanet creates a new NPC planet with generated content.
func NewNPCPlanet(id string, ownerID string, poiType PointOfInterestType) *NPCPlanet {
	npc := &NPCPlanet{
		ID:          id,
		OwnerID:     ownerID,
		Name:        GenerateName(poiType),
		Type:        poiType,
		Resources:   make(map[string]float64),
		EnemyFleet:  ship.NewFleet(),
		Discovered:  true,
		CreatedAt:   time.Now(),
	}
	npc.generate()
	return npc
}

// generate populates resources and enemy fleet based on POI type.
func (n *NPCPlanet) generate() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch n.Type {
	case POIAbandonedStation:
		n.generateResources(r, 10, 50)
	case POIDebris:
		n.generateResources(r, 20, 100)
	case POICosmicDebris:
		n.generateResources(r, 5, 30)
	case POIAsteroids:
		n.generateResources(r, 50, 200)
	case POIUnknownPlanet:
		n.generateResources(r, 100, 500)
		n.generateWeakFleet(r)
	case POIAlienBase:
		n.generateResources(r, 500, 2000)
		n.generateAlienFleet(r)
	}
}

// generateResources adds random resources to the NPC planet.
func (n *NPCPlanet) generateResources(r *rand.Rand, min, max float64) {
	resourceNames := []string{"food", "composite", "mechanisms", "reagents", "money", "alien_tech"}
	for _, resName := range resourceNames {
		amount := r.Float64()*(max-min) + min
		if resName == "money" {
			amount = r.Float64()*(max*10-min*10) + min*10
		}
		if resName == "alien_tech" && n.Type != POIAlienBase {
			amount = 0
		}
		n.Resources[resName] = amount
	}
}

// generateWeakFleet creates a weak enemy fleet.
func (n *NPCPlanet) generateWeakFleet(r *rand.Rand) {
	interceptor := ship.GetShipType("interceptor")
	if interceptor != nil {
		count := r.Intn(5) + 1
		n.EnemyFleet.AddShip(interceptor, count)
	}
}

// generateAlienFleet creates a strong alien fleet.
func (n *NPCPlanet) generateAlienFleet(r *rand.Rand) {
	ships := []ship.TypeID{"interceptor", "corvette", "frigate", "cruiser"}
	for _, st := range ships {
		shipType := ship.GetShipType(st)
		if shipType != nil {
			count := r.Intn(8) + 1
			n.EnemyFleet.AddShip(shipType, count)
		}
	}
}

// TotalResources returns the sum of all resources.
func (n *NPCPlanet) TotalResources() float64 {
	total := 0.0
	for _, v := range n.Resources {
		total += v
	}
	return total
}

// HasCombatShips returns true if the NPC planet has ships with weapons.
func (n *NPCPlanet) HasCombatShips() bool {
	return n.EnemyFleet.TotalShipCount() > 0
}

// TotalFleetStrength returns the combat strength of the enemy fleet.
func (n *NPCPlanet) TotalFleetStrength() float64 {
	snapshot := battle.NewFleetSnapshot(n.EnemyFleet)
	return snapshot.TotalDPS() + snapshot.TotalHP()*0.1
}

// ExplorationManager manages exploration expeditions and NPC planet discovery.
type ExplorationManager struct {
	npcPlanets map[string]*NPCPlanet // key = npc planet ID
	mu         int // simple counter for IDs (thread safety handled by caller)
}

// NewExplorationManager creates a new exploration manager.
func NewExplorationManager() *ExplorationManager {
	return &ExplorationManager{
		npcPlanets: make(map[string]*NPCPlanet),
	}
}

// DiscoverNPCPlanet creates a new NPC planet during exploration.
func (em *ExplorationManager) DiscoverNPCPlanet(expeditionID, playerID string) *NPCPlanet {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	npcID := expeditionID + "_npc_" + em.nextID()

	// Determine POI type based on random roll
	typeRoll := r.Intn(21)
	var poiType PointOfInterestType
	switch {
	case typeRoll < 2:
		poiType = POIAbandonedStation
	case typeRoll < 8:
		poiType = POIDebris
	case typeRoll < 10:
		poiType = POICosmicDebris
	case typeRoll < 15:
		poiType = POIAsteroids
	case typeRoll < 20:
		poiType = POIUnknownPlanet
	default:
		poiType = POIAlienBase
	}

	npc := NewNPCPlanet(npcID, playerID, poiType)
	em.npcPlanets[npcID] = npc
	return npc
}

// GetNPCPlanet returns an NPC planet by ID.
func (em *ExplorationManager) GetNPCPlanet(id string) *NPCPlanet {
	return em.npcPlanets[id]
}

// RemoveNPCPlanet removes an NPC planet (after resources collected or fleet destroyed).
func (em *ExplorationManager) RemoveNPCPlanet(id string) {
	delete(em.npcPlanets, id)
}

// GetAllNPCPlanets returns all known NPC planets.
func (em *ExplorationManager) GetAllNPCPlanets() []*NPCPlanet {
	result := make([]*NPCPlanet, 0, len(em.npcPlanets))
	for _, npc := range em.npcPlanets {
		result = append(result, npc)
	}
	return result
}

// nextID generates a unique ID suffix.
func (em *ExplorationManager) nextID() string {
	em.mu++
	return time.Now().Format("20060102150405") + "_" + string(rune(em.mu))
}

// CalculateDiscoveryChance calculates the chance of discovering an NPC planet.
// Based on expedition duration, fleet size, and elapsed time.
func CalculateDiscoveryChance(fleet *ship.Fleet, elapsedSeconds float64, totalDuration float64) float64 {
	fleetSize := float64(fleet.TotalShipCount())
	progress := elapsedSeconds / totalDuration

	// Base chance increases with progress
	baseChance := progress * 0.15 // max 15% per tick

	// Fleet size bonus
	fleetBonus := math.Min(fleetSize*0.02, 0.1) // up to 10% bonus

	// Time bonus: longer expeditions have higher chance
	timeBonus := math.Min(progress*0.2, 0.25) // up to 25% bonus

	chance := baseChance + fleetBonus + timeBonus
	return math.Min(chance, 0.5) // cap at 50%
}

// CollectResources transfers resources from NPC planet to expedition fleet cargo.
func CollectResources(npc *NPCPlanet, maxCargo float64) (map[string]float64, float64) {
	collected := make(map[string]float64)
	remaining := maxCargo

	for resName, amount := range npc.Resources {
		if amount <= 0 || remaining <= 0 {
			continue
		}
		take := math.Min(amount, remaining)
		collected[resName] = take
		npc.Resources[resName] -= take
		remaining -= take
	}

	return collected, maxCargo - remaining
}

// DeductExpeditionEnergy deducts energy from planet resources for expedition operation.
func DeductExpeditionEnergy(planetEnergy, fleetEnergy float64) (float64, bool) {
	if planetEnergy < fleetEnergy {
		return 0, false
	}
	return planetEnergy - fleetEnergy, true
}
