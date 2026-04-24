package api

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"spacegame/internal/game"

	"github.com/go-chi/chi/v5"
)

type contextKey string

const playerIDKey contextKey = "playerID"
const planetIDKey contextKey = "planetID"

// AuthPlayer represents an authenticated player with their ID.
type AuthPlayer struct {
	ID string
}

// AuthPlayerFromContext extracts the authenticated player from the request context.
// Returns nil if the request is not authenticated.
func AuthPlayerFromContext(r *http.Request) *AuthPlayer {
	playerID, ok := r.Context().Value(playerIDKey).(string)
	if !ok || playerID == "" {
		return nil
	}
	return &AuthPlayer{ID: playerID}
}

// requireAuth is a chi middleware that validates the auth token and attaches playerID to context.
func requireAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authToken := r.Header.Get("X-Auth-Token")
			if authToken == "" {
				JSON(w, http.StatusUnauthorized, map[string]string{"error": "Missing auth token"})
				return
			}

			var playerID string
			err := db.QueryRow("SELECT id FROM players WHERE auth_token = $1", authToken).Scan(&playerID)
			if err != nil {
				JSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid auth token"})
				return
			}

			ctx := context.WithValue(r.Context(), playerIDKey, playerID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// requirePlanetOwnership is a chi middleware that validates the planet belongs to the authenticated player.
func requirePlanetOwnership(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			playerID := r.Context().Value(playerIDKey).(string)
			planetID := chi.URLParam(r, "id")
			if planetID == "" {
				JSON(w, http.StatusBadRequest, map[string]string{"error": "Missing planet id"})
				return
			}

			var ownerID string
			err := db.QueryRow("SELECT player_id FROM planets WHERE id = $1", planetID).Scan(&ownerID)
			if err != nil {
				JSON(w, http.StatusNotFound, map[string]string{"error": "Planet not found"})
				return
			}
			if ownerID != playerID {
				JSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
				return
			}

			ctx := context.WithValue(r.Context(), planetIDKey, planetID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PlanetIDFromContext extracts the planet ID from the request context.
func PlanetIDFromContext(r *http.Request) string {
	planetID, _ := r.Context().Value(planetIDKey).(string)
	return planetID
}

// ensurePlanetLoaded loads a planet from DB if not already in memory.
// Returns the planet or nil if loading failed.
func ensurePlanetLoaded(planetID string) *game.Planet {
	p := game.Instance().GetPlanet(planetID)
	if p == nil {
		if err := game.Instance().LoadPlanetFromDB(planetID); err != nil {
			log.Printf("Error loading planet from DB: %v", err)
		}
		p = game.Instance().GetPlanet(planetID)
	}
	return p
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
