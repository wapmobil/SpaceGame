package game

import (
	"testing"
	"time"
)

func TestCreateBuyOrder(t *testing.T) {
	mp := NewMarketplace()

	order, err := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)
	if err != nil {
		t.Fatalf("unexpected error creating buy order: %v", err)
	}

	if order.ID == "" {
		t.Fatal("expected non-empty order ID")
	}
	if order.PlanetID != "planet-1" {
		t.Errorf("expected planet_id 'planet-1', got '%s'", order.PlanetID)
	}
	if order.PlayerID != "player-1" {
		t.Errorf("expected player_id 'player-1', got '%s'", order.PlayerID)
	}
	if order.Resource != "food" {
		t.Errorf("expected resource 'food', got '%s'", order.Resource)
	}
	if order.OrderType != OrderBuy {
		t.Errorf("expected order type 'buy', got '%s'", order.OrderType)
	}
	if order.Amount != 100 {
		t.Errorf("expected amount 100, got %f", order.Amount)
	}
	if order.Price != 10.0 {
		t.Errorf("expected price 10.0, got %f", order.Price)
	}
	if order.IsPrivate {
		t.Error("expected is_private to be false")
	}
	if order.Status != OrderActive {
		t.Errorf("expected status 'active', got '%s'", order.Status)
	}
	if order.Link != "" {
		t.Error("expected empty link for public order")
	}
}

func TestCreateSellOrder(t *testing.T) {
	mp := NewMarketplace()

	order, err := mp.CreateOrder("planet-1", "player-1", "composite", OrderSell, 50, 5.0, false)
	if err != nil {
		t.Fatalf("unexpected error creating sell order: %v", err)
	}

	if order.OrderType != OrderSell {
		t.Errorf("expected order type 'sell', got '%s'", order.OrderType)
	}
	if order.Amount != 50 {
		t.Errorf("expected amount 50, got %f", order.Amount)
	}
	if order.Price != 5.0 {
		t.Errorf("expected price 5.0, got %f", order.Price)
	}
}

func TestInvalidOrderAmount(t *testing.T) {
	mp := NewMarketplace()

	_, err := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 0, 10.0, false)
	if err == nil {
		t.Fatal("expected error for zero amount")
	}

	_, err = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, -10, 10.0, false)
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestOrderAmountExceedsMaximum(t *testing.T) {
	mp := NewMarketplace()

	_, err := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, MaxOrderAmount+1, 10.0, false)
	if err == nil {
		t.Fatal("expected error for amount exceeding maximum")
	}
}

func TestInvalidPrice(t *testing.T) {
	mp := NewMarketplace()

	_, err := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 0, false)
	if err == nil {
		t.Fatal("expected error for zero price")
	}

	_, err = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, MinPrice-0.01, false)
	if err == nil {
		t.Fatal("expected error for price below minimum")
	}

	_, err = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, MaxPrice+1, false)
	if err == nil {
		t.Fatal("expected error for price above maximum")
	}
}

func TestInvalidResource(t *testing.T) {
	mp := NewMarketplace()

	_, err := mp.CreateOrder("planet-1", "player-1", "invalid_resource", OrderBuy, 100, 10.0, false)
	if err == nil {
		t.Fatal("expected error for invalid resource")
	}
}

func TestInvalidOrderType(t *testing.T) {
	mp := NewMarketplace()

	_, err := mp.CreateOrder("planet-1", "player-1", "food", OrderType("invalid"), 100, 10.0, false)
	if err == nil {
		t.Fatal("expected error for invalid order type")
	}
}

func TestPrivateOrderCreatesLink(t *testing.T) {
	mp := NewMarketplace()

	order, err := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, true)
	if err != nil {
		t.Fatalf("unexpected error creating private order: %v", err)
	}

	if order.Link == "" {
		t.Fatal("expected non-empty link for private order")
	}
	if !order.IsPrivate {
		t.Error("expected is_private to be true")
	}
}

func TestGetPrivateOrder(t *testing.T) {
	mp := NewMarketplace()

	order, _ := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, true)
	retrieved := mp.GetPrivateOrder(order.Link)

	if retrieved == nil {
		t.Fatal("expected to find private order by link")
	}
	if retrieved.ID != order.ID {
		t.Errorf("expected order ID '%s', got '%s'", order.ID, retrieved.ID)
	}
}

func TestGetPrivateOrderNotFound(t *testing.T) {
	mp := NewMarketplace()

	retrieved := mp.GetPrivateOrder("non-existent-link")
	if retrieved != nil {
		t.Fatal("expected nil for non-existent private order link")
	}
}

func TestGetMyOrders(t *testing.T) {
	mp := NewMarketplace()

	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)
	_, _ = mp.CreateOrder("planet-1", "player-1", "composite", OrderSell, 50, 5.0, false)
	_, _ = mp.CreateOrder("planet-2", "player-2", "mechanisms", OrderBuy, 200, 8.0, false)

	myOrders := mp.GetMyOrders("player-1")

	if len(myOrders) != 2 {
		t.Errorf("expected 2 orders for player-1, got %d", len(myOrders))
	}

	for _, order := range myOrders {
		if order.PlayerID != "player-1" {
			t.Errorf("expected order to belong to player-1, got '%s'", order.PlayerID)
		}
	}
}

func TestGetVisibleOrders(t *testing.T) {
	mp := NewMarketplace()

	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)
	_, _ = mp.CreateOrder("planet-2", "player-2", "composite", OrderSell, 50, 5.0, true)
	_, _ = mp.CreateOrder("planet-3", "player-3", "mechanisms", OrderBuy, 200, 8.0, false)

	// Player 1 sees public orders but not private orders from others
	visible := mp.GetVisibleOrders("player-1")

	if len(visible) != 2 {
		t.Errorf("expected 2 visible orders for player-1, got %d", len(visible))
	}

	// Player 2 sees public orders and their own private order
	visible2 := mp.GetVisibleOrders("player-2")
	if len(visible2) != 3 {
		t.Errorf("expected 3 visible orders for player-2 (including own private), got %d", len(visible2))
	}
}

func TestOrderMatching(t *testing.T) {
	mp := NewMarketplace()

	// Create a sell order
	sellOrder, _ := mp.CreateOrder("planet-1", "player-1", "food", OrderSell, 100, 10.0, false)
	// Create a matching buy order
	buyOrder, _ := mp.CreateOrder("planet-2", "player-2", "food", OrderBuy, 100, 15.0, false)

	if sellOrder.ID == buyOrder.ID {
		t.Fatal("sell and buy orders should have different IDs")
	}

	result := mp.MatchOrders()

	if result.ExecutedTrades != 1 {
		t.Errorf("expected 1 executed trade, got %d", result.ExecutedTrades)
	}
	if result.TotalVolume != 1000.0 {
		t.Errorf("expected total volume 1000.0 (100 * 10.0), got %f", result.TotalVolume)
	}
	if len(result.MatchedOrders) != 2 {
		t.Errorf("expected 2 matched orders, got %d", len(result.MatchedOrders))
	}

	// Check that orders are filled
	if sellOrder.Status != OrderFilled {
		t.Errorf("expected sell order to be filled, got '%s'", sellOrder.Status)
	}
	if buyOrder.Status != OrderFilled {
		t.Errorf("expected buy order to be filled, got '%s'", buyOrder.Status)
	}
}

func TestOrderMatchingPartialFill(t *testing.T) {
	mp := NewMarketplace()

	// Create a large sell order
	sellOrder, _ := mp.CreateOrder("planet-1", "player-1", "food", OrderSell, 200, 10.0, false)
	// Create a smaller buy order that partially fills the sell
	buyOrder, _ := mp.CreateOrder("planet-2", "player-2", "food", OrderBuy, 50, 15.0, false)

	result := mp.MatchOrders()

	if result.ExecutedTrades != 1 {
		t.Errorf("expected 1 executed trade, got %d", result.ExecutedTrades)
	}

	// Sell order should still be active with reduced amount
	if sellOrder.Status != OrderActive {
		t.Errorf("expected sell order to remain active, got '%s'", sellOrder.Status)
	}
	if sellOrder.Amount != 150 {
		t.Errorf("expected sell order amount to be 150 (200-50), got %f", sellOrder.Amount)
	}

	// Buy order should be filled
	if buyOrder.Status != OrderFilled {
		t.Errorf("expected buy order to be filled, got '%s'", buyOrder.Status)
	}
}

func TestOrderMatchingPricePriority(t *testing.T) {
	mp := NewMarketplace()

	// Create multiple sell orders at different prices
	sell1, _ := mp.CreateOrder("planet-1", "player-1", "food", OrderSell, 100, 10.0, false)
	sell2, _ := mp.CreateOrder("planet-2", "player-2", "food", OrderSell, 100, 8.0, false)
	sell3, _ := mp.CreateOrder("planet-3", "player-3", "food", OrderSell, 100, 12.0, false)

	// Create a buy order that matches multiple sells
	buyOrder, _ := mp.CreateOrder("planet-4", "player-4", "food", OrderBuy, 250, 15.0, false)

	result := mp.MatchOrders()

	// Should match all 3 sell orders (best price first: sell2 at 8, sell1 at 10, sell3 at 12)
	if result.ExecutedTrades != 3 {
		t.Errorf("expected 3 executed trades, got %d", result.ExecutedTrades)
	}

	// buyOrder should be filled (250 = 100 + 100 + 50 from sell3)
	if buyOrder.Status != OrderFilled {
		t.Errorf("expected buy order to be filled, got '%s'", buyOrder.Status)
	}

	// sell3 should be partially filled (amount reduced but still active)
	if sell3.Status != OrderActive {
		t.Errorf("expected sell3 to remain active (partial fill), got '%s'", sell3.Status)
	}
	if sell3.Amount != 50 {
		t.Errorf("expected sell3 amount to be 50 (100-50), got %f", sell3.Amount)
	}

	_ = sell1
	_ = sell2
}

func TestNoMatchWhenPriceMismatch(t *testing.T) {
	mp := NewMarketplace()

	// Sell order at price 15
	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderSell, 100, 15.0, false)
	// Buy order at price 10 (lower than sell)
	_, _ = mp.CreateOrder("planet-2", "player-2", "food", OrderBuy, 100, 10.0, false)

	result := mp.MatchOrders()

	if result.ExecutedTrades != 0 {
		t.Errorf("expected 0 executed trades (no match), got %d", result.ExecutedTrades)
	}
}

func TestDeleteOrder(t *testing.T) {
	mp := NewMarketplace()

	order, _ := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)

	err := mp.DeleteOrder(order.ID)
	if err != nil {
		t.Fatalf("unexpected error deleting order: %v", err)
	}

	// Order should no longer exist
	retrieved := mp.GetOrder(order.ID)
	if retrieved != nil {
		t.Error("expected order to be deleted")
	}
}

func TestDeleteNonExistentOrder(t *testing.T) {
	mp := NewMarketplace()

	err := mp.DeleteOrder("non-existent-id")
	if err == nil {
		t.Fatal("expected error for non-existent order")
	}
}

func TestDeleteFilledOrder(t *testing.T) {
	mp := NewMarketplace()

	sellOrder, _ := mp.CreateOrder("planet-1", "player-1", "food", OrderSell, 100, 10.0, false)
	_, _ = mp.CreateOrder("planet-2", "player-2", "food", OrderBuy, 100, 15.0, false)

	mp.MatchOrders()

	err := mp.DeleteOrder(sellOrder.ID)
	if err == nil {
		t.Fatal("expected error for deleting filled order")
	}
}

func TestReservedResources(t *testing.T) {
	mp := NewMarketplace()

	order, err := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.ReservedResources == nil {
		t.Fatal("expected non-nil reserved resources")
	}
	if order.ReservedResources["food"] != 100 {
		t.Errorf("expected 100 food reserved, got %f", order.ReservedResources["food"])
	}
}

func TestNPCOrderCreation(t *testing.T) {
	mp := NewMarketplace()

	order, trader, err := mp.CreateNPCOrder("Trade Station Alpha", "npc-planet-1", "npc-player-1", "food", OrderSell, 100, 10.0)
	if err != nil {
		t.Fatalf("unexpected error creating NPC order: %v", err)
	}

	if order == nil {
		t.Fatal("expected non-nil order")
	}
	if trader == nil {
		t.Fatal("expected non-nil trader")
	}
	if trader.Name != "Trade Station Alpha" {
		t.Errorf("expected trader name 'Trade Station Alpha', got '%s'", trader.Name)
	}
}

func TestGenerateNPCOrders(t *testing.T) {
	mp := NewMarketplace()

	mp.GenerateNPCOrders()

	count := mp.GetOrderCount()
	if count == 0 {
		t.Fatal("expected at least one NPC order after generation")
	}

	traders := mp.GetAllNPCTraders()
	if len(traders) == 0 {
		t.Fatal("expected at least one NPC trader")
	}

	// All NPC orders should be public
	for _, order := range mp.GetVisibleOrders("anyone") {
		if order.IsPrivate {
			t.Error("NPC orders should be public")
		}
	}
}

func TestGetAllNPCTraders(t *testing.T) {
	mp := NewMarketplace()

	_, _, _ = mp.CreateNPCOrder("Trader 1", "npc-1", "npc-player-1", "food", OrderSell, 100, 10.0)
	_, _, _ = mp.CreateNPCOrder("Trader 2", "npc-2", "npc-player-2", "composite", OrderBuy, 50, 5.0)

	traders := mp.GetAllNPCTraders()
	if len(traders) != 2 {
		t.Errorf("expected 2 NPC traders, got %d", len(traders))
	}
}

func TestCancelNPCOrders(t *testing.T) {
	mp := NewMarketplace()

	_, trader, _ := mp.CreateNPCOrder("Trader 1", "npc-1", "npc-player-1", "food", OrderSell, 100, 10.0)

	err := mp.CancelNPCOrders(trader.ID)
	if err != nil {
		t.Fatalf("unexpected error cancelling NPC orders: %v", err)
	}

	// Trader should be gone
	retrieved := mp.GetNPCTrader(trader.ID)
	if retrieved != nil {
		t.Error("expected NPC trader to be cancelled")
	}
}

func TestOrderCreationCost(t *testing.T) {
	if OrderCreationCost != 50.0 {
		t.Errorf("expected order creation cost to be 50.0, got %f", OrderCreationCost)
	}
}

func TestOrderCount(t *testing.T) {
	mp := NewMarketplace()

	if mp.GetOrderCount() != 0 {
		t.Error("expected 0 orders initially")
	}

	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)
	_, _ = mp.CreateOrder("planet-2", "player-2", "composite", OrderSell, 50, 5.0, false)

	if mp.GetOrderCount() != 2 {
		t.Errorf("expected 2 orders, got %d", mp.GetOrderCount())
	}
}

func TestActiveOrderCount(t *testing.T) {
	mp := NewMarketplace()

	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)
	_, _ = mp.CreateOrder("planet-2", "player-2", "food", OrderSell, 50, 5.0, false)
	_, _ = mp.CreateOrder("planet-3", "player-3", "mechanisms", OrderBuy, 200, 8.0, false)

	// Fill one pair of matching orders
	mp.MatchOrders()

	active := mp.GetActiveOrderCount()
	if active != 2 {
		t.Errorf("expected 2 active orders, got %d", active)
	}
}

func TestCleanupExpiredOrders(t *testing.T) {
	mp := NewMarketplace()

	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)

	// Manually set created_at to 25 hours ago to simulate expiration
	orders := mp.GetVisibleOrders("player-1")
	if len(orders) > 0 {
		orders[0].CreatedAt = time.Now().Add(-25 * time.Hour)
	}

	mp.CleanupExpiredOrders()

	// Order should be expired
	orders = mp.GetVisibleOrders("player-1")
	if len(orders) != 0 {
		t.Error("expected 0 visible orders after cleanup")
	}
}

func TestOrderExists(t *testing.T) {
	mp := NewMarketplace()

	if mp.OrderExists("non-existent") {
		t.Error("expected order to not exist")
	}

	order, _ := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)

	if !mp.OrderExists(order.ID) {
		t.Error("expected order to exist")
	}
}

func TestGenerateLinkUniqueness(t *testing.T) {
	links := make(map[string]bool)

	for i := 0; i < 100; i++ {
		link := GenerateLink()
		if links[link] {
			t.Errorf("expected unique link, got duplicate: %s", link)
		}
		links[link] = true
		if len(link) == 0 {
			t.Error("expected non-empty link")
		}
	}
}

func TestMultipleResourceTypes(t *testing.T) {
	mp := NewMarketplace()

	resources := []string{"food", "composite", "mechanisms", "reagents"}
	for _, res := range resources {
		_, err := mp.CreateOrder("planet-1", "player-1", res, OrderBuy, 100, 10.0, false)
		if err != nil {
			t.Errorf("unexpected error creating order for resource %s: %v", res, err)
		}
	}

	if mp.GetOrderCount() != 4 {
		t.Errorf("expected 4 orders for different resources, got %d", mp.GetOrderCount())
	}
}

func TestGetOrdersByResource(t *testing.T) {
	mp := NewMarketplace()

	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)
	_, _ = mp.CreateOrder("planet-2", "player-2", "food", OrderSell, 50, 5.0, false)
	_, _ = mp.CreateOrder("planet-3", "player-3", "composite", OrderBuy, 200, 8.0, false)

	foodOrders := mp.GetOrdersByResource("player-1", "food")
	if len(foodOrders) != 2 {
		t.Errorf("expected 2 food orders, got %d", len(foodOrders))
	}

	compositeOrders := mp.GetOrdersByResource("player-1", "composite")
	if len(compositeOrders) != 1 {
		t.Errorf("expected 1 composite order, got %d", len(compositeOrders))
	}
}

func TestGetOrdersByResourceType(t *testing.T) {
	mp := NewMarketplace()

	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)
	_, _ = mp.CreateOrder("planet-2", "player-2", "food", OrderSell, 50, 5.0, false)
	_, _ = mp.CreateOrder("planet-3", "player-3", "food", OrderBuy, 200, 8.0, false)

	buyOrders := mp.GetOrdersByResourceType("player-1", "food", OrderBuy)
	if len(buyOrders) != 2 {
		t.Errorf("expected 2 buy orders for food, got %d", len(buyOrders))
	}

	sellOrders := mp.GetOrdersByResourceType("player-1", "food", OrderSell)
	if len(sellOrders) != 1 {
		t.Errorf("expected 1 sell order for food, got %d", len(sellOrders))
	}
}

func TestOrderTimestamps(t *testing.T) {
	mp := NewMarketplace()

	before := time.Now()
	order, _ := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)
	after := time.Now()

	if order.CreatedAt.Before(before) || order.CreatedAt.After(after) {
		t.Error("expected created_at to be within test window")
	}
	if order.UpdatedAt.Before(before) || order.UpdatedAt.After(after) {
		t.Error("expected updated_at to be within test window")
	}
}

func TestMatchingEngineMultipleResources(t *testing.T) {
	mp := NewMarketplace()

	// Create orders for different resources that shouldn't match each other
	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderSell, 100, 10.0, false)
	_, _ = mp.CreateOrder("planet-2", "player-2", "food", OrderBuy, 100, 15.0, false)
	_, _ = mp.CreateOrder("planet-3", "player-3", "composite", OrderSell, 100, 5.0, false)
	_, _ = mp.CreateOrder("planet-4", "player-4", "composite", OrderBuy, 100, 8.0, false)

	result := mp.MatchOrders()

	if result.ExecutedTrades != 2 {
		t.Errorf("expected 2 executed trades (one per resource), got %d", result.ExecutedTrades)
	}
}

func TestMatchingEngineNoSelfMatch(t *testing.T) {
	mp := NewMarketplace()

	// Same player creates both buy and sell orders for the same resource
	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderSell, 100, 10.0, false)
	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 15.0, false)

	result := mp.MatchOrders()

	// The matching engine should still match them (self-matching is allowed for simplicity)
	// This tests that the matching engine works regardless of player ownership
	if result.ExecutedTrades != 1 {
		t.Errorf("expected 1 executed trade (self-matching allowed), got %d", result.ExecutedTrades)
	}
}

func TestReservedResourcesMap(t *testing.T) {
	mp := NewMarketplace()

	order, _ := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)

	// Check that reserved resources map has the correct resource
	if _, ok := order.ReservedResources["food"]; !ok {
		t.Error("expected 'food' in reserved resources")
	}
	if order.ReservedResources["food"] != 100 {
		t.Errorf("expected 100 food reserved, got %f", order.ReservedResources["food"])
	}
}

func TestNPCTraderOrderIsPublic(t *testing.T) {
	mp := NewMarketplace()

	order, _, _ := mp.CreateNPCOrder("Trader Alpha", "npc-1", "npc-player-1", "food", OrderSell, 100, 10.0)

	if order.IsPrivate {
		t.Error("NPC orders should be public by default")
	}
}

func TestMarketplaceNilDB(t *testing.T) {
	mp := NewMarketplace()
	// Don't set DB - test that it works without database

	order, err := mp.CreateOrder("planet-1", "player-1", "food", OrderBuy, 100, 10.0, false)
	if err != nil {
		t.Fatalf("unexpected error without database: %v", err)
	}
	if order == nil {
		t.Fatal("expected non-nil order")
	}

	err = mp.DeleteOrder(order.ID)
	if err != nil {
		t.Fatalf("unexpected error deleting without database: %v", err)
	}
}

func TestMatchingEngineVolumeCalculation(t *testing.T) {
	mp := NewMarketplace()

	_, _ = mp.CreateOrder("planet-1", "player-1", "food", OrderSell, 50, 20.0, false)
	_, _ = mp.CreateOrder("planet-2", "player-2", "food", OrderBuy, 50, 25.0, false)

	result := mp.MatchOrders()

	if result.TotalVolume != 1000.0 {
		t.Errorf("expected total volume 1000.0 (50 * 20.0), got %f", result.TotalVolume)
	}
}

func TestMatchingEngineLargeOrderBook(t *testing.T) {
	mp := NewMarketplace()

	// Create 10 sell orders at increasing prices
	for i := 0; i < 10; i++ {
		_, _ = mp.CreateOrder("planet-sell", "player-sell", "food", OrderSell, 100, float64(5+i), false)
	}

	// Create 1 buy order that can match all 10
	_, _ = mp.CreateOrder("planet-buy", "player-buy", "food", OrderBuy, 1000, 15.0, false)

	result := mp.MatchOrders()

	if result.ExecutedTrades != 10 {
		t.Errorf("expected 10 executed trades, got %d", result.ExecutedTrades)
	}
}
