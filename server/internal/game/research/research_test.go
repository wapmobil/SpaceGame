package research

import (
	"testing"
)

func TestResearchProgressAdvances(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	tech := GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("planet_exploration tech not found")
	}

	// Start research with sufficient resources
	food := 100.0
	money := 100.0
	alienTech := 0.0
	err := rs.StartResearch(tech, &food, &money, &alienTech)
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
	food := 100.0
	money := 100.0
	alienTech := 0.0
	err := rs.StartResearch(tech, &food, &money, &alienTech)
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
}

func TestCannotStartWithoutResources(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	tech := GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("planet_exploration tech not found")
	}

	food := 50.0
	money := 50.0
	alienTech := 0.0
	err := rs.StartResearch(tech, &food, &money, &alienTech)
	if err == nil {
		t.Error("expected error when starting research without sufficient resources")
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

	food := 100.0
	money := 100.0
	alienTech := 0.0
	err := rs.StartResearch(tech, &food, &money, &alienTech)
	if err != nil {
		t.Fatalf("failed to start research: %v", err)
	}

	// Try to start again
	err = rs.StartResearch(tech, &food, &money, &alienTech)
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

	food := 100.0
	money := 100.0
	alienTech := 0.0
	err := rs.StartResearch(tech, &food, &money, &alienTech)
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

	tech := GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("planet_exploration tech not found")
	}

	food := 100.0
	money := 100.0
	alienTech := 0.0

	// Complete level 1
	err := rs.StartResearch(tech, &food, &money, &alienTech)
	if err != nil {
		t.Fatalf("failed to start research: %v", err)
	}
	for {
		rs.Tick()
		if rs.GetResearchState("planet_exploration").Completed {
			break
		}
	}

	// planet_exploration has max_level 1, should not be able to start again
	err = rs.StartResearch(tech, &food, &money, &alienTech)
	if err == nil {
		t.Error("expected error when starting max-level research")
	}
}

func TestResourceDeduction(t *testing.T) {
	rs := NewResearchSystem("test-planet", nil)

	tech := GetTechByID("planet_exploration")
	if tech == nil {
		t.Fatal("planet_exploration tech not found")
	}

	food := 100.0
	money := 100.0
	alienTech := 0.0

	// Start research
	err := rs.StartResearch(tech, &food, &money, &alienTech)
	if err != nil {
		t.Fatalf("failed to start research: %v", err)
	}

	// Resources should be deducted
	if food != 0.0 {
		t.Errorf("expected food to be 0, got %f", food)
	}
	if money != 0.0 {
		t.Errorf("expected money to be 0, got %f", money)
	}
}

func TestLocationBuildingsTechExists(t *testing.T) {
	tech := GetTechByID("location_buildings")
	if tech == nil {
		t.Fatal("expected location_buildings tech to exist")
	}
	if tech.Name != "Location Buildings" {
		t.Errorf("expected name 'Location Buildings', got '%s'", tech.Name)
	}
	if tech.MaxLevel != 1 {
		t.Errorf("expected max level 1, got %d", tech.MaxLevel)
	}
}

func TestAdvancedExplorationTechExists(t *testing.T) {
	tech := GetTechByID("advanced_exploration")
	if tech == nil {
		t.Fatal("expected advanced_exploration tech to exist")
	}
	if tech.Name != "Advanced Exploration" {
		t.Errorf("expected name 'Advanced Exploration', got '%s'", tech.Name)
	}
	if tech.MaxLevel != 3 {
		t.Errorf("expected max level 3, got %d", tech.MaxLevel)
	}
}

func TestAdvancedExplorationDependsOnLocationBuildings(t *testing.T) {
	tech := GetTechByID("advanced_exploration")
	if tech == nil {
		t.Fatal("expected advanced_exploration tech to exist")
	}
	found := false
	for _, dep := range tech.DependsOn {
		if dep == "location_buildings" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected advanced_exploration to depend on location_buildings, got %v", tech.DependsOn)
	}
}
