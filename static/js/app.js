document.addEventListener('DOMContentLoaded', function() {
    const chatMessages = document.getElementById('chat-messages');
    const userInput = document.getElementById('user-input');
    const sendButton = document.getElementById('send-button');

    // Backend API URL - make sure this matches your backend server
    const API_URL = 'http://127.0.0.1:8000/api/chat';

    // Configure marked.js with better code highlighting options
    marked.setOptions({
        highlight: function(code, lang) {
            if (lang && hljs.getLanguage(lang)) {
                return hljs.highlight(code, { language: lang }).value;
            }
            return hljs.highlightAuto(code).value;
        },
        breaks: true,
        gfm: true,
        pedantic: false,
        sanitize: false,
        smartLists: true,
        smartypants: false
    });

    // Override the renderer to add line numbers to code blocks
    const renderer = new marked.Renderer();
    const originalCodeRenderer = renderer.code;
    renderer.code = function(code, language, isEscaped) {
        return originalCodeRenderer.call(this, code, language, isEscaped);
    };

    marked.use({ renderer });

    // Function to add a message to the chat
    function addMessage(content, isUser = false) {
        const messageRow = document.createElement('div');
        messageRow.className = `message-row ${isUser ? 'user' : 'bot'}`;

        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${isUser ? 'user' : 'bot'}`;

        const messageContainer = document.createElement('div');
        messageContainer.className = 'message-container';

        // Create avatar
        const avatar = document.createElement('div');
        avatar.className = isUser ? 'user-avatar' : 'bot-avatar';
        avatar.textContent = isUser ? 'U' : 'AI';

        const contentDiv = document.createElement('div');
        contentDiv.className = 'message-content markdown-body';

        // Use markdown parsing for both user and bot messages
        contentDiv.innerHTML = marked.parse(content);

        // Apply syntax highlighting to code blocks
        contentDiv.querySelectorAll('pre code').forEach((block) => {
            hljs.highlightElement(block);

            // Add copy button to each code block (for both user and bot messages)
            const preElement = block.parentElement;
            const codeBlockWrapper = document.createElement('div');
            codeBlockWrapper.className = 'code-block-wrapper';
            codeBlockWrapper.style.position = 'relative';

            // Get the language from the class
            const languageClass = Array.from(block.classList).find(cls => cls.startsWith('language-'));
            if (languageClass) {
                const language = languageClass.replace('language-', '');
                codeBlockWrapper.setAttribute('data-language', language);
            }

            // Create copy button for this code block
            const copyCodeButton = document.createElement('button');
            copyCodeButton.className = 'copy-button code-copy-button';
            copyCodeButton.textContent = 'Copy';
            copyCodeButton.title = 'Copy code to clipboard';

            // Add click event to copy just this code block
            copyCodeButton.addEventListener('click', function(e) {
                e.stopPropagation(); // Prevent event bubbling

                // Get the text content of just this code block
                const codeText = block.textContent;

                // Copy to clipboard
                navigator.clipboard.writeText(codeText).then(function() {
                    copyCodeButton.textContent = 'Copied!';
                    copyCodeButton.classList.add('copy-success');

                    setTimeout(function() {
                        copyCodeButton.textContent = 'Copy';
                        copyCodeButton.classList.remove('copy-success');
                    }, 2000);
                }).catch(function(err) {
                    console.error('Could not copy code: ', err);
                    copyCodeButton.textContent = 'Error!';

                    setTimeout(function() {
                        copyCodeButton.textContent = 'Copy';
                    }, 2000);
                });
            });

            // Replace the pre element with our wrapper
            preElement.parentNode.insertBefore(codeBlockWrapper, preElement);
            codeBlockWrapper.appendChild(preElement);
            codeBlockWrapper.appendChild(copyCodeButton);
        });

        messageContainer.appendChild(avatar);
        messageContainer.appendChild(contentDiv);
        messageDiv.appendChild(messageContainer);

        // Only add the "Copy All" button to bot messages
        if (!isUser) {
            // Add copy button for the entire bot message
            const copyButton = document.createElement('button');
            copyButton.className = 'copy-button';
            copyButton.textContent = 'Copy All';
            copyButton.title = 'Copy entire message to clipboard';
            copyButton.addEventListener('click', function() {
                // Copy the original markdown content to clipboard
                navigator.clipboard.writeText(content).then(function() {
                    // Visual feedback for successful copy
                    copyButton.textContent = 'Copied!';
                    copyButton.classList.add('copy-success');

                    // Reset button text after 2 seconds
                    setTimeout(function() {
                        copyButton.textContent = 'Copy All';
                        copyButton.classList.remove('copy-success');
                    }, 2000);
                }).catch(function(err) {
                    console.error('Could not copy text: ', err);
                    copyButton.textContent = 'Error!';

                    // Reset button text after 2 seconds
                    setTimeout(function() {
                        copyButton.textContent = 'Copy All';
                    }, 2000);
                });
            });

            messageDiv.appendChild(copyButton);
        }

        messageRow.appendChild(messageDiv);
        chatMessages.appendChild(messageRow);

        // Scroll to the bottom of the chat
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    // Function to show typing indicator
    function showTypingIndicator() {
        const indicator = document.createElement('div');
        indicator.className = 'typing-indicator';
        indicator.id = 'typing-indicator';

        for (let i = 0; i < 3; i++) {
            const dot = document.createElement('span');
            indicator.appendChild(dot);
        }

        chatMessages.appendChild(indicator);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    // Function to remove typing indicator
    function removeTypingIndicator() {
        const indicator = document.getElementById('typing-indicator');
        if (indicator) {
            indicator.remove();
        }
    }

    // Function to send a message to the backend
    async function sendMessage(message) {
        try {
            showTypingIndicator();

            const response = await fetch(API_URL, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ message: message }),
            });

            if (!response.ok) {
                removeTypingIndicator();
                addMessage(`Error: Server returned status ${response.status}`, false);
                return;
            }

            const data = await response.json();
            removeTypingIndicator();

            // Add the bot's response to the chat
            addMessage(data.response, false);
        } catch (error) {
            console.error('Error sending message:', error);
            removeTypingIndicator();
            addMessage('Sorry, there was an error processing your request. Please try again.', false);
        }
    }

    // Event listener for send button
    sendButton.addEventListener('click', function() {
        const message = userInput.value.trim();
        if (message) {
            // Add user message to chat
            addMessage(message, true);

            // Clear input field
            userInput.value = '';

            // Send message to backend
            sendMessage(message);
        }
    });

    // Event listener for Enter key in input field
    userInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendButton.click();
        }
    });

    const previewToggle = document.getElementById('preview-toggle');
    const markdownPreview = document.getElementById('markdown-preview');
    let previewActive = false;

    // Toggle markdown preview
    if (previewToggle && markdownPreview) {
        previewToggle.addEventListener('click', function() {
            previewActive = !previewActive;

            if (previewActive) {
                // Update preview content
                markdownPreview.innerHTML = marked.parse(userInput.value);

                // Apply syntax highlighting
                markdownPreview.querySelectorAll('pre code').forEach((block) => {
                    hljs.highlightElement(block);
                });

                // Show preview
                markdownPreview.classList.add('active');
                userInput.style.opacity = '0';
            } else {
                // Hide preview
                markdownPreview.classList.remove('active');
                userInput.style.opacity = '1';
            }
        });

        // Update preview when typing
        userInput.addEventListener('input', function() {
            if (previewActive) {
                markdownPreview.innerHTML = marked.parse(userInput.value);

                // Apply syntax highlighting
                markdownPreview.querySelectorAll('pre code').forEach((block) => {
                    hljs.highlightElement(block);
                });
            }
        });
    }

    // Markdown toolbar functionality
    const toolbarButtons = document.querySelectorAll('.toolbar-button');

    toolbarButtons.forEach(button => {
        button.addEventListener('click', function() {
            const format = this.getAttribute('data-format');
            const textarea = document.getElementById('user-input');
            const start = textarea.selectionStart;
            const end = textarea.selectionEnd;
            const selectedText = textarea.value.substring(start, end);
            let replacement = '';

            switch(format) {
                case 'bold':
                    replacement = `**${selectedText}**`;
                    break;
                case 'italic':
                    replacement = `*${selectedText}*`;
                    break;
                case 'code':
                    replacement = `\`${selectedText}\``;
                    break;
                case 'codeblock':
                    replacement = '```\n' + selectedText + '\n```';
                    break;
            }

            textarea.value = textarea.value.substring(0, start) + replacement + textarea.value.substring(end);
            textarea.focus();
            textarea.selectionStart = start + replacement.length;
            textarea.selectionEnd = start + replacement.length;
        });
    });

    // Add copy buttons to existing bot messages (if any)
    document.querySelectorAll('.message.bot').forEach(function(botMessage) {
        // Extract the text content from the message
        const messageContent = botMessage.querySelector('.message-content').textContent;

        // Create and add the copy button for the entire message
        const copyButton = document.createElement('button');
        copyButton.className = 'copy-button';
        copyButton.textContent = 'Copy All';
        copyButton.title = 'Copy entire message to clipboard';
        copyButton.addEventListener('click', function() {
            navigator.clipboard.writeText(messageContent).then(function() {
                copyButton.textContent = 'Copied!';
                copyButton.classList.add('copy-success');

                setTimeout(function() {
                    copyButton.textContent = 'Copy All';
                    copyButton.classList.remove('copy-success');
                }, 2000);
            }).catch(function(err) {
                console.error('Could not copy text: ', err);
                copyButton.textContent = 'Error!';

                setTimeout(function() {
                    copyButton.textContent = 'Copy All';
                }, 2000);
            });
        });

        botMessage.appendChild(copyButton);

        // Add copy buttons to each code block in existing messages
        botMessage.querySelectorAll('pre code').forEach((block) => {
            const preElement = block.parentElement;
            const codeBlockWrapper = document.createElement('div');
            codeBlockWrapper.className = 'code-block-wrapper';
            codeBlockWrapper.style.position = 'relative';

            // Create copy button for this code block
            const copyCodeButton = document.createElement('button');
            copyCodeButton.className = 'copy-button code-copy-button';
            copyCodeButton.textContent = 'Copy';
            copyCodeButton.title = 'Copy code to clipboard';

            // Add click event to copy just this code block
            copyCodeButton.addEventListener('click', function(e) {
                e.stopPropagation(); // Prevent event bubbling

                // Get the text content of just this code block
                const codeText = block.textContent;

                // Copy to clipboard
                navigator.clipboard.writeText(codeText).then(function() {
                    copyCodeButton.textContent = 'Copied!';
                    copyCodeButton.classList.add('copy-success');

                    setTimeout(function() {
                        copyCodeButton.textContent = 'Copy';
                        copyCodeButton.classList.remove('copy-success');
                    }, 2000);
                }).catch(function(err) {
                    console.error('Could not copy code: ', err);
                    copyCodeButton.textContent = 'Error!';

                    setTimeout(function() {
                        copyCodeButton.textContent = 'Copy';
                    }, 2000);
                });
            });

            // Replace the pre element with our wrapper
            preElement.parentNode.insertBefore(codeBlockWrapper, preElement);
            codeBlockWrapper.appendChild(preElement);
            codeBlockWrapper.appendChild(copyCodeButton);
        });
    });
});