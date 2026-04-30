package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"spacegame/internal/game"
)

// WSUpgrader is the WebSocket upgrader configuration.
var WSUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

// handleWebSocketHealth returns the WebSocket health status.
func handleWebSocketHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":            "ok",
		"connected_clients": wsManager.Count(),
		"timestamp":         time.Now().Unix(),
	})
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
				Type:  "error",
				Error: "Missing auth token",
			})
			return
		}

		// Validate auth token against database
		var playerID string
		err = db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			conn.WriteJSON(WSMessage{
				Type:  "error",
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

	c.conn.SetReadDeadline(time.Now().Add(WSPingInterval * 2))

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error for player %s: %v", c.playerID, err)
			}
			break
		}
		c.conn.SetReadDeadline(time.Now().Add(WSPingInterval * 2))

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
		"planet_id":  c.planetID,
		"rows":       result.Rows,
		"food_gain":  result.FoodGain,
		"money_gain": result.MoneyGain,
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
	wsManager.connCounter.Add(1)
	return fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(), wsManager.connCounter.Load())
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
