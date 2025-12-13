# Quick Start: Alert System

## Overview

The payment gateway includes a structured alerting system that services can use to send alerts about critical events, failures, and anomalies.

## Quick Start (5 minutes)

### 1. Start Alert Receiver

```bash
go run ./cmd/alert-receiver
```

You'll see:
```
Alert webhook receiver listening on :9000
  POST /webhook - Receive alerts from services
  GET  /alerts  - View all received alerts
  POST /clear   - Clear alert history
```

### 2. Run the Demo

In another terminal:
```bash
./demo-alerts.sh
```

This sends 4 test alerts (info, warning, error, critical) and shows how they're received.

### 3. View Alerts

```bash
# View all alerts
curl http://localhost:9000/alerts | jq

# Count alerts
curl http://localhost:9000/alerts | jq '.count'

# Get just the latest alert
curl http://localhost:9000/alerts | jq '.alerts[-1]'
```

## Using AlertClient in Your Code

### Basic Usage

```go
import "github.com/BjornOnGit/payment-gateway/internal/util"

// Create client (usually in main.go)
alertClient := util.NewAlertClient("http://localhost:9000/webhook", logger)

// Send alerts
alertClient.SendWarning(ctx, "my-service", "Something happened", map[string]any{
    "key": "value",
})
```

### Severity Levels

```go
// Critical: Service down, immediate action needed
alertClient.SendCritical(ctx, "service", "Database down", details)

// Error: Significant problem
alertClient.SendError(ctx, "service", "Transaction failed", details)

// Warning: Notable issue
alertClient.SendWarning(ctx, "service", "Mismatch detected", details)

// Info: Informational
alertClient.SendInfo(ctx, "service", "Job started", details)
```

## Current Integration: Reconciliation Job

The reconciliation job has been updated to use AlertClient:

```bash
# Set webhook URL
export ALERT_WEBHOOK_URL=http://localhost:9000/webhook

# Run reconciliation
go run ./cmd/reconcile-job
```

On mismatch:
```
[warning] reconcile-job: Transaction amount mismatch detected | 
  Details: map[actual_amount:900 difference:100 expected_amount:1000 transaction_id:...]
```

## API Reference

### Alert Structure

```json
{
  "service": "my-service",
  "severity": "warning",
  "message": "Something happened",
  "details": {
    "transaction_id": "123",
    "amount": 1000
  },
  "timestamp": "2025-12-11T10:30:45Z"
}
```

### Endpoints

**POST /webhook** - Send alert
```bash
curl -X POST http://localhost:9000/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "service": "test",
    "severity": "warning",
    "message": "Test alert"
  }'
```

**GET /alerts** - Get all alerts
```bash
curl http://localhost:9000/alerts
```

**POST /clear** - Clear history
```bash
curl -X POST http://localhost:9000/clear
```

## Examples

### Send Different Severity Alerts

```bash
# Info
./demo-alerts.sh send test info "Info message"

# Warning
./demo-alerts.sh send test warning "Warning message"

# Error
./demo-alerts.sh send test error "Error message"

# Critical
./demo-alerts.sh send test critical "Critical message"
```

### Send with Details

```bash
./demo-alerts.sh send my-service error "Processing failed" \
  '{"transaction_id":"123","error":"timeout"}'
```

### Filter Alerts

```bash
# Get only warnings and errors
curl http://localhost:9000/alerts | \
  jq '.alerts[] | select(.severity | test("warning|error"))'

# Get alerts from specific service
curl http://localhost:9000/alerts | \
  jq '.alerts[] | select(.service == "reconcile-job")'

# Get recent alerts
curl http://localhost:9000/alerts | jq '.alerts[-3:]'
```

## Production Setup

### Connect to Email

Use a service like SendGrid:

```bash
export ALERT_WEBHOOK_URL=https://api.sendgrid.com/v3/mail/send
```

### Connect to Slack

Use Slack incoming webhooks:

```bash
export ALERT_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### Connect to PagerDuty

```bash
export ALERT_WEBHOOK_URL=https://events.pagerduty.com/v2/enqueue
```

### Custom Endpoint

Point to your own service:

```bash
export ALERT_WEBHOOK_URL=https://your-api.com/alerts
```

## Configuration

The AlertClient respects the webhook URL from environment:

```bash
# If not set, alerts are silently skipped (logged as debug)
export ALERT_WEBHOOK_URL=http://localhost:9000/webhook

# Services will auto-use this URL
go run ./cmd/reconcile-job
```

## Troubleshooting

### Alerts Not Received?

1. Check alert receiver is running:
   ```bash
   curl http://localhost:9000/alerts
   ```

2. Check webhook URL is correct:
   ```bash
   echo $ALERT_WEBHOOK_URL
   ```

3. Check service logs for errors:
   ```bash
   # Look for "failed to send alert"
   go run ./cmd/reconcile-job 2>&1 | grep -i alert
   ```

### Want to Extend DLQ Monitor?

```go
// In cmd/dlq-monitor/main.go

// Create alert client
alertClient := util.NewAlertClient(os.Getenv("ALERT_WEBHOOK_URL"), logger)

// In the handler
alertClient.SendCritical(ctx, "dlq-monitor",
    "Settlement failed permanently",
    map[string]any{
        "transaction_id": key,
        "topic": "dlq.settlement.requested",
    })
```

## Files

- **Client**: `internal/util/alert.go` - AlertClient implementation
- **Receiver**: `cmd/alert-receiver` - Test server
- **Demo**: `demo-alerts.sh` - Interactive demo
- **Examples**: `examples/alert-integration.go` - Integration patterns
- **Docs**: `docs/ALERT_SYSTEM.md` - Full documentation

## Next Steps

1. ✅ Understand alert structure and severity levels
2. ✅ Try the demo: `./demo-alerts.sh`
3. ✅ View alerts: `curl http://localhost:9000/alerts`
4. Configure production webhook URL (email, Slack, PagerDuty, etc.)
5. Extend other services (DLQ monitor, settlement worker, etc.) to send alerts
6. Set up monitoring and dashboards for alert metrics
