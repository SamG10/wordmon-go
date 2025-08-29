// Package core contient les types et fonctions principaux du jeu WordMon.
package core

import (
	"sort"
	"strings"
)

// Challenge définit l'interface pour tous les défis de combat WordMon.
type Challenge interface {
	Instructions() string
	Check(attempt string) (bool, error)
	DifficultyFor(rarity Rarity) int
	GetMaxAttempts() int
}

// AnagramChallenge implémente le défi anagramme pour WordMon.
type AnagramChallenge struct {
	TargetWord   string
	Rarity       Rarity
	MaxAttempts  int
	CurrentTries int
}

// NewAnagramChallenge crée un nouveau défi anagramme pour un mot donné.
func NewAnagramChallenge(word Word) *AnagramChallenge {
	challenge := &AnagramChallenge{
		TargetWord: word.Text,
		Rarity:     word.Rarity,
	}
	challenge.MaxAttempts = challenge.DifficultyFor(word.Rarity)
	return challenge
}

// Instructions retourne la consigne du défi anagramme.
func (ac *AnagramChallenge) Instructions() string {
	return "Défi : Donne un anagramme correct du mot '" + ac.TargetWord + "'"
}

// Check vérifie si la tentative est un anagramme valide.
func (ac *AnagramChallenge) Check(attempt string) (bool, error) {
	ac.CurrentTries++

	// Vérifier que ce n'est pas vide
	if attempt == "" {
		return false, InvalidAttemptError{
			Input:  attempt,
			Reason: "entrée vide",
		}
	}

	// Normaliser l'entrée
	attempt = strings.TrimSpace(strings.ToLower(attempt))
	target := strings.ToLower(ac.TargetWord)

	// Vérifier que l'entrée contient seulement des lettres
	for _, r := range attempt {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false, InvalidAttemptError{
				Input:  attempt,
				Reason: "doit contenir seulement des lettres",
			}
		}
	}

	// Vérifier que ce n'est pas le mot original
	if attempt == target {
		return false, InvalidAttemptError{
			Input:  attempt,
			Reason: "doit être différent du mot original",
		}
	}

	// Vérifier que c'est un anagramme (même lettres)
	if !isAnagram(attempt, target) {
		return false, InvalidAttemptError{
			Input:  attempt,
			Reason: "pas un anagramme valide",
		}
	}

	return true, nil
}

// DifficultyFor retourne le nombre d'essais selon la rareté du mot.
func (ac *AnagramChallenge) DifficultyFor(rarity Rarity) int {
	switch rarity {
	case Common:
		return 3 // 3 essais pour les mots communs
	case Rare:
		return 2 // 2 essais pour les mots rares
	case Legendary:
		return 1 // 1 seul essai pour les légendaires
	default:
		return 3
	}
}

// GetMaxAttempts retourne le nombre maximum d'essais autorisés.
func (ac *AnagramChallenge) GetMaxAttempts() int {
	return ac.MaxAttempts
}

// GetCurrentTries retourne le nombre d'essais déjà effectués.
func (ac *AnagramChallenge) GetCurrentTries() int {
	return ac.CurrentTries
}

// HasAttemptsLeft indique s'il reste des essais disponibles.
func (ac *AnagramChallenge) HasAttemptsLeft() bool {
	return ac.CurrentTries < ac.MaxAttempts
}

// isAnagram vérifie si deux mots sont des anagrammes.
func isAnagram(word1, word2 string) bool {
	if len(word1) != len(word2) {
		return false
	}

	// Convertir en slices de runes et trier
	runes1 := []rune(word1)
	runes2 := []rune(word2)

	sort.Slice(runes1, func(i, j int) bool { return runes1[i] < runes1[j] })
	sort.Slice(runes2, func(i, j int) bool { return runes2[i] < runes2[j] })

	return string(runes1) == string(runes2)
}
