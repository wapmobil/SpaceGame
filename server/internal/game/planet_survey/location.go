package planet_survey

import (
	"math"
	"math/rand"
	"time"

	"spacegame/internal/game/building"
)

type PlanetResourceType string

const (
	ResourceComposite  PlanetResourceType = "composite"
	ResourceMechanisms PlanetResourceType = "mechanisms"
	ResourceReagents   PlanetResourceType = "reagents"
)

type Location struct {
	ID              string
	PlanetID        string
	OwnerID         string
	Type            string
	Name            string
	BuildingType    string
	BuildingLevel   int
	BuildingActive  bool
	Buildings       []*LocationBuilding
	SourceResource  string
	SourceAmount    float64
	SourceRemaining float64
	Active          bool
	DiscoveredAt    time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type LocationBuildingDef struct {
	BuildingType      string
	Level1Production  building.ProductionResult
	Level2Production  building.ProductionResult
	Level3Production  building.ProductionResult
	SourceConsumption float64
}

type LocationType struct {
	Type          string
	Name          string
	Buildings     []LocationBuildingDef
	SourceResource string
	AmountRange   [2]float64
	RarityWeight  int
}

var locationTypes []LocationType

func init() {
	locationTypes = []LocationType{
		{
			Type: "pond", Name: "Pond",
			SourceResource: "food",
			AmountRange:    [2]float64{50, 150},
			RarityWeight:   30,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "fish_farm",
					Level1Production: building.ProductionResult{Food: 2},
					Level2Production: building.ProductionResult{Food: 4},
					Level3Production: building.ProductionResult{Food: 6},
					SourceConsumption: 1.0,
				},
				{
					BuildingType: "water_purifier",
					Level1Production: building.ProductionResult{Reagents: 1},
					Level2Production: building.ProductionResult{Reagents: 2},
					Level3Production: building.ProductionResult{Reagents: 3},
					SourceConsumption: 1.0,
				},
			},
		},
		{
			Type: "river", Name: "River",
			SourceResource: "food",
			AmountRange:    [2]float64{80, 250},
			RarityWeight:   30,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "fish_farm",
					Level1Production: building.ProductionResult{Food: 3},
					Level2Production: building.ProductionResult{Food: 6},
					Level3Production: building.ProductionResult{Food: 9},
					SourceConsumption: 1.5,
				},
				{
					BuildingType: "irrigation_system",
					Level1Production: building.ProductionResult{Food: 2},
					Level2Production: building.ProductionResult{Food: 4},
					Level3Production: building.ProductionResult{Food: 6},
					SourceConsumption: 2.0,
				},
				{
					BuildingType: "water_plant",
					Level1Production: building.ProductionResult{Reagents: 2},
					Level2Production: building.ProductionResult{Reagents: 4},
					Level3Production: building.ProductionResult{Reagents: 6},
					SourceConsumption: 2.0,
				},
			},
		},
		{
			Type: "forest", Name: "Forest",
			SourceResource: "composite",
			AmountRange:    [2]float64{100, 300},
			RarityWeight:   30,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "lumber_mill",
					Level1Production: building.ProductionResult{Composite: 2},
					Level2Production: building.ProductionResult{Composite: 4},
					Level3Production: building.ProductionResult{Composite: 6},
					SourceConsumption: 1.0,
				},
				{
					BuildingType: "herb_garden",
					Level1Production: building.ProductionResult{Reagents: 1},
					Level2Production: building.ProductionResult{Reagents: 2},
					Level3Production: building.ProductionResult{Reagents: 3},
					SourceConsumption: 1.5,
				},
				{
					BuildingType: "resin_tap",
					Level1Production: building.ProductionResult{Composite: 1, Reagents: 1},
					Level2Production: building.ProductionResult{Composite: 2, Reagents: 2},
					Level3Production: building.ProductionResult{Composite: 3, Reagents: 3},
					SourceConsumption: 2.0,
				},
			},
		},
		{
			Type: "mineral_deposit", Name: "Mineral Deposit",
			SourceResource: "iron",
			AmountRange:    [2]float64{100, 400},
			RarityWeight:   30,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "mineral_extractor",
					Level1Production: building.ProductionResult{Iron: 3},
					Level2Production: building.ProductionResult{Iron: 6},
					Level3Production: building.ProductionResult{Iron: 9},
					SourceConsumption: 1.0,
				},
				{
					BuildingType: "smelter",
					Level1Production: building.ProductionResult{Iron: 5, Money: 2},
					Level2Production: building.ProductionResult{Iron: 10, Money: 4},
					Level3Production: building.ProductionResult{Iron: 15, Money: 6},
					SourceConsumption: 1.5,
				},
			},
		},
		{
			Type: "dry_valley", Name: "Dry Valley",
			SourceResource: "energy",
			AmountRange:    [2]float64{50, 200},
			RarityWeight:   30,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "solar_farm",
					Level1Production: building.ProductionResult{Energy: 20},
					Level2Production: building.ProductionResult{Energy: 40},
					Level3Production: building.ProductionResult{Energy: 60},
					SourceConsumption: 1.0,
				},
				{
					BuildingType: "wind_turbine",
					Level1Production: building.ProductionResult{Energy: 15},
					Level2Production: building.ProductionResult{Energy: 30},
					Level3Production: building.ProductionResult{Energy: 45},
					SourceConsumption: 1.5,
				},
			},
		},
		{
			Type: "waterfall", Name: "Waterfall",
			SourceResource: "food",
			AmountRange:    [2]float64{100, 300},
			RarityWeight:   20,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "water_purifier",
					Level1Production: building.ProductionResult{Reagents: 2},
					Level2Production: building.ProductionResult{Reagents: 4},
					Level3Production: building.ProductionResult{Reagents: 6},
					SourceConsumption: 1.5,
				},
				{
					BuildingType: "hydro_turbine",
					Level1Production: building.ProductionResult{Energy: 25},
					Level2Production: building.ProductionResult{Energy: 50},
					Level3Production: building.ProductionResult{Energy: 75},
					SourceConsumption: 2.0,
				},
			},
		},
		{
			Type: "cave", Name: "Cave",
			SourceResource: "iron",
			AmountRange:    [2]float64{150, 500},
			RarityWeight:   20,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "mineral_extractor",
					Level1Production: building.ProductionResult{Iron: 4},
					Level2Production: building.ProductionResult{Iron: 8},
					Level3Production: building.ProductionResult{Iron: 12},
					SourceConsumption: 1.5,
				},
				{
					BuildingType: "crystal_harvester",
					Level1Production: building.ProductionResult{AlienTech: 1},
					Level2Production: building.ProductionResult{AlienTech: 2},
					Level3Production: building.ProductionResult{AlienTech: 3},
					SourceConsumption: 2.0,
				},
			},
		},
		{
			Type: "thermal_spring", Name: "Thermal Spring",
			SourceResource: "energy",
			AmountRange:    [2]float64{80, 300},
			RarityWeight:   20,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "geothermal_plant",
					Level1Production: building.ProductionResult{Energy: 30},
					Level2Production: building.ProductionResult{Energy: 60},
					Level3Production: building.ProductionResult{Energy: 90},
					SourceConsumption: 1.0,
				},
				{
					BuildingType: "hot_spring_resort",
					Level1Production: building.ProductionResult{Money: 5},
					Level2Production: building.ProductionResult{Money: 10},
					Level3Production: building.ProductionResult{Money: 15},
					SourceConsumption: 1.5,
				},
			},
		},
		{
			Type: "salt_lake", Name: "Salt Lake",
			SourceResource: "reagents",
			AmountRange:    [2]float64{100, 400},
			RarityWeight:   20,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "salt_mine",
					Level1Production: building.ProductionResult{Reagents: 3},
					Level2Production: building.ProductionResult{Reagents: 6},
					Level3Production: building.ProductionResult{Reagents: 9},
					SourceConsumption: 1.0,
				},
				{
					BuildingType: "chemical_plant",
					Level1Production: building.ProductionResult{Reagents: 2, Composite: 1},
					Level2Production: building.ProductionResult{Reagents: 4, Composite: 2},
					Level3Production: building.ProductionResult{Reagents: 6, Composite: 3},
					SourceConsumption: 1.5,
				},
			},
		},
		{
			Type: "wind_pass", Name: "Wind Pass",
			SourceResource: "energy",
			AmountRange:    [2]float64{100, 350},
			RarityWeight:   20,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "wind_turbine",
					Level1Production: building.ProductionResult{Energy: 20},
					Level2Production: building.ProductionResult{Energy: 40},
					Level3Production: building.ProductionResult{Energy: 60},
					SourceConsumption: 1.5,
				},
				{
					BuildingType: "weather_station",
					Level1Production: building.ProductionResult{Money: 3},
					Level2Production: building.ProductionResult{Money: 6},
					Level3Production: building.ProductionResult{Money: 9},
					SourceConsumption: 2.0,
				},
			},
		},
		{
			Type: "crystal_cave", Name: "Crystal Cave",
			SourceResource: "reagents",
			AmountRange:    [2]float64{200, 600},
			RarityWeight:   12,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "crystal_harvester",
					Level1Production: building.ProductionResult{AlienTech: 2},
					Level2Production: building.ProductionResult{AlienTech: 4},
					Level3Production: building.ProductionResult{AlienTech: 6},
					SourceConsumption: 1.5,
				},
				{
					BuildingType: "gem_cutter",
					Level1Production: building.ProductionResult{Money: 8},
					Level2Production: building.ProductionResult{Money: 16},
					Level3Production: building.ProductionResult{Money: 24},
					SourceConsumption: 2.0,
				},
			},
		},
		{
			Type: "meteor_crater", Name: "Meteor Crater",
			SourceResource: "iron",
			AmountRange:    [2]float64{200, 700},
			RarityWeight:   12,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "metal_scavenger",
					Level1Production: building.ProductionResult{Iron: 5},
					Level2Production: building.ProductionResult{Iron: 10},
					Level3Production: building.ProductionResult{Iron: 15},
					SourceConsumption: 1.5,
				},
				{
					BuildingType: "alloy_furnace",
					Level1Production: building.ProductionResult{Iron: 3, Composite: 2},
					Level2Production: building.ProductionResult{Iron: 6, Composite: 4},
					Level3Production: building.ProductionResult{Iron: 9, Composite: 6},
					SourceConsumption: 2.0,
				},
			},
		},
		{
			Type: "sunken_city", Name: "Sunken City",
			SourceResource: "reagents",
			AmountRange:    [2]float64{300, 800},
			RarityWeight:   12,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "artifact_excavator",
					Level1Production: building.ProductionResult{AlienTech: 3},
					Level2Production: building.ProductionResult{AlienTech: 6},
					Level3Production: building.ProductionResult{AlienTech: 9},
					SourceConsumption: 2.0,
				},
				{
					BuildingType: "ancient_library",
					Level1Production: building.ProductionResult{AlienTech: 2, Money: 5},
					Level2Production: building.ProductionResult{AlienTech: 4, Money: 10},
					Level3Production: building.ProductionResult{AlienTech: 6, Money: 15},
					SourceConsumption: 2.5,
				},
			},
		},
		{
			Type: "glacier", Name: "Glacier",
			SourceResource: "food",
			AmountRange:    [2]float64{150, 500},
			RarityWeight:   12,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "water_purifier",
					Level1Production: building.ProductionResult{Reagents: 3},
					Level2Production: building.ProductionResult{Reagents: 6},
					Level3Production: building.ProductionResult{Reagents: 9},
					SourceConsumption: 1.5,
				},
				{
					BuildingType: "ice_miner",
					Level1Production: building.ProductionResult{Reagents: 2, Energy: 10},
					Level2Production: building.ProductionResult{Reagents: 4, Energy: 20},
					Level3Production: building.ProductionResult{Reagents: 6, Energy: 30},
					SourceConsumption: 2.0,
				},
			},
		},
		{
			Type: "mushroom_forest", Name: "Mushroom Forest",
			SourceResource: "composite",
			AmountRange:    [2]float64{150, 500},
			RarityWeight:   12,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "herb_garden",
					Level1Production: building.ProductionResult{Reagents: 2},
					Level2Production: building.ProductionResult{Reagents: 4},
					Level3Production: building.ProductionResult{Reagents: 6},
					SourceConsumption: 1.5,
				},
				{
					BuildingType: "spore_collector",
					Level1Production: building.ProductionResult{Composite: 3, Reagents: 1},
					Level2Production: building.ProductionResult{Composite: 6, Reagents: 2},
					Level3Production: building.ProductionResult{Composite: 9, Reagents: 3},
					SourceConsumption: 2.0,
				},
			},
		},
		{
			Type: "crystal_field", Name: "Crystal Field",
			SourceResource: "reagents",
			AmountRange:    [2]float64{300, 1000},
			RarityWeight:   6,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "crystal_harvester",
					Level1Production: building.ProductionResult{AlienTech: 3, Energy: 10},
					Level2Production: building.ProductionResult{AlienTech: 6, Energy: 20},
					Level3Production: building.ProductionResult{AlienTech: 9, Energy: 30},
					SourceConsumption: 2.0,
				},
				{
					BuildingType: "energy_collector",
					Level1Production: building.ProductionResult{Energy: 40, Money: 5},
					Level2Production: building.ProductionResult{Energy: 80, Money: 10},
					Level3Production: building.ProductionResult{Energy: 120, Money: 15},
					SourceConsumption: 2.5,
				},
			},
		},
		{
			Type: "cloud_island", Name: "Cloud Island",
			SourceResource: "energy",
			AmountRange:    [2]float64{200, 800},
			RarityWeight:   6,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "weather_station",
					Level1Production: building.ProductionResult{Money: 10},
					Level2Production: building.ProductionResult{Money: 20},
					Level3Production: building.ProductionResult{Money: 30},
					SourceConsumption: 1.5,
				},
				{
					BuildingType: "sky_fish_farm",
					Level1Production: building.ProductionResult{Food: 5, Money: 3},
					Level2Production: building.ProductionResult{Food: 10, Money: 6},
					Level3Production: building.ProductionResult{Food: 15, Money: 9},
					SourceConsumption: 2.0,
				},
			},
		},
		{
			Type: "underground_lake", Name: "Underground Lake",
			SourceResource: "food",
			AmountRange:    [2]float64{200, 700},
			RarityWeight:   6,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "water_purifier",
					Level1Production: building.ProductionResult{Reagents: 4},
					Level2Production: building.ProductionResult{Reagents: 8},
					Level3Production: building.ProductionResult{Reagents: 12},
					SourceConsumption: 2.0,
				},
				{
					BuildingType: "rare_fish_habitat",
					Level1Production: building.ProductionResult{Food: 4, Reagents: 2},
					Level2Production: building.ProductionResult{Food: 8, Reagents: 4},
					Level3Production: building.ProductionResult{Food: 12, Reagents: 6},
					SourceConsumption: 2.5,
				},
			},
		},
		{
			Type: "radioactive_zone", Name: "Radioactive Zone",
			SourceResource: "reagents",
			AmountRange:    [2]float64{400, 1200},
			RarityWeight:   6,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "isotope_extractor",
					Level1Production: building.ProductionResult{AlienTech: 4},
					Level2Production: building.ProductionResult{AlienTech: 8},
					Level3Production: building.ProductionResult{AlienTech: 12},
					SourceConsumption: 2.0,
				},
				{
					BuildingType: "waste_processor",
					Level1Production: building.ProductionResult{AlienTech: 2, Money: 5},
					Level2Production: building.ProductionResult{AlienTech: 4, Money: 10},
					Level3Production: building.ProductionResult{AlienTech: 6, Money: 15},
					SourceConsumption: 2.5,
				},
			},
		},
		{
			Type: "anomaly_zone", Name: "Anomaly Zone",
			SourceResource: "reagents",
			AmountRange:    [2]float64{500, 1500},
			RarityWeight:   6,
			Buildings: []LocationBuildingDef{
				{
					BuildingType: "anomaly_analyzer",
					Level1Production: building.ProductionResult{AlienTech: 5, Energy: 15},
					Level2Production: building.ProductionResult{AlienTech: 10, Energy: 30},
					Level3Production: building.ProductionResult{AlienTech: 15, Energy: 45},
					SourceConsumption: 2.5,
				},
				{
					BuildingType: "quantum_generator",
					Level1Production: building.ProductionResult{Energy: 50, AlienTech: 3},
					Level2Production: building.ProductionResult{Energy: 100, AlienTech: 6},
					Level3Production: building.ProductionResult{Energy: 150, AlienTech: 9},
					SourceConsumption: 3.0,
				},
			},
		},
	}
}

type weightedLocation struct {
	locType LocationType
	weight  int
}

func GetLocationTypes() []LocationType {
	return locationTypes
}

func GetLocationRarityWeight(locType string) int {
	for _, lt := range locationTypes {
		if lt.Type == locType {
			return lt.RarityWeight
		}
	}
	return 30
}

func GetBuildingCostByRarity(rarityWeight int) (food, iron, money float64) {
	switch {
	case rarityWeight == 30:
		return 100, 50, 200
	case rarityWeight == 20:
		return 150, 75, 300
	case rarityWeight == 12:
		return 200, 100, 400
	case rarityWeight == 6:
		return 300, 150, 600
	default:
		return 100, 50, 200
	}
}

func getWeightedLocationTypes(planetType PlanetResourceType) []weightedLocation {
	result := make([]weightedLocation, 0, len(locationTypes))
	for _, lt := range locationTypes {
		weight := lt.RarityWeight
		if string(planetType) == lt.SourceResource {
			weight *= 6
		}
		result = append(result, weightedLocation{locType: lt, weight: weight})
	}
	return result
}

func SelectLocationType(planetType PlanetResourceType) string {
	weighted := getWeightedLocationTypes(planetType)

	totalWeight := 0
	for _, w := range weighted {
		totalWeight += w.weight
	}

	pick := rand.Intn(totalWeight)

	for _, w := range weighted {
		pick -= w.weight
		if pick < 0 {
			return w.locType.Type
		}
	}

	return weighted[0].locType.Type
}

func CalculateSourceAmount(locType string, planetType PlanetResourceType) float64 {
	var lt LocationType
	for _, t := range locationTypes {
		if t.Type == locType {
			lt = t
			break
		}
	}

	if lt.Type == "" {
		return 100
	}

	min := lt.AmountRange[0]
	max := lt.AmountRange[1]
	amount := min + rand.Float64()*(max-min)

	rarityMultiplier := 1.0
	switch {
	case lt.RarityWeight == 30:
		rarityMultiplier = 1.0
	case lt.RarityWeight == 20:
		rarityMultiplier = 1.2
	case lt.RarityWeight == 12:
		rarityMultiplier = 1.5
	case lt.RarityWeight == 6:
		rarityMultiplier = 2.0
	}
	amount *= rarityMultiplier

	if string(planetType) == lt.SourceResource {
		amount *= 1.5
	} else {
		amount *= 0.5
	}

	return math.Round(amount*100) / 100
}

func GenerateName(locType string) string {
	locNames := map[string][]string{
		"pond":            {"Silver Pond", "Crystal Pond", "Misty Pond", "Still Waters", "Mirror Pond"},
		"river":           {"Silver River", "Crystal River", "Swift River", "Winding River", "Ancient River"},
		"forest":          {"Oakwood Forest", "Misty Forest", "Ancient Forest", "Emerald Forest", "Whispering Forest"},
		"mineral_deposit": {"Iron Vein", "Crystal Deposit", "Stone Deposit", "Deep Deposit", "Rich Deposit"},
		"dry_valley":      {"Sun Valley", "Dust Valley", "Dry Basin", "Barren Valley", "Ash Valley"},
		"waterfall":       {"Thunder Falls", "Crystal Falls", "Mist Falls", "Silver Falls", "Ancient Falls"},
		"cave":            {"Deep Cave", "Crystal Cave", "Dark Hollow", "Stone Hollow", "Ancient Cave"},
		"thermal_spring":  {"Warm Spring", "Steam Spring", "Mist Spring", "Hot Spring", "Ancient Spring"},
		"salt_lake":       {"Salt Lake", "White Lake", "Crystal Lake", "Shallow Lake", "Deep Lake"},
		"wind_pass":       {"Wind Gap", "Storm Pass", "Breeze Pass", "High Pass", "Canyon Pass"},
		"crystal_cave":    {"Prism Cave", "Shimmer Cave", "Crystal Hollow", "Gem Cave", "Radiant Cave"},
		"meteor_crater":   {"Impact Crater", "Star Crater", "Scorched Crater", "Deep Crater", "Ancient Crater"},
		"sunken_city":     {"Lost City", "Beneath Waves", "Ancient Ruins", "Sunken Temple", "Deep Ruins"},
		"glacier":         {"Ice Glacier", "Frozen Expanse", "Crystal Glacier", "Northern Glacier", "Ancient Glacier"},
		"mushroom_forest": {"Spore Forest", "Glowing Forest", "Mushroom Glade", "Fungal Forest", "Deep Glade"},
		"crystal_field":   {"Crystal Field", "Radiant Field", "Shimmering Field", "Prism Field", "Luminous Field"},
		"cloud_island":    {"Cloud Isle", "Sky Island", "Floating Isle", "Mist Isle", "High Isle"},
		"underground_lake": {"Deep Lake", "Hidden Lake", "Subterranean Lake", "Dark Lake", "Crystal Lake"},
		"radioactive_zone": {"Toxic Zone", "Irradiated Zone", "Hazard Zone", "Green Zone", "Wasteland Zone"},
		"anomaly_zone":    {"Anomaly Site", "Void Zone", "Distortion Zone", "Strange Zone", "Mystic Zone"},
	}

	names := locNames[locType]
	if len(names) == 0 {
		return "Unknown Location"
	}

	return names[rand.Intn(len(names))]
}
