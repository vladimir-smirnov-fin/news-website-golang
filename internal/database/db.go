package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"  // Keep for local development
	_ "github.com/lib/pq"               // Add for PostgreSQL on Render
)

var DB *sql.DB

func InitDB() error {
	var err error
	
	// Get database URL from environment variable (Render sets this automatically)
	databaseURL := os.Getenv("DATABASE_URL")
	
	if databaseURL != "" {
		// Production mode - using PostgreSQL on Render
		fmt.Println("Connecting to PostgreSQL database...")
		DB, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return fmt.Errorf("error opening postgres connection: %v", err)
		}
		fmt.Println("Successfully connected to PostgreSQL")
	} else {
		// Local development mode - using MySQL
		fmt.Println("Connecting to MySQL database (local development)...")
		DB, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/go")
		if err != nil {
			return fmt.Errorf("error opening mysql connection: %v", err)
		}
		fmt.Println("Successfully connected to MySQL")
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection works
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	fmt.Println("Database connection pool configured successfully")
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		fmt.Println("Database connection closed")
	}
}