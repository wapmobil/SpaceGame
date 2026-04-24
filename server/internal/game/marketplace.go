package game

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"spacegame/internal/db"
)

// OrderType represents the type of market order.
type OrderType string

const (
	OrderBuy  OrderType = "buy"
	OrderSell OrderType = "sell"
)

// OrderStatus represents the status of a market order.
type OrderStatus string

const (
	OrderActive    OrderStatus = "active"
	OrderFilled    OrderStatus = "filled"
	OrderCancelled OrderStatus = "cancelled"
	OrderExpired   OrderStatus = "expired"
)

// MarketOrder represents an order on the marketplace.
type MarketOrder struct {
	ID               string
	PlanetID         string
	PlayerID         string
	Resource         string
	OrderType        OrderType
	Amount           float64
	Price            float64
	IsPrivate        bool
	Link             string
	Status           OrderStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ReservedResources map[string]float64
}

// MarketOrderResponse represents a market order in API responses.
type MarketOrderResponse struct {
	ID               string                 `json:"id"`
	PlanetID         string                 `json:"planet_id"`
	PlayerID         string                 `json:"player_id"`
	Resource         string                 `json:"resource"`
	OrderType        string                 `json:"order_type"`
	Amount           float64                `json:"amount"`
	Price            float64                `json:"price"`
	IsPrivate        bool                   `json:"is_private"`
	Link             string                 `json:"link,omitempty"`
	Status           string                 `json:"status"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	ReservedResources map[string]float64    `json:"reserved_resources,omitempty"`
}

// MarketOrderRequest is the request body for creating a market order.
type MarketOrderRequest struct {
	Resource string  `json:"resource"`
	OrderType string `json:"order_type"`
	Amount   float64 `json:"amount"`
	Price    float64 `json:"price"`
	IsPrivate bool   `json:"is_private"`
}

// NPCTrader represents an AI trader that places orders.
type NPCTrader struct {
	ID        string
	Name      string
	PlanetID  string
	OrderID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Marketplace manages all market orders.
type Marketplace struct {
	orders    map[string]*MarketOrder
	mu        sync.RWMutex
	db        *db.Database
	NPCTraders map[string]*NPCTrader
}

// OrderMatchingResult represents the result of a matching engine execution.
type OrderMatchingResult struct {
	MatchedOrders []string
	ExecutedTrades int
	TotalVolume    float64
}

const (
	OrderCreationCost = 50.0 // energy cost to create/delete an order
	MaxOrderAmount    = 1000000.0
	MinPrice          = 0.01
	MaxPrice          = 10000.0
)

// NewMarketplace creates a new marketplace instance.
func NewMarketplace() *Marketplace {
	return &Marketplace{
		orders:       make(map[string]*MarketOrder),
		NPCTraders:   make(map[string]*NPCTrader),
	}
}

// SetDB sets the database connection for the marketplace.
func (m *Marketplace) SetDB(d *db.Database) {
	m.db = d
}

// GenerateLink creates a unique link for private orders.
func GenerateLink() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateOrder creates a new market order with resource reservation.
func (m *Marketplace) CreateOrder(planetID, playerID, resource string, orderType OrderType, amount, price float64, isPrivate bool) (*MarketOrder, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("invalid order amount: must be positive")
	}
	if amount > MaxOrderAmount {
		return nil, fmt.Errorf("order amount exceeds maximum (%.0f)", MaxOrderAmount)
	}
	if price < MinPrice || price > MaxPrice {
		return nil, fmt.Errorf("price must be between %.2f and %.2f", MinPrice, MaxPrice)
	}
	if orderType != OrderBuy && orderType != OrderSell {
		return nil, fmt.Errorf("invalid order type: must be 'buy' or 'sell'")
	}

	// Validate resource name
	validResources := map[string]bool{
		"food": true, "composite": true, "mechanisms": true, "reagents": true,
	}
	if !validResources[resource] {
		return nil, fmt.Errorf("invalid resource: %s", resource)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate unique link for private orders
	link := ""
	if isPrivate {
		link = GenerateLink()
	}

	order := &MarketOrder{
		ID:        planetID + "_order_" + time.Now().Format("20060102150405") + "_" + fmt.Sprintf("%d", len(m.orders)),
		PlanetID:  planetID,
		PlayerID:  playerID,
		Resource:  resource,
		OrderType: orderType,
		Amount:    amount,
		Price:     price,
		IsPrivate: isPrivate,
		Link:      link,
		Status:    OrderActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ReservedResources: map[string]float64{
			resource: amount,
		},
	}

	m.orders[order.ID] = order

	// Save to database
	if m.db != nil {
		if err := m.saveOrderToDB(order); err != nil {
			delete(m.orders, order.ID)
			return nil, fmt.Errorf("failed to save order to database: %w", err)
		}
	}

	return order, nil
}

// DeleteOrder removes an order and refunds reserved resources.
func (m *Marketplace) DeleteOrder(orderID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	order, exists := m.orders[orderID]
	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}
	if order.Status != OrderActive {
		return fmt.Errorf("order is not active (status: %s)", order.Status)
	}

	// Mark as cancelled
	order.Status = OrderCancelled
	order.UpdatedAt = time.Now()

	// Save to database
	if m.db != nil {
		if err := m.saveOrderToDB(order); err != nil {
			return fmt.Errorf("failed to update order in database: %w", err)
		}
	}

	delete(m.orders, orderID)
	return nil
}

// GetOrder retrieves an order by ID.
func (m *Marketplace) GetOrder(orderID string) *MarketOrder {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.orders[orderID]
}

// GetPrivateOrder retrieves a private order by its link.
func (m *Marketplace) GetPrivateOrder(link string) *MarketOrder {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, order := range m.orders {
		if order.Link == link && order.IsPrivate && order.Status == OrderActive {
			return order
		}
	}
	return nil
}

// GetMyOrders returns all active orders for a player.
func (m *Marketplace) GetMyOrders(playerID string) []*MarketOrder {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*MarketOrder
	for _, order := range m.orders {
		if order.PlayerID == playerID && order.Status == OrderActive {
			result = append(result, order)
		}
	}
	return result
}

// GetVisibleOrders returns all visible orders for the global market view.
func (m *Marketplace) GetVisibleOrders(playerID string) []*MarketOrder {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*MarketOrder
	for _, order := range m.orders {
		if order.Status != OrderActive {
			continue
		}
		// Private orders are only visible to their owner
		if order.IsPrivate && order.PlayerID != playerID {
			continue
		}
		result = append(result, order)
	}
	return result
}

// GetOrdersByResource returns visible orders filtered by resource.
func (m *Marketplace) GetOrdersByResource(playerID, resource string) []*MarketOrder {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*MarketOrder
	for _, order := range m.orders {
		if order.Status != OrderActive {
			continue
		}
		if order.IsPrivate && order.PlayerID != playerID {
			continue
		}
		if order.Resource == resource {
			result = append(result, order)
		}
	}
	return result
}

// MatchOrders runs the matching engine to find and execute matching buy/sell orders.
func (m *Marketplace) MatchOrders() *OrderMatchingResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := &OrderMatchingResult{}

	// Group active orders by resource
	ordersByResource := make(map[string][]*MarketOrder)
	for _, order := range m.orders {
		if order.Status == OrderActive {
			ordersByResource[order.Resource] = append(ordersByResource[order.Resource], order)
		}
	}

	for _, orders := range ordersByResource {
		var buyOrders []*MarketOrder
		var sellOrders []*MarketOrder

		for _, order := range orders {
			if order.OrderType == OrderBuy {
				buyOrders = append(buyOrders, order)
			} else {
				sellOrders = append(sellOrders, order)
			}
		}

		// Sort buy orders by price descending (highest buyer first)
		for i := 0; i < len(buyOrders); i++ {
			for j := i + 1; j < len(buyOrders); j++ {
				if buyOrders[j].Price > buyOrders[i].Price {
					buyOrders[i], buyOrders[j] = buyOrders[j], buyOrders[i]
				}
			}
		}

		// Sort sell orders by price ascending (lowest seller first)
		for i := 0; i < len(sellOrders); i++ {
			for j := i + 1; j < len(sellOrders); j++ {
				if sellOrders[j].Price < sellOrders[i].Price {
					sellOrders[i], sellOrders[j] = sellOrders[j], sellOrders[i]
				}
			}
		}

		// Match orders
		buyIdx := 0
		sellIdx := 0
		for buyIdx < len(buyOrders) && sellIdx < len(sellOrders) {
			buy := buyOrders[buyIdx]
			sell := sellOrders[sellIdx]

			// Match if buyer's price >= seller's price
			if buy.Price >= sell.Price {
				// Calculate trade amount (minimum of both orders)
				tradeAmount := math.Min(buy.Amount, sell.Amount)

				// Execute trade at seller's price
				tradePrice := sell.Price
				tradeVolume := tradeAmount * tradePrice

				// Mark orders as filled (or partially filled)
				buy.Amount -= tradeAmount
				sell.Amount -= tradeAmount

				if buy.Amount <= 0.001 {
					buy.Status = OrderFilled
					buy.UpdatedAt = time.Now()
					buyIdx++
				} else {
					buy.UpdatedAt = time.Now()
				}

				if sell.Amount <= 0.001 {
					sell.Status = OrderFilled
					sell.UpdatedAt = time.Now()
					sellIdx++
				} else {
					sell.UpdatedAt = time.Now()
				}

				result.MatchedOrders = append(result.MatchedOrders, buy.ID, sell.ID)
				result.ExecutedTrades++
				result.TotalVolume += tradeVolume

				// Save to database
				if m.db != nil {
					m.saveOrderToDB(buy)
					m.saveOrderToDB(sell)
				}
			} else {
				break
			}
		}
	}

	return result
}

// CreateNPCOrder creates an order for an NPC trader.
func (m *Marketplace) CreateNPCOrder(name, planetID, playerID, resource string, orderType OrderType, amount, price float64) (*MarketOrder, *NPCTrader, error) {
	order, err := m.CreateOrder(planetID, playerID, resource, orderType, amount, price, false)
	if err != nil {
		return nil, nil, err
	}

	trader := &NPCTrader{
		ID:        planetID + "_npc_" + time.Now().Format("20060102150405"),
		Name:      name,
		PlanetID:  planetID,
		OrderID:   order.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.mu.Lock()
	m.NPCTraders[trader.ID] = trader
	m.mu.Unlock()

	if m.db != nil {
		if err := m.saveNPCTraderToDB(trader); err != nil {
			return nil, nil, fmt.Errorf("failed to save NPC trader to database: %w", err)
		}
	}

	return order, trader, nil
}

// GenerateNPCOrders generates random orders for NPC traders based on total market level.
// Base trader count is 3, each market level adds 1 trader.
func (m *Marketplace) GenerateNPCOrders(marketLevel int) {
	npcNames := []string{
		"Trade Station Alpha",
		"Merchant Outpost Beta",
		"Galactic Trading Post",
		"Stellar Market Delta",
		"Void Trader Epsilon",
		"Nebula Exchange Zeta",
		"Quantum Bazaar Eta",
		"Cosmic Depot Theta",
	}

	traderCount := 3 + marketLevel
	if traderCount > len(npcNames) {
		traderCount = len(npcNames)
	}

	validResources := []string{"food", "composite", "mechanisms", "reagents"}
	validTypes := []OrderType{OrderBuy, OrderSell}

	for i := 0; i < traderCount; i++ {
		name := npcNames[i]
		planetID := fmt.Sprintf("npc_%s", name)
		playerID := fmt.Sprintf("npc_%s_player", name)

		// Generate 1-3 orders per NPC
		orderCount := randInt(1, 4)
		for j := 0; j < orderCount; j++ {
			resource := validResources[randInt(0, len(validResources)-1)]
			orderType := validTypes[randInt(0, len(validTypes)-1)]
			amount := float64(randInt(10, 500))
			price := float64(randInt(10, 500))

			_, _, err := m.CreateNPCOrder(name, planetID, playerID, resource, orderType, amount, price)
			if err != nil {
				continue
			}
		}
	}
}

// GetNPCTrader returns an NPC trader by ID.
func (m *Marketplace) GetNPCTrader(id string) *NPCTrader {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.NPCTraders[id]
}

// GetAllNPCTraders returns all NPC traders.
func (m *Marketplace) GetAllNPCTraders() []*NPCTrader {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*NPCTrader
	for _, trader := range m.NPCTraders {
		result = append(result, trader)
	}
	return result
}

// CancelNPCOrders cancels all orders from a specific NPC trader.
func (m *Marketplace) CancelNPCOrders(npcID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	trader, exists := m.NPCTraders[npcID]
	if !exists {
		return fmt.Errorf("NPC trader not found: %s", npcID)
	}

	if trader.OrderID != "" {
		if order, ok := m.orders[trader.OrderID]; ok && order.Status == OrderActive {
			order.Status = OrderCancelled
			order.UpdatedAt = time.Now()
			if m.db != nil {
				m.saveOrderToDB(order)
			}
		}
	}

	delete(m.NPCTraders, npcID)
	return nil
}

// GetOrderCount returns the total number of orders.
func (m *Marketplace) GetOrderCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.orders)
}

// GetActiveOrderCount returns the number of active orders.
func (m *Marketplace) GetActiveOrderCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, order := range m.orders {
		if order.Status == OrderActive {
			count++
		}
	}
	return count
}

// CleanupExpiredOrders removes expired orders.
func (m *Marketplace) CleanupExpiredOrders() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, order := range m.orders {
		if order.Status == OrderActive && now.Sub(order.CreatedAt).Hours() > 24 {
			order.Status = OrderExpired
			order.UpdatedAt = now
			if m.db != nil {
				m.saveOrderToDB(order)
			}
		}
	}
}

// OrderExists checks if an order exists.
func (m *Marketplace) OrderExists(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.orders[id]
	return exists
}

// GetOrderByResourceType returns visible orders filtered by resource and type.
func (m *Marketplace) GetOrdersByResourceType(playerID, resource string, orderType OrderType) []*MarketOrder {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*MarketOrder
	for _, order := range m.orders {
		if order.Status != OrderActive {
			continue
		}
		if order.IsPrivate && order.PlayerID != playerID {
			continue
		}
		if order.Resource == resource && order.OrderType == orderType {
			result = append(result, order)
		}
	}
	return result
}

// Helper function for random integer generation (inclusive of min and max).
func randInt(min, max int) int {
	if min >= max {
		return min
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return min + int(n.Int64())
}

// Database helper functions

func (m *Marketplace) saveOrderToDB(order *MarketOrder) error {
	if m.db == nil {
		return nil
	}
	return m.saveOrderToDBUnsafe(order)
}

func (m *Marketplace) saveOrderToDBUnsafe(order *MarketOrder) error {
	if m.db == nil {
		return nil
	}

	var err error
	if order.Status == OrderActive {
		_, err = m.db.Exec(`
			INSERT INTO market_orders 
			(id, planet_id, player_id, resource, type, amount, price, is_private, link, status, 
			 reserved_food, reserved_composite, reserved_mechanisms, reserved_reagents, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			ON CONFLICT (id) DO UPDATE SET
				status = $10, updated_at = $16,
				reserved_food = EXCLUDED.reserved_food,
				reserved_composite = EXCLUDED.reserved_composite,
				reserved_mechanisms = EXCLUDED.reserved_mechanisms,
				reserved_reagents = EXCLUDED.reserved_reagents
		`, order.ID, order.PlanetID, order.PlayerID, order.Resource, string(order.OrderType),
			order.Amount, order.Price, order.IsPrivate, order.Link, string(order.Status),
			order.ReservedResources["food"], order.ReservedResources["composite"],
			order.ReservedResources["mechanisms"], order.ReservedResources["reagents"],
			order.CreatedAt, order.UpdatedAt)
	} else {
		_, err = m.db.Exec(`
			UPDATE market_orders 
			SET status = $1, updated_at = $2
			WHERE id = $3
		`, string(order.Status), order.UpdatedAt, order.ID)
	}

	return err
}

func (m *Marketplace) saveNPCTraderToDB(trader *NPCTrader) error {
	if m.db == nil {
		return nil
	}

	_, err := m.db.Exec(`
		INSERT INTO npc_traders 
		(id, name, planet_id, order_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, planet_id = EXCLUDED.planet_id,
			order_id = EXCLUDED.order_id, updated_at = EXCLUDED.updated_at
	`, trader.ID, trader.Name, trader.PlanetID, trader.OrderID, trader.CreatedAt, trader.UpdatedAt)

	return err
}
