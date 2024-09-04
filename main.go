package main

import (
	"log"

	_ "github.com/kihw/multy/docs"
	"github.com/kihw/multy/handlers"
	"github.com/kihw/multy/routes"

	"github.com/kihw/multy/services"

	"github.com/gin-gonic/gin"
	swagFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Multy API
// @version 1.0
// @description This is a sample server for managing shortcuts.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @host localhost:8080
// @BasePath /
func main() {
    r := gin.Default()


    // Initialize services.
    wheelClickService := &services.WheelClickService{}
    shortcutService := services.NewShortcutService()
    windowService := &services.WindowService{}
    startTurnService := services.NewStartTurnService(windowService)

    // Initialize handlers with their respective services.
    wheelClickHandler := &handlers.WheelClickHandler{
        WheelClickService: wheelClickService,
    }
    handlersService := &handlers.HandlersService{
        ShortcutService: shortcutService,
    }
    startTurnServiceHandler := handlers.NewStartTurnServiceHandler(startTurnService)

    // Log open windows for debugging.
    windows, err := windowService.GetWindows()
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Open windows:", windows)

    // Configure routes with respective handlers.
    routes.SetupShortcutRoutes(r, handlersService)
    routes.SetupWindowRoutes(r, windowService)
    routes.SetupWheelClickRoutes(r, wheelClickHandler)
    routes.SetupStartTurnServiceRoutes(r, startTurnServiceHandler)

    // Swagger route for API documentation.
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swagFiles.Handler))

    // Start the server on port 8080.
    r.Run(":8080")
}
