package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/SamG1008/wordmon-go/internal/config"
	"github.com/SamG1008/wordmon-go/internal/core"
	"github.com/SamG1008/wordmon-go/internal/store"
	"github.com/gin-gonic/gin"
)

// Server représente le serveur API
type Server struct {
	store      *store.MemoryStore
	gameConfig *config.GameConfig
	router     *gin.Engine
}

// NewServer crée un nouveau serveur API
func NewServer(store *store.MemoryStore, gameConfig *config.GameConfig) *Server {
	// Configuration de Gin
	gin.SetMode(gin.ReleaseMode) // Moins verbeux pour la production
	router := gin.New()

	// Middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	server := &Server{
		store:      store,
		gameConfig: gameConfig,
		router:     router,
	}

	// Configurer les routes
	server.setupRoutes()

	return server
}

// setupRoutes configure toutes les routes de l'API
func (s *Server) setupRoutes() {
	// Routes principales
	s.router.GET("/status", s.getStatus)

	// Routes des joueurs
	s.router.POST("/players", s.createPlayer)
	s.router.GET("/players/:id", s.getPlayer)

	// Routes du spawn
	s.router.GET("/spawn/current", s.getCurrentSpawn)

	// Routes des tentatives
	s.router.POST("/encounter/attempt", s.attemptCapture)

	// Routes du leaderboard
	s.router.GET("/leaderboard", s.getLeaderboard)
}

// Run démarre le serveur sur le port spécifié
func (s *Server) Run(port string) error {
	return s.router.Run(":" + port)
}

// Structures pour les réponses JSON

type StatusResponse struct {
	Game          string    `json:"game"`
	Version       string    `json:"version"`
	UptimeSeconds int       `json:"uptimeSeconds"`
	ActivePlayers int       `json:"activePlayers"`
	CurrentSpawn  *WordJSON `json:"currentSpawn"`
}

type PlayerJSON struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	XP        int            `json:"xp"`
	Level     int            `json:"level"`
	Inventory map[string]int `json:"inventory,omitempty"`
}

type WordJSON struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Rarity string `json:"rarity"`
	Points int    `json:"points"`
}

type CreatePlayerRequest struct {
	Name string `json:"name" binding:"required"`
}

type AttemptRequest struct {
	PlayerID string `json:"playerId" binding:"required"`
	Attempt  string `json:"attempt" binding:"required"`
}

type AttemptResponse struct {
	Status   string `json:"status"`
	Word     string `json:"word,omitempty"`
	Rarity   string `json:"rarity,omitempty"`
	XPGained int    `json:"xpGained,omitempty"`
	NewLevel int    `json:"newLevel,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// Handlers

// getStatus retourne le status du serveur
func (s *Server) getStatus(c *gin.Context) {
	currentSpawn := s.store.GetCurrentSpawn()
	var currentSpawnJSON *WordJSON = nil

	if currentSpawn != nil {
		currentSpawnJSON = &WordJSON{
			ID:     currentSpawn.ID,
			Text:   currentSpawn.Text,
			Rarity: string(currentSpawn.Rarity),
			Points: currentSpawn.Points,
		}
	}

	response := StatusResponse{
		Game:          s.gameConfig.Game.Name,
		Version:       s.gameConfig.Game.Version,
		UptimeSeconds: s.store.GetUptimeSeconds(),
		ActivePlayers: s.store.GetActivePlayersCount(),
		CurrentSpawn:  currentSpawnJSON,
	}

	c.JSON(http.StatusOK, response)
}

// createPlayer crée un nouveau joueur
func (s *Server) createPlayer(c *gin.Context) {
	var req CreatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nom requis"})
		return
	}

	// Valider le nom
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nom ne peut pas être vide"})
		return
	}

	player, err := s.store.CreatePlayer(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := PlayerJSON{
		ID:    player.ID,
		Name:  player.Name,
		XP:    player.XP,
		Level: player.Level,
	}

	c.JSON(http.StatusOK, response)
}

// getPlayer récupère un joueur par ID
func (s *Server) getPlayer(c *gin.Context) {
	id := c.Param("id")

	player, err := s.store.GetPlayer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "joueur non trouvé"})
		return
	}

	response := PlayerJSON{
		ID:        player.ID,
		Name:      player.Name,
		XP:        player.XP,
		Level:     player.Level,
		Inventory: player.Inventory,
	}

	c.JSON(http.StatusOK, response)
}

// getCurrentSpawn retourne le WordMon actuel
func (s *Server) getCurrentSpawn(c *gin.Context) {
	currentSpawn := s.store.GetCurrentSpawn()
	if currentSpawn == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "aucun WordMon actif"})
		return
	}

	response := WordJSON{
		ID:     currentSpawn.ID,
		Text:   currentSpawn.Text,
		Rarity: string(currentSpawn.Rarity),
		Points: currentSpawn.Points,
	}

	c.JSON(http.StatusOK, response)
}

// attemptCapture gère les tentatives de capture
func (s *Server) attemptCapture(c *gin.Context) {
	var req AttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "données requises manquantes"})
		return
	}

	// Vérifier que le joueur existe
	player, err := s.store.GetPlayer(req.PlayerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "joueur non trouvé"})
		return
	}

	// Vérifier qu'il y a un spawn actif
	currentSpawn := s.store.GetCurrentSpawn()
	if currentSpawn == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "aucun WordMon actif"})
		return
	}

	// Créer un challenge pour tester la réponse
	challenge := core.NewAnagramChallenge(*currentSpawn)
	isCorrect, err := challenge.Check(req.Attempt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur lors de la vérification"})
		return
	}

	response := AttemptResponse{
		Word:   currentSpawn.Text,
		Rarity: string(currentSpawn.Rarity),
	}

	if isCorrect {
		// Victoire - attribuer les XP
		oldLevel := player.Level
		player.AwardXP(currentSpawn.Points)
		player.Inventory[currentSpawn.ID]++

		// Sauvegarder le joueur
		s.store.UpdatePlayer(player)

		// Supprimer le spawn (capturé)
		s.store.ClearCurrentSpawn()

		response.Status = "captured"
		response.XPGained = currentSpawn.Points
		response.NewLevel = player.Level

		// Log de capture
		if player.Level > oldLevel {
			gin.DefaultWriter.Write([]byte(
				fmt.Sprintf("[player] %s a capturé \"%s\" (XP+%d, Level=%d)\n",
					player.Name, currentSpawn.Text, currentSpawn.Points, player.Level)))
		} else {
			gin.DefaultWriter.Write([]byte(
				fmt.Sprintf("[player] %s a capturé \"%s\" (XP+%d)\n",
					player.Name, currentSpawn.Text, currentSpawn.Points)))
		}
	} else {
		// Défaite
		response.Status = "fled"
		response.Reason = "wrong attempt"

		// Le WordMon s'enfuit après une mauvaise tentative
		s.store.ClearCurrentSpawn()
	}

	c.JSON(http.StatusOK, response)
}

// getLeaderboard retourne le classement des joueurs
func (s *Server) getLeaderboard(c *gin.Context) {
	// Récupérer le paramètre limit (défaut = 10, max = 50)
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	players := s.store.GetLeaderboard(limit)

	response := make([]PlayerJSON, len(players))
	for i, player := range players {
		response[i] = PlayerJSON{
			ID:    player.ID,
			Name:  player.Name,
			XP:    player.XP,
			Level: player.Level,
		}
	}

	c.JSON(http.StatusOK, response)
}

// Utilitaires de conversion

// CorePlayerToJSON convertit un core.Player en PlayerJSON
func CorePlayerToJSON(player *core.Player) PlayerJSON {
	return PlayerJSON{
		ID:        player.ID,
		Name:      player.Name,
		XP:        player.XP,
		Level:     player.Level,
		Inventory: player.Inventory,
	}
}

// CoreWordToJSON convertit un core.Word en WordJSON
func CoreWordToJSON(word *core.Word) WordJSON {
	return WordJSON{
		ID:     word.ID,
		Text:   word.Text,
		Rarity: string(word.Rarity),
		Points: word.Points,
	}
}
