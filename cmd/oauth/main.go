package main

import (
	"fmt"
	"log"
	"net/http"
	"relay/pkg/auth"
)

func main() {
	srv := auth.InitOAuthServer()
	db := auth.InitUserDB("oauth.db")
	http.HandleFunc("/login", auth.LoginHandler(db))
	http.HandleFunc("/authorize", auth.AuthorizeHandler(srv))
	http.HandleFunc("/token", auth.TokenHandler(srv))
	http.HandleFunc("/validate", auth.ValidateHandler(srv))

	fmt.Println("Starting Auth Server on :9096")
	log.Fatal(http.ListenAndServe(":9096", nil))
}
