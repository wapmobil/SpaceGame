package planet_survey

import (
	"testing"
	"time"
)

func TestGetLocationTypes_Count(t *testing.T) {
	types := GetLocationTypes()
	if len(types) != 20 {
		t.Errorf("expected 20 location types, got %d", len(types))
	}
}

func TestGetLocationTypes_RarityWeights(t *testing.T) {
	types := GetLocationTypes()
	ordinaryCount := 0
	uncommonCount := 0
	rareCount := 0
	exoticCount := 0

	for _, lt := range types {
		switch lt.RarityWeight {
		case 30:
			ordinaryCount++
		case 20:
			uncommonCount++
		case 12:
			rareCount++
		case 6:
			exoticCount++
		}
	}

	if ordinaryCount != 5 {
		t.Errorf("expected 5 ordinary types, got %d", ordinaryCount)
	}
	if uncommonCount != 5 {
		t.Errorf("expected 5 uncommon types, got %d", uncommonCount)
	}
	if rareCount != 5 {
		t.Errorf("expected 5 rare types, got %d", rareCount)
	}
	if exoticCount != 5 {
		t.Errorf("expected 5 exotic types, got %d", exoticCount)
	}
}

func TestGetLocationTypes_AllHaveBuildings(t *testing.T) {
	types := GetLocationTypes()
	for _, lt := range types {
		if len(lt.Buildings) == 0 {
			t.Errorf("location type '%s' has no buildings", lt.Type)
		}
	}
}

func TestGetLocationTypes_AllHaveSourceResource(t *testing.T) {
	types := GetLocationTypes()
	for _, lt := range types {
		if lt.SourceResource == "" {
			t.Errorf("location type '%s' has no source resource", lt.Type)
		}
	}
}

func TestGetLocationTypes_AllHaveAmountRange(t *testing.T) {
	types := GetLocationTypes()
	for _, lt := range types {
		if lt.AmountRange[0] >= lt.AmountRange[1] {
			t.Errorf("location type '%s' has invalid amount range [%f, %f]", lt.Type, lt.AmountRange[0], lt.AmountRange[1])
		}
	}
}

func TestSelectLocationType_ReturnsValidType(t *testing.T) {
	for range 100 {
		typ := SelectLocationType(ResourceComposite)
		found := false
		for _, lt := range GetLocationTypes() {
			if lt.Type == typ {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SelectLocationType returned unknown type: %s", typ)
		}
	}
}

func TestSelectLocationType_ResourceTypeBias(t *testing.T) {
	foodTypes := 0
	reagentTypes := 0
	totalRuns := 200

	for i := 0; i < totalRuns; i++ {
		typ := SelectLocationType(ResourceReagents)
		for _, lt := range GetLocationTypes() {
			if lt.Type == typ {
				if lt.SourceResource == "reagents" {
					reagentTypes++
				}
				if lt.SourceResource == "food" {
					foodTypes++
				}
			}
		}
	}

	reagentRatio := float64(reagentTypes) / float64(totalRuns)
	if reagentRatio < 0.3 {
		t.Errorf("expected reagent types to be selected at least 30%% of the time when ResourceType=reagents, got %.1f%%", reagentRatio*100)
	}
}

func TestSelectLocationType_ProducesVariety(t *testing.T) {
	typesFound := make(map[string]bool)
	for range 100 {
		typ := SelectLocationType(ResourceComposite)
		typesFound[typ] = true
	}
	if len(typesFound) < 5 {
		t.Errorf("expected at least 5 different types in 100 runs, got %d", len(typesFound))
	}
}

func TestCalculateSourceAmount_RarityMultiplier(t *testing.T) {
	ordinaryTypes := []string{"pond", "river", "forest", "mineral_deposit", "dry_valley"}
	exoticTypes := []string{"crystal_field", "cloud_island", "underground_lake", "radioactive_zone", "anomaly_zone"}

	for _, ot := range ordinaryTypes {
		amount := CalculateSourceAmount(ot, ResourceComposite)
		if amount < 25 {
			t.Errorf("expected ordinary type '%s' amount >= 25, got %f", ot, amount)
		}
	}

	for _, et := range exoticTypes {
		amount := CalculateSourceAmount(et, ResourceComposite)
		if amount < 125 {
			t.Errorf("expected exotic type '%s' amount >= 125, got %f", et, amount)
		}
	}
}

func TestCalculateSourceAmount_ResourceTypeBias(t *testing.T) {
	// Test with matching resource type - should have higher amounts
	matchingAmounts := make([]float64, 0)
	for range 20 {
		amount := CalculateSourceAmount("pond", ResourceReagents)
		matchingAmounts = append(matchingAmounts, amount)
	}

	matchingAvg := 0.0
	for _, a := range matchingAmounts {
		matchingAvg += a
	}
	matchingAvg /= float64(len(matchingAmounts))

	// Test with non-matching resource type
	nonMatchingAmounts := make([]float64, 0)
	for range 20 {
		amount := CalculateSourceAmount("pond", ResourceComposite)
		nonMatchingAmounts = append(nonMatchingAmounts, amount)
	}

	nonMatchingAvg := 0.0
	for _, a := range nonMatchingAmounts {
		nonMatchingAvg += a
	}
	nonMatchingAvg /= float64(len(nonMatchingAmounts))

	if matchingAvg > 0 && nonMatchingAvg > 0 {
		if matchingAvg < nonMatchingAvg {
			t.Errorf("expected matching resource type to have higher amounts (matching=%.1f, non-matching=%.1f)", matchingAvg, nonMatchingAvg)
		}
	}
}

func TestGenerateName_ReturnsNonEmpty(t *testing.T) {
	for _, lt := range GetLocationTypes() {
		name := GenerateName(lt.Type)
		if name == "" {
			t.Errorf("GenerateName returned empty string for type '%s'", lt.Type)
		}
		if name == "Unknown Location" {
			t.Errorf("GenerateName returned 'Unknown Location' for type '%s'", lt.Type)
		}
	}
}

func TestGenerateName_ProvidesVariety(t *testing.T) {
	names := make(map[string]bool)
	for range 20 {
		name := GenerateName("pond")
		names[name] = true
	}
	if len(names) < 2 {
		t.Errorf("expected at least 2 different names for pond in 20 runs, got %d", len(names))
	}
}

func TestTickLocationBuildings_ProductionApplied(t *testing.T) {
	loc := &Location{
		ID:              "loc-1",
		PlanetID:        "planet-1",
		OwnerID:         "owner-1",
		Type:            "pond",
		Name:            "Test Pond",
		BuildingType:    "fish_farm",
		BuildingLevel:   1,
		BuildingActive:  true,
		SourceResource:  "food",
		SourceAmount:    100,
		SourceRemaining: 100,
		Active:          true,
		DiscoveredAt:    time.Now(),
	}

	lbs := []*LocationBuilding{
		{
			ID:           "lb-1",
			LocationID:   "loc-1",
			BuildingType: "fish_farm",
			Level:        1,
			Active:       true,
		},
	}

	resources := map[string]float64{
		"food": 100, "iron": 50, "composite": 0, "mechanisms": 0, "reagents": 0,
		"energy": 20, "money": 500, "alien_tech": 0,
	}

	TickLocationBuildings(loc, lbs, resources)

	if resources["food"] <= 100 {
		t.Errorf("expected food to increase, initial=100, final=%f", resources["food"])
	}

	if loc.SourceRemaining >= 100 {
		t.Errorf("expected source remaining to decrease, initial=100, final=%f", loc.SourceRemaining)
	}
}

func TestTickLocationBuildings_Depletion(t *testing.T) {
	loc := &Location{
		ID:              "loc-1",
		PlanetID:        "planet-1",
		OwnerID:         "owner-1",
		Type:            "pond",
		Name:            "Test Pond",
		BuildingType:    "fish_farm",
		BuildingLevel:   1,
		BuildingActive:  true,
		SourceResource:  "food",
		SourceAmount:    100,
		SourceRemaining: 0.5,
		Active:          true,
		DiscoveredAt:    time.Now(),
	}

	lbs := []*LocationBuilding{
		{
			ID:           "lb-1",
			LocationID:   "loc-1",
			BuildingType: "fish_farm",
			Level:        1,
			Active:       true,
		},
	}

	resources := map[string]float64{
		"food": 100, "iron": 50, "composite": 0, "mechanisms": 0, "reagents": 0,
		"energy": 20, "money": 500, "alien_tech": 0,
	}

	TickLocationBuildings(loc, lbs, resources)

	if loc.BuildingActive {
		t.Error("expected building to be inactive after depletion")
	}
	if loc.Active {
		t.Error("expected location to be inactive after depletion")
	}
}

func TestTickLocationBuildings_InactiveBuilding(t *testing.T) {
	loc := &Location{
		ID:              "loc-1",
		PlanetID:        "planet-1",
		OwnerID:         "owner-1",
		Type:            "pond",
		Name:            "Test Pond",
		BuildingType:    "fish_farm",
		BuildingLevel:   1,
		BuildingActive:  true,
		SourceResource:  "food",
		SourceAmount:    100,
		SourceRemaining: 100,
		Active:          true,
		DiscoveredAt:    time.Now(),
	}

	lbs := []*LocationBuilding{
		{
			ID:           "lb-1",
			LocationID:   "loc-1",
			BuildingType: "fish_farm",
			Level:        1,
			Active:       false,
		},
	}

	resources := map[string]float64{
		"food": 100, "iron": 50, "composite": 0, "mechanisms": 0, "reagents": 0,
		"energy": 20, "money": 500, "alien_tech": 0,
	}

	TickLocationBuildings(loc, lbs, resources)

	if resources["food"] != 100 {
		t.Errorf("expected no production from inactive building, food should remain 100, got %f", resources["food"])
	}
}

func TestGetBuildingDef_AllBuildingsExist(t *testing.T) {
	expectedBuildings := []string{
		"fish_farm", "water_purifier", "irrigation_system", "water_plant",
		"lumber_mill", "herb_garden", "resin_tap",
		"mineral_extractor", "smelter",
		"solar_farm", "wind_turbine", "hydro_turbine",
		"geothermal_plant", "hot_spring_resort",
		"salt_mine", "chemical_plant",
		"weather_station",
		"crystal_harvester", "gem_cutter",
		"metal_scavenger", "alloy_furnace",
		"artifact_excavator", "ancient_library",
		"ice_miner",
		"spore_collector",
		"energy_collector", "sky_fish_farm",
		"rare_fish_habitat",
		"isotope_extractor", "waste_processor",
		"anomaly_analyzer", "quantum_generator",
	}

	for _, bt := range expectedBuildings {
		def := GetBuildingDef(bt)
		if def == nil {
			t.Errorf("expected building def for '%s', got nil", bt)
		}
	}
}

func TestGetProduction_Level1(t *testing.T) {
	prod := GetProduction("fish_farm", 1)
	if prod.Food != 2 {
		t.Errorf("expected fish_farm level 1 food production to be 2, got %f", prod.Food)
	}
}

func TestGetProduction_Level2(t *testing.T) {
	prod := GetProduction("fish_farm", 2)
	if prod.Food != 4 {
		t.Errorf("expected fish_farm level 2 food production to be 4, got %f", prod.Food)
	}
}

func TestGetProduction_Level3(t *testing.T) {
	prod := GetProduction("fish_farm", 3)
	if prod.Food != 6 {
		t.Errorf("expected fish_farm level 3 food production to be 6, got %f", prod.Food)
	}
}

func TestGetProduction_UnknownBuilding(t *testing.T) {
	prod := GetProduction("unknown_building", 1)
	if !prod.IsZero() {
		t.Errorf("expected zero production for unknown building, got %+v", prod)
	}
}

func TestGetProduction_InvalidLevel(t *testing.T) {
	prod := GetProduction("fish_farm", 0)
	expected := GetProduction("fish_farm", 1)
	if prod.Food != expected.Food {
		t.Errorf("expected level 0 to return level 1 production, got %f", prod.Food)
	}
}
