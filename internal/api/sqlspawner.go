package api

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/SamG1008/wordmon-go/internal/config"
	"github.com/SamG1008/wordmon-go/internal/core"
	"github.com/SamG1008/wordmon-go/internal/store"
)

// SQLSpawnerService gère les spawns pour l'API SQL
type SQLSpawnerService struct {
	store       *store.SQLStore
	server      *SQLServer
	gameConfig  *config.GameConfig
	interval    time.Duration
	fleeTimeout time.Duration
	ticker      *time.Ticker
	stopChan    chan struct{}
}

// NewSQLSpawnerService crée un nouveau service de spawn SQL
func NewSQLSpawnerService(sqlStore *store.SQLStore, server *SQLServer, gameConfig *config.GameConfig, interval, fleeTimeout time.Duration) *SQLSpawnerService {
	return &SQLSpawnerService{
		store:       sqlStore,
		server:      server,
		gameConfig:  gameConfig,
		interval:    interval,
		fleeTimeout: fleeTimeout,
		stopChan:    make(chan struct{}),
	}
}

// Start démarre le service de spawn
func (s *SQLSpawnerService) Start() {
	s.ticker = time.NewTicker(s.interval)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.spawnWordMon()
			case <-s.stopChan:
				return
			}
		}
	}()

	fmt.Printf("[spawn] SQL Spawner démarré - intervalle: %v, timeout: %v\n", s.interval, s.fleeTimeout)
}

// Stop arrête le service de spawn
func (s *SQLSpawnerService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
	fmt.Println("[spawn] SQL Spawner arrêté")
}

// spawnWordMon génère un nouveau WordMon
func (s *SQLSpawnerService) spawnWordMon() {
	// Vérifier s'il y a déjà un WordMon actif
	if s.server.GetCurrentSpawn() != nil {
		return
	}

	// Sélectionner une rareté selon les poids
	rarity := s.selectRarity()

	// Sélectionner un mot aléatoire de cette rareté depuis la DB
	word, err := s.store.RandomByRarity(string(rarity))
	if err != nil {
		fmt.Printf("[spawn] Erreur sélection mot: %v\n", err)
		return
	}

	// Définir le spawn actuel
	s.server.SetCurrentSpawn(word)

	fmt.Printf("[spawn] Nouveau WordMon: \"%s\" (%s, %d points)\n",
		word.Text, word.Rarity, word.Points)

	// Programmer la fuite automatique
	s.scheduleAutoFlee(word)
}

// selectRarity sélectionne une rareté selon les poids configurés
func (s *SQLSpawnerService) selectRarity() core.Rarity {
	weights := s.gameConfig.RarityWeights
	total := weights.Common + weights.Rare + weights.Legendary

	roll := rand.Intn(total)

	if roll < weights.Common {
		return core.Common
	} else if roll < weights.Common+weights.Rare {
		return core.Rare
	} else {
		return core.Legendary
	}
}

// scheduleAutoFlee programme la fuite automatique du WordMon
func (s *SQLSpawnerService) scheduleAutoFlee(word *core.Word) {
	go func() {
		time.Sleep(s.fleeTimeout)

		// Vérifier si le WordMon est toujours là
		current := s.server.GetCurrentSpawn()
		if current != nil && current.ID == word.ID {
			s.server.ClearCurrentSpawn()
			fmt.Printf("[spawn] \"%s\" s'est enfui (timeout)\n", word.Text)
		}
	}()
}
