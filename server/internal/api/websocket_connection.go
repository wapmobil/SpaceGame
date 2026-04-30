package api

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"spacegame/internal/game"
)

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

// WSConnectionManager manages all WebSocket connections.
type WSConnectionManager struct {
	mu           sync.RWMutex
	connections  map[string]*WSClient // playerID -> client
	planetMap    map[string]string    // connID -> playerID
	rateLimiter  *WSRateLimiter
	messageQueue *WSMessageQueue
	clients      map[string]*WSClient // connID -> client
	connCounter  atomic.Int64
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
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.mu.Unlock()
					return
				}
				c.mu.Unlock()
			}
		}
	}()
}
