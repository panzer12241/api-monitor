package main

import (
	"log"

	"api-monitor/app/services"
	"api-monitor/config"
	"api-monitor/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	// Connect to database
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize monitor service
	monitor := services.NewMonitorService(db)

	// Start monitor service
	monitor.Start()
	defer monitor.Stop()

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		AppName: "API Monitor v1.0",
	})

	// Setup routes
	routes.SetupRoutes(app, db, monitor)

	log.Println("🚀 API Monitor started on :8080")
	log.Fatal(app.Listen(":8080"))
}
