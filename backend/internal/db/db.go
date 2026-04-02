package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB initializes the SQLite database and runs migrations
func InitDB(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	return createTables()
}

func createTables() error {
	// 1. Create logs table
	createLogsTable := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		level TEXT,
		source_app TEXT,
		message TEXT
	);`

	// 2. Create FTS5 virtual table for fast full-text search
	// We use 'content=logs' to keep it synchronized without duplicating data heavily
	createLogsFTSTable := `
	CREATE VIRTUAL TABLE IF NOT EXISTS logs_fts USING fts5(
		message,
		content=logs,
		content_rowid=id
	);`

	// 3. Triggers to keep FTS table in sync
	createTriggers := `
	CREATE TRIGGER IF NOT EXISTS logs_ai AFTER INSERT ON logs BEGIN
	  INSERT INTO logs_fts(rowid, message) VALUES (new.id, new.message);
	END;

	CREATE TRIGGER IF NOT EXISTS logs_ad AFTER DELETE ON logs BEGIN
	  INSERT INTO logs_fts(logs_fts, rowid, message) VALUES ('delete', old.id, old.message);
	END;

	CREATE TRIGGER IF NOT EXISTS logs_au AFTER UPDATE ON logs BEGIN
	  INSERT INTO logs_fts(logs_fts, rowid, message) VALUES ('delete', old.id, old.message);
	  INSERT INTO logs_fts(rowid, message) VALUES (new.id, new.message);
	END;
	`

	// 4. Admin Users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password_hash TEXT,
		requires_password_change BOOLEAN DEFAULT 0
	);`

	queries := []string{createLogsTable, createLogsFTSTable, createTriggers, createUsersTable}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			log.Printf("Error executing query: %s\nError: %v", query, err)
			return err
		}
	}

	// Try to add column if it doesn't exist (for existing MVP DBs)
	DB.Exec("ALTER TABLE users ADD COLUMN requires_password_change BOOLEAN DEFAULT 0;")

	log.Println("Database initialized and migrated successfully.")
	return nil
}
