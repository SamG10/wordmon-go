package store

import "github.com/SamG1008/wordmon-go/internal/core"

// PlayerStore interface pour la gestion des joueurs
type PlayerStore interface {
	Create(name string) (*core.Player, error)
	Get(id string) (*core.Player, error)
	List(limit int) ([]core.Player, error)
	UpdateXP(id string, newXP int, newLevel int) error
}

// WordStore interface pour la gestion des mots
type WordStore interface {
	Seed(words []core.Word) error
	GetWord(id string) (*core.Word, error)
	RandomByRarity(rarity string) (*core.Word, error)
}

// CaptureStore interface pour la gestion des captures
type CaptureStore interface {
	Add(playerId, wordId string) error
	ListByPlayer(playerId string) ([]core.Word, error)
}

// Store interface complète
type Store interface {
	PlayerStore
	WordStore
	CaptureStore
	Close() error
}
