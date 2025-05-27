# Flow Chart Diagram for Ollama AI Assistant Application

Here's a flow chart diagram that illustrates how the Ollama AI Assistant application works:

```mermaid
flowchart TD
    A[Client Browser] -->|HTTP Request| B[Go Web Server]
    
    subgraph "Backend (Go Server)"
        B --> C{Request Type}
        
        C -->|GET /| D[Serve index.html]
        C -->|GET /api/models| E[Get Available Models]
        C -->|POST /api/chat| F[Handle Chat Request]
        C -->|GET /api/chat/ws| G[WebSocket Connection]
        C -->|GET /api/health| H[Health Check]
        
        E --> E1[Return AVAILABLE_MODELS list]
        
        F --> F1[Parse ChatMessage]
        F1 --> F2{Valid Model?}
        F2 -->|Yes| F3[Update currentModel]
        F2 -->|No| F4[Return Error]
        F3 --> F5[Process Query with LangChain]
        F5 --> F6[Return Response]
        F4 --> F6
        
        G --> G1[Establish WebSocket]
        G1 --> G2[Listen for Messages]
        G2 --> G3[Parse Message]
        G3 --> G4{Valid Model?}
        G4 -->|Yes| G5[Update currentModel]
        G4 -->|No| G6[Send Error]
        G5 --> G7[Process Query with LangChain]
        G7 --> G8[Send Response via WebSocket]
        G6 --> G8
        
        H --> H1[Check if Ollama is Running]
        H1 --> H2{Ollama Available?}
        H2 -->|Yes| H3[Return Status OK]
        H2 -->|No| H4[Return Status Error]
    end
    
    subgraph "Ollama LLM Processing"
        F5 --> I[Format Prompt with Template]
        G7 --> I
        I --> J[Initialize Ollama LLM Client]
        J --> K[Generate Response with LangChain]
        K --> L[Return Formatted Response]
    end
    
    D --> A
    E1 --> A
    F6 --> A
    G8 --> A
    H3 --> A
    H4 --> A
    L --> F5
    L --> G7
    
    subgraph "Frontend (Browser)"
        A --> M[Display UI]
        M --> N[User Selects Model]
        M --> O[User Enters Message]
        O --> P{Connection Type}
        P -->|HTTP| Q[Send POST Request]
        P -->|WebSocket| R[Send WebSocket Message]
        Q --> S[Display Response]
        R --> S
    end
```

## Explanation of the Flow

1. **Client Interaction**:
    - User accesses the application through a web browser
    - The frontend loads the UI with a model selector and chat interface
    - User can select a model from the dropdown menu and enter messages

2. **Request Handling**:
    - The Go web server handles different types of requests:
        - Serves static files (HTML, CSS, JS)
        - Provides API endpoints for models, chat, and health checks
        - Manages WebSocket connections for real-time chat

3. **Model Selection**:
    - Available models are defined in the `AVAILABLE_MODELS` array
    - The frontend displays these models in a dropdown
    - When a user selects a model, it's sent with the chat message

4. **Chat Processing**:
    - Messages can be sent via HTTP POST or WebSocket
    - The server validates the requested model
    - If valid, it updates the current model
    - The message is processed using the Ollama LLM through LangChain

5. **LLM Integration**:
    - The system formats the prompt using a template
    - Initializes the Ollama LLM client with the selected model
    - Generates a response using LangChain
    - Returns the formatted response to the client

6. **Response Handling**:
    - The response is sent back to the client
    - For HTTP requests, it's returned as JSON
    - For WebSocket connections, it's sent as a text message
    - The frontend displays the response with proper markdown formatting and syntax highlighting

This flow chart illustrates the complete lifecycle of a user interaction with your Ollama AI Assistant application, from model selection to receiving AI-generated responses.