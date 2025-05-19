package auth

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

//go:embed templates/login.html
var loginTemplate string

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "relay.tdccore.nl" {
			http.Error(w, "Invalid host", http.StatusForbidden)
			return
		}

		if r.Method == http.MethodGet {
			tmpl := template.Must(template.New("login").Parse(loginTemplate))
			tmpl.Execute(w, nil)
			return
		}

		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")
		token := r.FormValue("token") // optional: validate MFA separately

		var userID, hashed, mfa_secret string
		err := db.QueryRow("SELECT id, password, mfa_secret FROM users WHERE username = ?", username).Scan(&userID, &hashed, &mfa_secret)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) != nil {
			tmpl := template.Must(template.New("login").Parse(loginTemplate))
			tmpl.Execute(w, map[string]string{"Error": "Invalid credentials"})
			return
		}

		if !totp.Validate(token, mfa_secret) {
			tmpl := template.Must(template.New("login").Parse(loginTemplate))
			tmpl.Execute(w, map[string]string{"Error": "Invalid MFA code"})
			return
		}

		sid := fmt.Sprintf("sess-%d", time.Now().UnixNano())
		expiry := time.Now().Add(24 * time.Hour)

		_, err = db.Exec(`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
			sid, userID, expiry)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
		// Set session cookie
		http.SetCookie(w, &http.Cookie{Name: "session", Value: sid, Path: "/"})

		redirect := r.FormValue("redirect")
		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusFound)
		} else {
			fmt.Fprintln(w, "Logged in")
		}
	}
}

func GetUsernameFromSession(db *sql.DB, r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	var username string
	var expires time.Time

	err = db.QueryRow(`
  		SELECT u.username, s.expires_at
  		FROM sessions s
  		JOIN users u ON s.user_id = u.id
  		WHERE s.id = ?`, cookie.Value).Scan(&username, &expires)

	if err != nil || time.Now().After(expires) {
		return "", fmt.Errorf("invalid or expired session")
	}
	return username, nil
}
