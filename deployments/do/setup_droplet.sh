#!/bin/bash
set -e

# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
sudo apt install -y docker.io docker-compose git

# Enable Docker on boot
sudo systemctl enable docker
sudo systemctl start docker

# Add user to docker group
sudo usermod -aG docker $USER

echo "✅ Docker ready. Log out and log back in."
