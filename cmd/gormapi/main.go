package main

import (
	"log"
	"os"

	"github.com/SamG1008/wordmon-go/internal/api"
	"github.com/SamG1008/wordmon-go/internal/config"
	"github.com/SamG1008/wordmon-go/internal/core"
	"github.com/SamG1008/wordmon-go/internal/store"
)

func main() {
	log.Println("WordMon Go API Server version 0.8.1")
	log.Println("Exercice 08: Persistance avec PostgreSQL + GORM")

	// Configuration
	log.Println("=== Chargement des configurations ===")

	gameConfig, err := config.LoadGameConfig("configs/game.yaml")
	if err != nil {
		log.Fatalf("Erreur chargement game config: %v", err)
	}

	wordsConfig, err := config.LoadWordsConfig("configs/words.json")
	if err != nil {
		log.Fatalf("Erreur chargement words config: %v", err)
	}

	challengesConfig, err := config.LoadChallengesConfig("configs/challenges.yaml")
	if err != nil {
		log.Fatalf("Erreur chargement challenges config: %v", err)
	}

	// Store GORM
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/wordmon?sslmode=disable"
	}

	store, err := store.NewGORMStore(databaseURL)
	if err != nil {
		log.Fatalf("Erreur création store GORM: %v", err)
	}
	defer store.Close()

	// Seed initial - conversion des mots avec calcul des points
	coreWords := make([]core.Word, len(wordsConfig.Words))
	for i, w := range wordsConfig.Words {
		rarity := core.Rarity(w.Rarity)
		points := 0

		// Calcul des points selon la rareté
		switch rarity {
		case core.Common:
			points = gameConfig.XPRewards.Common
		case core.Rare:
			points = gameConfig.XPRewards.Rare
		case core.Legendary:
			points = gameConfig.XPRewards.Legendary
		}

		coreWords[i] = core.Word{
			ID:     w.ID,
			Text:   w.Text,
			Rarity: rarity,
			Points: points,
		}
	}

	if err := store.Seed(coreWords); err != nil {
		log.Fatalf("Erreur seed: %v", err)
	}

	// Serveur API universel
	server := api.NewUniversalServer(store, gameConfig, challengesConfig)

	log.Println("[server] Démarrage du serveur GORM sur :8080")
	log.Println("[server] Endpoints disponibles:")
	log.Println("  GET  /status")
	log.Println("  POST /players")
	log.Println("  GET  /players/:id")
	log.Println("  GET  /spawn/current")
	log.Println("  POST /encounter/attempt")
	log.Println("  GET  /leaderboard")

	if err := server.Run(":8080"); err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}
