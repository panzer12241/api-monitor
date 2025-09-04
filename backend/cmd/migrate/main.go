package main

import (
	"fmt"
	"log"
	"os"

	"api-monitor/config"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate/main.go fresh        # Drop all tables and run migrations")
		fmt.Println("  go run cmd/migrate/main.go fresh-seed   # Drop all tables, run migrations, and seed data")
		fmt.Println("  go run cmd/migrate/main.go run          # Run pending migrations only")
		fmt.Println("  go run cmd/migrate/main.go seed         # Run seeders only")
		fmt.Println("")
		fmt.Println("Or use Makefile:")
		fmt.Println("  make migrate-fresh        # Fresh migration")
		fmt.Println("  make migrate-fresh-seed   # Fresh migration with seeding")
		fmt.Println("  make migrate-run          # Run migrations")
		fmt.Println("  make seed                 # Run seeders")
		os.Exit(1)
	}

	command := os.Args[1]

	// Connect to database
	db, err := config.ConnectDBWithoutMigration()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	switch command {
	case "fresh":
		if err := config.FreshMigrate(db); err != nil {
			log.Fatal("Fresh migration failed:", err)
		}
		fmt.Println("✅ Fresh migration completed successfully!")

	case "fresh-seed":
		if err := config.FreshMigrateWithSeeder(db); err != nil {
			log.Fatal("Fresh migration with seeding failed:", err)
		}
		fmt.Println("✅ Fresh migration with seeding completed successfully!")

	case "run":
		if err := config.RunMigrations(db); err != nil {
			log.Fatal("Migration failed:", err)
		}
		fmt.Println("✅ Migrations completed successfully!")

	case "seed":
		if err := config.RunSeeders(db); err != nil {
			log.Fatal("Seeding failed:", err)
		}
		fmt.Println("✅ Seeding completed successfully!")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: fresh, fresh-seed, run, seed")
		os.Exit(1)
	}
}
