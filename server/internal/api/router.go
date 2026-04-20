package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// SetupRouter creates and configures the chi router with all routes.
func SetupRouter(db *sql.DB) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/api/register", handleRegister(db))
	r.Get("/api/planets", handleListPlanets(db))
	r.Post("/api/planets", handleCreatePlanet(db))
	r.Get("/ws", handleWebSocket(nil))
	r.Get("/health", handleHealth)
	r.Get("/ws/health", handleWebSocketHealth)

	r.Get("/api/planets/{id}/research", handleGetResearch(db))
	r.Post("/api/planets/{id}/research/start", handleStartResearch(db))

	r.Get("/api/planets/{id}/fleet", handleGetFleet(db))
	r.Post("/api/planets/{id}/ship/build", handleBuildShip(db))
	r.Get("/api/planets/{id}/ships/available", handleGetAvailableShips(db))
	r.Get("/api/planets/{id}/battles", handleGetBattles(db))

	// Expeditions
	r.Post("/api/planets/{id}/expeditions", handleCreateExpedition(db))
	r.Get("/api/planets/{id}/expeditions", handleGetExpeditions(db))
	r.Post("/api/expeditions/{id}/action", handleExpeditionAction(db))

	// Marketplace
	r.Post("/api/planets/{id}/market/orders", handleCreateMarketOrder(db))
	r.Get("/api/planets/{id}/market/orders", handleGetMyOrders(db))
	r.Get("/api/market", handleGetGlobalMarket(db))
	r.Delete("/api/market/orders/{id}", handleDeleteMarketOrder(db))
	r.Get("/api/market/traders", handleGetNPCTraders(db))
	r.Post("/api/market/match", handleMatchOrders(db))

	// Mining mini-game
	r.Post("/api/planets/{id}/mining/start", handleStartMining(db))
	r.Post("/api/planets/{id}/mining/move", handleMiningMove(db))
	r.Get("/api/planets/{id}/mining", handleGetMining(db))

	// Ratings / Leaderboards
	r.Get("/api/ratings", handleGetRatings(db))

	// Statistics
	r.Get("/api/stats", handleGetStats(db))
	r.Get("/api/planets/{id}/stats", handleGetStats(db))

	// Events
	r.Get("/api/events", handleGetEventHistory(db))
	r.Post("/api/planets/{id}/events/resolve", handleResolveEvent(db))

	return r
}

// NewServer creates a new HTTP server with the configured router.
func NewServer(db *sql.DB) *http.Server {
	r := SetupRouter(db)

	return &http.Server{
		Addr:         ":" + getEnv("PORT", "8080"),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}
