# Dead Letter Queue (DLQ) Implementation

## Overview

The payment gateway implements a robust retry and DLQ mechanism for handling failed settlement processing. When settlement operations fail, the system automatically retries up to 3 times before moving failed messages to a Dead Letter Queue for manual intervention.

## Architecture

### Components

1. **RabbitMQ Bus (`internal/bus/rabbitmq/bus_rmq.go`)**
   - Configures queues with DLQ support via `x-dead-letter-exchange` argument
   - Tracks retry counts using RabbitMQ's `x-death` header
   - Automatically routes failed messages to DLQ after max retries

2. **Settlement Worker (`cmd/settlement-worker/main.go`)**
   - Processes `settlement.requested` events
   - Retries failed settlements up to 3 times
   - Publishes permanently failed settlements to DLQ

3. **DLQ Monitor (`cmd/dlq-monitor/main.go`)**
   - Subscribes to `dlq.settlement.requested` topic
   - Logs permanently failed settlements
   - Can be extended for alerts, manual intervention, etc.

## How It Works

### Message Flow

```
1. API creates transaction
   ↓
2. Publishes to "transaction.created"
   ↓
3. Transaction worker processes → publishes "settlement.requested"
   ↓
4. Settlement worker attempts processing
   ↓
5a. SUCCESS → Mark completed
5b. FAILURE → Retry (Nack with requeue)
   ↓
6. After 3 failed attempts → Send to DLQ
   ↓
7. DLQ monitor logs for manual review
```

### Retry Logic

The retry mechanism is implemented at the RabbitMQ level using native features:

- **Queue Configuration**: Each consumer queue is configured with:
  ```go
  amqp.Table{
      "x-dead-letter-exchange": "dlq.settlement.requested",
      "x-message-ttl":          300000, // 5 minutes TTL
  }
  ```

- **Retry Tracking**: RabbitMQ automatically tracks redeliveries in the `x-death` header:
  ```go
  if xDeath, ok := m.Headers["x-death"].([]interface{}); ok {
      if death, ok := xDeath[0].(amqp.Table); ok {
          if count, ok := death["count"].(int64); ok {
              retryCount = int(count)
          }
      }
  }
  ```

- **Max Retries**: After 3 failed attempts, the message is rejected without requeue:
  ```go
  if retryCount >= 3 {
      m.Nack(false, false) // Don't requeue → goes to DLQ
  } else {
      m.Nack(false, true)  // Requeue for retry
  }
  ```

### DLQ Configuration

Each topic exchange has a corresponding DLQ:

- **Main Exchange**: `settlement.requested`
- **DLQ Exchange**: `dlq.settlement.requested`
- **DLQ Queue**: `dlq.settlement.requested` (same name as exchange)

The DLQ infrastructure is automatically created when the first consumer subscribes.

## Testing the DLQ Flow

### Manual Test

1. **Start the services**:
   ```bash
   # Terminal 1: Start API
   go run ./cmd/api
   
   # Terminal 2: Start transaction worker
   go run ./cmd/transaction-worker
   
   # Terminal 3: Start settlement worker
   go run ./cmd/settlement-worker
   
   # Terminal 4: Start DLQ monitor
   go run ./cmd/dlq-monitor
   ```

2. **Force settlement failures**:
   Edit `internal/service/settlement_service.go`:
   ```go
   func (s *SettlementService) simulateSettlementProcessing(...) bool {
       return false // Force failure
   }
   ```

3. **Create a transaction**:
   ```bash
   curl -X POST http://localhost:8080/transactions \
     -H "Content-Type: application/json" \
     -d '{
       "merchant_id": "...",
       "amount": 1000,
       "currency": "USD",
       "payment_method": "card"
     }'
   ```

4. **Observe the logs**:
   - Settlement worker will show 3 retry attempts
   - After 3rd failure, message goes to DLQ
   - DLQ monitor logs the permanently failed message

### Automated Test

Run the test script:
```bash
./test-dlq.sh
```

This script will:
- Temporarily modify the code to force failures
- Start the workers and DLQ monitor
- Show retry attempts and DLQ routing
- Clean up and restore code on exit

## Monitoring and Alerts

### Current Implementation

The DLQ monitor logs failed messages:
```go
logger.Error("DLQ message received - settlement failed permanently",
    zap.String("topic", topic),
    zap.String("key", key),
    zap.ByteString("payload", payload),
)
```

### Production Enhancements

For production, extend the DLQ monitor to:

1. **Store in Database**: Persist failed messages for analysis
   ```go
   dlqRepo.Create(ctx, &DLQMessage{
       Topic:     topic,
       Payload:   payload,
       FailedAt:  time.Now(),
   })
   ```

2. **Send Alerts**: Notify operations team
   ```go
   alertService.SendAlert("Settlement failed permanently", payload)
   ```

3. **Trigger Workflows**: Initiate manual review process
   ```go
   workflowService.CreateManualReviewTask(payload)
   ```

4. **Metrics**: Track DLQ message rates
   ```go
   metrics.Counter("dlq.messages.total").Inc()
   ```

## Configuration

### Environment Variables

- `RABBITMQ_URL`: RabbitMQ connection string (default: `amqp://guest:guest@localhost:5672/`)
- `LOG_SERVICE_NAME`: Service name for logging

### RabbitMQ Queue Settings

- **Message TTL**: 5 minutes (300,000 ms)
- **Max Retries**: 3 attempts
- **DLQ Durability**: DLQ queues are durable to prevent message loss

## Troubleshooting

### Messages Not Going to DLQ

**Symptom**: Messages keep retrying indefinitely

**Causes**:
- Queue created before DLQ configuration was added
- Exchange type mismatch

**Solution**:
```bash
# Delete existing queues/exchanges
docker exec rabbitmq rabbitmqctl purge_queue settlement.requested
docker exec rabbitmq rabbitmqctl delete_exchange settlement.requested

# Restart workers to recreate with correct config
```

### DLQ Messages Not Being Consumed

**Symptom**: Messages in DLQ but monitor not logging

**Causes**:
- DLQ monitor not running
- Incorrect topic name

**Solution**:
```bash
# Check RabbitMQ management UI
open http://localhost:15672

# Verify queue bindings
# Restart DLQ monitor
```

### High DLQ Rate

**Symptom**: Many messages going to DLQ

**Investigation**:
1. Check settlement service logs for error patterns
2. Review external API availability
3. Check database connectivity
4. Monitor network latency

## Best Practices

1. **Always Run DLQ Monitor**: Ensure the DLQ monitor is running in production to catch failures
2. **Set Up Alerts**: Configure alerts when DLQ receives messages
3. **Regular Review**: Periodically review DLQ messages for patterns
4. **Manual Intervention**: Create processes for handling DLQ messages
5. **Metrics**: Track retry rates and DLQ message counts
6. **Testing**: Regularly test the DLQ flow to ensure it works correctly

## Future Enhancements

1. **Configurable Retry Counts**: Make max retries configurable per topic
2. **Exponential Backoff**: Increase delay between retries
3. **Retry with Backoff**: Use TTL to implement delayed retries
4. **DLQ Reprocessing**: Add ability to replay DLQ messages after fixing issues
5. **Dead Letter Routing**: Route different failure types to different DLQs
6. **Monitoring Dashboard**: Web UI for viewing and managing DLQ messages
