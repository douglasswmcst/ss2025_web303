# Practical 5B: Complete gRPC Migration - HTTP Gateway to Pure gRPC Backend

## Objective

Building on Practical 5A, this practical completes the gRPC migration journey by:
1. Converting the API Gateway from HTTP reverse proxy to gRPC client
2. Removing HTTP endpoints from backend services (simplification)
3. Implementing HTTP ↔ gRPC protocol translation
4. Achieving a pure gRPC internal architecture while maintaining HTTP compatibility for external clients
5. Understanding production-grade microservices communication patterns

## Why This Matters

### The Evolution Journey

**Practical 5**: Monolith → HTTP Microservices
**Practical 5A**: Added gRPC for inter-service calls (Hybrid: HTTP + gRPC)
**Practical 5B**: Pure gRPC backend with HTTP translation layer

This mirrors real-world migrations at companies like:
- **Google**: All internal services use gRPC
- **Netflix**: gRPC for high-performance internal communication
- **Uber**: Protocol translation layers for mobile apps

### Key Benefits

1. **Simplified Services**: Backend services do one thing (gRPC) instead of two (HTTP + gRPC)
2. **Type Safety**: All internal communication uses proto-defined contracts
3. **Performance**: Binary protocol throughout the system
4. **Maintainability**: Less code, clearer purpose
5. **Backwards Compatibility**: External clients still use familiar HTTP/REST

## Architecture Evolution

### Before (Practical 5A): Hybrid Approach

```
External Client (HTTP)
    ↓
API Gateway (HTTP Reverse Proxy)
    ↓ HTTP forwarding
Backend Services (Dual: HTTP + gRPC)
    ↓ gRPC for inter-service
Other Services
```

**Problems**:
- Backend services run two servers (HTTP + gRPC)
- Gateway just forwards HTTP
- More code to maintain
- Potential port conflicts

### After (Practical 5B): Pure gRPC Backend

```
External Client (HTTP/REST - unchanged)
    ↓
API Gateway (HTTP→gRPC Translation Layer)
    ↓ gRPC calls
Backend Services (Pure gRPC only)
    ↓ gRPC
Inter-service calls
```

**Benefits**:
- Services are simpler (single protocol)
- Gateway handles all HTTP concerns
- Type-safe internal communication
- Clear separation of concerns

## Implementation Walkthrough

### Phase 1: Transform API Gateway to gRPC Client

The gateway evolution:
- **Before**: 41 lines, simple HTTP reverse proxy
- **After**: ~150 lines, intelligent protocol translator

#### Step 1.1: Add Proto Dependencies

**File**: `api-gateway/go.mod`

```go
module api-gateway

go 1.23

require (
    github.com/douglasswm/student-cafe-protos v0.0.0
    github.com/go-chi/chi/v5 v5.0.11
    google.golang.org/grpc v1.59.0
    google.golang.org/protobuf v1.31.0
)

replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos
```

**Why**: Gateway needs proto definitions to act as a gRPC client.

**Key Insight**: The `replace` directive lets the gateway import the local proto module during development.

---

#### Step 1.2: Create gRPC Client Manager

**File**: `api-gateway/grpc/clients.go`

```go
package grpc

import (
    "fmt"
    "log"
    "os"

    menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
    orderv1 "github.com/douglasswm/student-cafe-protos/gen/go/order/v1"
    userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type ServiceClients struct {
    UserClient  userv1.UserServiceClient
    MenuClient  menuv1.MenuServiceClient
    OrderClient orderv1.OrderServiceClient
}

func NewServiceClients() (*ServiceClients, error) {
    userAddr := getEnv("USER_SERVICE_GRPC_ADDR", "user-service:9091")
    menuAddr := getEnv("MENU_SERVICE_GRPC_ADDR", "menu-service:9092")
    orderAddr := getEnv("ORDER_SERVICE_GRPC_ADDR", "order-service:9093")

    log.Printf("Connecting to User Service at %s", userAddr)
    userConn, err := grpc.NewClient(userAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, fmt.Errorf("failed to connect to user service: %w", err)
    }

    // Similar for menu and order...

    return &ServiceClients{
        UserClient:  userv1.NewUserServiceClient(userConn),
        MenuClient:  menuv1.NewMenuServiceClient(menuConn),
        OrderClient: orderv1.NewOrderServiceClient(orderConn),
    }, nil
}
```

**What This Does**:
1. Creates gRPC connections to all backend services
2. Uses DNS names from Docker Compose (e.g., `user-service:9091`)
3. Wraps clients in a struct for easy injection
4. Uses environment variables for configuration

**Production Note**: In production, use `credentials.NewClientTLSFromFile()` instead of `insecure.NewCredentials()`.

---

#### Step 1.3: Implement HTTP → gRPC Translation

This is the **heart** of Practical 5B. Each handler:
1. Parses incoming HTTP request
2. Calls gRPC service
3. Translates gRPC response to HTTP JSON

**File**: `api-gateway/handlers/user_handlers.go`

```go
package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "strconv"

    userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
    "github.com/go-chi/chi/v5"
)

// CreateUser translates HTTP POST to gRPC CreateUser call
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
    // 1. Parse HTTP JSON request
    var req struct {
        Name        string `json:"name"`
        Email       string `json:"email"`
        IsCafeOwner bool   `json:"is_cafe_owner"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    // 2. Call gRPC service
    resp, err := h.clients.UserClient.CreateUser(context.Background(), &userv1.CreateUserRequest{
        Name:        req.Name,
        Email:       req.Email,
        IsCafeOwner: req.IsCafeOwner,
    })

    if err != nil {
        handleGRPCError(w, err)  // Translate gRPC errors
        return
    }

    // 3. Return HTTP JSON response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp.User)
}
```

**Let's Trace a Request**:

1. **Client sends**: `curl -X POST http://localhost:8080/api/users -d '{"name":"Alice"}'`
2. **Gateway receives**: HTTP POST with JSON body
3. **Gateway parses**: JSON → Go struct
4. **Gateway calls**: `UserClient.CreateUser()` with proto message
5. **gRPC happens**: Binary protocol over TCP to user-service:9091
6. **User service responds**: Proto message with user data
7. **Gateway translates**: Proto → JSON
8. **Client receives**: HTTP 201 Created with JSON body

**Key Pattern**: Gateway acts as a **protocol adapter**.

---

#### Step 1.4: Error Code Translation

gRPC uses different status codes than HTTP. We need to map them:

**File**: `api-gateway/handlers/handlers.go`

```go
func handleGRPCError(w http.ResponseWriter, err error) {
    st, ok := status.FromError(err)
    if !ok {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    var httpStatus int
    switch st.Code() {
    case codes.NotFound:
        httpStatus = http.StatusNotFound  // 404
    case codes.InvalidArgument:
        httpStatus = http.StatusBadRequest  // 400
    case codes.AlreadyExists:
        httpStatus = http.StatusConflict  // 409
    case codes.PermissionDenied:
        httpStatus = http.StatusForbidden  // 403
    case codes.Unauthenticated:
        httpStatus = http.StatusUnauthorized  // 401
    default:
        httpStatus = http.StatusInternalServerError  // 500
    }

    http.Error(w, st.Message(), httpStatus)
}
```

**Mapping Examples**:
| gRPC Code | HTTP Status | Use Case |
|-----------|-------------|----------|
| `codes.NotFound` | 404 | User/Order not found |
| `codes.InvalidArgument` | 400 | Invalid email format |
| `codes.AlreadyExists` | 409 | Email already registered |
| `codes.Unavailable` | 503 | Service down |

---

#### Step 1.5: Update Gateway Main

**File**: `api-gateway/main.go`

**Before (Reverse Proxy)**:
```go
func main() {
    r := chi.NewRouter()
    r.HandleFunc("/api/users*", proxyTo("http://user-service:8081"))
    http.ListenAndServe(":8080", r)
}
```

**After (gRPC Client)**:
```go
func main() {
    // Initialize gRPC clients
    clients, err := grpc.NewServiceClients()
    if err != nil {
        log.Fatalf("Failed to create gRPC clients: %v", err)
    }
    log.Println("gRPC clients initialized successfully")

    // Create handlers
    h := handlers.NewHandlers(clients)

    // Setup routes
    r := chi.NewRouter()
    r.Use(middleware.Logger)

    // User routes
    r.Post("/api/users", h.CreateUser)
    r.Get("/api/users/{id}", h.GetUser)
    r.Get("/api/users", h.GetUsers)

    // Menu routes
    r.Post("/api/menu", h.CreateMenuItem)
    r.Get("/api/menu/{id}", h.GetMenuItem)
    r.Get("/api/menu", h.GetMenu)

    // Order routes
    r.Post("/api/orders", h.CreateOrder)
    r.Get("/api/orders/{id}", h.GetOrder)
    r.Get("/api/orders", h.GetOrders)

    log.Println("API Gateway starting on :8080 (HTTP→gRPC translation layer)")
    http.ListenAndServe(":8080", r)
}
```

**Key Differences**:
- **Before**: Simple proxying, no business logic
- **After**: Explicit routing, protocol translation, error handling

---

### Phase 2: Simplify Backend Services

Now that the gateway handles HTTP, backend services can focus on gRPC only.

#### Step 2.1: Simplify User Service

**File**: `user-service/main.go`

**Before (75 lines, dual servers)**:
```go
func main() {
    database.Connect(dsn)
    go startGRPCServer()  // Background
    startHTTPServer()     // Blocking
}

func startHTTPServer() {
    r := chi.NewRouter()
    r.Post("/users", handlers.CreateUser)
    r.Get("/users/{id}", handlers.GetUser)
    http.ListenAndServe(":8081", r)
}

func startGRPCServer() {
    lis, _ := net.Listen("tcp", ":9091")
    s := grpc.NewServer()
    userv1.RegisterUserServiceServer(s, grpcserver.NewUserServer())
    s.Serve(lis)
}
```

**After (46 lines, gRPC only)**:
```go
func main() {
    // Connect to database
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "host=localhost user=postgres password=postgres dbname=user_db port=5432 sslmode=disable"
    }
    database.Connect(dsn)

    // Get gRPC port
    grpcPort := os.Getenv("GRPC_PORT")
    if grpcPort == "" {
        grpcPort = "9091"
    }

    // Start listening
    lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    // Register gRPC server
    s := grpc.NewServer()
    userv1.RegisterUserServiceServer(s, grpcserver.NewUserServer())

    log.Printf("User service (gRPC only) starting on :%s", grpcPort)
    s.Serve(lis)
}
```

**Code Reduction**: 75 lines → 46 lines (39% reduction)

**Removed**:
- ❌ HTTP server setup
- ❌ Chi router initialization
- ❌ HTTP handlers directory
- ❌ Goroutine complexity
- ❌ HTTP_PORT configuration

**Result**: Single-purpose service with clear responsibility.

---

#### Step 2.2: Simplify Menu & Order Services

Apply the same pattern:
- Remove `startHTTPServer()` function
- Remove HTTP handlers
- Keep only gRPC server code
- Update Dockerfiles to expose only gRPC port

**Order Service Special Case**:
Order service still needs gRPC **clients** (to call user/menu services), but no longer needs HTTP handlers.

---

#### Step 2.3: Update Docker Configuration

**File**: `docker-compose.yml`

**Before**:
```yaml
user-service:
  ports:
    - "8081:8081"  # HTTP
    - "9091:9091"  # gRPC
  environment:
    HTTP_PORT: "8081"
    GRPC_PORT: "9091"
```

**After**:
```yaml
user-service:
  ports:
    - "9091:9091"  # gRPC only
  environment:
    GRPC_PORT: "9091"

api-gateway:
  ports:
    - "8080:8080"  # HTTP for clients
  environment:
    USER_SERVICE_GRPC_ADDR: "user-service:9091"
    MENU_SERVICE_GRPC_ADDR: "menu-service:9092"
    ORDER_SERVICE_GRPC_ADDR: "order-service:9093"
```

**Key Changes**:
- Services expose only gRPC ports
- Gateway gets service addresses via environment variables
- Clear separation: External (HTTP) vs Internal (gRPC)

---

## Testing the New Architecture

### Test 1: Create a User

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com",
    "is_cafe_owner": false
  }'
```

**What Happens**:
1. Gateway receives HTTP POST
2. Parses JSON body
3. Calls `UserClient.CreateUser()` via gRPC
4. User service processes in pure gRPC
5. Gateway returns HTTP 201 Created

**Verify gRPC**:
```bash
docker-compose logs api-gateway | grep "gRPC clients initialized"
docker-compose logs user-service | grep "gRPC only"
```

---

### Test 2: Create an Order (End-to-End Flow)

```bash
# First create menu item and user
curl -X POST http://localhost:8080/api/menu \
  -H "Content-Type: application/json" \
  -d '{"name": "Coffee", "description": "Hot coffee", "price": 2.50}'

curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "bob@test.com", "is_cafe_owner": false}'

# Now create order
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "items": [{"menu_item_id": 1, "quantity": 2}]
  }'
```

**Complete gRPC Flow**:
1. Client → Gateway (HTTP)
2. Gateway → Order Service (gRPC)
3. Order Service → User Service (gRPC) to validate user
4. Order Service → Menu Service (gRPC) to get price
5. Order Service → Database (creates order)
6. Order Service → Gateway (gRPC response)
7. Gateway → Client (HTTP JSON)

**All internal communication is gRPC!**

---

### Test 3: Verify Services Are gRPC-Only

```bash
# These should FAIL (no HTTP servers):
curl http://localhost:9091  # gRPC port, not HTTP
curl http://localhost:9092
curl http://localhost:9093

# This should WORK (gateway HTTP):
curl http://localhost:8080/api/menu
```

**Expected**:
- Direct gRPC port access fails (not HTTP protocol)
- Gateway responds successfully

---

## Troubleshooting Guide

### Issue 1: Gateway Can't Connect to Services

**Symptom**:
```
Failed to connect to user service: connection refused
```

**Debug Steps**:
```bash
# 1. Check if services are running
docker-compose ps

# 2. Check service logs
docker-compose logs user-service | grep "gRPC server starting"

# 3. Verify network connectivity
docker-compose exec api-gateway ping user-service
```

**Common Causes**:
- Service hasn't started yet (add `depends_on` in docker-compose)
- Wrong service address (check `USER_SERVICE_GRPC_ADDR`)
- Service crashed on startup (check logs)

---

### Issue 2: Proto Import Errors in Gateway

**Symptom**:
```
could not import github.com/douglasswm/student-cafe-protos
```

**Solution**:
1. Ensure gateway Dockerfile copies proto module:
   ```dockerfile
   COPY ../student-cafe-protos /student-cafe-protos
   ```

2. Verify `go.mod` has replace directive:
   ```go
   replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos
   ```

3. Rebuild with no cache:
   ```bash
   docker-compose build --no-cache api-gateway
   ```

---

### Issue 3: HTTP Requests Return 500

**Symptom**: All requests return Internal Server Error

**Debug**:
```bash
docker-compose logs api-gateway
```

**Common Causes**:
- gRPC client not initialized properly
- Error in handler code
- Missing error handling in `handleGRPCError()`

**Fix**: Check gateway logs for specific error messages.

---

### Issue 4: Can Still Access HTTP Ports

**Symptom**: `curl http://localhost:8081` works (but shouldn't)

**Debug**: Check if you removed HTTP server code:
```bash
grep -r "startHTTPServer" user-service/main.go
```

**Solution**: Ensure `main.go` only starts gRPC server.

---

## Comparison: Practical 5A vs 5B

| Aspect | Practical 5A | Practical 5B |
|--------|--------------|--------------|
| **Gateway** | HTTP reverse proxy (41 lines) | gRPC client with translation (~150 lines) |
| **Backend Protocol** | Dual (HTTP + gRPC) | gRPC only |
| **Service Complexity** | ~75 lines (2 servers) | ~46 lines (1 server) |
| **External API** | HTTP/REST | HTTP/REST (same) |
| **Internal Communication** | gRPC | gRPC (same) |
| **Port Exposure** | 8081-8083 (HTTP), 9091-9093 (gRPC) | 9091-9093 (gRPC only) |
| **Use Case** | Gradual migration | Production architecture |

---

## Key Learnings

### 1. Protocol Translation Pattern

The gateway acts as a **protocol adapter** between HTTP (external) and gRPC (internal). This pattern:
- Maintains backwards compatibility
- Allows internal optimization
- Centralizes HTTP concerns

**Real World**: Google's API Gateway does this for all public APIs.

---

### 2. Service Simplification

Removing HTTP from backend services:
- **Reduces code**: ~40% fewer lines
- **Clearer purpose**: Single protocol, single responsibility
- **Easier testing**: Test gRPC server only

**Lesson**: Simplicity is a feature.

---

### 3. HTTP ↔ gRPC Error Mapping

Different protocols have different error models. Translating correctly is critical for client experience.

**Best Practice**: Create comprehensive mapping based on your API semantics.

---

### 4. Gradual Migration Path

The evolution shows how to migrate safely:
1. **Practical 5**: All HTTP (baseline)
2. **Practical 5A**: Add gRPC for inter-service (hybrid)
3. **Practical 5B**: Pure gRPC internal (final state)

Each step can be deployed and validated independently.

---

## Production Considerations

### Security

**Current**: `insecure.NewCredentials()`
**Production**: Use TLS

```go
creds, err := credentials.NewClientTLSFromFile("ca.pem", "")
conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
```

---

### Observability

Add logging and metrics:
```go
import (
    grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
    grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

s := grpc.NewServer(
    grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
        grpc_prometheus.UnaryServerInterceptor,
        loggingInterceptor,
    )),
)
```

---

### Load Balancing

For multiple gateway instances, use a load balancer:
```yaml
nginx:
  image: nginx
  ports:
    - "80:80"
  depends_on:
    - api-gateway-1
    - api-gateway-2
```

---

## Reflection Questions

1. **Why use gRPC for internal communication but HTTP for external?**
   - Hint: Consider client capabilities and backwards compatibility

2. **What are the trade-offs of the gateway pattern?**
   - Hint: Single point of failure vs. centralized logic

3. **How would you version the gateway as proto definitions evolve?**
   - Hint: Think about API versioning strategies

4. **When would you choose HTTP-only microservices instead?**
   - Hint: Consider team expertise and tooling ecosystem

5. **How does this architecture enable polyglot microservices?**
   - Hint: Proto definitions work across languages

---

## Next Steps

### For Learning
1. Add gRPC streaming for real-time order updates
2. Implement authentication using gRPC metadata
3. Add circuit breakers for service failures
4. Deploy to Kubernetes with service mesh

### For Production
1. Add comprehensive logging
2. Implement health checks
3. Set up Prometheus metrics
4. Add distributed tracing (OpenTelemetry)
5. Enable TLS for all gRPC connections

---

## Submission Requirements

### Required Deliverables

1. **Source Code**:
   - Updated api-gateway with gRPC clients
   - Simplified service main.go files
   - Updated docker-compose.yml
   - All handler files

2. **Documentation**:
   - This README with your observations
   - Screenshots showing:
     - `docker-compose ps` (only gRPC ports for services)
     - Successful order creation via HTTP
     - Gateway logs showing gRPC client initialization
     - Failed attempt to curl gRPC ports directly

3. **Reflection Essay (750 words minimum)**:
   - Explain the HTTP → gRPC translation process
   - Compare service complexity before/after
   - Discuss when to use this architecture pattern
   - Analyze the trade-offs of protocol translation
   - How would you extend this for real-time features?

### Grading Criteria

| Criteria | Weight |
|----------|--------|
| Gateway gRPC client implementation | 30% |
| HTTP → gRPC translation handlers | 25% |
| Service simplification (removed HTTP) | 20% |
| Docker configuration correctness | 10% |
| Documentation and reflection | 15% |

---

## Conclusion

Congratulations! You've completed the full microservices gRPC migration:

**Practical 5**: Monolith → HTTP Microservices
**Practical 5A**: Added gRPC inter-service communication
**Practical 5B**: Pure gRPC backend with HTTP gateway

You now understand:
- Protocol translation patterns
- gRPC client/server architecture
- Service simplification strategies
- Production-grade microservices communication

This pattern is used by **Google, Netflix, Uber, Slack** and countless other companies at scale.

**You're now equipped to build production-ready gRPC microservices!**

---

## Additional Resources

- [gRPC Gateway Pattern](https://grpc-ecosystem.github.io/grpc-gateway/)
- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)
- [Protocol Translation at Google](https://cloud.google.com/endpoints/docs/grpc/transcoding)
- [Microservices Communication Patterns](https://microservices.io/patterns/communication-style/messaging.html)
- [Production gRPC at Scale](https://www.youtube.com/watch?v=Z_yD7YPL2oE) (Netflix talk)
