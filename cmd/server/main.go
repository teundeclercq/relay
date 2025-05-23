package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"relay/pkg/auth"
	"relay/pkg/proxy"
)

func main() {
	port := flag.String("port", "8081", "Port to listen on")
	flag.Parse()

	db := auth.InitUserDB("oauth.db")
	auth.InitOAuthServer()

	http.HandleFunc("/login", auth.LoginHandler(db))

	http.Handle("/ws", proxy.HandleTunnel())

	http.HandleFunc("/", proxy.HandleProxy())

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Relay server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr,
		nil))
}
