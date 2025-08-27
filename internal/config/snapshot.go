package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PlayerSnapshot représente un joueur dans le snapshot
type PlayerSnapshot struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	XP        int            `json:"xp"`
	Level     int            `json:"level"`
	Inventory map[string]int `json:"inventory"`
}

// GameSnapshot représente l'état complet du jeu
type GameSnapshot struct {
	UpdatedAt string           `json:"updatedAt"`
	Players   []PlayerSnapshot `json:"players"`
}

// SaveSnapshot sauvegarde l'état des joueurs dans un fichier JSON
func SaveSnapshot(players []PlayerSnapshot, filePath string) error {
	snapshot := GameSnapshot{
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Players:   players,
	}

	// Créer le répertoire data s'il n'existe pas
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("erreur création répertoire %s: %w", dir, err)
	}

	// Écriture atomique: fichier temporaire puis rename
	tempFile := filePath + ".tmp"

	// Encoder en JSON avec indentation
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur encodage JSON: %w", err)
	}

	// Écrire dans le fichier temporaire
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("erreur écriture fichier temporaire: %w", err)
	}

	// Renommer pour remplacer atomiquement
	if err := os.Rename(tempFile, filePath); err != nil {
		// Nettoyer le fichier temporaire en cas d'erreur
		os.Remove(tempFile)
		return fmt.Errorf("erreur rename fichier: %w", err)
	}

	fmt.Printf("[snapshot] sauvegarde: %s (%d joueurs)\n", filePath, len(players))
	return nil
}

// LoadSnapshot charge l'état des joueurs depuis un fichier JSON
func LoadSnapshot(filePath string) (*GameSnapshot, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("erreur résolution chemin snapshot: %w", err)
	}

	// Vérifier si le fichier existe
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// Retourner un snapshot vide si le fichier n'existe pas
		return &GameSnapshot{
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			Players:   []PlayerSnapshot{},
		}, nil
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture fichier snapshot: %w", err)
	}

	var snapshot GameSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON snapshot: %w", err)
	}

	fmt.Printf("[snapshot] chargement: %s (%d joueurs)\n", absPath, len(snapshot.Players))
	return &snapshot, nil
}
