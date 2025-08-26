package core

// LevelFromXP calcule le niveau à partir de l'XP.
// Palier simple: niveau = 1 + XP/100
func LevelFromXP(xp int) int {
	if xp < 0 {
		return 1 // Niveau minimum
	}
	return 1 + xp/100
}

// AwardXP ajoute des points d'XP au joueur et met à jour son niveau.
func AwardXP(p *Player, points int) {
	// Validation: pas de points négatifs
	if points < 0 {
		return
	}
	
	// Ajouter les points d'XP
	p.XP += points
	
	// Mettre à jour le niveau
	p.Level = LevelFromXP(p.XP)
}

// Capture ajoute le mot à l'inventaire du joueur (count++) et retourne les points gagnés.
// La capture réussit toujours (mini-jeu viendra plus tard).
func Capture(p *Player, w Word) (gained int) {
	// Initialiser l'inventaire si nil
	if p.Inventory == nil {
		p.Inventory = make(map[string]int)
	}
	
	// Ajouter le mot à l'inventaire (ou incrémenter le count)
	p.Inventory[w.Text]++
	
	// Donner l'XP au joueur
	AwardXP(p, w.Points)
	
	// Retourner les points gagnés
	return w.Points
}

// Méthodes pour Player (récepteurs pointeur)

// Capture ajoute un WordMon à l'inventaire du joueur
func (p *Player) Capture(word Word) error {
	// Validation des données
	if word.Text == "" {
		return CaptureError{
			Word:   word.Text,
			Reason: "mot vide",
		}
	}
	
	if p.Inventory == nil {
		p.Inventory = make(map[string]int)
	}
	p.Inventory[word.Text]++
	return nil
}

// AwardXP ajoute de l'XP au joueur et met à jour son niveau
func (p *Player) AwardXP(points int) error {
	if points < 0 {
		return XPError{
			Points: points,
			Reason: "points négatifs interdits",
		}
	}
	p.XP += points
	p.Level = LevelFromXP(p.XP)
	return nil
}

// GetInventorySize retourne le nombre de mots différents dans l'inventaire
func (p *Player) GetInventorySize() int {
	return len(p.Inventory)
}

// GetTotalCaptures retourne le nombre total de captures
func (p *Player) GetTotalCaptures() int {
	total := 0
	for _, count := range p.Inventory {
		total += count
	}
	return total
}
