package game

import (
	"spacegame/internal/game/battle"
)

// GetBattleHistory returns the battle history.
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

// HasCombatFleet returns true if the planet has ships with weapons.
func (p *Planet) HasCombatFleet() bool {
	snapshot := p.GetFleetSnapshot()
	return snapshot.HasCombatShips()
}
