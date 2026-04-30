package api

import (
	"encoding/json"
	"time"
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

// WSMessage represents a WebSocket message.
type WSMessage struct {
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}
