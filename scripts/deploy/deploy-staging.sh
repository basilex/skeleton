#!/bin/bash

# Deploy to staging environment
# Usage: ./scripts/deploy/deploy-staging.sh

set -e

echo "========================================="
echo "Deploying to STAGING environment"
echo "========================================="

# Configuration
STAGING_HOST="${STAGING_HOST:-staging.skeleton.local}"
STAGING_USER="${STAGING_USER:-deploy}"
STAGING_PATH="${STAGING_PATH:-/var/www/skeleton-staging}"

# Pull latest images
echo "Pulling latest Docker images..."
docker-compose -f docker-compose.staging.yml pull

# Run database migrations
echo "Running database migrations..."
docker-compose -f docker-compose.staging.yml run --rm backend migrate

# Deploy services
echo "Deploying services..."
docker-compose -f docker-compose.staging.yml up -d --no-deps --build backend frontend

# Health check
echo "Waiting for services to start..."
sleep 10

# Check backend health
echo "Checking backend health..."
curl -f http://${STAGING_HOST}:8080/health || {
  echo "Backend health check failed!"
  exit 1
}

# Check frontend
echo "Checking frontend..."
curl -f http://${STAGING_HOST}:3000 || {
  echo "Frontend health check failed!"
  exit 1
}

echo ""
echo "========================================="
echo "✓ Staging deployment completed!"
echo "========================================="
echo ""
echo "Backend:  http://${STAGING_HOST}:8080"
echo "Frontend: http://${STAGING_HOST}:3000"
echo ""
echo "To view logs:"
echo "  docker-compose -f docker-compose.staging.yml logs -f"