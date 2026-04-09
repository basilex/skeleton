#!/bin/bash

# Production build script for Frontend
# Usage: ./scripts/docker/build-frontend.sh [tag]

set -e

TAG=${1:-latest}
IMAGE_NAME="${DOCKER_USERNAME:-skeleton}/skeleton-frontend:${TAG}"

echo "Building frontend Docker image: $IMAGE_NAME"

# Install dependencies first
echo "Installing dependencies..."
cd frontend
npm ci

# Build frontend
docker build \
  -t "$IMAGE_NAME" \
  -f frontend/Dockerfile \
  frontend/

echo "✓ Frontend image built successfully: $IMAGE_NAME"

# Optional: Push to registry
if [ "$PUSH" = "true" ]; then
  echo "Pushing to registry..."
  docker push "$IMAGE_NAME"
  echo "✓ Image pushed to registry"
fi