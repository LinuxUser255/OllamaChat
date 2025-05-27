# Detailed Analysis and Explanation of main.go

## Overview

The `main.go` file implements a web server that provides an interface to interact with Ollama's large language models. It creates a bridge between a web frontend and the Ollama LLM backend, offering both HTTP and WebSocket endpoints for real-time chat functionality.

## Key Components

### 1. Package Structure and Imports

```go
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
```

The code imports several standard Go packages for handling HTTP requests, JSON encoding/decoding, string manipulation, and template processing. It also imports third-party packages:
- `gorilla/mux`: For HTTP routing
- `gorilla/websocket`: For WebSocket support
- `tmc/langchaingo`: For interacting with language models through LangChain

### 2. Data Structures

```go
type ModelInfo struct {
    Name    string `json:"name"`
    Size    string `json:"size,omitempty"`
    ModTime string `json:"modified,omitempty"`
}

type ChatMessage struct {
    Message   string `json:"message"`
    ModelName string `json:"model_name,omitempty"`
}

type ModelInfoResponse struct {
    AvailableModels []string `json:"available_models"`
    CurrentModel    string   `json:"current_model"`
}

type ChatResponse struct {
    Response string `json:"response"`
}

type Message struct {
    Message string `json:"message"`
    Model   string `json:"model"`
}

type PromptData struct {
    Query string
}
```

These structures define the data formats used throughout the application:
- `ModelInfo`: Information about an Ollama model
- `ChatMessage`: A message from the user with the model to use
- `ModelInfoResponse`: Response containing available models and current model
- `ChatResponse`: Response from the chat API
- `Message`: Legacy structure for WebSocket communication
- `PromptData`: Data for template processing

### 3. Global Variables and Constants

```go
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

const DEFAULT_MODEL = "ollama3:8b"
var currentModel = DEFAULT_MODEL

const SystemTemplate = `You are a helpful coding assistant...`
const PORT = "8888"

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all connections in development
    },
}
```

These define:
- Available models that can be selected by users
- Default model and current model state
- System prompt template for the LLM
- Server port
- WebSocket upgrader configuration

### 4. Main Function

```go
func main() {
    r := mux.NewRouter()
    
    // Route definitions...
    
    // Start server
    fmt.Printf("Server starting on port %s...\n", PORT)
    log.Fatal(http.ListenAndServe(":"+PORT, r))
}
```

The main function:
1. Creates a new router using gorilla/mux
2. Defines various HTTP routes
3. Starts the HTTP server on the specified port

### 5. Route Handlers

#### Root and API Endpoints

```go
r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./static/index.html")
})

r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Ollama Chat Bot API is running",
    })
}).Methods("GET")
```

These handlers:
- Serve the main HTML file for the root path
- Provide a simple API status endpoint

#### Models Endpoint

```go
r.HandleFunc("/api/models", getAvailableModels).Methods("GET")

func getAvailableModels(w http.ResponseWriter, r *http.Request) {
    response := ModelInfoResponse{
        AvailableModels: AVAILABLE_MODELS,
        CurrentModel:    currentModel,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

This endpoint returns the list of available models and the currently selected model.

#### Chat Endpoint (HTTP)

```go
r.HandleFunc("/api/chat", handleChat).Methods("POST")

func handleChat(w http.ResponseWriter, r *http.Request) {
    var chatMsg ChatMessage
    
    // Decode request body
    // Check model availability
    // Process message with Ollama
    // Send response
}
```

This endpoint:
1. Parses the incoming JSON request
2. Validates the requested model
3. Updates the current model if needed
4. Processes the message using the Ollama LLM
5. Returns the response as JSON

#### WebSocket Endpoint

```go
r.HandleFunc("/api/chat/ws", handleWebSocket).Methods("GET")

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Upgrade HTTP connection to WebSocket
    // Listen for incoming messages in a loop
    // Parse messages and validate models
    // Process messages with Ollama
    // Send responses back through WebSocket
}
```

This handler:
1. Upgrades the HTTP connection to a WebSocket
2. Listens for incoming messages in a continuous loop
3. Parses the messages and validates the requested model
4. Processes the messages using the Ollama LLM
5. Sends responses back through the WebSocket

#### Health Check Endpoint

```go
r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
    // Check if Ollama is running
    cmd := exec.Command("ollama", "list")
    if err := cmd.Run(); err != nil {
        // Return error if Ollama is not available
    }
    
    // Return success if Ollama is available
}).Methods("GET")
```

This endpoint:
1. Checks if the Ollama service is running by executing `ollama list`
2. Returns appropriate status based on the result

### 6. Static File Serving

```go
fs := http.FileServer(http.Dir("./static"))
r.PathPrefix("/").Handler(http.StripPrefix("/", fs))
```

This sets up serving of static files (HTML, CSS, JS) from the `./static` directory.

### 7. Ollama Integration

```go
func processOllamaQueryWithLangChain(query string, modelName string) string {
    // Create context with timeout
    // Format prompt with system template
    // Initialize Ollama LLM client
    // Generate response from LLM
    // Return response
}
```

This function:
1. Creates a context with a timeout
2. Formats the user query with the system prompt template
3. Initializes the Ollama LLM client with the specified model
4. Generates a response from the LLM
5. Returns the formatted response

### 8. Helper Functions

```go
func FormatPrompt(query string) string {
    // Parse template
    // Execute template with query
    // Return formatted prompt
}

func getModels(w http.ResponseWriter, r *http.Request) {
    // Execute ollama list command
    // Parse output to extract model names
    // Return as JSON
}
```

These functions:
- Format the system prompt with the user query
- Get the list of installed Ollama models

## Technical Analysis

### Architecture Pattern
The application follows a classic web server architecture with RESTful API endpoints and WebSocket support. It uses a router to direct requests to appropriate handlers and maintains a stateful connection for real-time chat through WebSockets.

### State Management
The application maintains minimal state:
- `currentModel`: Tracks the currently selected model
- WebSocket connections: Maintained for each client

### Error Handling
The code includes error handling for:
- JSON parsing errors
- Model validation
- Ollama service availability
- WebSocket communication errors
- LLM processing errors

### Concurrency
The application handles concurrent requests through Go's built-in HTTP server concurrency model. Each request is processed in its own goroutine. WebSocket connections are maintained in separate goroutines as well.

### Security Considerations
- The WebSocket upgrader allows all origins (`CheckOrigin` returns true), which is suitable for development but should be restricted in production.
- There's no authentication or rate limiting implemented.

### Integration with Ollama
The application integrates with Ollama in two ways:
1. Through the LangChain Go library for generating responses
2. By executing the `ollama list` command to check service availability

## Potential Improvements

1. **Authentication**: Add user authentication for secure access
2. **Rate Limiting**: Implement rate limiting to prevent abuse
3. **Streaming Responses**: Implement streaming for large responses
4. **Conversation History**: Add support for maintaining conversation context
5. **Error Recovery**: Implement more robust error recovery mechanisms
6. **Configuration**: Move hardcoded values to configuration files
7. **Metrics and Logging**: Add comprehensive logging and metrics collection
8. **Testing**: Add unit and integration tests

## Conclusion

The `main.go` file implements a well-structured web server that provides a bridge between a web frontend and Ollama's language models. It offers both HTTP and WebSocket endpoints for flexible integration and real-time chat capabilities. The code follows Go best practices and provides a solid foundation for a chat application powered by large language models.