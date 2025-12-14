package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	// Get migrations directory (default: ./configs/migrations)
	migrationsDir := "./configs/migrations"
	if len(os.Args) > 1 {
		migrationsDir = os.Args[1]
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Verify connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("✓ Connected to database")

	// Read and execute migrations
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Filter and sort SQL files
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	if len(sqlFiles) == 0 {
		fmt.Println("No migration files found")
		return
	}

	// Execute each migration
	for _, filename := range sqlFiles {
		filepath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filepath)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", filename, err)
		}

		sql := string(content)
		if strings.TrimSpace(sql) == "" {
			fmt.Printf("⊘ Skipping empty migration: %s\n", filename)
			continue
		}

		// Execute the migration
		if _, err := db.ExecContext(ctx, sql); err != nil {
			log.Fatalf("Failed to execute migration %s: %v", filename, err)
		}
		fmt.Printf("✓ Applied migration: %s\n", filename)
	}

	fmt.Printf("\n✓ All %d migrations completed successfully!\n", len(sqlFiles))
}
