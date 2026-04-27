package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"spacegame/internal/game"
	"spacegame/internal/game/building"

	"github.com/go-chi/chi/v5"
)

func handleGetBuildings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		for i := range p.Buildings {
			p.PopulateBuildingEntry(i)
		}

		details := p.GetBuildDetails()

		buildings := make([]BuildingDetail, len(details.Buildings))
		for i, b := range details.Buildings {
			deltas := building.NextLevelDeltas(b.Type, b.Level)
			buildings[i] = BuildingDetail{
				Type:          b.Type,
				Level:         b.Level,
				BuildProgress: b.BuildProgress,
				Enabled:       b.Enabled,
				BuildTime:     b.BuildTime,
				Cost: CostDetail{
					Food:  b.Cost.Food,
					Iron:  b.Cost.Iron,
					Money: b.Cost.Money,
				},
				NextCost: CostDetail{
					Food:  b.NextCost.Food,
					Iron:  b.NextCost.Iron,
					Money: b.NextCost.Money,
				},
				Production: ProdDetail{
					Food:       b.Production.Food,
					Iron:       b.Production.Iron,
					Composite:  b.Production.Composite,
					Mechanisms: b.Production.Mechanisms,
					Reagents:   b.Production.Reagents,
					Energy:     b.Production.Energy,
					Money:      b.Production.Money,
					AlienTech:  b.Production.AlienTech,
				},
				NextProduction: ProdDetail{
					Food:       b.Production.Food + deltas.Food,
					Iron:       b.Production.Iron + deltas.Iron,
					Composite:  b.Production.Composite + deltas.Composite,
					Mechanisms: b.Production.Mechanisms + deltas.Mechanisms,
					Reagents:   b.Production.Reagents + deltas.Reagents,
					Energy:     b.Production.Energy + deltas.Energy,
					Money:      b.Production.Money + deltas.Money,
					AlienTech:  b.Production.AlienTech + deltas.AlienTech,
				},
				Deltas: ProdDetail{
					Food:       deltas.Food,
					Iron:       deltas.Iron,
					Composite:  deltas.Composite,
					Mechanisms: deltas.Mechanisms,
					Reagents:   deltas.Reagents,
					Energy:     deltas.Energy,
					Money:      deltas.Money,
					AlienTech:  deltas.AlienTech,
				},
			}
		}

		JSON(w, http.StatusOK, buildings)
	}
}

func handleBuildBuilding(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		ownerID := AuthPlayerFromContext(r).ID

		var req BuildBuildingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Type == "" {
			Error(w, http.StatusBadRequest, "Missing building type")
			return
		}

		validBuildings := map[string]bool{
			"farm": true, "solar": true, "storage": true, "base": true,
			"energy_storage": true, "shipyard": true,
			"market": true, "dynamo": true, "mine": true,
		}
		if !validBuildings[req.Type] {
			Error(w, http.StatusBadRequest, "Unknown building type")
			return
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		foodCost, moneyCost, err := p.AddBuilding(req.Type)
		if err != nil {
			resp := map[string]interface{}{
				"error":               err.Error(),
				"active_constructions": p.ActiveConstruction,
				"max_constructions":   p.GetMaxConcurrentBuildings(),
			}
			if pe, ok := err.(*game.PlanetError); ok && pe.Extra != "" {
				resp["extra"] = pe.Extra
			}
			JSON(w, http.StatusBadRequest, resp)
			return
		}

		idx := p.FindBuildingIndex(req.Type)
		if idx >= 0 {
			p.PopulateBuildingEntry(idx)
		}

		level := p.GetBuildingLevel(req.Type)
		wsBroadcast.BroadcastBuildingUpdate(ownerID, planetID, req.Type, level)

		Created(w, map[string]interface{}{
			"status":              "started",
			"type":                req.Type,
			"progress":            p.Buildings[idx].BuildProgress,
			"food_cost":           foodCost,
			"money_cost":          moneyCost,
			"active_constructions": p.ActiveConstruction,
			"max_constructions":   p.GetMaxConcurrentBuildings(),
		})
	}
}

func handleConfirmBuilding(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		ownerID := AuthPlayerFromContext(r).ID
		buildingType := chi.URLParam(r, "buildingType")

		if buildingType == "" {
			Error(w, http.StatusBadRequest, "Missing building type")
			return
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		if err := p.ConfirmBuilding(buildingType); err != nil {
			JSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		idx := p.FindBuildingIndex(buildingType)
		if idx >= 0 {
			p.PopulateBuildingEntry(idx)
		}

		level := p.GetBuildingLevel(buildingType)
		wsBroadcast.BroadcastBuildingUpdate(ownerID, planetID, buildingType, level)

		JSON(w, http.StatusOK, map[string]interface{}{
			"status": "confirmed",
			"type":   buildingType,
		})
	}
}

func handleToggleBuilding(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		ownerID := AuthPlayerFromContext(r).ID
		buildingType := chi.URLParam(r, "buildingType")

		if buildingType == "" {
			Error(w, http.StatusBadRequest, "Missing building type")
			return
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		idx := p.FindBuildingIndex(buildingType)
		if idx < 0 {
			Error(w, http.StatusNotFound, "Building not found")
			return
		}

		b := &p.Buildings[idx]
		if b.IsBuilding() || b.IsBuildComplete() {
			Error(w, http.StatusBadRequest, "Building not ready")
			return
		}

		if b.Type == "storage" {
			b.Enabled = true
		} else {
			b.Enabled = !b.Enabled
		}

		level := p.GetBuildingLevel(buildingType)
		wsBroadcast.BroadcastBuildingUpdate(ownerID, planetID, buildingType, level)

		JSON(w, http.StatusOK, map[string]interface{}{
			"status":  "toggled",
			"type":    buildingType,
			"enabled": b.Enabled,
		})
	}
}

func handleGetBuildDetails(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		for i := range p.Buildings {
			p.PopulateBuildingEntry(i)
		}

		details := p.GetBuildDetails()

		buildings := make([]BuildingDetail, len(details.Buildings))
		for i, b := range details.Buildings {
			deltas := building.NextLevelDeltas(b.Type, b.Level)
			buildings[i] = BuildingDetail{
				Type:          b.Type,
				Level:         b.Level,
				BuildProgress: b.BuildProgress,
				Enabled:       b.Enabled,
				BuildTime:     b.BuildTime,
				Cost: CostDetail{
					Food:  b.Cost.Food,
					Iron:  b.Cost.Iron,
					Money: b.Cost.Money,
				},
				NextCost: CostDetail{
					Food:  b.NextCost.Food,
					Iron:  b.NextCost.Iron,
					Money: b.NextCost.Money,
				},
				Production: ProdDetail{
					Food:       b.Production.Food,
					Iron:       b.Production.Iron,
					Composite:  b.Production.Composite,
					Mechanisms: b.Production.Mechanisms,
					Reagents:   b.Production.Reagents,
					Energy:     b.Production.Energy,
					Money:      b.Production.Money,
					AlienTech:  b.Production.AlienTech,
				},
				NextProduction: ProdDetail{
					Food:       b.Production.Food + deltas.Food,
					Iron:       b.Production.Iron + deltas.Iron,
					Composite:  b.Production.Composite + deltas.Composite,
					Mechanisms: b.Production.Mechanisms + deltas.Mechanisms,
					Reagents:   b.Production.Reagents + deltas.Reagents,
					Energy:     b.Production.Energy + deltas.Energy,
					Money:      b.Production.Money + deltas.Money,
					AlienTech:  b.Production.AlienTech + deltas.AlienTech,
				},
				Deltas: ProdDetail{
					Food:       deltas.Food,
					Iron:       deltas.Iron,
					Composite:  deltas.Composite,
					Mechanisms: deltas.Mechanisms,
					Reagents:   deltas.Reagents,
					Energy:     deltas.Energy,
					Money:      deltas.Money,
					AlienTech:  deltas.AlienTech,
				},
			}
		}

		production := ProdDetail{
			Food:       details.ResourceProduction.Food,
			Iron:       details.ResourceProduction.Iron,
			Composite:  details.ResourceProduction.Composite,
			Mechanisms: details.ResourceProduction.Mechanisms,
			Reagents:   details.ResourceProduction.Reagents,
			Energy:     details.ResourceProduction.Energy,
			Money:      details.ResourceProduction.Money,
			AlienTech:  details.ResourceProduction.AlienTech,
			EnergyNet:  details.EnergyBalance,
		}

		buildingCosts := make(map[string]BuildingCostDetail)
		existingTypes := make(map[string]bool)
		for _, b := range buildings {
			existingTypes[b.Type] = true
		}
		for _, bt := range game.BuildingsOrder {
			if !existingTypes[bt] && game.IsBuildingUnlocked(bt, p.Research.GetCompleted(), p.Resources.ResearchUnlocks) {
				cost := p.GetBuildingCost(bt, 0)
				p1 := building.Production(bt, 1)
				e1 := -building.EnergyConsumption(bt, 1)
				deltas := building.NextLevelDeltas(bt, 0)
				nextP := building.Production(bt, 2)
				nextE := -building.EnergyConsumption(bt, 2)
				buildingCosts[bt] = BuildingCostDetail{
					Cost: CostDetail{
						Food:  cost.Food,
						Iron:  cost.Iron,
						Money: cost.Money,
					},
					Production: ProdDetail{
						Food:       p1.Food,
						Iron:       p1.Iron,
						Composite:  p1.Composite,
						Mechanisms: p1.Mechanisms,
						Reagents:   p1.Reagents,
						Energy:     e1,
						Money:      p1.Money,
						AlienTech:  p1.AlienTech,
					},
					NextProduction: ProdDetail{
						Food:       nextP.Food,
						Iron:       nextP.Iron,
						Composite:  nextP.Composite,
						Mechanisms: nextP.Mechanisms,
						Reagents:   nextP.Reagents,
						Energy:     nextE,
						Money:      nextP.Money,
						AlienTech:  nextP.AlienTech,
					},
					Deltas: ProdDetail{
						Food:       deltas.Food,
						Iron:       deltas.Iron,
						Composite:  deltas.Composite,
						Mechanisms: deltas.Mechanisms,
						Reagents:   deltas.Reagents,
						Energy:     deltas.Energy,
						Money:      deltas.Money,
						AlienTech:  deltas.AlienTech,
					},
				}
			}
		}

		var gardenBedStateResp GardenBedStateResponse
		if p.GardenBedState != nil && p.GardenBedState.RowCount > 0 {
			gardenBedStateResp = GardenBedStateResponse{
				Rows:     p.GardenBedState.Rows,
				RowCount: p.GardenBedState.RowCount,
			}
		} else {
			farmLevel := p.GetBuildingLevel("farm")
			if farmLevel > 0 {
				p.GardenBedState = game.NewGardenBedState(farmLevel)
				gardenBedStateResp = GardenBedStateResponse{
					Rows:     p.GardenBedState.Rows,
					RowCount: p.GardenBedState.RowCount,
				}
			}
		}

		resp := BuildDetailsResponse{
			Resources: PlanetResources{
				Food:            details.Resources.Food,
				Iron:            details.Resources.Iron,
				Composite:       details.Resources.Composite,
				Mechanisms:      details.Resources.Mechanisms,
				Reagents:        details.Resources.Reagents,
				Energy:          details.EnergyBuffer.Value,
				MaxEnergy:       details.Resources.MaxEnergy,
				Money:           details.Resources.Money,
				AlienTech:       details.Resources.AlienTech,
				StorageCapacity: p.CalculateStorageCapacity(),
			},
			EnergyBuffer: EnergyBufferDetail{
				Value:   details.EnergyBuffer.Value,
				Max:     details.EnergyBuffer.Max,
				Deficit: details.EnergyBuffer.Deficit,
			},
			Buildings: buildings,
			EnergyBalance: EnergyBalanceDetail{
				Production:  details.ResourceProduction.Energy,
				Consumption: 0,
				Net:         details.EnergyBalance,
			},
			ResourceProduction: production,
			ActiveConstruction: details.ActiveConstruction,
			MaxConstruction:    details.MaxConstruction,
			BaseOperational:    details.BaseOperational,
			CanResearch:        details.CanResearch,
			CanExpedition:      details.CanExpedition,
			BuildingCosts:      buildingCosts,
			ResearchUnlocks:    p.Resources.ResearchUnlocks,
			GardenBedState:       gardenBedStateResp,
		}

		JSON(w, http.StatusOK, resp)
	}
}
