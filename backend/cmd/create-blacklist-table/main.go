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

	// Read and execute blacklisted tokens migration
	content, err := ioutil.ReadFile("database/migrations/004_create_blacklisted_tokens.sql")
	if err != nil {
		log.Fatal("Failed to read migration file:", err)
	}

	log.Println("Creating blacklisted_tokens table...")
	if _, err := db.Exec(string(content)); err != nil {
		log.Fatal("Failed to execute migration:", err)
	}

	log.Println("✅ Blacklisted tokens table created successfully!")
}
