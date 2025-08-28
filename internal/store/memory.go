package store

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/SamG1008/wordmon-go/internal/core"
)

// MemoryStore stocke les données en mémoire
type MemoryStore struct {
	mu           sync.RWMutex
	players      map[string]*core.Player
	currentSpawn *core.Word
	startTime    time.Time
	nextPlayerID int
}

// NewMemoryStore crée un nouveau store en mémoire
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		players:      make(map[string]*core.Player),
		startTime:    time.Now(),
		nextPlayerID: 1,
	}
}

// CreatePlayer crée un nouveau joueur
func (s *MemoryStore) CreatePlayer(name string) (*core.Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Vérifier si le nom est déjà pris
	for _, player := range s.players {
		if player.Name == name {
			return nil, fmt.Errorf("nom déjà pris: %s", name)
		}
	}

	// Créer le joueur avec un ID unique
	playerID := fmt.Sprintf("p%d", s.nextPlayerID)
	s.nextPlayerID++

	player := core.NewPlayer(playerID, name)
	s.players[playerID] = &player

	return &player, nil
}

// GetPlayer récupère un joueur par ID
func (s *MemoryStore) GetPlayer(id string) (*core.Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	player, exists := s.players[id]
	if !exists {
		return nil, fmt.Errorf("joueur non trouvé: %s", id)
	}

	return player, nil
}

// UpdatePlayer met à jour un joueur
func (s *MemoryStore) UpdatePlayer(player *core.Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.players[player.ID]; !exists {
		return fmt.Errorf("joueur non trouvé: %s", player.ID)
	}

	s.players[player.ID] = player
	return nil
}

// GetAllPlayers retourne tous les joueurs
func (s *MemoryStore) GetAllPlayers() []*core.Player {
	s.mu.RLock()
	defer s.mu.RUnlock()

	players := make([]*core.Player, 0, len(s.players))
	for _, player := range s.players {
		players = append(players, player)
	}

	return players
}

// GetLeaderboard retourne le classement des joueurs
func (s *MemoryStore) GetLeaderboard(limit int) []*core.Player {
	players := s.GetAllPlayers()

	// Trier par XP décroissant
	sort.Slice(players, func(i, j int) bool {
		return players[i].XP > players[j].XP
	})

	// Limiter le résultat
	if limit > 0 && limit < len(players) {
		players = players[:limit]
	}

	return players
}

// SetCurrentSpawn définit le WordMon actuel
func (s *MemoryStore) SetCurrentSpawn(word *core.Word) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentSpawn = word
}

// GetCurrentSpawn retourne le WordMon actuel
func (s *MemoryStore) GetCurrentSpawn() *core.Word {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentSpawn
}

// ClearCurrentSpawn supprime le WordMon actuel
func (s *MemoryStore) ClearCurrentSpawn() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentSpawn = nil
}

// GetActivePlayersCount retourne le nombre de joueurs actifs
func (s *MemoryStore) GetActivePlayersCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.players)
}

// GetUptimeSeconds retourne le temps depuis le démarrage
func (s *MemoryStore) GetUptimeSeconds() int {
	return int(time.Since(s.startTime).Seconds())
}
