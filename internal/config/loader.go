package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// GameConfig représente la configuration principale du jeu
type GameConfig struct {
	Game          GameInfo      `yaml:"game" toml:"game"`
	RarityWeights RarityWeights `yaml:"rarityWeights" toml:"rarityWeights"`
	XPRewards     XPRewards     `yaml:"xpRewards" toml:"xpRewards"`
	Spawner       SpawnerConfig `yaml:"spawner" toml:"spawner"`
	Level         LevelConfig   `yaml:"level" toml:"level"`
}

type GameInfo struct {
	Name    string `yaml:"name" toml:"name"`
	Version string `yaml:"version" toml:"version"`
}

type RarityWeights struct {
	Common    int `yaml:"Common" toml:"Common"`
	Rare      int `yaml:"Rare" toml:"Rare"`
	Legendary int `yaml:"Legendary" toml:"Legendary"`
}

type XPRewards struct {
	Common    int `yaml:"Common" toml:"Common"`
	Rare      int `yaml:"Rare" toml:"Rare"`
	Legendary int `yaml:"Legendary" toml:"Legendary"`
}

type SpawnerConfig struct {
	IntervalSeconds      int `yaml:"intervalSeconds" toml:"intervalSeconds"`
	AutoFleeAfterSeconds int `yaml:"autoFleeAfterSeconds" toml:"autoFleeAfterSeconds"`
}

type LevelConfig struct {
	Base       int `yaml:"base" toml:"base"`
	XPPerLevel int `yaml:"xpPerLevel" toml:"xpPerLevel"`
}

// ChallengesConfig représente la configuration des défis
type ChallengesConfig struct {
	Anagram AnagramConfig `yaml:"anagram"`
	ATrou   ATrouConfig   `yaml:"aTrou"`
}

type AnagramConfig struct {
	MinLenByRarity       map[string]int `yaml:"minLenByRarity"`
	MustDifferFromSource bool           `yaml:"mustDifferFromSource"`
}

type ATrouConfig struct {
	RevealedLetters map[string]int `yaml:"revealedLetters"`
	MaxAttempts     int            `yaml:"maxAttempts"`
}

// WordEntry représente une entrée du dictionnaire de mots
type WordEntry struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Rarity string `json:"rarity"`
}

// WordsConfig contient la liste des mots par rareté
type WordsConfig struct {
	Words    []WordEntry            `json:"words,omitempty"`
	ByRarity map[string][]WordEntry `json:"-"`
}

// LoadGameConfig charge la configuration principale du jeu
func LoadGameConfig(path string) (*GameConfig, error) {
	// Vérifier l'override par variable d'environnement
	if envPath := os.Getenv("WORDMON_CONFIG_PATH"); envPath != "" {
		path = envPath
	}

	// Résoudre le chemin absolu
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("erreur résolution chemin config: %w", err)
	}

	fmt.Printf("[config] chargement game config: %s\n", absPath)

	// Lire le fichier
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture fichier config: %w", err)
	}

	var config GameConfig

	// Détecter le format par l'extension
	ext := strings.ToLower(filepath.Ext(absPath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("erreur parsing YAML: %w", err)
		}
	case ".toml":
		if err := toml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("erreur parsing TOML: %w", err)
		}
	default:
		return nil, fmt.Errorf("format de fichier non supporté: %s (attendu: .yaml, .yml, .toml)", ext)
	}

	// Appliquer les valeurs par défaut
	applyGameDefaults(&config)

	// Appliquer les overrides ENV
	applyGameEnvOverrides(&config)

	// Valider la configuration
	if err := validateGameConfig(&config); err != nil {
		return nil, fmt.Errorf("validation config échouée: %w", err)
	}

	// Log de confirmation
	fmt.Printf("[config] game: %s v%s\n", config.Game.Name, config.Game.Version)
	fmt.Printf("[config] rarity weights: C=%d R=%d L=%d (OK sum=%d)\n",
		config.RarityWeights.Common, config.RarityWeights.Rare, config.RarityWeights.Legendary,
		config.RarityWeights.Common+config.RarityWeights.Rare+config.RarityWeights.Legendary)
	fmt.Printf("[config] xp rewards: C=%d R=%d L=%d\n",
		config.XPRewards.Common, config.XPRewards.Rare, config.XPRewards.Legendary)

	return &config, nil
}

// LoadChallengesConfig charge la configuration des défis
func LoadChallengesConfig(path string) (*ChallengesConfig, error) {
	// Vérifier l'override par variable d'environnement
	if envPath := os.Getenv("WORDMON_CHALLENGES_PATH"); envPath != "" {
		path = envPath
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("erreur résolution chemin challenges: %w", err)
	}

	fmt.Printf("[config] chargement challenges config: %s\n", absPath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture fichier challenges: %w", err)
	}

	var config ChallengesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("erreur parsing YAML challenges: %w", err)
	}

	// Appliquer les valeurs par défaut
	applyChallengesDefaults(&config)

	// Valider
	if err := validateChallengesConfig(&config); err != nil {
		return nil, fmt.Errorf("validation challenges échouée: %w", err)
	}

	fmt.Printf("[config] challenges: anagram + a-trou chargés\n")

	return &config, nil
}

// LoadWordsConfig charge le dictionnaire de mots
func LoadWordsConfig(path string) (*WordsConfig, error) {
	// Vérifier l'override par variable d'environnement
	if envPath := os.Getenv("WORDMON_WORDS_PATH"); envPath != "" {
		path = envPath
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("erreur résolution chemin words: %w", err)
	}

	fmt.Printf("[config] chargement words config: %s\n", absPath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture fichier words: %w", err)
	}

	var words []WordEntry
	if err := json.Unmarshal(data, &words); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON words: %w", err)
	}

	config := &WordsConfig{
		Words:    words,
		ByRarity: make(map[string][]WordEntry),
	}

	// Organiser par rareté
	for _, word := range words {
		config.ByRarity[word.Rarity] = append(config.ByRarity[word.Rarity], word)
	}

	// Valider
	if err := validateWordsConfig(config); err != nil {
		return nil, fmt.Errorf("validation words échouée: %w", err)
	}

	fmt.Printf("[config] words: Common=%d Rare=%d Legendary=%d (total=%d)\n",
		len(config.ByRarity["Common"]), len(config.ByRarity["Rare"]),
		len(config.ByRarity["Legendary"]), len(words))

	return config, nil
}

// applyGameDefaults applique les valeurs par défaut
func applyGameDefaults(config *GameConfig) {
	if config.Game.Name == "" {
		config.Game.Name = "WordMon Go"
	}
	if config.Game.Version == "" {
		config.Game.Version = "0.6.0"
	}
	if config.Level.XPPerLevel == 0 {
		config.Level.XPPerLevel = 100
	}
	if config.Level.Base == 0 {
		config.Level.Base = 1
	}
	if config.Spawner.IntervalSeconds == 0 {
		config.Spawner.IntervalSeconds = 10
	}
	if config.Spawner.AutoFleeAfterSeconds == 0 {
		config.Spawner.AutoFleeAfterSeconds = 5
	}
}

// applyGameEnvOverrides applique les overrides ENV
func applyGameEnvOverrides(config *GameConfig) {
	if interval := os.Getenv("WORDMON_SPAWN_INTERVAL"); interval != "" {
		if val, err := strconv.Atoi(interval); err == nil {
			config.Spawner.IntervalSeconds = val
			fmt.Printf("[config] override ENV: spawner.intervalSeconds=%d\n", val)
		}
	}
}

// applyChallengesDefaults applique les valeurs par défaut pour challenges
func applyChallengesDefaults(config *ChallengesConfig) {
	if config.Anagram.MinLenByRarity == nil {
		config.Anagram.MinLenByRarity = map[string]int{
			"Common": 3, "Rare": 5, "Legendary": 7,
		}
	}
	if config.ATrou.RevealedLetters == nil {
		config.ATrou.RevealedLetters = map[string]int{
			"Common": 2, "Rare": 1, "Legendary": 0,
		}
	}
	if config.ATrou.MaxAttempts == 0 {
		config.ATrou.MaxAttempts = 4
	}
}

// validateGameConfig valide la configuration du jeu
func validateGameConfig(config *GameConfig) error {
	// Vérifier que la somme des poids fait 100
	sum := config.RarityWeights.Common + config.RarityWeights.Rare + config.RarityWeights.Legendary
	if sum != 100 {
		return fmt.Errorf("somme des poids de rareté doit être 100, obtenu: %d", sum)
	}

	// Vérifier que les récompenses XP sont positives
	if config.XPRewards.Common <= 0 || config.XPRewards.Rare <= 0 || config.XPRewards.Legendary <= 0 {
		return fmt.Errorf("les récompenses XP doivent être strictement positives")
	}

	// Vérifier que les poids sont positifs
	if config.RarityWeights.Common < 0 || config.RarityWeights.Rare < 0 || config.RarityWeights.Legendary < 0 {
		return fmt.Errorf("les poids de rareté doivent être positifs")
	}

	return nil
}

// validateChallengesConfig valide la configuration des défis
func validateChallengesConfig(config *ChallengesConfig) error {
	rarities := []string{"Common", "Rare", "Legendary"}

	// Vérifier anagram minLen
	for _, rarity := range rarities {
		if len, ok := config.Anagram.MinLenByRarity[rarity]; !ok || len < 0 {
			return fmt.Errorf("anagram.minLenByRarity.%s manquant ou négatif", rarity)
		}
	}

	// Vérifier aTrou revealedLetters
	for _, rarity := range rarities {
		if letters, ok := config.ATrou.RevealedLetters[rarity]; !ok || letters < 0 {
			return fmt.Errorf("aTrou.revealedLetters.%s manquant ou négatif", rarity)
		}
	}

	if config.ATrou.MaxAttempts <= 0 {
		return fmt.Errorf("aTrou.maxAttempts doit être positif")
	}

	return nil
}

// validateWordsConfig valide la configuration des mots
func validateWordsConfig(config *WordsConfig) error {
	validRarities := map[string]bool{"Common": true, "Rare": true, "Legendary": true}
	seenIDs := make(map[string]bool)

	for _, word := range config.Words {
		// Vérifier ID unique
		if seenIDs[word.ID] {
			return fmt.Errorf("ID dupliqué: %s", word.ID)
		}
		seenIDs[word.ID] = true

		// Vérifier text non vide
		if strings.TrimSpace(word.Text) == "" {
			return fmt.Errorf("text vide pour ID: %s", word.ID)
		}

		// Vérifier pas d'espaces dans text
		if strings.Contains(word.Text, " ") {
			return fmt.Errorf("text contient des espaces pour ID: %s", word.ID)
		}

		// Vérifier rareté valide
		if !validRarities[word.Rarity] {
			return fmt.Errorf("rareté inconnue: %s (ID: %s)", word.Rarity, word.ID)
		}
	}

	// Vérifier minimum 5 mots par rareté
	for rarity := range validRarities {
		if len(config.ByRarity[rarity]) < 5 {
			return fmt.Errorf("minimum 5 mots requis pour rareté %s, obtenu: %d",
				rarity, len(config.ByRarity[rarity]))
		}
	}

	return nil
}
