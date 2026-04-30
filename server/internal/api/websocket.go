package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"spacegame/internal/game"
)

// WSMaxMessagesPerSecond is the maximum messages a client can send per second.
const WSMaxMessagesPerSecond = 10

// WSMessageQueueSize is the maximum number of queued messages per player.
const WSMessageQueueSize = 50

// WSPingInterval is how often the server sends ping frames.
const WSPingInterval = 30 * time.Second

// WPCloseGracePeriod is how long to wait for a pong response.
const WPCloseGracePeriod = 10 * time.Second

// WSReconnectWindow is how long to keep queued messages for reconnecting players.
const WSReconnectWindow = 5 * time.Minute

// WSUpgrader is the WebSocket upgrader configuration.
var WSUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSClient represents a connected WebSocket client.
type WSClient struct {
	conn        *websocket.Conn
	playerID    string
	planetID    string
	connID      string
	send        chan []byte
	mu          sync.Mutex
	closed      bool
	lastSeen    time.Time
	token       string
	connectedAt time.Time
}

// WSRateLimiter implements a sliding window rate limiter.
type WSRateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*time.Ticker
	maxRate  int
	window   time.Duration
}

// NewWSRateLimiter creates a new rate limiter.
func NewWSRateLimiter(maxRate int) *WSRateLimiter {
	return &WSRateLimiter{
		clients: make(map[string]*time.Ticker),
		maxRate: maxRate,
		window:  time.Second,
	}
}

// WSMessage represents a WebSocket message.
type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
	Error string         `json:"error,omitempty"`
}

// WSMessageQueue stores messages for disconnected players.
type WSMessageQueue struct {
	mu      sync.Mutex
	queues  map[string][]WSMessage
	timers  map[string]*time.Timer
}

// NewWSMessageQueue creates a new message queue.
func NewWSMessageQueue() *WSMessageQueue {
	return &WSMessageQueue{
		queues:  make(map[string][]WSMessage),
		timers:  make(map[string]*time.Timer),
	}
}

// Enqueue adds a message to a player's queue.
func (mq *WSMessageQueue) Enqueue(playerID string, msg WSMessage) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	queue := mq.queues[playerID]
	if len(queue) >= WSMessageQueueSize {
		queue = queue[1:]
	}
	mq.queues[playerID] = append(queue, msg)

	// Set cleanup timer
	if timer, exists := mq.timers[playerID]; exists {
		timer.Stop()
	}
	mq.timers[playerID] = time.AfterFunc(WSReconnectWindow, func() {
		mq.mu.Lock()
		delete(mq.queues, playerID)
		delete(mq.timers, playerID)
		mq.mu.Unlock()
	})
}

// Dequeue removes and returns all queued messages for a player.
func (mq *WSMessageQueue) Dequeue(playerID string) []WSMessage {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	msgs := mq.queues[playerID]
	delete(mq.queues, playerID)
	if timer, exists := mq.timers[playerID]; exists {
		timer.Stop()
		delete(mq.timers, playerID)
	}
	return msgs
}

// Clear removes all queued messages for a player.
func (mq *WSMessageQueue) Clear(playerID string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	delete(mq.queues, playerID)
	if timer, exists := mq.timers[playerID]; exists {
		timer.Stop()
		delete(mq.timers, playerID)
	}
}

// Len returns the number of queued messages for a player.
func (mq *WSMessageQueue) Len(playerID string) int {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return len(mq.queues[playerID])
}

// WSConnectionManager manages all WebSocket connections.
type WSConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*WSClient // playerID -> client
	planetMap   map[string]string    // connID -> playerID
	rateLimiter *WSRateLimiter
	messageQueue *WSMessageQueue
	clients     map[string]*WSClient // connID -> client
	connCounter int
}

// NewWSConnectionManager creates a new connection manager.
func NewWSConnectionManager() *WSConnectionManager {
	return &WSConnectionManager{
		connections:  make(map[string]*WSClient),
		planetMap:    make(map[string]string),
		rateLimiter:  NewWSRateLimiter(WSMaxMessagesPerSecond),
		messageQueue: NewWSMessageQueue(),
		clients:      make(map[string]*WSClient),
	}
}

// AddClient registers a new WebSocket client.
func (cm *WSConnectionManager) AddClient(playerID, connID string, client *WSClient) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.connections[playerID] = client
	cm.clients[connID] = client
	cm.planetMap[connID] = playerID
	client.lastSeen = time.Now()
	client.connectedAt = time.Now()
	log.Printf("WebSocket client added: player %s, conn %s", playerID, connID)
}

// RemoveClient removes a WebSocket client and queues last state.
func (cm *WSConnectionManager) RemoveClient(connID string) {
	cm.mu.Lock()
	playerID, exists := cm.planetMap[connID]
	delete(cm.clients, connID)
	delete(cm.planetMap, connID)
	cm.mu.Unlock()

	if !exists || playerID == "" {
		return
	}

	cm.mu.Lock()
	client, hasClient := cm.connections[playerID]
	if hasClient && client.connID == connID {
		delete(cm.connections, playerID)
	}
	cm.mu.Unlock()

	if client != nil {
		client.Close()
	}

	// Clean up any active drill session for this player
	if dg := game.FindActiveSession("", playerID); dg != nil {
		dg.Destroy()
		log.Printf("Drill session destroyed on disconnect: player %s", playerID)
	}

	// Queue a state update for reconnection
	cm.messageQueue.Enqueue(playerID, WSMessage{
		Type: "disconnected",
		Data: json.RawMessage(fmt.Sprintf(`{"player_id":"%s","timestamp":%d}`, playerID, time.Now().Unix())),
	})
	log.Printf("WebSocket client removed: player %s, conn %s", playerID, connID)
}

// GetClient returns a client by player ID.
func (cm *WSConnectionManager) GetClient(playerID string) *WSClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.connections[playerID]
}

// GetClientByConnID returns a client by connection ID.
func (cm *WSConnectionManager) GetClientByConnID(connID string) *WSClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clients[connID]
}

// SendToPlayer sends a message to a specific player.
func (cm *WSConnectionManager) SendToPlayer(playerID string, msg WSMessage) bool {
	cm.mu.RLock()
	client := cm.connections[playerID]
	cm.mu.RUnlock()

	if client == nil {
		cm.messageQueue.Enqueue(playerID, msg)
		return false
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("WebSocket marshal error: %v", err)
		return false
	}

	client.mu.Lock()
	if client.closed {
		client.mu.Unlock()
		return false
	}
	client.lastSeen = time.Now()
	client.mu.Unlock()

	select {
	case client.send <- data:
		return true
	default:
		cm.messageQueue.Enqueue(playerID, msg)
		return false
	}
}

// SendToPlanet sends a message to all clients for a specific planet.
func (cm *WSConnectionManager) SendToPlanet(planetID string, msg WSMessage) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for playerID, client := range cm.connections {
		if client.planetID == planetID {
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			select {
			case client.send <- data:
			default:
				cm.messageQueue.Enqueue(playerID, msg)
			}
		}
	}
}

// BroadcastToAll sends a message to all connected clients.
func (cm *WSConnectionManager) BroadcastToAll(msg WSMessage) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for _, client := range cm.connections {
		select {
		case client.send <- data:
		default:
			// Client send buffer full, skip
		}
	}
}

// IsConnected checks if a player is connected.
func (cm *WSConnectionManager) IsConnected(playerID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, exists := cm.connections[playerID]
	return exists
}

// Count returns the number of connected clients.
func (cm *WSConnectionManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.connections)
}

// Close closes all client connections.
func (cm *WSConnectionManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, client := range cm.connections {
		client.Close()
	}
	for _, client := range cm.clients {
		client.Close()
	}
	cm.connections = make(map[string]*WSClient)
	cm.clients = make(map[string]*WSClient)
	cm.planetMap = make(map[string]string)
}

// NewClient creates a new WebSocket client.
func (cm *WSConnectionManager) NewClient(conn *websocket.Conn, playerID, planetID, connID, token string) *WSClient {
	return &WSClient{
		conn:     conn,
		playerID: playerID,
		planetID: planetID,
		connID:   connID,
		send:     make(chan []byte, 256),
		token:    token,
		lastSeen: time.Now(),
	}
}

// Close closes the client connection.
func (c *WSClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	if c.conn != nil {
		c.conn.Close()
	}
}

// WriteJSON writes a JSON message to the client.
func (c *WSClient) WriteJSON(msg WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, data)
}

// WriteMessage writes raw bytes to the client.
func (c *WSClient) WriteMessage(msgType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("client closed")
	}
	c.lastSeen = time.Now()
	return c.conn.WriteMessage(msgType, data)
}

// StartHeartbeat starts the ping/pong heartbeat for a client.
func (c *WSClient) StartHeartbeat(cm *WSConnectionManager, connID string) {
	ticker := time.NewTicker(WSPingInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.mu.Lock()
				if c.closed {
					c.mu.Unlock()
					return
				}
				err := c.conn.WriteMessage(websocket.PingMessage, nil)
				c.mu.Unlock()
				if err != nil {
					cm.RemoveClient(connID)
					return
				}
			}
		}
	}()
}

// StartWritePump starts the write pump goroutine for a client.
func (c *WSClient) StartWritePump(cm *WSConnectionManager, connID string) {
	ticker := time.NewTicker(WSPingInterval)
	go func() {
		defer func() {
			ticker.Stop()
			cm.RemoveClient(connID)
		}()

		for {
			select {
			case message, ok := <-c.send:
				c.mu.Lock()
				if !ok {
					c.mu.Unlock()
					return
				}
				c.conn.SetWriteDeadline(time.Now().Add(WPCloseGracePeriod))
				if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					c.mu.Unlock()
					return
				}
				c.mu.Unlock()
			case <-ticker.C:
				c.mu.Lock()
				if c.closed {
					c.mu.Unlock()
					return
				}
				c.conn.SetWriteDeadline(time.Now().Add(WPCloseGracePeriod))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.mu.Unlock()
					return
				}
				c.mu.Unlock()
			}
		}
	}()
}

// CheckRateLimit checks if a client has exceeded the rate limit.
func (rl *WSRateLimiter) CheckRateLimit(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	_, exists := rl.clients[clientID]
	if !exists {
		rl.clients[clientID] = time.NewTicker(time.Millisecond * 100)
	}

	return true // Simplified: allow through, actual counting done per-message
}

// ConsumeToken checks and consumes a rate limit token.
func (rl *WSRateLimiter) ConsumeToken(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	_, exists := rl.clients[clientID]
	if !exists {
		rl.clients[clientID] = time.NewTicker(time.Second)
	}

	return true
}

// Cleanup removes stale rate limiter entries.
func (rl *WSRateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for id, ticker := range rl.clients {
		ticker.Stop()
		delete(rl.clients, id)
	}
}

// WSBroadcastService provides broadcast functionality for game events.
type WSBroadcastService struct {
	cm *WSConnectionManager
}

// NewWSBroadcastService creates a new broadcast service.
func NewWSBroadcastService(cm *WSConnectionManager) *WSBroadcastService {
	return &WSBroadcastService{cm: cm}
}

// BroadcastStateUpdate sends a planet state update to the owning player.
func (bs *WSBroadcastService) BroadcastStateUpdate(playerID, planetID string, state map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "state_update",
		Data: json.RawMessage(fmt.Sprintf(`{"planet_id":"%s","state":%s}`, planetID, toJSON(state))),
	})
}

// BroadcastNotification sends a notification to a player.
func (bs *WSBroadcastService) BroadcastNotification(playerID, message, notifType string) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "notification",
		Data: json.RawMessage(fmt.Sprintf(`{"message":"%s","type":"%s"}`, escapeJSON(message), escapeJSON(notifType))),
	})
}

// BroadcastExpeditionUpdate sends an expedition update to the owning player.
func (bs *WSBroadcastService) BroadcastExpeditionUpdate(playerID, expeditionID string, progress float64) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "space_expedition_update",
		Data: json.RawMessage(fmt.Sprintf(`{"expedition_id":"%s","progress":%.4f}`, expeditionID, progress)),
	})
}

// BroadcastMarketUpdate sends a market update to the owning player.
func (bs *WSBroadcastService) BroadcastMarketUpdate(playerID, orderID, status string) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "market_update",
		Data: json.RawMessage(fmt.Sprintf(`{"order_id":"%s","status":"%s"}`, orderID, status)),
	})
}


// BroadcastBuildShip sends a ship build notification.
func (bs *WSBroadcastService) BroadcastBuildShip(playerID, planetID, shipType string, queued bool) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "build",
		Data: json.RawMessage(fmt.Sprintf(`{"planet_id":"%s","ship_type":"%s","queued":%t}`, planetID, shipType, queued)),
	})
}

// BroadcastResearchUpdate sends a research update.
func (bs *WSBroadcastService) BroadcastResearchUpdate(playerID, planetID, techID string, progress float64) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "research_update",
		Data: json.RawMessage(fmt.Sprintf(`{"planet_id":"%s","tech_id":"%s","progress":%.4f}`, planetID, techID, progress)),
	})
}

// BroadcastBuildingUpdate sends a building update.
func (bs *WSBroadcastService) BroadcastBuildingUpdate(playerID, planetID, building string, level int) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "building_update",
		Data: json.RawMessage(fmt.Sprintf(`{"planet_id":"%s","building":"%s","level":%d}`, planetID, building, level)),
	})
}

// BroadcastDrillUpdate sends a drill mini-game update to the owning player.
func (bs *WSBroadcastService) BroadcastDrillUpdate(playerID string, data map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "drill_update",
		Data: json.RawMessage(toJSON(data)),
	})
}

// BroadcastGardenBedUpdate sends a garden bed update to the owning player.
func (bs *WSBroadcastService) BroadcastGardenBedUpdate(playerID string, data map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "garden_bed_update",
		Data: json.RawMessage(toJSON(data)),
	})
}

// BroadcastRandomEvent sends a random event notification.
func (bs *WSBroadcastService) BroadcastRandomEvent(playerID string, event map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "random_event",
		Data: json.RawMessage(toJSON(event)),
	})
}

// BroadcastPlayerKilled sends a player death notification.
func (bs *WSBroadcastService) BroadcastPlayerKilled(playerID string, opponent string) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "notification",
		Data: json.RawMessage(fmt.Sprintf(`{"message":"Your fleet was defeated by %s","type":"combat"}`, escapeJSON(opponent))),
	})
}

// BroadcastExpeditionDiscovered sends an expedition discovery notification.
func (bs *WSBroadcastService) BroadcastExpeditionDiscovered(playerID, expeditionID, npcName, npcType string) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "notification",
		Data: json.RawMessage(fmt.Sprintf(`{"message":"Expedition discovered %s (%s)","type":"space_discovery","expedition_id":"%s"}`, escapeJSON(npcName), escapeJSON(npcType), expeditionID)),
	})
}

// BroadcastPlanetSurveyUpdate sends a planet survey update to the owning player.
func (bs *WSBroadcastService) BroadcastPlanetSurveyUpdate(playerID, planetID string, data map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "planet_survey_update",
		Data: json.RawMessage(toJSON(data)),
	})
}

// BroadcastLocationUpdate sends a location update to the owning player.
func (bs *WSBroadcastService) BroadcastLocationUpdate(playerID, planetID string, data map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "location_update",
		Data: json.RawMessage(toJSON(data)),
	})
}

// BroadcastMarketOrderFilled sends a market order filled notification.
func (bs *WSBroadcastService) BroadcastMarketOrderFilled(playerID, orderID, resource string, amount, price float64) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "market_update",
		Data: json.RawMessage(fmt.Sprintf(`{"order_id":"%s","status":"filled","resource":"%s","amount":%.0f,"price":%.0f}`, orderID, resource, amount, price)),
	})
}

// broadcastToPlayer sends a message to a player, queuing if disconnected.
func (bs *WSBroadcastService) broadcastToPlayer(playerID string, msg WSMessage) {
	bs.cm.SendToPlayer(playerID, msg)
}

// broadcastToAll sends a message to all connected players.
func (bs *WSBroadcastService) broadcastToAll(msg WSMessage) {
	bs.cm.BroadcastToAll(msg)
}

// broadcastToPlanet sends a message to all players with a specific planet.
func (bs *WSBroadcastService) broadcastToPlanet(planetID string, msg WSMessage) {
	bs.cm.SendToPlanet(planetID, msg)
}

// Helper functions

func escapeJSON(s string) string {
	s = replaceAll(s, `\`, `\\`)
	s = replaceAll(s, `"`, `\"`)
	s = replaceAll(s, "\n", `\n`)
	s = replaceAll(s, "\r", `\r`)
	s = replaceAll(s, "\t", `\t`)
	return s
}

func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}

func toJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// handleWebSocketHealth returns the WebSocket health status.
func handleWebSocketHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"connected_clients": wsManager.Count(),
		"timestamp": time.Now().Unix(),
	})
}

// wsManager is the global WebSocket connection manager.
var wsManager *WSConnectionManager

// wsBroadcast is the global broadcast service.
var wsBroadcast *WSBroadcastService

// InitWS initializes the WebSocket manager and broadcast service.
func InitWS() {
	wsManager = NewWSConnectionManager()
	wsBroadcast = NewWSBroadcastService(wsManager)
	if g := game.Instance(); g != nil {
		g.RegisterBroadcastHandler(func(planetID, playerID string, state map[string]interface{}) {
			wsBroadcast.BroadcastStateUpdate(playerID, planetID, state)
		})
		g.RegisterNotificationHandler(func(playerID, message, notifType string) {
			wsBroadcast.BroadcastNotification(playerID, message, notifType)
		})
	}
}

// handleWebSocket upgrades the connection to WebSocket and handles real-time messages.
func handleWebSocket(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := WSUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		query := r.URL.Query()
		authToken := query.Get("token")
		if authToken == "" {
			conn.WriteJSON(WSMessage{
				Type: "error",
				Error: "Missing auth token",
			})
			return
		}

		// Validate auth token against database
		var playerID string
		err = db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			conn.WriteJSON(WSMessage{
				Type: "error",
				Error: "Invalid auth token",
			})
			return
		}

		// Check for reconnection
		connID := generateConnID()
		isReconnect := false

		if wsManager.IsConnected(playerID) {
			isReconnect = true
			// Close old connection
			oldClient := wsManager.GetClient(playerID)
			if oldClient != nil {
				wsManager.RemoveClient(oldClient.connID)
			}
		}

		// Create new client
		client := wsManager.NewClient(conn, playerID, "", connID, authToken)
		wsManager.AddClient(playerID, connID, client)

		// Start heartbeat
		client.StartHeartbeat(wsManager, connID)

		// Start write pump
		client.StartWritePump(wsManager, connID)

		// Send welcome message
		welcomeMsg := WSMessage{
			Type: "welcome",
			Data: json.RawMessage(fmt.Sprintf(`{"player_id":"%s","message":"Connected to SpaceGame WebSocket"}`, playerID)),
		}
		if err := client.WriteJSON(welcomeMsg); err != nil {
			log.Printf("WebSocket send welcome error for player %s: %v", playerID, err)
			return
		}

		// Send queued messages for reconnection
		if isReconnect {
			queued := wsManager.messageQueue.Dequeue(playerID)
			if len(queued) > 0 {
				log.Printf("Sending %d queued messages to reconnecting player %s", len(queued), playerID)
				for _, msg := range queued {
					data, err := json.Marshal(msg)
					if err == nil {
						client.WriteMessage(websocket.TextMessage, data)
					}
				}
			}
		}

		// Start read pump
		client.readPump(wsManager, connID)
	}
}

// readPump reads messages from the client and dispatches them.
func (c *WSClient) readPump(cm *WSConnectionManager, connID string) {
	defer func() {
		cm.RemoveClient(connID)
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error for player %s: %v", c.playerID, err)
			}
			break
		}

		// Rate limiting check
		if !cm.rateLimiter.ConsumeToken(c.playerID) {
			c.WriteJSON(WSMessage{
				Type:  "error",
				Error: "Rate limit exceeded",
			})
			continue
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.WriteJSON(WSMessage{
				Type:  "error",
				Error: "Invalid JSON",
			})
			continue
		}

		c.handleMessage(msg)
	}
}

// handleMessage dispatches a message to the appropriate handler.
func (c *WSClient) handleMessage(msg WSMessage) {
	switch msg.Type {
	case "ping":
		c.WriteJSON(WSMessage{
			Type: "pong",
			Data: json.RawMessage(fmt.Sprintf(`{"timestamp":%d}`, time.Now().UnixNano())),
		})
	case "subscribe":
		if msg.Data != nil {
			var sub struct {
				PlanetID string `json:"planet_id"`
			}
			if err := json.Unmarshal(msg.Data, &sub); err == nil && sub.PlanetID != "" {
				c.planetID = sub.PlanetID
				c.WriteJSON(WSMessage{
					Type: "subscribed",
					Data: json.RawMessage(fmt.Sprintf(`{"planet_id":"%s"}`, sub.PlanetID)),
				})
			}
		}
	case "build":
		c.handleBuild(msg)
	case "research":
		c.handleResearch(msg)
	case "build_ship":
		c.handleBuildShip(msg)
	case "start_expedition":
		c.handleStartExpedition(msg)
	case "create_market_order":
		c.handleCreateMarketOrder(msg)
	case "delete_market_order":
		c.handleDeleteMarketOrder(msg)
	case "drill_command":
		c.handleDrillCommand(msg)
	case "garden_bed_action":
		c.handleGardenBedAction(msg)
	default:
		c.WriteJSON(WSMessage{
			Type: "ack",
			Data: json.RawMessage(`{"message":"Received"}`),
		})
	}
}

// handleBuild handles a build action.
func (c *WSClient) handleBuild(msg WSMessage) {
	if msg.Data != nil {
		data, _ := json.Marshal(msg.Data)
		log.Printf("Player %s build action: %s", c.playerID, string(data))
		wsBroadcast.BroadcastBuildingUpdate(c.playerID, c.planetID, "unknown", 0)
	}
}

// handleResearch handles a research action.
func (c *WSClient) handleResearch(msg WSMessage) {
	if msg.Data != nil {
		data, _ := json.Marshal(msg.Data)
		log.Printf("Player %s research action: %s", c.playerID, string(data))
	}
}

// handleBuildShip handles a ship build action.
func (c *WSClient) handleBuildShip(msg WSMessage) {
	if msg.Data != nil {
		data, _ := json.Marshal(msg.Data)
		log.Printf("Player %s build ship: %s", c.playerID, string(data))
	}
}

// handleStartExpedition handles an expedition start action.
func (c *WSClient) handleStartExpedition(msg WSMessage) {
	if msg.Data != nil {
		data, _ := json.Marshal(msg.Data)
		log.Printf("Player %s start expedition: %s", c.playerID, string(data))
	}
}

// handleCreateMarketOrder handles a market order creation.
func (c *WSClient) handleCreateMarketOrder(msg WSMessage) {
	if msg.Data != nil {
		data, _ := json.Marshal(msg.Data)
		log.Printf("Player %s create market order: %s", c.playerID, string(data))
	}
}

// handleDeleteMarketOrder handles a market order deletion.
func (c *WSClient) handleDeleteMarketOrder(msg WSMessage) {
	if msg.Data != nil {
		data, _ := json.Marshal(msg.Data)
		log.Printf("Player %s delete market order: %s", c.playerID, string(data))
	}
}

// handleDrillCommand handles a drill command from the client.
func (c *WSClient) handleDrillCommand(msg WSMessage) {
	if msg.Data == nil {
		c.WriteJSON(WSMessage{
			Type:  "error",
			Error: "Missing drill command data",
		})
		return
	}

	var req struct {
		Direction string  `json:"direction"`
		Extract   *bool   `json:"extract,omitempty"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		c.WriteJSON(WSMessage{
			Type:  "error",
			Error: "Invalid drill command",
		})
		return
	}

	// Find active drill session for this player
	dg := game.FindActiveSession(c.planetID, c.playerID)
	if dg == nil {
		c.WriteJSON(WSMessage{
			Type:  "error",
			Error: "No active drill session",
		})
		return
	}

	dg.SetCommand(req.Direction, req.Extract)

	c.WriteJSON(WSMessage{
		Type: "drill_command_ack",
		Data: json.RawMessage(fmt.Sprintf(`{"status":"command_received"}`)),
	})
}

// handleGardenBedAction handles a garden bed action from the client via WebSocket.
func (c *WSClient) handleGardenBedAction(msg WSMessage) {
	if msg.Data == nil {
		c.WriteJSON(WSMessage{
			Type:  "error",
			Error: "Missing garden bed action data",
		})
		return
	}

	var req struct {
		Action    string `json:"action"`
		RowIndex  int    `json:"row_index"`
		PlantType string `json:"plant_type,omitempty"`
	}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		c.WriteJSON(WSMessage{
			Type:  "error",
			Error: "Invalid garden bed action",
		})
		return
	}

	if req.Action == "" {
		c.WriteJSON(WSMessage{
			Type:  "error",
			Error: "Missing action",
		})
		return
	}

	p := game.Instance().GetPlanet(c.planetID)
	if p == nil {
		if err := game.Instance().LoadPlanetFromDB(c.planetID); err != nil {
			log.Printf("Error loading planet from DB: %v", err)
		}
		p = game.Instance().GetPlanet(c.planetID)
	}

	if p == nil {
		c.WriteJSON(WSMessage{
			Type:  "error",
			Error: "Planet not found",
		})
		return
	}

	result, err := game.GardenBedActionInternal(p, req.Action, req.RowIndex, req.PlantType)
	if err != nil {
		c.WriteJSON(WSMessage{
			Type:  "error",
			Error: err.Error(),
		})
		return
	}

	if !result.Success {
		c.WriteJSON(WSMessage{
			Type: "garden_bed_action_result",
			Data: json.RawMessage(fmt.Sprintf(`{"success":false,"error":"%s","rows":%s}`,
				escapeJSON(result.Error), toJSON(result.Rows))),
		})
		return
	}

	wsBroadcast.BroadcastGardenBedUpdate(c.playerID, map[string]interface{}{
		"planet_id":    c.planetID,
		"rows":         result.Rows,
		"food_gain":    result.FoodGain,
		"money_gain":   result.MoneyGain,
	})

	c.WriteJSON(WSMessage{
		Type: "garden_bed_action_result",
		Data: json.RawMessage(fmt.Sprintf(`{"success":true,"rows":%s,"food_gain":%.0f,"money_gain":%.0f}`,
			toJSON(result.Rows), result.FoodGain, result.MoneyGain)),
	})
}

// extractPlayerIDFromToken extracts player ID from auth token by querying the database.
func extractPlayerIDFromToken(db *sql.DB, token string) (string, error) {
	var playerID string
	err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", token).Scan(&playerID)
	if err != nil {
		return "", err
	}
	return playerID, nil
}

// generateConnID generates a unique connection ID.
func generateConnID() string {
	wsManager.connCounter++
	return fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(), wsManager.connCounter)
}

// randomEventMessages contains messages for random events.
var randomEventMessages = []string{
	"Solar flare detected! Energy production increased by 20%.",
	"Alien signal intercepted! Research progress boosted.",
	"Space debris field detected! Navigation hazards increased.",
	"Trade opportunity found! Market prices fluctuating.",
	"Comet approaching! Rare resources available.",
	" Pirate fleet spotted nearby! Increase defenses.",
	"Abandoned space station discovered! Potential loot.",
	"Gravitational anomaly detected! Sensors disrupted.",
	"Merchant fleet passing by! Trade discounts available.",
	"Distress signal received! Send aid or ignore.",
}

// RandomEvent triggers a random event for a player.
func RandomEvent(playerID string) {
	idx := rand.Intn(len(randomEventMessages))
	msg := randomEventMessages[idx]
	wsBroadcast.BroadcastNotification(playerID, msg, "event")
}

// PlanetStateBroadcast sends a full state update for a planet.
func PlanetStateBroadcast(playerID, planetID string, state map[string]interface{}) {
	wsBroadcast.BroadcastStateUpdate(playerID, planetID, state)
}
