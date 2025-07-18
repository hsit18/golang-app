#!/bin/bash

# Test script for Kafka integration
# Make sure your Kafka server is running and the application is started

echo "Testing Kafka Integration..."

# Set the base URL (adjust port if needed)
BASE_URL="http://localhost:4002"

# Check if server is running
echo "Checking if server is running on ${BASE_URL}..."
if ! curl -s --connect-timeout 5 "${BASE_URL}/api/health" >/dev/null; then
    echo "❌ Server is not responding on ${BASE_URL}"
    echo "Make sure your Go application is running with MUX_HTTP_PORT=4002"
    exit 1
else
    echo "✅ Server is responding"
fi

# Function to safely format JSON output
safe_jq() {
    local response="$1"
    if echo "$response" | jq . >/dev/null 2>&1; then
        echo "$response" | jq .
    else
        echo "Raw response (not JSON): $response"
    fi
}

echo -e "\n1. Testing Health Endpoint..."
response=$(curl -s "${BASE_URL}/api/health")
safe_jq "$response"

echo -e "\n2. Testing Consumer Status..."
response=$(curl -s "${BASE_URL}/api/kafka/consumer/status")
safe_jq "$response"

echo -e "\n3. Testing Synchronous Kafka Message (Text)..."
response=$(curl -s -X POST "${BASE_URL}/api/kafka/send" \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "test-topic",
    "key": "test-key-1",
    "message": "Hello from Kafka producer!"
  }')
safe_jq "$response"

echo -e "\n4. Testing Synchronous Kafka Message (JSON)..."
response=$(curl -s -X POST "${BASE_URL}/api/kafka/send" \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "user-events",
    "key": "user123",
    "message": {
      "user_id": 123,
      "event": "login",
      "timestamp": "2025-07-16T10:00:00Z",
      "ip_address": "192.168.1.1"
    }
  }')
safe_jq "$response"

echo -e "\n5. Testing Asynchronous Kafka Message..."
response=$(curl -s -X POST "${BASE_URL}/api/kafka/send-async" \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "logs",
    "key": "app-log",
    "message": {
      "level": "info",
      "service": "golang-app",
      "message": "Application test message",
      "timestamp": "2025-07-16T10:00:00Z"
    }
  }')
safe_jq "$response"

echo -e "\n6. Testing Error Handling (Missing Topic)..."
response=$(curl -s -X POST "${BASE_URL}/api/kafka/send" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "test-key",
    "message": "This should fail"
  }')
safe_jq "$response"

echo -e "\n\nKafka integration tests completed!"
