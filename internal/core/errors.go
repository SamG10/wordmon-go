package core

import "fmt"

// Types d'erreurs personnalisées

// InvalidStateError représente une transition d'état interdite
type InvalidStateError struct {
	From     EncounterState
	Expected EncounterState
	Action   string
}

func (e InvalidStateError) Error() string {
	return fmt.Sprintf("transition interdite: état=%s, attendu=%s pour %s", 
		e.From, e.Expected, e.Action)
}

// InvalidAttemptError représente une tentative invalide du joueur
type InvalidAttemptError struct {
	Input  string
	Reason string
}

func (e InvalidAttemptError) Error() string {
	if e.Input == "" {
		return fmt.Sprintf("tentative invalide: entrée vide (%s)", e.Reason)
	}
	return fmt.Sprintf("tentative invalide: '%s' (%s)", e.Input, e.Reason)
}

// CaptureError représente une erreur lors de la capture
type CaptureError struct {
	Word   string
	Reason string
}

func (e CaptureError) Error() string {
	return fmt.Sprintf("impossible de capturer '%s' (%s)", e.Word, e.Reason)
}

// XPError représente une erreur liée aux points d'expérience
type XPError struct {
	Points int
	Reason string
}

func (e XPError) Error() string {
	return fmt.Sprintf("erreur XP: %d points (%s)", e.Points, e.Reason)
}

// ChallengeError représente une erreur de défi
type ChallengeError struct {
	Input  string
	Reason string
}

func (e ChallengeError) Error() string {
	return fmt.Sprintf("erreur de défi: '%s' (%s)", e.Input, e.Reason)
}
