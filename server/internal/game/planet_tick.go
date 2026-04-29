package game

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"spacegame/internal/game/planet_survey"
	"spacegame/internal/game/ship"
)

// Tick processes one game tick (1 second).
func (p *Planet) Tick(gameTick int64) {
	p.LastTick = time.Now()

	// 0. Recalculate build speed from research
	p.RecalculateBuildSpeed()

	// 1. Step building construction progress
	for i := range p.Buildings {
		p.stepBuildingEntry(i)
	}
	p.RecalculateActiveConstruction()
	// Disable buildings that are under construction or ready for confirmation
	for i := range p.Buildings {
		b := &p.Buildings[i]
		if (b.IsBuilding() || b.IsBuildComplete()) && b.Enabled {
			b.Enabled = false
		}
	}

	// 2. Energy tick
	p.tickEnergy()

	// 3. Resource production and clamping
	p.tickResources()
	// 3.5 Auto-disable dynamo when food is depleted
	p.autoDisableDynamo()
	// 3.6 Auto-disable base when food is depleted
	p.autoDisableBase()

	// 3.7 Garden bed tick (every 10 ticks, only if farm is working)
	if p.GardenBedState != nil && p.GetBuildingLevel("farm") > 0 {
		farmIdx := p.FindBuildingIndex("farm")
		if farmIdx >= 0 && farmIdx < len(p.Buildings) && p.Buildings[farmIdx].IsWorking() {
			ProcessGardenBedTick(p, gameTick)
		}
	}

	// 4. Advance research progress (pause if base is not operational)
	if p.HasOperationalBase() {
		p.Research.Tick()
		p.applyResearchEffects()
	}

	// 5. Advance ship construction
	if completed := p.Shipyard.Tick(); completed != nil {
		st := ship.GetShipType(*completed)
		if st != nil {
			p.Fleet.AddShip(st, 1)
		}
	}

	// 6. Advance expeditions
	p.TickExpeditions()

	// 6.5 Tick surface expeditions
	p.TickSurfaceExpeditions()

	// 6.6 Tick location buildings
	p.TickLocationBuildings()

	// 7. Save to DB (throttled)
	if p.game != nil && p.game.shouldSave(p.ID) {
		p.game.savePlanet(p)
	}

	// 8. Broadcast state update
	p.broadcastPlanetUpdate()
}

func (p *Planet) TickSurfaceExpeditions() {
	if !p.BaseOperational() {
		return
	}

	for i := len(p.SurfaceExpeditions) - 1; i >= 0; i-- {
		exp := p.SurfaceExpeditions[i]

		if exp.Status == "completed" || exp.Status == "failed" || exp.Status == "abandoned" {
			continue
		}

		if exp.Status == "discovered" {
			continue
		}

	

		planet_survey.Tick(exp, 1)

		if int(exp.ElapsedTime)%60 == 0 && exp.Discovered == nil {
			count := len(p.Locations)
			chance := planet_survey.CalculateDiscoveryChance(count)
			if rand.Float64() < chance {
				locType := planet_survey.SelectLocationType(planet_survey.PlanetResourceType(string(p.ResourceType)))
				sourceAmount := planet_survey.CalculateSourceAmount(locType, planet_survey.PlanetResourceType(string(p.ResourceType)))
				var sourceResource string
				for _, lt := range planet_survey.GetLocationTypes() {
					if lt.Type == locType {
						sourceResource = lt.SourceResource
						break
					}
				}
				loc := &planet_survey.Location{
					ID:              p.ID + "_loc_" + time.Now().Format("20060102150405"),
					PlanetID:        p.ID,
					OwnerID:         p.OwnerID,
					Type:            locType,
					Name:            planet_survey.GenerateName(locType),
					SourceResource:  sourceResource,
					SourceAmount:    sourceAmount,
					SourceRemaining: sourceAmount,
					Active:          true,
					BuildingActive:  false,
					DiscoveredAt:    time.Now(),
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
					Buildings:       make([]*planet_survey.LocationBuilding, 0),
				}
				exp.Discovered = loc
				exp.Status = "discovered"
				p.Locations = append(p.Locations, loc)
			}
		}

		if planet_survey.IsExpired(exp) {
			rangeStats := p.RangeStats[exp.Range]
			if rangeStats == nil {
				rangeStats = &planet_survey.ExpeditionRangeStats{}
				p.RangeStats[exp.Range] = rangeStats
			}
			rangeStats.TotalExpeditions++

			isSuccess := exp.Discovered != nil
			if isSuccess {
				rangeStats.LocationsFound++
			}

			resourceRecovery := planet_survey.CalculateResourceRecovery(p.Level, rangeStats.TotalExpeditions, isSuccess)

			for resName, amount := range resourceRecovery {
				if amount <= 0 {
					continue
				}
				switch resName {
				case "food":
					p.Resources.Food += amount
				case "iron":
					p.Resources.Iron += amount
				case "money":
					p.Resources.Money += amount
				case "reagents":
					p.Resources.Reagents += amount
				case "composite":
					p.Resources.Composite += amount
				case "mechanisms":
					p.Resources.Mechanisms += amount
				}
			}

			historyEntry := planet_survey.ExpeditionHistoryEntry{
				ID:             exp.ID,
				PlanetID:       p.ID,
				ExpeditionType: "surface",
				Status:         exp.Status,
				Result:         "success",
				Discovered:     "",
				ResourcesGained: resourceRecovery,
				CreatedAt:      exp.CreatedAt,
				CompletedAt:    time.Now(),
			}

			if exp.Discovered != nil {
				historyEntry.Result = "success"
				historyEntry.Discovered = exp.Discovered.Name
			} else {
				historyEntry.Result = "failed"
			}

			p.ExpeditionHistory = append(p.ExpeditionHistory, historyEntry)

			var notifyMsg string
			if isSuccess {
				var resParts []string
				for res, amt := range resourceRecovery {
					if amt > 0 {
						resParts = append(resParts, fmt.Sprintf("%s: +%.0f", res, amt))
					}
				}
				notifyMsg = fmt.Sprintf("Экспедиция завершена! Обнаружена локация: %s. Получено: %s", exp.Discovered.Name, strings.Join(resParts, ", "))
			} else {
				notifyMsg = "Экспедиция завершена без результатов."
			}
			if p.game != nil && p.game.notifyFunc != nil {
				p.game.notifyFunc(p.OwnerID, notifyMsg, "expedition_complete")
			}

			p.SurfaceExpeditions = append(p.SurfaceExpeditions[:i], p.SurfaceExpeditions[i+1:]...)
		}
	}
}

func (p *Planet) TickLocationBuildings() {
	resources := map[string]float64{
		"food":       p.Resources.Food,
		"iron":       p.Resources.Iron,
		"composite":  p.Resources.Composite,
		"mechanisms": p.Resources.Mechanisms,
		"reagents":   p.Resources.Reagents,
		"energy":     p.Resources.Energy,
		"money":      p.Resources.Money,
		"alien_tech": p.Resources.AlienTech,
	}

	for _, loc := range p.Locations {
		if !loc.Active || !loc.BuildingActive {
			continue
		}

		locBuildings := make([]*planet_survey.LocationBuilding, 0)
		for _, lb := range loc.Buildings {
			locBuildings = append(locBuildings, lb)
		}

		planet_survey.TickLocationBuildings(loc, locBuildings, resources)
	}

	p.Resources.Food = resources["food"]
	p.Resources.Iron = resources["iron"]
	p.Resources.Composite = resources["composite"]
	p.Resources.Mechanisms = resources["mechanisms"]
	p.Resources.Reagents = resources["reagents"]
	p.Resources.Energy = resources["energy"]
	p.Resources.Money = resources["money"]
	p.Resources.AlienTech = resources["alien_tech"]
}
