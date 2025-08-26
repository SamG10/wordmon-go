package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Attempt représente une tentative de capture d'un joueur
type Attempt struct {
	PlayerID string
	Player   *Player
	Answer   string
	Word     Word
}

// BattleResult représente le résultat d'un combat
type BattleResult struct {
	Success   bool
	Winner    *Player
	Word      Word
	Message   string
}

// Spawner gère l'apparition et les combats des WordMon
type Spawner struct {
	spawnCh   chan Word
	battleCh  chan Attempt
	resultCh  chan BattleResult
	players   []*Player
	timeout   time.Duration
	mutex     sync.Mutex
}

// NewSpawner crée un nouveau spawner
func NewSpawner(players []*Player, timeout time.Duration) *Spawner {
	return &Spawner{
		spawnCh:  make(chan Word, 10),
		battleCh: make(chan Attempt, 100),
		resultCh: make(chan BattleResult, 10),
		players:  players,
		timeout:  timeout,
	}
}

// StartSpawner lance le processus de spawn des WordMon
func (s *Spawner) StartSpawner(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Println("[Spawner] Démarrage du spawner de WordMon...")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[Spawner] Arrêt du spawner...")
			close(s.spawnCh)
			return
		case <-ticker.C:
			word := SpawnWord()
			fmt.Printf("[Spawner] Un WordMon apparaît: \"%s\" (%s, +%d XP)\n", 
				word.Text, word.Rarity, word.Points)
			
			select {
			case s.spawnCh <- word:
				// WordMon envoyé avec succès
			case <-ctx.Done():
				fmt.Println("[Spawner] Arrêt du spawner...")
				close(s.spawnCh)
				return
			}
		}
	}
}

// StartBattleManager gère les combats pour chaque WordMon qui apparaît
func (s *Spawner) StartBattleManager(ctx context.Context) {
	fmt.Println("[BattleManager] Démarrage du gestionnaire de combats...")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[BattleManager] Arrêt du gestionnaire de combats...")
			close(s.battleCh)
			close(s.resultCh)
			return
		case word, ok := <-s.spawnCh:
			if !ok {
				fmt.Println("[BattleManager] Canal de spawn fermé, arrêt...")
				close(s.battleCh)
				close(s.resultCh)
				return
			}
			go s.handleBattle(ctx, word)
		}
	}
}

// handleBattle gère un combat individuel avec timeout et premier arrivé
func (s *Spawner) handleBattle(ctx context.Context, word Word) {
	battleTimeout := time.After(s.timeout)
	
	fmt.Printf("[Battle] Combat ouvert pour \"%s\" - timeout dans %v\n", 
		word.Text, s.timeout)

	select {
	case <-ctx.Done():
		return
	case attempt := <-s.battleCh:
		// Première tentative reçue
		result := s.processBattleAttempt(attempt, word)
		s.resultCh <- result
		
		// Vider les autres tentatives en attente pour ce WordMon
		go s.drainAttempts(word)
		
	case <-battleTimeout:
		// Timeout - le WordMon s'échappe
		result := BattleResult{
			Success: false,
			Winner:  nil,
			Word:    word,
			Message: fmt.Sprintf("Personne n'a répondu à temps... \"%s\" disparaît", word.Text),
		}
		s.resultCh <- result
	}
}

// processBattleAttempt traite une tentative de capture
func (s *Spawner) processBattleAttempt(attempt Attempt, word Word) BattleResult {
	challenge := NewAnagramChallenge(word)
	
	isCorrect, err := challenge.Check(attempt.Answer)
	if err != nil {
		return BattleResult{
			Success: false,
			Winner:  nil,
			Word:    word,
			Message: fmt.Sprintf("Erreur de validation: %v", err),
		}
	}
	
	if isCorrect {
		// Bonne réponse - capture réussie
		s.mutex.Lock()
		err := attempt.Player.Capture(word)
		s.mutex.Unlock()
		
		if err != nil {
			return BattleResult{
				Success: false,
				Winner:  nil,
				Word:    word,
				Message: fmt.Sprintf("Erreur lors de la capture: %v", err),
			}
		}
		
		return BattleResult{
			Success: true,
			Winner:  attempt.Player,
			Word:    word,
			Message: fmt.Sprintf("%s capture \"%s\" ! XP +%d, niveau %d", 
				attempt.Player.Name, word.Text, word.Points, attempt.Player.Level),
		}
	} else {
		// Mauvaise réponse - le WordMon fuit
		return BattleResult{
			Success: false,
			Winner:  nil,
			Word:    word,
			Message: fmt.Sprintf("[%s] tente une capture avec réponse: \"%s\" (anagramme incorrecte)\nMauvaise tentative! \"%s\" s'enfuit...", 
				attempt.Player.Name, attempt.Answer, word.Text),
		}
	}
}

// drainAttempts vide les tentatives en trop pour éviter les blocages
func (s *Spawner) drainAttempts(word Word) {
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case attempt := <-s.battleCh:
			fmt.Printf("[%s] trop tard... le WordMon s'est échappé\n", attempt.Player.Name)
		case <-timeout:
			return
		}
	}
}

// GetSpawnChannel retourne le canal de spawn pour les joueurs
func (s *Spawner) GetSpawnChannel() <-chan Word {
	return s.spawnCh
}

// GetBattleChannel retourne le canal de bataille pour envoyer des tentatives
func (s *Spawner) GetBattleChannel() chan<- Attempt {
	return s.battleCh
}

// GetResultChannel retourne le canal des résultats
func (s *Spawner) GetResultChannel() <-chan BattleResult {
	return s.resultCh
}
