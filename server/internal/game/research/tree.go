package research

// ResearchState tracks the state of a single tech research on a planet.
type ResearchState struct {
	TechID      string  `json:"tech_id"`
	Level       int     `json:"level"`
	Completed   bool    `json:"completed"`
	InProgress  bool    `json:"in_progress"`
	Progress    float64 `json:"progress"`       // seconds remaining
	TotalTime   float64 `json:"total_time"`     // total build time
	StartTime   float64 `json:"start_time"`     // when research started (game tick)
}

// ResearchTree represents a connected tree of technologies with parent-child relationships.
type ResearchTree struct {
	Techs   []*Tech
	Children map[string][]string // techID -> list of child tech IDs
}

// BuildTree constructs a ResearchTree from a list of techs.
func BuildTree(techs []*Tech) *ResearchTree {
	tree := &ResearchTree{
		Techs:    techs,
		Children: make(map[string][]string),
	}

	// Build parent-child relationships from dependencies
	for _, tech := range techs {
		found := false
		for _, other := range techs {
			if other.ID == tech.ID {
				continue
			}
			for _, dep := range tech.DependsOn {
				if dep == other.ID {
					tree.Children[other.ID] = append(tree.Children[other.ID], tech.ID)
					found = true
				}
			}
		}
		if !found && len(tech.DependsOn) == 0 {
			// Root node
			tree.Children[tech.ID] = append(tree.Children[tech.ID], tech.ID)
		}
	}

	return tree
}

// IsUnlocked checks if a tech is available to research (all prerequisites completed).
func (rt *ResearchTree) IsUnlocked(tech *Tech, completed map[string]int) bool {
	for _, depID := range tech.DependsOn {
		if completed[depID] <= 0 {
			return false
		}
	}
	return true
}

// IsMaxLevel checks if a tech has reached its maximum level.
func (rt *ResearchTree) IsMaxLevel(tech *Tech, currentLevel int) bool {
	return currentLevel >= tech.MaxLevel
}

// GetAvailable returns techs that can be started (prerequisites met, not max level, not in progress).
func (rt *ResearchTree) GetAvailable(techs []*Tech, completed map[string]int, inProgress map[string]bool) []*Tech {
	var available []*Tech
	for _, tech := range techs {
		if rt.IsMaxLevel(tech, completed[tech.ID]) {
			continue
		}
		if inProgress[tech.ID] {
			continue
		}
		if rt.IsUnlocked(tech, completed) {
			available = append(available, tech)
		}
	}
	return available
}

// HasPrerequisites checks if all prerequisites for a tech are met.
func (rt *ResearchTree) HasPrerequisites(tech *Tech, completed map[string]int) bool {
	return rt.IsUnlocked(tech, completed)
}

// TraverseDepthFirst visits all nodes in depth-first order, calling the callback.
func (rt *ResearchTree) TraverseDepthFirst(callback func(*Tech, int)) {
	visited := make(map[string]bool)
	for _, tech := range rt.Techs {
		if len(tech.DependsOn) == 0 {
			rt.visit(tech, callback, visited, 0)
		}
	}
}

func (rt *ResearchTree) visit(tech *Tech, callback func(*Tech, int), visited map[string]bool, depth int) {
	if visited[tech.ID] {
		return
	}
	visited[tech.ID] = true
	callback(tech, depth)
	for _, childID := range rt.Children[tech.ID] {
		if child := GetTechByID(childID); child != nil {
			rt.visit(child, callback, visited, depth+1)
		}
	}
}

// GetAncestors returns all prerequisite tech IDs for a given tech (transitive).
func (rt *ResearchTree) GetAncestors(techID string) []string {
	var result []string
	visited := make(map[string]bool)
	rt.collectAncestors(techID, &result, visited)
	return result
}

func (rt *ResearchTree) collectAncestors(techID string, result *[]string, visited map[string]bool) {
	if visited[techID] {
		return
	}
	visited[techID] = true

	tech := GetTechByID(techID)
	if tech == nil {
		return
	}

	for _, depID := range tech.DependsOn {
		*result = append(*result, depID)
		rt.collectAncestors(depID, result, visited)
	}
}

// GetDescendants returns all tech IDs that depend on the given tech (transitive).
func (rt *ResearchTree) GetDescendants(techID string) []string {
	var result []string
	visited := make(map[string]bool)
	rt.collectDescendants(techID, &result, visited)
	return result
}

func (rt *ResearchTree) collectDescendants(techID string, result *[]string, visited map[string]bool) {
	if visited[techID] {
		return
	}
	visited[techID] = true

	for _, childID := range rt.Children[techID] {
		*result = append(*result, childID)
		rt.collectDescendants(childID, result, visited)
	}
}
