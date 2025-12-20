# Payment Gateway

A production-ready payment gateway system built with Go, featuring asynchronous event-driven architecture, automatic retries, and dead letter queue (DLQ) support.

## Features

- **Async Processing**: RabbitMQ-based message bus for distributed event processing
- **Transaction Management**: Full transaction lifecycle from creation to settlement
- **Automatic Retries**: Up to 3 automatic retries for transient failures
- **Dead Letter Queue**: Permanent failure handling with manual intervention support
- **Structured Logging**: Comprehensive logging with trace context using zap
- **Reconciliation**: Batch job for transaction vs settlement reconciliation

## Architecture

### Services

1. **API Server** (`cmd/api`)

- REST API for transaction creation
- JWT authentication
- Publishes events to RabbitMQ

1. **Transaction Worker** (`cmd/transaction-worker`)

- Processes `transaction.created` events
- Routes to appropriate payment provider
- Publishes `settlement.requested` events

1. **Settlement Worker** (`cmd/settlement-worker`)

- Processes `settlement.requested` events
- Creates settlement records
- Automatic retry on failure (max 3 attempts)
- Routes permanent failures to DLQ

1. **DLQ Monitor** (`cmd/dlq-monitor`)

- Monitors dead letter queue
- Logs permanently failed settlements
- Foundation for manual intervention

2. **Reconciliation Job** (`cmd/reconcile-job`)

- Batch job for settlement reconciliation
- Detects transaction vs settlement mismatches
- Webhook notifications

### Message Flow

```
API → transaction.created → Transaction Worker → settlement.requested → Settlement Worker
																										│
																			  ┌──────────────────┴──────────────┐
																			  │                                 │
																		 SUCCESS                           FAILURE
																			  │                                 │
																	 settlement.completed              Retry (up to 3x)
																														  │
																												  Max retries
																														  │
																											  dlq.settlement.requested
																														  │
																												  DLQ Monitor
```

## Getting Started

### Prerequisites

- Go 1.23.5+
- PostgreSQL 13+
- RabbitMQ 3.x
- Docker (optional, for running services)

### Setup

1. **Clone the repository**
	```bash
	git clone <repository-url>
	cd payment-gateway
	```

2. **Start RabbitMQ**
	```bash
	docker run -d --name rabbitmq \
	  -p 5672:5672 \
	  -p 15672:15672 \
	  rabbitmq:3-management
	```

3. **Configure environment**
	```bash
	cp .env.example .env
	# Edit .env with your database and RabbitMQ settings
	```

4. **Run database migrations**
	```bash
	# Apply migrations in configs/migrations/
	psql -d payment_gateway -f configs/migrations/0001_init.sql
	```

5. **Build all services**
	```bash
	go build -o bin/api ./cmd/api
	go build -o bin/transaction-worker ./cmd/transaction-worker
	go build -o bin/settlement-worker ./cmd/settlement-worker
	go build -o bin/dlq-monitor ./cmd/dlq-monitor
	go build -o bin/reconcile-job ./cmd/reconcile-job
	```

6. **Start services**
	```bash
	# Terminal 1: API Server
	./bin/api

	# Terminal 2: Transaction Worker
	./bin/transaction-worker

	# Terminal 3: Settlement Worker
	./bin/settlement-worker

	# Terminal 4: DLQ Monitor
	./bin/dlq-monitor
	```

## DLQ Implementation

The system includes comprehensive Dead Letter Queue support for handling permanent failures.

### How It Works

1. **Automatic Retries**: Settlement failures trigger up to 3 automatic retries
2. **Retry Tracking**: RabbitMQ `x-death` header tracks delivery attempts
3. **DLQ Routing**: After max retries, messages automatically route to DLQ
4. **Monitoring**: DLQ monitor logs all permanent failures

### Testing DLQ Flow

Run the automated test:
```bash
./test-dlq.sh
```

Or manually:
1. Force failures by editing `internal/service/settlement_service.go`:
	```go
	func (s *SettlementService) simulateSettlementProcessing(...) bool {
		 return false // Force failure
	}
	```
2. Start all services
3. Create a transaction via API
4. Watch logs for retry attempts and DLQ routing

### Documentation

- [Complete DLQ Implementation Guide](docs/DLQ_IMPLEMENTATION.md)
- [DLQ Architecture & Flow Diagrams](docs/DLQ_FLOW_DIAGRAM.md)
- [Implementation Summary](DLQ_SUMMARY.md)
- [Testing Checklist](DLQ_CHECKLIST.md)

## Configuration

### Environment Variables

```bash
# Database
DB_URL=postgres://user:pass@localhost/payment_gateway

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Logging
LOG_SERVICE_NAME=api  # or transaction-worker, settlement-worker, dlq-monitor
LOG_LEVEL=info

# Server
PORT=8080
```

### RabbitMQ Settings

- **Max Retries**: 3 attempts (configured in `bus_rmq.go`)
- **Message TTL**: 5 minutes (300,000 ms)
- **Exchange Type**: Topic
- **DLQ Pattern**: `dlq.<topic-name>` (e.g., `dlq.settlement.requested`)

## Development

### Project Structure

```
payment-gateway/
├── cmd/
│   ├── api/                    # API server
│   ├── transaction-worker/     # Transaction processing worker
│   ├── settlement-worker/      # Settlement processing worker
│   ├── dlq-monitor/           # DLQ monitoring service
│   └── reconcile-job/         # Reconciliation batch job
├── internal/
│   ├── bus/
│   │   └── rabbitmq/          # RabbitMQ implementation with DLQ support
│   ├── db/
│   │   └── repo/              # Database repositories
│   ├── model/                 # Data models
│   ├── service/               # Business logic (routing, settlement)
│   └── util/                  # Logger and utilities
├── configs/
│   └── migrations/            # Database migrations
├── docs/                      # Documentation
├── test-dlq.sh               # DLQ testing script
└── README.md
```

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests (requires RabbitMQ)
go test ./... -tags=integration

# DLQ flow test
./test-dlq.sh
```

### Building for Production

```bash
# Build with optimizations
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api ./cmd/api
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/settlement-worker ./cmd/settlement-worker
# ... repeat for other services
```

## Monitoring & Operations

### Key Metrics

- Settlement success rate
- Average retry count
- DLQ message rate
- Transaction processing time
- Settlement processing time

### Alerts

Configure alerts for:
- DLQ messages received (immediate attention needed)
- High retry rate (> 30%)
- Settlement processing failures
- Queue depth anomalies

### RabbitMQ Management

Access RabbitMQ management UI:
```bash
open http://localhost:15672
# Default credentials: guest/guest
```

Monitor:
- Message rates
- Queue depths
- Consumer status
- DLQ message counts

## Troubleshooting

### DLQ Messages Not Routing

**Issue**: Messages retry indefinitely, never reach DLQ

**Solution**:
```bash
# Delete existing queues/exchanges
docker exec rabbitmq rabbitmqctl purge_queue settlement.requested
docker exec rabbitmq rabbitmqctl delete_exchange settlement.requested

# Restart workers to recreate with correct DLQ config
```

### High Retry Rate

**Issue**: Many messages failing and retrying

**Investigate**:
1. Check settlement service logs for error patterns
2. Verify external API availability
3. Check database connectivity
4. Review network latency

### DLQ Monitor Not Receiving Messages

**Issue**: Messages in DLQ but not being logged

**Solution**:
1. Verify DLQ monitor is running
2. Check topic name matches (`dlq.settlement.requested`)
3. Review RabbitMQ bindings in management UI
4. Restart DLQ monitor

## Production Deployment

### Checklist

- [ ] Configure RabbitMQ cluster for HA
- [ ] Set up monitoring and alerting
- [ ] Deploy DLQ monitor with alerting integration
- [ ] Create runbook for DLQ incidents
- [ ] Train on-call team
- [ ] Document rollback procedures
- [ ] Load test with realistic traffic
- [ ] Security review completed

### Scaling

**Horizontal Scaling**:
- Run multiple instances of each worker
- RabbitMQ distributes messages across consumers
- Scale workers independently based on load

**Vertical Scaling**:
- Increase worker resources for CPU-intensive operations
- Scale database for high transaction volumes
- Consider read replicas for reconciliation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Why Contribute
- Impact: Help shape a production-style payment gateway with async processing, DLQ, and reconciliation flows.
- Learning: Build real-world experience with Go, RabbitMQ, and a modern Next.js frontend.
- Collaboration: Improve reliability, docs, and test coverage alongside maintainers.
- Recognition: Meaningful contributions are acknowledged; maintainership is open over time.

More details: see [CONTRIBUTING.md](CONTRIBUTING.md), [CODE OF CONDUCT.md](CODE%20OF%20CONDUCT.md), and [SECURITY.md](SECURITY.md).

## License

[LICENSE](LICENSE)

## Support

For issues and questions:
- Check the [documentation](docs/)
- Review [troubleshooting guide](#troubleshooting)
- Open an issue on GitHub

## Documentation Index
- [docs/ALERT_CHECKLIST.md](docs/ALERT_CHECKLIST.md)
- [docs/ALERT_IMPLEMENTATION_SUMMARY.md](docs/ALERT_IMPLEMENTATION_SUMMARY.md)
- [docs/ALERT_QUICKSTART.md](docs/ALERT_QUICKSTART.md)
- [docs/ALERT_SYSTEM.md](docs/ALERT_SYSTEM.md)
- [docs/DLQ_CHECKLIST.md](docs/DLQ_CHECKLIST.md)
- [docs/DLQ_FLOW_DIAGRAM.md](docs/DLQ_FLOW_DIAGRAM.md)
- [docs/DLQ_IMPLEMENTATION.md](docs/DLQ_IMPLEMENTATION.md)
- [docs/DLQ_SUMMARY.md](docs/DLQ_SUMMARY.md)
- [docs/METRICS_CHECKLIST.md](docs/METRICS_CHECKLIST.md)
- [docs/METRICS_INTEGRATION.md](docs/METRICS_INTEGRATION.md)
- [docs/METRICS_QUICKSTART.md](docs/METRICS_QUICKSTART.md)
- [docs/POSTMAN_COLLECTION_GUIDE.md](docs/POSTMAN_COLLECTION_GUIDE.md)
- [docs/POSTMAN_FIXES.md](docs/POSTMAN_FIXES.md)
