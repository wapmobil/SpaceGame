package planet_survey

import (
	"testing"
)

func TestAll20LocationTypesDefined(t *testing.T) {
	types := GetLocationTypes()
	expectedTypes := []string{
		"pond", "river", "forest", "mineral_deposit", "dry_valley",
		"waterfall", "cave", "thermal_spring", "salt_lake", "wind_pass",
		"crystal_cave", "meteor_crater", "sunken_city", "glacier", "mushroom_forest",
		"crystal_field", "cloud_island", "underground_lake", "radioactive_zone", "anomaly_zone",
	}

	if len(types) != len(expectedTypes) {
		t.Fatalf("expected %d location types, got %d", len(expectedTypes), len(types))
	}

	existingTypes := make(map[string]bool)
	for _, lt := range types {
		existingTypes[lt.Type] = true
	}

	for _, expected := range expectedTypes {
		if !existingTypes[expected] {
			t.Errorf("expected location type '%s' to be defined", expected)
		}
	}
}

func TestAllBuildingsHaveProductionPerLevel(t *testing.T) {
	types := GetLocationTypes()
	for _, lt := range types {
		for _, bldg := range lt.Buildings {
			l1 := bldg.Level1Production
			l2 := bldg.Level2Production
			l3 := bldg.Level3Production

			if l1.IsZero() && l2.IsZero() && l3.IsZero() {
				t.Errorf("building '%s' in location type '%s' has zero production at all levels",
					bldg.BuildingType, lt.Type)
			}

			if l2.Food < l1.Food && l1.Food > 0 {
				t.Errorf("building '%s' level 2 food (%f) should be >= level 1 food (%f)",
					bldg.BuildingType, l2.Food, l1.Food)
			}

			if l3.Food < l2.Food && l2.Food > 0 {
				t.Errorf("building '%s' level 3 food (%f) should be >= level 2 food (%f)",
					bldg.BuildingType, l3.Food, l2.Food)
			}
		}
	}
}

func TestAllBuildingsHaveSourceConsumption(t *testing.T) {
	types := GetLocationTypes()
	for _, lt := range types {
		for _, bldg := range lt.Buildings {
			if bldg.SourceConsumption <= 0 {
				t.Errorf("building '%s' in location type '%s' has non-positive source consumption: %f",
					bldg.BuildingType, lt.Type, bldg.SourceConsumption)
			}
		}
	}
}

func TestLocationTypeBuildingsCount(t *testing.T) {
	types := GetLocationTypes()
	for _, lt := range types {
		if len(lt.Buildings) < 1 {
			t.Errorf("location type '%s' should have at least 1 building, got %d", lt.Type, len(lt.Buildings))
		}
		if len(lt.Buildings) > 3 {
			t.Errorf("location type '%s' should have at most 3 buildings, got %d", lt.Type, len(lt.Buildings))
		}
	}
}

func TestGetBuildingDef_AllTypesHaveDefs(t *testing.T) {
	types := GetLocationTypes()
	for _, lt := range types {
		for _, bldg := range lt.Buildings {
			def := GetBuildingDef(bldg.BuildingType)
			if def == nil {
				t.Errorf("expected building def for '%s' (from location type '%s')", bldg.BuildingType, lt.Type)
			}
			if def.BuildingType != bldg.BuildingType {
				t.Errorf("building def type mismatch: expected '%s', got '%s'", bldg.BuildingType, def.BuildingType)
			}
		}
	}
}

func TestBuildingProductionScalesWithLevel(t *testing.T) {
	buildings := []string{
		"fish_farm", "water_purifier", "mineral_extractor",
		"solar_farm", "wind_turbine", "crystal_harvester",
		"isotope_extractor", "anomaly_analyzer",
	}

	for _, bt := range buildings {
		l1 := GetProduction(bt, 1)
		l2 := GetProduction(bt, 2)
		l3 := GetProduction(bt, 3)

		total1 := l1.Food + l1.Iron + l1.Composite + l1.Mechanisms + l1.Reagents + l1.Energy + l1.Money + l1.AlienTech
		total2 := l2.Food + l2.Iron + l2.Composite + l2.Mechanisms + l2.Reagents + l2.Energy + l2.Money + l2.AlienTech
		total3 := l3.Food + l3.Iron + l3.Composite + l3.Mechanisms + l3.Reagents + l3.Energy + l3.Money + l3.AlienTech

		if total1 > 0 {
			if total2 < total1 {
				t.Errorf("building '%s': level 2 total production (%f) < level 1 total (%f)", bt, total2, total1)
			}
			if total3 < total2 {
				t.Errorf("building '%s': level 3 total production (%f) < level 2 total (%f)", bt, total3, total2)
			}
		}
	}
}
