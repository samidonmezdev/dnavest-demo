#!/bin/bash

# Server Setup Script for Finscope (Hetzner/Ubuntu)
# Run this script on your server once to prepare it for CI/CD deployment.

set -e

echo "🚀 Starting Server Setup..."

# 1. Update System
echo "📦 Updating system packages..."
sudo apt-get update
sudo apt-get upgrade -y

# 2. Install Docker
if ! command -v docker &> /dev/null; then
    echo "🐳 Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
    echo "✅ Docker installed successfully."
else
    echo "✅ Docker is already installed."
fi

# 3. Install Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "🐙 Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    echo "✅ Docker Compose installed successfully."
else
    echo "✅ Docker Compose is already installed."
fi

# 4. Create Project Directory
echo "📂 Creating project directory..."
mkdir -p /root/dnavest-demo
echo "✅ Directory /root/dnavest-demo created."

# 5. Clean up
sudo apt-get autoremove -y

echo "
✨ Server Setup Complete!
---------------------------------------------
Next Steps:
1. Ensure your GitHub Secrets (HOST, USERNAME, SSH_KEY, GHCR_PAT) are set.
2. Push your code to the 'main' branch.
3. The CI/CD pipeline will automatically deploy to this server.
"
