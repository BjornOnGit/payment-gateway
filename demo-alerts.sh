#!/bin/bash

# Demonstration of the Alert System

set -e

echo "=================================================="
echo "Payment Gateway - Alert System Demo"
echo "=================================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if alert receiver is running
check_alert_receiver() {
    if ! curl -s http://localhost:9000/alerts > /dev/null 2>&1; then
        echo -e "${RED}Alert receiver not running on :9000${NC}"
        echo "Start it with: go run ./cmd/alert-receiver"
        exit 1
    fi
}

# Send test alert
send_alert() {
    local service=$1
    local severity=$2
    local message=$3
    local details=$4

    echo -e "${BLUE}Sending alert...${NC}"
    curl -s -X POST http://localhost:9000/webhook \
        -H "Content-Type: application/json" \
        -d "{
            \"service\": \"$service\",
            \"severity\": \"$severity\",
            \"message\": \"$message\",
            \"details\": $details,
            \"timestamp\": \"$(date -u +'%Y-%m-%dT%H:%M:%SZ')\"
        }" | jq '.'
    echo ""
}

# View alerts
view_alerts() {
    echo -e "${BLUE}Current alerts:${NC}"
    curl -s http://localhost:9000/alerts | jq '.'
    echo ""
}

# Clear alerts
clear_alerts() {
    echo -e "${BLUE}Clearing alert history...${NC}"
    curl -s -X POST http://localhost:9000/clear | jq '.'
    echo ""
}

# Demo flow
demo() {
    echo -e "${YELLOW}Step 1: Check alert receiver status${NC}"
    check_alert_receiver
    echo -e "${GREEN}✓ Alert receiver is running${NC}"
    echo ""

    echo -e "${YELLOW}Step 2: Clear any existing alerts${NC}"
    clear_alerts

    echo -e "${YELLOW}Step 3: Send INFO alert${NC}"
    send_alert "demo-service" "info" "Demo started" '{"demo":"true"}'

    echo -e "${YELLOW}Step 4: Send WARNING alert${NC}"
    send_alert "reconcile-job" "warning" "Transaction mismatch detected" '{"transaction_id":"123e4567-e89b-12d3-a456-426614174000","expected":1000,"actual":900,"difference":100}'

    echo -e "${YELLOW}Step 5: Send ERROR alert${NC}"
    send_alert "settlement-worker" "error" "Settlement processing failed" '{"transaction_id":"456f7890-a90c-23d3-b567-537725285111","reason":"external_api_timeout","retry_count":3}'

    echo -e "${YELLOW}Step 6: Send CRITICAL alert${NC}"
    send_alert "api-server" "critical" "Database connection lost" '{"error":"connection_refused","host":"localhost","port":5432}'

    echo -e "${YELLOW}Step 7: View all received alerts${NC}"
    view_alerts

    echo -e "${YELLOW}Step 8: Query specific severity alerts${NC}"
    echo -e "${BLUE}Warning and Error alerts:${NC}"
    curl -s http://localhost:9000/alerts | jq '.alerts[] | select(.severity | test("warning|error")) | {service, severity, message}' 
    echo ""

    echo -e "${YELLOW}Step 9: Query by service${NC}"
    echo -e "${BLUE}Alerts from reconcile-job:${NC}"
    curl -s http://localhost:9000/alerts | jq '.alerts[] | select(.service == "reconcile-job")'
    echo ""

    echo -e "${GREEN}✓ Demo completed!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Review the alerts received above"
    echo "2. Try sending custom alerts:"
    echo "   curl -X POST http://localhost:9000/webhook \\"
    echo "     -H 'Content-Type: application/json' \\"
    echo "     -d '{\"service\":\"my-service\",\"severity\":\"warning\",\"message\":\"Custom alert\"}'"
    echo "3. View alerts: curl http://localhost:9000/alerts | jq"
    echo "4. Clear history: curl -X POST http://localhost:9000/clear"
}

# Main
case "${1:-demo}" in
    demo)
        demo
        ;;
    send)
        if [ $# -lt 4 ]; then
            echo "Usage: $0 send <service> <severity> <message> [details_json]"
            exit 1
        fi
        send_alert "$2" "$3" "$4" "${5:-{}}"
        ;;
    view)
        view_alerts
        ;;
    clear)
        clear_alerts
        ;;
    *)
        echo "Usage: $0 [demo|send|view|clear]"
        echo ""
        echo "Commands:"
        echo "  demo - Run full demo"
        echo "  send <service> <severity> <message> [details] - Send alert"
        echo "  view - View all alerts"
        echo "  clear - Clear alert history"
        exit 1
        ;;
esac
