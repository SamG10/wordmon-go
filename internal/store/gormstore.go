package store

import (
	"fmt"
	"log"

	"github.com/SamG1008/wordmon-go/internal/core"
	"github.com/SamG1008/wordmon-go/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GORMStore implémente la persistance avec GORM
type GORMStore struct {
	db *gorm.DB
}

// NewGORMStore crée un nouveau store GORM
func NewGORMStore(databaseURL string) (*GORMStore, error) {
	// Configuration GORM avec logs
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, fmt.Errorf("erreur connexion GORM: %w", err)
	}

	// Test de la connexion
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("erreur récupération DB: %w", err)
	}
	
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("erreur ping DB: %w", err)
	}

	log.Printf("[db] Connected to Postgres via GORM (wordmon)")

	// Auto-migration des modèles
	err = db.AutoMigrate(&models.Player{}, &models.Word{}, &models.Capture{})
	if err != nil {
		return nil, fmt.Errorf("erreur auto-migration: %w", err)
	}

	log.Printf("[migrate] GORM auto-migration completed")

	return &GORMStore{db: db}, nil
}

// Close ferme la connexion à la base
func (s *GORMStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// === PLAYER REPOSITORY ===

// Create crée un nouveau joueur
func (s *GORMStore) Create(name string) (*core.Player, error) {
	player := &models.Player{
		Name:  name,
		XP:    0,
		Level: 1,
	}

	if err := s.db.Create(player).Error; err != nil {
		return nil, fmt.Errorf("erreur création joueur: %w", err)
	}

	return &core.Player{
		ID:    player.ID,
		Name:  player.Name,
		XP:    player.XP,
		Level: player.Level,
	}, nil
}

// Get récupère un joueur par ID
func (s *GORMStore) Get(id string) (*core.Player, error) {
	var player models.Player
	
	if err := s.db.First(&player, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("joueur introuvable: %s", id)
		}
		return nil, fmt.Errorf("erreur récupération joueur: %w", err)
	}

	return &core.Player{
		ID:    player.ID,
		Name:  player.Name,
		XP:    player.XP,
		Level: player.Level,
	}, nil
}

// List récupère une liste de joueurs
func (s *GORMStore) List(limit int) ([]core.Player, error) {
	var players []models.Player
	
	query := s.db.Order("xp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&players).Error; err != nil {
		return nil, fmt.Errorf("erreur liste joueurs: %w", err)
	}

	result := make([]core.Player, len(players))
	for i, p := range players {
		result[i] = core.Player{
			ID:    p.ID,
			Name:  p.Name,
			XP:    p.XP,
			Level: p.Level,
		}
	}

	return result, nil
}

// UpdateXP met à jour l'XP et le niveau d'un joueur
func (s *GORMStore) UpdateXP(id string, newXP int, newLevel int) error {
	result := s.db.Model(&models.Player{}).Where("id = ?", id).Updates(map[string]interface{}{
		"xp":    newXP,
		"level": newLevel,
	})

	if result.Error != nil {
		return fmt.Errorf("erreur mise à jour XP: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("joueur introuvable pour mise à jour XP: %s", id)
	}

	return nil
}

// === WORD REPOSITORY ===

// Seed ajoute les mots en base
func (s *GORMStore) Seed(words []core.Word) error {
	// Vérifier si des mots existent déjà
	var count int64
	s.db.Model(&models.Word{}).Count(&count)
	
	if count > 0 {
		log.Printf("[seed] %d words already in DB, skipping seed", count)
		return nil
	}

	// Conversion vers modèles GORM
	gormWords := make([]models.Word, len(words))
	for i, w := range words {
		gormWords[i] = models.Word{
			ID:     w.ID,
			Text:   w.Text,
			Rarity: string(w.Rarity),
			Points: w.Points,
		}
	}

	// Insertion en batch
	if err := s.db.Create(&gormWords).Error; err != nil {
		return fmt.Errorf("erreur seed words: %w", err)
	}

	log.Printf("[seed] %d words loaded into DB", len(words))
	return nil
}

// Get récupère un mot par ID
func (s *GORMStore) GetWord(id string) (*core.Word, error) {
	var word models.Word
	
	if err := s.db.First(&word, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("mot introuvable: %s", id)
		}
		return nil, fmt.Errorf("erreur récupération mot: %w", err)
	}

	return &core.Word{
		ID:     word.ID,
		Text:   word.Text,
		Rarity: core.Rarity(word.Rarity),
		Points: word.Points,
	}, nil
}

// RandomByRarity récupère un mot aléatoire selon la rareté
func (s *GORMStore) RandomByRarity(rarity string) (*core.Word, error) {
	var word models.Word
	
	if err := s.db.Where("rarity = ?", rarity).Order("RANDOM()").First(&word).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("aucun mot trouvé pour rareté: %s", rarity)
		}
		return nil, fmt.Errorf("erreur sélection mot aléatoire: %w", err)
	}

	return &core.Word{
		ID:     word.ID,
		Text:   word.Text,
		Rarity: core.Rarity(word.Rarity),
		Points: word.Points,
	}, nil
}

// === CAPTURE REPOSITORY ===

// Add ajoute une capture (avec transaction)
func (s *GORMStore) Add(playerId, wordId string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Vérifier que le joueur existe
		var player models.Player
		if err := tx.First(&player, "id = ?", playerId).Error; err != nil {
			return fmt.Errorf("joueur introuvable: %s", playerId)
		}

		// Vérifier que le mot existe
		var word models.Word
		if err := tx.First(&word, "id = ?", wordId).Error; err != nil {
			return fmt.Errorf("mot introuvable: %s", wordId)
		}

		// Créer la capture
		capture := &models.Capture{
			PlayerID: playerId,
			WordID:   wordId,
		}

		if err := tx.Create(capture).Error; err != nil {
			return fmt.Errorf("erreur création capture: %w", err)
		}

		// Mettre à jour l'XP du joueur
		newXP := player.XP + word.Points
		newLevel := (newXP / 100) + 1

		if err := tx.Model(&player).Updates(map[string]interface{}{
			"xp":    newXP,
			"level": newLevel,
		}).Error; err != nil {
			return fmt.Errorf("erreur mise à jour XP: %w", err)
		}

		log.Printf("[api] Capture success: %s +%dXP (level=%d)", player.Name, word.Points, newLevel)
		return nil
	})
}

// ListByPlayer récupère les captures d'un joueur
func (s *GORMStore) ListByPlayer(playerId string) ([]core.Word, error) {
	var captures []models.Capture
	
	if err := s.db.Preload("Word").Where("player_id = ?", playerId).Find(&captures).Error; err != nil {
		return nil, fmt.Errorf("erreur récupération captures: %w", err)
	}

	words := make([]core.Word, len(captures))
	for i, c := range captures {
		words[i] = core.Word{
			ID:     c.Word.ID,
			Text:   c.Word.Text,
			Rarity: core.Rarity(c.Word.Rarity),
			Points: c.Word.Points,
		}
	}

	return words, nil
}
