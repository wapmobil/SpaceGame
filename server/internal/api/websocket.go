package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleWebSocket upgrades the connection to WebSocket and handles real-time messages.
func handleWebSocket(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		query := r.URL.Query()
		authToken := query.Get("token")
		if authToken == "" {
			conn.WriteJSON(map[string]string{"error": "Missing auth token"})
			return
		}

		var playerID string
		err = db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
		if err != nil {
			conn.WriteJSON(map[string]string{"error": "Invalid auth token"})
			return
		}

		log.Printf("WebSocket connected: player %s", playerID)

		conn.WriteJSON(map[string]interface{}{
			"type":      "welcome",
			"player_id": playerID,
			"message":   "Connected to SpaceGame WebSocket",
		})

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error for player %s: %v", playerID, err)
				break
			}

			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				conn.WriteJSON(map[string]string{"error": "Invalid JSON"})
				continue
			}

			msgType, _ := msg["type"].(string)
			log.Printf("WebSocket message from %s: %s", playerID, msgType)

			switch msgType {
			case "ping":
				conn.WriteJSON(map[string]string{"type": "pong"})
			case "subscribe":
				planetID, _ := msg["planet_id"].(string)
				if planetID != "" {
					conn.WriteJSON(map[string]string{
						"type":      "subscribed",
						"planet_id": planetID,
					})
				}
			default:
				conn.WriteJSON(map[string]interface{}{
					"type":    "ack",
					"message": "Received",
				})
			}
		}
	}
}
