package building

// ProductionResult holds the resource changes from one tick.
type ProductionResult struct {
	Food      float64
	Composite float64
	Mechanisms float64
	Reagents  float64
	Energy    float64
	Money     float64
	AlienTech float64
}

// Add adds another ProductionResult to this one.
func (p *ProductionResult) Add(o ProductionResult) {
	p.Food += o.Food
	p.Composite += o.Composite
	p.Mechanisms += o.Mechanisms
	p.Reagents += o.Reagents
	p.Energy += o.Energy
	p.Money += o.Money
	p.AlienTech += o.AlienTech
}

// IsZero returns true if all deltas are zero.
func (p *ProductionResult) IsZero() bool {
	return p.Food == 0 && p.Composite == 0 && p.Mechanisms == 0 &&
		p.Reagents == 0 && p.Energy == 0 && p.Money == 0 && p.AlienTech == 0
}
