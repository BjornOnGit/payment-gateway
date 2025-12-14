#!/bin/bash
set -e

# Add 2GB swap if not exists
if [ ! -f /swapfile ]; then
  echo "Creating 2GB swap file..."
  sudo fallocate -l 2G /swapfile
  sudo chmod 600 /swapfile
  sudo mkswap /swapfile
  sudo swapon /swapfile
  echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
  echo "✅ Swap added"
else
  echo "Swap already exists"
fi

# Verify
free -h
