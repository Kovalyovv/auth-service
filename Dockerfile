# auth-service/Dockerfile
# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the auth service binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/auth/main.go

# Stage 2: Create a minimal final image
FROM alpine:latest

# Copy the CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the built binary from the builder stage
COPY --from=builder /auth-service /auth-service

# Expose the HTTP and gRPC ports
EXPOSE 8001 50001

# Command to run the service
CMD ["/auth-service"]
