package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SamG1008/wordmon-go/internal/config"
	"github.com/SamG1008/wordmon-go/internal/core"
	"github.com/SamG1008/wordmon-go/internal/store"
	"github.com/gin-gonic/gin"
)

// SQLServer représente le serveur API avec base SQL
type SQLServer struct {
	store        *store.SQLStore
	gameConfig   *config.GameConfig
	router       *gin.Engine
	startTime    time.Time
	currentSpawn *core.Word
}

// NewSQLServer crée un nouveau serveur API SQL
func NewSQLServer(sqlStore *store.SQLStore, gameConfig *config.GameConfig) *SQLServer {
	// Configuration de Gin
	gin.SetMode(gin.ReleaseMode) // Moins verbeux pour la production
	router := gin.New()

	// Middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	server := &SQLServer{
		store:      sqlStore,
		gameConfig: gameConfig,
		router:     router,
		startTime:  time.Now(),
	}

	// Configurer les routes
	server.setupRoutes()

	return server
}

// setupRoutes configure toutes les routes de l'API
func (s *SQLServer) setupRoutes() {
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
func (s *SQLServer) Run(port string) error {
	return s.router.Run(":" + port)
}

// isValidAnagram vérifie si deux mots sont des anagrammes
func isValidAnagram(attempt, target string) bool {
	// Normaliser
	attempt = strings.TrimSpace(strings.ToLower(attempt))
	target = strings.ToLower(target)

	// Vérifier que ce n'est pas le même mot
	if attempt == target {
		return false
	}

	if len(attempt) != len(target) {
		return false
	}

	// Convertir en slices de runes et trier
	runes1 := []rune(attempt)
	runes2 := []rune(target)

	sort.Slice(runes1, func(i, j int) bool { return runes1[i] < runes1[j] })
	sort.Slice(runes2, func(i, j int) bool { return runes2[i] < runes2[j] })

	return string(runes1) == string(runes2)
}

// === HANDLERS ===

// getStatus retourne le statut du serveur
func (s *SQLServer) getStatus(c *gin.Context) {
	// Compter les joueurs actifs
	players, err := s.store.List(0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur récupération stats"})
		return
	}

	var currentSpawnJSON *WordJSON
	if s.currentSpawn != nil {
		currentSpawnJSON = &WordJSON{
			ID:     s.currentSpawn.ID,
			Text:   s.currentSpawn.Text,
			Rarity: string(s.currentSpawn.Rarity),
			Points: s.currentSpawn.Points,
		}
	}

	response := StatusResponse{
		Game:          s.gameConfig.Game.Name,
		Version:       s.gameConfig.Game.Version,
		UptimeSeconds: int(time.Since(s.startTime).Seconds()),
		ActivePlayers: len(players),
		CurrentSpawn:  currentSpawnJSON,
	}

	c.JSON(http.StatusOK, response)
}

// createPlayer crée un nouveau joueur
func (s *SQLServer) createPlayer(c *gin.Context) {
	var req CreatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nom requis"})
		return
	}

	player, err := s.store.Create(req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("erreur création: %v", err)})
		return
	}

	response := PlayerJSON{
		ID:    player.ID,
		Name:  player.Name,
		XP:    player.XP,
		Level: player.Level,
	}

	c.JSON(http.StatusCreated, response)
}

// getPlayer récupère un joueur par ID
func (s *SQLServer) getPlayer(c *gin.Context) {
	playerID := c.Param("id")

	player, err := s.store.Get(playerID)
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
func (s *SQLServer) getCurrentSpawn(c *gin.Context) {
	if s.currentSpawn == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "aucun WordMon actif"})
		return
	}

	response := WordJSON{
		ID:     s.currentSpawn.ID,
		Text:   s.currentSpawn.Text,
		Rarity: string(s.currentSpawn.Rarity),
		Points: s.currentSpawn.Points,
	}

	c.JSON(http.StatusOK, response)
}

// attemptCapture gère les tentatives de capture
func (s *SQLServer) attemptCapture(c *gin.Context) {
	var req AttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "playerId et attempt requis"})
		return
	}

	// Vérifier qu'il y a un WordMon actif
	if s.currentSpawn == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "aucun WordMon actif"})
		return
	}

	// Récupérer le joueur
	player, err := s.store.Get(req.PlayerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "joueur non trouvé"})
		return
	}

	// Vérifier la tentative (anagramme)
	if !isValidAnagram(req.Attempt, s.currentSpawn.Text) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tentative incorrecte"})
		return
	}

	// Enregistrer la capture (transaction atomique dans SQLStore)
	err = s.store.Add(player.ID, s.currentSpawn.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur enregistrement capture"})
		return
	}

	// Calcul des récompenses
	xpGained := s.currentSpawn.Points
	newXP := player.XP + xpGained
	newLevel := newXP/100 + 1

	// Log de succès
	fmt.Printf("[api] Capture success: %s +%dXP (level=%d)\n", player.Name, xpGained, newLevel)

	// Réponse de succès
	response := AttemptResponse{
		Status:   "captured",
		Word:     s.currentSpawn.Text,
		Rarity:   string(s.currentSpawn.Rarity),
		XPGained: xpGained,
		NewLevel: newLevel,
	}

	// Effacer le spawn actuel
	s.currentSpawn = nil

	c.JSON(http.StatusOK, response)
}

// getLeaderboard retourne le classement des joueurs
func (s *SQLServer) getLeaderboard(c *gin.Context) {
	limitParam := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		limit = 10
	}

	players, err := s.store.List(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur récupération leaderboard"})
		return
	}

	var response []PlayerJSON
	for _, player := range players {
		response = append(response, PlayerJSON{
			ID:    player.ID,
			Name:  player.Name,
			XP:    player.XP,
			Level: player.Level,
		})
	}

	c.JSON(http.StatusOK, response)
}

// === SPAWN MANAGEMENT ===

// SetCurrentSpawn définit le WordMon actuel
func (s *SQLServer) SetCurrentSpawn(word *core.Word) {
	s.currentSpawn = word
}

// GetCurrentSpawn retourne le WordMon actuel
func (s *SQLServer) GetCurrentSpawn() *core.Word {
	return s.currentSpawn
}

// ClearCurrentSpawn supprime le WordMon actuel
func (s *SQLServer) ClearCurrentSpawn() {
	s.currentSpawn = nil
}
