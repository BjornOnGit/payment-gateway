#!/bin/bash

set -e

echo "Running migrations..."

for file in ./configs/migrations/*.sql; do
    echo "Applying migration: $file"
    # echo "DEBUG: psql -h localhost -p 5432 -U $POSTGRES_USER -d $POSTGRES_DB -f $file"
    PGPASSWORD=$POSTGRES_PASSWORD psql \
      -h localhost \
      -p 5432 \
      -U $POSTGRES_USER \
      -d $POSTGRES_DB \
      -f "$file"
done

echo "Migrations applied."
