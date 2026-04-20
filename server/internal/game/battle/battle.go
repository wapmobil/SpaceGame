package battle

import (
	"math"
	"math/rand"
)

// BattleResult holds the outcome of a battle.
type BattleResult struct {
	Winner       string            // "attacker", "defender", "draw"
	Rounds       int
	AttackerLost map[string]int    // type -> count destroyed
	DefenderLost map[string]int
	AttackerLoot map[string]float64 // resources won by attacker
	AttackerRefund map[string]float64 // 50% cost return for attacker's lost ships
	DefenderRefund map[string]float64 // 50% cost return for defender's lost ships
}

// BattleState holds the mutable state during battle calculation.
type BattleState struct {
	attacker *FleetSnapshot
	defender *FleetSnapshot
	rounds   int
}

// CalculateBattle simulates an auto-battle between two fleets.
// Returns the result after all rounds are resolved.
func CalculateBattle(attacker, defender *FleetSnapshot) *BattleResult {
	result := &BattleResult{
		AttackerLost:   make(map[string]int),
		DefenderLost:   make(map[string]int),
		AttackerLoot:   make(map[string]float64),
		AttackerRefund: make(map[string]float64),
		DefenderRefund: make(map[string]float64),
	}

	// Handle empty fleets
	if attacker.IsEmpty() && defender.IsEmpty() {
		result.Winner = "draw"
		result.Rounds = 0
		return result
	}

	if attacker.IsEmpty() {
		result.Winner = "defender"
		result.Rounds = 0
		return result
	}

	if defender.IsEmpty() {
		result.Winner = "attacker"
		result.Rounds = 0
		result.AttackerLoot = calculateLoot(defender)
		return result
	}

	// Check if both have no weapons
	attackerHasWeapons := attacker.HasCombatShips()
	defenderHasWeapons := defender.HasCombatShips()

	if !attackerHasWeapons && !defenderHasWeapons {
		result.Winner = "draw"
		result.Rounds = 0
		return result
	}

	state := &BattleState{
		attacker: attacker.Clone(),
		defender: defender.Clone(),
	}

	maxRounds := 50 // Prevent infinite loops
	for state.rounds < maxRounds {
		state.rounds++

		// Both sides attack before checking for destruction
		// This ensures fair combat when fleets are equal

		// Attacker attacks defender (only if attacker has weapons)
		var attackerDmg float64
		if attackerHasWeapons {
			attackerHits := calculateHits(state.attacker, state.defender)
			defenderHP := state.defender.TotalHP()

			if defenderHP > 0 && attackerHits > 0 {
				attackerDmg = attackerHits * avgWeaponDamage(state.attacker)
				_, attackerDestroyed := state.defender.ApplyDamage(attackerDmg)
				for typeID, count := range attackerDestroyed {
					result.DefenderLost[typeID] += count
				}
			}
		}

		// Defender attacks attacker (only if defender has weapons)
		var defenderDmg float64
		if defenderHasWeapons {
			defenderHits := calculateHits(state.defender, state.attacker)
			attackerHP := state.attacker.TotalHP()

			if attackerHP > 0 && defenderHits > 0 {
				defenderDmg = defenderHits * avgWeaponDamage(state.defender)
				_, defenderDestroyed := state.attacker.ApplyDamage(defenderDmg)
				for typeID, count := range defenderDestroyed {
					result.AttackerLost[typeID] += count
				}
			}
		}

		// Check for draw (both destroyed in same round)
		if state.attacker.IsEmpty() && state.defender.IsEmpty() {
			result.Winner = "draw"
			break
		}

		// Check if defender is destroyed
		if state.defender.IsEmpty() {
			result.Winner = "attacker"
			result.AttackerLoot = calculateLoot(state.defender)
			break
		}

		// Check if attacker is destroyed
		if state.attacker.IsEmpty() {
			result.Winner = "defender"
			break
		}
	}

	// If battle exceeded max rounds, declare based on remaining strength
	if state.rounds >= maxRounds {
		if state.attacker.TotalHP() > state.defender.TotalHP() {
			result.Winner = "attacker"
			result.AttackerLoot = calculateLoot(state.defender)
		} else if state.defender.TotalHP() > state.attacker.TotalHP() {
			result.Winner = "defender"
		} else {
			result.Winner = "draw"
		}
	}

	result.Rounds = state.rounds

	// Calculate refunds for destroyed ships
	if len(result.AttackerLost) > 0 {
		result.AttackerRefund = GetDestroyedShipCosts(result.AttackerLost)
	}
	if len(result.DefenderLost) > 0 {
		result.DefenderRefund = GetDestroyedShipCosts(result.DefenderLost)
	}

	return result
}

// calculateHits determines how many hits a fleet lands on target.
// Formula: hits = totalDPS / (targetTotalHP × 0.1)
// This creates a scaling where stronger fleets hit more ships per round.
func calculateHits(fleet *FleetSnapshot, target *FleetSnapshot) float64 {
	targetHP := target.TotalHP()
	if targetHP <= 0 {
		return 0
	}

	dps := fleet.TotalDPS()
	hits := dps / (targetHP * 0.1)

	// Add small randomness (±20%)
	randomness := 0.8 + rand.Float64()*0.4
	hits *= randomness

	return math.Max(1, hits) // Minimum 1 hit if fleet has weapons
}

// avgWeaponDamage returns the average weapon damage per ship in the fleet.
func avgWeaponDamage(fleet *FleetSnapshot) float64 {
	totalShips := fleet.TotalShipCount()
	if totalShips == 0 {
		return 0
	}

	totalDmg := 0.0
	for _, sc := range fleet.Ships {
		if sc.Count > 0 && (sc.WeaponMin > 0 || sc.WeaponMax > 0) {
			totalDmg += float64(sc.Count) * (sc.WeaponMin + sc.WeaponMax) / 2
		}
	}

	return totalDmg / float64(totalShips)
}

// calculateLoot determines what resources the attacker wins.
func calculateLoot(defender *FleetSnapshot) map[string]float64 {
	loot := make(map[string]float64)

	// Base loot: money based on fleet strength
	moneyLoot := defender.TotalHP() * 2
	loot["money"] = moneyLoot

	// Alien tech chance based on cruiser count
	cruisers := 0
	if sc, ok := defender.Ships["cruiser"]; ok {
		cruisers = sc.Count
	}
	if cruisers > 0 {
		loot["alien_tech"] = float64(cruisers) * 2
	}

	// Small resource loot
	loot["food"] = moneyLoot * 0.1
	loot["composite"] = moneyLoot * 0.05
	loot["mechanisms"] = moneyLoot * 0.05
	loot["reagents"] = moneyLoot * 0.05

	return loot
}
