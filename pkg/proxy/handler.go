package proxy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

var (
	mu         sync.Mutex
	clientConn io.ReadWriteCloser // You may need to adjust the type as needed
	// ws should be defined elsewhere or passed in
)

func HandleProxy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func HandleTunnel() websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		fmt.Println("Client connected")
		mu.Lock()
		clientConn = ws
		mu.Unlock()

		select {}
	})
}
