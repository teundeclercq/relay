package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/websocket"
)

func main() {
	relayHost := flag.String("server", "localhost:8081", "Relay server address (host:port)")
	localTarget := flag.String("local", "localhost:8080", "Local service to tunnel (host:port)")
	flag.Parse()

	wsUrl := fmt.Sprintf("wss://%s/ws", *relayHost)
	origin := fmt.Sprintf("https://%s", *relayHost)

	log.Printf("Connecting to relay server at %s\n", wsUrl)
	ws, err := websocket.Dial(wsUrl, "", origin)
	if err != nil {
		log.Fatalf("WebSocket dial failed: %v", err)
	}
	defer ws.Close()

	log.Println("Tunnel established. Waiting for requests...")

	for {
		req, err := http.ReadRequest(bufio.NewReader(ws))
		if err != nil {
			log.Printf("Read request error: %v", err)
			return
		}

		localConn, err := net.Dial("tcp", *localTarget)
		if err != nil {
			log.Printf("Local connection failed: %v", err)
			return
		}

		err = req.Write(localConn)
		if err != nil {
			log.Printf("Write to local failed: %v", err)
			localConn.Close()
			continue
		}

		resp, err := http.ReadResponse(bufio.NewReader(localConn), req)
		localConn.Close()
		if err != nil {
			log.Printf("Read local response failed: %v", err)
			continue
		}

		err = resp.Write(ws)
		if err != nil {
			log.Printf("Write back to relay failed: %v", err)
			return
		}
	}
}
