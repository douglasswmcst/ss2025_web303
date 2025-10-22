# Practical 5B: Complete gRPC Migration - API Gateway Implementation Plan

**Date**: 2025-10-22
**Building On**: Practical 5A (gRPC inter-service communication)
**Goal**: Migrate API Gateway from HTTP reverse proxy to gRPC client architecture

---

## Overview

This practical completes the gRPC migration journey by converting the API Gateway from an HTTP reverse proxy to a gRPC client that calls backend services. External clients will continue to use familiar HTTP/REST, but all internal communication will be pure gRPC. Backend services will be simplified to expose only gRPC endpoints.

**Duration**: ~3-4 hours for students
**Difficulty**: Intermediate (building on Practical 5A knowledge)

---

## Current State Analysis (Practical 5A)

### Architecture
```
External Client (HTTP)
    ↓
API Gateway (HTTP Reverse Proxy)
    ↓ HTTP forwarding
Backend Services (Dual: HTTP + gRPC)
    ↓ gRPC
Inter-service calls
```

### Current API Gateway (`practical5a/api-gateway/main.go`)
- **Lines of Code**: 41 lines
- **Technology**: `httputil.ReverseProxy`
- **Function**: Simple HTTP forwarding:
  ```go
  r.HandleFunc("/api/users*", proxyTo("http://user-service:8081", "/users"))
  r.HandleFunc("/api/menu*", proxyTo("http://menu-service:8082", "/menu"))
  r.HandleFunc("/api/orders*", proxyTo("http://order-service:8083", "/orders"))
  ```

### Current Backend Services
- **User Service**: main.go (75 lines) - Runs both HTTP (8081) and gRPC (9091) servers
- **Menu Service**: main.go (74 lines) - Runs both HTTP (8082) and gRPC (9092) servers
- **Order Service**: main.go (98 lines) - Runs both HTTP (8083) and gRPC (9093) servers

Each service has:
```go
go startHTTPServer()   // REST endpoints for Gateway
startGRPCServer()      // gRPC for inter-service + future Gateway
```

### Key Discoveries
- Gateway currently has **no proto dependencies**
- Services already expose gRPC - just need to remove HTTP
- Proto module already contains all necessary service definitions
- Docker configuration supports dual protocols

---

## Desired End State (Practical 5B)

### New Architecture
```
External Client (HTTP/REST - unchanged)
    ↓
API Gateway (HTTP→gRPC Translation Layer)
    ↓ gRPC calls
Backend Services (Pure gRPC only)
    ↓ gRPC
Inter-service calls
```

### Benefits
1. **Simpler Backend Services**: Remove dual server complexity
2. **Pure gRPC Internal Architecture**: All service-to-service communication uses gRPC
3. **HTTP Compatibility**: External clients still use curl/browsers
4. **Type Safety**: Gateway uses proto-generated clients
5. **Performance**: Binary protocol for all internal calls

### What Changes
| Component | Current | New |
|-----------|---------|-----|
| API Gateway | HTTP reverse proxy | HTTP server + gRPC clients |
| Gateway Protocol | HTTP forwarding | HTTP → gRPC translation |
| Backend Services | Dual (HTTP + gRPC) | gRPC only |
| Service main.go | ~75 lines (dual servers) | ~40 lines (gRPC only) |
| External API | HTTP/REST | HTTP/REST (unchanged) |

---

## What We're NOT Doing

To keep scope focused and educational:

1. ❌ **gRPC-Web**: Not implementing browser-native gRPC (would require special client libraries)
2. ❌ **Load Balancing**: Single gateway instance (can add later with Kubernetes)
3. ❌ **Authentication/Authorization**: Keeping focus on gRPC migration
4. ❌ **Rate Limiting**: Not adding new features, just migrating
5. ❌ **Streaming**: Sticking with unary RPCs for simplicity
6. ❌ **TLS**: Using insecure credentials for local development

---

## Implementation Approach

### Strategy: Gateway as Protocol Translator

The new gateway will:
1. **Accept HTTP** requests from external clients (curl, browsers)
2. **Parse** HTTP request (path, method, body)
3. **Call** appropriate gRPC service method
4. **Translate** gRPC response to HTTP JSON
5. **Return** HTTP response to client

This is a common pattern called the **gRPC Gateway Pattern** or **Protocol Translation Layer**.

### Why This Approach

**Educational Value**:
- Students learn how to build gRPC clients
- Understand HTTP ↔ gRPC translation
- See complete migration path (HTTP → Hybrid → Pure gRPC)
- Can still use curl for testing (familiar)

**Practical Value**:
- Real companies use this pattern (Twitch, Slack, etc.)
- Allows gradual client migration
- Backwards compatible with existing clients
- Simpler than gRPC-Web for learning

---

## Phase 1: Update API Gateway to Use gRPC Clients

### Overview
Convert the gateway from HTTP reverse proxy to gRPC client that translates HTTP requests into gRPC calls.

### Changes Required

#### 1. Update `api-gateway/go.mod`

**File**: `practical5a/api-gateway/go.mod`
**Changes**: Add proto module and gRPC dependencies

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

**Why**: Gateway needs to import proto definitions and gRPC libraries to act as a client.

---

#### 2. Create gRPC Client Manager

**File**: `practical5a/api-gateway/grpc/clients.go` (NEW)
**Changes**: Create a centralized client manager

```go
package grpc

import (
    "fmt"
    "os"

    userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
    menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
    orderv1 "github.com/douglasswm/student-cafe-protos/gen/go/order/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// ServiceClients holds all gRPC clients
type ServiceClients struct {
    UserClient  userv1.UserServiceClient
    MenuClient  menuv1.MenuServiceClient
    OrderClient orderv1.OrderServiceClient
}

// NewServiceClients creates gRPC clients for all backend services
func NewServiceClients() (*ServiceClients, error) {
    // Get service addresses from environment or use defaults
    userAddr := getEnv("USER_SERVICE_GRPC_ADDR", "user-service:9091")
    menuAddr := getEnv("MENU_SERVICE_GRPC_ADDR", "menu-service:9092")
    orderAddr := getEnv("ORDER_SERVICE_GRPC_ADDR", "order-service:9093")

    // Create gRPC connections
    userConn, err := grpc.NewClient(userAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, fmt.Errorf("failed to connect to user service: %w", err)
    }

    menuConn, err := grpc.NewClient(menuAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, fmt.Errorf("failed to connect to menu service: %w", err)
    }

    orderConn, err := grpc.NewClient(orderAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, fmt.Errorf("failed to connect to order service: %w", err)
    }

    return &ServiceClients{
        UserClient:  userv1.NewUserServiceClient(userConn),
        MenuClient:  menuv1.NewMenuServiceClient(menuConn),
        OrderClient: orderv1.NewOrderServiceClient(orderConn),
    }, nil
}

func getEnv(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}
```

**Why**: Centralized client management makes it easy to inject clients into handlers and manage connections.

---

#### 3. Create HTTP→gRPC Translation Handlers

**File**: `practical5a/api-gateway/handlers/user_handlers.go` (NEW)
**Changes**: Translate HTTP user requests to gRPC calls

```go
package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "strconv"

    userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
    "github.com/go-chi/chi/v5"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// CreateUser handles POST /api/users
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name        string `json:"name"`
        Email       string `json:"email"`
        IsCafeOwner bool   `json:"is_cafe_owner"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Call gRPC service
    resp, err := h.clients.UserClient.CreateUser(context.Background(), &userv1.CreateUserRequest{
        Name:        req.Name,
        Email:       req.Email,
        IsCafeOwner: req.IsCafeOwner,
    })

    if err != nil {
        handleGRPCError(w, err)
        return
    }

    // Return HTTP JSON response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp.User)
}

// GetUser handles GET /api/users/{id}
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "invalid user ID", http.StatusBadRequest)
        return
    }

    resp, err := h.clients.UserClient.GetUser(context.Background(), &userv1.GetUserRequest{
        Id: uint32(id),
    })

    if err != nil {
        handleGRPCError(w, err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp.User)
}

// GetUsers handles GET /api/users
func (h *Handlers) GetUsers(w http.ResponseWriter, r *http.Request) {
    resp, err := h.clients.UserClient.GetUsers(context.Background(), &userv1.GetUsersRequest{})

    if err != nil {
        handleGRPCError(w, err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp.Users)
}
```

**File**: `practical5a/api-gateway/handlers/menu_handlers.go` (NEW)
Similar pattern for menu endpoints.

**File**: `practical5a/api-gateway/handlers/order_handlers.go` (NEW)
Similar pattern for order endpoints.

**File**: `practical5a/api-gateway/handlers/handlers.go` (NEW)
**Changes**: Base handler struct with shared error handling

```go
package handlers

import (
    "net/http"
    "api-gateway/grpc"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type Handlers struct {
    clients *grpc.ServiceClients
}

func NewHandlers(clients *grpc.ServiceClients) *Handlers {
    return &Handlers{clients: clients}
}

// handleGRPCError converts gRPC errors to appropriate HTTP status codes
func handleGRPCError(w http.ResponseWriter, err error) {
    st, ok := status.FromError(err)
    if !ok {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    var httpStatus int
    switch st.Code() {
    case codes.NotFound:
        httpStatus = http.StatusNotFound
    case codes.InvalidArgument:
        httpStatus = http.StatusBadRequest
    case codes.AlreadyExists:
        httpStatus = http.StatusConflict
    case codes.PermissionDenied:
        httpStatus = http.StatusForbidden
    case codes.Unauthenticated:
        httpStatus = http.StatusUnauthorized
    default:
        httpStatus = http.StatusInternalServerError
    }

    http.Error(w, st.Message(), httpStatus)
}
```

**Why**: This pattern provides:
- Clean separation of concerns
- Consistent error handling
- Easy to test
- Clear HTTP → gRPC mapping

---

#### 4. Update `api-gateway/main.go`

**File**: `practical5a/api-gateway/main.go`
**Changes**: Replace reverse proxy with gRPC client handlers

```go
package main

import (
    "log"
    "net/http"

    "api-gateway/grpc"
    "api-gateway/handlers"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    // Initialize gRPC clients
    clients, err := grpc.NewServiceClients()
    if err != nil {
        log.Fatalf("Failed to create gRPC clients: %v", err)
    }
    log.Println("gRPC clients initialized successfully")

    // Create handlers with clients
    h := handlers.NewHandlers(clients)

    // Setup HTTP router
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

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

    log.Println("API Gateway starting on :8080 (HTTP→gRPC translation)")
    http.ListenAndServe(":8080", r)
}
```

**Before**: 41 lines, simple reverse proxy
**After**: ~50 lines, gRPC client gateway

**Key Difference**:
- **Before**: `httputil.ReverseProxy` forwarded HTTP → HTTP
- **After**: Explicit handlers translate HTTP → gRPC

---

#### 5. Update `api-gateway/Dockerfile`

**File**: `practical5a/api-gateway/Dockerfile`
**Changes**: Add proto module copy

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /build

# Copy proto module first
COPY ../student-cafe-protos /student-cafe-protos

# Copy gateway files
WORKDIR /build/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway .

FROM alpine:latest
WORKDIR /
COPY --from=builder /api-gateway /api-gateway
EXPOSE 8080
CMD ["/api-gateway"]
```

**Why**: Gateway now needs access to proto module for generated Go code.

---

### Success Criteria - Phase 1

#### Automated Verification:
- [ ] Gateway builds successfully: `cd api-gateway && go build`
- [ ] Proto imports resolve: `go mod verify`
- [ ] No compilation errors
- [ ] Docker image builds: `docker build -t api-gateway:grpc ./api-gateway`

#### Manual Verification:
- [ ] Gateway starts and logs "gRPC clients initialized successfully"
- [ ] Can create a user via: `curl -X POST http://localhost:8080/api/users ...`
- [ ] Can fetch menu via: `curl http://localhost:8080/api/menu`
- [ ] Can create order via: `curl -X POST http://localhost:8080/api/orders ...`
- [ ] Error responses are proper HTTP status codes
- [ ] Gateway logs show gRPC calls being made

---

## Phase 2: Simplify Backend Services (Remove HTTP)

### Overview
Now that the gateway uses gRPC, backend services no longer need HTTP endpoints. We'll simplify them to pure gRPC servers.

### Changes Required

#### 1. Simplify `user-service/main.go`

**File**: `practical5a/user-service/main.go`
**Changes**: Remove HTTP server, keep only gRPC

**Before** (75 lines with dual servers):
```go
func main() {
    // ... database connection ...
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
    // gRPC setup
}
```

**After** (~40 lines, gRPC only):
```go
func main() {
    // Connect to database
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "host=localhost user=postgres password=postgres dbname=user_db port=5432 sslmode=disable"
    }

    if err := database.Connect(dsn); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Start gRPC server
    grpcPort := os.Getenv("GRPC_PORT")
    if grpcPort == "" {
        grpcPort = "9091"
    }

    lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    s := grpc.NewServer()
    userv1.RegisterUserServiceServer(s, grpcserver.NewUserServer())

    log.Printf("User service (gRPC only) starting on :%s", grpcPort)
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
```

**Removals**:
- Delete `startHTTPServer()` function
- Delete `startGRPCServer()` function
- Delete `handlers/` directory (no longer needed)
- Simplify `go.mod` (remove chi dependency)

**Result**: ~45% code reduction, clearer purpose

---

#### 2. Simplify `menu-service/main.go`

**File**: `practical5a/menu-service/main.go`
**Changes**: Same pattern as user-service

- Remove HTTP server code
- Keep only gRPC server
- Delete `handlers/` directory
- Update `go.mod`

---

#### 3. Simplify `order-service/main.go`

**File**: `practical5a/order-service/main.go`
**Changes**: Remove HTTP server, keep gRPC server and clients

**Note**: Order service is special - it still needs:
- gRPC **server** for handling order requests from gateway
- gRPC **clients** for calling user-service and menu-service

But it no longer needs:
- HTTP server
- HTTP handlers (delete `handlers/` directory)

---

#### 4. Update Service Dockerfiles

**Files**:
- `practical5a/user-service/Dockerfile`
- `practical5a/menu-service/Dockerfile`
- `practical5a/order-service/Dockerfile`

**Changes**: Update EXPOSE to only show gRPC port

**Before**:
```dockerfile
EXPOSE 8081 9091
```

**After**:
```dockerfile
EXPOSE 9091
```

---

#### 5. Update `docker-compose.yml`

**File**: `practical5a/docker-compose.yml`
**Changes**: Remove HTTP port mappings from services, add env vars to gateway

```yaml
services:
  user-service:
    build:
      context: .
      dockerfile: user-service/Dockerfile
    ports:
      - "9091:9091"  # Only gRPC port
    environment:
      GRPC_PORT: "9091"
      # Remove HTTP_PORT

  menu-service:
    build:
      context: .
      dockerfile: menu-service/Dockerfile
    ports:
      - "9092:9092"  # Only gRPC port
    environment:
      GRPC_PORT: "9092"

  order-service:
    build:
      context: .
      dockerfile: order-service/Dockerfile
    ports:
      - "9093:9093"  # Only gRPC port
    environment:
      GRPC_PORT: "9093"
      USER_SERVICE_GRPC_ADDR: "user-service:9091"
      MENU_SERVICE_GRPC_ADDR: "menu-service:9092"

  api-gateway:
    build:
      context: .
      dockerfile: api-gateway/Dockerfile
    ports:
      - "8080:8080"  # HTTP for external clients
    depends_on:
      - user-service
      - menu-service
      - order-service
    environment:
      USER_SERVICE_GRPC_ADDR: "user-service:9091"
      MENU_SERVICE_GRPC_ADDR: "menu-service:9092"
      ORDER_SERVICE_GRPC_ADDR: "order-service:9093"
```

**Key Changes**:
- Services expose only gRPC ports
- Gateway has gRPC service addresses
- Gateway still exposes 8080 for external HTTP

---

### Success Criteria - Phase 2

#### Automated Verification:
- [ ] All services build: `cd user-service && go build` (etc.)
- [ ] No unused dependencies in go.mod files
- [ ] Docker images build successfully
- [ ] docker-compose up succeeds

#### Manual Verification:
- [ ] Services log "gRPC server starting" (not HTTP)
- [ ] Services respond only on gRPC ports (9091/9092/9093)
- [ ] HTTP ports (8081/8082/8083) are closed
- [ ] Verify with: `curl http://localhost:8081` (should fail)
- [ ] Verify with: `curl http://localhost:8080/api/users` (should work via gateway)

---

## Phase 3: Update Documentation

### Overview
Create comprehensive student documentation explaining the complete migration and new architecture.

### Changes Required

#### 1. Create `practical5b.md`

**File**: `practicals/practical5b.md` (NEW)
**Changes**: Complete walkthrough document

**Sections**:
1. **Objective**: What students will learn
2. **Architecture Evolution**: Diagram showing Practical 5 → 5A → 5B
3. **Why This Matters**: Real-world relevance
4. **Implementation Walkthrough**:
   - Phase 1: Gateway migration (detailed)
   - Phase 2: Service simplification (detailed)
5. **Testing Guide**: Step-by-step testing instructions
6. **Troubleshooting**: Common issues and solutions
7. **Comparison Table**: HTTP vs gRPC Gateway
8. **Reflection Questions**: For student essays

**Key Educational Points to Explain**:
- **HTTP → gRPC Translation**: How the gateway bridges protocols
- **Error Mapping**: gRPC status codes → HTTP status codes
- **Pure gRPC Benefits**: Why removing HTTP simplifies services
- **Migration Strategy**: Gradual vs big-bang migrations
- **Production Patterns**: How companies like Google do this

---

#### 2. Update `practical5a/README.md`

**File**: `practicals/practical5a/README.md`
**Changes**: Add "Next Steps" section pointing to Practical 5B

```markdown
## Next Steps: Practical 5B

In Practical 5B, you'll complete the gRPC migration by:
- Converting the API Gateway to use gRPC clients
- Removing HTTP endpoints from backend services
- Achieving a pure gRPC internal architecture

See `../practical5b.md` for details.
```

---

#### 3. Create `practical5b/README.md`

**File**: `practicals/practical5b/README.md` (NEW)
**Changes**: Quick start guide for Practical 5B

**Sections**:
- Quick Start
- Architecture Diagram
- Testing Commands
- Troubleshooting
- Submission Requirements

---

#### 4. Update `deploy.sh`

**File**: `practicals/practical5b/deploy.sh` (NEW)
**Changes**: Deployment script for new architecture

```bash
#!/bin/bash

echo "===================================="
echo "Student Cafe - Practical 5B Deployment"
echo "Pure gRPC Backend with HTTP Gateway"
echo "===================================="

# ... deployment steps ...

echo "Architecture:"
echo "  External: HTTP/REST on :8080"
echo "  Internal: Pure gRPC"
echo "  - User Service (gRPC):   :9091"
echo "  - Menu Service (gRPC):   :9092"
echo "  - Order Service (gRPC):  :9093"
```

---

### Success Criteria - Phase 3

#### Automated Verification:
- [ ] All markdown files have no broken links
- [ ] Code examples in docs compile
- [ ] deploy.sh is executable

#### Manual Verification:
- [ ] Documentation clearly explains HTTP → gRPC translation
- [ ] Step-by-step instructions are tested and accurate
- [ ] Diagrams show architecture evolution
- [ ] Troubleshooting section covers common issues
- [ ] Student reflection questions are thought-provoking

---

## Testing Strategy

### Unit Tests
Not required for this practical (focus is on architecture), but could add:
- Gateway handler tests (mock gRPC clients)
- Error mapping tests

### Integration Tests
**End-to-End Flow**:

1. **Test User Creation**:
   ```bash
   curl -X POST http://localhost:8080/api/users \
     -H "Content-Type: application/json" \
     -d '{"name": "Test User", "email": "test@test.com"}'
   ```
   **Verify**: Gateway calls user-service via gRPC, returns HTTP 201

2. **Test Menu Retrieval**:
   ```bash
   curl http://localhost:8080/api/menu
   ```
   **Verify**: Gateway calls menu-service via gRPC, returns HTTP 200 JSON

3. **Test Order Creation**:
   ```bash
   curl -X POST http://localhost:8080/api/orders \
     -H "Content-Type: application/json" \
     -d '{"user_id": 1, "items": [{"menu_item_id": 1, "quantity": 2}]}'
   ```
   **Verify**:
   - Gateway calls order-service via gRPC
   - Order-service calls user-service via gRPC
   - Order-service calls menu-service via gRPC
   - Returns HTTP 201

4. **Test Error Handling**:
   ```bash
   curl http://localhost:8080/api/users/99999
   ```
   **Verify**: Returns HTTP 404 (gRPC NotFound → HTTP 404)

### Manual Testing Steps

1. **Verify Services are gRPC-Only**:
   ```bash
   # These should fail (no HTTP servers):
   curl http://localhost:9091
   curl http://localhost:9092
   curl http://localhost:9093

   # This should work (gateway):
   curl http://localhost:8080/api/menu
   ```

2. **Check Logs**:
   ```bash
   docker-compose logs api-gateway | grep "gRPC clients initialized"
   docker-compose logs user-service | grep "gRPC only"
   ```

3. **Performance Comparison** (optional):
   - Benchmark Practical 5A (HTTP between gateway and services)
   - Benchmark Practical 5B (gRPC between gateway and services)
   - Show gRPC is faster

---

## Performance Considerations

### Expected Improvements
- **Latency**: 20-30% reduction in service-to-service calls (gRPC vs HTTP)
- **Throughput**: Higher requests/second due to binary protocol
- **Resource Usage**: Lower CPU for serialization/deserialization

### Measurements (Optional Student Exercise)
```bash
# Benchmark tool
go install github.com/bojand/ghz/cmd/ghz@latest

# Benchmark gateway
ghz --insecure \
  --proto proto/user/v1/user.proto \
  --call user.v1.UserService/GetUsers \
  localhost:9091
```

---

## Migration Notes

### Backward Compatibility
**Breaking Changes**:
- Backend services no longer accept HTTP on ports 8081/8082/8083
- Direct service calls must use gRPC

**Non-Breaking**:
- External API unchanged (still HTTP on :8080)
- curl commands still work
- No client changes needed

### Rollback Strategy
If issues arise, can quickly revert:
1. Keep Practical 5A branch intact
2. Use git to switch between versions
3. Docker compose down/up to switch

### Migration Lessons
Students learn:
1. **Gradual Migration**: Changed one layer at a time
2. **Protocol Translation**: Gateway bridges HTTP and gRPC
3. **Service Simplification**: Removing unused code
4. **Production Patterns**: How real migrations happen

---

## Common Issues and Solutions

### Issue 1: Gateway Can't Connect to Services

**Symptom**:
```
Failed to connect to user service: connection refused
```

**Debug**:
```bash
docker-compose logs user-service | grep "gRPC server starting"
docker-compose ps  # Check service status
```

**Solution**: Verify GRPC_PORT environment variables and service startup

---

### Issue 2: Proto Import Errors

**Symptom**:
```
could not import github.com/douglasswm/student-cafe-protos
```

**Solution**: Ensure gateway Dockerfile copies proto module:
```dockerfile
COPY ../student-cafe-protos /student-cafe-protos
```

---

### Issue 3: HTTP Requests Return 500

**Symptom**: All requests return Internal Server Error

**Debug**: Check gateway logs for gRPC errors

**Solution**: Verify error handling in `handleGRPCError()` function

---

### Issue 4: Services Still Have HTTP Ports

**Symptom**: Can still curl http://localhost:8081

**Debug**: Check service main.go for startHTTPServer()

**Solution**: Ensure HTTP server code is removed, only gRPC remains

---

## References

- **Practical 5**: Original monolith → microservices
- **Practical 5A**: Inter-service gRPC migration
- **Practical 5B**: This plan (gateway gRPC migration)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/)
- [gRPC Gateway Pattern](https://grpc-ecosystem.github.io/grpc-gateway/)

---

## Summary

This plan transforms the Student Cafe application into a **pure gRPC architecture** internally while maintaining **HTTP compatibility** externally. Students learn:

1. **How to build gRPC clients** in a gateway
2. **Protocol translation** (HTTP ↔ gRPC)
3. **Service simplification** (removing dual protocols)
4. **Real-world migration patterns**
5. **Complete microservices evolution**: REST → Hybrid → Pure gRPC

**Estimated Student Time**: 3-4 hours
**Complexity**: Intermediate (requires understanding Practical 5A)
**Value**: High - demonstrates production-ready patterns

---

**End of Implementation Plan**
