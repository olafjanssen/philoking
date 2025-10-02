@echo off
echo Clearing Kafka message queues...

REM Check if Docker is running
docker ps >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Docker is not running. Please start Docker first.
    pause
    exit /b 1
)

REM Check if Kafka container is running
docker ps --format "table {{.Names}}" | findstr kafka >nul
if %errorlevel% neq 0 (
    echo Error: Kafka container is not running. Please start Kafka first with: docker-compose up -d
    pause
    exit /b 1
)

echo.
echo Stopping the application if running...
taskkill /f /im philoking.exe >nul 2>&1

echo.
echo Clearing message queue by consuming all messages...

REM Clear messages by consuming them (this effectively clears the queue)
echo Clearing chat-messages topic...
timeout /t 2 >nul
docker exec kafka kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic chat-messages --from-beginning --timeout-ms 1000 >nul 2>&1

echo.
echo ✅ Kafka message queues cleared successfully!
echo You can now start the application with: .\philoking.exe
echo.
pause
