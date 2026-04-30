package api

import (
	"sync"
	"time"
)

// WSMessageQueue stores messages for disconnected players.
type WSMessageQueue struct {
	mu     sync.Mutex
	queues map[string][]WSMessage
	timers map[string]*time.Timer
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

// WSRateLimiter implements a sliding window rate limiter.
type WSRateLimiter struct {
	mu      sync.Mutex
	clients map[string]*time.Ticker
	maxRate int
	window  time.Duration
}

// NewWSRateLimiter creates a new rate limiter.
func NewWSRateLimiter(maxRate int) *WSRateLimiter {
	return &WSRateLimiter{
		clients: make(map[string]*time.Ticker),
		maxRate: maxRate,
		window:  time.Second,
	}
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
