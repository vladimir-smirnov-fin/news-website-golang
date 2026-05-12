package database

import (
	"os"
)

var isPostgreSQL bool

func init() {
	// If DATABASE_URL exists, we're on Render using PostgreSQL
	_, isPostgreSQL = os.LookupEnv("DATABASE_URL")
}

// GetPlaceholder returns the appropriate placeholder for the database
func GetPlaceholder(index int) string {
	if isPostgreSQL {
		return "$" + string(rune('1'+index))
	}
	return "?"
}