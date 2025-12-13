#!/usr/bin/env bash
set -euo pipefail

COLLECTION_FILE="$(pwd)/postman_collection.json"
ENV_FILE="$(pwd)/tests/e2e/env.json" # optional env file
NEWMAN_CMD="$(command -v newman || true)"

if [ -z "$NEWMAN_CMD" ]; then
  echo "Newman not installed. Install via: npm install -g newman"
  exit 2
fi

if [ ! -f "$COLLECTION_FILE" ]; then
  echo "Collection not found at $COLLECTION_FILE"
  exit 1
fi

if [ -f "$ENV_FILE" ]; then
  newman run "$COLLECTION_FILE" -e "$ENV_FILE" --bail --insecure
else
  newman run "$COLLECTION_FILE" --bail --insecure
fi
