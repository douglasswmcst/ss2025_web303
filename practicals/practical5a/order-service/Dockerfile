FROM golang:1.23-alpine AS builder
WORKDIR /build

# Copy proto module first (needed for go mod download)
COPY ../student-cafe-protos /student-cafe-protos

# Copy service files
WORKDIR /build/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service .

FROM alpine:latest
WORKDIR /
COPY --from=builder /order-service /order-service

# Expose only gRPC port
EXPOSE 9093

CMD ["/order-service"]
