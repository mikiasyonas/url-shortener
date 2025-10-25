#!/bin/bash

echo "Production Deployment Process"
echo "======================================"

# 2. Check Docker Compose
echo ""
echo "1. 🔍 Checking Docker Compose:"
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
    echo "✅ Using docker-compose"
elif docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
    echo "✅ Using docker compose"
else
    echo "Docker Compose not found"
    echo "Install from: https://docs.docker.com/compose/install/"
    exit 1
fi

# 3. Build and deploy
echo ""
echo "2. Building and Deploying:"
$DOCKER_COMPOSE_CMD -f deployments/docker-compose.prod.yml --env-file .env.production up -d --build

# 4. Wait for services
echo ""
echo "3. Waiting for services to start..."
sleep 25

# 5. Check status
echo ""
echo "4. Checking Service Status:"
$DOCKER_COMPOSE_CMD -f deployments/docker-compose.prod.yml ps

echo "🗄️ Running database migrations..."
atlas migrate apply --env gorm

# 6. Test the application
echo ""
echo "5. Testing Application:"
echo "Health check:"
curl -s http://localhost/health || echo "Health check failed"

echo ""
echo "Testing URL shortening:"
SHORT_URL=$(curl -s -X POST http://localhost/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}' | grep -o '"short_url":"[^"]*"' | cut -d'"' -f4)

if [ -n "$SHORT_URL" ]; then
    echo "Short URL created: $SHORT_URL"
    echo "Testing redirect:"
    curl -I -s $SHORT_URL | head -n 1
else
    echo "URL shortening failed"
fi

echo ""
echo "Deployement completed!"
echo "Access your URL shortener at: http://localhost"