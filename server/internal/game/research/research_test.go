package research

import (
	"testing"
)

func TestPrerequisitesMustBeMet(t *testing.T) {
	techs := TechsByTree(TreeStandard)
	tree := BuildTree(techs)

	// energy_storage depends on planet_exploration
	energyStorage := GetTechByID("energy_storage")
	if energyStorage == nil {
		t.Fatal("energy_storage tech not found")
	}

	// Without prerequisites, should not be unlocked
	completed := map[string]int{}
	if tree.IsUnlocked(energyStorage, completed) {
		t.Error("energy_storage should not be unlocked without planet_exploration")
	}

	// With prerequisites met, should be unlocked
	completed["planet_exploration"] = 1
	if !tree.IsUnlocked(energyStorage, completed) {
		t.Error("energy_storage should be unlocked after planet_exploration")
	}
}

func TestResearchProgressAdvances(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	tech := GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("planet_exploration tech not found")
	}

	// Start research with sufficient resources
	err := rs.StartResearch(tech, 100, 100, 0)
	if err != nil {
		t.Fatalf("failed to start research: %v", err)
	}

	state := rs.GetResearchState("planet_exploration")
	if state == nil {
		t.Fatal("research state not found")
	}
	if !state.InProgress {
		t.Error("research should be in progress")
	}
	if state.Progress != tech.BuildTime {
		t.Errorf("expected progress %f, got %f", tech.BuildTime, state.Progress)
	}

	// Tick once
	rs.Tick()

	state = rs.GetResearchState("planet_exploration")
	if state == nil {
		t.Fatal("research state not found after tick")
	}
	if state.Progress != tech.BuildTime-1 {
		t.Errorf("expected progress %f after tick, got %f", tech.BuildTime-1, state.Progress)
	}

	// Tick until complete
	for state.Progress > 0 {
		rs.Tick()
		state = rs.GetResearchState("planet_exploration")
	}

	if !state.Completed {
		t.Error("research should be completed after progress reaches 0")
	}
	if state.InProgress {
		t.Error("research should no longer be in progress when completed")
	}
}

func TestCompletedResearchAppliesEffects(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	tech := GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("planet_exploration tech not found")
	}

	// Start and complete research
	err := rs.StartResearch(tech, 100, 100, 0)
	if err != nil {
		t.Fatalf("failed to start research: %v", err)
	}

	// Tick until complete
	state := rs.GetResearchState("planet_exploration")
	for state.Progress > 0 {
		rs.Tick()
		state = rs.GetResearchState("planet_exploration")
	}

	if !state.Completed {
		t.Fatal("research should be completed")
	}

	// Check that the effect was applied (level recorded)
	if rs.Completed["planet_exploration"] != 1 {
		t.Errorf("expected completed level 1, got %d", rs.Completed["planet_exploration"])
	}

	// Effect should have been logged (no panic = success)
	tech.Effect("test-planet", 1)
}

func TestCannotStartWithoutResources(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	tech := GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("planet_exploration tech not found")
	}

	err := rs.StartResearch(tech, 50, 50, 0)
	if err == nil {
		t.Error("expected error when starting research without sufficient resources")
	}

	if _, ok := err.(*ResearchError); !ok {
		t.Errorf("expected *ResearchError, got %T", err)
	}
}

func TestCannotStartWithoutPrerequisites(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	// energy_storage depends on planet_exploration
	tech := GetTechByID("energy_storage")
	if tech == nil {
		t.Fatal("energy_storage tech not found")
	}

	err := rs.StartResearch(tech, 1000, 1000, 0)
	if err == nil {
		t.Error("expected error when starting research without prerequisites")
	}

	if _, ok := err.(*ResearchError); !ok {
		t.Errorf("expected *ResearchError, got %T", err)
	}
}

func TestCannotStartAlreadyInProgress(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	tech := GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("planet_exploration tech not found")
	}

	err := rs.StartResearch(tech, 100, 100, 0)
	if err != nil {
		t.Fatalf("failed to start research: %v", err)
	}

	// Try to start again
	err = rs.StartResearch(tech, 100, 100, 0)
	if err == nil {
		t.Error("expected error when starting already in-progress research")
	}
}

func TestGetResearchProgress(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	tech := GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("planet_exploration tech not found")
	}

	// Progress should be 0 before starting
	pct := rs.GetResearchProgress("planet_exploration")
	if pct != 0 {
		t.Errorf("expected 0%% progress before starting, got %f%%", pct)
	}

	err := rs.StartResearch(tech, 100, 100, 0)
	if err != nil {
		t.Fatalf("failed to start research: %v", err)
	}

	// After starting, progress should be 0%
	pct = rs.GetResearchProgress("planet_exploration")
	if pct != 0 {
		t.Errorf("expected 0%% progress right after starting, got %f%%", pct)
	}

	// Tick halfway
	for i := 0; i < int(tech.BuildTime/2); i++ {
		rs.Tick()
	}

	pct = rs.GetResearchProgress("planet_exploration")
	expected := 50.0
	if pct < expected-1 || pct > expected+1 {
		t.Errorf("expected ~%f%% progress, got %f%%", expected, pct)
	}
}

func TestMultiLevelTech(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	// energy_saving has max level 4
	tech := GetTechByID("energy_saving")
	if tech == nil {
		t.Fatal("energy_saving tech not found")
	}

	// Must complete planet_exploration first
	planetTech := GetTechByID("planet_exploration")
	err := rs.StartResearch(planetTech, 1000, 1000, 0)
	if err != nil {
		t.Fatalf("failed to start planet_exploration: %v", err)
	}
	for {
		rs.Tick()
		if rs.GetResearchState("planet_exploration").Completed {
			break
		}
	}

	// Must complete energy_storage first
	energyTech := GetTechByID("energy_storage")
	err = rs.StartResearch(energyTech, 1000, 1000, 0)
	if err != nil {
		t.Fatalf("failed to start energy_storage: %v", err)
	}
	for {
		rs.Tick()
		if rs.GetResearchState("energy_storage").Completed {
			break
		}
	}

	// Now energy_saving should be available
	err = rs.StartResearch(tech, 1000, 1000, 0)
	if err != nil {
		t.Fatalf("failed to start energy_saving: %v", err)
	}

	// Complete level 1
	for {
		rs.Tick()
		if rs.GetResearchState("energy_saving").Completed {
			break
		}
	}

	// Should be able to start level 2
	err = rs.StartResearch(tech, 1000, 1000, 0)
	if err != nil {
		t.Fatalf("failed to start energy_saving level 2: %v", err)
	}
}

func TestTreeTraversal(t *testing.T) {
	techs := TechsByTree(TreeStandard)
	tree := BuildTree(techs)

	var visited []string
	tree.TraverseDepthFirst(func(tech *Tech, depth int) {
		visited = append(visited, tech.ID)
		_ = depth
	})

	// planet_exploration should be visited first (it's a root node)
	if len(visited) == 0 || visited[0] != "planet_exploration" {
		t.Errorf("expected planet_exploration first, got %v", visited)
	}
}

func TestGetAncestors(t *testing.T) {
	tree := BuildTree(TechsByTree(TreeStandard))

	// energy_storage depends on planet_exploration
	ancestors := tree.GetAncestors("energy_storage")
	found := false
	for _, a := range ancestors {
		if a == "planet_exploration" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected planet_exploration in ancestors of energy_storage, got %v", ancestors)
	}
}

func TestAlienTechRequiresCommandCenter(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	alienTech := GetTechByID("alien_technologies")
	if alienTech == nil {
		t.Fatal("alien_technologies tech not found")
	}

	// Should fail without command_center
	err := rs.StartResearch(alienTech, 0, 0, 10)
	if err == nil {
		t.Error("expected error when starting alien tech without command_center")
	}
}

func TestStandardAndAlienTreesIndependent(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	// Standard tree techs should not be affected by alien tree state
	standardTech := GetTechByID("planet_exploration")
	if standardTech == nil {
		t.Fatal("planet_exploration not found")
	}

	// Should be able to start standard research without any alien tech
	err := rs.StartResearch(standardTech, 100, 100, 0)
	if err != nil {
		t.Fatalf("failed to start standard research: %v", err)
	}

	// Alien tree should have no available techs (no command_center)
	alien := rs.GetAvailableAlienTechs()
	if len(alien) != 0 {
		t.Errorf("expected 0 available alien techs, got %d", len(alien))
	}
}
