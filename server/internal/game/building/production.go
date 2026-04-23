package building

// ProdInfo is an alias for ProductionResult for backwards compatibility.
type ProdInfo = ProductionResult

// ProductionResult holds the resource changes from one tick.
type ProductionResult struct {
	Food      float64 `json:"food"`
	Composite float64 `json:"composite"`
	Mechanisms float64 `json:"mechanisms"`
	Reagents  float64 `json:"reagents"`
	Energy    float64 `json:"energy"`
	Money     float64 `json:"money"`
	AlienTech float64 `json:"alien_tech"`
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
