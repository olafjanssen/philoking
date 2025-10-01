#!/bin/bash

echo "Starting PhiloKing Multi-Agent Chat System Test"
echo "=============================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker first."
    exit 1
fi

echo "1. Starting Kafka infrastructure..."
docker-compose up -d zookeeper kafka

echo "Waiting for Kafka to be ready..."
sleep 10

echo "2. Building the application..."
go build -o philoking.exe main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
else
    echo "❌ Build failed!"
    exit 1
fi

echo "3. Starting the application..."
echo "The application will start on http://localhost:8080"
echo "Press Ctrl+C to stop the application"

# Set a dummy API key for testing
export LLM_API_KEY="test-key"

# Run the application
./philoking.exe

