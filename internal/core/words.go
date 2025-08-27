package core

import (
	"math/rand"
)

// Variables globales pour les pools configurables
var (
	poolCommon    []Word
	poolRare      []Word
	poolLegendary []Word
	rarityWeights struct {
		Common    int
		Rare      int
		Legendary int
	}
	xpRewards struct {
		Common    int
		Rare      int
		Legendary int
	}
)

// ConfigureWords configure les pools de mots et les paramètres depuis la config
func ConfigureWords(words []WordEntry, weights RarityWeights, rewards XPRewards) {
	// Réinitialiser les pools
	poolCommon = make([]Word, 0)
	poolRare = make([]Word, 0)
	poolLegendary = make([]Word, 0)

	// Organiser les mots par rareté
	for _, entry := range words {
		word := Word{
			ID:     entry.ID,
			Text:   entry.Text,
			Rarity: parseRarity(entry.Rarity),
		}

		// Assigner les points selon la rareté et la config
		switch word.Rarity {
		case Common:
			word.Points = rewards.Common
			poolCommon = append(poolCommon, word)
		case Rare:
			word.Points = rewards.Rare
			poolRare = append(poolRare, word)
		case Legendary:
			word.Points = rewards.Legendary
			poolLegendary = append(poolLegendary, word)
		}
	}

	// Configurer les poids
	rarityWeights.Common = weights.Common
	rarityWeights.Rare = weights.Rare
	rarityWeights.Legendary = weights.Legendary

	// Configurer les récompenses
	xpRewards.Common = rewards.Common
	xpRewards.Rare = rewards.Rare
	xpRewards.Legendary = rewards.Legendary
}

// Types pour la configuration (pour éviter les imports circulaires)
type WordEntry struct {
	ID     string
	Text   string
	Rarity string
}

type RarityWeights struct {
	Common    int
	Rare      int
	Legendary int
}

type XPRewards struct {
	Common    int
	Rare      int
	Legendary int
}

// parseRarity convertit une string en enum Rarity
func parseRarity(rarity string) Rarity {
	switch rarity {
	case "Common":
		return Common
	case "Rare":
		return Rare
	case "Legendary":
		return Legendary
	default:
		return Common // Valeur par défaut
	}
}

// SpawnWord retourne un mot aléatoire en respectant les pondérations configurées
func SpawnWord() Word {
	// Utiliser les pools par défaut si pas configuré
	if len(poolCommon) == 0 {
		initDefaultPools()
	}

	// Générer un nombre aléatoire entre 0 et 99
	roll := rand.Intn(100)

	// Utiliser les poids configurés
	commonThreshold := rarityWeights.Common
	rareThreshold := commonThreshold + rarityWeights.Rare

	switch {
	case roll < commonThreshold && len(poolCommon) > 0:
		return poolCommon[rand.Intn(len(poolCommon))]
	case roll < rareThreshold && len(poolRare) > 0:
		return poolRare[rand.Intn(len(poolRare))]
	case len(poolLegendary) > 0:
		return poolLegendary[rand.Intn(len(poolLegendary))]
	default:
		// Fallback si pas de mots dans la catégorie
		if len(poolCommon) > 0 {
			return poolCommon[rand.Intn(len(poolCommon))]
		}
		// Fallback ultime
		return Word{ID: "fallback", Text: "mot", Rarity: Common, Points: 1}
	}
}

// initDefaultPools initialise les pools par défaut (rétrocompatibilité)
func initDefaultPools() {
	poolCommon = []Word{
		{ID: "c001", Text: "chat", Rarity: Common, Points: 5},
		{ID: "c002", Text: "chien", Rarity: Common, Points: 5},
		{ID: "c003", Text: "maison", Rarity: Common, Points: 5},
		{ID: "c004", Text: "soleil", Rarity: Common, Points: 5},
		{ID: "c005", Text: "eau", Rarity: Common, Points: 5},
	}

	poolRare = []Word{
		{ID: "r001", Text: "licorne", Rarity: Rare, Points: 20},
		{ID: "r002", Text: "phoenix", Rarity: Rare, Points: 20},
		{ID: "r003", Text: "cristal", Rarity: Rare, Points: 20},
		{ID: "r004", Text: "tempête", Rarity: Rare, Points: 20},
		{ID: "r005", Text: "étoile", Rarity: Rare, Points: 20},
	}

	poolLegendary = []Word{
		{ID: "l001", Text: "dragon", Rarity: Legendary, Points: 100},
		{ID: "l002", Text: "excalibur", Rarity: Legendary, Points: 100},
		{ID: "l003", Text: "atlantide", Rarity: Legendary, Points: 100},
		{ID: "l004", Text: "immortel", Rarity: Legendary, Points: 100},
		{ID: "l005", Text: "cosmos", Rarity: Legendary, Points: 100},
	}

	// Poids par défaut
	rarityWeights.Common = 80
	rarityWeights.Rare = 18
	rarityWeights.Legendary = 2

	// Récompenses par défaut
	xpRewards.Common = 5
	xpRewards.Rare = 20
	xpRewards.Legendary = 100
}
