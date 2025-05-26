package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// ModelInfo represents information about an Ollama model
type ModelInfo struct {
	Name    string `json:"name"`
	Size    string `json:"size,omitempty"`
	ModTime string `json:"modified,omitempty"`
}

// Available models
var models = []string{
	"gemma3", "qwen3", "devstral", "deepseek-r1", "deepseek-coder-v2", "llama4", "qwen2.5vl", "llama3.3", "phi4", "mistral",
}

const SystemTemplate = `You are a helpful coding assistant. When providing code examples:
1. Always use proper markdown formatting with language-specific syntax highlighting
2. Use triple backticks with the language name for code blocks (e.g. "` + "```" + `python")
3. Format code in a clean, readable way with proper indentation
4. Use VSCode-style syntax highlighting conventions

User Query: {{.Query}}
`

// PromptData holds the data to be inserted into the template
type PromptData struct {
	Query string
}

// FormatPrompt formats the system prompt with the user query
func FormatPrompt(query string) string {
	tmpl, err := template.New("prompt").Parse(SystemTemplate)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, PromptData{Query: query})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return ""
	}

	return buf.String()
}

// PORT Configuration
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
			log.Printf("Error closing WebSocket: %v", err)
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

		// Process the message with Ollama using langchaingo
		response := processOllamaQueryWithLangChain(msg.Message, msg.Model)

		// Send response back
		if err := conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
			log.Println(err)
			return
		}
	}
}

// Get available models from Ollama and return as JSON
func getModels(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("ollama", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, "Failed to get models", http.StatusInternalServerError)
		log.Printf("Error getting models: %v", err)
		return
	}

	// Parse the output to extract model names
	lines := strings.Split(string(output), "\n")
	var modelList []ModelInfo
	
	// Skip header line and process each model line
	for i, line := range lines {
		if i == 0 || len(line) == 0 {
			continue // Skip header or empty lines
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			modelList = append(modelList, ModelInfo{
				Name:    fields[0],
				Size:    fields[1],
				ModTime: strings.Join(fields[2:], " "),
			})
		}
	}
	
	// Convert to JSON and send response
	jsonResponse, err := json.Marshal(modelList)
	if err != nil {
		http.Error(w, "Failed to marshal model list", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// Initialize an Ollama model client
func initializeOllamaModel(modelName string) (*ollama.LLM, error) {
	if modelName == "" {
		modelName = "llama3" // Default model
	}
	
	// Create a new Ollama LLM client
	ollamaLLM, err := ollama.New(
		ollama.WithModel(modelName),
		ollama.WithServerURL("http://localhost:11434"),
	)
	
	if err != nil {
		log.Printf("Error initializing Ollama model %s: %v", modelName, err)
		return nil, err
	}
	
	return ollamaLLM, nil
}

// Process query with Ollama using langchaingo
func processOllamaQueryWithLangChain(query string, modelName string) string {
	// Initialize the model
	model, err := initializeOllamaModel(modelName)
	if err != nil {
		return fmt.Sprintf("Error initializing model: %v", err)
	}
	
	// Create a formatted prompt
	prompt := FormatPrompt(query)
	
	// Call the model
	ctx := context.Background()
	response, err := model.Call(ctx, prompt, llms.WithTemperature(0.7))
	
	if err != nil {
		log.Printf("Error calling Ollama model: %v", err)
		return fmt.Sprintf("Error: %v", err)
	}
	
	return response
}
