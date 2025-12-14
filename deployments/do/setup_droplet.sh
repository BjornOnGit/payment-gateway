#!/bin/bash
set -e

# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
sudo apt install -y docker.io git

# Enable Docker on boot
sudo systemctl enable docker
sudo systemctl start docker

# Add user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose v2 standalone (replaces v1)
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Create app directory with proper structure
mkdir -p /srv/payment-gateway/{dev-certs,secrets}

echo "✅ Docker and Docker Compose v2 ready. Log out and log back in."
