package game

// EnergyBuffer manages the planet's energy buffer.
// Energy is produced by solar panels and consumed by other buildings.
// The buffer accumulates excess energy and can go negative (deficit).
type EnergyBuffer struct {
	Value   float64 // current charge (can be negative = deficit)
	Max     float64 // max capacity = 100 + energy_storage_level * 100
	Deficit bool    // true when Value <= 0
}

// NewEnergyBuffer creates an energy buffer with default max capacity.
func NewEnergyBuffer() EnergyBuffer {
	return EnergyBuffer{
		Value: 100,
		Max:   100,
	}
}

// UpdateMax recalculates max capacity based on energy storage building level.
func (e *EnergyBuffer) UpdateMax(energyStorageLevel int) {
	e.Max = 100 + float64(energyStorageLevel)*100
	if e.Value > e.Max {
		e.Value = e.Max
	}
}
