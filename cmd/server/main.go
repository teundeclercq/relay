package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

var mu sync.Mutex
var clientConn *websocket.Conn

func main() {
	port := flag.String("port", "8081", "Port to listen on")
	flag.Parse()

	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		fmt.Println("Client connected")
		mu.Lock()
		clientConn = ws
		mu.Unlock()
	}))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if clientConn == nil {
			mu.Unlock()
			http.Error(w, "No client connected", http.StatusServiceUnavailable)
			return
		}

		log.Printf("Received request: %s %s", r.Method, r.URL)
		// Forward the full HTTP request to the client
		err := r.Write(clientConn)
		if err != nil {
			mu.Unlock()
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
			return
		}

		// Read the HTTP response from the client
		resp, err := http.ReadResponse(bufio.NewReader(clientConn), r)
		mu.Unlock()
		if err != nil {
			log.Printf("Failed to read response from client: %v\n", err)
			http.Error(w, "Failed to read response from client", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Relay server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
