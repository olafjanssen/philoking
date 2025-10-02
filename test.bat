@echo off
echo Starting PhiloKing Multi-Agent Conversation System
echo ==================================================

echo 1. Starting Kafka...
docker-compose up -d zookeeper kafka

echo Waiting for Kafka to be ready...
timeout /t 10 /nobreak > nul

echo 2. Building application...
go build -o philoking.exe main.go

if %errorlevel% neq 0 (
    echo ❌ Build failed!
    exit /b 1
)

echo ✅ Build successful!

echo 3. Starting multi-agent conversation system...
echo.
echo 🤖 This system features:
echo    - Multiple agent types (LLM, Echo, etc.)
echo    - Configurable agent types via YAML
echo    - Natural conversation flow with selective responses
echo    - Web interface for real-time chat
echo.
echo The application will start on http://localhost:8080
echo Configure agents in config.yaml
echo To clear Kafka messages, run: clear-messages.bat
echo Press Ctrl+C to stop the application

REM Set Ollama configuration
set PROVIDER=ollama
set MODEL=gpt-oss:20b
set OLLAMA_URL=http://localhost:11434

REM Run the application
philoking.exe
