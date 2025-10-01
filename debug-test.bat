@echo off
echo Debug Test for PhiloKing
echo ========================

echo 1. Starting Kafka...
docker-compose up -d zookeeper kafka

echo Waiting for Kafka to be ready...
timeout /t 10 /nobreak > nul

echo 2. Building application with debug logging...
go build -o philoking.exe main.go

if %errorlevel% neq 0 (
    echo ❌ Build failed!
    exit /b 1
)

echo ✅ Build successful!

echo 3. Starting application with debug logging...
echo The application will start on http://localhost:8080
echo Watch the console for debug messages
echo Press Ctrl+C to stop the application

REM Set Ollama configuration (but we'll use mock responses for now)
set PROVIDER=ollama
set MODEL=gpt-oss:20b
set OLLAMA_URL=http://localhost:11434

REM Run the application
philoking.exe
