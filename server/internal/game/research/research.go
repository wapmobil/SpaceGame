package research

import (
	"encoding/json"
	"log"

	"spacegame/internal/db"
)

// ResearchSystem manages research state for a planet.
type ResearchSystem struct {
	PlanetID  string
	States    map[string]*ResearchState // techID -> ResearchState
	Completed map[string]int            // techID -> highest level reached
	db        *db.Database
}

// NewResearchSystem creates a new ResearchSystem for a planet.
func NewResearchSystem(planetID string, d *db.Database) *ResearchSystem {
	return &ResearchSystem{
		PlanetID:  planetID,
		States:    make(map[string]*ResearchState),
		Completed: make(map[string]int),
		db:        d,
	}
}

// LoadFromDB loads research state from the database.
func (rs *ResearchSystem) LoadFromDB() error {
	if rs.db == nil {
		return nil
	}

	rows, err := rs.db.Query(
		`SELECT tech_id, completed, in_progress, progress, total_time, start_time, level
		 FROM research WHERE planet_id = $1`,
		rs.PlanetID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var state ResearchState
		var completed bool
		var inProgress bool
		var level int

		if err := rows.Scan(&state.TechID, &completed, &inProgress, &state.Progress, &state.TotalTime, &state.StartTime, &level); err != nil {
			log.Printf("Error scanning research row: %v", err)
			continue
		}

		state.Completed = completed
		state.InProgress = inProgress
		rs.States[state.TechID] = &state

		if completed && level > 0 {
			rs.Completed[state.TechID] = level
		}
	}

	return nil
}

// SaveToDB saves all research states to the database.
func (rs *ResearchSystem) SaveToDB() error {
	if rs.db == nil {
		return nil
	}

	for techID, state := range rs.States {
		_, err := rs.db.Exec(`
			INSERT INTO research (planet_id, tech_id, completed, in_progress, progress, total_time, start_time, level)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (planet_id, tech_id)
			DO UPDATE SET
				completed = EXCLUDED.completed,
				in_progress = EXCLUDED.in_progress,
				progress = EXCLUDED.progress,
				total_time = EXCLUDED.total_time,
				start_time = EXCLUDED.start_time,
				level = EXCLUDED.level,
				updated_at = NOW()
		`, rs.PlanetID, techID, state.Completed, state.InProgress, state.Progress, state.TotalTime, state.StartTime, rs.Completed[techID])

		if err != nil {
			log.Printf("Error saving research %s: %v", techID, err)
		}
	}

	return nil
}

// StartResearch begins researching a technology. Returns error if prerequisites not met or resources insufficient.
func (rs *ResearchSystem) StartResearch(tech *Tech, food, money, alienTech float64) error {
	// Check max level
	if rs.Completed[tech.ID] >= tech.MaxLevel {
		return &ResearchError{techID: tech.ID, reason: "max_level"}
	}

	// Check if already in progress
	if state, ok := rs.States[tech.ID]; ok {
		if state.InProgress {
			return &ResearchError{techID: tech.ID, reason: "already_in_progress"}
		}
	}

	// Check prerequisites
	tree := BuildTree(TechsByTree(tech.Tree))
	if !tree.HasPrerequisites(tech, rs.Completed) {
		return &ResearchError{techID: tech.ID, reason: "prerequisites_not_met"}
	}

	// Check resources
	if !tech.CanStart(food, money, alienTech) {
		return &ResearchError{techID: tech.ID, reason: "insufficient_resources"}
	}

	// Deduct resources and start research
	tech.DeductCost(&food, &money, &alienTech)

	rs.States[tech.ID] = &ResearchState{
		TechID:     tech.ID,
		Level:      rs.Completed[tech.ID] + 1,
		Completed:  false,
		InProgress: true,
		Progress:   tech.BuildTime,
		TotalTime:  tech.BuildTime,
	}

	return nil
}

// Tick advances all in-progress research by 1 second.
func (rs *ResearchSystem) Tick() {
	for techID, state := range rs.States {
		if !state.InProgress || state.Completed {
			continue
		}

		state.Progress--
		if state.Progress <= 0 {
			state.Progress = 0
			state.InProgress = false
			state.Completed = true
			rs.Completed[techID] = state.Level
		}
	}
}

// GetResearchState returns the state of a specific tech research.
func (rs *ResearchSystem) GetResearchState(techID string) *ResearchState {
	if state, ok := rs.States[techID]; ok {
		return state
	}
	return nil
}

// GetResearchProgress returns progress as a percentage (0-100).
func (rs *ResearchSystem) GetResearchProgress(techID string) float64 {
	state := rs.GetResearchState(techID)
	if state == nil || state.TotalTime == 0 {
		return 0
	}
	return ((state.TotalTime - state.Progress) / state.TotalTime) * 100
}

// GetAllStates returns all research states.
func (rs *ResearchSystem) GetAllStates() map[string]*ResearchState {
	result := make(map[string]*ResearchState)
	for k, v := range rs.States {
		result[k] = v
	}
	return result
}

// GetCompleted returns the completed tech map.
func (rs *ResearchSystem) GetCompleted() map[string]int {
	return rs.Completed
}

// GetAvailableTechs returns techs that can be researched.
func (rs *ResearchSystem) GetAvailableTechs() []*Tech {
	tree := BuildTree(TechsByTree(TreeStandard))
	available := tree.GetAvailable(TechsByTree(TreeStandard), rs.Completed, rs.inProgressIDs())
	return available
}

// GetAvailableAlienTechs returns alien techs that can be researched.
func (rs *ResearchSystem) GetAvailableAlienTechs() []*Tech {
	tree := BuildTree(TechsByTree(TreeAlien))
	available := tree.GetAvailable(TechsByTree(TreeAlien), rs.Completed, rs.inProgressIDs())
	return available
}

func (rs *ResearchSystem) inProgressIDs() map[string]bool {
	inProgress := make(map[string]bool)
	for _, state := range rs.States {
		if state.InProgress {
			inProgress[state.TechID] = true
		}
	}
	return inProgress
}

// GetResearchJSON returns research state as JSON for API responses.
func (rs *ResearchSystem) GetResearchJSON() ([]byte, error) {
	type ResearchEntry struct {
		TechID      string  `json:"tech_id"`
		Name        string  `json:"name"`
		Level       int     `json:"level"`
		Completed   bool    `json:"completed"`
		InProgress  bool    `json:"in_progress"`
		Progress    float64 `json:"progress"`
		TotalTime   float64 `json:"total_time"`
		ProgressPct float64 `json:"progress_pct"`
	}

	var entries []ResearchEntry
	tree := BuildTree(TechsByTree(TreeStandard))
	tree.TraverseDepthFirst(func(tech *Tech, depth int) {
		state := rs.GetResearchState(tech.ID)
		entry := ResearchEntry{
			TechID:    tech.ID,
			Name:      tech.Name,
			Level:     rs.Completed[tech.ID],
			Completed: false,
			InProgress: false,
		}

		if state != nil {
			entry.Completed = state.Completed
			entry.InProgress = state.InProgress
			entry.Progress = state.TotalTime - state.Progress
			entry.TotalTime = state.TotalTime
			entry.ProgressPct = rs.GetResearchProgress(tech.ID)
		} else {
			entry.TotalTime = tech.BuildTime
		}

		entries = append(entries, entry)
	})

	// Add alien tree
	tree2 := BuildTree(TechsByTree(TreeAlien))
	tree2.TraverseDepthFirst(func(tech *Tech, depth int) {
		state := rs.GetResearchState(tech.ID)
		entry := ResearchEntry{
			TechID:    tech.ID,
			Name:      tech.Name,
			Level:     rs.Completed[tech.ID],
			Completed: false,
			InProgress: false,
		}

		if state != nil {
			entry.Completed = state.Completed
			entry.InProgress = state.InProgress
			entry.Progress = state.TotalTime - state.Progress
			entry.TotalTime = state.TotalTime
			entry.ProgressPct = rs.GetResearchProgress(tech.ID)
		} else {
			entry.TotalTime = tech.BuildTime
		}

		entries = append(entries, entry)
	})

	return json.Marshal(entries)
}

// GetAvailableForAPI returns available techs for API responses.
func (rs *ResearchSystem) GetAvailableForAPI() ([]byte, error) {
	type AvailableEntry struct {
		TechID    string  `json:"tech_id"`
		Name      string  `json:"name"`
		CostFood  float64 `json:"cost_food,omitempty"`
		CostMoney float64 `json:"cost_money,omitempty"`
		CostAlien float64 `json:"cost_alien,omitempty"`
		BuildTime float64 `json:"build_time"`
		Tree      int     `json:"tree"`
	}

	standard := rs.GetAvailableTechs()
	var entries []AvailableEntry
	for _, tech := range standard {
		entries = append(entries, AvailableEntry{
			TechID:    tech.ID,
			Name:      tech.Name,
			CostFood:  tech.CostFood,
			CostMoney: tech.CostMoney,
			BuildTime: tech.BuildTime,
			Tree:      int(tech.Tree),
		})
	}

	// Check if alien tree is unlocked (command_center completed)
	if rs.Completed["command_center"] > 0 {
		alien := rs.GetAvailableAlienTechs()
		for _, tech := range alien {
			entries = append(entries, AvailableEntry{
				TechID:    tech.ID,
				Name:      tech.Name,
				CostAlien: tech.CostAlien,
				BuildTime: tech.BuildTime,
				Tree:      int(tech.Tree),
			})
		}
	}

	return json.Marshal(entries)
}

// ResearchError represents an error that occurred during research.
type ResearchError struct {
	techID string
	reason string
}

func (e *ResearchError) Error() string {
	return "research error: " + e.reason + " (tech: " + e.techID + ")"
}

// Effect functions that modify planet behavior when research completes.

func applyPlanetExploration(planetID string, level int) {
	log.Printf("[Research] Planet %s unlocked Factory building (level %d)", planetID, level)
}

func applyEnergyStorage(planetID string, level int) {
	log.Printf("[Research] Planet %s unlocked Energy Storage building (level %d)", planetID, level)
}

func applyEnergySaving(planetID string, level int) {
	log.Printf("[Research] Planet %s Energy Saving level %d: -10%% energy consumption", planetID, level)
}

func applyTrade(planetID string, level int) {
	log.Printf("[Research] Planet %s unlocked Marketplace (level %d)", planetID, level)
}

func applyShips(planetID string, level int) {
	log.Printf("[Research] Planet %s unlocked Shipyard (level %d)", planetID, level)
}

func applyUpgradedEnergyStorage(planetID string, level int) {
	log.Printf("[Research] Planet %s Upgraded Energy Storage level %d: +20%% capacity", planetID, level)
}

func applyFastConstruction(planetID string, level int) {
	log.Printf("[Research] Planet %s Fast Construction level %d: building speed bonus", planetID, level)
}

func applyParallelConstruction(planetID string, level int) {
	log.Printf("[Research] Planet %s Parallel Construction level %d: +%d simultaneous construction", planetID, level, level)
}

func applyCompactStorage(planetID string, level int) {
	log.Printf("[Research] Planet %s Compact Storage level %d: 2x storage capacity", planetID, level)
}

func applyExpeditions(planetID string, level int) {
	log.Printf("[Research] Planet %s unlocked Expedition system (level %d)", planetID, level)
}

func applyCommandCenter(planetID string, level int) {
	log.Printf("[Research] Planet %s unlocked Command Center - Alien Technology tree available (level %d)", planetID, level)
}

func applyAlienTechnologies(planetID string, level int) {
	log.Printf("[Research] Planet %s unlocked Alien Technologies tree (level %d)", planetID, level)
}

func applyAdditionalExpedition(planetID string, level int) {
	log.Printf("[Research] Planet %s Additional Expedition level %d: +1 concurrent expedition", planetID, level)
}

func applySuperEnergyStorage(planetID string, level int) {
	log.Printf("[Research] Planet %s Super Energy Storage level %d: +20%% capacity", planetID, level)
}
