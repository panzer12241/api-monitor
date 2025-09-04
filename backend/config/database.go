package config

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/lib/pq"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetJWTSecret() []byte {
	return []byte(GetEnv("JWT_SECRET", "your-secret-key-change-this-in-production"))
}

func ConnectDB() (*sql.DB, error) {
	host := GetEnv("DB_HOST", "localhost")
	port := GetEnv("DB_PORT", "5432")
	user := GetEnv("DB_USERNAME", "postgres")
	password := GetEnv("DB_PASSWORD", "postgres123")
	dbname := GetEnv("DB_DATABASE", "api_monitor")

	// Use sslmode=require for production databases like DigitalOcean
	sslmode := GetEnv("DB_SSLMODE", "disable")
	if host != "localhost" && host != "127.0.0.1" {
		sslmode = "require"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connected successfully")

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Printf("Warning: Failed to run migrations: %v", err)
	}

	return db, nil
}

// ConnectDBWithoutMigration connects to database without running migrations
func ConnectDBWithoutMigration() (*sql.DB, error) {
	host := GetEnv("DB_HOST", "localhost")
	port := GetEnv("DB_PORT", "5432")
	user := GetEnv("DB_USERNAME", "postgres")
	password := GetEnv("DB_PASSWORD", "postgres123")
	dbname := GetEnv("DB_DATABASE", "api_monitor")

	// Use sslmode=require for production databases like DigitalOcean
	sslmode := GetEnv("DB_SSLMODE", "disable")
	if host != "localhost" && host != "127.0.0.1" {
		sslmode = "require"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connected successfully")
	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrationsDir := "database/migrations"

	// Read all .sql files from migrations directory
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %v", err)
	}

	// Sort files to ensure they run in order
	sort.Strings(files)

	for _, file := range files {
		log.Printf("Running migration: %s", filepath.Base(file))

		// Read migration file
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", file, err)
		}

		// Execute migration
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %v", file, err)
		}

		log.Printf("Migration completed: %s", filepath.Base(file))
	}

	log.Println("All migrations completed successfully")
	return nil
}

// RunMigrations - exported version of runMigrations for CLI tool
func RunMigrations(db *sql.DB) error {
	return runMigrations(db)
}

// FreshMigrate drops all tables and runs migrations fresh (like Laravel's migrate:fresh)
func FreshMigrate(db *sql.DB) error {
	log.Println("Starting fresh migration (dropping all tables)...")

	// Drop all tables in the correct order to avoid foreign key constraints
	dropStatements := []string{
		"DROP TABLE IF EXISTS api_check_logs CASCADE;",
		"DROP TABLE IF EXISTS api_endpoints CASCADE;",
		"DROP TABLE IF EXISTS users CASCADE;",
		"DROP TABLE IF EXISTS proxies CASCADE;",
	}

	for _, stmt := range dropStatements {
		log.Printf("Executing: %s", stmt)
		if _, err := db.Exec(stmt); err != nil {
			log.Printf("Warning: Failed to execute %s: %v", stmt, err)
		}
	}

	log.Println("All tables dropped. Running migrations...")

	// Run migrations
	return runMigrations(db)
}

// RunSeeders runs all SQL seeder files
func RunSeeders(db *sql.DB) error {
	seedersDir := "database/seeders"

	log.Println("Starting to run seeders...")

	// Read all .sql files from seeders directory
	files, err := filepath.Glob(filepath.Join(seedersDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read seeder files: %v", err)
	}

	// Sort files to ensure they run in order
	sort.Strings(files)

	if len(files) == 0 {
		log.Println("No seeder files found")
		return nil
	}

	for _, file := range files {
		log.Printf("Running seeder: %s", filepath.Base(file))

		// Read seeder file
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read seeder file %s: %v", file, err)
		}

		// Execute seeder
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute seeder %s: %v", file, err)
		}

		log.Printf("Seeder completed: %s", filepath.Base(file))
	}

	log.Println("All seeders completed successfully")
	return nil
}

// FreshMigrateWithSeeder drops all tables, runs migrations, and seeds data (like Laravel's migrate:fresh --seed)
func FreshMigrateWithSeeder(db *sql.DB) error {
	// Run fresh migration first
	if err := FreshMigrate(db); err != nil {
		return fmt.Errorf("fresh migration failed: %v", err)
	}

	// Then run seeders
	if err := RunSeeders(db); err != nil {
		return fmt.Errorf("seeders failed: %v", err)
	}

	return nil
}
