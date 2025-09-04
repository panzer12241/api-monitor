package routes

import (
	"api-monitor/app/controllers"
	"api-monitor/app/middleware"
	"api-monitor/app/services"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupRoutes(app *fiber.App, db *sql.DB, monitor *services.MonitorService) {
	// Enable CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Initialize controllers
	authController := controllers.NewAuthController(db)
	endpointController := controllers.NewEndpointController(db, monitor)
	proxyController := controllers.NewProxyController(db)

	// Public routes (no auth required)
	auth := app.Group("/api/v1/auth")
	auth.Post("/login", authController.Login)
	auth.Post("/register", authController.Register)

	// Protected auth routes (require JWT)
	authProtected := app.Group("/api/v1/auth")
	authProtected.Use(middleware.JWTMiddleware())
	authProtected.Post("/logout", authController.Logout)

	// Protected API endpoints (require JWT)
	api := app.Group("/api/v1", middleware.JWTMiddleware())
	{
		// Endpoint management
		api.Get("/endpoints", endpointController.GetEndpoints)
		api.Post("/endpoints", endpointController.CreateEndpoint)
		api.Put("/endpoints/:id", endpointController.UpdateEndpoint)
		api.Delete("/endpoints/:id", endpointController.DeleteEndpoint)
		api.Post("/endpoints/:id/toggle", endpointController.ToggleEndpoint)
		api.Get("/endpoints/:id/logs", endpointController.GetEndpointLogs)
		// api.Post("/endpoints/:id/check", endpointController.ManualCheck)
		// api.Post("/cleanup-logs", endpointController.ManualCleanup)

		// Proxy management
		api.Get("/proxies", proxyController.GetProxies)
		api.Post("/proxies", proxyController.CreateProxy)
		api.Put("/proxies/:id", proxyController.UpdateProxy)
		api.Delete("/proxies/:id", proxyController.DeleteProxy)
		api.Post("/proxies/:id/toggle", proxyController.ToggleProxy)

		// User management (admin only)
		users := api.Group("/users", middleware.AdminMiddleware())
		_ = users
		// users.Get("/", userController.GetUsers)
		// users.Post("/", userController.CreateUser)
		// users.Put("/:id", userController.UpdateUser)
		// users.Delete("/:id", userController.DeleteUser)
	}
}
