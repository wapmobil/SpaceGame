package ship

// FleetShip represents a single ship in the fleet.
type FleetShip struct {
	TypeID TypeID `json:"type_id"`
	Count  int    `json:"count"`
	HP     float64 `json:"hp"`
}

// Fleet represents the collection of ships on a planet.
type Fleet struct {
	Ships map[string]*FleetShip // key = TypeID string
}

// NewFleet creates an empty fleet.
func NewFleet() *Fleet {
	return &Fleet{
		Ships: make(map[string]*FleetShip),
	}
}

// AddShip adds a ship to the fleet.
func (f *Fleet) AddShip(st *ShipType, count int) {
	key := string(st.TypeID)
	if existing, ok := f.Ships[key]; ok {
		existing.Count += count
		existing.HP = st.HP * float64(existing.Count)
	} else {
		f.Ships[key] = &FleetShip{
			TypeID: st.TypeID,
			Count:  count,
			HP:     st.HP * float64(count),
		}
	}
}

// RemoveShip removes ships from the fleet. Returns remaining count.
func (f *Fleet) RemoveShip(typeID TypeID, count int) int {
	key := string(typeID)
	if existing, ok := f.Ships[key]; ok {
		existing.Count -= count
		if existing.Count <= 0 {
			delete(f.Ships, key)
			return 0
		}
		st := GetShipType(typeID)
		if st != nil {
			existing.HP = st.HP * float64(existing.Count)
		}
		return existing.Count
	}
	return 0
}

// GetShipCount returns the number of ships of a given type.
func (f *Fleet) GetShipCount(typeID TypeID) int {
	key := string(typeID)
	if s, ok := f.Ships[key]; ok {
		return s.Count
	}
	return 0
}

// TotalShipCount returns the total number of ships.
func (f *Fleet) TotalShipCount() int {
	total := 0
	for _, s := range f.Ships {
		total += s.Count
	}
	return total
}

// TotalSlots returns total slots used by all ships.
func (f *Fleet) TotalSlots() int {
	total := 0
	for _, s := range f.Ships {
		st := GetShipType(TypeID(s.TypeID))
		if st != nil {
			total += st.Slots * s.Count
		}
	}
	return total
}

// TotalCargoCapacity returns the combined cargo capacity.
func (f *Fleet) TotalCargoCapacity() float64 {
	total := 0.0
	for _, s := range f.Ships {
		st := GetShipType(TypeID(s.TypeID))
		if st != nil {
			total += st.Cargo * float64(s.Count)
		}
	}
	return total
}

// TotalEnergyConsumption returns the combined energy consumption.
func (f *Fleet) TotalEnergyConsumption() float64 {
	total := 0.0
	for _, s := range f.Ships {
		st := GetShipType(TypeID(s.TypeID))
		if st != nil {
			total += st.Energy * float64(s.Count)
		}
	}
	return total
}

// TotalDamage returns the combined damage of all ships.
func (f *Fleet) TotalDamage() float64 {
	total := 0.0
	for _, s := range f.Ships {
		st := GetShipType(TypeID(s.TypeID))
		if st != nil {
			total += st.WeaponMaxDmg * float64(s.Count)
		}
	}
	return total
}

// TotalHP returns the combined HP of all ships.
func (f *Fleet) TotalHP() float64 {
	total := 0.0
	for _, s := range f.Ships {
		total += s.HP
	}
	return total
}

// CanAddShip checks if a ship can be added without exceeding max slots.
func (f *Fleet) CanAddShip(st *ShipType, count int, maxSlots int) bool {
	return f.TotalSlots()+st.Slots*count <= maxSlots
}

// GetShipState returns the fleet state as a map for API responses.
func (f *Fleet) GetShipState() map[string]interface{} {
	result := make(map[string]interface{})
	for _, s := range f.Ships {
		st := GetShipType(TypeID(s.TypeID))
		if st != nil {
			result[string(s.TypeID)] = map[string]interface{}{
				"type_id":        s.TypeID,
				"name":           st.Name,
				"description":    st.Description,
				"count":          s.Count,
				"hp":             s.HP,
				"max_hp":         st.HP * float64(s.Count),
				"slots_used":     st.Slots * s.Count,
				"cargo_capacity": st.Cargo * float64(s.Count),
				"energy":         st.Energy * float64(s.Count),
				"weapon_min":     st.WeaponMinDmg * float64(s.Count),
				"weapon_max":     st.WeaponMaxDmg * float64(s.Count),
				"armor":          st.Armor * float64(s.Count),
			}
		}
	}
	return result
}
