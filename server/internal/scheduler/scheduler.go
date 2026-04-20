package scheduler

import (
	"log"
	"time"

	"spacegame/internal/game"
)

// Scheduler manages periodic game tasks.
type Scheduler struct {
	game *game.Game
}

// New creates a new Scheduler.
func New(g *game.Game) *Scheduler {
	return &Scheduler{game: g}
}

// Start begins the scheduled tasks.
func (s *Scheduler) Start() {
	go s.gameTick()
	go s.ratingsUpdate()
	go s.randomEventsTick()
	log.Println("Scheduler started")
}

func (s *Scheduler) gameTick() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.game.Tick()
	}
}

func (s *Scheduler) ratingsUpdate() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		log.Println("Computing ratings...")
		s.game.ComputeRatings()
	}
}

func (s *Scheduler) randomEventsTick() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.game.TriggerRandomEvents()
	}
}
