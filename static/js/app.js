document.addEventListener('DOMContentLoaded', function() {
    // DOM Elements
    const chatMessages = document.getElementById('chat-messages');
    const chatInput = document.getElementById('chat-input');
    const sendButton = document.getElementById('send-button');
    const modelSelect = document.getElementById('model-select');
    const newChatButton = document.getElementById('new-chat');

    // WebSocket connection
    let socket;
    let currentModel = modelSelect.value;

    // Initialize WebSocket connection
    function initWebSocket() {
        socket = new WebSocket(`ws://${window.location.host}/api/ws`);

        socket.onopen = function(e) {
            console.log("WebSocket connection established");
        };

        socket.onmessage = function(event) {
            displayMessage(event.data, false);
        };

        socket.onclose = function(event) {
            if (event.wasClean) {
                console.log(`Connection closed cleanly, code=${event.code}, reason=${event.reason}`);
            } else {
                console.log('Connection died');
                // Try to reconnect after a delay
                setTimeout(initWebSocket, 3000);
            }
        };

        socket.onerror = function(error) {
            console.error(`WebSocket error: ${error.message}`);
        };
    }

    // Initialize WebSocket on page load
    initWebSocket();

    // Send message function
    function sendMessage() {
        const message = chatInput.value.trim();
        if (message === '') return;

        // Display user message
        displayMessage(message, true);

        // Prepare message data with model selection
        const messageData = JSON.stringify({
            message: message,
            model: currentModel
        });

        // Send to server
        socket.send(messageData);

        // Clear input
        chatInput.value = '';
    }

    // Display message in chat
    function displayMessage(message, isUser) {
        const messageDiv = document.createElement('div');
        messageDiv.className = isUser ? 'message user-message' : 'message bot-message';

        if (!isUser) {
            // Process markdown for bot responses
            try {
                // Configure marked to use highlight.js
                marked.setOptions({
                    highlight: function(code, lang) {
                        if (lang && hljs.getLanguage(lang)) {
                            return hljs.highlight(code, { language: lang }).value;
                        }
                        return hljs.highlightAuto(code).value;
                    }
                });

                messageDiv.innerHTML = marked.parse(message);
            } catch (e) {
                console.error("Markdown parsing error:", e);
                messageDiv.textContent = message;
            }
        } else {
            messageDiv.textContent = message;
        }

        chatMessages.appendChild(messageDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    // Event listeners
    sendButton.addEventListener('click', sendMessage);

    chatInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });

    modelSelect.addEventListener('change', function() {
        currentModel = this.value;
    });

    newChatButton.addEventListener('click', function() {
        // Clear chat history
        while (chatMessages.firstChild) {
            chatMessages.removeChild(chatMessages.firstChild);
        }

        // Add welcome message
        const welcomeDiv = document.createElement('div');
        welcomeDiv.className = 'welcome-message';
        welcomeDiv.innerHTML = `
            <h2>Welcome to Ollama Chat</h2>
            <p>Start a conversation with your selected AI model.</p>
        `;
        chatMessages.appendChild(welcomeDiv);
    });
});