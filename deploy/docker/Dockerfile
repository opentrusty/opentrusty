# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache make git

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o opentrusty ./cmd/server

# Stage 2: Runtime
FROM alpine:3.23

# Create a non-root user
RUN addgroup -S opentrusty && adduser -S opentrusty -G opentrusty

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/opentrusty .

# Ensure binary is executable and owned by non-root user
RUN chown -R opentrusty:opentrusty /app

USER opentrusty

# Expose the default port (8080)
EXPOSE 8080

# The app is configured via environment variables
ENTRYPOINT ["./opentrusty"]
