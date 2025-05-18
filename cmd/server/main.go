package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func main() {
	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		fmt.Println("Client connected")
		defer ws.Close()

		buffer := make([]byte, 4096)
		for {
			n, err := ws.Read(buffer)
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			message := buffer[:n]
			log.Printf("Received: %s\n", message)

			_, err = ws.Write([]byte("Echo: " + string(message)))
			if err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}))

	fmt.Println("Relay server listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
