#!/bin/bash

# nginx_setup.sh
# Interactive Nginx Setup Script for Finscope
# Installs Nginx, Configures Proxy, and sets up SSL (Certbot)

set -e

echo "🚀 Starting Nginx Setup..."

# 1. Install Nginx if not exists
if ! command -v nginx &> /dev/null; then
    echo "📦 Installing Nginx..."
    sudo apt-get update
    sudo apt-get install -y nginx certbot python3-certbot-nginx
    echo "✅ Nginx installed."
else
    echo "✅ Nginx is already installed."
fi

# 2. Collect Configuration Details
echo ""
echo "Please enter the domain configuration details:"
read -p "🌐 Enter Domain Name (e.g., example.com or api.example.com): " DOMAIN_NAME
read -p "🔌 Enter Backend Port (Default 8000 for API Gateway): " PORT
PORT=${PORT:-8000}

# 3. Create Nginx Configuration
CONFIG_FILE="/etc/nginx/sites-available/$DOMAIN_NAME"

echo "📝 Creating Nginx configuration at $CONFIG_FILE..."

# Detailed Nginx Config
sudo bash -c "cat > $CONFIG_FILE" <<EOF
server {
    server_name $DOMAIN_NAME;

    # Backend API Proxy
    location / {
        proxy_pass http://localhost:$PORT;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # Optional: Increase body size for upload endpoints
    client_max_body_size 10M;
}
EOF

# 4. Enable Site
echo "🔗 Enabling configuration..."
if [ -L "/etc/nginx/sites-enabled/$DOMAIN_NAME" ]; then
    echo "   Config already linked."
else
    sudo ln -s $CONFIG_FILE /etc/nginx/sites-enabled/
fi

# 5. Disable Default Config if exists (optional but recommended)
if [ -f "/etc/nginx/sites-enabled/default" ]; then
    read -p "Would you like to disable the default Nginx site? (y/n) " DISABLE_DEFAULT
    if [[ "$DISABLE_DEFAULT" =~ ^[Yy]$ ]]; then
        sudo rm /etc/nginx/sites-enabled/default
        echo "   Default site disabled."
    fi
fi

# 6. Test and Reload
echo "🧪 Testing Nginx configuration..."
sudo nginx -t
echo "🔄 Reloading Nginx..."
sudo systemctl reload nginx

# 7. SSL Setup (Certbot)
echo ""
read -p "🔒 Would you like to setup SSL using Certbot? (y/n) " SETUP_SSL
if [[ "$SETUP_SSL" =~ ^[Yy]$ ]]; then
    echo "Running Certbot... (Follow the instructions)"
    sudo certbot --nginx -d $DOMAIN_NAME
fi

echo "
🎉 Nginx Setup Complete!
---------------------------------------------
Domain: http://$DOMAIN_NAME (or https:// if SSL set)
Proxy To: localhost:$PORT
"
