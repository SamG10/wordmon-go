// Package core contient les types et fonctions principaux du jeu WordMon.
package core

// Rarity représente la rareté d'un WordMon (common, rare, legendary).
type Rarity string

// Constantes de rareté pour les WordMon.
const (
	Common    Rarity = "common"
	Rare      Rarity = "rare"
	Legendary Rarity = "legendary"
)

// Player représente un joueur dans WordMon.
type Player struct {
	ID        string
	Name      string
	XP        int
	Level     int
	Inventory map[string]int // clé = mot, valeur = nb capturés
}

// Word représente un mot/monstre dans WordMon.
type Word struct {
	ID     string
	Text   string
	Rarity Rarity
	Points int
}

// NewPlayer crée un nouveau joueur avec l'inventaire initialisé.
func NewPlayer(id, name string) Player {
	return Player{
		ID:        id,
		Name:      name,
		XP:        0,
		Level:     1,
		Inventory: make(map[string]int), // Initialisation correcte de la map
	}
}
