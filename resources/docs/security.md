# Security Recommendations for Ollama Chat Application

To strengthen your application against common security vulnerabilities, here are specific recommendations for your Ollama Chat application:

## API Security

### 1. Input Validation and Sanitization

**Current Risk**: Your application accepts user input and passes it directly to the Ollama CLI, which could allow command injection.

**Recommendation**:
```go
// Add to main.go
func validateInput(input string) bool {
    // Check for maximum length
    if len(input) > 4000 {
        return false
    }
    
    // Check for potentially dangerous patterns
    dangerousPatterns := []string{";", "&&", "||", "`", "$", "|"}
    for _, pattern := range dangerousPatterns {
        if strings.Contains(input, pattern) {
            return false
        }
    }
    
    return true
}

// Then in processOllamaQuery:
func processOllamaQuery(query string, model string) string {
    if !validateInput(query) {
        return "Invalid input detected"
    }
    
    // Validate model name against whitelist
    validModels := map[string]bool{
        "llama3.1": true, "deepseek-coder-v2": true, "gemma3": true,
        "mistral": true, "phi4": true, "llama4": true, "qwen3": true,
    }
    
    if !validModels[model] {
        model = "llama3.1" // Default to safe model
    }
    
    // Continue with processing...
}
```

### 2. Rate Limiting

**Current Risk**: Without rate limiting, your API could be vulnerable to DoS attacks or excessive resource consumption.

**Recommendation**:
```go
// Add to main.go
type IPRateLimiter struct {
    ips map[string]time.Time
    mu  sync.Mutex
}

func NewIPRateLimiter() *IPRateLimiter {
    return &IPRateLimiter{
        ips: make(map[string]time.Time),
    }
}

func (r *IPRateLimiter) Allow(ip string) bool {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    now := time.Now()
    lastAttempt, found := r.ips[ip]
    
    if !found || now.Sub(lastAttempt) > 1*time.Second {
        r.ips[ip] = now
        return true
    }
    
    return false
}

// Create a global rate limiter
var rateLimiter = NewIPRateLimiter()

// Then in handleWebSocket:
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Get client IP
    ip := r.RemoteAddr
    
    // Apply rate limiting
    if !rateLimiter.Allow(ip) {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }
    
    // Continue with WebSocket handling...
}
```

### 3. CORS Protection

**Current Risk**: Your WebSocket allows connections from any origin, which could enable cross-site attacks.

**Recommendation**:
```go
// Replace the current upgrader with:
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // In production, restrict to your domain
        origin := r.Header.Get("Origin")
        return origin == "http://yourdomain.com" || 
               origin == "https://yourdomain.com" ||
               strings.HasPrefix(origin, "http://localhost")
    },
}
```

## Front-End Security

### 1. Cross-Site Scripting (XSS) Protection

**Current Risk**: Your application renders markdown from the LLM, which could contain malicious scripts.

**Recommendation**:

```javascript
// In app.js, modify the displayMessage function:
function displayMessage(message, isUser) {
    const messageDiv = document.createElement('div');
    messageDiv.className = isUser ? 'message user-message' : 'message bot-message';
    
    if (!isUser) {
        try {
            // Configure DOMPurify and marked for safe rendering
            marked.setOptions({
                highlight: function(code, lang) {
                    if (lang && hljs.getLanguage(lang)) {
                        return hljs.highlight(code, { language: lang }).value;
                    }
                    return hljs.highlightAuto(code).value;
                },
                sanitize: true, // Enable built-in sanitizer
            });
            
            // Use DOMPurify to sanitize the HTML
            const sanitizedHTML = DOMPurify.sanitize(marked.parse(message));
            messageDiv.innerHTML = sanitizedHTML;
        } catch (e) {
            console.error("Markdown parsing error:", e);
            messageDiv.textContent = message;
        }
    } else {
        // For user messages, use textContent to prevent XSS
        messageDiv.textContent = message;
    }
    
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}
```

Add DOMPurify to your HTML:
```html
<script src="https://cdnjs.cloudflare.com/ajax/libs/dompurify/2.4.0/purify.min.js"></script>
```

### 2. Content Security Policy (CSP)

**Current Risk**: Without CSP, your application is vulnerable to script injection attacks.

**Recommendation**:

Add this to your HTML `<head>` section:
```html
<meta http-equiv="Content-Security-Policy" content="default-src 'self'; script-src 'self' https://cdnjs.cloudflare.com; style-src 'self' https://cdnjs.cloudflare.com; connect-src 'self' ws: wss:;">
```

## Server Security

### 1. HTTP Security Headers

**Current Risk**: Missing security headers can expose your application to various attacks.

**Recommendation**:
```go
// Add middleware to set security headers
func securityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevent MIME type sniffing
        w.Header().Set("X-Content-Type-Options", "nosniff")
        
        // Prevent clickjacking
        w.Header().Set("X-Frame-Options", "DENY")
        
        // Enable XSS protection in browsers
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        
        // Strict Transport Security (when using HTTPS)
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        
        next.ServeHTTP(w, r)
    })
}

// Apply middleware in main():
r.Use(securityHeadersMiddleware)
```

### 2. Command Execution Protection

**Current Risk**: Direct execution of system commands with user input is dangerous.

**Recommendation**:
```go
// Replace direct command execution with a safer approach
func processOllamaQuery(query string, model string) string {
    // Validate inputs first (as shown earlier)
    
    // Use explicit arguments instead of shell string
    cmd := exec.Command("ollama", "run", model)
    
    // Create pipes for stdin/stdout
    stdin, err := cmd.StdinPipe()
    if err != nil {
        return fmt.Sprintf("Error: %v", err)
    }
    
    // Write query to stdin
    go func() {
        defer stdin.Close()
        io.WriteString(stdin, query)
    }()
    
    // Get output
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Sprintf("Error: %v", err)
    }
    
    return strings.TrimSpace(string(output))
}
```

## Additional Security Measures

### 1. Implement Request Timeouts

**Current Risk**: Long-running requests could lead to resource exhaustion.

**Recommendation**:
```go
// In main():
srv := &http.Server{
    Handler:      r,
    Addr:         ":" + PORT,
    WriteTimeout: 30 * time.Second,
    ReadTimeout:  15 * time.Second,
    IdleTimeout:  60 * time.Second,
}
log.Fatal(srv.ListenAndServe())
```

### 2. Secure WebSocket Messages

**Current Risk**: WebSocket messages are transmitted in plaintext.

**Recommendation**:
```javascript
// In app.js, add message encryption/decryption
// This is a simple example - consider using a proper encryption library
function encryptMessage(message) {
    // Simple encoding for demonstration - use proper encryption in production
    return btoa(message);
}

function decryptMessage(message) {
    // Simple decoding for demonstration
    return atob(message);
}

// Then when sending:
socket.send(encryptMessage(messageData));

// And when receiving:
socket.onmessage = function(event) {
    const decryptedMessage = decryptMessage(event.data);
    displayMessage(decryptedMessage, false);
};
```

### 3. Input Length Restrictions

**Current Risk**: Extremely long inputs could cause performance issues or buffer overflows.

**Recommendation**:
```javascript
// In app.js
chatInput.addEventListener('input', function() {
    const maxLength = 4000;
    if (this.value.length > maxLength) {
        this.value = this.value.substring(0, maxLength);
        // Notify user
        alert(`Message too long. Limited to ${maxLength} characters.`);
    }
});
```

## Security Testing Recommendations

1. **Regular Dependency Scanning**: Use tools like `go list -m all -json | nancy sleuth` for Go dependencies and `npm audit` for JavaScript dependencies.

2. **Static Code Analysis**: Implement tools like `gosec` for Go code and ESLint with security plugins for JavaScript.

3. **Penetration Testing**: Regularly test your application with tools like OWASP ZAP or Burp Suite.

4. **Security Headers Check**: Use online tools like SecurityHeaders.com to verify your HTTP security headers.

By implementing these security measures, you'll significantly reduce the risk of common vulnerabilities in your Ollama Chat application while maintaining its functionality and performance.