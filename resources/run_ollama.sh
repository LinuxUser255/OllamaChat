#!/usr/bin/env bash

# Model urls
# a https://ollama.com/library/gemma3
# b https://ollama.com/library/qwen3
# c https://ollama.com/library/devstral
# d https://ollama.com/library/deepseek-r1
# e https://ollama.com/library/deepseek-coder-v2
# f https://ollama.com/library/llama4
# g https://ollama.com/library/qwen2.5vl
# h https://ollama.com/library/llama3.3
# i https://ollama.com/library/codellama
# j https://ollama.com/library/starcoder2
# k https://ollama.com/library/codegemma
# l https://ollama.com/library/phi4
# m https://ollama.com/library/mistral

echo ''
echo 'Select an Ollama Model to run.'
echo ''

# Function to check if model is pulled, pull if not, and then run
check_pull_and_run() {
    local model=$1

    # Check if model is already downloaded, and if not install it
    if ! ollama list | rg -q "$model"; then
        echo "Model $model not found. Pulling..."
        ollama pull $model
    fi

    # Try to run the model
    echo "Running $model..."
    if ! ollama run $model; then
        # If running fails due to version mismatch, update Ollama
        echo "Updating Ollama..."
        curl -fsSL https://ollama.com/install.sh | sh
        echo "Retrying with $model..."
        ollama run $model
    fi
}

# Function to handle model selection
handle_model() {
    case $1 in
        A) check_pull_and_run gemma3 ;;
        B) check_pull_and_run qwen3 ;;
        C) check_pull_and_run devstral ;;
        D) check_pull_and_run deepseek-r1 ;;
        E) check_pull_and_run deepseek-coder-v2 ;;
        F) check_pull_and_run llama4 ;;
        G) check_pull_and_run qwen2.5vl ;;
        H) check_pull_and_run llama3.3 ;;
        I) check_pull_and_run codellama ;;
        J) check_pull_and_run starcoder2 ;;
        K) check_pull_and_run codegemma ;;
        L) check_pull_and_run phi4 ;;
        M) check_pull_and_run mistral ;;
        *) echo "Invalid selection" ;;
    esac
}

# Display menu
echo "A: gemma3"
echo "B: qwen3"
echo "C: devstral"
echo "D: deepseek-r1"
echo "E: deepseek-coder-v2"
echo "F: llama4"
echo "G: qwen2.5vl"
echo "H: llama3.3"
echo "I: codellama"
echo "J: starcoder2"
echo "K: codegemma"
echo "L: phi4"
echo "M: mistral"
echo ''

# Prompt user for selection
#read -p "Enter your choice (A-M): " choice
read -p "Enter your choice (A/B/C/D/E/F/G/H/I/J/K/L/M): " choice

# Convert to uppercase
choice=$(echo $choice | tr '[:lower:]' '[:upper:]')

# Handle the selected model
handle_model $choice
