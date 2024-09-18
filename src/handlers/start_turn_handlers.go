package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kihw/multy/src/services"
)

// StartTurnServiceHandler gère les requêtes liées à StartTurnService.
type StartTurnServiceHandler struct {
	StartTurnService *services.StartTurnService
}

// NewStartTurnServiceHandler crée un nouveau gestionnaire pour StartTurnService.
func NewStartTurnServiceHandler(startTurnService *services.StartTurnService) *StartTurnServiceHandler {
	return &StartTurnServiceHandler{
		StartTurnService: startTurnService,
	}
}

// StartService démarre StartTurnService.
// @Summary Démarrer le service de détection d'événements pour une fenêtre spécifique
// @Description Démarre le service pour écouter les événements d'une fenêtre spécifiée par son titre
// @Tags StartTurn
// @Accept  json
// @Produce  json
// @Param windowTitle query string true "Titre de la fenêtre à surveiller"
// @Success 200 {object} map[string]string "Service démarré avec succès"
// @Failure 400 {object} map[string]string "Erreur si le service est déjà en cours ou si windowTitle est manquant"
// @Router /start-turn/start [get]
func (h *StartTurnServiceHandler) StartService(c *gin.Context) {

	// Récupérer le titre de la fenêtre depuis les paramètres de la requête
	windowTitle := c.Query("windowTitle")
	if windowTitle == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "windowTitle is required"})
		return
	}

	// Démarrer le service avec le titre de la fenêtre spécifié
	h.StartTurnService.Start(windowTitle)
	c.JSON(http.StatusOK, gin.H{"message": "StartTurnService started for window", "windowTitle": windowTitle})
}

// StopService arrête StartTurnService.
// @Summary Arrêter le service de détection d'événements
// @Description Arrête le service d'écoute des événements sur une fenêtre
// @Tags StartTurn
// @Produce  json
// @Success 200 {object} map[string]string "Service arrêté avec succès"
// @Failure 400 {object} map[string]string "Erreur si le service n'est pas en cours"
// @Router /start-turn/stop [get]
func (h *StartTurnServiceHandler) StopService(c *gin.Context) {

	// Arrêter le service
	h.StartTurnService.Stop()
	c.JSON(http.StatusOK, gin.H{"message": "StartTurnService stopped"})
}
