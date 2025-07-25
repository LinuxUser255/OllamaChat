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
	"time"

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

// ChatMessage represents a message from the user with the model to use
type ChatMessage struct {
	Message   string `json:"message"`
	ModelName string `json:"model_name,omitempty"`
}

// ModelInfoResponse represents the response with available models and current model
type ModelInfoResponse struct {
	AvailableModels []string `json:"available_models"`
	CurrentModel    string   `json:"current_model"`
}

// ChatResponse represents the response from the chat API
type ChatResponse struct {
	Response string `json:"response"`
	Action   string
}

// Available models
var AVAILABLE_MODELS = []string{
	"gemma3",
	"qwen3",
	"devstral",
	"deepseek-r1",
	"deepseek-coder-v2",
	"llama4",
	"qwen2.5vl",
	"llama3.3",
	"codellama",
	"starcoder2",
	"codegemma",
	"phi4",
	"mistral",
}

// Default model
const DEFAULT_MODEL = "ollama3:8b"

// Current model
var currentModel = DEFAULT_MODEL

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
	PORT = "8888"
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

	// Get request to the homepage
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// Root endpoint
	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"message": "Ollama Chat Bot API is running",
		}); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	}).Methods("GET")

	// Models endpoint
	r.HandleFunc("/api/models", getAvailableModels).Methods("GET")

	// Retrieve actual installed models from ollama
	r.HandleFunc("/api/models/installed-models", getModels).Methods("GET")

	// Chat endpoint
	r.HandleFunc("/api/chat", handleChat).Methods("POST")

	// WebSocket endpoint for real-time chat
	r.HandleFunc("/api/chat/ws", handleWebSocket).Methods("GET")

	// Add health check endpoint
	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		// Check if Ollama is running
		cmd := exec.Command("ollama", "list")
		if err := cmd.Run(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			err := json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Ollama service is not available",
				"error":   err.Error(),
			})
			if err != nil {
				return
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "Server is running and Ollama is available",
		})
		if err != nil {
			return
		}
	}).Methods("GET")

	// Add model pull endpoint
	r.HandleFunc("/api/models/pull", handleModelPull).Methods("POST")

	// Static files
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/").Handler(http.StripPrefix("/", fs))

	// Start server
	fmt.Printf("Server starting on port %s...\n", PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, r))
}

// Get available models endpoint
func getAvailableModels(w http.ResponseWriter, r *http.Request) {
	response := ModelInfoResponse{
		AvailableModels: AVAILABLE_MODELS,
		CurrentModel:    currentModel,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

// Handle chat endpoint
func handleChat(w http.ResponseWriter, r *http.Request) {
	var chatMsg ChatMessage

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&chatMsg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(ChatResponse{
			Response: fmt.Sprintf("Error parsing request: %v", err),
		})
		if err != nil {
			return
		}
		return
	}

	// Check if we need to switch models
	if chatMsg.ModelName != "" && chatMsg.ModelName != currentModel {
		// Check if the requested model is available
		modelAvailable := false
		for _, model := range AVAILABLE_MODELS {
			if model == chatMsg.ModelName {
				modelAvailable = true
				break
			}
		}

		if !modelAvailable {
			// Check if the model exists in Ollama's repository
			cmd := exec.Command("ollama", "list")
			output, err := cmd.CombinedOutput()

			// If we can run the command, check if the model is installed
			modelInstalled := false
			if err == nil {
				lines := strings.Split(string(output), "\n")
				for i, line := range lines {
					if i == 0 || len(line) == 0 {
						continue // Skip header or empty lines
					}

					fields := strings.Fields(line)
					if len(fields) >= 1 && fields[0] == chatMsg.ModelName {
						modelInstalled = true
						break
					}
				}
			}

			if modelInstalled {
				// Model is installed but not in our list, add it
				AVAILABLE_MODELS = append(AVAILABLE_MODELS, chatMsg.ModelName)
				currentModel = chatMsg.ModelName
			} else {
				// Model is not installed, suggest pulling it, then offer to pull it
				w.WriteHeader(http.StatusNotFound)
				err := json.NewEncoder(w).Encode(ChatResponse{
					Response: fmt.Sprintf("Model %s is not available. Would you like to pull it from Ollama's repository?", chatMsg.ModelName),
					Action:   fmt.Sprintf("pull:%s", chatMsg.ModelName),
				})
				if err != nil {
					return
				}
				return
			}
		} else {
			// Update the current model
			currentModel = chatMsg.ModelName
		}
	}

	// Process the message with Ollama
	response := processOllamaQueryWithLangChain(chatMsg.Message, currentModel)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(ChatResponse{
		Response: response,
	})
	if err != nil {
		return
	}
}

// WebSocket handler for real-time chat
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Println("WebSocket connection attempt...")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	log.Println("WebSocket connection established successfully")
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing WebSocket: %v", err)
		}
	}(conn)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			return
		}

		log.Printf("Received message type: %d, content: %s", messageType, string(p))

		// Parse the incoming message
		var chatMsg ChatMessage
		if err := json.Unmarshal(p, &chatMsg); err != nil {
			log.Printf("Error parsing message: %v", err)

			// Try parsing with the old Message structure as fallback
			var oldMsg Message
			if jsonErr := json.Unmarshal(p, &oldMsg); jsonErr != nil {
				log.Printf("Also failed to parse as old message format: %v", jsonErr)
				continue
			}

			// Convert old format to new format
			chatMsg.Message = oldMsg.Message
			chatMsg.ModelName = oldMsg.Model
			log.Println("Successfully parsed using old message format")
		}

		// Check if we need to switch models
		if chatMsg.ModelName != "" && chatMsg.ModelName != currentModel {
			// Check if the requested model is available
			modelAvailable := false
			for _, model := range AVAILABLE_MODELS {
				if model == chatMsg.ModelName {
					modelAvailable = true
					break
				}
			}

			if !modelAvailable {
				// Check if the model exists in Ollama's repository
				cmd := exec.Command("ollama", "list")
				output, err := cmd.CombinedOutput()

				// If we can run the command, check if the model is installed
				modelInstalled := false
				if err == nil {
					lines := strings.Split(string(output), "\n")
					for i, line := range lines {
						if i == 0 || len(line) == 0 {
							continue // Skip header or empty lines
						}

						fields := strings.Fields(line)
						if len(fields) >= 1 && fields[0] == chatMsg.ModelName {
							modelInstalled = true
							break
						}
					}
				}

				if modelInstalled {
					// Model is installed but not in our list, add it
					AVAILABLE_MODELS = append(AVAILABLE_MODELS, chatMsg.ModelName)
					currentModel = chatMsg.ModelName
				} else {
					// Model is not installed, ask user if they want to pull it
					response := ChatResponse{
						Response: fmt.Sprintf("Model %s is not available. Would you like to pull it from Ollama's repository?", chatMsg.ModelName),
						Action:   fmt.Sprintf("pull:%s", chatMsg.ModelName),
					}

					responseJSON, err := json.Marshal(response)
					if err != nil {
						log.Printf("Error marshaling response: %v", err)
						continue
					}

					if err := conn.WriteMessage(websocket.TextMessage, responseJSON); err != nil {
						log.Println("WebSocket write error:", err)
						return
					}
					continue
				}
			} else {
				// Update the current model
				currentModel = chatMsg.ModelName
			}
		}

		log.Printf("Processing query with model: %s", currentModel)

		// Process the message with Ollama using langchaingo
		response := processOllamaQueryWithLangChain(chatMsg.Message, currentModel)

		log.Printf("Got response (length: %d)", len(response))

		// Send response back
		if err := conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
			log.Println("WebSocket write error:", err)
			return
		}

		log.Println("Response sent successfully")
	}
}

// Get installed models from Ollama and return as JSON
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
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// pullOllamaModel pulls a model by running this shell script located in this repo: resources/run_ollama.sh
func pullOllamaModel(modelName string) (string, error) {
	log.Printf("Attempting to pull model: %s", modelName)

	// Create command to pull the model using resources/run_ollama.sh
	cmd := exec.Command("sh", "./resources/run_ollama.sh", modelName)
	cmd.Dir = "./resources/run_ollama.sh"

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error pulling model %s: %v", modelName, err)
		return "", fmt.Errorf("failed to pull model: %v", err)
	}

	log.Printf("Successfully pulled model %s", modelName)

	// Add the model to available models list
	AVAILABLE_MODELS = append(AVAILABLE_MODELS, modelName)

	return string(output), nil
}

// handleModelPull handles the API endpoint for pulling a model
func handleModelPull(w http.ResponseWriter, r *http.Request) {
	// Extract model name from request
	var pullRequest struct {
		ModelName string `json:"model_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&pullRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request format",
		})
		if err != nil {
			return
		}
		return
	}

	// Pull the model
	output, err := pullOllamaModel(pullRequest.ModelName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		if err != nil {
			return
		}
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Successfully pulled model %s", pullRequest.ModelName),
		"details": output,
	})
	if err != nil {
		return
	}
}

// processOllamaQueryWithLangChain processes a query using the Ollama LLM through langchaingo
func processOllamaQueryWithLangChain(query string, modelName string) string {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Format the prompt with the system template
	formattedPrompt := FormatPrompt(query)

	// Initialize the Ollama LLM client
	llm, err := ollama.New(
		ollama.WithModel(modelName),
	)
	if err != nil {
		log.Printf("Error initializing Ollama: %v", err)
		return fmt.Sprintf("Error initializing Ollama: %v", err)
	}

	// Generate a response from the LLM
	response, err := llm.Call(ctx, formattedPrompt, llms.WithTemperature(0.7))
	if err != nil {
		log.Printf("Error calling Ollama: %v", err)
		return fmt.Sprintf("Error generating response: %v", err)
	}

	return response
}
