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
	r.Get("/ws", handleWebSocket(db))
	r.Get("/health", handleHealth)

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
