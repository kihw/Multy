package handlers

import (
	"log"
	"net/http"

	"github.com/kihw/multy/src/services"

	"github.com/gin-gonic/gin"
)

type HandlersService struct {
	ShortcutService *services.ShortcutService
}

// @Summary Register a hotkey
// @Description Registers a hotkey to focus on a window
// @Tags Shortcut
// @Accept json
// @Produce json
// @Param key path string true "Key to register"
// @Param windowName path string true "Name of the window to focus"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /shortcut/register/{key}/{windowName} [post]
func (hs *HandlersService) RegisterHotKeyHandler(c *gin.Context) {
	key := c.Param("key")
	windowName := c.Param("windowName") // This should come from the request

	shortcut := services.Shortcut{
		ID:         hs.ShortcutService.GenerateHotkeyID(),
		Key:        key,
		WindowName: windowName, // Set the window name here
	}

	err := hs.ShortcutService.RegisterShortcut(shortcut)
	if err != nil {
		log.Printf("Error registering shortcut: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Shortcut registered successfully"})
}

// UnregisterHotKeyHandler handles the unregistration of an existing hotkey.
// @Summary Unregister an existing hotkey
// @Description Unregisters a previously registered keyboard shortcut
// @Tags Shortcut
// @Success 200 {string} string "Raccourci désenregistré avec succès"
// @Failure 500 {string} string "Failed to unregister shortcut"
// @Router /shortcut/unregister [delete]
func (hs *HandlersService) UnregisterHotKeyHandler(c *gin.Context) {
	hs.ShortcutService.UnregisterShortcut()
	c.JSON(http.StatusOK, gin.H{"message": "Raccourci désenregistré avec succès"})
}
