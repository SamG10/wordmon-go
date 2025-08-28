package store

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/SamG1008/wordmon-go/internal/core"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// SQLStore implémente la persistance avec PostgreSQL
type SQLStore struct {
	db *sql.DB
}

// NewSQLStore crée un nouveau store SQL
func NewSQLStore(databaseURL string) (*SQLStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erreur connexion DB: %w", err)
	}

	// Test de la connexion
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erreur ping DB: %w", err)
	}

	log.Printf("[db] Connected to Postgres (wordmon)")

	return &SQLStore{db: db}, nil
}

// Close ferme la connexion à la base
func (s *SQLStore) Close() error {
	return s.db.Close()
}

// === PLAYER REPOSITORY ===

// Create crée un nouveau joueur
func (s *SQLStore) Create(name string) (*core.Player, error) {
	id := uuid.New().String()

	query := `INSERT INTO players (id, name, xp, level) VALUES ($1, $2, 0, 1)`
	_, err := s.db.Exec(query, id, name)
	if err != nil {
		return nil, fmt.Errorf("erreur création joueur: %w", err)
	}

	log.Printf("[api] Player created: %s (id=%s)", name, id)

	return &core.Player{
		ID:    id,
		Name:  name,
		XP:    0,
		Level: 1,
	}, nil
}

// Get récupère un joueur par ID
func (s *SQLStore) Get(id string) (*core.Player, error) {
	player := &core.Player{}
	query := `SELECT id, name, xp, level FROM players WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(&player.ID, &player.Name, &player.XP, &player.Level)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("joueur non trouvé: %s", id)
		}
		return nil, fmt.Errorf("erreur récupération joueur: %w", err)
	}

	// Récupération de l'inventaire (captures)
	inventory, err := s.getPlayerInventory(id)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération inventaire: %w", err)
	}
	player.Inventory = inventory

	return player, nil
}

// getPlayerInventory récupère l'inventaire d'un joueur
func (s *SQLStore) getPlayerInventory(playerID string) (map[string]int, error) {
	inventory := make(map[string]int)

	query := `
		SELECT w.id, COUNT(*) 
		FROM captures c 
		JOIN words w ON c.word_id = w.id 
		WHERE c.player_id = $1 
		GROUP BY w.id`

	rows, err := s.db.Query(query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var wordID string
		var count int
		if err := rows.Scan(&wordID, &count); err != nil {
			return nil, err
		}
		inventory[wordID] = count
	}

	return inventory, nil
}

// List récupère la liste des joueurs (pour le leaderboard)
func (s *SQLStore) List(limit int) ([]core.Player, error) {
	query := `SELECT id, name, xp, level FROM players ORDER BY xp DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération leaderboard: %w", err)
	}
	defer rows.Close()

	var players []core.Player
	for rows.Next() {
		var player core.Player
		if err := rows.Scan(&player.ID, &player.Name, &player.XP, &player.Level); err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

// UpdateXP met à jour l'XP et le niveau d'un joueur
func (s *SQLStore) UpdateXP(id string, newXP int, newLevel int) error {
	query := `UPDATE players SET xp = $1, level = $2 WHERE id = $3`
	_, err := s.db.Exec(query, newXP, newLevel, id)
	if err != nil {
		return fmt.Errorf("erreur mise à jour XP: %w", err)
	}
	return nil
}

// === WORD REPOSITORY ===

// Seed insère les mots depuis la configuration
func (s *SQLStore) Seed(words []core.Word) error {
	// Vérifier si les mots existent déjà
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM words").Scan(&count)
	if err != nil {
		return fmt.Errorf("erreur vérification seed: %w", err)
	}

	if count > 0 {
		log.Printf("[seed] %d words already in DB, skipping seed", count)
		return nil
	}

	// Insérer les mots
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("erreur transaction seed: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO words (id, text, rarity, points) VALUES ($1, $2, $3, $4)`
	for _, word := range words {
		_, err := tx.Exec(query, word.ID, word.Text, string(word.Rarity), word.Points)
		if err != nil {
			return fmt.Errorf("erreur insertion mot %s: %w", word.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur commit seed: %w", err)
	}

	log.Printf("[seed] %d words loaded into DB", len(words))
	return nil
}

// GetWord récupère un mot par ID
func (s *SQLStore) GetWord(id string) (*core.Word, error) {
	word := &core.Word{}
	query := `SELECT id, text, rarity, points FROM words WHERE id = $1`

	var rarityStr string
	err := s.db.QueryRow(query, id).Scan(&word.ID, &word.Text, &rarityStr, &word.Points)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mot non trouvé: %s", id)
		}
		return nil, fmt.Errorf("erreur récupération mot: %w", err)
	}

	word.Rarity = core.Rarity(rarityStr)
	return word, nil
}

// RandomByRarity sélectionne un mot aléatoire par rareté
func (s *SQLStore) RandomByRarity(rarity string) (*core.Word, error) {
	word := &core.Word{}
	query := `SELECT id, text, rarity, points FROM words WHERE rarity = $1 ORDER BY RANDOM() LIMIT 1`

	var rarityStr string
	err := s.db.QueryRow(query, rarity).Scan(&word.ID, &word.Text, &rarityStr, &word.Points)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("aucun mot trouvé pour rareté: %s", rarity)
		}
		return nil, fmt.Errorf("erreur sélection mot aléatoire: %w", err)
	}

	word.Rarity = core.Rarity(rarityStr)
	return word, nil
}

// === CAPTURE REPOSITORY ===

// Add enregistre une capture (avec transaction)
func (s *SQLStore) Add(playerID, wordID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("erreur transaction capture: %w", err)
	}
	defer tx.Rollback()

	// 1. Insérer la capture
	captureID := uuid.New().String()
	captureQuery := `INSERT INTO captures (id, player_id, word_id) VALUES ($1, $2, $3)`
	_, err = tx.Exec(captureQuery, captureID, playerID, wordID)
	if err != nil {
		return fmt.Errorf("erreur insertion capture: %w", err)
	}

	// 2. Récupérer les points du mot
	var points int
	pointsQuery := `SELECT points FROM words WHERE id = $1`
	err = tx.QueryRow(pointsQuery, wordID).Scan(&points)
	if err != nil {
		return fmt.Errorf("erreur récupération points: %w", err)
	}

	// 3. Mettre à jour l'XP du joueur
	updateQuery := `UPDATE players SET xp = xp + $1, level = (xp + $1) / 100 + 1 WHERE id = $2`
	_, err = tx.Exec(updateQuery, points, playerID)
	if err != nil {
		return fmt.Errorf("erreur mise à jour XP: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur commit capture: %w", err)
	}

	return nil
}

// ListByPlayer récupère les captures d'un joueur
func (s *SQLStore) ListByPlayer(playerID string) ([]core.Word, error) {
	query := `
		SELECT w.id, w.text, w.rarity, w.points 
		FROM captures c 
		JOIN words w ON c.word_id = w.id 
		WHERE c.player_id = $1 
		ORDER BY c.captured_at DESC`

	rows, err := s.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération captures: %w", err)
	}
	defer rows.Close()

	var words []core.Word
	for rows.Next() {
		var word core.Word
		var rarityStr string
		if err := rows.Scan(&word.ID, &word.Text, &rarityStr, &word.Points); err != nil {
			return nil, err
		}
		word.Rarity = core.Rarity(rarityStr)
		words = append(words, word)
	}

	return words, nil
}
