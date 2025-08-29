package core

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// AIPlayer représente un joueur IA qui participe automatiquement aux combats
type AIPlayer struct {
	*Player
	participationRate float64 // Probabilité de participer (0.0 à 1.0)
	responseTime      time.Duration
	skillLevel        float64 // Probabilité de donner la bonne réponse (0.0 à 1.0)
}

// NewAIPlayer crée un nouveau joueur IA
func NewAIPlayer(name string, participationRate, skillLevel float64, responseTime time.Duration) *AIPlayer {
	player := NewPlayer(fmt.Sprintf("ai_%s", name), name)
	return &AIPlayer{
		Player:            &player,
		participationRate: participationRate,
		responseTime:      responseTime,
		skillLevel:        skillLevel,
	}
}

// StartAIPlayer lance la goroutine du joueur IA
func (ai *AIPlayer) StartAIPlayer(ctx context.Context, spawnCh <-chan Word, battleCh chan<- Attempt) {
	fmt.Printf("[%s] Joueur IA démarré (participation: %.0f%%, skill: %.0f%%)\n",
		ai.Name, ai.participationRate*100, ai.skillLevel*100)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] Joueur IA arrêté\n", ai.Name)
			return
		case word, ok := <-spawnCh:
			if !ok {
				fmt.Printf("[%s] Canal de spawn fermé, arrêt du joueur IA\n", ai.Name)
				return
			}

			// Décider si le joueur participe
			if rand.Float64() > ai.participationRate {
				fmt.Printf("[%s] ignore le WordMon \"%s\"\n", ai.Name, word.Text)
				continue
			}

			// Simuler le temps de réflexion
			go ai.attemptCapture(ctx, word, battleCh)
		}
	}
}

// attemptCapture simule une tentative de capture avec délai de réponse
func (ai *AIPlayer) attemptCapture(ctx context.Context, word Word, battleCh chan<- Attempt) {
	// Temps de réaction variable (±25%)
	variation := time.Duration(float64(ai.responseTime) * (0.5 - rand.Float64()) * 0.5)
	responseDelay := ai.responseTime + variation

	select {
	case <-ctx.Done():
		return
	case <-time.After(responseDelay):
		// Générer une réponse
		answer := ai.generateAnswer(word)

		attempt := Attempt{
			PlayerID: ai.ID,
			Player:   ai.Player,
			Answer:   answer,
			Word:     word,
		}

		fmt.Printf("[%s] tente de capturer \"%s\" avec: \"%s\"\n",
			ai.Name, word.Text, answer)

		select {
		case battleCh <- attempt:
			// Tentative envoyée
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			// Timeout - le combat est probablement terminé
			fmt.Printf("[%s] tentative trop tardive pour \"%s\"\n", ai.Name, word.Text)
		}
	}
}

// generateAnswer génère une réponse (anagramme) selon le niveau de skill
func (ai *AIPlayer) generateAnswer(word Word) string {
	if rand.Float64() < ai.skillLevel {
		// Bonne réponse - générer un anagramme correct
		return ai.generateCorrectAnagram(word.Text)
	} else {
		// Mauvaise réponse - générer une réponse incorrecte
		return ai.generateIncorrectAnswer(word.Text)
	}
}

// generateCorrectAnagram génère un anagramme correct
func (ai *AIPlayer) generateCorrectAnagram(text string) string {
	runes := []rune(text)

	// Mélanger les lettres plusieurs fois pour créer un anagramme
	for i := 0; i < 10; i++ {
		for j := len(runes) - 1; j > 0; j-- {
			k := rand.Intn(j + 1)
			runes[j], runes[k] = runes[k], runes[j]
		}
	}

	result := string(runes)

	// S'assurer qu'on ne retourne pas le mot original
	if result == text && len(text) > 1 {
		// Inverser simplement les deux premières lettres
		if len(runes) >= 2 {
			runes[0], runes[1] = runes[1], runes[0]
			result = string(runes)
		}
	}

	return result
}

// generateIncorrectAnswer génère une réponse incorrecte
func (ai *AIPlayer) generateIncorrectAnswer(text string) string {
	incorrectAnswers := []string{
		strings.ToUpper(text),       // Tout en majuscules
		text + "x",                  // Ajouter une lettre
		text[:len(text)-1],          // Enlever la dernière lettre
		strings.Repeat(text[:1], 3), // Répéter la première lettre
		"wrong",                     // Réponse générique
		text + text[:1],             // Dupliquer la première lettre
	}

	// Filtrer les réponses qui pourraient être correctes
	validIncorrect := make([]string, 0)
	for _, answer := range incorrectAnswers {
		if len(answer) > 0 && answer != text {
			validIncorrect = append(validIncorrect, answer)
		}
	}

	if len(validIncorrect) > 0 {
		return validIncorrect[rand.Intn(len(validIncorrect))]
	}

	return "wrong"
}

// CreateDefaultAIPlayers crée un ensemble de joueurs IA avec des profils variés
func CreateDefaultAIPlayers() []*AIPlayer {
	return []*AIPlayer{
		NewAIPlayer("Alice", 0.8, 0.7, 1*time.Second),          // Participative et compétente
		NewAIPlayer("Bob", 0.6, 0.5, 2*time.Second),            // Modéré
		NewAIPlayer("Charlie", 0.9, 0.3, 500*time.Millisecond), // Très rapide mais peu précis
		NewAIPlayer("Diana", 0.4, 0.9, 3*time.Second),          // Rare mais très précise
	}
}
