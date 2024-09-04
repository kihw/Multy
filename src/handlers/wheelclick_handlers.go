package handlers

import (
	"net/http"

	"github.com/kihw/multy/src/services"

	"github.com/gin-gonic/gin"
)

type WheelClickHandler struct {
	WheelClickService *services.WheelClickService
}

// StartWheelClick listens for middle mouse clicks.
// @Summary Start middle mouse click detection
// @Description Listens for middle mouse clicks and triggers click simulation on Dofus windows.
// @Tags WheelClick
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /wheelclick/start [post]
func (h *WheelClickHandler) StartWheelClick(c *gin.Context) {
	h.WheelClickService.Start()
	c.JSON(http.StatusOK, gin.H{"message": "Wheel click detection started"})
}

// StopWheelClick stops listening for middle mouse clicks.
// @Summary Stop middle mouse click detection
// @Description Stops the detection of middle mouse clicks.
// @Tags WheelClick
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /wheelclick/stop [post]
func (h *WheelClickHandler) StopWheelClick(c *gin.Context) {
	h.WheelClickService.Stop()
	c.JSON(http.StatusOK, gin.H{"message": "Wheel click detection stopped"})
}
