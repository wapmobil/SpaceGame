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

	// Non-authenticated routes
	r.Post("/api/register", handleRegister(db))
	r.Post("/api/login", handleLogin(db))
	r.Get("/api/planets", handleListPlanets(db))
	r.Post("/api/planets", handleCreatePlanet(db))
	r.Get("/ws", handleWebSocket(db))
	r.Get("/health", handleHealth)
	r.Get("/ws/health", handleWebSocketHealth)
	r.Get("/api/market", handleGetGlobalMarket(db))
	r.Get("/api/ratings", handleGetRatings(db))

	// Planet-authenticated routes (auth + planet ownership)
	r.Route("/api/planets", func(rr chi.Router) {
		rr.Use(requireAuth(db), requirePlanetOwnership(db))
		rr.Get("/{id}", handleGetPlanet(db))
		rr.Get("/{id}/buildings", handleGetBuildings(db))
		rr.Post("/{id}/buildings", handleBuildBuilding(db))
		rr.Post("/{id}/buildings/{buildingType}/confirm", handleConfirmBuilding(db))
		rr.Post("/{id}/buildings/{buildingType}/toggle", handleToggleBuilding(db))
		rr.Get("/{id}/build-details", handleGetBuildDetails(db))
		rr.Get("/{id}/research", handleGetResearch(db))
		rr.Post("/{id}/research/start", handleStartResearch(db))
		rr.Get("/{id}/fleet", handleGetFleet(db))
		rr.Post("/{id}/ship/build", handleBuildShip(db))
		rr.Get("/{id}/ships/available", handleGetAvailableShips(db))
		rr.Post("/{id}/expeditions", handleCreateExpedition(db))
		rr.Get("/{id}/expeditions", handleGetExpeditions(db))
		rr.Post("/{id}/market/orders", handleCreateMarketOrder(db))
		rr.Get("/{id}/market/orders", handleGetMyOrders(db))
		rr.Post("/{id}/sell-food", handleSellFood(db))
		rr.Post("/{id}/drill/start", handleStartDrill(db))
		rr.Post("/{id}/drill", handleDrillCommand(db))
		rr.Get("/{id}/drill/chunk", handleDrillChunk(db))
		rr.Post("/{id}/drill/complete", handleCompleteDrill(db))
		rr.Post("/{id}/drill/destroy", handleDestroyDrill(db))
		rr.Post("/{id}/drill/cleanup", handleCleanupDrill(db))
		rr.Get("/{id}/stats", handleGetStats(db))
		rr.Post("/{id}/events/resolve", handleResolveEvent(db))
		rr.Get("/{id}/farm", handleGetFarm(db))
		rr.Post("/{id}/farm/action", handleFarmAction(db))
	})

	// Auth-only routes (auth but no planet ownership check)
	r.Route("/api", func(rr chi.Router) {
		rr.Use(requireAuth(db))
		rr.Post("/expeditions/{id}/action", handleExpeditionAction(db))
		rr.Delete("/market/orders/{id}", handleDeleteMarketOrder(db))
	})

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
