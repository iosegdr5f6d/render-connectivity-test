package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type ServerMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Time string `json:"time,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// This disposable tester accepts connections from embedded pages
	// regardless of their browser origin. Do not reuse this setting
	// unchanged for a real application.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/health", healthHandler)
	http.HandleFunc("/ws", websocketHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	address := "0.0.0.0:" + port

	log.Printf("Connectivity test server listening on %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "Connectivity test backend is running.")
	fmt.Fprintln(w, "HTTP health endpoint: /api/health")
	fmt.Fprintln(w, "WebSocket endpoint: /ws")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	response := map[string]interface{}{
		"ok":      true,
		"service": "connectivity-test",
		"time":    time.Now().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Health response error: %v", err)
	}
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Limit incoming messages so this test endpoint cannot be used
	// to send arbitrarily large payloads.
	conn.SetReadLimit(2048)

	log.Printf("WebSocket client connected")

	err = conn.WriteJSON(ServerMessage{
		Type: "status",
		Text: "WebSocket connected. Send a test message.",
	})
	if err != nil {
		return
	}

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket client disconnected: %v", err)
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		message := strings.TrimSpace(string(payload))
		if message == "" {
			continue
		}

		if len(message) > 1000 {
			err := conn.WriteJSON(ServerMessage{
				Type: "error",
				Text: "Message too long. The limit is 1000 bytes.",
			})
			if err != nil {
				return
			}
			continue
		}

		// Echo the message to the same client that sent it.
		// Nothing is stored or forwarded to other clients.
		err = conn.WriteJSON(ServerMessage{
			Type: "echo",
			Text: "Backend received: " + message,
			Time: time.Now().Format("15:04:05"),
		})
		if err != nil {
			return
		}
	}
}
