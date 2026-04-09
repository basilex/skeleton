#!/bin/bash

# Build all Docker images for production
# Usage: ./scripts/docker/build-all.sh [tag]

set -e

TAG=${1:-latest}

echo "========================================="
echo "Building all Docker images for production"
echo "Tag: $TAG"
echo "========================================="

# Build backend
echo ""
echo "1/2 Building backend..."
./scripts/docker/build-backend.sh "$TAG"

# Build frontend
echo ""
echo "2/2 Building frontend..."
./scripts/docker/build-frontend.sh "$TAG"

echo ""
echo "========================================="
echo "✓ All images built successfully!"
echo "========================================="
echo ""
docker images | grep skeleton