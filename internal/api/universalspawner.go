package api

import (
	"log"
	"sync"
	"time"

	"github.com/SamG1008/wordmon-go/internal/config"
	"github.com/SamG1008/wordmon-go/internal/core"
	"github.com/SamG1008/wordmon-go/internal/store"
)

// UniversalSpawner spawner générique qui fonctionne avec n'importe quel Store
type UniversalSpawner struct {
	store       store.Store
	gameConfig  *config.GameConfig
	currentSpawn *core.Word
	mutex       sync.RWMutex
	ticker      *time.Ticker
}

// NewUniversalSpawner crée un nouveau spawner universel
func NewUniversalSpawner(s store.Store, gameConfig *config.GameConfig) *UniversalSpawner {
	return &UniversalSpawner{
		store:      s,
		gameConfig: gameConfig,
	}
}

// Start démarre le spawner
func (s *UniversalSpawner) Start() {
	log.Printf("[spawn] Universal Spawner démarré - intervalle: %ds, timeout: %ds", 
		s.gameConfig.Spawner.IntervalSeconds, s.gameConfig.Spawner.AutoFleeAfterSeconds)

	// Premier spawn immédiat
	s.spawnNewWordMon()

	// Spawn périodique
	s.ticker = time.NewTicker(time.Duration(s.gameConfig.Spawner.IntervalSeconds) * time.Second)
	go func() {
		for range s.ticker.C {
			s.spawnNewWordMon()
		}
	}()
}

// Stop arrête le spawner
func (s *UniversalSpawner) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
}

// GetCurrentSpawn retourne le spawn actuel
func (s *UniversalSpawner) GetCurrentSpawn() *core.Word {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.currentSpawn
}

// ForceNewSpawn force un nouveau spawn
func (s *UniversalSpawner) ForceNewSpawn() {
	s.spawnNewWordMon()
}

// spawnNewWordMon génère un nouveau WordMon selon les poids de rareté
func (s *UniversalSpawner) spawnNewWordMon() {
	// Sélection de la rareté selon les poids
	rarity := s.selectRarityByWeight()
	
	// Récupération d'un mot aléatoire de cette rareté
	word, err := s.store.RandomByRarity(rarity)
	if err != nil {
		log.Printf("[spawn] Erreur sélection mot: %v", err)
		return
	}

	s.mutex.Lock()
	s.currentSpawn = word
	s.mutex.Unlock()

	log.Printf("[spawn] Nouveau WordMon spawné: %s (%s, %d pts)", 
		word.Text, word.Rarity, word.Points)
}

// selectRarityByWeight sélectionne une rareté selon les poids configurés
func (s *UniversalSpawner) selectRarityByWeight() string {
	weights := s.gameConfig.RarityWeights
	totalWeight := weights.Common + weights.Rare + weights.Legendary
	
	// Simple distribution basée sur les poids (version simplifiée)
	// Pour une vraie randomisation, il faudrait utiliser math/rand
	commonThreshold := weights.Common * 100 / totalWeight
	rareThreshold := commonThreshold + (weights.Rare * 100 / totalWeight)
	
	// Simulation simple (toujours common pour maintenant)
	if commonThreshold > 50 {
		return "common"
	} else if rareThreshold > 75 {
		return "rare"  
	} else {
		return "legendary"
	}
}
