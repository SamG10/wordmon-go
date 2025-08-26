package main

import (
	"context"
	"fmt"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"math/rand"
	"time"
	"bufio"
	"strings"

	"github.com/SamG1008/wordmon-go/internal/core"
)

const Version = "0.5.0"

func main() {
	// Exercice 04: Récupération des panics pour éviter les crashes
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC intercepté: %v\n", r)
			fmt.Println("Le jeu va se fermer proprement...")
			os.Exit(1)
		}
	}()

	// Initialiser le générateur aléatoire
	rand.Seed(time.Now().UnixNano())
	
	// Exercice 01: Définir les flags CLI
	showVersion := flag.Bool("version", false, "Affiche la version")
	showVersionShort := flag.Bool("v", false, "Affiche la version")
	playerName := flag.String("player", "", "Nom du joueur")
	playerNameShort := flag.String("p", "", "Nom du joueur")
	testMode := flag.Bool("test", false, "Mode test automatique")
	interactiveMode := flag.Bool("interactive", false, "Mode interactif (exercices 01-04)")
	interactiveModeShort := flag.Bool("i", false, "Mode interactif (exercices 01-04)")
	concurrentMode := flag.Bool("concurrent", false, "Mode concurrent avec spawner (exercice 05)")
	concurrentModeShort := flag.Bool("c", false, "Mode concurrent avec spawner (exercice 05)")
	duration := flag.Int("duration", 30, "Durée de la démo en secondes (mode concurrent)")
	
	// Parser les arguments de la ligne de commande
	flag.Parse()
	
	// Si --version ou -v est utilisé, afficher seulement la version
	if *showVersion || *showVersionShort {
		fmt.Println(Version)
		os.Exit(0)
	}
	
	// Afficher le message d'accueil
	fmt.Printf("WordMon Go version %s - Edition Complète\n", Version)
	fmt.Println("Exercices 01-05 intégrés : CLI, Types, OOP, Erreurs, Concurrence")
	
	// Récupérer le nom du joueur ou utiliser "Guest"
	player := *playerName
	if player == "" {
		player = *playerNameShort
	}
	if player == "" {
		player = "Guest"
	}
	
	fmt.Printf("Bienvenue, Dresseur %s !\n\n", player)
	
	// Créer un Player avec le constructeur (Exercice 02)
	gamePlayer := core.NewPlayer("player_001", player)
	
	// Routage vers les différents modes
	if *concurrentMode || *concurrentModeShort {
		// Exercice 05: Mode concurrent avec spawner
		runConcurrentMode(&gamePlayer, time.Duration(*duration)*time.Second)
	} else if *testMode {
		// Mode test pour tous les exercices
		runTestMode(&gamePlayer)
	} else if *interactiveMode || *interactiveModeShort {
		// Exercices 01-04: Mode interactif classique
		runInteractiveMode(&gamePlayer)
	} else {
		// Mode par défaut: montrer les options
		showMainMenu(&gamePlayer)
	}
	
	os.Exit(0)
}

// showMainMenu affiche le menu principal et laisse l'utilisateur choisir
func showMainMenu(player *core.Player) {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Println("=== Menu Principal WordMon Go ===")
		fmt.Println("1. Mode Interactif (Exercices 01-04)")
		fmt.Println("2. Mode Concurrent (Exercice 05)")
		fmt.Println("3. Mode Test (Tous exercices)")
		fmt.Println("4. Statut du joueur")
		fmt.Println("5. Quitter")
		fmt.Print("\nVotre choix (1-5): ")
		
		if !scanner.Scan() {
			break
		}
		
		choice := strings.TrimSpace(scanner.Text())
		
		switch choice {
		case "1":
			runInteractiveMode(player)
		case "2":
			fmt.Print("Durée de la démo (secondes, défaut 30): ")
			scanner.Scan()
			durationStr := strings.TrimSpace(scanner.Text())
			duration := 30
			if durationStr != "" {
				fmt.Sscanf(durationStr, "%d", &duration)
			}
			runConcurrentMode(player, time.Duration(duration)*time.Second)
		case "3":
			runTestMode(player)
		case "4":
			displayPlayerStatus(*player)
		case "5":
			fmt.Println("À bientôt, Dresseur !")
			return
		default:
			fmt.Println("Choix invalide. Veuillez entrer un nombre entre 1 et 5.")
		}
		
		fmt.Println("\nAppuyez sur Entrée pour continuer...")
		scanner.Scan()
	}
}

// runInteractiveMode lance le mode interactif (Exercices 01-04)
func runInteractiveMode(player *core.Player) {
	encounter := core.NewEncounter()
	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Println("=== Mode Interactif (Exercices 01-04) ===")
	fmt.Println("Commandes: 'rencontre' pour démarrer, 'statut' pour voir vos stats, 'quit' pour quitter")
	
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		
		command := strings.TrimSpace(strings.ToLower(scanner.Text()))
		
		switch command {
		case "rencontre", "r":
			handleEncounter(encounter, player, scanner)
		case "statut", "s":
			displayPlayerStatus(*player)
		case "quit", "q":
			fmt.Println("Retour au menu principal...")
			return
		case "help", "h":
			fmt.Println("Commandes disponibles:")
			fmt.Println("  rencontre (r) - Démarrer une nouvelle rencontre")
			fmt.Println("  statut (s)    - Afficher vos statistiques") 
			fmt.Println("  help (h)      - Afficher cette aide")
			fmt.Println("  quit (q)      - Retour au menu principal")
		default:
			fmt.Println("Commande inconnue. Tapez 'help' pour voir les commandes.")
		}
	}
}

// runConcurrentMode lance le mode concurrent (Exercice 05)
func runConcurrentMode(humanPlayer *core.Player, duration time.Duration) {
	fmt.Println("=== Mode Concurrent (Exercice 05) ===")
	fmt.Printf("Durée de la démo: %v\n", duration)
	fmt.Println("Spawning des WordMon toutes les 3 secondes...")
	fmt.Println("Timeout par combat: 5 secondes")
	fmt.Println("Appuyez sur Ctrl+C pour arrêter prématurément")
	fmt.Println()

	// Créer le contexte avec timeout pour l'arrêt propre
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	
	// Channels centraux
	spawnCh := make(chan core.Word, 10)
	battleCh := make(chan PlayerAttempt, 100)
	
	// Créer 3 joueurs IA
	aiPlayers := []AIPlayer{
		{ID: "alice", Name: "Alice", SkillLevel: 0.7, ResponseTime: 1 * time.Second},
		{ID: "bob", Name: "Bob", SkillLevel: 0.5, ResponseTime: 2 * time.Second},
		{ID: "charlie", Name: "Charlie", SkillLevel: 0.9, ResponseTime: 500 * time.Millisecond},
	}
	
	fmt.Printf("Joueurs IA participants:\n")
	for _, p := range aiPlayers {
		fmt.Printf("  - %s (compétence: %.0f%%, vitesse: %v)\n", 
			p.Name, p.SkillLevel*100, p.ResponseTime)
	}
	fmt.Println()
	
	// Démarrer le spawner concurrent (toutes les 3 secondes pour la démo)
	go StartSpawner(ctx, spawnCh, 3*time.Second)
	
	// Démarrer les joueurs IA
	for _, player := range aiPlayers {
		go player.StartPlayer(ctx, spawnCh, battleCh)
	}
	
	// Démarrer le gestionnaire de combat
	go BattleCoordinator(ctx, battleCh, 5*time.Second)
	
	// Gérer l'arrêt propre avec Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case <-ctx.Done():
		fmt.Println("\n[Système] Fin de la démo concurrent...")
	case <-sigChan:
		fmt.Println("\n[Système] Arrêt demandé par l'utilisateur...")
		cancel()
	}
	
	// Attendre un peu pour que tous les goroutines se terminent proprement
	time.Sleep(500 * time.Millisecond)
	
	// Afficher les statistiques finales
	fmt.Println("\n=== Statistiques Finales ===")
	for _, aiPlayer := range aiPlayers {
		fmt.Printf("  %s: Tentatives effectuées\n", aiPlayer.Name)
	}
	
	fmt.Println("[Système] Retour au menu principal...")
}

// Types pour l'exercice 05
type AIPlayer struct {
	ID           string
	Name         string
	SkillLevel   float64
	ResponseTime time.Duration
}

type PlayerAttempt struct {
	PlayerName string
	Answer     string
	Word       core.Word
	Timestamp  time.Time
}

// StartSpawner implémente le spawner concurrent (Exercice 05)
func StartSpawner(ctx context.Context, ch chan<- core.Word, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	fmt.Printf("[Spawner] Démarrage - spawn toutes les %v\n", interval)
	
	for {
		select {
		case <-ctx.Done():
			fmt.Println("[Spawner] Arrêt propre du spawner")
			close(ch)
			return
		case <-ticker.C:
			word := core.SpawnWord()
			fmt.Printf("[Spawner] Un WordMon apparaît: \"%s\" (%s, +%d XP)\n", 
				word.Text, word.Rarity, word.Points)
			
			select {
			case ch <- word:
				// Envoyé avec succès
			case <-ctx.Done():
				fmt.Println("[Spawner] Arrêt propre du spawner")
				close(ch)
				return
			}
		}
	}
}

// StartPlayer démarre un joueur IA
func (p AIPlayer) StartPlayer(ctx context.Context, spawnCh <-chan core.Word, battleCh chan<- PlayerAttempt) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] Joueur arrêté\n", p.Name)
			return
		case word, ok := <-spawnCh:
			if !ok {
				fmt.Printf("[%s] Canal fermé\n", p.Name)
				return
			}
			
			// Décider de participer (90% de chance)
			if rand.Float64() > 0.9 {
				fmt.Printf("[%s] ignore \"%s\"\n", p.Name, word.Text)
				continue
			}
			
			// Lancer la tentative en arrière-plan
			go p.AttemptCapture(ctx, word, battleCh)
		}
	}
}

// AttemptCapture tente de capturer un WordMon
func (p AIPlayer) AttemptCapture(ctx context.Context, word core.Word, battleCh chan<- PlayerAttempt) {
	// Simuler le temps de réflexion
	delay := p.ResponseTime + time.Duration(rand.Intn(500))*time.Millisecond
	
	select {
	case <-ctx.Done():
		return
	case <-time.After(delay):
		// Générer une réponse
		answer := p.GenerateAnswer(word.Text)
		
		attempt := PlayerAttempt{
			PlayerName: p.Name,
			Answer:     answer,
			Word:       word,
			Timestamp:  time.Now(),
		}
		
		fmt.Printf("[%s] tente une capture avec réponse: \"%s\"\n", 
			p.Name, answer)
		
		select {
		case battleCh <- attempt:
			// Tentative envoyée
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			fmt.Printf("[%s] trop tard... le WordMon s'est échappé\n", p.Name)
		}
	}
}

// GenerateAnswer génère une réponse selon le niveau de compétence
func (p AIPlayer) GenerateAnswer(word string) string {
	if rand.Float64() < p.SkillLevel {
		// Bonne réponse - anagramme correct
		return GenerateAnagram(word)
	} else {
		// Mauvaise réponse
		return "wrong_answer"
	}
}

// GenerateAnagram génère un anagramme simple
func GenerateAnagram(word string) string {
	runes := []rune(word)
	
	// Mélanger
	for i := len(runes) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	
	result := string(runes)
	
	// Éviter de retourner le mot original
	if result == word && len(word) > 1 {
		// Inverser les deux premiers caractères
		runes[0], runes[1] = runes[1], runes[0]
		result = string(runes)
	}
	
	return result
}

// BattleCoordinator gère les combats avec select (Exercice 05)
func BattleCoordinator(ctx context.Context, battleCh <-chan PlayerAttempt, timeout time.Duration) {
	fmt.Printf("[BattleCoordinator] Démarrage - timeout: %v\n", timeout)
	
	// Map pour tracker les combats en cours
	activeBattles := make(map[string]bool) // wordID -> combat en cours
	
	for {
		select {
		case <-ctx.Done():
			fmt.Println("[BattleCoordinator] Arrêt du coordinateur")
			return
		case attempt, ok := <-battleCh:
			if !ok {
				fmt.Println("[BattleCoordinator] Canal fermé")
				return
			}
			
			wordID := attempt.Word.ID
			
			// Vérifier si un combat est déjà résolu pour ce WordMon
			if activeBattles[wordID] {
				fmt.Printf("[%s] trop tard... le WordMon s'est échappé\n", attempt.PlayerName)
				continue
			}
			
			// Marquer ce combat comme en cours
			activeBattles[wordID] = true
			
			// Traiter le combat
			go HandleBattle(attempt, timeout)
		}
	}
}

// HandleBattle traite un combat individuel
func HandleBattle(attempt PlayerAttempt, timeout time.Duration) {
	fmt.Printf("[Battle] Combat pour \"%s\" - %s tente sa chance\n", 
		attempt.Word.Text, attempt.PlayerName)
	
	// Créer le challenge
	challenge := core.NewAnagramChallenge(attempt.Word)
	
	// Vérifier la réponse
	isCorrect, err := challenge.Check(attempt.Answer)
	
	if err != nil {
		fmt.Printf("[Résultat] Erreur: %v\n", err)
		return
	}
	
	if isCorrect {
		fmt.Printf("[Résultat] %s capture \"%s\" ! (+%d XP)\n", 
			attempt.PlayerName, attempt.Word.Text, attempt.Word.Points)
	} else {
		fmt.Printf("[Résultat] Mauvaise tentative! \"%s\" s'enfuit...\n", attempt.Word.Text)
	}
	
	fmt.Println("---")
}

// handleEncounter gère une rencontre complète (Exercices 01-04)
func handleEncounter(encounter *core.Encounter, player *core.Player, scanner *bufio.Scanner) {
	// Démarrer la rencontre
	if err := encounter.Start(player); err != nil {
		fmt.Printf("Erreur lors du démarrage: %v\n", err)
		return
	}
	
	// Afficher les logs
	for _, log := range encounter.GetBattleLog() {
		fmt.Println(log)
	}
	
	// Commencer le combat
	if err := encounter.BeginBattle(); err != nil {
		fmt.Printf("Erreur lors du début du combat: %v\n", err)
		return
	}
	
	// Afficher les nouvelles instructions
	logs := encounter.GetBattleLog()
	if len(logs) > 0 {
		fmt.Println(logs[len(logs)-1])
	}
	
	// Boucle de combat
	for encounter.GetCurrentState() == core.IN_BATTLE {
		fmt.Print("Votre tentative: ")
		if !scanner.Scan() {
			break
		}
		
		attempt := strings.TrimSpace(scanner.Text())
		if attempt == "" {
			fmt.Println("Veuillez entrer une tentative valide.")
			continue
		}
		
		if err := encounter.SubmitAttempt(attempt); err != nil {
			fmt.Printf("Erreur: %v\n", err)
			continue
		}
		
		// Afficher le dernier log
		logs = encounter.GetBattleLog()
		if len(logs) > 0 {
			fmt.Println(logs[len(logs)-1])
		}
	}
	
	// Résoudre la rencontre
	if err := encounter.Resolve(); err != nil {
		fmt.Printf("Erreur lors de la résolution: %v\n", err)
		return
	}
	
	// Afficher le résultat final
	logs = encounter.GetBattleLog()
	if len(logs) > 0 {
		fmt.Println(logs[len(logs)-1])
	}
}

// runTestMode lance les scénarios de test automatiques
func runTestMode(player *core.Player) {
	fmt.Println("=== Mode Test Automatique (Tous exercices) ===")
	
	// Test 1: Exercice 02 - Types et Spawning
	fmt.Println("\n1. Test: Types et Spawning (Exercice 02)")
	testSpawning()
	
	// Test 2: Exercice 03 - Combat simple
	fmt.Println("\n2. Test: Combat simple (Exercice 03)")
	testCombat(player)
	
	// Test 3: Exercice 04 - Gestion d'erreurs
	fmt.Println("\n3. Test: Gestion d'erreurs (Exercice 04)")
	testErrorHandling(player)
	
	// Test 4: Exercice 05 - Concurrence (court)
	fmt.Println("\n4. Test: Concurrence (Exercice 05)")
	testConcurrency()
	
	fmt.Println("\n=== Tests terminés ===")
}

func testSpawning() {
	fmt.Println("  Test de spawn de 5 WordMon...")
	for i := 0; i < 5; i++ {
		word := core.SpawnWord()
		fmt.Printf("    %d. %s (%s, %d points)\n", i+1, word.Text, word.Rarity, word.Points)
	}
}

func testCombat(player *core.Player) {
	fmt.Println("  Test de combat automatique...")
	encounter := core.NewEncounter()
	
	if err := encounter.Start(player); err != nil {
		fmt.Printf("    Erreur Start: %v\n", err)
		return
	}
	
	if err := encounter.BeginBattle(); err != nil {
		fmt.Printf("    Erreur BeginBattle: %v\n", err)
		return
	}
	
	// Obtenir le mot et créer un anagramme simple
	currentMon := encounter.GetCurrentMon()
	if currentMon != nil {
		word := currentMon.Word.Text
		anagram := GenerateAnagram(word)
		fmt.Printf("    Mot: %s, Anagramme généré: %s\n", word, anagram)
		
		if err := encounter.SubmitAttempt(anagram); err != nil {
			fmt.Printf("    Erreur SubmitAttempt: %v\n", err)
		}
	}
	
	if err := encounter.Resolve(); err != nil {
		fmt.Printf("    Erreur Resolve: %v\n", err)
	}
	
	fmt.Printf("    Résultat: XP = %d, Niveau = %d\n", player.XP, player.Level)
}

func testErrorHandling(player *core.Player) {
	fmt.Println("  Test de gestion d'erreurs...")
	
	// Test transition invalide
	encounter := core.NewEncounter()
	err := encounter.BeginBattle() // Sans Start()
	if err != nil {
		fmt.Printf("    Erreur détectée correctement: %v\n", err)
	}
	
	// Test XP négative
	err = player.AwardXP(-10)
	if err != nil {
		fmt.Printf("    XP négative rejetée: %v\n", err)
	}
}

func testConcurrency() {
	fmt.Println("  Test de concurrence (5 secondes)...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	spawnCh := make(chan core.Word, 5)
	battleCh := make(chan PlayerAttempt, 10)
	
	// Spawner rapide pour test
	go func() {
		for i := 0; i < 3; i++ {
			word := core.SpawnWord()
			fmt.Printf("    [Test] WordMon: %s\n", word.Text)
			spawnCh <- word
			time.Sleep(1 * time.Second)
		}
		close(spawnCh)
	}()
	
	// Joueur test
	testPlayer := AIPlayer{
		ID: "test", Name: "TestBot", SkillLevel: 0.8, ResponseTime: 200 * time.Millisecond,
	}
	go testPlayer.StartPlayer(ctx, spawnCh, battleCh)
	
	// Battle coordinator simple
	go func() {
		for attempt := range battleCh {
			fmt.Printf("    [Test] %s répond: %s\n", attempt.PlayerName, attempt.Answer)
		}
	}()
	
	<-ctx.Done()
	fmt.Println("    Test de concurrence terminé")
}

// displayPlayerStatus affiche l'état complet du joueur
func displayPlayerStatus(p core.Player) {
	fmt.Printf("=== Statut du Dresseur %s ===\n", p.Name)
	fmt.Printf("XP: %d | Niveau: %d\n", p.XP, p.Level)
	fmt.Printf("Pokédex: %d mots différents | Total captures: %d\n", 
		p.GetInventorySize(), p.GetTotalCaptures())
	
	if len(p.Inventory) == 0 {
		fmt.Println("Inventaire: (vide)")
	} else {
		fmt.Println("Inventaire:")
		for word, count := range p.Inventory {
			fmt.Printf("  - %s: %d\n", word, count)
		}
	}
}
