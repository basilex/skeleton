#!/bin/bash

# Deploy to production environment
# Usage: ./scripts/deploy/deploy-production.sh

set -e

echo "========================================="
echo "Deploying to PRODUCTION environment"
echo "========================================="
echo ""
echo "⚠️  WARNING: This will deploy to production!"
echo ""
read -p "Are you sure? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
  echo "Deployment cancelled."
  exit 0
fi

# Configuration
PROD_HOST="${PROD_HOST:-skeleton.local}"
PROD_USER="${PROD_USER:-deploy}"
PROD_PATH="${PROD_PATH:-/var/www/skeleton}"

# Tag version
VERSION=$(git rev-parse --short HEAD)
echo "Deploying version: $VERSION"

# Backup current deployment
echo "Creating backup..."
ssh ${PROD_USER}@${PROD_HOST} "cd ${PROD_PATH} && docker-compose -f docker-compose.prod.yml exec -T postgres pg_dump -U skeleton skeleton > backup_${VERSION}.sql"

# Pull latest images
echo "Pulling latest Docker images..."
ssh ${PROD_USER}@${PROD_HOST} "cd ${PROD_PATH} && docker-compose -f docker-compose.prod.yml pull"

# Run database migrations
echo "Running database migrations..."
ssh ${PROD_USER}@${PROD_HOST} "cd ${PROD_PATH} && docker-compose -f docker-compose.prod.yml run --rm backend migrate"

# Deploy services with rolling update
echo "Deploying backend..."
ssh ${PROD_USER}@${PROD_HOST} "cd ${PROD_PATH} && docker-compose -f docker-compose.prod.yml up -d --no-deps --build backend"

echo "Waiting for backend to be ready..."
sleep 30

# Health check backend
echo "Checking backend health..."
ssh ${PROD_USER}@${PROD_HOST} "curl -f http://localhost:8080/health" || {
  echo "Backend health check failed! Rolling back..."
  # Rollback logic here
  exit 1
}

# Deploy frontend with cache clear
echo "Deploying frontend..."
ssh ${PROD_USER}@${PROD_HOST} "cd ${PROD_PATH} && docker-compose -f docker-compose.prod.yml up -d --no-deps --build frontend"

echo "Waiting for frontend to be ready..."
sleep 15

# Health check frontend
echo "Checking frontend..."
ssh ${PROD_USER}@${PROD_HOST} "curl -f http://localhost:3000" || {
  echo "Frontend health check failed!"
  exit 1
}

# Verify deployment
echo ""
echo "========================================="
echo "✓ Production deployment completed!"
echo "========================================="
echo ""
echo "Version:  $VERSION"
echo "Backend:  http://${PROD_HOST}:8080"
echo "Frontend: http://${PROD_HOST}:3000"
echo ""
echo "To view logs:"
echo "  ssh ${PROD_USER}@${PROD_HOST} 'cd ${PROD_PATH} && docker-compose -f docker-compose.prod.yml logs -f'"
echo ""
echo "To rollback:"
echo "  ssh ${PROD_USER}@${PROD_HOST} 'cd ${PROD_PATH} && psql -U skeleton skeleton < backup_${VERSION}.sql'"