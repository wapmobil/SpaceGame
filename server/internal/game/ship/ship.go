package ship

// TypeID represents a unique ship type identifier.
type TypeID string

const (
	TypeTradeShip   TypeID = "trade_ship"
	TypeSmallShip   TypeID = "small_ship"
	TypeInterceptor TypeID = "interceptor"
	TypeCorvette    TypeID = "corvette"
	TypeFrigate     TypeID = "frigate"
	TypeCruiser     TypeID = "cruiser"
)

// ShipType defines the blueprint for a ship class.
type ShipType struct {
	TypeID       TypeID    `json:"type_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Slots        int       `json:"slots"`
	Cargo        float64   `json:"cargo"`
	Energy       float64   `json:"energy"`
	HP           float64   `json:"hp"`
	Armor        float64   `json:"armor"`
	WeaponMinDmg float64   `json:"weapon_min_damage"`
	WeaponMaxDmg float64   `json:"weapon_max_damage"`
	Cost         Cost      `json:"cost"`
	BuildTime    float64   `json:"build_time"`
	MinShipyard  int       `json:"min_shipyard_level"`
}

// Cost holds the resource costs to build a ship.
type Cost struct {
	Food       float64 `json:"food"`
	Composite  float64 `json:"composite"`
	Mechanisms float64 `json:"mechanisms"`
	Reagents   float64 `json:"reagents"`
	Money      float64 `json:"money"`
}

// CanAfford checks if the given resource amounts cover this cost.
func (c Cost) CanAfford(food, composite, mechanisms, reagents, money float64) bool {
	return food >= c.Food &&
		composite >= c.Composite &&
		mechanisms >= c.Mechanisms &&
		reagents >= c.Reagents &&
		money >= c.Money
}

// Deduct returns a map of resource deltas.
func (c Cost) Deduct() map[string]float64 {
	return map[string]float64{
		"food":      -c.Food,
		"composite": -c.Composite,
		"mechanisms": -c.Mechanisms,
		"reagents":  -c.Reagents,
		"money":     -c.Money,
	}
}

// AllShipTypes returns all available ship types.
func AllShipTypes() []*ShipType {
	return []*ShipType{
		{
			TypeID:       TypeTradeShip,
			Name:         "Trade Ship",
			Description:  "Cargo vessel for trade expeditions",
			Slots:        2,
			Cargo:        100,
			Energy:       100,
			HP:           20,
			Armor:        0,
			WeaponMinDmg: 0,
			WeaponMaxDmg: 0,
			Cost: Cost{
				Food: 100, Composite: 100, Mechanisms: 100, Reagents: 100, Money: 100,
			},
			BuildTime:     100,
			MinShipyard:   1,
		},
		{
			TypeID:       TypeSmallShip,
			Name:         "Small Ship",
			Description:  "Basic combat vessel",
			Slots:        1,
			Cargo:        2,
			Energy:       10,
			HP:           4,
			Armor:        0,
			WeaponMinDmg: 1,
			WeaponMaxDmg: 1,
			Cost: Cost{
				Food: 10, Composite: 10, Mechanisms: 10, Reagents: 10, Money: 10,
			},
			BuildTime:     10,
			MinShipyard:   1,
		},
		{
			TypeID:       TypeInterceptor,
			Name:         "Interceptor",
			Description:  "Fast attack vessel",
			Slots:        2,
			Cargo:        0,
			Energy:       100,
			HP:           20,
			Armor:        0,
			WeaponMinDmg: 8,
			WeaponMaxDmg: 10,
			Cost: Cost{
				Food: 100, Composite: 100, Mechanisms: 100, Reagents: 100, Money: 100,
			},
			BuildTime:     100,
			MinShipyard:   2,
		},
		{
			TypeID:       TypeCorvette,
			Name:         "Corvette",
			Description:  "Balanced combat vessel",
			Slots:        4,
			Cargo:        10,
			Energy:       200,
			HP:           200,
			Armor:        10,
			WeaponMinDmg: 2,
			WeaponMaxDmg: 4,
			Cost: Cost{
				Food: 250, Composite: 250, Mechanisms: 250, Reagents: 250, Money: 250,
			},
			BuildTime:     250,
			MinShipyard:   3,
		},
		{
			TypeID:       TypeFrigate,
			Name:         "Frigate",
			Description:  "Heavy combat vessel",
			Slots:        5,
			Cargo:        20,
			Energy:       400,
			HP:           100,
			Armor:        5,
			WeaponMinDmg: 6,
			WeaponMaxDmg: 4,
			Cost: Cost{
				Food: 500, Composite: 500, Mechanisms: 500, Reagents: 500, Money: 500,
			},
			BuildTime:     500,
			MinShipyard:   3,
		},
		{
			TypeID:       TypeCruiser,
			Name:         "Cruiser",
			Description:  "Ultimate combat vessel",
			Slots:        6,
			Cargo:        0,
			Energy:       500,
			HP:           500,
			Armor:        20,
			WeaponMinDmg: 2,
			WeaponMaxDmg: 16,
			Cost: Cost{
				Food: 1000, Composite: 1000, Mechanisms: 1000, Reagents: 1000, Money: 1000,
			},
			BuildTime:     1000,
			MinShipyard:   4,
		},
	}
}

// GetShipType returns a ship type by its ID.
func GetShipType(id TypeID) *ShipType {
	for _, st := range AllShipTypes() {
		if st.TypeID == id {
			return st
		}
	}
	return nil
}

// IsPeaceful returns true if the ship has no weapons.
func (st *ShipType) IsPeaceful() bool {
	return st.WeaponMinDmg == 0 && st.WeaponMaxDmg == 0
}
