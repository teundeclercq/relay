package main

import (
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

func main() {
	origin := "http://localhost/"
	url := "ws://your-server-ip-or-domain:8081/ws" // change this to your real IP or domain

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatalf("WebSocket dial failed: %v", err)
	}
	defer ws.Close()

	message := "Hello from client"
	_, err = ws.Write([]byte(message))
	if err != nil {
		log.Fatalf("Write error: %v", err)
	}

	var reply = make([]byte, 4096)
	n, err := ws.Read(reply)
	if err != nil {
		log.Fatalf("Read error: %v", err)
	}

	fmt.Printf("Server replied: %s\n", reply[:n])
}
