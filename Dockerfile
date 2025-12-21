# Build stage
FROM golang:1.21-alpine AS builder

RUN apk --no-cache add ca-certificates git

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o webhook ./cmd/webhook

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /workspace/webhook .

# Create directories for configuration
RUN mkdir -p /etc/webhook/certs /etc/webhook/casbin

# Run as non-root user
RUN addgroup -g 1000 webhook && \
    adduser -D -u 1000 -G webhook webhook && \
    chown -R webhook:webhook /app /etc/webhook

USER webhook

EXPOSE 8443

ENTRYPOINT ["/app/webhook"]
