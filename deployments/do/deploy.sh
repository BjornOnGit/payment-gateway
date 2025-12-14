#!/bin/bash
set -e

APP_DIR=/srv/payment-gateway

if [ ! -d "$APP_DIR" ]; then
  sudo mkdir -p $APP_DIR
  sudo chown $USER:$USER $APP_DIR
fi

cd $APP_DIR

# Fix git ownership issue
git config --global --add safe.directory $APP_DIR

# No need for git pull since rsync already synced files
# git pull origin main

docker build -t payment-gateway:api -f deployments/Dockerfile .

docker compose \
  -f docker-compose.prod.yml \
  up -d --build

docker system prune -f
