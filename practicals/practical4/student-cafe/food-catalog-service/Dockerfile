# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
# Download all dependencies.
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o /food-catalog-service .

# Stage 2: Create a minimal final image
FROM alpine:latest

# We need consul binary for the health check
RUN apk --no-cache add curl

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /food-catalog-service /food-catalog-service

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["/food-catalog-service"]
