# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service .

# Stage 2: Final image
FROM alpine:latest
WORKDIR /
COPY --from=builder /order-service /order-service
EXPOSE 8081
CMD ["/order-service"]
