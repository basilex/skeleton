# Development Dockerfile (optimized for development)
FROM golang:1.23-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make gcc musl-dev postgresql-client

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Development mode - use air for hot reload
CMD ["go", "run", "./cmd/api"]