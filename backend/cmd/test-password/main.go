package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Test password verification
	storedHash := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"
	testPassword := "admin123"

	fmt.Printf("Testing password verification:\n")
	fmt.Printf("Password: %s\n", testPassword)
	fmt.Printf("Hash: %s\n", storedHash)

	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(testPassword))
	if err == nil {
		fmt.Println("✅ Password verification: SUCCESS")
	} else {
		fmt.Printf("❌ Password verification: FAILED - %v\n", err)
	}

	// Generate new hash for admin123
	fmt.Println("\nGenerating new hash for 'admin123':")
	newHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to generate hash:", err)
	}
	fmt.Printf("New hash: %s\n", string(newHash))

	// Test the new hash
	err = bcrypt.CompareHashAndPassword(newHash, []byte("admin123"))
	if err == nil {
		fmt.Println("✅ New hash verification: SUCCESS")
	} else {
		fmt.Printf("❌ New hash verification: FAILED - %v\n", err)
	}
}
