package game

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"sync"

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
	tickCount     int64
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

// LoadPlanetFromDB loads a single planet from the database and registers it.
func (g *Game) LoadPlanetFromDB(planetID string) error {
	if g.db == nil {
		return nil
	}

	hasEnergyBuffer, err := g.db.ColumnExists(context.Background(), "planets", "energy_buffer")
	if err != nil {
		hasEnergyBuffer = false
	}

	var query string
	if hasEnergyBuffer {
		query = `SELECT id, player_id, name, level, resources, energy_buffer FROM planets WHERE id = $1`
	} else {
		query = `SELECT id, player_id, name, level, resources FROM planets WHERE id = $1`
	}

	var id, ownerID, name string
	var level int
	var resourcesJSON []byte
	var energyBufferValue *float64

	if hasEnergyBuffer {
		err = g.db.QueryRow(query, planetID).Scan(&id, &ownerID, &name, &level, &resourcesJSON, &energyBufferValue)
	} else {
		err = g.db.QueryRow(query, planetID).Scan(&id, &ownerID, &name, &level, &resourcesJSON)
	}

	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	planet := NewPlanet(id, ownerID, name, g)
	planet.Level = level
	if energyBufferValue != nil {
		planet.EnergyBuffer.Value = *energyBufferValue
	}

	var resources PlanetResources
	if err := json.Unmarshal(resourcesJSON, &resources); err == nil {
		planet.Resources = resources
	}
	

	buildingRows, err := g.db.Query(`
		SELECT type, level, build_progress, enabled FROM buildings WHERE planet_id = $1
	`, planetID)
	if err == nil {
		for buildingRows.Next() {
			var bType string
			var bLevel int
			var bProgress float32
			var bEnabled bool
			if err := buildingRows.Scan(&bType, &bLevel, &bProgress, &bEnabled); err == nil {
				idx := planet.FindBuildingIndex(bType)
				if idx >= 0 {
					planet.Buildings[idx].Level = bLevel
					planet.Buildings[idx].BuildProgress = float64(bProgress)
					planet.Buildings[idx].Enabled = bEnabled
				} else {
					planet.Buildings = append(planet.Buildings, BuildingEntry{
						Type:          bType,
						Level:         bLevel,
						BuildProgress: float64(bProgress),
						Enabled:       bEnabled,
					})
				}
			}
		}
		buildingRows.Close()
	}

	for i := range planet.Buildings {
		planet.PopulateBuildingEntry(i)
	}

	// If planet_exploration is completed but no random unlock, generate one
	if planet.Research.Completed["planet_exploration"] > 0 && planet.Resources.ResearchUnlocks == "" {
		buildings := []string{"composite_drone", "mechanism_factory", "reagent_lab"}
		planet.Resources.ResearchUnlocks = buildings[rand.Intn(len(buildings))]
	}

	// Load farm state
	if err := LoadFarmFromDB(planet); err != nil {
		log.Printf("Error loading farm for planet %s: %v", planet.ID, err)
	}

	g.AddPlanet(planet)
	return nil
}

// LoadPlanetsFromDB loads all active planets from the database.
func (g *Game) LoadPlanetsFromDB() error {
	if g.db == nil {
		log.Println("No database connection, skipping planet load")
		return nil
	}

	hasEnergyBuffer, err := g.db.ColumnExists(context.Background(), "planets", "energy_buffer")
	if err != nil {
		hasEnergyBuffer = false
	}
	var query string
	if hasEnergyBuffer {
		query = `SELECT id, player_id, name, level, resources, energy_buffer FROM planets p ORDER BY updated_at DESC`
	} else {
		query = `SELECT id, player_id, name, level, resources FROM planets p ORDER BY updated_at DESC`
	}

	rows, err := g.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, ownerID, name string
		var level int
		var resourcesJSON []byte
		var energyBufferValue *float64

		if hasEnergyBuffer {
			if err := rows.Scan(&id, &ownerID, &name, &level, &resourcesJSON, &energyBufferValue); err != nil {
				log.Printf("Error scanning planet row: %v", err)
				continue
			}
		} else {
			if err := rows.Scan(&id, &ownerID, &name, &level, &resourcesJSON); err != nil {
				log.Printf("Error scanning planet row: %v", err)
				continue
			}
		}

		planet := NewPlanet(id, ownerID, name, g)
		planet.Level = level
		if energyBufferValue != nil {
			planet.EnergyBuffer.Value = *energyBufferValue
		}

		// Parse resources from JSON
		var resources PlanetResources
		if err := json.Unmarshal(resourcesJSON, &resources); err == nil {
			planet.Resources = resources
		}
		

		// Load buildings from DB
		buildingRows, err := g.db.Query(`
			SELECT type, level, build_progress, enabled FROM buildings WHERE planet_id = $1
		`, id)
		if err == nil {
			for buildingRows.Next() {
				var bType string
				var bLevel int
				var bProgress float32
				var bEnabled bool
				if err := buildingRows.Scan(&bType, &bLevel, &bProgress, &bEnabled); err == nil {
					idx := planet.FindBuildingIndex(bType)
					if idx >= 0 {
						planet.Buildings[idx].Level = bLevel
						planet.Buildings[idx].BuildProgress = float64(bProgress)
						planet.Buildings[idx].Enabled = bEnabled
					} else {
						planet.Buildings = append(planet.Buildings, BuildingEntry{
							Type:          bType,
							Level:         bLevel,
							BuildProgress: float64(bProgress),
							Enabled:       bEnabled,
						})
					}
				}
			}
			buildingRows.Close()
		} else {
			log.Printf("Error loading buildings for planet %s: %v", id, err)
		}

		// Populate computed fields for each building
		for i := range planet.Buildings {
			planet.PopulateBuildingEntry(i)
		}

		// Load farm state
		if err := LoadFarmFromDB(planet); err != nil {
			log.Printf("Error loading farm for planet %s: %v", id, err)
		}

		g.AddPlanet(planet)
	}

	log.Printf("Loaded %d planets from database", len(g.planets))
	return nil
}

// Tick processes one game tick for all planets.
func (g *Game) Tick() {
	g.mu.Lock()
	g.tickCount++
	tick := g.tickCount
	g.mu.Unlock()

	g.mu.RLock()
	planets := make([]*Planet, 0, len(g.planets))
	for _, p := range g.planets {
		planets = append(planets, p)
	}
	g.mu.RUnlock()

	for _, p := range planets {
		p.Tick(tick)
	}

	// Check for random events
	g.TriggerRandomEvents()
}

// GetTotalMarketLevel returns the sum of all market building levels across all planets.
func (g *Game) GetTotalMarketLevel() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	total := 0
	for _, p := range g.planets {
		total += p.GetBuildingLevel("market")
	}
	return total
}

// shouldSave checks if a planet should be saved to DB this tick.
func (g *Game) shouldSave(planetID string) bool {
	g.saveCount[planetID]++
	return g.saveCount[planetID] >= 10 // Save every 10 ticks
}

// dbPlanetResources is the resource struct for database storage (no energy fields).
type dbPlanetResources struct {
	Food            float64 `json:"food"`
	Iron            float64 `json:"iron"`
	Composite       float64 `json:"composite"`
	Mechanisms      float64 `json:"mechanisms"`
	Reagents        float64 `json:"reagents"`
	Money           float64 `json:"money"`
	AlienTech       float64 `json:"alien_tech"`
	StorageCapacity float64 `json:"storage_capacity"`
	ResearchUnlocks string  `json:"research_unlocks"`
}

// savePlanet saves planet state to the database.
func (g *Game) savePlanet(p *Planet) {
	if g.db == nil {
		return
	}

	dbResources := dbPlanetResources{
		Food:            p.Resources.Food,
		Iron:            p.Resources.Iron,
		Composite:       p.Resources.Composite,
		Mechanisms:      p.Resources.Mechanisms,
		Reagents:        p.Resources.Reagents,
		Money:           p.Resources.Money,
		AlienTech:       p.Resources.AlienTech,
		StorageCapacity: p.Resources.StorageCapacity,
		ResearchUnlocks: p.Resources.ResearchUnlocks,
	}
	resourcesJSON, err := json.Marshal(dbResources)
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
		enabled := b.Enabled
		_, err = g.db.Exec(`
			INSERT INTO buildings (planet_id, type, level, build_progress, enabled)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (planet_id, type) DO UPDATE
			SET level = $3, build_progress = $4, enabled = $5, updated_at = NOW()
		`, p.ID, b.Type, b.Level, float32(progress), enabled)
		if err != nil {
			log.Printf("Error saving building %s for planet %s: %v", b.Type, p.ID, err)
		}
	}

	// Save research state
	if err := p.Research.SaveToDB(); err != nil {
		log.Printf("Error saving research for planet %s: %v", p.ID, err)
	}

	// Save farm state
	if p.FarmState != nil && p.FarmState.RowCount > 0 {
		SaveFarmToDB(p)
	}

	// Save fleet state
	if fleetState, _ := g.db.ColumnExists(context.Background(), "planets", "fleet_state"); fleetState {
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
	}

	// Save shipyard state
	if shipyardState, _ := g.db.ColumnExists(context.Background(), "planets", "shipyard_state"); shipyardState {
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
	}

	// Save energy buffer
	if energyBuf, _ := g.db.ColumnExists(context.Background(), "planets", "energy_buffer"); energyBuf {
		_, err = g.db.Exec(`
			UPDATE planets 
			SET energy_buffer = $1, updated_at = NOW()
			WHERE id = $2
		`, p.EnergyBuffer.Value, p.ID)
		if err != nil {
			log.Printf("Error saving energy buffer for planet %s: %v", p.ID, err)
		}
	}

	// Save farm state
	if farmGrid, _ := g.db.ColumnExists(context.Background(), "planets", "farm_grid"); farmGrid {
		if p.FarmState != nil {
			farmData, err := json.Marshal(p.FarmState.Rows)
			if err != nil {
				log.Printf("Error marshaling farm for planet %s: %v", p.ID, err)
			} else {
				_, err = g.db.Exec(`
					UPDATE planets 
					SET farm_grid = $1::jsonb, farm_last_tick = $2, updated_at = NOW()
					WHERE id = $3
				`, string(farmData), p.FarmState.LastTick, p.ID)
				if err != nil {
					log.Printf("Error saving farm for planet %s: %v", p.ID, err)
				}
			}
		}
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

// SavePlanet saves planet state to the database.
func (g *Game) SavePlanet(p *Planet) {
	g.savePlanet(p)
}
