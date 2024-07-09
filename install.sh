#!/bin/bash

set -e

if [ -z "$1" ]; then
  echo "Usage: $0 DOMAIN_NAME"
  exit 1
fi

DOMAIN_NAME=$1

# Create a production nginx configuration file
cat <<EOL > nginx_prod.conf
events {}

http {
    server {
        listen 80;
        server_name $DOMAIN_NAME;

        location /r/ {
            proxy_pass http://app:8000;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }

        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        listen 443 ssl; # managed by Certbot
        ssl_certificate /etc/letsencrypt/live/$DOMAIN_NAME/fullchain.pem; # managed by Certbot
        ssl_certificate_key /etc/letsencrypt/live/$DOMAIN_NAME/privkey.pem; # managed by Certbot
        include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
        ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot
    }
}
EOL

# Install Certbot
sudo apt-get update
sudo apt-get install -y certbot

# Download options-ssl-nginx.conf and ssl-dhparams.pem
sudo mkdir -p /etc/letsencrypt
sudo wget https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem -O /etc/letsencrypt/ssl-dhparams.pem
sudo wget https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/options-ssl-nginx.conf -O /etc/letsencrypt/options-ssl-nginx.conf

# Obtain SSL certificate
sudo certbot certonly --standalone -d $DOMAIN_NAME

# Set up Docker containers
docker compose down
docker compose up --build -d

# Restart Nginx container to apply SSL certificates
docker compose restart nginx

echo "Setup complete. Your URL shortener is available at https://$DOMAIN_NAME"
