package ship

// ConstructionQueueEntry represents a ship being built.
type ConstructionQueueEntry struct {
	TypeID       TypeID  `json:"type_id"`
	BuildTime    float64 `json:"build_time"`
	Progress     float64 `json:"progress"`
}

// Shipyard manages ship construction on a planet.
type Shipyard struct {
	Level      int                         `json:"level"`
	Queue      []ConstructionQueueEntry    `json:"queue"`
	BuildSpeed float64                     `json:"build_speed"`
}

// NewShipyard creates a new shipyard with level 0.
func NewShipyard() *Shipyard {
	return &Shipyard{
		Level:      0,
		Queue:      []ConstructionQueueEntry{},
		BuildSpeed: 1.0,
	}
}

// MaxSlots returns the maximum ship slots based on base and shipyard levels.
func (s *Shipyard) MaxSlots(baseLevel int) int {
	return 10 * (baseLevel + s.Level)
}

// CanQueueShip checks if a ship can be queued (resources and shipyard level).
func (s *Shipyard) CanQueueShip(st *ShipType, food, composite, mechanisms, reagents, money float64) bool {
	if st.MinShipyard > s.Level {
		return false
	}
	return st.Cost.CanAfford(food, composite, mechanisms, reagents, money)
}

// QueueShip adds a ship to the construction queue. Returns error if can't queue.
// The caller (planet) should verify resources and shipyard level before calling.
func (s *Shipyard) QueueShip(st *ShipType) error {
	if st.MinShipyard > s.Level {
		return &ShipyardError{reason: "shipyard_level_insufficient"}
	}

	s.Queue = append(s.Queue, ConstructionQueueEntry{
		TypeID:    st.TypeID,
		BuildTime: st.BuildTime,
		Progress:  0,
	})

	return nil
}

// DeductCost deducts the cost of a ship from the given resources.
// Call this after QueueShip succeeds.
func (s *Shipyard) DeductCost(st *ShipType, food, composite, mechanisms, reagents, money *float64) {
	*food -= st.Cost.Food
	*composite -= st.Cost.Composite
	*mechanisms -= st.Cost.Mechanisms
	*reagents -= st.Cost.Reagents
	*money -= st.Cost.Money
}

// Tick advances construction progress for all queued ships.
func (s *Shipyard) Tick() *TypeID {
	if len(s.Queue) == 0 {
		return nil
	}

	entry := &s.Queue[0]
	entry.Progress += s.BuildSpeed

	if entry.Progress >= entry.BuildTime {
		shipTypeID := entry.TypeID
		s.Queue = s.Queue[1:]
		return &shipTypeID
	}

	return nil
}

// GetQueuedCount returns the number of ships in the queue.
func (s *Shipyard) GetQueuedCount() int {
	return len(s.Queue)
}

// GetQueueProgress returns the progress percentage of the current build.
func (s *Shipyard) GetQueueProgress() float64 {
	if len(s.Queue) == 0 {
		return 0
	}
	entry := s.Queue[0]
	if entry.BuildTime == 0 {
		return 100
	}
	return (entry.Progress / entry.BuildTime) * 100
}

// CancelShip removes the first ship from the queue and returns its cost for refund.
func (s *Shipyard) CancelShip() (*ShipType, Cost, bool) {
	if len(s.Queue) == 0 {
		return nil, Cost{}, false
	}

	entry := s.Queue[0]
	s.Queue = s.Queue[1:]

	st := GetShipType(entry.TypeID)
	if st == nil {
		return nil, Cost{}, false
	}
	return st, st.Cost, true
}

// ShipyardError represents an error in shipyard operations.
type ShipyardError struct {
	reason string
}

func (e *ShipyardError) Error() string {
	return "shipyard error: " + e.reason
}
