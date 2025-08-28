package main

import (
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

const Version = "0.8.0"

// convertConfigWordsToCore convertit les mots de config vers core.Word
func convertConfigWordsToCore(configWords []config.WordEntry) []core.Word {
	var coreWords []core.Word
	for _, configWord := range configWords {
		coreWord := core.Word{
			ID:     configWord.ID,
			Text:   configWord.Text,
			Rarity: core.Rarity(configWord.Rarity),
			Points: getPointsForRarity(configWord.Rarity),
		}
		coreWords = append(coreWords, coreWord)
	}
	return coreWords
}

// getPointsForRarity retourne les points selon la rareté
func getPointsForRarity(rarity string) int {
	switch rarity {
	case "Common":
		return 5
	case "Rare":
		return 20
	case "Legendary":
		return 100
	default:
		return 5
	}
}

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
	fmt.Printf("Exercice 08: Persistance avec PostgreSQL\n")
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

	fmt.Printf("[config] API configurée avec %d mots\n", len(wordsConfig.Words))
	fmt.Println()

	// Créer le store SQL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/wordmon?sslmode=disable"
	}

	sqlStore, err := store.NewSQLStore(databaseURL)
	if err != nil {
		fmt.Printf("Erreur connexion DB: %v\n", err)
		os.Exit(1)
	}
	defer sqlStore.Close()

	// Seed des mots dans la DB
	coreWords := convertConfigWordsToCore(wordsConfig.Words)
	if err := sqlStore.Seed(coreWords); err != nil {
		fmt.Printf("Erreur seed words: %v\n", err)
		os.Exit(1)
	}

	// Créer le serveur API avec adapter SQL
	server := api.NewSQLServer(sqlStore, gameConfig)

	// Créer et démarrer le spawner SQL
	spawnInterval := time.Duration(gameConfig.Spawner.IntervalSeconds) * time.Second
	fleeTimeout := time.Duration(gameConfig.Spawner.AutoFleeAfterSeconds) * time.Second

	spawner := api.NewSQLSpawnerService(sqlStore, server, gameConfig, spawnInterval, fleeTimeout)

	// Démarrer le spawner en arrière-plan
	go spawner.Start()

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
		}
	}()

	// Attendre le signal d'arrêt
	<-sigChan
	fmt.Println("\n[server] Arrêt demandé par l'utilisateur...")
	spawner.Stop()

	// Laisser le temps aux goroutines de se terminer
	time.Sleep(500 * time.Millisecond)

	fmt.Println("[server] Serveur arrêté proprement")
}
