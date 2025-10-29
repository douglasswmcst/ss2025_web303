# Practical 6 Example: Comprehensive Testing for Microservices

## Overview

This is the complete example implementation for **Practical 6**, demonstrating comprehensive testing strategies for gRPC microservices. It builds on Practical 5A and adds unit tests, integration tests, end-to-end tests, and test automation.

### Key Learning Objectives

1. **Unit Testing**: Test individual service methods in isolation using in-memory databases
2. **Mocking**: Use mocks to simulate external dependencies in unit tests
3. **Integration Testing**: Test multiple services working together using in-memory gRPC connections
4. **End-to-End Testing**: Validate the entire system through HTTP API requests
5. **Test Automation**: Use Makefile for consistent test execution and CI/CD integration

## What's Included

This example includes comprehensive tests at three levels:

### ✅ Unit Tests
- **User Service**: `user-service/grpc/server_test.go`
  - Tests for CreateUser, GetUser, GetUsers
  - In-memory SQLite database
  - Table-driven tests

- **Menu Service**: `menu-service/grpc/server_test.go`
  - Tests for CreateMenuItem, GetMenuItem, GetMenu
  - Price handling tests
  - Edge case validation

- **Order Service**: `order-service/grpc/server_test.go`
  - Tests with mocked gRPC clients
  - Validation scenarios (invalid user, invalid menu item)
  - Price snapshotting tests

### ✅ Integration Tests
Located in `tests/integration/integration_test.go`:
- Complete order flow across all services
- Order validation tests
- Concurrent request handling
- Uses bufconn (in-memory gRPC connections)

### ✅ End-to-End Tests
Located in `tests/e2e/e2e_test.go`:
- Full system validation via HTTP API
- User creation and retrieval
- Menu item management
- Complete order workflow
- Error handling validation
- Requires running Docker containers

## Project Structure

```
practical5a/
├── student-cafe-protos/          # Centralized proto repository
│   ├── proto/                    # Proto definition files
│   │   ├── user/v1/user.proto
│   │   ├── menu/v1/menu.proto
│   │   └── order/v1/order.proto
│   ├── gen/go/                   # Generated Go code
│   ├── go.mod                    # Go module definition
│   ├── Makefile                  # Proto generation commands
│   └── README.md                 # Proto repo documentation
├── user-service/                 # User microservice
│   ├── grpc/server.go           # gRPC server implementation
│   ├── handlers/                 # HTTP handlers (REST)
│   ├── main.go                   # Runs both HTTP and gRPC servers
│   └── Dockerfile
├── menu-service/                 # Menu microservice
│   ├── grpc/server.go
│   ├── handlers/
│   ├── main.go
│   └── Dockerfile
├── order-service/                # Order microservice
│   ├── grpc/
│   │   ├── server.go            # gRPC server
│   │   └── clients.go           # gRPC clients for other services
│   ├── handlers/                 # HTTP handlers (use gRPC internally)
│   ├── main.go
│   └── Dockerfile
├── api-gateway/                  # REST API Gateway
├── docker-compose.yml            # Orchestration config
├── deploy.sh                     # Deployment script
└── README.md                     # This file
```

## Prerequisites

### Required Tools

1. **Go** (1.23+)
   ```bash
   go version
   ```

2. **Protocol Buffer Compiler (protoc)**
   ```bash
   # macOS
   brew install protobuf

   # Linux
   sudo apt-get install protobuf-compiler

   # Verify
   protoc --version
   ```

3. **Go Proto Plugins**
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

4. **Docker & Docker Compose**
   ```bash
   docker --version
   docker-compose --version
   ```

## Quick Start

### Option 1: Automated Deployment (Recommended)

```bash
./deploy.sh
```

This script will:
1. Generate Go code from proto files
2. Clean up existing containers
3. Build Docker images
4. Start all services
5. Display access information

### Option 2: Manual Deployment

#### Step 1: Generate Proto Code

```bash
cd student-cafe-protos
make generate
cd ..
```

#### Step 2: Build and Start Services

```bash
docker-compose build
docker-compose up -d
```

#### Step 3: Verify Services

```bash
docker-compose ps
```

All services should show as "Up".

## Testing the Application

### 1. Create a Menu Item

```bash
curl -X POST http://localhost:8080/api/menu \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Cappuccino",
    "description": "Italian coffee with steamed milk",
    "price": 3.50
  }'
```

### 2. Create a User

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com",
    "is_cafe_owner": false
  }'
```

### 3. Create an Order (Demonstrates gRPC Communication)

```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "items": [
      {"menu_item_id": 1, "quantity": 2}
    ]
  }'
```

**What happens behind the scenes:**
1. Client sends HTTP request to API Gateway
2. API Gateway forwards to Order Service (HTTP)
3. Order Service validates user via **gRPC** call to User Service
4. Order Service fetches menu item price via **gRPC** call to Menu Service
5. Order Service creates order and returns HTTP response

### 4. Verify gRPC Communication

Check the order-service logs to see gRPC calls:

```bash
docker-compose logs order-service | grep gRPC
```

You should see messages like:
```
gRPC clients initialized successfully
gRPC server starting on :9093
```

## Understanding the gRPC Implementation

### 1. Centralized Proto Repository

**Location**: `student-cafe-protos/`

**Key Files**:
- `proto/user/v1/user.proto`: User service definitions
- `proto/menu/v1/menu.proto`: Menu service definitions
- `proto/order/v1/order.proto`: Order service definitions

**Why This Solves Previous Issues**:

In previous practicals, students often faced:
- ❌ Proto files duplicated across services
- ❌ Version mismatches between services
- ❌ Docker build errors when copying proto files
- ❌ Circular dependencies

Our solution:
- ✅ Single source of truth for all proto definitions
- ✅ Versioned Go module that services import
- ✅ Proto code generated once, used everywhere
- ✅ Docker builds work reliably with `replace` directive

### 2. How Services Import Proto Code

Each service's `go.mod` includes:

```go
require (
    github.com/douglasswm/student-cafe-protos v0.0.0
    google.golang.org/grpc v1.59.0
)

// Local development: point to the local proto module
replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos
```

**The `replace` directive** tells Go to use the local proto module instead of fetching from GitHub. This is perfect for development!

### 3. Dual Server Implementation

Each service runs **two servers concurrently**:

```go
func main() {
    // ... database connection ...

    // Start gRPC server in background
    go startGRPCServer()

    // Start HTTP server (blocks)
    startHTTPServer()
}
```

**HTTP Server** (port 8081/8082/8083):
- Handles REST API requests
- Used by API Gateway and external clients

**gRPC Server** (port 9091/9092/9093):
- Handles internal service-to-service calls
- More efficient than HTTP/REST
- Strongly typed with proto definitions

### 4. gRPC Client Usage (Order Service Example)

The order service creates gRPC clients to call other services:

```go
// In main.go
clients, err := grpcserver.NewClients()
handlers.GrpcClients = clients

// In handlers/order_handlers.go
// Validate user via gRPC
userResp, err := GrpcClients.UserClient.GetUser(ctx, &userv1.GetUserRequest{
    Id: uint32(req.UserID),
})

// Get menu item via gRPC
menuResp, err := GrpcClients.MenuClient.GetMenuItem(ctx, &menuv1.GetMenuItemRequest{
    Id: uint32(item.MenuItemID),
})
```

**Benefits over HTTP**:
- Type safety (compile-time checks)
- Better performance (binary protocol)
- Streaming support (not used here, but available)
- Built-in load balancing and retries

## Troubleshooting

### Issue 1: Proto Generation Fails

**Symptom**:
```
protoc-gen-go: program not found
```

**Solution**:
```bash
cd student-cafe-protos
make install-tools
```

### Issue 2: Docker Build Fails - Proto Module Not Found

**Symptom**:
```
go: github.com/douglasswm/student-cafe-protos@v0.0.0: invalid version
```

**Solution**:
Ensure the Dockerfile copies the proto module:
```dockerfile
COPY ../student-cafe-protos /student-cafe-protos
```

And the service `go.mod` has the `replace` directive.

### Issue 3: gRPC Connection Refused

**Symptom**:
```
failed to connect to user service: connection refused
```

**Solutions**:
1. Verify services are running:
   ```bash
   docker-compose ps
   ```

2. Check gRPC ports are exposed:
   ```bash
   docker-compose logs user-service | grep "gRPC server"
   ```

3. Verify service names in `docker-compose.yml` match code:
   ```yaml
   environment:
     USER_SERVICE_GRPC_ADDR: "user-service:9091"
   ```

### Issue 4: Changes Not Reflected

**Symptom**: Code changes don't appear after rebuild

**Solution**:
```bash
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## Updating Proto Definitions

### Step-by-Step Guide

1. **Modify Proto Files**

   Edit the relevant `.proto` file in `student-cafe-protos/proto/`:

   ```protobuf
   // Add a new field to User
   message User {
       uint32 id = 1;
       string name = 2;
       string email = 3;
       bool is_cafe_owner = 4;
       string phone_number = 5;  // NEW FIELD
   }
   ```

2. **Regenerate Code**

   ```bash
   cd student-cafe-protos
   make clean && make generate
   ```

3. **Update Service Implementation**

   Update the model conversion in the affected service:

   ```go
   // In user-service/grpc/server.go
   func modelToProto(user *models.User) *userv1.User {
       return &userv1.User{
           // ... existing fields ...
           PhoneNumber: user.PhoneNumber,  // NEW
       }
   }
   ```

4. **Rebuild and Deploy**

   ```bash
   cd ..
   ./deploy.sh
   ```

## Advanced Topics

### Versioning Proto Definitions

For production, you'd tag the proto module:

```bash
cd student-cafe-protos
git add .
git commit -m "Add phone_number field to User"
git tag v1.1.0
git push origin v1.1.0
```

Then services can pin to specific versions:

```go
require (
    github.com/douglasswm/student-cafe-protos v1.1.0
)
```

### Adding gRPC Interceptors

For logging, authentication, or error handling:

```go
s := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),
)
```

### gRPC Streaming

The proto definitions support streaming (not implemented here):

```protobuf
service OrderService {
    // Server streaming - watch order updates
    rpc WatchOrder(WatchOrderRequest) returns (stream Order);
}
```

## Comparison with Practical 5

| Aspect | Practical 5 (HTTP) | Practical 5A (gRPC) |
|--------|-------------------|-------------------|
| **Inter-Service Protocol** | HTTP/REST | gRPC |
| **External API** | HTTP/REST | HTTP/REST (same) |
| **Type Safety** | JSON (runtime) | Protobuf (compile-time) |
| **Performance** | Good | Better (binary protocol) |
| **Proto Management** | N/A | Centralized module |
| **Complexity** | Lower | Higher (but more scalable) |

## Key Takeaways

1. **Centralized Proto Repository**: Solves version sync and build issues by treating proto definitions as a versioned Go module

2. **Dual Protocol Support**: Services can speak both HTTP (for clients) and gRPC (for internal communication)

3. **gRPC Benefits**:
   - Type safety via proto definitions
   - Better performance than REST
   - Built-in features (streaming, deadlines, cancellation)

4. **Production Ready**: This pattern is used by companies like Google, Netflix, and Uber

5. **Trade-offs**: More complex than pure REST, but scales better for large systems

## Next Steps

1. **Add gRPC Health Checks**: Implement the gRPC health checking protocol
2. **Add Metrics**: Collect gRPC metrics with Prometheus
3. **Deploy to Kubernetes**: Migrate from Docker Compose to K8s
4. **Add Service Mesh**: Integrate Istio for advanced traffic management
5. **Implement Streaming**: Add real-time order updates using server streaming

## Resources

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://protobuf.dev/)
- [Go gRPC Tutorial](https://grpc.io/docs/languages/go/quickstart/)
- [Microservices Patterns](https://microservices.io/patterns/index.html)

## Submission Requirements

### What to Submit

1. **All Source Code**:
   - `student-cafe-protos/` directory
   - All service directories
   - `docker-compose.yml`
   - `deploy.sh`

2. **Documentation**:
   - This `README.md` with your observations
   - Screenshots showing:
     - Successful proto generation
     - All services running (docker-compose ps)
     - Successful order creation (demonstrating gRPC communication)
     - Service logs showing gRPC connections

3. **Reflection Essay (500 words minimum)**:
   - How does the centralized proto repository solve the issues from previous practicals?
   - Compare HTTP vs gRPC for inter-service communication
   - What are the trade-offs of running dual servers (HTTP + gRPC)?
   - When would you choose gRPC over REST?
   - How would you version the proto module in production?

### Grading Criteria

| Criteria | Weight |
|----------|--------|
| Proto repository structure and generation | 20% |
| gRPC server implementations | 25% |
| gRPC client usage in order-service | 25% |
| Docker configuration and deployment | 15% |
| Documentation and reflection | 15% |

---

**Good luck!** This practical demonstrates production-grade microservices communication patterns. Understanding these concepts will make you valuable in any microservices organization.
