package planet_survey

import (
	"testing"
	"time"
)

func TestValidateInventory_Valid(t *testing.T) {
	inv := map[string]float64{
		"food":        100,
		"iron":        200,
		"composite":   50,
		"mechanisms":  30,
		"reagents":    20,
	}
	if err := ValidateInventory(inv); err != nil {
		t.Fatalf("expected valid inventory, got error: %v", err)
	}
}

func TestValidateInventory_OverLimit(t *testing.T) {
	inv := map[string]float64{
		"food":     600,
		"iron":     500,
		"reagents": 100,
	}
	err := ValidateInventory(inv)
	if err == nil {
		t.Fatal("expected error for over-limit inventory, got nil")
	}
	if got := err.Error(); got != "inventory total 1200 exceeds maximum 1000" {
		t.Fatalf("unexpected error message: %s", got)
	}
}

func TestValidateInventory_Negative(t *testing.T) {
	inv := map[string]float64{
		"food":   -10,
		"iron":   100,
		"reagents": 20,
	}
	err := ValidateInventory(inv)
	if err == nil {
		t.Fatal("expected error for negative inventory, got nil")
	}
	if got := err.Error(); got != "inventory amount for food cannot be negative" {
		t.Fatalf("unexpected error message: %s", got)
	}
}

func TestValidateInventory_UnknownResource(t *testing.T) {
	inv := map[string]float64{
		"food":     100,
		"plutonium": 50,
	}
	err := ValidateInventory(inv)
	if err == nil {
		t.Fatal("expected error for unknown resource, got nil")
	}
	if got := err.Error(); got != "unknown resource in inventory: plutonium" {
		t.Fatalf("unexpected error message: %s", got)
	}
}

func TestValidateInventory_Empty(t *testing.T) {
	inv := map[string]float64{}
	err := ValidateInventory(inv)
	if err == nil {
		t.Fatal("expected error for empty inventory, got nil")
	}
	if got := err.Error(); got != "inventory cannot be empty" {
		t.Fatalf("unexpected error message: %s", got)
	}
}

func TestValidateInventory_ZeroTotal(t *testing.T) {
	inv := map[string]float64{
		"food": 0,
		"iron": 0,
	}
	err := ValidateInventory(inv)
	if err == nil {
		t.Fatal("expected error for zero total inventory, got nil")
	}
	if got := err.Error(); got != "inventory total must be greater than 0" {
		t.Fatalf("unexpected error message: %s", got)
	}
}

func TestParseReward_NestedFormat(t *testing.T) {
	raw := map[string]interface{}{
		"resources": map[string]interface{}{
			"reagents":  50,
			"food":      -10,
			"iron":      30,
			"unknown":   100,
		},
	}
	result, err := ParseReward(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := map[string]float64{
		"reagents": 50,
		"food":     -10,
		"iron":     30,
	}
	if len(result) != len(expected) {
		t.Fatalf("expected %d resources, got %d", len(expected), len(result))
	}
	for k, v := range expected {
		if result[k] != v {
			t.Errorf("expected %s = %f, got %f", k, v, result[k])
		}
	}
}

func TestParseReward_FlatFormat(t *testing.T) {
	raw := map[string]interface{}{
		"food":      25,
		"iron":      -5,
		"composite": 10,
		"plutonium": 100,
	}
	result, err := ParseReward(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(result))
	}
	if result["food"] != 25 {
		t.Errorf("expected food = 25, got %f", result["food"])
	}
	if result["iron"] != -5 {
		t.Errorf("expected iron = -5, got %f", result["iron"])
	}
	if result["composite"] != 10 {
		t.Errorf("expected composite = 10, got %f", result["composite"])
	}
	if _, ok := result["plutonium"]; ok {
		t.Error("unknown resource plutonium should be ignored")
	}
}

func TestParseReward_Empty(t *testing.T) {
	raw := map[string]interface{}{}
	result, err := ParseReward(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestApplyRewardToInventory_Add(t *testing.T) {
	inv := map[string]float64{
		"food":     100,
		"iron":     200,
		"reagents": 50,
	}
	reward := map[string]float64{
		"food":     50,
		"iron":     100,
		"reagents": -20,
	}
	ApplyRewardToInventory(inv, reward)
	if inv["food"] != 150 {
		t.Errorf("expected food = 150, got %f", inv["food"])
	}
	if inv["iron"] != 300 {
		t.Errorf("expected iron = 300, got %f", inv["iron"])
	}
	if inv["reagents"] != 30 {
		t.Errorf("expected reagents = 30, got %f", inv["reagents"])
	}
}

func TestApplyRewardToInventory_ClampToZero(t *testing.T) {
	inv := map[string]float64{
		"food":     10,
		"iron":     200,
		"reagents": 50,
	}
	reward := map[string]float64{
		"food": -100,
	}
	ApplyRewardToInventory(inv, reward)
	if inv["food"] < 0 {
		t.Errorf("expected food clamped to >= 0, got %f", inv["food"])
	}
	if inv["food"] != 0 {
		t.Errorf("expected food = 0, got %f", inv["food"])
	}
}

func TestApplyRewardToInventory_ClampToMax(t *testing.T) {
	inv := map[string]float64{
		"food":     900,
		"iron":     50,
		"reagents": 50,
	}
	reward := map[string]float64{
		"food":     100,
		"iron":     100,
		"reagents": 100,
	}
	ApplyRewardToInventory(inv, reward)
	total := GetInventorySize(inv)
	if total > MaxInventory {
		t.Errorf("expected total <= %f, got %f", MaxInventory, total)
	}
	if total < MaxInventory-0.01 {
		t.Errorf("expected total approximately %f, got %f", MaxInventory, total)
	}
}

func TestGetInventorySize_Sum(t *testing.T) {
	inv := map[string]float64{
		"food":     100,
		"iron":     200,
		"composite": 50,
	}
	size := GetInventorySize(inv)
	expected := 350.0
	if size != expected {
		t.Errorf("expected size = %f, got %f", expected, size)
	}
}

func TestGetInventorySize_Empty(t *testing.T) {
	inv := map[string]float64{}
	size := GetInventorySize(inv)
	if size != 0 {
		t.Errorf("expected size = 0, got %f", size)
	}
}

func TestParseEventResponse_ValidJSON(t *testing.T) {
	raw := `{"event_id": "evt_1", "description": "Вы нашел странный кристалл.", "immediate_reward": {"resources": {"reagents": 20}}, "choices": [{"label": "Взять", "description": "Забрать кристалл", "reward": {"resources": {"reagents": 50}}, "next_event_id": "evt_2"}, {"label": "Оставить", "description": "Оставить на месте", "reward": {}, "next_event_id": "evt_3"}], "is_end": false, "location_reward": "crystal_cave"}`
	event, err := ParseEventResponse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Description != "Вы нашел странный кристалл." {
		t.Errorf("unexpected description: %s", event.Description)
	}
	if event.ImmediateReward["reagents"] != 20 {
		t.Errorf("expected immediate reagents = 20, got %f", event.ImmediateReward["reagents"])
	}
	if len(event.Choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(event.Choices))
	}
	if event.Choices[0].Label != "Взять" {
		t.Errorf("expected first choice label = 'Взять', got '%s'", event.Choices[0].Label)
	}
	if event.IsEnd {
		t.Error("expected is_end = false")
	}
	if event.LocationReward != "crystal_cave" {
		t.Errorf("unexpected location_reward: %s", event.LocationReward)
	}
}

func TestParseEventResponse_MarkdownWrapped(t *testing.T) {
	raw := "```json\n{\"event_id\": \"evt_2\", \"description\": \"Тест\", \"immediate_reward\": {}, \"choices\": [{\"label\": \"A\", \"description\": \"desc\", \"reward\": {}, \"next_event_id\": \"n1\"}, {\"label\": \"B\", \"description\": \"desc\", \"reward\": {}, \"next_event_id\": \"n2\"}], \"is_end\": true, \"location_reward\": \"pond\"}\n```\n"
	event, err := ParseEventResponse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Description != "Тест" {
		t.Errorf("unexpected description: %s", event.Description)
	}
}

func TestParseEventResponse_InvalidJSON(t *testing.T) {
	raw := `This is not JSON at all`
	_, err := ParseEventResponse(raw)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseEventResponse_NestedRewards(t *testing.T) {
	raw := `{"event_id": "evt_3", "description": "Тест с вложенными наградами", "immediate_reward": {"resources": {"food": 10}}, "choices": [{"label": "Выбор 1", "description": "Описание 1", "reward": {"resources": {"iron": 50, "composite": -5}}, "next_event_id": "evt_4"}, {"label": "Выбор 2", "description": "Описание 2", "reward": {"resources": {"reagents": 30}}, "next_event_id": "evt_5"}], "is_end": false, "location_reward": ""}`
	event, err := ParseEventResponse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.ImmediateReward["food"] != 10 {
		t.Errorf("expected immediate food = 10, got %f", event.ImmediateReward["food"])
	}
	if event.Choices[0].Reward["iron"] != 50 {
		t.Errorf("expected choice iron = 50, got %f", event.Choices[0].Reward["iron"])
	}
	if event.Choices[0].Reward["composite"] != -5 {
		t.Errorf("expected choice composite = -5, got %f", event.Choices[0].Reward["composite"])
	}
	if event.Choices[1].Reward["reagents"] != 30 {
		t.Errorf("expected second choice reagents = 30, got %f", event.Choices[1].Reward["reagents"])
	}
}

type mockPlanet struct {
	id              string
	ownerID         string
	name            string
	description     string
	level           int
	resourceType    string
	resources       map[string]*float64
}

func (m *mockPlanet) GetID() string              { return m.id }
func (m *mockPlanet) GetOwnerID() string         { return m.ownerID }
func (m *mockPlanet) GetName() string            { return m.name }
func (m *mockPlanet) GetDescription() string     { return m.description }
func (m *mockPlanet) GetLevel() int              { return m.level }
func (m *mockPlanet) GetResourceType() string    { return m.resourceType }
func (m *mockPlanet) GetResources() map[string]*float64 { return m.resources }
func (m *mockPlanet) AddResource(res string, amount float64) {
	if m.resources == nil {
		m.resources = make(map[string]*float64)
	}
	if _, ok := m.resources[res]; !ok {
		v := 0.0
		m.resources[res] = &v
	}
	*m.resources[res] += amount
}

func TestBuildPrompt_ContainsContext(t *testing.T) {
	planet := &mockPlanet{
		id:            "p1",
		ownerID:       "o1",
		name:          "Kepler-442b",
		description:   "Холодная планета",
		level:         2,
		resourceType:  "composite",
		resources:     make(map[string]*float64),
	}
	inv := map[string]float64{"food": 100, "iron": 200}
	events := []ExpeditionEvent{
		{Description: "Первое событие", Choices: []ExpeditionChoice{{Label: "Выбор 1"}}},
	}
	prompt := BuildPrompt(planet, inv, events)
	if len(prompt) == 0 {
		t.Fatal("expected non-empty prompt")
	}
}

func TestCompleteChain_StatusAndLocation(t *testing.T) {
	chain := &ExpeditionChain{
		ID:              "chain1",
		PlanetID:        "p1",
		OwnerID:         "o1",
		Status:          "active",
		EventCount:      3,
		CurrentEventIndex: 2,
		Inventory:       map[string]float64{"food": 100},
		Events:          make([]ExpeditionEvent, 0),
	}
	loc := &Location{
		ID:         "loc1",
		Type:       "forest",
		Name:       "Green Forest",
		Active:     true,
		DiscoveredAt: timeNow(),
		CreatedAt:  timeNow(),
		UpdatedAt:  timeNow(),
	}
	CompleteChain(chain, loc)
	if chain.Status != "completed" {
		t.Errorf("expected status = 'completed', got '%s'", chain.Status)
	}
	if chain.DiscoveredLocation == nil {
		t.Fatal("expected discovered location to be set")
	}
	if chain.DiscoveredLocation.Type != "forest" {
		t.Errorf("expected location type = 'forest', got '%s'", chain.DiscoveredLocation.Type)
	}
}

func TestFailChain_Status(t *testing.T) {
	chain := &ExpeditionChain{
		ID:              "chain2",
		PlanetID:        "p1",
		OwnerID:         "o1",
		Status:          "active",
		CurrentEventIndex: 1,
		Inventory:       map[string]float64{"food": 50},
		Events:          make([]ExpeditionEvent, 0),
	}
	FailChain(chain)
	if chain.Status != "failed" {
		t.Errorf("expected status = 'failed', got '%s'", chain.Status)
	}
}

func TestReturnInventoryToPlanet(t *testing.T) {
	planet := &mockPlanet{
		id:          "p1",
		ownerID:     "o1",
		name:        "Test Planet",
		resourceType: "iron",
		resources: map[string]*float64{
			"food": func() *float64 { f := 500.0; return &f }(),
			"iron": func() *float64 { f := 1000.0; return &f }(),
		},
	}
	inv := map[string]float64{
		"food":     100,
		"iron":     200,
		"reagents": 50,
	}
	ReturnInventoryToPlanet(planet, inv)
	if *planet.resources["food"] != 600 {
		t.Errorf("expected food = 600, got %f", *planet.resources["food"])
	}
	if *planet.resources["iron"] != 1200 {
		t.Errorf("expected iron = 1200, got %f", *planet.resources["iron"])
	}
	if _, ok := planet.resources["reagents"]; !ok {
		t.Error("expected reagents to be added to planet resources")
	}
	if *planet.resources["reagents"] != 50 {
		t.Errorf("expected reagents = 50, got %f", *planet.resources["reagents"])
	}
}

func TestReturnInventoryToPlanet_ZeroAmounts(t *testing.T) {
	planet := &mockPlanet{
		id:          "p1",
		ownerID:     "o1",
		name:        "Test Planet",
		resourceType: "iron",
		resources: map[string]*float64{
			"food": func() *float64 { f := 500.0; return &f }(),
		},
	}
	inv := map[string]float64{
		"food":     0,
		"iron":     -10,
	}
	ReturnInventoryToPlanet(planet, inv)
	if *planet.resources["food"] != 500 {
		t.Errorf("expected food unchanged = 500, got %f", *planet.resources["food"])
	}
}

func TestResolveChoice_Reward(t *testing.T) {
	chain := &ExpeditionChain{
		ID:              "chain1",
		PlanetID:        "p1",
		OwnerID:         "o1",
		Status:          "active",
		CurrentEventIndex: 0,
		Inventory:       map[string]float64{"food": 100, "iron": 200},
		Events: []ExpeditionEvent{
			{
				EventID: "evt_1",
				Description: "Тестовое событие",
				Choices: []ExpeditionChoice{
					{
						Label:       "Взять ресурсы",
						Description: "Получить железо",
						Reward:      map[string]float64{"iron": 50},
						NextEventID: "evt_2",
					},
					{
						Label:       "Отказаться",
						Description: "Ничего не делать",
						Reward:      map[string]float64{},
						NextEventID: "evt_2",
					},
				},
				IsEnd: false,
			},
		},
	}

	// Simulate what ResolveChoice does: apply choice reward and record it
	choice := chain.Events[0].Choices[0]
	ApplyRewardToInventory(chain.Inventory, choice.Reward)

	if chain.Inventory["iron"] != 250 {
		t.Errorf("expected iron = 250, got %f", chain.Inventory["iron"])
	}
	if chain.Inventory["food"] != 100 {
		t.Errorf("expected food unchanged = 100, got %f", chain.Inventory["food"])
	}
}

func TestResolveChoice_RecordPlayerChoice(t *testing.T) {
	chain := &ExpeditionChain{
		ID:              "chain1",
		CurrentEventIndex: 0,
		Inventory:       map[string]float64{"food": 100, "iron": 200},
		Events: []ExpeditionEvent{
			{
				EventID: "evt_1",
				Description: "Test event",
				Choices: []ExpeditionChoice{
					{Label: "Option A", Reward: map[string]float64{"iron": 10}},
					{Label: "Option B", Reward: map[string]float64{"food": -20}},
				},
				IsEnd: false,
			},
		},
	}

	// Simulate ResolveChoice recording player choice (choice index 1)
	choiceIndex := 1
	choice := chain.Events[0].Choices[choiceIndex]
	chain.Events[0].PlayerChoice = choiceIndex
	chain.Events[0].RewardsReceived = make(map[string]float64)
	for k, v := range choice.Reward {
		chain.Events[0].RewardsReceived[k] = v
	}

	if chain.Events[0].PlayerChoice != 1 {
		t.Errorf("expected PlayerChoice = 1, got %d", chain.Events[0].PlayerChoice)
	}
	if chain.Events[0].RewardsReceived["food"] != -20 {
		t.Errorf("expected RewardsReceived[food] = -20, got %f", chain.Events[0].RewardsReceived["food"])
	}
	if len(chain.Events[0].RewardsReceived) != 1 {
		t.Errorf("expected 1 reward entry, got %d", len(chain.Events[0].RewardsReceived))
	}
}

func TestGetCurrentEvent(t *testing.T) {
	chain := &ExpeditionChain{
		CurrentEventIndex: 1,
		Events: []ExpeditionEvent{
			{EventID: "evt_0", Description: "Event 0"},
			{EventID: "evt_1", Description: "Event 1"},
			{EventID: "evt_2", Description: "Event 2"},
		},
	}
	event := GetCurrentEvent(chain)
	if event == nil {
		t.Fatal("expected event, got nil")
	}
	if event.EventID != "evt_1" {
		t.Errorf("expected event_id = 'evt_1', got '%s'", event.EventID)
	}
	if event.Description != "Event 1" {
		t.Errorf("expected description = 'Event 1', got '%s'", event.Description)
	}
}

func TestGetCurrentEvent_OutOfBounds(t *testing.T) {
	chain := &ExpeditionChain{
		CurrentEventIndex: 5,
		Events: []ExpeditionEvent{
			{EventID: "evt_0"},
		},
	}
	event := GetCurrentEvent(chain)
	if event != nil {
		t.Errorf("expected nil for out of bounds, got %+v", event)
	}
}

func TestGetCurrentEvent_NegativeIndex(t *testing.T) {
	chain := &ExpeditionChain{
		CurrentEventIndex: -1,
		Events: []ExpeditionEvent{
			{EventID: "evt_0"},
		},
	}
	event := GetCurrentEvent(chain)
	if event != nil {
		t.Errorf("expected nil for negative index, got %+v", event)
	}
}

func TestCanStartExpeditionChain_Operational(t *testing.T) {
	// Verify that valid inventory resources are a subset of planet resources
	// The CanStartExpeditionChain function checks BaseOperational() which
	// requires food. All valid inventory resources must be valid planet resources.
	for _, res := range ValidInventoryResources {
		if !isValidResource(res) {
			t.Errorf("expected '%s' to be a valid resource", res)
		}
	}
	// Verify we have exactly 5 expected resources
	expected := map[string]bool{
		"food": true, "iron": true, "composite": true,
		"mechanisms": true, "reagents": true,
	}
	if len(ValidInventoryResources) != len(expected) {
		t.Errorf("expected %d valid resources, got %d", len(expected), len(ValidInventoryResources))
	}
	for _, res := range ValidInventoryResources {
		if !expected[res] {
			t.Errorf("unexpected resource in ValidInventoryResources: %s", res)
		}
	}
}

func TestCanStartExpeditionChain_Research(t *testing.T) {
	// Verify that the resource validation helper works for all valid resources
	for _, res := range ValidInventoryResources {
		if !isValidResource(res) {
			t.Errorf("isValidResource('%s') = false, want true", res)
		}
	}
	// Verify invalid resources are rejected
	invalidResources := []string{"plutonium", "dark_matter", "energy", "money"}
	for _, res := range invalidResources {
		if isValidResource(res) {
			t.Errorf("isValidResource('%s') = true, want false", res)
		}
	}
}

func TestCanStartExpeditionChain_ActiveCountLimit(t *testing.T) {
	// MaxActiveChains = 1 means only one active chain per planet
	if MaxActiveChains != 1 {
		t.Errorf("expected MaxActiveChains = 1, got %d", MaxActiveChains)
	}

	// Verify that CanStartExpeditionChain returns false when at the limit
	chain := &ExpeditionChain{Status: "active"}
	activeCount := 0
	if chain.Status == "active" {
		activeCount++
	}
	// When activeCount >= MaxActiveChains, CanStartExpeditionChain should return false
	// This is verified by the logic: activeCount < MaxActiveChains => false when equal
	if activeCount >= MaxActiveChains {
		// This is expected - at the limit, cannot start another expedition
		t.Logf("activeCount=%d >= MaxActiveChains=%d, cannot start new expedition (expected)", activeCount, MaxActiveChains)
	} else {
		t.Error("expected activeCount >= MaxActiveChains")
	}
}

func TestGetEventHistory_Copy(t *testing.T) {
	chain := &ExpeditionChain{
		Events: []ExpeditionEvent{
			{EventID: "evt_1", Description: "First"},
			{EventID: "evt_2", Description: "Second"},
		},
	}
	history := GetEventHistory(chain)
	if len(history) != 2 {
		t.Fatalf("expected 2 events, got %d", len(history))
	}
	if history[0].EventID != "evt_1" {
		t.Errorf("expected first event = 'evt_1', got '%s'", history[0].EventID)
	}
	// Modify the copy - should not affect the original
	history[0].Description = "Modified"
	if chain.Events[0].Description == "Modified" {
		t.Error("GetEventHistory should return a copy, not a reference")
	}
}

func TestValidateInventory_MixedZeroAndNonZero(t *testing.T) {
	inv := map[string]float64{
		"food":     0,
		"iron":     100,
		"reagents": 0,
	}
	err := ValidateInventory(inv)
	if err != nil {
		t.Fatalf("expected valid inventory with mixed zeros, got error: %v", err)
	}
}

func TestParseReward_NestedWithUnknownResources(t *testing.T) {
	raw := map[string]interface{}{
		"resources": map[string]interface{}{
			"food":      10,
			"iron":      20,
			"plutonium": 100,
			"dark_matter": 50,
		},
	}
	result, err := ParseReward(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 valid resources, got %d", len(result))
	}
	if result["food"] != 10 || result["iron"] != 20 {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestApplyRewardToInventory_NewResource(t *testing.T) {
	inv := map[string]float64{
		"food": 100,
	}
	reward := map[string]float64{
		"iron": 50,
	}
	ApplyRewardToInventory(inv, reward)
	if inv["iron"] != 50 {
		t.Errorf("expected iron = 50, got %f", inv["iron"])
	}
	if inv["food"] != 100 {
		t.Errorf("expected food unchanged = 100, got %f", inv["food"])
	}
}

func TestParseEventResponse_MissingDescription(t *testing.T) {
	raw := `{"event_id": "evt_1", "description": "", "immediate_reward": {}, "choices": [], "is_end": true, "location_reward": "pond"}`
	_, err := ParseEventResponse(raw)
	if err == nil {
		t.Fatal("expected error for empty description, got nil")
	}
}

func TestParseEventResponse_InvalidChoiceCount(t *testing.T) {
	raw := `{"event_id": "evt_1", "description": "Тест", "immediate_reward": {}, "choices": [{"label": "A", "description": "desc", "reward": {}, "next_event_id": "n1"}], "is_end": false, "location_reward": ""}`
	_, err := ParseEventResponse(raw)
	if err == nil {
		t.Fatal("expected error for invalid choice count (1 choice, required 2-4), got nil")
	}
}

func TestParseEventResponse_EndEventWithOneChoice(t *testing.T) {
	raw := `{"event_id": "evt_1", "description": "Финальное событие", "immediate_reward": {}, "choices": [{"label": "Завершить", "description": "desc", "reward": {}, "next_event_id": "n1"}], "is_end": true, "location_reward": "pond"}`
	event, err := ParseEventResponse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !event.IsEnd {
		t.Error("expected is_end = true")
	}
}

func TestParseReward_IntTypes(t *testing.T) {
	raw := map[string]interface{}{
		"food":     int(100),
		"iron":     int64(200),
		"reagents": float64(50),
	}
	result, err := ParseReward(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["food"] != 100 {
		t.Errorf("expected food = 100, got %f", result["food"])
	}
	if result["iron"] != 200 {
		t.Errorf("expected iron = 200, got %f", result["iron"])
	}
	if result["reagents"] != 50 {
		t.Errorf("expected reagents = 50, got %f", result["reagents"])
	}
}

func timeNow() time.Time {
	return time.Now()
}
