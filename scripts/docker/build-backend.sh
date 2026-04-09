#!/bin/bash

# Production build script for Backend
# Usage: ./scripts/docker/build-backend.sh [tag]

set -e

TAG=${1:-latest}
IMAGE_NAME="${DOCKER_USERNAME:-skeleton}/skeleton-backend:${TAG}"

echo "Building backend Docker image: $IMAGE_NAME"

# Build backend
docker build \
  -t "$IMAGE_NAME" \
  -f backend/Dockerfile \
  --build-arg GO_VERSION=1.25 \
  backend/

echo "✓ Backend image built successfully: $IMAGE_NAME"

# Optional: Push to registry
if [ "$PUSH" = "true" ]; then
  echo "Pushing to registry..."
  docker push "$IMAGE_NAME"
  echo "✓ Image pushed to registry"
fi