import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from langchain.prompts import PromptTemplate
# Updated import for Ollama in LangChain
from langchain_ollama import OllamaLLM
from pydantic import BaseModel
from typing import List, Optional

# Initialize the Ollama model with updated class
model = OllamaLLM(model="deepseek-coder-v2")

# Define available models
AVAILABLE_MODELS = [
    "deepseek-coder-v2",
    "codellama:7b",
    "codellama:13b",
    "llama3:8b",
    "mistral:7b"
]

# Define a system prompt that encourages proper code formatting
SYSTEM_TEMPLATE = """You are a helpful coding assistant. When providing code examples:
1. Always use proper markdown formatting with language-specific syntax highlighting
2. Use triple backticks with the language name for code blocks (e.g. ```python)
3. Format code in a clean, readable way with proper indentation
4. Use VSCode-style syntax highlighting conventions

User Query: {query}
"""

prompt_template = PromptTemplate(
    input_variables=["query"],
    template=SYSTEM_TEMPLATE
)

# Define request model
class ChatMessage(BaseModel):
    message: str
    model_name: Optional[str] = "deepseek-coder-v2"

# Define model info response
class ModelInfo(BaseModel):
    available_models: List[str]
    current_model: str

# Create FastAPI app
app = FastAPI(title="Ollama Chat Bot API")

# Configure CORS
app.add_middleware(
    CORSMiddleware,  # type: ignore
    allow_origins=["*"],  # In production, replace with specific origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/")
async def root():
    return {"message": "Ollama Chat Bot API is running"}

@app.get("/api/models")
async def get_models():
    """Get available models and current model"""
    return ModelInfo(
        available_models=AVAILABLE_MODELS,
        current_model=model.model
    )

@app.post("/api/chat")
async def chat(chat_message: ChatMessage):
    # Process the message here
    user_message = chat_message.message

    # Check if we need to switch models
    global model
    if chat_message.model_name and chat_message.model_name != model.model:
        if chat_message.model_name in AVAILABLE_MODELS:
            try:
                model = OllamaLLM(model=chat_message.model_name)
            except Exception as e:
                return {"response": f"Error switching to model {chat_message.model_name}: {str(e)}"}

    try:
        # Format the prompt with the system instructions
        formatted_prompt = prompt_template.format(query=user_message)

        # Call the Ollama model with the formatted prompt
        llm_response = model.invoke(formatted_prompt)
        return {"response": llm_response}
    except Exception as e:
        print(f"Error calling Ollama: {str(e)}")
        return {"response": f"Error processing your request: {str(e)}"}


# Add this to make the script runnable
if __name__ == "__main__":
    uvicorn.run("app:app", host="127.0.0.1", port=8000, reload=True)
