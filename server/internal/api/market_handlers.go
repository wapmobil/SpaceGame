package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"spacegame/internal/game"

	"github.com/go-chi/chi/v5"
)

func getMarketplace() *game.Marketplace {
	g := game.Instance()
	if g == nil {
		return nil
	}
	return g.Marketplace
}

func handleCreateMarketOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		playerID := AuthPlayerFromContext(r).ID

		var req CreateMarketOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Resource == "" {
			Error(w, http.StatusBadRequest, "Missing resource")
			return
		}
		if req.OrderType == "" {
			Error(w, http.StatusBadRequest, "Missing order_type")
			return
		}
		if req.Amount <= 0 {
			Error(w, http.StatusBadRequest, "Amount must be positive")
			return
		}
		if req.Price <= 0 {
			Error(w, http.StatusBadRequest, "Price must be positive")
			return
		}

		if req.OrderType != "buy" && req.OrderType != "sell" {
			Error(w, http.StatusBadRequest, "Invalid order_type: must be 'buy' or 'sell'")
			return
		}

		validResources := map[string]bool{
			"food": true, "composite": true, "mechanisms": true, "reagents": true,
		}
		if !validResources[req.Resource] {
			Error(w, http.StatusBadRequest, "Invalid resource")
			return
		}

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		if p.EnergyBuffer.Value < game.OrderCreationCost {
			Error(w, http.StatusConflict, fmt.Sprintf("Insufficient energy. Need %.0f energy, have %.0f", game.OrderCreationCost, p.EnergyBuffer.Value))
			return
		}

		orderType := game.OrderType(req.OrderType)
		if orderType == game.OrderSell {
			switch req.Resource {
			case "food":
				if p.Resources.Food < req.Amount {
					Error(w, http.StatusConflict, fmt.Sprintf("Insufficient %s. Need %.0f, have %.0f", req.Resource, req.Amount, p.Resources.Food))
					return
				}
			case "composite":
				if p.Resources.Composite < req.Amount {
					Error(w, http.StatusConflict, fmt.Sprintf("Insufficient %s. Need %.0f, have %.0f", req.Resource, req.Amount, p.Resources.Composite))
					return
				}
			case "mechanisms":
				if p.Resources.Mechanisms < req.Amount {
					Error(w, http.StatusConflict, fmt.Sprintf("Insufficient %s. Need %.0f, have %.0f", req.Resource, req.Amount, p.Resources.Mechanisms))
					return
				}
			case "reagents":
				if p.Resources.Reagents < req.Amount {
					Error(w, http.StatusConflict, fmt.Sprintf("Insufficient %s. Need %.0f, have %.0f", req.Resource, req.Amount, p.Resources.Reagents))
					return
				}
			}
		}

		mp := getMarketplace()
		if mp == nil {
			Error(w, http.StatusInternalServerError, "Marketplace not initialized")
			return
		}

		order, err := mp.CreateOrder(planetID, playerID, req.Resource, orderType, req.Amount, req.Price, req.IsPrivate)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "exceeds maximum"):
				Error(w, http.StatusBadRequest, errMsg)
			case strings.Contains(errMsg, "price must be"):
				Error(w, http.StatusBadRequest, errMsg)
			case strings.Contains(errMsg, "invalid resource"):
				Error(w, http.StatusBadRequest, errMsg)
			case strings.Contains(errMsg, "invalid order type"):
				Error(w, http.StatusBadRequest, errMsg)
			default:
				Error(w, http.StatusInternalServerError, "Failed to create order")
			}
			return
		}

		p.EnergyBuffer.Value -= game.OrderCreationCost
		if p.EnergyBuffer.Value < 0 {
			p.EnergyBuffer.Value = 0
		}

		if orderType == game.OrderSell {
			switch req.Resource {
			case "food":
				p.Resources.Food -= req.Amount
			case "composite":
				p.Resources.Composite -= req.Amount
			case "mechanisms":
				p.Resources.Mechanisms -= req.Amount
			case "reagents":
				p.Resources.Reagents -= req.Amount
			}
		}

		if orderType == game.OrderBuy {
			switch req.Resource {
			case "food":
				p.Resources.Food -= req.Amount * req.Price
			case "composite":
				p.Resources.Composite -= req.Amount * req.Price
			case "mechanisms":
				p.Resources.Mechanisms -= req.Amount * req.Price
			case "reagents":
				p.Resources.Reagents -= req.Amount * req.Price
			}
		}

		resp := MarketOrderResponse{
			ID:               order.ID,
			PlanetID:         order.PlanetID,
			PlayerID:         order.PlayerID,
			Resource:         order.Resource,
			OrderType:        string(order.OrderType),
			Amount:           order.Amount,
			Price:            order.Price,
			IsPrivate:        order.IsPrivate,
			Link:             order.Link,
			Status:           string(order.Status),
			CreatedAt:        order.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        order.UpdatedAt.Format(time.RFC3339),
			ReservedResources: order.ReservedResources,
		}

		Created(w, resp)
	}
}

func handleGetMyOrders(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID := AuthPlayerFromContext(r).ID

		mp := getMarketplace()
		if mp == nil {
			Error(w, http.StatusInternalServerError, "Marketplace not initialized")
			return
		}

		orders := mp.GetMyOrders(playerID)

		resp := make([]MarketOrderResponse, 0, len(orders))
		for _, order := range orders {
			resp = append(resp, MarketOrderResponse{
				ID:               order.ID,
				PlanetID:         order.PlanetID,
				PlayerID:         order.PlayerID,
				Resource:         order.Resource,
				OrderType:        string(order.OrderType),
				Amount:           order.Amount,
				Price:            order.Price,
				IsPrivate:        order.IsPrivate,
				Link:             order.Link,
				Status:           string(order.Status),
				CreatedAt:        order.CreatedAt.Format(time.RFC3339),
				UpdatedAt:        order.UpdatedAt.Format(time.RFC3339),
				ReservedResources: order.ReservedResources,
			})
		}

		JSON(w, http.StatusOK, resp)
	}
}

func handleGetGlobalMarket(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mp := getMarketplace()
		if mp == nil {
			Error(w, http.StatusInternalServerError, "Marketplace not initialized")
			return
		}

		orders := mp.GetVisibleOrders("")

		buyOrders := make(map[string][]MarketOrderResponse)
		sellOrders := make(map[string][]MarketOrderResponse)

		for _, order := range orders {
			resp := MarketOrderResponse{
				ID:               order.ID,
				PlanetID:         order.PlanetID,
				PlayerID:         order.PlayerID,
				Resource:         order.Resource,
				OrderType:        string(order.OrderType),
				Amount:           order.Amount,
				Price:            order.Price,
				IsPrivate:        order.IsPrivate,
				Link:             order.Link,
				Status:           string(order.Status),
				CreatedAt:        order.CreatedAt.Format(time.RFC3339),
				UpdatedAt:        order.UpdatedAt.Format(time.RFC3339),
				ReservedResources: order.ReservedResources,
			}
			if order.OrderType == game.OrderBuy {
				buyOrders[order.Resource] = append(buyOrders[order.Resource], resp)
			} else {
				sellOrders[order.Resource] = append(sellOrders[order.Resource], resp)
			}
		}

		bestBuyPrice := 0.0
		bestSellPrice := math.MaxFloat64

		for _, orders := range buyOrders {
			for _, order := range orders {
				if order.Price > bestBuyPrice {
					bestBuyPrice = order.Price
				}
			}
		}
		for _, orders := range sellOrders {
			for _, order := range orders {
				if order.Price < bestSellPrice {
					bestSellPrice = order.Price
				}
			}
		}
		if bestSellPrice == math.MaxFloat64 {
			bestSellPrice = 0
		}

		totalVolume := 0.0
		for _, orders := range buyOrders {
			for _, order := range orders {
				totalVolume += order.Amount * order.Price
			}
		}
		for _, orders := range sellOrders {
			for _, order := range orders {
				totalVolume += order.Amount * order.Price
			}
		}

		npcTraderCount := len(mp.GetAllNPCTraders())

		JSON(w, http.StatusOK, map[string]interface{}{
			"buy_orders":       buyOrders,
			"sell_orders":      sellOrders,
			"best_buy_price":   bestBuyPrice,
			"best_sell_price":  bestSellPrice,
			"total_volume":     totalVolume,
			"active_orders":    len(orders),
			"npc_trader_count": npcTraderCount,
		})
	}
}

func handleDeleteMarketOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID := AuthPlayerFromContext(r).ID
		orderID := chi.URLParam(r, "id")

		mp := getMarketplace()
		if mp == nil {
			Error(w, http.StatusInternalServerError, "Marketplace not initialized")
			return
		}

		order := mp.GetOrder(orderID)
		if order == nil {
			Error(w, http.StatusNotFound, "Order not found")
			return
		}

		if order.PlayerID != playerID {
			Error(w, http.StatusForbidden, "Forbidden")
			return
		}

		p := game.Instance().GetPlanet(order.PlanetID)
		if p == nil {
			var planetID string
			err := db.QueryRow("SELECT id FROM planets WHERE player_id = $1 LIMIT 1", playerID).Scan(&planetID)
			if err == nil {
				p = game.Instance().GetPlanet(planetID)
				if p == nil {
					p = game.NewPlanet(planetID, playerID, "", game.Instance())
				}
			}
		}

		if p != nil {
			for resource, amount := range order.ReservedResources {
				switch resource {
				case "food":
					p.Resources.Food += amount
				case "composite":
					p.Resources.Composite += amount
				case "mechanisms":
					p.Resources.Mechanisms += amount
				case "reagents":
					p.Resources.Reagents += amount
				}
			}

			p.EnergyBuffer.Value += game.OrderCreationCost
		}

		err := mp.DeleteOrder(orderID)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "not found"):
				Error(w, http.StatusNotFound, "Order not found")
			case strings.Contains(errMsg, "not active"):
				Error(w, http.StatusConflict, "Order is not active and cannot be deleted")
			default:
				Error(w, http.StatusInternalServerError, "Failed to delete order")
			}
			return
		}

		JSON(w, http.StatusOK, map[string]interface{}{
			"status":          "deleted",
			"order_id":        orderID,
			"refunded_energy": game.OrderCreationCost,
		})
	}
}

func handleSellFood(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planetID := PlanetIDFromContext(r)
		_ = AuthPlayerFromContext(r)

		p := ensurePlanetLoaded(planetID)
		if p == nil {
			Error(w, http.StatusInternalServerError, "Failed to load planet")
			return
		}

		mp := getMarketplace()
		if mp == nil {
			Error(w, http.StatusInternalServerError, "Marketplace not initialized")
			return
		}

		bestSellPrice := 0.0
		for _, order := range mp.GetVisibleOrders("") {
			if order.Resource == "food" && order.OrderType == game.OrderBuy && order.Status == "active" {
				if order.Price > bestSellPrice {
					bestSellPrice = order.Price
				}
			}
		}

		if bestSellPrice <= 0 {
			bestSellPrice = 0.01
		}

		var req struct {
			Amount float64 `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
			Error(w, http.StatusBadRequest, "Invalid amount")
			return
		}

		sellAmount := math.Min(req.Amount, p.Resources.Food)
		if sellAmount <= 0 {
			Error(w, http.StatusBadRequest, "No food to sell")
			return
		}

		revenue := sellAmount * bestSellPrice
		p.Resources.Food -= sellAmount
		p.Resources.Money += revenue

		JSON(w, http.StatusOK, map[string]interface{}{
			"status":     "sold",
			"amount":     sellAmount,
			"price":      bestSellPrice,
			"revenue":    revenue,
			"food_left":  p.Resources.Food,
			"money":      p.Resources.Money,
		})
	}
}
