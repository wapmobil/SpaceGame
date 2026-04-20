package battle

import (
	"math"

	"spacegame/internal/game/ship"
)

// FleetSnapshot holds immutable fleet data for battle calculations.
type FleetSnapshot struct {
	Ships map[string]*ShipCount
}

// ShipCount represents a group of identical ships.
type ShipCount struct {
	Type      string
	Count     int
	HP        float64
	Armor     float64
	WeaponMin float64
	WeaponMax float64
	TotalHP   float64
	TotalDmg  float64
	Cost      ship.Cost
}

// NewFleetSnapshot creates a FleetSnapshot from a ship.Fleet.
func NewFleetSnapshot(f *ship.Fleet) *FleetSnapshot {
	fs := &FleetSnapshot{
		Ships: make(map[string]*ShipCount),
	}

	for typeID, fship := range f.Ships {
		st := ship.GetShipType(ship.TypeID(typeID))
		if st == nil || fship.Count <= 0 {
			continue
		}

		sc := &ShipCount{
			Type:      string(st.TypeID),
			Count:     fship.Count,
			HP:        st.HP,
			Armor:     st.Armor,
			WeaponMin: st.WeaponMinDmg,
			WeaponMax: st.WeaponMaxDmg,
			TotalHP:   float64(fship.Count) * st.HP,
			TotalDmg:  float64(fship.Count) * (st.WeaponMinDmg + st.WeaponMaxDmg) / 2,
			Cost:      st.Cost,
		}

		fs.Ships[typeID] = sc
	}

	return fs
}

// TotalShipCount returns the total number of ships.
func (fs *FleetSnapshot) TotalShipCount() int {
	total := 0
	for _, sc := range fs.Ships {
		total += sc.Count
	}
	return total
}

// TotalHP returns the sum of all ship HP.
func (fs *FleetSnapshot) TotalHP() float64 {
	total := 0.0
	for _, sc := range fs.Ships {
		total += sc.TotalHP
	}
	return total
}

// TotalDPS returns the average damage per round.
func (fs *FleetSnapshot) TotalDPS() float64 {
	total := 0.0
	for _, sc := range fs.Ships {
		total += sc.TotalDmg
	}
	return total
}

// HasCombatShips returns true if the fleet has any ships with weapons.
func (fs *FleetSnapshot) HasCombatShips() bool {
	for _, sc := range fs.Ships {
		if sc.WeaponMax > 0 || sc.WeaponMin > 0 {
			return true
		}
	}
	return false
}

// Clone creates a deep copy of the FleetSnapshot.
func (fs *FleetSnapshot) Clone() *FleetSnapshot {
	clone := &FleetSnapshot{
		Ships: make(map[string]*ShipCount),
	}

	for typeID, sc := range fs.Ships {
		scClone := *sc
		clone.Ships[typeID] = &scClone
	}

	return clone
}

// ApplyDamage distributes damage across ship types proportionally to their HP.
func (fs *FleetSnapshot) ApplyDamage(damage float64) (actualDamage float64, destroyed map[string]int) {
	actualDamage = 0
	destroyed = make(map[string]int)

	if damage <= 0 {
		return 0, destroyed
	}

	// Calculate how much damage each ship type can absorb
	totalHP := fs.TotalHP()
	if totalHP <= 0 {
		return 0, destroyed
	}

	remainingDamage := damage

	// Distribute damage proportionally
	for typeID, sc := range fs.Ships {
		if sc.Count <= 0 || sc.TotalHP <= 0 {
			continue
		}

		// This type's share of the damage
		typeShare := sc.TotalHP / totalHP
		typeDamage := remainingDamage * typeShare

		// Actual damage after armor (armor reduces damage to minimum 0)
		armorReduction := math.Min(sc.Armor, typeDamage)
		typeDamage = math.Max(0, typeDamage-armorReduction)

		actualDamage += typeDamage

		// Calculate how many ships of this type are destroyed
		shipsDestroyed := int(math.Floor(typeDamage / sc.HP))
		shipsDestroyed = min(shipsDestroyed, sc.Count)

		if shipsDestroyed > 0 {
			destroyed[typeID] = shipsDestroyed
			sc.Count -= shipsDestroyed
			sc.TotalHP = float64(sc.Count) * sc.HP
			sc.TotalDmg = float64(sc.Count) * (sc.WeaponMin + sc.WeaponMax) / 2
		}

		remainingDamage -= typeDamage
	}

	return actualDamage, destroyed
}

// RemoveShips removes ships from the fleet (for recycling).
func (fs *FleetSnapshot) RemoveShips(destroyed map[string]int) {
	for typeID, count := range destroyed {
		if sc, ok := fs.Ships[typeID]; ok {
			sc.Count -= count
			sc.TotalHP = float64(sc.Count) * sc.HP
			sc.TotalDmg = float64(sc.Count) * (sc.WeaponMin + sc.WeaponMax) / 2
		}
	}
}

// IsEmpty returns true if the fleet has no ships.
func (fs *FleetSnapshot) IsEmpty() bool {
	return fs.TotalShipCount() == 0
}

// GetDestroyedShipCosts calculates the refund (50% of cost) for destroyed ships.
func GetDestroyedShipCosts(destroyed map[string]int) map[string]float64 {
	refund := make(map[string]float64)

	for typeID, count := range destroyed {
		st := ship.GetShipType(ship.TypeID(typeID))
		if st == nil {
			continue
		}

		refund["food"] += st.Cost.Food * float64(count) * 0.5
		refund["composite"] += st.Cost.Composite * float64(count) * 0.5
		refund["mechanisms"] += st.Cost.Mechanisms * float64(count) * 0.5
		refund["reagents"] += st.Cost.Reagents * float64(count) * 0.5
		refund["money"] += st.Cost.Money * float64(count) * 0.5
	}

	return refund
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
