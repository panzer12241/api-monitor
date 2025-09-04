package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Generate hashes for both admin and user passwords
	passwords := map[string]string{
		"admin": "admin123",
		"user":  "user123",
	}

	for username, password := range passwords {
		fmt.Printf("Generating hash for %s (password: %s):\n", username, password)

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to generate hash:", err)
		}

		fmt.Printf("Hash: %s\n", string(hash))

		// Verify the hash
		err = bcrypt.CompareHashAndPassword(hash, []byte(password))
		if err == nil {
			fmt.Println("✅ Verification: SUCCESS")
		} else {
			fmt.Printf("❌ Verification: FAILED - %v\n", err)
		}
		fmt.Println()
	}
}
