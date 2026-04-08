#!/bin/bash
# Clean Docker cache and rebuild

set -e

echo "🧹 Cleaning Docker cache..."
echo "================================"

# Stop all containers
echo "1. Stopping containers..."
docker-compose down --remove-orphans 2>/dev/null || true
docker-compose -f docker-compose.staging.yml down 2>/dev/null || true
docker-compose -f docker-compose.prod.yml down 2>/dev/null || true
docker-compose -f docker-compose.test.yml down 2>/dev/null || true

# Remove old images
echo "2. Removing old images..."
docker rmi skeleton-api 2>/dev/null || true
docker rmi skeleton-api-dev 2>/dev/null || true
docker rmi skeleton-api-staging 2>/dev/null || true
docker rmi skeleton-api-prod 2>/dev/null || true

# Remove dangling images
echo "3. Removing dangling images..."
docker image prune -f

# Remove build cache
echo "4. Removing build cache..."
docker builder prune -f

# Remove volumes (optional - uncomment if needed)
# echo "5. Removing volumes..."
# docker volume prune -f

echo ""
echo "✅ Docker cache cleaned!"
echo ""
echo "🚀 To rebuild:"
echo "   make docker-dev    # Development environment"
echo "   make docker-staging # Staging environment"
echo "   make docker-prod    # Production environment"
echo ""
echo "📝 Note: First build will take 2-3 minutes"