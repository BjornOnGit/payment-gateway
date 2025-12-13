# DLQ Architecture Diagram

## Complete Message Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Payment Gateway System                          │
└─────────────────────────────────────────────────────────────────────────┘

┌──────────┐
│   API    │  Creates transaction
│  Server  │─────────────────┐
└──────────┘                 │
                             ▼
                    ┌─────────────────┐
                    │   RabbitMQ      │
                    │   Exchange:     │
                    │ transaction.    │
                    │   created       │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │ Transaction     │  Processes transaction
                    │   Worker        │  Updates status
                    └────────┬────────┘
                             │
                             │ Publishes settlement.requested
                             ▼
                    ┌─────────────────┐
                    │   RabbitMQ      │
                    │   Exchange:     │
                    │  settlement.    │
                    │   requested     │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  Main Queue     │
                    │  (auto-named)   │
                    │                 │
                    │ Config:         │
                    │ ┌─────────────┐ │
                    │ │x-dead-letter│ │
                    │ │-exchange:   │ │
                    │ │dlq.settle...│ │
                    │ │x-message-ttl│ │
                    │ │300000 (5min)│ │
                    │ └─────────────┘ │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  Settlement     │
                    │    Worker       │
                    └────────┬────────┘
                             │
                  ┌──────────┴──────────┐
                  │                     │
            SUCCESS │                   │ FAILURE
                  │                     │
                  ▼                     ▼
         ┌──────────────┐     ┌──────────────────┐
         │   ACK        │     │  NACK(requeue)   │
         │   Message    │     │  Retry Count++   │
         │   removed    │     └────────┬─────────┘
         └──────────────┘              │
                                       │
                            ┌──────────┴─────────┐
                            │                    │
                     Retry 1,2 │              Retry 3 │
                            │                    │
                            ▼                    ▼
                    ┌───────────────┐   ┌──────────────────┐
                    │  Re-delivered │   │ NACK(no requeue) │
                    │  to worker    │   │   Max retries!   │
                    └───────┬───────┘   └────────┬─────────┘
                            │                    │
                            └────────────────────┘
                                       │
                                       │ x-dead-letter-exchange routes
                                       │
                                       ▼
                            ┌─────────────────────┐
                            │   RabbitMQ          │
                            │   DLQ Exchange:     │
                            │ dlq.settlement.     │
                            │    requested        │
                            └──────────┬──────────┘
                                       │
                                       ▼
                            ┌─────────────────────┐
                            │   DLQ Queue         │
                            │ (durable, named)    │
                            └──────────┬──────────┘
                                       │
                                       ▼
                            ┌─────────────────────┐
                            │   DLQ Monitor       │
                            │   - Logs failure    │
                            │   - Sends alerts    │
                            │   - Manual review   │
                            └─────────────────────┘
```

## Retry Flow Detail

```
Attempt 1: Worker receives message (retry_count=0)
           │
           ├─ SUCCESS → ACK → Done ✓
           │
           └─ FAILURE → NACK(requeue=true)
                        │
                        ▼
Attempt 2: Worker receives message (retry_count=1)
           │
           ├─ SUCCESS → ACK → Done ✓
           │
           └─ FAILURE → NACK(requeue=true)
                        │
                        ▼
Attempt 3: Worker receives message (retry_count=2)
           │
           ├─ SUCCESS → ACK → Done ✓
           │
           └─ FAILURE → NACK(requeue=true)
                        │
                        ▼
Attempt 4: Worker receives message (retry_count=3)
           │
           └─ Check: retry_count >= maxRetries (3)
              │
              └─ NACK(requeue=false)
                 │
                 ▼
              RabbitMQ routes to DLQ via x-dead-letter-exchange
              │
              ▼
           DLQ Monitor logs permanent failure
```

## RabbitMQ Header Tracking

```
Message Headers Evolution:

Initial Delivery:
{
  "x-death": null  // No x-death header yet
}

After 1st Nack:
{
  "x-death": [{
    "count": 1,
    "reason": "rejected",
    "queue": "amq.gen-xyz",
    "time": "2024-01-15T10:30:00Z",
    "exchange": "settlement.requested"
  }]
}

After 2nd Nack:
{
  "x-death": [{
    "count": 2,     ← Incremented!
    "reason": "rejected",
    "queue": "amq.gen-xyz",
    "time": "2024-01-15T10:31:00Z",
    "exchange": "settlement.requested"
  }]
}

After 3rd Nack (final):
{
  "x-death": [{
    "count": 3,     ← Max retries reached
    "reason": "rejected",
    "queue": "amq.gen-xyz",
    "time": "2024-01-15T10:32:00Z",
    "exchange": "settlement.requested"
  }]
}
→ NACK(requeue=false) → Routes to DLQ
```

## Code Flow

```go
// In bus_rmq.go Subscribe method:

for m := range msgs {
    // Extract retry count from x-death header
    retryCount := 0
    if xDeath, ok := m.Headers["x-death"].([]interface{}); ok {
        if death, ok := xDeath[0].(amqp.Table); ok {
            if count, ok := death["count"].(int64); ok {
                retryCount = int(count)
            }
        }
    }

    // Store in context for handler
    msgCtx = context.WithValue(msgCtx, "retry_count", retryCount)

    // Call handler
    if err := handler(msgCtx, topic, m.RoutingKey, m.Body); err != nil {
        if retryCount >= 3 {
            m.Nack(false, false)  // → DLQ
        } else {
            m.Nack(false, true)   // → Retry
        }
        continue
    }

    m.Ack(false)  // Success
}
```

```go
// In settlement_service.go ProcessSettlement method:

func (s *SettlementService) ProcessSettlement(ctx context.Context, payload []byte) error {
    // Extract retry count from context
    retryCount := 0
    if rc, ok := ctx.Value("retry_count").(int); ok {
        retryCount = rc
    }

    // Check if exhausted
    if retryCount >= maxRetries {
        // Publish to DLQ for manual intervention
        s.bus.Publish(context.Background(), "dlq.settlement.requested", ...)
        return fmt.Errorf("settlement failed permanently")
    }

    // Process settlement...
    if success := s.simulateSettlementProcessing(ctx, settlement); !success {
        return fmt.Errorf("settlement processing failed")  // Triggers retry
    }

    return nil  // Success
}
```

## Key Benefits

1. **Automatic**: RabbitMQ handles retry tracking and DLQ routing
2. **Reliable**: Uses native RabbitMQ features (no custom state)
3. **Visible**: x-death header provides complete delivery history
4. **Flexible**: Easy to adjust max retries and TTL
5. **Scalable**: Workers and DLQ monitor scale independently

## Production Considerations

- **Message TTL**: 5 minutes prevents stale messages
- **DLQ Durability**: Messages persisted to prevent loss
- **Monitoring**: DLQ monitor provides operational visibility
- **Manual Intervention**: Failed messages available for replay
- **Metrics**: Track retry rates and DLQ message counts
