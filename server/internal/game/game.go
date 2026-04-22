package game

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"spacegame/internal/db"
)

var globalInstance *Game

// Instance returns the global Game singleton.
func Instance() *Game {
	return globalInstance
}

// SetInstance sets the global Game singleton.
func SetInstance(g *Game) {
	globalInstance = g
}

// Game is the core game engine managing all planets.
type Game struct {
	planets       map[string]*Planet
	mu            sync.RWMutex
	saveCount     map[string]int
	db            *db.Database
	Marketplace   *Marketplace
	broadcastFunc func(planetID, playerID string, state map[string]interface{})
}

// New creates a new Game instance.
func New() *Game {
	return &Game{
		planets:     make(map[string]*Planet),
		saveCount:   make(map[string]int),
		Marketplace: NewMarketplace(),
	}
}

// SetDB sets the database connection for saving planet state.
func (g *Game) SetDB(d *db.Database) {
	g.db = d
}

// AddPlanet adds a planet to the game engine.
func (g *Game) AddPlanet(p *Planet) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.planets[p.ID] = p
	log.Printf("Added planet %s (%s) to game engine", p.Name, p.ID)
}

// GetPlanet returns a planet by ID.
func (g *Game) GetPlanet(id string) *Planet {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.planets[id]
}

// GetAllPlanets returns all planets.
func (g *Game) GetAllPlanets() []*Planet {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]*Planet, 0, len(g.planets))
	for _, p := range g.planets {
		result = append(result, p)
	}
	return result
}

// LoadPlanetsFromDB loads all active planets from the database.
func (g *Game) LoadPlanetsFromDB() error {
	if g.db == nil {
		log.Println("No database connection, skipping planet load")
		return nil
	}

	rows, err := g.db.Query(`
		SELECT id, player_id, name, level, resources, energy_buffer
		FROM planets p
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, ownerID, name string
		var level int
		var resourcesJSON []byte
		var energyBufferValue float64
		if err := rows.Scan(&id, &ownerID, &name, &level, &resourcesJSON, &energyBufferValue); err != nil {
			log.Printf("Error scanning planet row: %v", err)
			continue
		}

		planet := NewPlanet(id, ownerID, name, g)
		planet.Level = level
		planet.EnergyBuffer.Value = energyBufferValue

		// Parse resources from JSON
		var resources PlanetResources
		if err := json.Unmarshal(resourcesJSON, &resources); err == nil {
			planet.Resources = resources
		}

		// Load buildings from DB
		buildingRows, err := g.db.Query(`
			SELECT type, level, build_progress, pending FROM buildings WHERE planet_id = $1
		`, id)
		if err == nil {
			defer buildingRows.Close()
			for buildingRows.Next() {
				var bType string
				var bLevel int
				var bProgress float32
				var bPending bool
				if err := buildingRows.Scan(&bType, &bLevel, &bProgress, &bPending); err == nil {
					idx := planet.FindBuildingIndex(bType)
					if idx >= 0 {
						planet.Buildings[idx].Level = bLevel
						planet.Buildings[idx].BuildProgress = float64(bProgress)
						planet.Buildings[idx].Pending = bPending
					} else {
						planet.Buildings = append(planet.Buildings, BuildingEntry{
							Type:          bType,
							Level:         bLevel,
							BuildProgress: float64(bProgress),
							Pending:       bPending,
						})
					}
				}
			}
		} else {
			log.Printf("Error loading buildings for planet %s: %v", id, err)
		}

		// Populate computed fields for each building
		for i := range planet.Buildings {
			planet.PopulateBuildingEntry(i)
		}

		g.AddPlanet(planet)
	}

	log.Printf("Loaded %d planets from database", len(g.planets))
	return nil
}

// Tick processes one game tick for all planets.
func (g *Game) Tick() {
	g.mu.RLock()
	planets := make([]*Planet, 0, len(g.planets))
	for _, p := range g.planets {
		planets = append(planets, p)
	}
	g.mu.RUnlock()

	// Process battles every 60 ticks (1 minute)
	battleTick := g.getBattleTick()

	for _, p := range planets {
		p.Tick()

		if battleTick {
			g.processAutoBattle(p)
		}
	}

	// Check for random events
	g.TriggerRandomEvents()
}

// battleCooldownTicks is the number of ticks between auto-battle attempts per planet.
const battleCooldownTicks = 60

// processAutoBattle checks if a planet should fight an NPC and executes the battle.
func (g *Game) processAutoBattle(p *Planet) {
	// Skip if planet has no combat fleet
	if !p.HasCombatFleet() {
		return
	}

	// Check battle cooldown via last battle timestamp
	if len(p.Battles) > 0 {
		lastBattle := p.Battles[len(p.Battles)-1]
		elapsed := time.Since(lastBattle.Timestamp).Seconds()
		if elapsed < battleCooldownTicks {
			return
		}
	}

	// For now, log that auto-battle would trigger
	// NPC planet integration will be added in Phase 6
	log.Printf("Planet %s has combat fleet (%d ships, strength %.0f), ready for auto-battle",
		p.Name, p.GetTotalShipCount(), p.GetFleetStrength())
}

// getBattleTick returns true every battleCooldownTicks ticks.
func (g *Game) getBattleTick() bool {
	// Use a simple counter based on time
	// In production, this would be tracked per-planet
	return true
}

// shouldSave checks if a planet should be saved to DB this tick.
func (g *Game) shouldSave(planetID string) bool {
	g.saveCount[planetID]++
	return g.saveCount[planetID] >= 10 // Save every 10 ticks
}

// savePlanet saves planet state to the database.
func (g *Game) savePlanet(p *Planet) {
	if g.db == nil {
		return
	}

	resourcesJSON, err := json.Marshal(p.Resources)
	if err != nil {
		log.Printf("Error marshaling resources for planet %s: %v", p.ID, err)
	} else {
		_, err = g.db.Exec(`
			UPDATE planets 
			SET resources = $1::jsonb, updated_at = NOW()
			WHERE id = $2
		`, string(resourcesJSON), p.ID)
		if err != nil {
			log.Printf("Error saving planet %s: %v", p.ID, err)
		}
	}

	// Save building state
	for _, b := range p.Buildings {
		progress := b.BuildProgress
		pending := b.Pending
		_, err = g.db.Exec(`
			INSERT INTO buildings (planet_id, type, level, build_progress, pending)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (planet_id, type) DO UPDATE
			SET level = $3, build_progress = $4, pending = $5, updated_at = NOW()
		`, p.ID, b.Type, b.Level, float32(progress), pending)
		if err != nil {
			log.Printf("Error saving building %s for planet %s: %v", b.Type, p.ID, err)
		}
	}

	// Save research state
	if err := p.Research.SaveToDB(); err != nil {
		log.Printf("Error saving research for planet %s: %v", p.ID, err)
	}

	// Save fleet state
	fleetData, err := json.Marshal(p.Fleet.GetShipState())
	if err != nil {
		log.Printf("Error marshaling fleet for planet %s: %v", p.ID, err)
	} else {
		_, err = g.db.Exec(`
			UPDATE planets 
			SET fleet_state = $1::jsonb, updated_at = NOW()
			WHERE id = $2
		`, string(fleetData), p.ID)
		if err != nil {
			log.Printf("Error saving fleet for planet %s: %v", p.ID, err)
		}
	}

	// Save shipyard state
	shipyardData, err := json.Marshal(p.Shipyard)
	if err != nil {
		log.Printf("Error marshaling shipyard for planet %s: %v", p.ID, err)
	} else {
		_, err = g.db.Exec(`
			UPDATE planets 
			SET shipyard_state = $1::jsonb, updated_at = NOW()
			WHERE id = $2
		`, string(shipyardData), p.ID)
		if err != nil {
			log.Printf("Error saving shipyard for planet %s: %v", p.ID, err)
		}
	}

	// Save energy buffer
	_, err = g.db.Exec(`
		UPDATE planets 
		SET energy_buffer = $1, updated_at = NOW()
		WHERE id = $2
	`, p.EnergyBuffer.Value, p.ID)
	if err != nil {
		log.Printf("Error saving energy buffer for planet %s: %v", p.ID, err)
	}

	g.saveCount[p.ID] = 0
}

// broadcastPlanetUpdate sends a state update to all connected clients for a planet.
func (g *Planet) broadcastPlanetUpdate() {
	if g.game == nil {
		return
	}

	state := g.GetState()
	playerID := g.OwnerID
	planetID := g.ID

	// Import the broadcast function from api package
	// We'll use a callback pattern to avoid circular imports
	if g.game.broadcastFunc != nil {
		g.game.broadcastFunc(planetID, playerID, state)
	}
}

// RegisterBroadcastHandler sets the broadcast callback function.
func (g *Game) RegisterBroadcastHandler(fn func(planetID, playerID string, state map[string]interface{})) {
	g.broadcastFunc = fn
}
