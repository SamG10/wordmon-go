// Package core contient les types et fonctions principaux du jeu WordMon.
package core

import "fmt"

// EncounterState représente l'état d'une rencontre WordMon.
type EncounterState string

const (
	IDLE        EncounterState = "IDLE"
	ENCOUNTERED EncounterState = "ENCOUNTERED"
	IN_BATTLE   EncounterState = "IN_BATTLE"
	WON         EncounterState = "WON"
	CAPTURED    EncounterState = "CAPTURED"
	LOST        EncounterState = "LOST"
	FLED        EncounterState = "FLED"
)

// Encounter orchestre une rencontre WordMon avec une machine à états.
type Encounter struct {
	State      EncounterState
	CurrentMon *WordMon
	Player     *Player
	BattleLog  []string
}

// NewEncounter crée une nouvelle instance de rencontre WordMon.
func NewEncounter() *Encounter {
	return &Encounter{
		State:     IDLE,
		BattleLog: make([]string, 0),
	}
}

// Start démarre une nouvelle rencontre (IDLE → ENCOUNTERED).
func (e *Encounter) Start(player *Player) error {
	if e.State != IDLE {
		return InvalidStateError{
			From:     e.State,
			Expected: IDLE,
			Action:   "Start",
		}
	}

	// Spawn un WordMon
	word := SpawnWord()
	wordmon := NewWordMon(word)

	e.CurrentMon = &wordmon
	e.Player = player
	e.State = ENCOUNTERED
	e.BattleLog = make([]string, 0)

	e.addLog(e.CurrentMon.Presentation())
	return nil
}

// BeginBattle lance le combat (ENCOUNTERED → IN_BATTLE).
func (e *Encounter) BeginBattle() error {
	if e.State != ENCOUNTERED {
		return InvalidStateError{
			From:     e.State,
			Expected: ENCOUNTERED,
			Action:   "BeginBattle",
		}
	}

	e.State = IN_BATTLE
	instructions := e.CurrentMon.GetChallenge().Instructions()
	e.addLog(instructions)

	return nil
}

// SubmitAttempt soumet une tentative de réponse du joueur.
func (e *Encounter) SubmitAttempt(input string) error {
	if e.State != IN_BATTLE {
		return InvalidStateError{
			From:     e.State,
			Expected: IN_BATTLE,
			Action:   "SubmitAttempt",
		}
	}

	if input == "" {
		return InvalidAttemptError{
			Input:  input,
			Reason: "entrée vide",
		}
	}

	challenge := e.CurrentMon.GetChallenge()

	success, err := challenge.Check(input)
	if err != nil {
		e.addLog(fmt.Sprintf("Tentative '%s' → ERREUR: %s", input, err.Error()))
		return fmt.Errorf("erreur de tentative: %w", err)
	}

	if success {
		e.State = WON
		e.addLog(fmt.Sprintf("Tentative '%s' → VICTOIRE !", input))
		return nil
	}

	// Vérifier s'il reste des essais
	if anagramChallenge, ok := challenge.(*AnagramChallenge); ok {
		if anagramChallenge.HasAttemptsLeft() {
			remaining := anagramChallenge.GetMaxAttempts() - anagramChallenge.GetCurrentTries()
			e.addLog(fmt.Sprintf("Tentative '%s' → ÉCHEC... Il reste %d essai(s)", input, remaining))
			return nil
		}
	}

	// Plus d'essais
	e.State = LOST
	e.addLog(fmt.Sprintf("Tentative '%s' → ÉCHEC FINAL ! Plus d'essais...", input))
	return nil
}

// Resolve résout la rencontre selon l'état actuel (victoire ou défaite).
func (e *Encounter) Resolve() error {
	switch e.State {
	case WON:
		return e.resolveVictory()
	case LOST:
		return e.resolveDefeat()
	default:
		return InvalidStateError{
			From:     e.State,
			Expected: WON,
			Action:   "Resolve (attendu WON ou LOST)",
		}
	}
}

// resolveVictory gère la victoire (WON → CAPTURED → IDLE)
func (e *Encounter) resolveVictory() error {
	// Capturer le WordMon
	if err := e.Player.Capture(e.CurrentMon.Word); err != nil {
		return fmt.Errorf("erreur lors de la capture: %w", err)
	}

	// Donner l'XP
	if err := e.Player.AwardXP(e.CurrentMon.Word.Points); err != nil {
		return fmt.Errorf("erreur lors de l'attribution d'XP: %w", err)
	}

	e.addLog(fmt.Sprintf("Capture réussie : inventaire +1 ('%s'), XP +%d, niveau = %d",
		e.CurrentMon.Word.Text, e.CurrentMon.Word.Points, e.Player.Level))

	e.State = CAPTURED

	// Retour à IDLE
	e.reset()
	return nil
}

// resolveDefeat gère la défaite (LOST → FLED → IDLE)
func (e *Encounter) resolveDefeat() error {
	e.addLog("Le WordMon s'échappe dans un nuage de lettres...")
	e.State = FLED

	// Retour à IDLE
	e.reset()
	return nil
}

// reset remet la rencontre à l'état initial
func (e *Encounter) reset() {
	e.State = IDLE
	e.CurrentMon = nil
	e.Player = nil
}

// addLog ajoute un message au log de bataille
func (e *Encounter) addLog(message string) {
	e.BattleLog = append(e.BattleLog, message)
}

// GetBattleLog retourne le log complet de la bataille.
func (e *Encounter) GetBattleLog() []string {
	return e.BattleLog
}

// GetCurrentState retourne l'état actuel de la rencontre.
func (e *Encounter) GetCurrentState() EncounterState {
	return e.State
}

// GetCurrentMon retourne le WordMon actuellement rencontré (peut être nil).
func (e *Encounter) GetCurrentMon() *WordMon {
	return e.CurrentMon
}
