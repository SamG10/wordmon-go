package api

import (
	"context"
	"fmt"
	"time"

	"github.com/SamG1008/wordmon-go/internal/core"
	"github.com/SamG1008/wordmon-go/internal/store"
)

// SpawnerService gère le spawning automatique de WordMon
type SpawnerService struct {
	store    *store.MemoryStore
	interval time.Duration
	timeout  time.Duration
}

// NewSpawnerService crée un nouveau service de spawning
func NewSpawnerService(store *store.MemoryStore, interval, timeout time.Duration) *SpawnerService {
	return &SpawnerService{
		store:    store,
		interval: interval,
		timeout:  timeout,
	}
}

// Start démarre le spawner en arrière-plan
func (s *SpawnerService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	fmt.Printf("[spawn] Spawner démarré - intervalle: %v, timeout: %v\n", s.interval, s.timeout)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[spawn] Spawner arrêté")
			return
		case <-ticker.C:
			s.spawnWordMon()
		}
	}
}

// spawnWordMon génère un nouveau WordMon
func (s *SpawnerService) spawnWordMon() {
	// Vérifier s'il y a déjà un WordMon actif
	if s.store.GetCurrentSpawn() != nil {
		// Ne pas spawner s'il y en a déjà un
		return
	}

	// Générer un nouveau WordMon
	word := core.SpawnWord()
	s.store.SetCurrentSpawn(&word)

	fmt.Printf("[spawn] Nouveau WordMon: \"%s\" (%s, %d points)\n",
		word.Text, word.Rarity, word.Points)

	// Programmer la fuite automatique après timeout
	go s.scheduleAutoFlee(&word)
}

// scheduleAutoFlee programme la fuite automatique d'un WordMon
func (s *SpawnerService) scheduleAutoFlee(word *core.Word) {
	time.Sleep(s.timeout)

	// Vérifier si le WordMon est toujours actif
	currentSpawn := s.store.GetCurrentSpawn()
	if currentSpawn != nil && currentSpawn.ID == word.ID {
		s.store.ClearCurrentSpawn()
		fmt.Printf("[spawn] \"%s\" s'est enfui (timeout)\n", word.Text)
	}
}
