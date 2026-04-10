FROM golang:1.25-alpine AS builder

WORKDIR /app

# Enable Go toolchain auto-download for required version
ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/keys ./keys
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./main"]