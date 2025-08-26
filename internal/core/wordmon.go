package core

import "fmt"

// WordMon représente un mot sauvage combattable (comme un Pokémon)
type WordMon struct {
	Word      Word
	Challenge Challenge
}

// NewWordMon crée un nouveau WordMon avec son challenge associé
func NewWordMon(word Word) WordMon {
	// Vérifications critiques - panic si les données sont corrompues
	if word.Text == "" {
		panic("WordMon invalide: mot vide")
	}
	if word.Rarity != Common && word.Rarity != Rare && word.Rarity != Legendary {
		panic(fmt.Sprintf("WordMon invalide: rareté inconnue '%s'", word.Rarity))
	}
	if word.Points <= 0 {
		panic(fmt.Sprintf("WordMon invalide: points invalides %d", word.Points))
	}
	
	return WordMon{
		Word:      word,
		Challenge: NewAnagramChallenge(word),
	}
}

// Presentation retourne une description lisible du WordMon
func (wm WordMon) Presentation() string {
	var rarityDesc string
	switch wm.Word.Rarity {
	case Common:
		rarityDesc = "Un WordMon ordinaire"
	case Rare:
		rarityDesc = "Un WordMon rare brille faiblement"
	case Legendary:
		rarityDesc = "Un WordMon légendaire rayonne de puissance"
	}
	
	return fmt.Sprintf("%s apparaît : '%s' [%s] (%d points)",
		rarityDesc, wm.Word.Text, wm.Word.Rarity, wm.Word.Points)
}

// GetChallenge retourne le défi associé à ce WordMon
func (wm WordMon) GetChallenge() Challenge {
	return wm.Challenge
}
