package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SamG1008/wordmon-go/internal/api"
	"github.com/SamG1008/wordmon-go/internal/config"
	"github.com/SamG1008/wordmon-go/internal/core"
	"github.com/SamG1008/wordmon-go/internal/store"
	"github.com/joho/godotenv"
)

const Version = "0.7.0"

func init() {
	// Charger les variables d'environnement depuis le fichier .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Avertissement: fichier .env non trouvé")
	}
}

func main() {
	// Gestion des panics
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC intercepté: %v\n", r)
			fmt.Println("Le serveur va s'arrêter proprement...")
			os.Exit(1)
		}
	}()

	// Initialiser le générateur aléatoire
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Flags CLI
	showVersion := flag.Bool("version", false, "Affiche la version")
	showVersionShort := flag.Bool("v", false, "Affiche la version")
	port := flag.String("port", "8080", "Port du serveur")
	portShort := flag.String("p", "8080", "Port du serveur")

	flag.Parse()

	// Si --version ou -v est utilisé, afficher seulement la version
	if *showVersion || *showVersionShort {
		fmt.Println(Version)
		os.Exit(0)
	}

	// Déterminer le port
	serverPort := *port
	if serverPort == "8080" && *portShort != "8080" {
		serverPort = *portShort
	}

	fmt.Printf("WordMon Go API Server version %s\n", Version)
	fmt.Printf("Exercice 07: API REST avec Gin\n")
	fmt.Println("=== Chargement des configurations ===")

	// Charger les configurations
	gameConfig, err := config.LoadGameConfig("configs/game.yaml")
	if err != nil {
		fmt.Printf("Erreur chargement config: %v\n", err)
		os.Exit(1)
	}

	wordsConfig, err := config.LoadWordsConfig("configs/words.json")
	if err != nil {
		fmt.Printf("Erreur chargement words: %v\n", err)
		os.Exit(1)
	}

	challengesConfig, err := config.LoadChallengesConfig("configs/challenges.yaml")
	if err != nil {
		fmt.Printf("Erreur chargement challenges: %v\n", err)
		os.Exit(1)
	}
	_ = challengesConfig // Utilisé dans de futurs exercices

	// Configurer le système de mots
	configureGameSystems(wordsConfig, gameConfig)

	fmt.Printf("[config] API configurée avec %d mots\n", len(wordsConfig.Words))
	fmt.Println()

	// Créer le store en mémoire
	memStore := store.NewMemoryStore()

	// Créer le serveur API
	server := api.NewServer(memStore, gameConfig)

	// Créer et démarrer le spawner
	spawnInterval := time.Duration(gameConfig.Spawner.IntervalSeconds) * time.Second
	fleeTimeout := time.Duration(gameConfig.Spawner.AutoFleeAfterSeconds) * time.Second
	spawner := api.NewSpawnerService(memStore, spawnInterval, fleeTimeout)

	// Context pour arrêt propre
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Démarrer le spawner en arrière-plan
	go spawner.Start(ctx)

	// Gérer l'arrêt propre avec Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Démarrer le serveur dans une goroutine
	go func() {
		fmt.Printf("[server] Démarrage du serveur sur :%s\n", serverPort)
		fmt.Printf("[server] Endpoints disponibles:\n")
		fmt.Printf("  GET  /status\n")
		fmt.Printf("  POST /players\n")
		fmt.Printf("  GET  /players/:id\n")
		fmt.Printf("  GET  /spawn/current\n")
		fmt.Printf("  POST /encounter/attempt\n")
		fmt.Printf("  GET  /leaderboard\n")
		fmt.Println()

		if err := server.Run(serverPort); err != nil {
			fmt.Printf("Erreur serveur: %v\n", err)
			cancel()
		}
	}()

	// Attendre le signal d'arrêt
	<-sigChan
	fmt.Println("\n[server] Arrêt demandé par l'utilisateur...")
	cancel()

	// Laisser le temps aux goroutines de se terminer
	time.Sleep(500 * time.Millisecond)

	fmt.Println("[server] Serveur arrêté proprement")
}

// configureGameSystems configure le système de mots avec les données chargées
func configureGameSystems(wordsConfig *config.WordsConfig, gameConfig *config.GameConfig) {
	// Convertir les WordEntry du config vers les types du core
	words := make([]core.WordEntry, len(wordsConfig.Words))
	for i, entry := range wordsConfig.Words {
		words[i] = core.WordEntry{
			ID:     entry.ID,
			Text:   entry.Text,
			Rarity: entry.Rarity,
		}
	}

	// Configurer les poids et récompenses
	weights := core.RarityWeights{
		Common:    gameConfig.RarityWeights.Common,
		Rare:      gameConfig.RarityWeights.Rare,
		Legendary: gameConfig.RarityWeights.Legendary,
	}

	rewards := core.XPRewards{
		Common:    gameConfig.XPRewards.Common,
		Rare:      gameConfig.XPRewards.Rare,
		Legendary: gameConfig.XPRewards.Legendary,
	}

	// Appliquer la configuration au système de mots
	core.ConfigureWords(words, weights, rewards)
}
