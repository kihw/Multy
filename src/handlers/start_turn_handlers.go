package handlers

import (
	"net/http"

	"github.com/kihw/multy/services"

	"github.com/gin-gonic/gin"
)

// StartTurnServiceHandler structure
type StartTurnServiceHandler struct {
    service *services.StartTurnService
}

// NewStartTurnServiceHandler crée un nouveau handler pour le StartTurnService
func NewStartTurnServiceHandler(service *services.StartTurnService) *StartTurnServiceHandler {
    return &StartTurnServiceHandler{
        service: service,
    }
}

// Start détecte les messages StartTurn
// @Summary Démarrer la détection du StartTurn
// @Description Démarre la détection des messages StartTurn
// @Tags Start Turn
// @Success 200 {string} string "StartTurn detection started"
// @Failure 500 {string} string "Failed to start StartTurn detection"
// @Router /StartTurn/start [post]
func (h *StartTurnServiceHandler) Start(c *gin.Context) {
    h.service.Start()
    c.JSON(200, gin.H{"status": "StartTurn detection started"})
}

// Stop détecte les messages StartTurn
// @Summary Arrêter la détection du StartTurn
// @Description Arrête la détection des messages StartTurn
// @Tags Start Turn
// @Success 200 {string} string "StartTurn detection stopped"
// @Failure 500 {string} string "Failed to stop StartTurn detection"
// @Router /StartTurn/stop [post]
func (h *StartTurnServiceHandler) Stop(c *gin.Context) {
    h.service.Stop()
    c.JSON(http.StatusOK, gin.H{"message": "StartTurn detection stopped"})
}
