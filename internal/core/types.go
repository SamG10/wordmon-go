package core

// Type Rarity comme alias de string
type Rarity string

// Constantes de rareté avec iota
const (
	Common    Rarity = "common"
	Rare      Rarity = "rare"
	Legendary Rarity = "legendary"
)

// Struct Player
type Player struct {
	ID        string
	Name      string
	XP        int
	Level     int
	Inventory map[string]int // clé = mot, valeur = nb capturés
}

// Struct Word
type Word struct {
	ID     string
	Text   string
	Rarity Rarity
	Points int
}

// NewPlayer crée un nouveau joueur avec l'inventaire initialisé
func NewPlayer(id, name string) Player {
	return Player{
		ID:        id,
		Name:      name,
		XP:        0,
		Level:     1,
		Inventory: make(map[string]int), // Initialisation correcte de la map
	}
}