package building

import "math"

// Shipyard enables ship construction.
type Shipyard struct {
	Building
	ShipQueue     []int
	ShipBuildTime float64
}

// NewShipyard creates a new Shipyard building.
func NewShipyard(planetID string) *Shipyard {
	return &Shipyard{
		Building: Building{
			BuildingType:  TypeShipyard,
			BuildingLevel: 0,
			BuildProgress: 0,
			PlanetID:      planetID,
		},
		ShipQueue:     []int{},
		ShipBuildTime: 0,
	}
}

// Consumption returns energy consumed per level (16 per level).
func (s *Shipyard) Consumption() int {
	return 16
}

// BuildTime returns the time to build next level.
func (s *Shipyard) BuildTime() float64 {
	return math.Pow(2, float64(s.Building.BuildingLevel+7)) + 3000
}

// Cost returns the food cost to build next level.
func (s *Shipyard) Cost() float64 {
	return math.Pow(2, float64(s.Building.BuildingLevel+7))
}

// Produce returns no production (shipyard only enables ship building).
func (s *Shipyard) Produce(level int) ProductionResult {
	return ProductionResult{}
}

// BuildShip advances ship construction. Returns ship type index if a ship is completed.
func (s *Shipyard) BuildShip() int {
	if len(s.ShipQueue) > 0 {
		s.ShipBuildTime--
		if s.ShipBuildTime <= 0 {
			shipType := s.ShipQueue[0]
			s.ShipQueue = s.ShipQueue[1:]
			if len(s.ShipQueue) > 0 {
				s.ShipBuildTime = 100
			}
			return shipType
		}
	}
	return -1
}

// QueueShip adds a ship type to the construction queue.
func (s *Shipyard) QueueShip(shipType int) {
	s.ShipQueue = append(s.ShipQueue, shipType)
	if len(s.ShipQueue) == 1 {
		s.ShipBuildTime = 100
	}
}

// MaxShips returns the max ships capacity based on base and shipyard levels.
func (s *Shipyard) MaxShips(baseLevel int) int {
	return 10 * (baseLevel + s.BuildingLevel)
}
