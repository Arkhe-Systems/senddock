#!/bin/sh
set -e

if [ -f .env ]; then
    echo "Error: .env already exists. Delete it first if you want to re-run setup."
    exit 1
fi

JWT_SECRET=$(openssl rand -hex 32)
POSTGRES_PASSWORD=$(openssl rand -hex 16)

cat > .env <<EOF
JWT_SECRET=${JWT_SECRET}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
FRONTEND_URL=http://localhost:8080
DEPLOYMENT_MODE=self-hosted
EOF

echo "SendDock setup complete."
echo ""
echo "Starting services..."
docker compose -f docker-compose.prod.yml up -d

echo ""
echo "SendDock is running at http://localhost:8080"
echo "Open it in your browser to create your admin account."
