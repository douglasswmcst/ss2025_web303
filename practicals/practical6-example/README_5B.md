# Practical 5B: Pure gRPC Backend with HTTP Gateway

## Quick Start Guide

This directory contains the complete implementation of Practical 5B, building on Practical 5A to achieve a pure gRPC internal architecture.

### What Changed from 5A → 5B

**API Gateway**:
- ❌ Removed: HTTP reverse proxy
- ✅ Added: gRPC clients + HTTP→gRPC translation

**Backend Services** (user, menu, order):
- ❌ Removed: HTTP servers and handlers
- ✅ Kept: gRPC servers only
- 📉 Code reduction: ~40% fewer lines per service

**Architecture**:
```
External: HTTP/REST (unchanged for clients)
    ↓
Gateway: HTTP→gRPC translation layer
    ↓
Internal: Pure gRPC communication
```

---

## Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local development)
- Protocol Buffer compiler (protoc)

---

## Deployment

### Option 1: Automated (Recommended)

```bash
cd practicals/practical5a
./deploy_5b.sh
```

### Option 2: Manual

```bash
# 1. Generate proto code
cd student-cafe-protos
make generate
cd ..

# 2. Build and start
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d

# 3. Wait for services
sleep 10

# 4. Check status
docker-compose ps
```

---

## Testing

### Create Test Data

```bash
# Create menu item
curl -X POST http://localhost:8080/api/menu \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Cappuccino",
    "description": "Italian coffee with steamed milk",
    "price": 3.50
  }'

# Create user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com",
    "is_cafe_owner": false
  }'

# Create order (triggers full gRPC flow)
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "items": [{"menu_item_id": 1, "quantity": 2}]
  }'
```

### Verify Architecture

```bash
# Check gateway initialized gRPC clients
docker-compose logs api-gateway | grep "gRPC clients initialized"

# Check services are gRPC-only
docker-compose logs user-service | grep "gRPC only"
docker-compose logs menu-service | grep "gRPC only"
docker-compose logs order-service | grep "gRPC only"

# Verify HTTP ports are closed (these should FAIL)
curl http://localhost:9091  # gRPC port, not HTTP
curl http://localhost:9092
curl http://localhost:9093

# Verify gateway HTTP works (should SUCCESS)
curl http://localhost:8080/api/menu
```

---

## Service Ports

### External (HTTP)
- **API Gateway**: http://localhost:8080

### Internal (gRPC - not directly accessible via HTTP)
- **User Service**: localhost:9091
- **Menu Service**: localhost:9092
- **Order Service**: localhost:9093

### Databases
- **User DB**: localhost:5434
- **Menu DB**: localhost:5433
- **Order DB**: localhost:5435

---

## Architecture Details

### API Gateway (HTTP→gRPC Translation)

**Responsibilities**:
- Accept HTTP requests from external clients
- Parse JSON payloads
- Call appropriate gRPC service methods
- Translate gRPC responses to HTTP JSON
- Map gRPC status codes to HTTP status codes

**Files**:
- `api-gateway/grpc/clients.go` - gRPC client manager
- `api-gateway/handlers/*.go` - HTTP→gRPC handlers
- `api-gateway/main.go` - Server setup

### Backend Services (Pure gRPC)

**User Service** (`user-service:9091`):
- CreateUser, GetUser, GetUsers

**Menu Service** (`menu-service:9092`):
- CreateMenuItem, GetMenuItem, GetMenu

**Order Service** (`order-service:9093`):
- CreateOrder, GetOrder, GetOrders
- Has gRPC clients to User and Menu services

---

## Request Flow Example

**Client creates an order**:

1. `curl -X POST http://localhost:8080/api/orders ...` (HTTP)
2. Gateway receives HTTP, parses JSON
3. Gateway → Order Service via gRPC: `CreateOrder()`
4. Order Service → User Service via gRPC: `GetUser()` (validate)
5. Order Service → Menu Service via gRPC: `GetMenuItem()` (get price)
6. Order Service creates order in database
7. Order Service → Gateway via gRPC: response
8. Gateway → Client via HTTP: JSON response

**All internal calls use gRPC!**

---

## Troubleshooting

### Gateway Can't Connect to Services

**Error**: `Failed to connect to user service`

**Fix**:
```bash
# Check services are running
docker-compose ps

# Check logs
docker-compose logs user-service

# Restart
docker-compose restart api-gateway
```

### Proto Import Errors

**Error**: `could not import student-cafe-protos`

**Fix**:
```bash
# Regenerate proto code
cd student-cafe-protos
make clean && make generate
cd ..

# Rebuild without cache
docker-compose build --no-cache
```

### All Requests Return 500

**Debug**:
```bash
docker-compose logs api-gateway
```

**Common causes**:
- gRPC clients not initialized
- Service address misconfiguration
- Proto version mismatch

---

## Code Comparison

### Service Complexity

| Service | Practical 5A | Practical 5B | Reduction |
|---------|--------------|--------------|-----------|
| main.go lines | ~75 (dual) | ~46 (gRPC) | 39% |
| Servers | 2 (HTTP+gRPC) | 1 (gRPC) | 50% |
| Ports | 2 | 1 | 50% |

### Gateway Complexity

| Metric | Practical 5A | Practical 5B | Change |
|--------|--------------|--------------|--------|
| Lines of code | 41 | ~150 | +267% |
| Responsibility | Proxy | Translation | More complex |
| Business logic | None | Protocol adapter | Added |

**Insight**: Complexity moved to the gateway (good) - services are simpler.

---

## File Structure

```
practical5a/
├── student-cafe-protos/       # Shared proto repository
├── api-gateway/               # HTTP→gRPC translation
│   ├── grpc/
│   │   └── clients.go        # gRPC client manager
│   ├── handlers/
│   │   ├── handlers.go       # Error mapping
│   │   ├── user_handlers.go  # User translation
│   │   ├── menu_handlers.go  # Menu translation
│   │   └── order_handlers.go # Order translation
│   ├── main.go               # Gateway server
│   └── Dockerfile
├── user-service/             # gRPC-only
│   ├── grpc/server.go
│   ├── main.go              # Simplified!
│   └── Dockerfile
├── menu-service/            # gRPC-only
│   ├── grpc/server.go
│   ├── main.go             # Simplified!
│   └── Dockerfile
├── order-service/          # gRPC-only + clients
│   ├── grpc/
│   │   ├── server.go
│   │   └── clients.go     # Calls user/menu
│   ├── main.go           # Simplified!
│   └── Dockerfile
├── docker-compose.yml     # Updated for gRPC-only
└── deploy_5b.sh          # Deployment script
```

---

## Key Learnings

1. **Protocol Translation**: Gateway bridges HTTP (external) and gRPC (internal)
2. **Service Simplification**: Single protocol = simpler services
3. **Error Mapping**: gRPC codes → HTTP status codes
4. **Backwards Compatibility**: External API unchanged (HTTP/REST)
5. **Production Pattern**: Used by Google, Netflix, Uber

---

## Next Steps

### To Learn More
- Read `../practical5b.md` for detailed walkthrough
- Experiment with gRPC streaming
- Add authentication using gRPC metadata
- Implement circuit breakers

### To Deploy to Production
- Add TLS for gRPC connections
- Implement health checks
- Set up Prometheus metrics
- Add distributed tracing
- Deploy to Kubernetes

---

## Resources

- [Complete Walkthrough](../practical5b.md)
- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://protobuf.dev/)
- [gRPC Gateway Pattern](https://grpc-ecosystem.github.io/grpc-gateway/)

---

## Support

For issues or questions:
1. Check `../practical5b.md` troubleshooting section
2. Review service logs: `docker-compose logs <service-name>`
3. Verify proto generation: `cd student-cafe-protos && make generate`
4. Rebuild from scratch: `docker-compose down -v && docker-compose build --no-cache`

---

**Good luck!** You're now building production-grade gRPC microservices! 🚀
