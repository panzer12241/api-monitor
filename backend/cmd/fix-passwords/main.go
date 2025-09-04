package main

import (
	"io/ioutil"
	"log"

	"api-monitor/config"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	// Connect to database
	db, err := config.ConnectDBWithoutMigration()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Read and execute password fix migration
	content, err := ioutil.ReadFile("database/migrations/003_fix_user_passwords.sql")
	if err != nil {
		log.Fatal("Failed to read migration file:", err)
	}

	log.Println("Executing password fix migration...")
	if _, err := db.Exec(string(content)); err != nil {
		log.Fatal("Failed to execute migration:", err)
	}

	log.Println("✅ Password fix migration completed successfully!")
}
