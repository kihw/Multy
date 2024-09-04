package routes

import (
	"net/http"

	"github.com/kihw/multy/services"

	"github.com/kihw/multy/handlers"

	"github.com/gin-gonic/gin"
)

// SetupWindowRoutes configure les routes liées aux fenêtres
func SetupWindowRoutes(r *gin.Engine, ws *services.WindowService) {
	// Route to get the list of open windows
	r.GET("/windows", func(c *gin.Context) {
		windows, err := ws.GetWindows()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, windows)
		}
	})

	// Ensure the windowService is correctly passed
	r.POST("/focus/:keyword", func(c *gin.Context) {
		keyword := c.Param("keyword")
		err := ws.FocusWindowWithTitle(keyword) // Use the same window service instance
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Window focused successfully"})
	})
}

// SetupShortcutRoutes configure les routes liées aux raccourcis
func SetupShortcutRoutes(r *gin.Engine, hs *handlers.HandlersService) {
	r.POST("/shortcut/register/:key/:windowName", hs.RegisterHotKeyHandler)
	r.DELETE("/shortcut/unregister/:id", hs.UnregisterHotKeyHandler)
}

func SetupWheelClickRoutes(r *gin.Engine, wh *handlers.WheelClickHandler) {
	r.POST("/wheelclick/start", wh.StartWheelClick)
}

func SetupStartTurnServiceRoutes(r *gin.Engine, handler *handlers.StartTurnServiceHandler) {
    r.POST("/StartTurn/start", handler.Start)
    r.POST("/StartTurn/stop", handler.Stop)
}
