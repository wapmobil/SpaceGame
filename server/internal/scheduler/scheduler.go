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
	go s.marketRefresh()
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

func (s *Scheduler) marketRefresh() {
	ticker := time.NewTicker(20 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		marketLevel := s.game.GetTotalMarketLevel()
		s.game.Marketplace.GenerateNPCOrders(marketLevel)
		log.Printf("NPC traders refreshed (market level: %d)", marketLevel)
	}
}
