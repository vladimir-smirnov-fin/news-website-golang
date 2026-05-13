package database

import (
	"fmt"
	"os"
)

func MigrateDB() error {
	// Detect which database we're using
	isPostgres := os.Getenv("DATABASE_URL") != ""
	
	if isPostgres {
		return migratePostgreSQL()
	} else {
		return migrateMySQL()
	}
}

func migratePostgreSQL() error {
	// Create users table (PostgreSQL syntax)
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Create articles table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			anons TEXT NOT NULL,
			full_text TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create articles table: %v", err)
	}

	// Create indexes (PostgreSQL supports IF NOT EXISTS)
	_, err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_articles_user_id ON articles(user_id);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}

	fmt.Println("Database migration completed successfully (PostgreSQL)")
	return nil
}

func migrateMySQL() error {
	// Create users table (MySQL syntax)
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Create articles table (MySQL syntax)
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			anons TEXT NOT NULL,
			full_text TEXT NOT NULL,
			user_id INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create articles table: %v", err)
	}

	// Create indexes (MySQL doesn't support IF NOT EXISTS, so check first)
	// Check if index exists before creating
	var indexExists int
	err = DB.QueryRow(`
		SELECT COUNT(*) FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		AND table_name = 'articles' 
		AND index_name = 'idx_articles_user_id'
	`).Scan(&indexExists)
	if err == nil && indexExists == 0 {
		_, err = DB.Exec("CREATE INDEX idx_articles_user_id ON articles(user_id)")
		if err != nil {
			return fmt.Errorf("failed to create articles index: %v", err)
		}
	}

	err = DB.QueryRow(`
		SELECT COUNT(*) FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		AND table_name = 'users' 
		AND index_name = 'idx_users_email'
	`).Scan(&indexExists)
	if err == nil && indexExists == 0 {
		_, err = DB.Exec("CREATE INDEX idx_users_email ON users(email)")
		if err != nil {
			return fmt.Errorf("failed to create users index: %v", err)
		}
	}

	fmt.Println("Database migration completed successfully (MySQL)")
	return nil
}