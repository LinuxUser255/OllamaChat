package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Configuration
const (
	PORT = "8080"
)

// Message structure for WebSocket communication
type Message struct {
	Message string `json:"message"`
	Model   string `json:"model"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections in development
	},
}

// Main function
func main() {
	// Make sure to run "go get github.com/gorilla/mux" first if not installed
	r := mux.NewRouter()

	// API Routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/models", getModels).Methods("GET")
	api.HandleFunc("/ws", handleWebSocket)

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/").Handler(http.StripPrefix("/", fs))

	// Start server
	fmt.Printf("Server starting on port %s...\n", PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, r))
}

// WebSocket handler for real-time chat
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// Parse the incoming message
		var msg Message
		if err := json.Unmarshal(p, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Process the message with Ollama
		response := processOllamaQuery(msg.Message, msg.Model)

		// Send response back
		if err := conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
			log.Println(err)
			return
		}
	}
}

// Get available models from Ollama
func getModels(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("ollama", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, "Failed to get models", http.StatusInternalServerError)
		log.Printf("Error getting models: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

// Process query with Ollama CLI
func processOllamaQuery(query string, model string) string {
	if model == "" {
		model = "llama3.1" // Default model
	}

	cmd := exec.Command("ollama", "run", model, query)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error running Ollama: %v", err)
		return fmt.Sprintf("Error: %v", err)
	}

	// Clean up the output
	response := strings.TrimSpace(string(output))
	return response
}
