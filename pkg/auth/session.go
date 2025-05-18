package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func CreateSession(db *sql.DB, userID string) (string, error) {
	sessionID := generateSessionID()
	expires := time.Now().Add(24 * time.Hour)
	_, err := db.Exec(`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		sessionID, userID, expires)
	return sessionID, err
}

func GetUsernameFromSessionDB(db *sql.DB, r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	var username string
	var expires time.Time
	err = db.QueryRow(`
		SELECT u.username, s.expires_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = ?`, cookie.Value).Scan(&username, &expires)
	if err != nil {
		return "", err
	}
	if time.Now().After(expires) {
		return "", errors.New("session expired")
	}
	return username, nil
}

func generateSessionID() string {
	return fmt.Sprintf("sess-%d", time.Now().UnixNano())
}
