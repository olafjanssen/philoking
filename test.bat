@echo off
echo Starting PhiloKing Multi-Agent Chat System Test
echo ==============================================

echo 1. Starting Kafka infrastructure...
docker-compose up -d zookeeper kafka

echo Waiting for Kafka to be ready...
timeout /t 10 /nobreak > nul

echo 2. Building the application...
go build -o philoking.exe main.go

if %errorlevel% equ 0 (
    echo ✅ Build successful!
) else (
    echo ❌ Build failed!
    exit /b 1
)

echo 3. Starting the application...
echo The application will start on http://localhost:8080
echo Press Ctrl+C to stop the application

REM Set a dummy API key for testing
set LLM_API_KEY=test-key

REM Run the application
philoking.exe

