package auth

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitUserDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			mfa_secret TEXT NOT NULL
		)`)
	if err != nil {
		log.Fatalf("failed to create users table: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
  		id TEXT PRIMARY KEY,
  		user_id TEXT NOT NULL,
  		expires_at DATETIME NOT NULL,
  		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatalf("failed to create sessions table: %v", err)
	}

	return db
}
