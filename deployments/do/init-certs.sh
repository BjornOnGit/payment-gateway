#!/bin/bash
set -e

# Generate SSL certificates if they don't exist
if [ ! -f "dev-certs/server.crt" ]; then
    echo "Generating SSL certificates..."
    mkdir -p dev-certs
    openssl req -x509 -newkey rsa:4096 -keyout dev-certs/server.key -out dev-certs/server.crt -days 365 -nodes -subj "/CN=localhost"
    echo "✅ SSL certificates generated"
fi

# Generate JWT keys if they don't exist
if [ ! -f "secrets/jwt_private.pem" ]; then
    echo "Generating JWT keys..."
    mkdir -p secrets
    openssl genrsa -out secrets/jwt_private.pem 4096 > /dev/null 2>&1
    openssl rsa -in secrets/jwt_private.pem -pubout -out secrets/jwt_public.pem > /dev/null 2>&1
    echo "✅ JWT keys generated"
fi

echo "All credentials ready"
