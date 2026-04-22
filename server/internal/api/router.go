package api

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"strings"
	"time"

	"spacegame/internal/web"

	"github.com/go-chi/chi/v5"
)

// SetupRouter creates and configures the chi router with all routes.
func SetupRouter(db *sql.DB) *chi.Mux {
	r := chi.NewRouter()

	// Serve Flutter web app (SPA with static asset fallback)
	webSub := web.Sub()
	r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip API, WS, health, and static asset paths
			if strings.HasPrefix(r.URL.Path, "/api/") ||
				strings.HasPrefix(r.URL.Path, "/ws") ||
				r.URL.Path == "/health" ||
				r.URL.Path == "/ws/health" {
				h.ServeHTTP(w, r)
				return
			}

			path := r.URL.Path
			if path == "/" || path == "" {
				path = "/index.html"
			}

			// Try to serve as static asset from embedded FS
			testPath := path[1:] // strip leading /
			f, err := webSub.Open(testPath)
			if err == nil {
				defer f.Close()
				info, _ := f.Stat()
				if info.IsDir() {
					path = "/index.html"
					testPath = path[1:]
					f, _ = webSub.Open(testPath)
				}
				data, _ := io.ReadAll(f)
				http.ServeContent(w, r, path, info.ModTime(), bytes.NewReader(data))
				return
			}

			// Fallback to index.html for SPA routing
			f, err = webSub.Open("index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer f.Close()
			info, _ := f.Stat()
			data, _ := io.ReadAll(f)
			http.ServeContent(w, r, "/index.html", info.ModTime(), bytes.NewReader(data))
		})
	})

	r.Post("/api/register", handleRegister(db))
	r.Post("/api/login", handleLogin(db))
	r.Get("/api/planets", handleListPlanets(db))
	r.Post("/api/planets", handleCreatePlanet(db))
	r.Get("/ws", handleWebSocket(db))
	r.Get("/health", handleHealth)
	r.Get("/ws/health", handleWebSocketHealth)

	r.Get("/api/planets/{id}", handleGetPlanet(db))
	r.Get("/api/planets/{id}/buildings", handleGetBuildings(db))
	r.Post("/api/planets/{id}/buildings", handleBuildBuilding(db))
	r.Post("/api/planets/{id}/buildings/{buildingType}/confirm", handleConfirmBuilding(db))
	r.Post("/api/planets/{id}/buildings/{buildingType}/toggle", handleToggleBuilding(db))
	r.Get("/api/planets/{id}/build-details", handleGetBuildDetails(db))
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
	r.Post("/api/planets/{id}/sell-food", handleSellFood(db))

	// Mining mini-game
	r.Post("/api/planets/{id}/mining/start", handleStartMining(db))
	r.Post("/api/planets/{id}/mining/move", handleMiningMove(db))
	r.Get("/api/planets/{id}/mining", handleGetMining(db))

	// Ratings / Leaderboards
	r.Get("/api/ratings", handleGetRatings(db))

	// Statistics
	r.Get("/api/planets/{id}/stats", handleGetStats(db))

	// Events
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
