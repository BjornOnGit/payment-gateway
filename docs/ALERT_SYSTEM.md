# Alert System

## Overview

The payment gateway includes a comprehensive alert system for monitoring critical events and failures. This system uses webhook-based alerts that can be integrated with any alerting platform (PagerDuty, Slack, email services, custom dashboards, etc.).

## Architecture

### Components

1. **Alert Client** (`internal/util/alert.go`)
   - Sends structured alerts to a webhook endpoint
   - Supports multiple severity levels (critical, error, warning, info)
   - Includes service name, timestamp, and contextual details
   - Non-blocking: failures to send alerts are logged but don't block operations

2. **Alert Receiver** (`cmd/alert-receiver`)
   - Simple HTTP server for receiving and storing alerts
   - Perfect for testing and development
   - Provides REST API to query alert history
   - Can be extended for production use

3. **Alert Producers**
   - Reconciliation job: Sends alerts on transaction mismatches
   - DLQ monitor: Can be extended to alert on permanent failures
   - Any service can use AlertClient to send alerts

## Usage

### For Services (Sending Alerts)

```go
import "github.com/BjornOnGit/payment-gateway/internal/util"

// Create alert client
alertClient := util.NewAlertClient(
    "http://localhost:9000/webhook",
    logger,
)

// Send alerts
alertClient.SendWarning(ctx, "my-service", "Something concerning happened", map[string]any{
    "transaction_id": "123e4567-e89b-12d3-a456-426614174000",
    "amount": 1000,
})

alertClient.SendCritical(ctx, "my-service", "Critical failure", map[string]any{
    "error": "external API down",
})

// Convenience methods
alertClient.SendError(ctx, "service", "message", details)
alertClient.SendInfo(ctx, "service", "message", details)
```

### Alert Structure

```json
{
  "service": "reconcile-job",
  "severity": "warning",
  "message": "Transaction amount mismatch detected",
  "details": {
    "transaction_id": "123e4567...",
    "expected_amount": 1000,
    "actual_amount": 900,
    "difference": 100
  },
  "timestamp": "2025-12-11T10:30:45Z"
}
```

### Severity Levels

- **critical**: Immediate action required (service down, data loss risk)
- **error**: Significant problem (failed transaction, API error)
- **warning**: Notable issue (mismatch, retry failure, unusual pattern)
- **info**: Informational (job started, threshold reached)

## Running the Alert Receiver

### Start Alert Receiver

```bash
go run ./cmd/alert-receiver
```

Output:
```
Alert webhook receiver listening on :9000
  POST /webhook - Receive alerts from services
  GET  /alerts  - View all received alerts
  POST /clear   - Clear alert history
```

### API Endpoints

**POST /webhook** - Receive alert
```bash
curl -X POST http://localhost:9000/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "service": "test-service",
    "severity": "warning",
    "message": "Test alert",
    "timestamp": "2025-12-11T10:30:45Z"
  }'
```

**GET /alerts** - View all alerts
```bash
curl http://localhost:9000/alerts
```

Response:
```json
{
  "count": 3,
  "alerts": [
    {
      "service": "reconcile-job",
      "severity": "warning",
      "message": "Transaction amount mismatch detected",
      "details": {...},
      "timestamp": "2025-12-11T10:30:45Z"
    },
    ...
  ]
}
```

**POST /clear** - Clear alert history
```bash
curl -X POST http://localhost:9000/clear
```

## Integration with Services

### Reconciliation Job

The reconciliation job now uses AlertClient to send alerts on transaction mismatches:

```bash
export ALERT_WEBHOOK_URL=http://localhost:9000/webhook
go run ./cmd/reconcile-job
```

On mismatch detection:
```
[warning] reconcile-job: Transaction amount mismatch detected | Details: map[actual_amount:900 difference:100 expected_amount:1000 transaction_id:123e4567...]
```

### DLQ Monitor Enhancement

You can extend the DLQ monitor to send alerts:

```go
alertClient := util.NewAlertClient(os.Getenv("ALERT_WEBHOOK_URL"), logger)

if err := alertClient.SendCritical(ctx, "dlq-monitor",
    "Settlement failed permanently - in DLQ",
    map[string]any{
        "transaction_id": key,
        "payload": string(payload),
    }); err != nil {
    logger.Error("failed to send alert", zap.Error(err))
}
```

## Production Integration

### Email Alerts

Use a service like SendGrid or AWS SES:

```go
// Configure webhook to email service
ALERT_WEBHOOK_URL=https://your-email-service.com/alerts
```

### Slack Integration

Use Slack webhooks:

```bash
export ALERT_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

Format alerts in AlertClient to match Slack's message format (optional).

### PagerDuty Integration

Use PagerDuty's webhook API:

```bash
export ALERT_WEBHOOK_URL=https://events.pagerduty.com/v2/enqueue
```

Create a middleware service to transform alerts to PagerDuty format.

### Custom Dashboard

Create a service that receives alerts and stores them:

```go
// Store alert in database
db.InsertAlert(ctx, alert)

// Trigger UI refresh
hub.Broadcast(alert)

// Send to time-series DB for metrics
prometheus.AlertsCounter.Inc()
```

## Configuration

### Environment Variables

```bash
# Alert webhook URL (optional, alerts are disabled if not set)
ALERT_WEBHOOK_URL=http://localhost:9000/webhook

# Service name for logging
LOG_SERVICE_NAME=reconcile-job

# Log level
LOG_LEVEL=info
```

### Timeout

The AlertClient has a 10-second timeout for HTTP requests. This prevents blocking operations when the alert endpoint is slow or unreachable.

### Non-Blocking

Failed alert sends don't stop operations:
- Logged as warnings
- Operation continues normally
- Retry logic only applies to the HTTP request itself

## Monitoring

### Key Metrics

- Alert count by severity
- Alert latency (send time)
- Alert delivery failures
- Alerts by service
- Top alert messages

### Dashboards

Create dashboards showing:
- Real-time alert stream
- Alert frequency over time
- Alert distribution by service and severity
- Alert response time

### Querying Alerts

```bash
# Get all alerts
curl http://localhost:9000/alerts | jq '.alerts | length'

# Get warning/error alerts
curl http://localhost:9000/alerts | jq '.alerts[] | select(.severity | test("warning|error"))'

# Get alerts from specific service
curl http://localhost:9000/alerts | jq '.alerts[] | select(.service == "reconcile-job")'

# Get recent alerts
curl http://localhost:9000/alerts | jq '.alerts[-5:]'
```

## Testing

### Test Alert Flow

```bash
# Terminal 1: Start alert receiver
go run ./cmd/alert-receiver

# Terminal 2: Send test alert
curl -X POST http://localhost:9000/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "service": "test-service",
    "severity": "warning",
    "message": "Test alert",
    "timestamp": "'$(date -u +'%Y-%m-%dT%H:%M:%SZ')'"
  }'

# Terminal 1: Should show
# [warning] test-service: Test alert | Details: map[]

# Terminal 3: Verify alert received
curl http://localhost:9000/alerts | jq '.alerts[-1]'
```

### Test with Reconciliation Job

```bash
# Terminal 1: Start alert receiver
go run ./cmd/alert-receiver

# Terminal 2: Run reconciliation job
export ALERT_WEBHOOK_URL=http://localhost:9000/webhook
go run ./cmd/reconcile-job

# Should see alerts for any mismatches
curl http://localhost:9000/alerts | jq '.count'
```

## Best Practices

1. **Always configure alerts in production**
   - Use appropriate webhook endpoints for your platform
   - Test integration before deployment

2. **Set appropriate severity levels**
   - Critical: Service down, data loss
   - Error: Transaction failed, API error
   - Warning: Mismatch, retry exceeded
   - Info: Job started, threshold met

3. **Include meaningful details**
   - Transaction IDs for tracing
   - Error messages for debugging
   - Context for quick understanding

4. **Monitor the monitoring**
   - Track alert delivery success
   - Alert on high alert rates
   - Regular review of alert patterns

5. **Handle alert failures gracefully**
   - Don't block operations
   - Log failures for investigation
   - Implement fallback alerting

## Troubleshooting

### Alerts Not Being Sent

**Check configuration**:
```bash
echo $ALERT_WEBHOOK_URL
```

**Verify webhook URL is reachable**:
```bash
curl -v http://localhost:9000/webhook
```

**Check logs**:
```bash
# Look for "failed to send alert" messages
go run ./cmd/reconcile-job 2>&1 | grep -i alert
```

### Alerts Not Being Received

**Verify alert receiver is running**:
```bash
curl http://localhost:9000/alerts
```

**Check alert receiver logs**:
- Should show `[severity] service: message` for each alert received

**Verify webhook URL matches**:
```bash
# In service config
export ALERT_WEBHOOK_URL=http://localhost:9000/webhook

# In alert receiver
curl http://localhost:9000/alerts
```

### High Latency

**Check alert receiver performance**:
- Is it processing alerts slowly?
- Are there many alerts arriving at once?

**Adjust timeout if needed**:
- Default: 10 seconds
- Can be modified in `NewAlertClient`

## Future Enhancements

1. **Alert Deduplication**: Prevent duplicate alerts in short time windows
2. **Alert Routing**: Different severity levels to different endpoints
3. **Alert Grouping**: Batch related alerts together
4. **Alert Enrichment**: Add more context (user, region, environment)
5. **Alert Templates**: Customize message format per integration
6. **Alert History**: Persist alerts in database
7. **Alert Dashboard**: Web UI for viewing and managing alerts
8. **Alert Acknowledgment**: Mark alerts as handled
