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

var sessionStore = make(map[string]string)

//go:embed templates/login.html
var loginTemplate string

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tmpl := template.Must(template.New("login").Parse(loginTemplate))
			tmpl.Execute(w, nil)
			return
		}

		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")
		token := r.FormValue("token") // optional: validate MFA separately

		var hashed, mfa_secret string
		err := db.QueryRow("SELECT password, mfa_secret FROM users WHERE username = ?", username).Scan(&hashed, &mfa_secret)
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
		http.SetCookie(w, &http.Cookie{Name: "session", Value: sid, Path: "/"})
		sessionStore[sid] = username

		redirect := r.FormValue("redirect")
		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusFound)
		} else {
			fmt.Fprintln(w, "Logged in")
		}
	}
}

func GetUsernameFromSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	username, ok := sessionStore[cookie.Value]
	if !ok {
		return "", fmt.Errorf("invalid session")
	}
	return username, nil
}
