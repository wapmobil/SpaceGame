package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"spacegame/internal/api"
	"spacegame/internal/db"
	"spacegame/internal/game"
	"spacegame/internal/scheduler"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := db.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "spacegame"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	log.Println("Connecting to database...")
	database, err := db.New(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("Database connected and migrated")

	gameEngine := game.New()
	gameEngine.SetDB(database)
	game.SetInstance(gameEngine)

	// Load existing planets from database
	if err := gameEngine.LoadPlanetsFromDB(); err != nil {
		log.Printf("Warning: failed to load planets from DB: %v", err)
	}

	sched := scheduler.New(gameEngine)
	sched.Start()

	httpServer := api.NewServer(database.DB)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		_ = httpServer.Shutdown(context.Background())
	}()

	log.Printf("Starting HTTP server on %s", httpServer.Addr)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	log.Println("SpaceGame server is running")
	<-ctx.Done()
	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
