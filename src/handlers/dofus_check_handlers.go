package handlers

import (
	"log"
	"net/http"

	"github.com/kihw/multy/src/services" // Update this with your actual project import path

	"github.com/gin-gonic/gin"
)

// DofusCheckHandler contains the DofusCheckService instance.
type DofusCheckHandler struct {
	dofusCheckService *services.DofusCheckService
}

// NewDofusCheckHandler creates a new instance of DofusCheckHandler.
func NewDofusCheckHandler(dcs *services.DofusCheckService) *DofusCheckHandler {
	return &DofusCheckHandler{dofusCheckService: dcs}
}

// StartDofusCheck starts the monitoring of the Dofus window.
// @Summary Start monitoring Dofus window
// @Description Starts the DofusCheckService which monitors the Dofus window state
// @Tags DofusCheck
// @Success 200 {string} string "DofusCheck service started successfully"
// @Failure 500 {object} string "Error occurred while starting the service"
// @Router /dofus-check/start [post]
func (h *DofusCheckHandler) StartDofusCheck(c *gin.Context) {
	go h.dofusCheckService.StartMonitoring()
	log.Println("DofusCheck service started.")
	c.JSON(http.StatusOK, "DofusCheck service started successfully")
}

// StopDofusCheck stops the monitoring of the Dofus window.
// @Summary Stop monitoring Dofus window
// @Description Stops the DofusCheckService which monitors the Dofus window state
// @Tags DofusCheck
// @Success 200 {string} string "DofusCheck service stopped successfully"
// @Failure 500 {object} string "Error occurred while stopping the service"
// @Router /dofus-check/stop [post]
func (h *DofusCheckHandler) StopDofusCheck(c *gin.Context) {
	h.dofusCheckService.StopMonitoring()
	log.Println("DofusCheck service stopped.")
	c.JSON(http.StatusOK, "DofusCheck service stopped successfully")
}
