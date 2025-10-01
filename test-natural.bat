@echo off
echo Starting PhiloKing Natural Conversation System
echo =============================================

echo 1. Starting Kafka...
docker-compose up -d zookeeper kafka

echo Waiting for Kafka to be ready...
timeout /t 10 /nobreak > nul

echo 2. Building natural conversation system...
go build -o philoking-natural.exe main-natural.go

if %errorlevel% neq 0 (
    echo ❌ Build failed!
    exit /b 1
)

echo ✅ Build successful!

echo 3. Starting natural conversation system...
echo.
echo 🤖 This system features:
echo    - 4 different personality agents
echo    - Selective responses based on relevance
echo    - Natural conversation flow
echo    - User treated as equal participant
echo.
echo The application will start on http://localhost:8080
echo Press Ctrl+C to stop the application

REM Set Ollama configuration (but we'll use mock responses for now)
set PROVIDER=ollama
set MODEL=gpt-oss:20b
set OLLAMA_URL=http://localhost:11434

REM Run the natural conversation system
philoking-natural.exe
