FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY main.go ./

# Build
RUN go build -o kanm-now-playing main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/kanm-now-playing .

# Expose port
EXPOSE 8000

# Run
CMD ["./kanm-now-playing"]
