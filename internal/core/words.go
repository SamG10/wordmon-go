package core

import (
	"math/rand"
	"time"
)

// Pools de mots par rareté
var poolCommon = []Word{
	{ID: "c001", Text: "chat", Rarity: Common, Points: 5},
	{ID: "c002", Text: "chien", Rarity: Common, Points: 5},
	{ID: "c003", Text: "maison", Rarity: Common, Points: 5},
	{ID: "c004", Text: "soleil", Rarity: Common, Points: 5},
	{ID: "c005", Text: "eau", Rarity: Common, Points: 5},
}

var poolRare = []Word{
	{ID: "r001", Text: "licorne", Rarity: Rare, Points: 20},
	{ID: "r002", Text: "phoenix", Rarity: Rare, Points: 20},
	{ID: "r003", Text: "cristal", Rarity: Rare, Points: 20},
	{ID: "r004", Text: "tempête", Rarity: Rare, Points: 20},
	{ID: "r005", Text: "étoile", Rarity: Rare, Points: 20},
}

var poolLegendary = []Word{
	{ID: "l001", Text: "dragon", Rarity: Legendary, Points: 100},
	{ID: "l002", Text: "excalibur", Rarity: Legendary, Points: 100},
	{ID: "l003", Text: "atlantide", Rarity: Legendary, Points: 100},
	{ID: "l004", Text: "immortel", Rarity: Legendary, Points: 100},
	{ID: "l005", Text: "cosmos", Rarity: Legendary, Points: 100},
}

// SpawnWord retourne un mot aléatoire en respectant des pondérations de rareté.
// Pondération : Common 80%, Rare 18%, Legendary 2%
func SpawnWord() Word {
	// Initialiser le générateur aléatoire avec le temps actuel
	rand.Seed(time.Now().UnixNano())
	
	// Générer un nombre aléatoire entre 0 et 99
	roll := rand.Intn(100)
	
	switch {
	case roll < 80: // 0-79 : Common (80%)
		return poolCommon[rand.Intn(len(poolCommon))]
	case roll < 98: // 80-97 : Rare (18%)
		return poolRare[rand.Intn(len(poolRare))]
	default: // 98-99 : Legendary (2%)
		return poolLegendary[rand.Intn(len(poolLegendary))]
	}
}
