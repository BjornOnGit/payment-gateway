# Metrics Integration - Payment Gateway API

## Overview

The payment gateway API now includes Prometheus metrics support for monitoring request patterns, transaction creation, and settlement success rates.

## What Was Integrated

### Metrics Utility (`internal/util/metrics.go`)

**Metrics Defined**:
- `requests_total` - Total HTTP requests, labeled by method and path
- `transactions_created_total` - Total transactions created
- `settlements_succeeded_total` - Total successful settlements

**Handlers**:
- `MetricsHandler()` - Returns Prometheus HTTP handler for `/metrics` endpoint
- `MetricsMiddleware(handler)` - Middleware to track all requests

### API Server Integration (`cmd/api/main.go`)

**Setup**:
```go
// Create a mux for routing including metrics
mux := http.NewServeMux()

// Add metrics endpoint (public, no auth required)
mux.Handle("/metrics", util.MetricsHandler())

// Delegate all other requests to the main router
mux.Handle("/", router)

// Wrap entire mux with metrics middleware to track all requests
handler := util.MetricsMiddleware(mux)

// Use metrics-wrapped handler in server
server := &http.Server{
    Addr:    ":" + port,
    Handler: handler,
}
```

## Usage

### Accessing Metrics

Once the API server is running:

```bash
curl http://localhost:8080/metrics
```

This returns Prometheus-formatted metrics.

### Example Output

```
# HELP requests_total Total HTTP requests processed, labeled by method and path.
# TYPE requests_total counter
requests_total{method="POST",path="/v1/transactions"} 5
requests_total{method="GET",path="/health"} 10
requests_total{method="POST",path="/oauth/token"} 3

# HELP transactions_created_total Total transactions created
# TYPE transactions_created_total counter
transactions_created_total 5

# HELP settlements_succeeded_total Total successful settlements
# TYPE settlements_succeeded_total counter
settlements_succeeded_total 3
```

## How It Works

### Request Tracking

When a request comes in:

1. **MetricsMiddleware** intercepts the request
2. Increments `requests_total` counter with method and path labels
3. Passes request to the main handler
4. Request is processed normally

```go
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        RequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()
        next.ServeHTTP(w, r)
    })
}
```

### Custom Metrics

Services can increment metrics directly:

**In Transaction Service**:
```go
util.TransactionsCreatedTotal.Inc()
```

**In Settlement Service**:
```go
util.SettlementsSucceededTotal.Inc()
```

## Architecture

```
HTTP Request
    ↓
MetricsMiddleware (tracks request)
    ↓
Main Router
    ├─ /metrics → MetricsHandler (Prometheus format)
    ├─ /health → Health check
    ├─ /oauth/token → OAuth token endpoint
    └─ /v1/transactions → Transaction creation
    ↓
Response
```

## Monitoring

### Prometheus Integration

Add to `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'payment-gateway'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Key Metrics to Monitor

1. **Request Rate**: `rate(requests_total[1m])`
   - Understand API usage patterns
   - Detect anomalies

2. **Transaction Rate**: `rate(transactions_created_total[1m])`
   - Monitor transaction throughput
   - Track business metrics

3. **Settlement Success Rate**: `settlements_succeeded_total / transactions_created_total`
   - Monitor settlement completion
   - Track operational health

4. **Request Distribution**: `requests_total` by method and path
   - Identify most used endpoints
   - Detect unusual traffic patterns

## Grafana Dashboards

### Example Queries

**Request Rate (per minute)**:
```
rate(requests_total[1m])
```

**Transaction Rate (per minute)**:
```
rate(transactions_created_total[1m])
```

**Settlement Success Rate**:
```
settlements_succeeded_total / transactions_created_total
```

**Request Count by Endpoint**:
```
sum(requests_total) by (path)
```

**Request Count by Method**:
```
sum(requests_total) by (method)
```

## Performance Considerations

### Minimal Overhead

- **MetricsMiddleware**: O(1) operation per request
- **Counter increment**: Microsecond scale
- **No request blocking**: Metrics are updated asynchronously by Prometheus

### Memory Impact

- Static metrics defined at startup
- Memory grows with unique label combinations (method × path)
- Typical payload: < 1KB per endpoint

## Configuration

### Custom Metrics

To add new metrics:

1. **Define in `internal/util/metrics.go`**:
```go
var CustomMetric = prometheus.NewCounter(
    prometheus.CounterOpts{
        Name: "custom_metric_total",
        Help: "Description",
    },
)
```

2. **Register in init()**:
```go
func init() {
    prometheus.MustRegister(CustomMetric)
}
```

3. **Use in code**:
```go
util.CustomMetric.Inc()
```

### Metrics Endpoint

By default: `http://localhost:8080/metrics`

Configure via `API_PORT`:
```bash
export API_PORT=9000
# Metrics then available at: http://localhost:9000/metrics
```

## Debugging

### Check Metrics are Being Recorded

```bash
# While API is running
curl -s http://localhost:8080/metrics | grep -v "^#"
```

### View Specific Metric

```bash
curl -s http://localhost:8080/metrics | grep "requests_total"
```

### Monitor in Real-Time

```bash
watch -n 1 "curl -s http://localhost:8080/metrics | grep -v '^#'"
```

## Best Practices

1. **Keep Metrics Focused**: Track meaningful business and operational metrics
2. **Use Labels Wisely**: Avoid high-cardinality labels (like user IDs) that create many metrics
3. **Monitor the Metrics**: Set up alerts on metric rate changes
4. **Export to Time-Series DB**: Use Prometheus for scraping and storage
5. **Visualize**: Use Grafana for dashboards
6. **Alert**: Configure alert rules based on metric thresholds

## Integration with Alert System

Combine with the alert system for automated incident response:

```go
// In metrics handler or middleware
if settlementFailureRate > 0.1 { // > 10% failure
    alertClient.SendCritical(ctx, "metrics-monitor",
        "High settlement failure rate detected",
        map[string]any{
            "failure_rate": settlementFailureRate,
            "threshold": 0.1,
        })
}
```

## Future Enhancements

1. **Histogram Metrics**: Track request latency distribution
2. **Gauge Metrics**: Monitor active connections, queue depth
3. **Custom Business Metrics**: Revenue, customer metrics
4. **Error Tracking**: Error rates by type
5. **Database Metrics**: Query latency, connection pool stats
6. **Message Queue Metrics**: RabbitMQ message rates
7. **Automatic Alerting**: Alert on metric anomalies

## Verification

Check that metrics are working:

```bash
# 1. Start API server
go run ./cmd/api

# 2. Make some requests
curl -k https://localhost:8080/health

# 3. View metrics
curl http://localhost:8080/metrics | head -20
```

You should see:
- `requests_total` with your health check request
- `# TYPE requests_total counter`
- Counter values and labels

## Dependencies

- `github.com/prometheus/client_golang` - Already included in go.mod
- Standard library (`net/http`)

No additional dependencies needed beyond what's already configured.
