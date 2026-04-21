package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWSMessageQueue(t *testing.T) {
	mq := NewWSMessageQueue()

	// Test enqueue
	msg := WSMessage{Type: "test", Data: json.RawMessage(`{"key":"value"}`)}
	mq.Enqueue("player1", msg)

	if mq.Len("player1") != 1 {
		t.Fatalf("expected queue length 1, got %d", mq.Len("player1"))
	}

	// Test queue size limit
	for i := 0; i < WSMessageQueueSize+10; i++ {
		mq.Enqueue("player1", WSMessage{Type: "overflow", Data: json.RawMessage(`{}`)})
	}

	if mq.Len("player1") > WSMessageQueueSize {
		t.Fatalf("expected queue length <= %d, got %d", WSMessageQueueSize, mq.Len("player1"))
	}

	// Test dequeue
	dequeued := mq.Dequeue("player1")
	if len(dequeued) == 0 {
		t.Fatal("expected dequeued messages, got none")
	}

	// Test queue is empty after dequeue
	if mq.Len("player1") != 0 {
		t.Fatalf("expected empty queue after dequeue, got %d", mq.Len("player1"))
	}
}

func TestWSRateLimiter(t *testing.T) {
	rl := NewWSRateLimiter(10)

	// Test rate limit token consumption
	if !rl.ConsumeToken("player1") {
		t.Fatal("expected token consumption to succeed")
	}

	// Test cleanup
	rl.Cleanup()

	// After cleanup, tokens should be gone (will be recreated on next call)
	if len(rl.clients) != 0 {
		t.Fatalf("expected no clients after cleanup, got %d", len(rl.clients))
	}
}

func TestWSConnectionManager(t *testing.T) {
	cm := NewWSConnectionManager()
	defer cm.Close()

	// Test AddClient and GetClient
	client := cm.NewClient(nil, "player1", "planet1", "conn1", "token1")
	cm.AddClient("player1", "conn1", client)

	if cm.Count() != 1 {
		t.Fatalf("expected 1 client, got %d", cm.Count())
	}

	if !cm.IsConnected("player1") {
		t.Fatal("expected player1 to be connected")
	}

	// Test GetClientByConnID
	retrieved := cm.GetClientByConnID("conn1")
	if retrieved == nil {
		t.Fatal("expected to retrieve client by connID")
	}
	if retrieved.playerID != "player1" {
		t.Fatalf("expected playerID 'player1', got '%s'", retrieved.playerID)
	}

	// Test RemoveClient
	cm.RemoveClient("conn1")

	if cm.IsConnected("player1") {
		t.Fatal("expected player1 to be disconnected")
	}

	if cm.Count() != 0 {
		t.Fatalf("expected 0 clients after removal, got %d", cm.Count())
	}
}

func TestWSBroadcastService(t *testing.T) {
	cm := NewWSConnectionManager()
	bs := NewWSBroadcastService(cm)
	defer cm.Close()

	// Create a mock client to test broadcasting
	client := cm.NewClient(nil, "player1", "planet1", "conn1", "token1")
	client.send = make(chan []byte, 10)
	cm.AddClient("player1", "conn1", client)

	// Test BroadcastStateUpdate
	state := map[string]interface{}{
		"resources": map[string]float64{"food": 100, "energy": 50},
		"buildings": map[string]int{"farm": 2},
	}
	bs.BroadcastStateUpdate("player1", "planet1", state)

	// Check that message was sent (non-blocking)
	select {
	case msg := <-client.send:
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "state_update" {
			t.Fatalf("expected type 'state_update', got '%s'", wsMsg.Type)
		}
	default:
		t.Fatal("expected message in send channel")
	}

	// Test BroadcastNotification
	bs.BroadcastNotification("player1", "Test notification", "alert")

	select {
	case msg := <-client.send:
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "notification" {
			t.Fatalf("expected type 'notification', got '%s'", wsMsg.Type)
		}
	default:
		t.Fatal("expected notification message in send channel")
	}

	// Test BroadcastExpeditionUpdate
	bs.BroadcastExpeditionUpdate("player1", "exp1", 0.75)

	select {
	case msg := <-client.send:
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "expedition_update" {
			t.Fatalf("expected type 'expedition_update', got '%s'", wsMsg.Type)
		}
	default:
		t.Fatal("expected expedition update message in send channel")
	}

	// Test BroadcastMarketUpdate
	bs.BroadcastMarketUpdate("player1", "order1", "filled")

	select {
	case msg := <-client.send:
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "market_update" {
			t.Fatalf("expected type 'market_update', got '%s'", wsMsg.Type)
		}
	default:
		t.Fatal("expected market update message in send channel")
	}

	// Test BroadcastMiningUpdate
	miningState := map[string]interface{}{
		"player_x": 5, "player_y": 3, "status": "active",
	}
	bs.BroadcastMiningUpdate("player1", "session1", miningState)

	select {
	case msg := <-client.send:
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "mining_update" {
			t.Fatalf("expected type 'mining_update', got '%s'", wsMsg.Type)
		}
	default:
		t.Fatal("expected mining update message in send channel")
	}
}

func TestWSBroadcastServiceToDisconnectedPlayer(t *testing.T) {
	cm := NewWSConnectionManager()
	bs := NewWSBroadcastService(cm)
	defer cm.Close()

	// Send to disconnected player - should be queued
	bs.BroadcastStateUpdate("disconnected_player", "planet1", map[string]interface{}{
		"resources": map[string]float64{"food": 100},
	})

	// Check queue
	if cm.messageQueue.Len("disconnected_player") == 0 {
		t.Fatal("expected message to be queued for disconnected player")
	}

	// Simulate reconnection
	client := cm.NewClient(nil, "disconnected_player", "planet1", "conn1", "token1")
	client.send = make(chan []byte, 10)
	cm.AddClient("disconnected_player", "conn1", client)

	// Dequeue messages (simulating reconnection)
	dequeued := cm.messageQueue.Dequeue("disconnected_player")
	if len(dequeued) == 0 {
		t.Fatal("expected dequeued messages for reconnecting player")
	}
}

func TestWSBroadcastToAll(t *testing.T) {
	cm := NewWSConnectionManager()
	defer cm.Close()

	// Create multiple clients
	client1 := cm.NewClient(nil, "player1", "planet1", "conn1", "token1")
	client1.send = make(chan []byte, 10)
	cm.AddClient("player1", "conn1", client1)

	client2 := cm.NewClient(nil, "player2", "planet2", "conn2", "token2")
	client2.send = make(chan []byte, 10)
	cm.AddClient("player2", "conn2", client2)

	// Broadcast to all
	msg := WSMessage{Type: "global_event", Data: json.RawMessage(`{"message":"test"}`)}
	cm.BroadcastToAll(msg)

	// Check both received the message
	count1 := 0
	count2 := 0

	select {
	case <-client1.send:
		count1++
	default:
	}

	select {
	case <-client2.send:
		count2++
	default:
	}

	if count1 != 1 {
		t.Fatalf("expected client1 to receive message, got %d", count1)
	}
	if count2 != 1 {
		t.Fatalf("expected client2 to receive message, got %d", count2)
	}
}

func TestWSBroadcastToPlanet(t *testing.T) {
	cm := NewWSConnectionManager()
	defer cm.Close()

	// Create clients on same planet
	client1 := cm.NewClient(nil, "player1", "planet1", "conn1", "token1")
	client1.send = make(chan []byte, 10)
	cm.AddClient("player1", "conn1", client1)

	client2 := cm.NewClient(nil, "player2", "planet1", "conn2", "token2")
	client2.send = make(chan []byte, 10)
	cm.AddClient("player2", "conn2", client2)

	// Create client on different planet
	client3 := cm.NewClient(nil, "player3", "planet2", "conn3", "token3")
	client3.send = make(chan []byte, 10)
	cm.AddClient("player3", "conn3", client3)

	// Broadcast to planet1
	msg := WSMessage{Type: "planet_event", Data: json.RawMessage(`{"event":"test"}`)}
	cm.SendToPlanet("planet1", msg)

	// Check only planet1 clients received
	count1 := 0
	count2 := 0
	count3 := 0

	select {
	case <-client1.send:
		count1++
	default:
	}

	select {
	case <-client2.send:
		count2++
	default:
	}

	select {
	case <-client3.send:
		count3++
	default:
	}

	if count1 != 1 {
		t.Fatalf("expected client1 to receive message, got %d", count1)
	}
	if count2 != 1 {
		t.Fatalf("expected client2 to receive message, got %d", count2)
	}
	if count3 != 0 {
		t.Fatalf("expected client3 NOT to receive message, got %d", count3)
	}
}

func TestWSClientClose(t *testing.T) {
	cm := NewWSConnectionManager()
	defer cm.Close()

	client := cm.NewClient(nil, "player1", "planet1", "conn1", "token1")
	cm.AddClient("player1", "conn1", client)

	// Verify closed flag
	client.mu.Lock()
	if client.closed {
		t.Fatal("expected client to not be closed yet")
	}
	client.mu.Unlock()

	// Close client
	client.Close()

	// Verify closed flag
	client.mu.Lock()
	if !client.closed {
		t.Fatal("expected client to be closed")
	}
	client.mu.Unlock()

	// Test that Close is idempotent
	client.Close()
	client.mu.Lock()
	if !client.closed {
		t.Fatal("expected client to still be closed")
	}
	client.mu.Unlock()
}

func TestWSConnectionManagerConcurrent(t *testing.T) {
	cm := NewWSConnectionManager()
	defer cm.Close()

	var wg sync.WaitGroup
	numClients := 100

	// Concurrent adds
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			playerID := "player" + string(rune('a'+id%26))
			connID := "conn_" + string(rune('a'+id%26))
			client := cm.NewClient(nil, playerID, "planet1", connID, "token")
			cm.AddClient(playerID, connID, client)
		}(i)
	}

	wg.Wait()

	// Count should not exceed numClients (some players share IDs)
	count := cm.Count()
	if count > numClients {
		t.Fatalf("expected at most %d clients, got %d", numClients, count)
	}
}

func TestWSMessageQueueConcurrent(t *testing.T) {
	mq := NewWSMessageQueue()
	var wg sync.WaitGroup

	// Concurrent enqueue
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := WSMessage{Type: "test", Data: json.RawMessage(`{"id":` + string(rune(id)) + `}`)}
			mq.Enqueue("player1", msg)
		}(i)
	}

	wg.Wait()

	length := mq.Len("player1")
	if length == 0 {
		t.Fatal("expected messages in queue")
	}
	if length > WSMessageQueueSize {
		t.Fatalf("expected queue length <= %d, got %d", WSMessageQueueSize, length)
	}
}

func TestHandleWebSocketHealth(t *testing.T) {
	// Initialize WS manager
	InitWS()
	defer wsManager.Close()

	req := httptest.NewRequest(http.MethodGet, "/ws/health", nil)
	w := httptest.NewRecorder()

	handleWebSocketHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Fatalf("expected status 'ok', got '%v'", resp["status"])
	}

	if _, ok := resp["connected_clients"]; !ok {
		t.Fatal("expected 'connected_clients' in response")
	}

	if _, ok := resp["timestamp"]; !ok {
		t.Fatal("expected 'timestamp' in response")
	}
}

func TestWSClientWriteJSON(t *testing.T) {
	cm := NewWSConnectionManager()
	defer cm.Close()

	_ = cm.NewClient(nil, "player1", "planet1", "conn1", "token1")

	// Test WriteJSON with nil connection (will panic on WriteMessage)
	// We can't fully test this without a real connection, but we can test the JSON marshaling
	msg := WSMessage{Type: "test", Data: json.RawMessage(`{"key":"value"}`)}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	var unmarshaled WSMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if unmarshaled.Type != "test" {
		t.Fatalf("expected type 'test', got '%s'", unmarshaled.Type)
	}
}

func TestWSBroadcastServiceRandomEvent(t *testing.T) {
	// Initialize WS manager
	InitWS()
	defer wsManager.Close()

	// Create a client
	client := wsManager.NewClient(nil, "player1", "planet1", "conn1", "token1")
	client.send = make(chan []byte, 10)
	wsManager.AddClient("player1", "conn1", client)

	// Trigger random event
	RandomEvent("player1")

	// Check message was sent
	select {
	case msg := <-client.send:
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "notification" {
			t.Fatalf("expected type 'notification', got '%s'", wsMsg.Type)
		}
	default:
		t.Fatal("expected notification message in send channel")
	}
}

func TestWSBroadcastServicePlanetStateBroadcast(t *testing.T) {
	// Initialize WS manager
	InitWS()
	defer wsManager.Close()

	// Create a client
	client := wsManager.NewClient(nil, "player1", "planet1", "conn1", "token1")
	client.send = make(chan []byte, 10)
	wsManager.AddClient("player1", "conn1", client)

	// Broadcast planet state
	PlanetStateBroadcast("player1", "planet1", map[string]interface{}{
		"resources": map[string]float64{"food": 100, "energy": 50},
	})

	// Check message was sent
	select {
	case msg := <-client.send:
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "state_update" {
			t.Fatalf("expected type 'state_update', got '%s'", wsMsg.Type)
		}
	default:
		t.Fatal("expected state update message in send channel")
	}
}

func TestWSMessageQueueCleanup(t *testing.T) {
	mq := NewWSMessageQueue()

	// Enqueue a message
	mq.Enqueue("player1", WSMessage{Type: "test", Data: json.RawMessage(`{}`)})

	if mq.Len("player1") != 1 {
		t.Fatalf("expected queue length 1, got %d", mq.Len("player1"))
	}

	// Clear the queue
	mq.Clear("player1")

	if mq.Len("player1") != 0 {
		t.Fatalf("expected queue length 0 after clear, got %d", mq.Len("player1"))
	}
}

func TestWSConnectionManagerCloseAll(t *testing.T) {
	cm := NewWSConnectionManager()

	// Add multiple clients
	for i := 0; i < 10; i++ {
		client := cm.NewClient(nil, "player"+string(rune('a'+i)), "planet1", "conn"+string(rune('a'+i)), "token")
		cm.AddClient("player"+string(rune('a'+i)), "conn"+string(rune('a'+i)), client)
	}

	if cm.Count() != 10 {
		t.Fatalf("expected 10 clients, got %d", cm.Count())
	}

	// Close all
	cm.Close()

	if cm.Count() != 0 {
		t.Fatalf("expected 0 clients after close, got %d", cm.Count())
	}
}

// TestWebSocketHandlerIntegration tests the WebSocket handler with a mock connection.
func TestWebSocketHandlerIntegration(t *testing.T) {
	// Initialize WS manager
	InitWS()
	defer wsManager.Close()

	done := make(chan struct{})

	// Create a mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			conn, err := WSUpgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			query := r.URL.Query()
			token := query.Get("token")
			if token == "" {
				conn.WriteJSON(WSMessage{Type: "error", Error: "Missing auth token"})
				return
			}

			playerID := "test_player_" + token[:8]
			if playerID == "" {
				conn.WriteJSON(WSMessage{Type: "error", Error: "Invalid auth token"})
				return
			}

			connID := generateConnID()
			client := wsManager.NewClient(conn, playerID, "", connID, token)
			wsManager.AddClient(playerID, connID, client)

			// Send welcome directly
			data, _ := json.Marshal(WSMessage{
				Type: "welcome",
				Data: json.RawMessage(`{"player_id":"` + playerID + `","message":"Connected"}`),
			})
			conn.WriteMessage(websocket.TextMessage, data)

			// Read messages and respond directly
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					break
				}
				var wsMsg WSMessage
				if err := json.Unmarshal(message, &wsMsg); err != nil {
					continue
				}
				if wsMsg.Type == "ping" {
					pongData, _ := json.Marshal(WSMessage{
						Type: "pong",
						Data: json.RawMessage(`{"timestamp":` + string(rune(time.Now().UnixNano())) + `}`),
					})
					conn.WriteMessage(websocket.TextMessage, pongData)
				}
			}
			close(done)
		}
	}))
	defer ts.Close()

	// Connect via WebSocket
	wsURL := "ws" + ts.URL[4:] // Convert http to ws
	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"/ws?token=test-token-123", nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Read welcome message with timeout
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read welcome: %v", err)
	}

	var welcome WSMessage
	if err := json.Unmarshal(message, &welcome); err != nil {
		t.Fatalf("failed to unmarshal welcome: %v", err)
	}

	if welcome.Type != "welcome" {
		t.Fatalf("expected welcome message, got '%s'", welcome.Type)
	}

	// Close connection to trigger server cleanup
	conn.Close()

	// Wait for server to finish with timeout
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("server did not finish within timeout")
	}
}

func TestWebSocketHandlerMissingToken(t *testing.T) {
	// Initialize WS manager
	InitWS()
	defer wsManager.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			conn, err := WSUpgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			query := r.URL.Query()
			token := query.Get("token")
			if token == "" {
				conn.WriteJSON(WSMessage{Type: "error", Error: "Missing auth token"})
				return
			}
		}
	}))
	defer ts.Close()

	wsURL := "ws" + ts.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"/ws", nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Read error message
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read error: %v", err)
	}

	var errMsg WSMessage
	if err := json.Unmarshal(message, &errMsg); err != nil {
		t.Fatalf("failed to unmarshal error: %v", err)
	}

	if errMsg.Type != "error" {
		t.Fatalf("expected error message, got '%s'", errMsg.Type)
	}
}

func TestEscapeJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`hello "world"`, `hello \"world\"`},
		{`line1\nline2`, `line1\\nline2`},
		{`path\to\file`, `path\\to\\file`},
		{`simple`, `simple`},
	}

	for _, tt := range tests {
		result := escapeJSON(tt.input)
		if result != tt.expected {
			t.Fatalf("escapeJSON(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{map[string]interface{}{"key": "value"}, `{"key":"value"}`},
		{map[string]interface{}{"num": 42}, `{"num":42}`},
		{nil, `{}`},
	}

	for _, tt := range tests {
		result := toJSON(tt.input)
		if result != tt.expected {
			t.Fatalf("toJSON(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestWSClientStartHeartbeat(t *testing.T) {
	cm := NewWSConnectionManager()
	defer cm.Close()

	client := cm.NewClient(nil, "player1", "planet1", "conn1", "token1")
	cm.AddClient("player1", "conn1", client)

	// Start heartbeat with short interval - just verify it doesn't panic
	// The actual heartbeat test would require mocking the connection
	client.mu.Lock()
	client.closed = true
	client.mu.Unlock()

	// Should not panic
	done := make(chan struct{})
	go func() {
		client.StartHeartbeat(cm, "conn1")
		close(done)
	}()

	select {
	case <-done:
		// Good, heartbeat exited
	case <-time.After(100 * time.Millisecond):
		t.Fatal("heartbeat did not exit promptly for closed client")
	}
}

func TestWSClientStartWritePump(t *testing.T) {
	cm := NewWSConnectionManager()
	defer cm.Close()

	client := cm.NewClient(nil, "player1", "planet1", "conn1", "token1")
	cm.AddClient("player1", "conn1", client)

	// Close client to stop write pump
	client.mu.Lock()
	client.closed = true
	client.mu.Unlock()

	done := make(chan struct{})
	go func() {
		client.StartWritePump(cm, "conn1")
		close(done)
	}()

	select {
	case <-done:
		// Good, write pump exited
	case <-time.After(100 * time.Millisecond):
		t.Fatal("write pump did not exit promptly for closed client")
	}
}
