package api

import (
	"net/http"
	"strconv"

	"github.com/SamG1008/wordmon-go/internal/config"
	"github.com/SamG1008/wordmon-go/internal/store"
	"github.com/gin-gonic/gin"
)

// UniversalServer serveur API générique qui fonctionne avec n'importe quel Store
type UniversalServer struct {
	store      store.Store
	gameConfig *config.GameConfig
	spawner    *UniversalSpawner
}

// NewUniversalServer crée un serveur API générique
func NewUniversalServer(s store.Store, gameConfig *config.GameConfig, challengesConfig *config.ChallengesConfig) *UniversalServer {
	server := &UniversalServer{
		store:      s,
		gameConfig: gameConfig,
	}

	// Démarrer le spawner
	server.spawner = NewUniversalSpawner(s, gameConfig)
	server.spawner.Start()

	return server
}

// Run démarre le serveur
func (s *UniversalServer) Run(addr string) error {
	router := gin.Default()

	// Routes
	router.GET("/status", s.handleStatus)
	router.POST("/players", s.handleCreatePlayer)
	router.GET("/players/:id", s.handleGetPlayer)
	router.GET("/spawn/current", s.handleGetCurrentSpawn)
	router.POST("/encounter/attempt", s.handleEncounterAttempt)
	router.GET("/leaderboard", s.handleLeaderboard)

	return router.Run(addr)
}

// Handlers
func (s *UniversalServer) handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "0.8.1-universal",
		"store":   "universal",
	})
}

func (s *UniversalServer) handleCreatePlayer(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player, err := s.store.Create(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, player)
}

func (s *UniversalServer) handleGetPlayer(c *gin.Context) {
	id := c.Param("id")

	player, err := s.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Récupération de l'inventaire
	inventory, err := s.store.ListByPlayer(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"id":        player.ID,
		"name":      player.Name,
		"xp":        player.XP,
		"level":     player.Level,
		"inventory": inventory,
	}

	c.JSON(http.StatusOK, response)
}

func (s *UniversalServer) handleGetCurrentSpawn(c *gin.Context) {
	spawn := s.spawner.GetCurrentSpawn()
	if spawn == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aucun WordMon disponible actuellement"})
		return
	}

	c.JSON(http.StatusOK, spawn)
}

func (s *UniversalServer) handleEncounterAttempt(c *gin.Context) {
	var req struct {
		PlayerID string `json:"playerId" binding:"required"`
		Attempt  string `json:"attempt" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	spawn := s.spawner.GetCurrentSpawn()
	if spawn == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aucun WordMon disponible"})
		return
	}

	// Validation de la tentative (logique simplifiée)
	success := req.Attempt == spawn.Text

	if success {
		// Ajouter la capture
		if err := s.store.Add(req.PlayerID, spawn.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Générer un nouveau spawn
		s.spawner.ForceNewSpawn()

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Capture réussie !",
			"xp":      spawn.Points,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Tentative échouée",
		})
	}
}

func (s *UniversalServer) handleLeaderboard(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	players, err := s.store.List(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, players)
}
