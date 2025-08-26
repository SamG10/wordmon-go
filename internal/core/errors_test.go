package core

import (
	"errors"
	"testing"
)

// TestInvalidStateTransitions teste les transitions d'état invalides
func TestInvalidStateTransitions(t *testing.T) {
	encounter := NewEncounter()
	player := NewPlayer("test", "TestPlayer")

	// Test 1: BeginBattle sans Start
	err := encounter.BeginBattle()
	if err == nil {
		t.Error("BeginBattle devrait échouer sans Start")
	}
	var stateErr InvalidStateError
	if !errors.As(err, &stateErr) {
		t.Errorf("Erreur attendue: InvalidStateError, reçu: %T", err)
	}

	// Test 2: SubmitAttempt sans BeginBattle
	err = encounter.SubmitAttempt("test")
	if err == nil {
		t.Error("SubmitAttempt devrait échouer sans BeginBattle")
	}
	if !errors.As(err, &stateErr) {
		t.Errorf("Erreur attendue: InvalidStateError, reçu: %T", err)
	}

	// Test 3: Resolve sans combat
	err = encounter.Resolve()
	if err == nil {
		t.Error("Resolve devrait échouer sans combat")
	}
	if !errors.As(err, &stateErr) {
		t.Errorf("Erreur attendue: InvalidStateError, reçu: %T", err)
	}
}

// TestInvalidAttempts teste les tentatives invalides
func TestInvalidAttempts(t *testing.T) {
	encounter := NewEncounter()
	player := NewPlayer("test", "TestPlayer")

	// Initialiser une rencontre valide
	encounter.Start(&player)
	encounter.BeginBattle()

	// Test 1: Tentative vide
	err := encounter.SubmitAttempt("")
	if err == nil {
		t.Error("SubmitAttempt devrait échouer avec entrée vide")
	}
	var attemptErr InvalidAttemptError
	if !errors.As(err, &attemptErr) {
		t.Errorf("Erreur attendue: InvalidAttemptError, reçu: %T", err)
	}

	// Test 2: Tentative avec caractères invalides
	err = encounter.SubmitAttempt("123")
	if err == nil {
		t.Error("SubmitAttempt devrait échouer avec caractères non-alphabétiques")
	}
	if !errors.As(err, &attemptErr) {
		t.Errorf("Erreur attendue: InvalidAttemptError, reçu: %T", err)
	}
}

// TestCaptureErrors teste les erreurs de capture
func TestCaptureErrors(t *testing.T) {
	player := NewPlayer("test", "TestPlayer")

	// Test: Capture d'un mot vide
	emptyWord := Word{ID: "empty", Text: "", Rarity: Common, Points: 5}
	err := player.Capture(emptyWord)
	if err == nil {
		t.Error("Capture devrait échouer avec mot vide")
	}
	var captureErr CaptureError
	if !errors.As(err, &captureErr) {
		t.Errorf("Erreur attendue: CaptureError, reçu: %T", err)
	}
}

// TestXPErrors teste les erreurs d'XP
func TestXPErrors(t *testing.T) {
	player := NewPlayer("test", "TestPlayer")

	// Test: XP négative
	err := player.AwardXP(-10)
	if err == nil {
		t.Error("AwardXP devrait échouer avec points négatifs")
	}
	var xpErr XPError
	if !errors.As(err, &xpErr) {
		t.Errorf("Erreur attendue: XPError, reçu: %T", err)
	}
}

// TestWordMonPanic teste les panics pour données corrompues
func TestWordMonPanic(t *testing.T) {
	// Test 1: Mot vide (doit paniquer)
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewWordMon devrait paniquer avec mot vide")
		}
	}()

	emptyWord := Word{ID: "empty", Text: "", Rarity: Common, Points: 5}
	NewWordMon(emptyWord)
}

// TestWordMonPanicInvalidRarity teste le panic avec rareté invalide
func TestWordMonPanicInvalidRarity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewWordMon devrait paniquer avec rareté invalide")
		}
	}()

	invalidWord := Word{ID: "invalid", Text: "test", Rarity: "INVALID", Points: 5}
	NewWordMon(invalidWord)
}

// TestWordMonPanicInvalidPoints teste le panic avec points invalides
func TestWordMonPanicInvalidPoints(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewWordMon devrait paniquer avec points invalides")
		}
	}()

	invalidWord := Word{ID: "invalid", Text: "test", Rarity: Common, Points: 0}
	NewWordMon(invalidWord)
}
