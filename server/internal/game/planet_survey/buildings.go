package planet_survey

import (
	"spacegame/internal/game/building"
)

var buildingDefs map[string]LocationBuildingDef

func init() {
	buildingDefs = map[string]LocationBuildingDef{
		"fish_farm": {
			BuildingType: "fish_farm",
			Level1Production: building.ProductionResult{Food: 2},
			Level2Production: building.ProductionResult{Food: 4},
			Level3Production: building.ProductionResult{Food: 6},
			SourceConsumption: 1.0,
		},
		"water_purifier": {
			BuildingType: "water_purifier",
			Level1Production: building.ProductionResult{Reagents: 2},
			Level2Production: building.ProductionResult{Reagents: 4},
			Level3Production: building.ProductionResult{Reagents: 6},
			SourceConsumption: 1.0,
		},
		"irrigation_system": {
			BuildingType: "irrigation_system",
			Level1Production: building.ProductionResult{Food: 2},
			Level2Production: building.ProductionResult{Food: 4},
			Level3Production: building.ProductionResult{Food: 6},
			SourceConsumption: 2.0,
		},
		"water_plant": {
			BuildingType: "water_plant",
			Level1Production: building.ProductionResult{Reagents: 2},
			Level2Production: building.ProductionResult{Reagents: 4},
			Level3Production: building.ProductionResult{Reagents: 6},
			SourceConsumption: 2.0,
		},
		"lumber_mill": {
			BuildingType: "lumber_mill",
			Level1Production: building.ProductionResult{Composite: 2},
			Level2Production: building.ProductionResult{Composite: 4},
			Level3Production: building.ProductionResult{Composite: 6},
			SourceConsumption: 1.0,
		},
		"herb_garden": {
			BuildingType: "herb_garden",
			Level1Production: building.ProductionResult{Reagents: 1},
			Level2Production: building.ProductionResult{Reagents: 2},
			Level3Production: building.ProductionResult{Reagents: 3},
			SourceConsumption: 1.5,
		},
		"resin_tap": {
			BuildingType: "resin_tap",
			Level1Production: building.ProductionResult{Composite: 1, Reagents: 1},
			Level2Production: building.ProductionResult{Composite: 2, Reagents: 2},
			Level3Production: building.ProductionResult{Composite: 3, Reagents: 3},
			SourceConsumption: 2.0,
		},
		"mineral_extractor": {
			BuildingType: "mineral_extractor",
			Level1Production: building.ProductionResult{Iron: 3},
			Level2Production: building.ProductionResult{Iron: 6},
			Level3Production: building.ProductionResult{Iron: 9},
			SourceConsumption: 1.0,
		},
		"smelter": {
			BuildingType: "smelter",
			Level1Production: building.ProductionResult{Iron: 5, Money: 2},
			Level2Production: building.ProductionResult{Iron: 10, Money: 4},
			Level3Production: building.ProductionResult{Iron: 15, Money: 6},
			SourceConsumption: 1.5,
		},
		"solar_farm": {
			BuildingType: "solar_farm",
			Level1Production: building.ProductionResult{Energy: 20},
			Level2Production: building.ProductionResult{Energy: 40},
			Level3Production: building.ProductionResult{Energy: 60},
			SourceConsumption: 1.0,
		},
		"wind_turbine": {
			BuildingType: "wind_turbine",
			Level1Production: building.ProductionResult{Energy: 15},
			Level2Production: building.ProductionResult{Energy: 30},
			Level3Production: building.ProductionResult{Energy: 45},
			SourceConsumption: 1.5,
		},
		"hydro_turbine": {
			BuildingType: "hydro_turbine",
			Level1Production: building.ProductionResult{Energy: 25},
			Level2Production: building.ProductionResult{Energy: 50},
			Level3Production: building.ProductionResult{Energy: 75},
			SourceConsumption: 2.0,
		},
		"geothermal_plant": {
			BuildingType: "geothermal_plant",
			Level1Production: building.ProductionResult{Energy: 30},
			Level2Production: building.ProductionResult{Energy: 60},
			Level3Production: building.ProductionResult{Energy: 90},
			SourceConsumption: 1.0,
		},
		"hot_spring_resort": {
			BuildingType: "hot_spring_resort",
			Level1Production: building.ProductionResult{Money: 5},
			Level2Production: building.ProductionResult{Money: 10},
			Level3Production: building.ProductionResult{Money: 15},
			SourceConsumption: 1.5,
		},
		"salt_mine": {
			BuildingType: "salt_mine",
			Level1Production: building.ProductionResult{Reagents: 3},
			Level2Production: building.ProductionResult{Reagents: 6},
			Level3Production: building.ProductionResult{Reagents: 9},
			SourceConsumption: 1.0,
		},
		"chemical_plant": {
			BuildingType: "chemical_plant",
			Level1Production: building.ProductionResult{Reagents: 2, Composite: 1},
			Level2Production: building.ProductionResult{Reagents: 4, Composite: 2},
			Level3Production: building.ProductionResult{Reagents: 6, Composite: 3},
			SourceConsumption: 1.5,
		},
		"weather_station": {
			BuildingType: "weather_station",
			Level1Production: building.ProductionResult{Money: 3},
			Level2Production: building.ProductionResult{Money: 6},
			Level3Production: building.ProductionResult{Money: 9},
			SourceConsumption: 2.0,
		},
		"crystal_harvester": {
			BuildingType: "crystal_harvester",
			Level1Production: building.ProductionResult{AlienTech: 2},
			Level2Production: building.ProductionResult{AlienTech: 4},
			Level3Production: building.ProductionResult{AlienTech: 6},
			SourceConsumption: 1.5,
		},
		"gem_cutter": {
			BuildingType: "gem_cutter",
			Level1Production: building.ProductionResult{Money: 8},
			Level2Production: building.ProductionResult{Money: 16},
			Level3Production: building.ProductionResult{Money: 24},
			SourceConsumption: 2.0,
		},
		"metal_scavenger": {
			BuildingType: "metal_scavenger",
			Level1Production: building.ProductionResult{Iron: 5},
			Level2Production: building.ProductionResult{Iron: 10},
			Level3Production: building.ProductionResult{Iron: 15},
			SourceConsumption: 1.5,
		},
		"alloy_furnace": {
			BuildingType: "alloy_furnace",
			Level1Production: building.ProductionResult{Iron: 3, Composite: 2},
			Level2Production: building.ProductionResult{Iron: 6, Composite: 4},
			Level3Production: building.ProductionResult{Iron: 9, Composite: 6},
			SourceConsumption: 2.0,
		},
		"artifact_excavator": {
			BuildingType: "artifact_excavator",
			Level1Production: building.ProductionResult{AlienTech: 3},
			Level2Production: building.ProductionResult{AlienTech: 6},
			Level3Production: building.ProductionResult{AlienTech: 9},
			SourceConsumption: 2.0,
		},
		"ancient_library": {
			BuildingType: "ancient_library",
			Level1Production: building.ProductionResult{AlienTech: 2, Money: 5},
			Level2Production: building.ProductionResult{AlienTech: 4, Money: 10},
			Level3Production: building.ProductionResult{AlienTech: 6, Money: 15},
			SourceConsumption: 2.5,
		},
		"ice_miner": {
			BuildingType: "ice_miner",
			Level1Production: building.ProductionResult{Reagents: 2, Energy: 10},
			Level2Production: building.ProductionResult{Reagents: 4, Energy: 20},
			Level3Production: building.ProductionResult{Reagents: 6, Energy: 30},
			SourceConsumption: 2.0,
		},
		"spore_collector": {
			BuildingType: "spore_collector",
			Level1Production: building.ProductionResult{Composite: 3, Reagents: 1},
			Level2Production: building.ProductionResult{Composite: 6, Reagents: 2},
			Level3Production: building.ProductionResult{Composite: 9, Reagents: 3},
			SourceConsumption: 2.0,
		},
		"energy_collector": {
			BuildingType: "energy_collector",
			Level1Production: building.ProductionResult{Energy: 40, Money: 5},
			Level2Production: building.ProductionResult{Energy: 80, Money: 10},
			Level3Production: building.ProductionResult{Energy: 120, Money: 15},
			SourceConsumption: 2.5,
		},
		"sky_fish_farm": {
			BuildingType: "sky_fish_farm",
			Level1Production: building.ProductionResult{Food: 5, Money: 3},
			Level2Production: building.ProductionResult{Food: 10, Money: 6},
			Level3Production: building.ProductionResult{Food: 15, Money: 9},
			SourceConsumption: 2.0,
		},
		"rare_fish_habitat": {
			BuildingType: "rare_fish_habitat",
			Level1Production: building.ProductionResult{Food: 4, Reagents: 2},
			Level2Production: building.ProductionResult{Food: 8, Reagents: 4},
			Level3Production: building.ProductionResult{Food: 12, Reagents: 6},
			SourceConsumption: 2.5,
		},
		"isotope_extractor": {
			BuildingType: "isotope_extractor",
			Level1Production: building.ProductionResult{AlienTech: 4},
			Level2Production: building.ProductionResult{AlienTech: 8},
			Level3Production: building.ProductionResult{AlienTech: 12},
			SourceConsumption: 2.0,
		},
		"waste_processor": {
			BuildingType: "waste_processor",
			Level1Production: building.ProductionResult{AlienTech: 2, Money: 5},
			Level2Production: building.ProductionResult{AlienTech: 4, Money: 10},
			Level3Production: building.ProductionResult{AlienTech: 6, Money: 15},
			SourceConsumption: 2.5,
		},
		"anomaly_analyzer": {
			BuildingType: "anomaly_analyzer",
			Level1Production: building.ProductionResult{AlienTech: 5, Energy: 15},
			Level2Production: building.ProductionResult{AlienTech: 10, Energy: 30},
			Level3Production: building.ProductionResult{AlienTech: 15, Energy: 45},
			SourceConsumption: 2.5,
		},
		"quantum_generator": {
			BuildingType: "quantum_generator",
			Level1Production: building.ProductionResult{Energy: 50, AlienTech: 3},
			Level2Production: building.ProductionResult{Energy: 100, AlienTech: 6},
			Level3Production: building.ProductionResult{Energy: 150, AlienTech: 9},
			SourceConsumption: 3.0,
		},
	}
}

func GetBuildingDef(buildingType string) *LocationBuildingDef {
	def, ok := buildingDefs[buildingType]
	if !ok {
		return nil
	}
	return &def
}

func GetProduction(buildingType string, level int) building.ProductionResult {
	def, ok := buildingDefs[buildingType]
	if !ok {
		return building.ProductionResult{}
	}

	switch level {
	case 1:
		return def.Level1Production
	case 2:
		return def.Level2Production
	case 3:
		return def.Level3Production
	default:
		return def.Level1Production
	}
}
