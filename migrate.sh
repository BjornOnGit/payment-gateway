#!/bin/bash
set -e

# Run database migrations using Go
# Usage: ./migrate.sh <DATABASE_URL> [migrations_dir]

if [ -z "$1" ]; then
    echo "Usage: ./migrate.sh <DATABASE_URL> [migrations_dir]"
    echo ""
    echo "Example:"
    echo "  ./migrate.sh 'postgres://user:pass@neon.tech:5432/db' ./configs/migrations"
    echo ""
    echo "Or set DATABASE_URL environment variable and run:"
    echo "  ./migrate.sh"
    exit 1
fi

# Set DATABASE_URL if provided as argument
if [ -n "$1" ]; then
    export DATABASE_URL="$1"
fi

# Set migrations directory (default: ./configs/migrations)
MIGRATIONS_DIR="${2:-./configs/migrations}"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "Error: Migrations directory not found: $MIGRATIONS_DIR"
    exit 1
fi

echo "Running migrations from: $MIGRATIONS_DIR"
go run cmd/migrate/main.go "$MIGRATIONS_DIR"
