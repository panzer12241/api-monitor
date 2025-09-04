package main

import (
	"fmt"
	"log"
	"strings"

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

	// Query all users
	rows, err := db.Query("SELECT id, username, password, role, is_active FROM users")
	if err != nil {
		log.Fatal("Failed to query users:", err)
	}
	defer rows.Close()

	fmt.Println("=== Users in database ===")
	fmt.Printf("%-5s %-15s %-70s %-10s %-8s\n", "ID", "Username", "Password Hash", "Role", "Active")
	fmt.Println(strings.Repeat("-", 110))

	for rows.Next() {
		var id int
		var username, passwordHash, role string
		var isActive bool

		if err := rows.Scan(&id, &username, &passwordHash, &role, &isActive); err != nil {
			log.Fatal("Failed to scan row:", err)
		}

		fmt.Printf("%-5d %-15s %-70s %-10s %-8t\n", id, username, passwordHash, role, isActive)
	}

	if err := rows.Err(); err != nil {
		log.Fatal("Row iteration error:", err)
	}
}
