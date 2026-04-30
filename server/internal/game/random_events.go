package game

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// RandomEventType defines the type of random event.
type RandomEventType string

const (
	RandomEventShortCircuit    RandomEventType = "short_circuit"
	RandomEventTheft           RandomEventType = "theft"
	RandomEventStorageCollapse RandomEventType = "storage_collapse"
	RandomEventMineCollapse    RandomEventType = "mine_collapse"
)

// RandomEvent represents a triggered random event.
type RandomEvent struct {
	Type        RandomEventType `json:"type"`
	Description string          `json:"description"`
	PlanetID    string          `json:"planet_id"`
	PlayerID    string          `json:"player_id"`
	Timestamp   time.Time       `json:"timestamp"`
	Resolved    bool            `json:"resolved"`
	ResolveCost map[string]float64 `json:"resolve_cost,omitempty"`
}

// EventDef defines a random event configuration.
type EventDef struct {
	Type        RandomEventType
	Chance      float64       // probability per tick (e.g., 0.02 = 2%)
	Description string
	ResolveCost map[string]float64
	ApplyFn     func(p *Planet) (string, error)
}

// GetRandomEvents returns the list of configured random events.
func GetRandomEvents() []EventDef {
	return []EventDef{
		{
			Type:        RandomEventShortCircuit,
			Chance:      0, // disabled — for future special modes
			Description: "Short Circuit: Energy production disrupted for 1 tick. Pay resources to fix.",
			ResolveCost: map[string]float64{
				"money": 100,
			},
			ApplyFn: applyShortCircuit,
		},
		{
			Type:        RandomEventTheft,
			Chance:      0, // disabled
			Description: "Theft: Lost 5-20% of money to space pirates.",
			ResolveCost: map[string]float64{},
			ApplyFn:     applyTheft,
		},
		{
			Type:        RandomEventStorageCollapse,
			Chance:      0, // disabled
			Description: "Storage Roof Collapse: Lost 5-20% of stored resources.",
			ResolveCost: map[string]float64{
				"money": 50,
			},
			ApplyFn: applyStorageCollapse,
		},
	}
}

// TriggerRandomEvents checks all planets for random events and applies them.
func (g *Game) TriggerRandomEvents() {
	g.mu.RLock()
	planets := make([]*Planet, 0, len(g.planets))
	for _, p := range g.planets {
		planets = append(planets, p)
	}
	g.mu.RUnlock()

	events := GetRandomEvents()

	for _, p := range planets {
		for _, eventDef := range events {
			if rand.Float64() < eventDef.Chance {
				log.Printf("Random event triggered on planet %s: %s", p.ID, eventDef.Description)

				description, err := eventDef.ApplyFn(p)
				if err != nil {
					log.Printf("Error applying event %s to planet %s: %v", eventDef.Type, p.ID, err)
					continue
				}

				// Log event to database
				if g.db != nil {
					_, err := g.db.Exec(
						"INSERT INTO events (planet_id, player_id, event_type, description) VALUES ($1, $2, $3, $4)",
						p.ID, p.OwnerID, string(eventDef.Type), description,
					)
					if err != nil {
						log.Printf("Error logging event for planet %s: %v", p.ID, err)
					} else {
						log.Printf("Event logged for planet %s: %s", p.ID, string(eventDef.Type))
					}
				}

				// Broadcast event via WebSocket
				if g.broadcastFunc != nil {
					state := p.GetState()
					g.broadcastFunc(p.ID, p.OwnerID, state)
				}
			}
		}
	}
}

// ResolveEvent attempts to resolve a random event by paying the required cost.
func (g *Game) ResolveEvent(planetID string, eventType string) (string, error) {
	p := g.GetPlanet(planetID)
	if p == nil {
		return "", &PlanetError{PlanetID: planetID, Reason: "planet_not_found"}
	}

	events := GetRandomEvents()
	var eventDef *EventDef
	for i := range events {
		if events[i].Type == RandomEventType(eventType) {
			eventDef = &events[i]
			break
		}
	}

	if eventDef == nil {
		return "", &PlanetError{PlanetID: planetID, Reason: "unknown_event_type"}
	}

	// Check if player can afford the resolve cost and pay
	for resource, cost := range eventDef.ResolveCost {
		switch resource {
		case "money":
			if p.Resources.Money < cost {
				return fmt.Sprintf("Cannot resolve event: need %.0f money, have %.0f", cost, p.Resources.Money), nil
			}
			p.Resources.Money -= cost
		case "food":
			if p.Resources.Food < cost {
				return fmt.Sprintf("Cannot resolve event: need %.0f food, have %.0f", cost, p.Resources.Food), nil
			}
			p.Resources.Food -= cost
		case "composite":
			if p.Resources.Composite < cost {
				return fmt.Sprintf("Cannot resolve event: need %.0f composite, have %.0f", cost, p.Resources.Composite), nil
			}
			p.Resources.Composite -= cost
		case "mechanisms":
			if p.Resources.Mechanisms < cost {
				return fmt.Sprintf("Cannot resolve event: need %.0f mechanisms, have %.0f", cost, p.Resources.Mechanisms), nil
			}
			p.Resources.Mechanisms -= cost
		case "reagents":
			if p.Resources.Reagents < cost {
				return fmt.Sprintf("Cannot resolve event: need %.0f reagents, have %.0f", cost, p.Resources.Reagents), nil
			}
			p.Resources.Reagents -= cost
		}
	}

	// Resolve the event effects
	switch eventDef.Type {
	case RandomEventShortCircuit:
		// Restore energy production
		energyStorageLevel := 0
		for _, b := range p.Buildings {
			if b.Type == "energy_storage" {
				energyStorageLevel += b.Level
			}
		}
		p.EnergyBuffer.UpdateMax(energyStorageLevel)
		p.Resources.MaxEnergy = p.EnergyBuffer.Max
		production, consumption := p.calculateEnergy()
		if production >= consumption {
			p.EnergyBuffer.Value = p.Resources.MaxEnergy
		}
	case RandomEventTheft:
		// No recovery needed - money is lost
	case RandomEventStorageCollapse:
		// No recovery needed - resources are lost
	case RandomEventMineCollapse:
		// Level is reduced, player can rebuild
	}

	// Log resolved event in database
	if g.db != nil {
		_, err := g.db.Exec(
			"UPDATE events SET resolved = TRUE, resolved_at = NOW() WHERE planet_id = $1 AND event_type = $2 AND resolved = FALSE ORDER BY created_at DESC LIMIT 1",
			planetID, eventType,
		)
		if err != nil {
			log.Printf("Error updating event as resolved: %v", err)
		}
	}

	return fmt.Sprintf("Event resolved: %s. Paid %.0f money in repair costs.", eventDef.Description, eventDef.ResolveCost["money"]), nil
}

func applyShortCircuit(p *Planet) (string, error) {
	p.EnergyBuffer.Value = 0
	p.Resources.MaxEnergy = 0
	return fmt.Sprintf("Short circuit: Energy production disrupted for 1 tick. Pay resources to fix."), nil
}

func applyTheft(p *Planet) (string, error) {
	// Lose 5-20% of money
	lossPercent := 0.05 + rand.Float64()*0.15 // 5% to 20%
	loss := p.Resources.Money * lossPercent
	p.Resources.Money -= loss
	if p.Resources.Money < 0 {
		p.Resources.Money = 0
	}
	return fmt.Sprintf("Theft: Lost %.0f money (%.0f%%) to space pirates.", loss, lossPercent*100), nil
}

func applyStorageCollapse(p *Planet) (string, error) {
	// Lose 5-20% of a random stored resource
	resources := []struct {
		name  string
		value *float64
	}{
		{"food", &p.Resources.Food},
		{"composite", &p.Resources.Composite},
		{"mechanisms", &p.Resources.Mechanisms},
		{"reagents", &p.Resources.Reagents},
	}

	// Only consider resources that have some amount
	var available []struct {
		name  string
		value *float64
	}
	for _, r := range resources {
		if *r.value > 0 {
			available = append(available, r)
		}
	}

	if len(available) == 0 {
		return "Storage collapse: No resources to lose.", nil
	}

	// Pick a random resource
	selected := available[rand.Intn(len(available))]
	lossPercent := 0.05 + rand.Float64()*0.15
	loss := *selected.value * lossPercent
	*selected.value -= loss
	if *selected.value < 0 {
		*selected.value = 0
	}

	return fmt.Sprintf("Storage collapse: Lost %.0f %s (%.0f%%) due to roof collapse.", loss, selected.name, lossPercent*100), nil
}
