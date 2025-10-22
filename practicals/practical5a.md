# Practical 5A: gRPC Migration with Centralized Proto Repository

## Objective

Building on Practical 5, this practical teaches you how to:
1. Refactor REST/HTTP microservices to use gRPC for inter-service communication
2. Create a centralized, versioned protocol buffer repository
3. Solve common proto synchronization and build issues
4. Implement dual protocol support (REST + gRPC) in microservices
5. Apply production-grade patterns for service communication

## Why This Matters

Real-world microservices often need:
- **Performance**: gRPC is faster than REST (binary vs text protocol)
- **Type Safety**: Protocol buffers provide compile-time type checking
- **Efficiency**: Less bandwidth usage with binary serialization
- **Flexibility**: Support both external REST APIs and internal gRPC communication

### Common Issues This Solves

In previous practicals, students encountered:
1. **Proto Sync Issues**: Each service had its own copy of proto files, leading to version mismatches
2. **Build Errors**: Docker builds failed when trying to copy proto files from other directories
3. **Circular Dependencies**: Services couldn't properly import each other's proto definitions
4. **Manual Synchronization**: Updating proto files required copying to multiple locations

## Solution Architecture

### Centralized Proto Repository Pattern

```
student-cafe-protos/          ← Standalone Go module
├── proto/                    ← Source of truth
│   ├── user/v1/
│   ├── menu/v1/
│   └── order/v1/
├── gen/go/                   ← Generated once
├── go.mod                    ← Versioned module
└── Makefile                  ← Generation automation

user-service/
├── go.mod                    ← Imports proto module
│   replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos
└── grpc/server.go           ← Uses generated code

menu-service/
└── (same pattern)

order-service/
└── (same pattern)
```

**Key Insight**: The proto repository is a separate Go module that all services import, just like any other dependency!

## Implementation Walkthrough

### Phase 1: Create Centralized Proto Repository

See complete implementation in `practicals/practical5a/student-cafe-protos/`

#### Step 1.1: Directory Structure

```bash
mkdir -p student-cafe-protos/{proto/{user,menu,order}/v1,gen/go}
```

Why this structure?
- `proto/`: Contains source `.proto` files organized by service and version
- `gen/go/`: Contains generated Go code
- Versioning (`/v1/`): Allows backward-compatible evolution

#### Step 1.2: Define Protocol Buffers

**Key Points**:
1. Each service gets its own proto package
2. Use consistent naming conventions
3. Include all CRUD operations as RPC methods
4. Define both request and response messages

Example from `user.proto`:
```protobuf
syntax = "proto3";
package user.v1;
option go_package = "github.com/douglasswm/student-cafe-protos/gen/go/user/v1;userv1";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc GetUsers(GetUsersRequest) returns (GetUsersResponse);
}
```

**Important**: The `go_package` option specifies where generated code goes and how to import it.

#### Step 1.3: Configure Code Generation

The `Makefile` automates proto generation:

```makefile
generate:
    @protoc \
        --go_out=gen/go \
        --go_opt=paths=source_relative \
        --go-grpc_out=gen/go \
        --go-grpc_opt=paths=source_relative \
        --proto_path=proto \
        proto/user/v1/user.proto \
        proto/menu/v1/menu.proto \
        proto/order/v1/order.proto
```

**Why this matters**:
- `--go_out`: Generates Protocol Buffer messages (structs)
- `--go-grpc_out`: Generates gRPC service interfaces and clients
- `paths=source_relative`: Keeps generated files organized by service

#### Step 1.4: Create Go Module

The proto repository is a proper Go module:

```go
// go.mod
module github.com/douglasswm/student-cafe-protos

go 1.23

require (
    google.golang.org/grpc v1.59.0
    google.golang.org/protobuf v1.31.0
)
```

This allows services to import it like any dependency!

### Phase 2: Add gRPC to Services

Each service needs three changes:
1. Update `go.mod` to import proto module
2. Create gRPC server implementation
3. Update `main.go` to run both HTTP and gRPC servers

#### Step 2.1: Import Proto Module

In each service's `go.mod`:

```go
require (
    github.com/douglasswm/student-cafe-protos v0.0.0
    google.golang.org/grpc v1.59.0
    google.golang.org/protobuf v1.31.0
)

// THIS IS CRITICAL for local development
replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos
```

**The `replace` directive**:
- During development: points to local proto module
- In production: remove and use tagged versions (v1.0.0, v1.1.0, etc.)
- Solves the "module not found" Docker build error!

#### Step 2.2: Implement gRPC Server

Example from `user-service/grpc/server.go`:

```go
type UserServer struct {
    userv1.UnimplementedUserServiceServer  // Important: embeds unimplemented methods
}

func (s *UserServer) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    var user models.User
    if err := database.DB.First(&user, req.Id).Error; err != nil {
        return nil, status.Errorf(codes.NotFound, "user not found")
    }

    return &userv1.GetUserResponse{
        User: modelToProto(&user),
    }, nil
}
```

**Key patterns**:
1. Embed `Unimplemented*Server` for forward compatibility
2. Use `status.Errorf()` for gRPC errors with proper codes
3. Convert between GORM models and proto messages

#### Step 2.3: Run Dual Servers

In `main.go`:

```go
func main() {
    database.Connect(dsn)

    // Start gRPC server in background goroutine
    go startGRPCServer()  // Port 9091

    // Start HTTP server (blocks)
    startHTTPServer()     // Port 8081
}
```

**Why dual servers?**
- External clients use HTTP/REST (familiar, easy to test with curl)
- Internal services use gRPC (faster, type-safe)
- Gradual migration path (can phase out HTTP later)

### Phase 3: gRPC Client Implementation

The order service needs to call user-service and menu-service via gRPC.

#### Step 3.1: Create gRPC Clients

`order-service/grpc/clients.go`:

```go
func NewClients() (*Clients, error) {
    userConn, err := grpc.NewClient(
        "user-service:9091",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )

    return &Clients{
        UserClient: userv1.NewUserServiceClient(userConn),
        MenuClient: menuv1.NewMenuServiceClient(menuConn),
    }, nil
}
```

**Important details**:
- Service discovery via DNS names (Docker Compose networking)
- `insecure.NewCredentials()`: No TLS for local development (use TLS in production!)
- Clients are reusable across requests (connection pooling)

#### Step 3.2: Use Clients in Handlers

`order-service/handlers/order_handlers.go`:

```go
func CreateOrder(w http.ResponseWriter, r *http.Request) {
    // ... parse request ...

    // Validate user via gRPC (instead of HTTP!)
    _, err := GrpcClients.UserClient.GetUser(ctx, &userv1.GetUserRequest{
        Id: uint32(req.UserID),
    })

    // Fetch menu item price via gRPC
    menuResp, err := GrpcClients.MenuClient.GetMenuItem(ctx, &menuv1.GetMenuItemRequest{
        Id: uint32(item.MenuItemID),
    })

    // ... create order ...
}
```

**Before (Practical 5)**: HTTP call with JSON parsing
**After (Practical 5A)**: gRPC call with type-safe proto messages

### Phase 4: Docker Configuration

#### The Docker Build Challenge

Problem: Services need the proto module at build time, but it's in a sibling directory.

**Solution**: Multi-stage build with context from parent directory

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /build

# Copy proto module first
COPY ../student-cafe-protos /student-cafe-protos

# Copy service files
WORKDIR /build/app
COPY go.mod go.sum ./
RUN go mod download  # Now works because replace directive finds module!
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service .
```

**Critical insight**: The `replace` directive in `go.mod` makes this work:
```go
replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos
```

Docker sees this and uses `/student-cafe-protos` (which we copied in the Dockerfile).

#### Docker Compose Configuration

```yaml
user-service:
  build:
    context: .  # Build context is root directory
    dockerfile: user-service/Dockerfile
  ports:
    - "8081:8081"  # HTTP
    - "9091:9091"  # gRPC
  environment:
    HTTP_PORT: "8081"
    GRPC_PORT: "9091"
```

**Key changes from Practical 5**:
1. Context is parent directory (so we can copy proto module)
2. Two ports exposed per service
3. Environment variables distinguish HTTP and gRPC ports

## How the gRPC Flow Works

Let's trace a complete order creation request:

```
1. Client → API Gateway
   HTTP POST /api/orders
   {"user_id": 1, "items": [{"menu_item_id": 1, "quantity": 2}]}

2. API Gateway → Order Service
   HTTP POST /orders (forwarded)

3. Order Service → User Service
   gRPC: userv1.UserService/GetUser
   Request: GetUserRequest{Id: 1}
   Response: GetUserResponse{User: {...}}
   ✅ User validated

4. Order Service → Menu Service
   gRPC: menuv1.MenuService/GetMenuItem
   Request: GetMenuItemRequest{Id: 1}
   Response: GetMenuItemResponse{MenuItem: {Price: 2.50}}
   ✅ Price fetched

5. Order Service → Database
   Creates order with snapshotted price

6. Order Service → Client
   HTTP 201 Created
   {"id": 1, "user_id": 1, "status": "pending", ...}
```

**Compare with Practical 5**:
- Steps 3 & 4 were HTTP GET requests with JSON
- Now they're gRPC calls with binary proto messages
- Type safety: compiler checks at build time
- Performance: ~2-5x faster than REST

## Deployment

### Automated Deployment

```bash
cd practicals/practical5a
./deploy.sh
```

The script:
1. Generates proto code (`make generate`)
2. Builds Docker images
3. Starts all services
4. Shows test commands

### Manual Deployment

```bash
# Generate proto code
cd student-cafe-protos && make generate && cd ..

# Build and start
docker-compose build
docker-compose up -d

# Verify
docker-compose ps
docker-compose logs -f order-service
```

## Testing & Verification

### Test 1: Create Menu Item

```bash
curl -X POST http://localhost:8080/api/menu \
  -H "Content-Type: application/json" \
  -d '{"name": "Latte", "description": "Espresso with milk", "price": 4.00}'
```

### Test 2: Create User

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "bob@test.com", "is_cafe_owner": false}'
```

### Test 3: Create Order (gRPC Communication)

```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "items": [{"menu_item_id": 1, "quantity": 2}]}'
```

**Verify gRPC is being used**:

```bash
docker-compose logs order-service | grep -i grpc
```

You should see:
```
gRPC clients initialized successfully
gRPC server starting on :9093
```

### Test 4: Direct gRPC Testing (Advanced)

Install grpcurl:
```bash
brew install grpcurl  # macOS
```

Call gRPC endpoint directly:
```bash
grpcurl -plaintext -d '{"id": 1}' \
  localhost:9091 user.v1.UserService/GetUser
```

## Troubleshooting Guide

### Problem 1: Proto Generation Fails

**Error**:
```
protoc-gen-go: program not found
```

**Solution**:
```bash
cd student-cafe-protos
make install-tools
export PATH=$PATH:$(go env GOPATH)/bin
make generate
```

### Problem 2: Docker Build - Module Not Found

**Error**:
```
go: github.com/douglasswm/student-cafe-protos@v0.0.0: invalid version
```

**Root Cause**: The `replace` directive can't find the proto module.

**Solution**:
1. Check Dockerfile copies proto module:
   ```dockerfile
   COPY ../student-cafe-protos /student-cafe-protos
   ```

2. Verify build context in docker-compose.yml:
   ```yaml
   build:
     context: .  # Must be parent directory
     dockerfile: user-service/Dockerfile
   ```

3. Verify `go.mod` has correct path:
   ```go
   replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos
   ```

### Problem 3: gRPC Connection Refused

**Error in logs**:
```
failed to connect to user service: connection refused
```

**Debugging steps**:

1. Check service is running:
   ```bash
   docker-compose ps user-service
   ```

2. Verify gRPC server started:
   ```bash
   docker-compose logs user-service | grep "gRPC server starting"
   ```

3. Check port configuration:
   ```yaml
   environment:
     GRPC_PORT: "9091"  # Must match client connection
   ```

4. Test connectivity:
   ```bash
   docker-compose exec order-service ping user-service
   ```

### Problem 4: Proto Changes Not Reflected

**Symptom**: Changed proto file but services still use old version.

**Solution**:
```bash
# Regenerate proto code
cd student-cafe-protos
make clean && make generate

# Rebuild services
cd ..
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## Key Learnings

### 1. Centralized Proto Repository Benefits

**Before (Practical 5)**:
- ❌ No proto files
- ❌ JSON schema validation at runtime
- ❌ No type safety

**After (Practical 5A)**:
- ✅ Single source of truth
- ✅ Versioned Go module
- ✅ Compile-time type checking
- ✅ No sync issues

### 2. gRPC vs REST Comparison

| Aspect | REST/HTTP | gRPC |
|--------|-----------|------|
| Protocol | Text (JSON) | Binary (Protobuf) |
| Performance | Good | Excellent (~2-5x faster) |
| Type Safety | Runtime | Compile-time |
| Tooling | curl, Postman | grpcurl, specialized tools |
| Browser Support | Native | Requires grpc-web |
| Human Readable | Yes | No (binary) |
| Bandwidth | Higher | Lower |

**When to use each**:
- **REST**: Public APIs, browser clients, simple CRUD
- **gRPC**: Internal microservices, high-performance, type safety critical

### 3. Production Considerations

#### Versioning Strategy

```bash
# Tag proto repository
cd student-cafe-protos
git tag v1.0.0
git push origin v1.0.0

# Services use tagged version
require github.com/douglasswm/student-cafe-protos v1.0.0
```

#### Adding TLS

```go
// Production: use TLS
creds, err := credentials.NewClientTLSFromFile("ca.pem", "")
conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
```

#### Health Checks

Implement gRPC health checking protocol:

```protobuf
import "google.golang.org/grpc/health/grpc_health_v1/health.proto";
```

## Extending the Implementation

### Add Server Streaming

Update proto:
```protobuf
service OrderService {
    rpc WatchOrders(WatchOrdersRequest) returns (stream Order);
}
```

Implement:
```go
func (s *OrderServer) WatchOrders(req *orderv1.WatchOrdersRequest, stream orderv1.OrderService_WatchOrdersServer) error {
    for {
        // Send order updates
        stream.Send(&orderv1.Order{...})
        time.Sleep(time.Second)
    }
}
```

### Add Interceptors

Logging interceptor:
```go
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    log.Printf("gRPC call: %s", info.FullMethod)
    return handler(ctx, req)
}

s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
```

## Conclusion

This practical demonstrates:

1. **Professional Proto Management**: Centralized repository solves real-world sync issues
2. **Hybrid Architecture**: Support both REST (external) and gRPC (internal)
3. **Type Safety**: Protocol buffers catch errors at compile-time
4. **Performance**: Binary serialization is faster than JSON
5. **Scalability**: Pattern used by Google, Netflix, Uber

You now understand how to build production-grade microservices with gRPC!

## Additional Resources

- [gRPC Core Concepts](https://grpc.io/docs/what-is-grpc/core-concepts/)
- [Protocol Buffers Guide](https://protobuf.dev/programming-guides/proto3/)
- [Go gRPC Best Practices](https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md)
- [Microservices Communication Patterns](https://microservices.io/patterns/communication-style/messaging.html)

---

**Congratulations!** You've successfully implemented gRPC microservices with a centralized proto repository. This pattern is production-ready and used by leading tech companies worldwide.
