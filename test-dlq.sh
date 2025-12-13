#!/bin/bash

# Test DLQ flow by forcing settlement failures

echo "Starting DLQ flow test..."
echo "This script will:"
echo "1. Force settlement service to fail"
echo "2. Watch retries happen (3 attempts)"
echo "3. Verify message goes to DLQ"
echo ""

# Temporarily modify settlement_service.go to force failures
echo "Step 1: Forcing settlement failures by modifying code..."
sed -i 's/return true \/\/ Change to false to test DLQ flow/return false \/\/ Forcing failure for DLQ test/' internal/service/settlement_service.go

echo "Step 2: Rebuild settlement-worker..."
go build -o bin/settlement-worker ./cmd/settlement-worker

echo "Step 3: Start settlement-worker in background..."
./bin/settlement-worker &
WORKER_PID=$!

echo "Step 4: Start DLQ monitor in background..."
go run ./cmd/dlq-monitor &
DLQ_PID=$!

echo ""
echo "Settlement worker PID: $WORKER_PID"
echo "DLQ monitor PID: $DLQ_PID"
echo ""
echo "Step 5: Trigger a transaction via API (this will trigger settlement)..."
echo "Watch the logs for:"
echo "  - Settlement attempts (will fail)"
echo "  - Retry attempts (3 times)"
echo "  - Message moved to DLQ"
echo ""
echo "Press Ctrl+C to stop and cleanup..."

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    kill $WORKER_PID 2>/dev/null
    kill $DLQ_PID 2>/dev/null
    
    # Restore original code
    sed -i 's/return false \/\/ Forcing failure for DLQ test/return true \/\/ Change to false to test DLQ flow/' internal/service/settlement_service.go
    
    echo "Cleanup complete. Original code restored."
    exit 0
}

trap cleanup SIGINT SIGTERM

# Wait for interrupt
wait
