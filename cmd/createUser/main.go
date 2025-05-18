package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: createuser <username> <password> <issuer>")
		return
	}
	username := os.Args[1]
	password := os.Args[2]
	issuer := os.Args[3]

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	mfaSecret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
	})
	if err != nil {
		log.Fatalf("Failed to generate MFA secret: %v", err)
	}

	db, err := sql.Open("sqlite3", "oauth.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE,
		password TEXT,
		mfa_secret TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO users (id, username, password, mfa_secret) VALUES (?, ?, ?, ?)`,
		username+"-id", username, string(hash), mfaSecret.Secret())
	if err != nil {
		log.Fatalf("Insert error: %v", err)
	}

	fmt.Println("User created successfully.")
	fmt.Println("MFA secret:", mfaSecret.Secret())
	fmt.Println("Scan QR code using this URL in your authenticator app:")
	fmt.Println(mfaSecret.URL())
}
