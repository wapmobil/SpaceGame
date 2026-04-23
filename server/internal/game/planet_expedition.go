package game

import (
	"log"
	"math/rand"
	"time"

	"spacegame/internal/game/battle"
	"spacegame/internal/game/expedition"
	"spacegame/internal/game/ship"
)

// CanStartExpedition checks if the planet can start a new expedition.
func (p *Planet) CanStartExpedition(expType expedition.Type, fleet *ship.Fleet) error {
	if !p.BaseOperational() {
		return &PlanetError{PlanetID: p.ID, Reason: "base_not_operational", Extra: "Planet base requires food to operate. Produce food to start expeditions."}
	}

	// Check if expeditions research is completed
	if _, ok := p.Research.GetCompleted()["expeditions"]; !ok {
		return &PlanetError{PlanetID: p.ID, Reason: "expeditions_not_researched"}
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
		return &PlanetError{PlanetID: p.ID, Reason: "max_expeditions_reached"}
	}

	// Check fleet has ships
	if fleet.TotalShipCount() == 0 {
		return &PlanetError{PlanetID: p.ID, Reason: "no_ships_available"}
	}

	// Check energy
	energyCost := fleet.TotalEnergyConsumption()
	if p.EnergyBuffer.Value < energyCost {
		return &PlanetError{PlanetID: p.ID, Reason: "insufficient_energy"}
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
	p.EnergyBuffer.Value -= expFleetEnergy
	if p.EnergyBuffer.Value < 0 {
		p.EnergyBuffer.Value = 0
	}

	// Create expedition
	id := p.ID + "_exp_" + time.Now().Format("20060102150405")
	exp := expedition.CreateExpedition(id, p.ID, target, expType, expFleet, duration)

	p.Expeditions = append(p.Expeditions, exp)

	log.Printf("Planet %s started %s expedition with %d ships, duration: %.0fs",
		p.Name, expType, expFleet.TotalShipCount(), duration)

	return exp, nil
}

// TickExpeditions advances all active expeditions.
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
		return &PlanetError{PlanetID: p.ID, Reason: "expedition_not_found"}
	}

	if exp.Status != expedition.StatusAtPoint {
		return &PlanetError{PlanetID: p.ID, Reason: "expedition_not_at_point"}
	}

	if exp.DiscoveredNPC == nil {
		return &PlanetError{PlanetID: p.ID, Reason: "no_npc_discovered"}
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
		return &PlanetError{PlanetID: p.ID, Reason: "unknown_action"}
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
		return &PlanetError{PlanetID: p.ID, Reason: "no_combat_ships"}
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
				"id":              exp.DiscoveredNPC.ID,
				"name":            exp.DiscoveredNPC.Name,
				"type":            exp.DiscoveredNPC.Type,
				"resources":       exp.DiscoveredNPC.Resources,
				"total_resources": exp.DiscoveredNPC.TotalResources(),
				"has_combat":      exp.DiscoveredNPC.HasCombatShips(),
				"fleet_strength":  exp.DiscoveredNPC.TotalFleetStrength(),
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
