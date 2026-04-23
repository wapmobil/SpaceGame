package research

import (
	"context"
	"encoding/json"
	"log"

	"spacegame/internal/db"
)

// ResearchSystem manages research state for a planet.
type ResearchSystem struct {
	PlanetID      string
	States        map[string]*ResearchState // techID -> ResearchState
	Completed     map[string]int            // techID -> highest level reached
	lastCompleted map[string]bool           // techs completed in the last tick
	db            *db.Database
}

// NewResearchSystem creates a new ResearchSystem for a planet.
func NewResearchSystem(planetID string, d *db.Database) *ResearchSystem {
	return &ResearchSystem{
		PlanetID:      planetID,
		States:        make(map[string]*ResearchState),
		Completed:     make(map[string]int),
		lastCompleted: make(map[string]bool),
		db:            d,
	}
}

// LoadFromDB loads research state from the database.
func (rs *ResearchSystem) LoadFromDB() error {
	if rs.db == nil {
		return nil
	}

	hasTotalTime, _ := rs.db.ColumnExists(context.Background(), "research", "total_time")
	hasStartTime, _ := rs.db.ColumnExists(context.Background(), "research", "start_time")
	hasLevel, _ := rs.db.ColumnExists(context.Background(), "research", "level")

	var query string
	var scanArgs []interface{}
	if hasTotalTime && hasStartTime && hasLevel {
		query = `SELECT tech_id, completed, in_progress, progress, total_time, start_time, level FROM research WHERE planet_id = $1`
		scanArgs = []interface{}{rs.PlanetID}
	} else {
		query = `SELECT tech_id, completed, in_progress, progress FROM research WHERE planet_id = $1`
		scanArgs = []interface{}{rs.PlanetID}
	}

	rows, err := rs.db.Query(query, scanArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var state ResearchState
		var completed bool
		var inProgress bool

		if hasTotalTime && hasStartTime && hasLevel {
			var level int
			if err := rows.Scan(&state.TechID, &completed, &inProgress, &state.Progress, &state.TotalTime, &state.StartTime, &level); err != nil {
				log.Printf("Error scanning research row: %v", err)
				continue
			}
			if completed && level > 0 {
				rs.Completed[state.TechID] = level
			}
		} else {
			if err := rows.Scan(&state.TechID, &completed, &inProgress, &state.Progress); err != nil {
				log.Printf("Error scanning research row: %v", err)
				continue
			}
		}

		state.Completed = completed
		state.InProgress = inProgress
		rs.States[state.TechID] = &state
	}

	return nil
}

// SaveToDB saves all research states to the database.
func (rs *ResearchSystem) SaveToDB() error {
	if rs.db == nil {
		return nil
	}

	hasTotalTime, _ := rs.db.ColumnExists(context.Background(), "research", "total_time")
	hasStartTime, _ := rs.db.ColumnExists(context.Background(), "research", "start_time")
	hasLevel, _ := rs.db.ColumnExists(context.Background(), "research", "level")

	for techID, state := range rs.States {
		var err error
		if hasTotalTime && hasStartTime && hasLevel {
			_, err = rs.db.Exec(`
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
		} else {
			_, err = rs.db.Exec(`
				INSERT INTO research (planet_id, tech_id, completed, in_progress, progress)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (planet_id, tech_id)
				DO UPDATE SET
					completed = EXCLUDED.completed,
					in_progress = EXCLUDED.in_progress,
					progress = EXCLUDED.progress,
					updated_at = NOW()
			`, rs.PlanetID, techID, state.Completed, state.InProgress, state.Progress)
		}

		if err != nil {
			log.Printf("Error saving research %s: %v", techID, err)
		}
	}

	return nil
}

// StartResearch begins researching a technology. Returns error if prerequisites not met or resources insufficient.
func (rs *ResearchSystem) StartResearch(tech *Tech, food, money, alienTech *float64) error {
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
	if !tech.CanStart(*food, *money, *alienTech) {
		return &ResearchError{techID: tech.ID, reason: "insufficient_resources"}
	}

	// Deduct resources and start research
	tech.DeductCost(food, money, alienTech)

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
	rs.lastCompleted = make(map[string]bool)
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
			rs.lastCompleted[techID] = true
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

// GetLastCompleted returns techs that completed in the last tick.
func (rs *ResearchSystem) GetLastCompleted() map[string]bool {
	return rs.lastCompleted
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
		Description string  `json:"description"`
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
			Description: tech.Description,
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
			Description: tech.Description,
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


