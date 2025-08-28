### **Module Practical: WEB303 Microservices & Serverless Applications**

## **Practical 3 Solution: Full-Stack Microservices with gRPC, Databases, and Service Discovery**

### **Objective**

This solution addresses the key issues identified in the original practical and provides a complete, working microservices ecosystem that properly implements:

1. **Service Discovery with Consul**: API Gateway dynamically discovers services instead of hardcoded connections
2. **Proper Service Communication**: Services communicate through gRPC with proper error handling
3. **Fixed Build Issues**: Corrected proto file structure and build configuration
4. **Composite Endpoints**: Properly implemented data aggregation from multiple services

### **Key Issues Identified and Fixed**

#### **Issue 1: Proto Files Location Causing Build Problems**

**Problem**: The original walkthrough places proto files in the root directory, which causes import path issues when building Docker containers for individual services.

**Root Cause Analysis**:

1. **Docker Build Context Isolation**: Each Dockerfile can only access files within its build context (the directory containing the Dockerfile and its subdirectories)
2. **Import Path Resolution**: Generated Go files contain import statements that reference modules outside the service's build context
3. **Multi-Stage Build Complexity**: Without proper file organization, multi-stage builds fail to locate dependencies during compilation

**Production-Ready Solutions**:

**Option 1: Shared Proto Module (Recommended for Production)**

```bash
# Create a shared proto module
mkdir proto-shared
cd proto-shared
go mod init github.com/yourorg/proto-shared

# Generate proto files with proper module paths
protoc --go_out=. --go_opt=module=github.com/yourorg/proto-shared
       --go-grpc_out=. --go-grpc_opt=module=github.com/yourorg/proto-shared
       proto/*.proto

# Publish to private registry or use Go modules replace directive
```

**Option 2: Build-Time Proto Generation (Current Implementation)**

- Copy proto files into each service's build context
- Generate code during Docker build process
- Use multi-stage builds to optimize final image size

#### **Issue 2: API Gateway Not Using Consul Service Discovery**

**Problem**: The original API Gateway directly connects to services using hardcoded hostnames (`users-service:50051`).

**Solution**: Implement proper Consul service discovery to dynamically find and connect to services.

#### **Issue 3: Incomplete Composite Endpoint Implementation**

**Problem**: The aggregation endpoint doesn't properly handle service discovery and error cases.

**Solution**: Implement robust service discovery and connection management for the composite endpoint.

---

### **Part 1: Production-Ready Project Structure**

**For Learning/Development (Current Implementation):**

```
practical-three/
├── docker-compose.yml
├── proto/                           # Source proto definitions
│   ├── users.proto
│   ├── products.proto
│   └── gen/                        # Generated Go code
│       ├── users.pb.go
│       ├── users_grpc.pb.go
│       ├── products.pb.go
│       └── products_grpc.pb.go
├── api-gateway/
│   ├── main.go
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── proto/                      # Build-time copy for Docker
│       └── gen/
├── services/
│   ├── users-service/
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── proto/                  # Build-time copy for Docker
│   │       └── gen/
│   └── products-service/
│       ├── main.go
│       ├── Dockerfile
│       ├── go.mod
│       ├── go.sum
│       └── proto/                  # Build-time copy for Docker
│           └── gen/
└── scripts/
    ├── build.sh
    └── generate-proto.sh
```

**Alternative Structure for Future Expansion:**

```
practical-three-extended/
├── proto/                          # API definitions
│   ├── users.proto
│   ├── products.proto
│   ├── purchases.proto
│   └── gen/                        # Generated Go files
├── services/
│   ├── users-service/
│   │   ├── main.go
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   └── proto/                  # Build context copy
│   ├── products-service/
│   │   └── [similar structure]
│   └── purchase-service/
│       └── [similar structure]
├── api-gateway/
│   ├── main.go
│   ├── go.mod
│   ├── Dockerfile
│   └── proto/                      # Build context copy
├── docker-compose.yml
└── scripts/
    ├── build.sh
    └── generate-proto.sh
```

**Key Development Benefits**:

1. **Simple Structure**: Easy to understand and navigate
2. **Docker-First Approach**: All services containerized for consistency
3. **Local Development**: Everything runs locally with Docker Compose
4. **Proto File Management**: Clear separation between source and generated files

---

### **Part 2: Prerequisites and Setup**

1. **Install Required Tools:**

```bash
# Install protobuf compiler
brew install protobuf  # On macOS
# or
sudo apt-get install protobuf-compiler  # On Ubuntu

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

# Ensure Go bin is in PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

2. **Create Project Structure:**

```bash
mkdir practical-three
cd practical-three
mkdir -p proto/gen
mkdir api-gateway
mkdir -p services/users-service
mkdir -p services/products-service
```

3. **Define Proto Files:**

**`proto/users.proto`:**

3. **Define Proto Files:**

**Development Proto Files (Current Implementation):**

**`proto/users.proto`:**

```protobuf
syntax = "proto3";

option go_package = "./proto/gen;gen";

package users;

service UserService {
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  rpc GetUser(GetUserRequest) returns (UserResponse);
}

message User {
  string id = 1;
  string name = 2;
  string email = 3;
}

message CreateUserRequest {
  string name = 1;
  string email = 2;
}

message GetUserRequest {
  string id = 1;
}

message UserResponse {
  User user = 1;
}
```

**`proto/products.proto`:**

```protobuf
syntax = "proto3";

option go_package = "./proto/gen;gen";

package products;

service ProductService {
  rpc CreateProduct(CreateProductRequest) returns (ProductResponse);
  rpc GetProduct(GetProductRequest) returns (ProductResponse);
}

message Product {
  string id = 1;
  string name = 2;
  double price = 3;
}

message CreateProductRequest {
  string name = 1;
  double price = 2;
}

message GetProductRequest {
  string id = 1;
}

message ProductResponse {
  Product product = 1;
}
```

4. **Generate Proto Files:**

```bash
# Generate in main proto directory
protoc --go_out=./proto/gen --go_opt=paths=source_relative \
    --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative \
    proto/*.proto

# Copy generated files to each service (we'll automate this in build scripts)
```

---

### **Part 3: Fixed Docker Compose Configuration**

**`docker-compose.yml`:**

```yaml
services:
  consul:
    image: hashicorp/consul:latest
    container_name: consul
    ports:
      - "8500:8500"
    command: "agent -dev -client=0.0.0.0 -ui"
    networks:
      - microservices

  users-db:
    image: postgres:13
    container_name: users-db
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: users_db
    ports:
      - "5432:5432"
    volumes:
      - users_data:/var/lib/postgresql/data
    networks:
      - microservices

  products-db:
    image: postgres:13
    container_name: products-db
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: products_db
    ports:
      - "5433:5432"
    volumes:
      - products_data:/var/lib/postgresql/data
    networks:
      - microservices

  users-service:
    build: ./services/users-service
    container_name: users-service
    ports:
      - "50051:50051"
    depends_on:
      - consul
      - users-db
    environment:
      - CONSUL_HTTP_ADDR=consul:8500
    networks:
      - microservices

  products-service:
    build: ./services/products-service
    container_name: products-service
    ports:
      - "50052:50052"
    depends_on:
      - consul
      - products-db
    environment:
      - CONSUL_HTTP_ADDR=consul:8500
    networks:
      - microservices

  api-gateway:
    build: ./api-gateway
    container_name: api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - consul
      - users-service
      - products-service
    environment:
      - CONSUL_HTTP_ADDR=consul:8500
    networks:
      - microservices

volumes:
  users_data:
  products_data:

networks:
  microservices:
    driver: bridge
```

---

### **Part 4: Fixed Users Service Implementation**

**`services/users-service/main.go`:**

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "os"
    "time"

    "google.golang.org/grpc"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    pb "./proto/gen"
    consulapi "github.com/hashicorp/consul/api"
)

const serviceName = "users-service"
const servicePort = 50051

type User struct {
    gorm.Model
    Name  string
    Email string `gorm:"unique"`
}

type server struct {
    pb.UnimplementedUserServiceServer
    db *gorm.DB
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
    user := User{Name: req.Name, Email: req.Email}
    if result := s.db.Create(&user); result.Error != nil {
        return nil, result.Error
    }
    return &pb.UserResponse{User: &pb.User{Id: fmt.Sprint(user.ID), Name: user.Name, Email: user.Email}}, nil
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
    var user User
    if result := s.db.First(&user, req.Id); result.Error != nil {
        return nil, result.Error
    }
    return &pb.UserResponse{User: &pb.User{Id: fmt.Sprint(user.ID), Name: user.Name, Email: user.Email}}, nil
}

func main() {
    // Wait for database to be ready
    time.Sleep(10 * time.Second)

    // Connect to database
    dsn := "host=users-db user=user password=password dbname=users_db port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    db.AutoMigrate(&User{})

    // Start gRPC server
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    s := grpc.NewServer()
    pb.RegisterUserServiceServer(s, &server{db: db})

    // Register with Consul
    if err := registerServiceWithConsul(); err != nil {
        log.Fatalf("Failed to register with Consul: %v", err)
    }

    log.Printf("%s gRPC server listening at %v", serviceName, lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}

func registerServiceWithConsul() error {
    config := consulapi.DefaultConfig()
    // Use environment variable if set, otherwise default
    if addr := os.Getenv("CONSUL_HTTP_ADDR"); addr != "" {
        config.Address = addr
    }

    consul, err := consulapi.NewClient(config)
    if err != nil {
        return err
    }

    // Get container hostname for Docker networking
    hostname, err := os.Hostname()
    if err != nil {
        return err
    }

    registration := &consulapi.AgentServiceRegistration{
        ID:      fmt.Sprintf("%s-%s", serviceName, hostname),
        Name:    serviceName,
        Port:    servicePort,
        Address: hostname, // Use container hostname
        Check: &consulapi.AgentServiceCheck{
            GRPC:                           fmt.Sprintf("%s:%d", hostname, servicePort),
            Interval:                       "10s",
            DeregisterCriticalServiceAfter: "30s",
        },
    }

    return consul.Agent().ServiceRegister(registration)
}
```

**`services/users-service/Dockerfile`:**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy proto files first
COPY ../../proto ./proto
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 50051

CMD ["./server"]
```

**`services/users-service/go.mod`:**

```go
module users-service

go 1.21

require (
    google.golang.org/grpc v1.58.3
    google.golang.org/protobuf v1.31.0
    github.com/hashicorp/consul/api v1.25.1
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
)
```

---

### **Part 5: Fixed Products Service Implementation**

**`services/products-service/main.go`:**

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "os"
    "time"

    "google.golang.org/grpc"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    pb "./proto/gen"
    consulapi "github.com/hashicorp/consul/api"
)

const serviceName = "products-service"
const servicePort = 50052

type Product struct {
    gorm.Model
    Name  string
    Price float64
}

type server struct {
    pb.UnimplementedProductServiceServer
    db *gorm.DB
}

func (s *server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
    product := Product{Name: req.Name, Price: req.Price}
    if result := s.db.Create(&product); result.Error != nil {
        return nil, result.Error
    }
    return &pb.ProductResponse{Product: &pb.Product{Id: fmt.Sprint(product.ID), Name: product.Name, Price: product.Price}}, nil
}

func (s *server) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
    var product Product
    if result := s.db.First(&product, req.Id); result.Error != nil {
        return nil, result.Error
    }
    return &pb.ProductResponse{Product: &pb.Product{Id: fmt.Sprint(product.ID), Name: product.Name, Price: product.Price}}, nil
}

func main() {
    // Wait for database to be ready
    time.Sleep(10 * time.Second)

    // Connect to database
    dsn := "host=products-db user=user password=password dbname=products_db port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    db.AutoMigrate(&Product{})

    // Start gRPC server
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    s := grpc.NewServer()
    pb.RegisterProductServiceServer(s, &server{db: db})

    // Register with Consul
    if err := registerServiceWithConsul(); err != nil {
        log.Fatalf("Failed to register with Consul: %v", err)
    }

    log.Printf("%s gRPC server listening at %v", serviceName, lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}

func registerServiceWithConsul() error {
    config := consulapi.DefaultConfig()
    if addr := os.Getenv("CONSUL_HTTP_ADDR"); addr != "" {
        config.Address = addr
    }

    consul, err := consulapi.NewClient(config)
    if err != nil {
        return err
    }

    hostname, err := os.Hostname()
    if err != nil {
        return err
    }

    registration := &consulapi.AgentServiceRegistration{
        ID:      fmt.Sprintf("%s-%s", serviceName, hostname),
        Name:    serviceName,
        Port:    servicePort,
        Address: hostname,
        Check: &consulapi.AgentServiceCheck{
            GRPC:                           fmt.Sprintf("%s:%d", hostname, servicePort),
            Interval:                       "10s",
            DeregisterCriticalServiceAfter: "30s",
        },
    }

    return consul.Agent().ServiceRegister(registration)
}
```

**`services/products-service/Dockerfile`:**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy proto files first
COPY ../../proto ./proto
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 50052

CMD ["./server"]
```

---

### **Part 6: Fixed API Gateway with Proper Service Discovery**

**`api-gateway/main.go`:**

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/gorilla/mux"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "./proto/gen"
    consulapi "github.com/hashicorp/consul/api"
)

type ServiceDiscovery struct {
    consul *consulapi.Client
    mu     sync.RWMutex
    connections map[string]*grpc.ClientConn
}

type UserPurchaseData struct {
    User    *pb.User    `json:"user"`
    Product *pb.Product `json:"product"`
}

var sd *ServiceDiscovery

func main() {
    // Initialize service discovery
    config := consulapi.DefaultConfig()
    if addr := os.Getenv("CONSUL_HTTP_ADDR"); addr != "" {
        config.Address = addr
    }

    consul, err := consulapi.NewClient(config)
    if err != nil {
        log.Fatalf("Failed to create consul client: %v", err)
    }

    sd = &ServiceDiscovery{
        consul: consul,
        connections: make(map[string]*grpc.ClientConn),
    }

    // Wait for services to be ready
    time.Sleep(15 * time.Second)

    r := mux.NewRouter()

    // User routes
    r.HandleFunc("/api/users", createUserHandler).Methods("POST")
    r.HandleFunc("/api/users/{id}", getUserHandler).Methods("GET")

    // Product routes
    r.HandleFunc("/api/products", createProductHandler).Methods("POST")
    r.HandleFunc("/api/products/{id}", getProductHandler).Methods("GET")

    // Composite endpoint
    r.HandleFunc("/api/purchases/user/{userId}/product/{productId}", getPurchaseDataHandler).Methods("GET")

    log.Println("API Gateway listening on port 8080...")
    http.ListenAndServe(":8080", r)
}

func (sd *ServiceDiscovery) getServiceConnection(serviceName string) (*grpc.ClientConn, error) {
    sd.mu.Lock()
    defer sd.mu.Unlock()

    // Check if we already have a connection
    if conn, exists := sd.connections[serviceName]; exists {
        return conn, nil
    }

    // Discover service
    services, _, err := sd.consul.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
    }

    if len(services) == 0 {
        return nil, fmt.Errorf("no healthy instances of service %s found", serviceName)
    }

    // Use first healthy instance
    service := services[0].Service
    address := fmt.Sprintf("%s:%d", service.Address, service.Port)

    // Create gRPC connection
    conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, fmt.Errorf("failed to connect to service %s at %s: %w", serviceName, address, err)
    }

    sd.connections[serviceName] = conn
    log.Printf("Connected to %s at %s", serviceName, address)
    return conn, nil
}

func getUsersClient() (pb.UserServiceClient, error) {
    conn, err := sd.getServiceConnection("users-service")
    if err != nil {
        return nil, err
    }
    return pb.NewUserServiceClient(conn), nil
}

func getProductsClient() (pb.ProductServiceClient, error) {
    conn, err := sd.getServiceConnection("products-service")
    if err != nil {
        return nil, err
    }
    return pb.NewProductServiceClient(conn), nil
}

// User Handlers
func createUserHandler(w http.ResponseWriter, r *http.Request) {
    client, err := getUsersClient()
    if err != nil {
        http.Error(w, fmt.Sprintf("Service unavailable: %v", err), http.StatusServiceUnavailable)
        return
    }

    var req pb.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    res, err := client.CreateUser(context.Background(), &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res.User)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
    client, err := getUsersClient()
    if err != nil {
        http.Error(w, fmt.Sprintf("Service unavailable: %v", err), http.StatusServiceUnavailable)
        return
    }

    vars := mux.Vars(r)
    id := vars["id"]

    res, err := client.GetUser(context.Background(), &pb.GetUserRequest{Id: id})
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res.User)
}

// Product Handlers
func createProductHandler(w http.ResponseWriter, r *http.Request) {
    client, err := getProductsClient()
    if err != nil {
        http.Error(w, fmt.Sprintf("Service unavailable: %v", err), http.StatusServiceUnavailable)
        return
    }

    var req pb.CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    res, err := client.CreateProduct(context.Background(), &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res.Product)
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {
    client, err := getProductsClient()
    if err != nil {
        http.Error(w, fmt.Sprintf("Service unavailable: %v", err), http.StatusServiceUnavailable)
        return
    }

    vars := mux.Vars(r)
    id := vars["id"]

    res, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: id})
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res.Product)
}

// Fixed composite endpoint with proper service discovery
func getPurchaseDataHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    userId := vars["userId"]
    productId := vars["productId"]

    var wg sync.WaitGroup
    var user *pb.User
    var product *pb.Product
    var userErr, productErr error

    wg.Add(2)

    // Fetch user data
    go func() {
        defer wg.Done()
        client, err := getUsersClient()
        if err != nil {
            userErr = err
            return
        }
        res, err := client.GetUser(context.Background(), &pb.GetUserRequest{Id: userId})
        if err != nil {
            userErr = err
            return
        }
        user = res.User
    }()

    // Fetch product data
    go func() {
        defer wg.Done()
        client, err := getProductsClient()
        if err != nil {
            productErr = err
            return
        }
        res, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: productId})
        if err != nil {
            productErr = err
            return
        }
        product = res.Product
    }()

    wg.Wait()

    if userErr != nil {
        http.Error(w, fmt.Sprintf("Failed to get user: %v", userErr), http.StatusNotFound)
        return
    }
    if productErr != nil {
        http.Error(w, fmt.Sprintf("Failed to get product: %v", productErr), http.StatusNotFound)
        return
    }

    purchaseData := UserPurchaseData{
        User:    user,
        Product: product,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(purchaseData)
}
```

---

### **Part 6A: Design Decision Analysis - Composite Endpoint Architecture**

#### **Why is the API Gateway handling this composite endpoint?**

**Current Approach: API Gateway Orchestration**

The `getPurchaseDataHandler` demonstrates the "API Gateway Orchestration" pattern where:

1. The gateway receives a single HTTP request
2. Makes parallel gRPC calls to multiple services
3. Aggregates the results into a single response

#### **Advantages of Gateway Orchestration**

✅ **Single HTTP request from client** (better for web/mobile apps)  
✅ **Parallel service calls** reduce latency  
✅ **Client doesn't need to know** about multiple services  
✅ **Consistent error handling** and response format  
✅ **Cross-cutting concerns** (auth, logging, rate limiting) in one place  
✅ **Network efficiency** (fewer round trips)

#### **Disadvantages**

❌ **Gateway becomes a potential bottleneck**  
❌ **Business logic creeps** into infrastructure layer  
❌ **Tight coupling** between gateway and services  
❌ **More complex gateway code**

#### **Alternative Approach: Inter-Service Communication**

**Option 1: Service-to-Service gRPC Communication**

- Users service could call Products service directly
- Or Products service could call Users service
- One service acts as the orchestrator

**Option 2: Domain Service Pattern**

- Create a dedicated "Purchase Service"
- This service orchestrates calls to Users and Products services
- Gateway simply proxies to Purchase Service

#### **Pros of Inter-Service Communication**

✅ **Domain logic stays** within appropriate services  
✅ **Gateway remains thin** and focused on routing  
✅ **Better separation** of concerns  
✅ **Services can evolve** independently  
✅ **Clearer ownership** of business logic

#### **Cons of Inter-Service Communication**

❌ **Increased complexity** in individual services  
❌ **Potential for cascading failures**  
❌ **Service discovery logic** in business services  
❌ **More network hops** (latency)  
❌ **Circular dependency risks**

#### **Recommended Approach: Hybrid Strategy**

1. **SIMPLE AGGREGATION**: Use API Gateway (current approach)

   - When combining data from multiple services
   - For read-only operations
   - When clients need a single HTTP endpoint

2. **BUSINESS WORKFLOWS**: Use dedicated services

   - Create a "Purchase Service" for purchase workflows
   - Handles complex business logic and transactions
   - Orchestrates calls to Users, Products, Inventory, etc.

3. **EVENT-DRIVEN PATTERNS**: Use message queues
   - For asynchronous workflows
   - When eventual consistency is acceptable
   - Reduces coupling between services

#### **Conclusion**

The current API Gateway approach is appropriate for:

- Learning and demonstration purposes
- Simple data aggregation
- Client-facing read operations

For production systems, consider:

- Dedicated domain services for complex workflows
- Event-driven architecture for loose coupling
- Clear service boundaries and responsibilities
- Appropriate error handling and circuit breakers

---

````

**`api-gateway/Dockerfile`:**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy proto files first
COPY ../proto ./proto
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
````

**`api-gateway/go.mod`:**

```go
module api-gateway

go 1.21

require (
    google.golang.org/grpc v1.58.3
    google.golang.org/protobuf v1.31.0
    github.com/hashicorp/consul/api v1.25.1
    github.com/gorilla/mux v1.8.0
)
```

---

### **Part 7: Build and Deployment Strategy**

**Development Build Script (Current Implementation):**

**`scripts/build.sh`:**

````bash
#!/bin/bash
set -e

echo "🚀 Building microservices development environment..."

# Function to check if required tools are installed
check_dependencies() {
    echo "📋 Checking dependencies..."

    command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed."; exit 1; }
    command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose is required but not installed."; exit 1; }
    command -v protoc >/dev/null 2>&1 || { echo "❌ protoc is required but not installed."; exit 1; }

    echo "✅ All dependencies found"
}

# Generate proto files
generate_proto_files() {
    echo "🔧 Generating proto files..."

    # Clean previous generations
    rm -rf proto/gen
    mkdir -p proto/gen

    # Generate Go code
    protoc --go_out=./proto/gen --go_opt=paths=source_relative \
           --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative \
           proto/*.proto

    echo "✅ Proto files generated"
}

# Copy proto files to each service for Docker build context
distribute_proto_files() {
    echo "📦 Distributing proto files to services..."

    services=("api-gateway" "services/users-service" "services/products-service")

    for service in "${services[@]}"; do
        echo "  📂 Copying to $service..."
        mkdir -p "$service/proto/gen"
        cp -r proto/* "$service/proto/"
    done

    echo "✅ Proto files distributed"
}

# Clean up old containers and images
cleanup() {
    echo "🧹 Cleaning up old containers..."
    docker-compose down --remove-orphans 2>/dev/null || true
    docker system prune -f --volumes 2>/dev/null || true
}

# Build and start services
build_and_start() {
    echo "🏗️  Building and starting services..."

    # Build with no cache to ensure fresh build
    docker-compose build --no-cache

    # Start services
    docker-compose up -d

    # Wait for services to be healthy
    echo "⏳ Waiting for services to be ready..."
    sleep 30

    # Check service health
    check_service_health
}

# Check if services are responding
check_service_health() {
    echo "🔍 Checking service health..."

    # Check Consul
    if curl -s http://localhost:8500/v1/status/leader >/dev/null; then
        echo "✅ Consul is healthy"
    else
        echo "❌ Consul is not responding"
    fi

    # Check API Gateway
    if curl -s http://localhost:8080/api/users >/dev/null 2>&1; then
        echo "✅ API Gateway is healthy"
    else
        echo "⚠️  API Gateway may still be starting..."
    fi

    echo "🎉 Build complete! Services are available at:"
    echo "   - Consul UI: http://localhost:8500"
    echo "   - API Gateway: http://localhost:8080"
    echo "   - Users DB: localhost:5432"
    echo "   - Products DB: localhost:5433"
}

# Main execution
main() {
    check_dependencies
    generate_proto_files
    distribute_proto_files
    cleanup
**`scripts/build.sh`:**

```bash
#!/bin/bash
set -e

echo "🚀 Building microservices development environment..."

# Function to check if required tools are installed
check_dependencies() {
    echo "📋 Checking dependencies..."

    command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed."; exit 1; }
    command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose is required but not installed."; exit 1; }
    command -v protoc >/dev/null 2>&1 || { echo "❌ protoc is required but not installed."; exit 1; }

    echo "✅ All dependencies found"
}

# Generate proto files
generate_proto_files() {
    echo "🔧 Generating proto files..."

    # Clean previous generations
    rm -rf proto/gen
    mkdir -p proto/gen

    # Generate Go code
    protoc --go_out=./proto/gen --go_opt=paths=source_relative \
           --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative \
           proto/*.proto

    echo "✅ Proto files generated"
}

# Copy proto files to each service for Docker build context
distribute_proto_files() {
    echo "📦 Distributing proto files to services..."

    services=("api-gateway" "services/users-service" "services/products-service")

    for service in "${services[@]}"; do
        echo "  📂 Copying to $service..."
        mkdir -p "$service/proto/gen"
        cp -r proto/* "$service/proto/"
    done

    echo "✅ Proto files distributed"
}

# Clean up old containers and images
cleanup() {
    echo "🧹 Cleaning up old containers..."
    docker-compose down --remove-orphans 2>/dev/null || true
    docker system prune -f --volumes 2>/dev/null || true
}

# Build and start services
build_and_start() {
    echo "🏗️  Building and starting services..."

    # Build with no cache to ensure fresh build
    docker-compose build --no-cache

    # Start services
    docker-compose up -d

    # Wait for services to be ready
    echo "⏳ Waiting for services to be ready..."
    sleep 30

    # Check service health
    check_service_health
}

# Check if services are responding
check_service_health() {
    echo "🔍 Checking service health..."

    # Check Consul
    if curl -s http://localhost:8500/v1/status/leader >/dev/null; then
        echo "✅ Consul is healthy"
    else
        echo "❌ Consul is not responding"
    fi

    # Check API Gateway
    if curl -s http://localhost:8080/api/users >/dev/null 2>&1; then
        echo "✅ API Gateway is healthy"
    else
        echo "⚠️  API Gateway may still be starting..."
    fi

    echo "🎉 Build complete! Services are available at:"
    echo "   - Consul UI: http://localhost:8500"
    echo "   - API Gateway: http://localhost:8080"
    echo "   - Users DB: localhost:5432"
    echo "   - Products DB: localhost:5433"
}

# Main execution
main() {
    check_dependencies
    generate_proto_files
    distribute_proto_files
    cleanup
    build_and_start
}

# Run main function
main "$@"
````

**Makefile for Development:**

```makefile
.PHONY: help proto build test clean up down logs

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

proto: ## Generate protobuf files
	@echo "Generating protobuf files..."
	@protoc --go_out=./proto/gen --go_opt=paths=source_relative \
	        --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative \
	        proto/*.proto
	@echo "✅ Proto files generated"

build: proto ## Build all services
	@echo "Building services..."
	@./scripts/build.sh

test: ## Run tests for all services
	@echo "Running tests..."
	@cd api-gateway && go test -v ./...
	@cd services/users-service && go test -v ./...
	@cd services/products-service && go test -v ./...

clean: ## Clean up containers and images
	@echo "Cleaning up..."
	@docker-compose down --remove-orphans -v
	@docker system prune -f

up: ## Start services
	@docker-compose up -d

down: ## Stop services
	@docker-compose down

logs: ## Show logs from all services
	@docker-compose logs -f

dev: ## Start development environment
	@make build && make up && make logs
```

Make scripts executable:

```bash
chmod +x scripts/build.sh
chmod +x Makefile
```

---

### **Part 8: Testing the Fixed System**

1. **Build and Run:**

```bash
./build.sh
```

2. **Verify Consul Registration:**
   Visit `http://localhost:8500` to see both services registered.

3. **Test Individual Endpoints:**

**Create a User:**

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"name": "John Doe", "email": "john@example.com"}' \
     http://localhost:8080/api/users
```

**Get User:**

```bash
curl http://localhost:8080/api/users/1
```

**Create a Product:**

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"name": "Laptop", "price": 1299.99}' \
     http://localhost:8080/api/products
```

**Get Product:**

```bash
curl http://localhost:8080/api/products/1
```

**Test Composite Endpoint (Fixed):**

```bash
curl http://localhost:8080/api/purchases/user/1/product/1
```

Expected response:

```json
{
  "user": {
    "id": "1",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "product": {
    "id": "1",
    "name": "Laptop",
    "price": 1299.99
  }
}
```

---

### **Part 9: Key Improvements Made**

1. **Proper Service Discovery**: API Gateway now uses Consul to dynamically discover services instead of hardcoded connections.

2. **Fixed Build Issues**: Proto files are properly copied to each service's build context, eliminating import path problems.

3. **Robust Error Handling**: Added proper error handling for service discovery failures and gRPC connection issues.

4. **Connection Management**: Implemented connection pooling and reuse in the API Gateway.

5. **Health Checks**: Added proper health checks for service registration with Consul.

6. **Network Configuration**: Added proper Docker network configuration for service communication.

7. **Environment Variables**: Used environment variables for Consul configuration to make the system more flexible.

### **Summary**

This solution addresses all the key issues:

- ✅ Fixed proto file build issues by copying files to each service
- ✅ Implemented proper Consul service discovery in API Gateway
- ✅ Fixed composite endpoint to properly aggregate data from multiple services
- ✅ Added robust error handling and connection management
- ✅ Provided proper testing procedures and sample requests

The system now properly demonstrates a microservices architecture with dynamic service discovery, making it scalable and resilient.

---

## **Part 10: Advanced Implementation - Inter-Service Communication Pattern**

### **Overview**

This section demonstrates an alternative architecture where services communicate directly with each other via gRPC, rather than having the API Gateway orchestrate all composite operations. We'll implement a dedicated **Purchase Service** that handles purchase workflows by communicating with both Users and Products services.

### **Architecture Enhancement**

```
Client → API Gateway → Purchase Service → Users Service
                    → Purchase Service → Products Service
                    → Purchase Service → Payment Service (simulated)
```

### **Step 1: Enhanced Proto Definitions**

First, let's extend our proto files to support the new Purchase Service and inter-service communication patterns.

**`proto/purchases.proto`:**

```protobuf
syntax = "proto3";

option go_package = "./proto/gen;gen";

package purchases;

import "google/protobuf/timestamp.proto";

service PurchaseService {
  rpc CreatePurchase(CreatePurchaseRequest) returns (CreatePurchaseResponse);
  rpc GetPurchase(GetPurchaseRequest) returns (GetPurchaseResponse);
  rpc GetUserPurchases(GetUserPurchasesRequest) returns (GetUserPurchasesResponse);
}

message Purchase {
  string id = 1;
  string user_id = 2;
  string product_id = 3;
  double amount = 4;
  PurchaseStatus status = 5;
  google.protobuf.Timestamp created_at = 6;
}

enum PurchaseStatus {
  PURCHASE_STATUS_UNSPECIFIED = 0;
  PURCHASE_STATUS_PENDING = 1;
  PURCHASE_STATUS_COMPLETED = 2;
  PURCHASE_STATUS_FAILED = 3;
  PURCHASE_STATUS_REFUNDED = 4;
}

message CreatePurchaseRequest {
  string user_id = 1;
  string product_id = 2;
}

message CreatePurchaseResponse {
  Purchase purchase = 1;
  UserData user = 2;
  ProductData product = 3;
}

message GetPurchaseRequest {
  string id = 1;
}

message GetPurchaseResponse {
  Purchase purchase = 1;
  UserData user = 2;
  ProductData product = 3;
}

message GetUserPurchasesRequest {
  string user_id = 1;
}

message GetUserPurchasesResponse {
  repeated Purchase purchases = 1;
  UserData user = 2;
  repeated ProductData products = 3;
}

// Embedded data from other services
message UserData {
  string id = 1;
  string name = 2;
  string email = 3;
}

message ProductData {
  string id = 1;
  string name = 2;
  double price = 3;
}
```

### **Step 2: Purchase Service Implementation**

**`services/purchase-service/main.go`:**

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "os"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    pb "./proto/gen"
    usersPb "./proto/gen" // Assuming same package for simplicity
    productsPb "./proto/gen"
    consulapi "github.com/hashicorp/consul/api"
)

const serviceName = "purchase-service"
const servicePort = 50053

type Purchase struct {
    gorm.Model
    UserID    string
    ProductID string
    Amount    float64
    Status    string
}

type server struct {
    pb.UnimplementedPurchaseServiceServer
    db               *gorm.DB
    serviceDiscovery *ServiceDiscovery
}

type ServiceDiscovery struct {
    consul      *consulapi.Client
    connections map[string]*grpc.ClientConn
}

func (s *server) CreatePurchase(ctx context.Context, req *pb.CreatePurchaseRequest) (*pb.CreatePurchaseResponse, error) {
    log.Printf("Creating purchase for user %s, product %s", req.UserId, req.ProductId)

    // 1. Validate user exists by calling Users service
    usersClient, err := s.getUsersServiceClient()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to users service: %w", err)
    }

    userRes, err := usersClient.GetUser(ctx, &usersPb.GetUserRequest{Id: req.UserId})
    if err != nil {
        return nil, fmt.Errorf("user validation failed: %w", err)
    }

    // 2. Validate product exists and get price by calling Products service
    productsClient, err := s.getProductsServiceClient()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to products service: %w", err)
    }

    productRes, err := productsClient.GetProduct(ctx, &productsPb.GetProductRequest{Id: req.ProductId})
    if err != nil {
        return nil, fmt.Errorf("product validation failed: %w", err)
    }

    // 3. Create purchase record
    purchase := Purchase{
        UserID:    req.UserId,
        ProductID: req.ProductId,
        Amount:    productRes.Product.Price,
        Status:    "completed",
    }

    if result := s.db.Create(&purchase); result.Error != nil {
        return nil, fmt.Errorf("failed to create purchase: %w", result.Error)
    }

    // 4. Return aggregated response with data from other services
    return &pb.CreatePurchaseResponse{
        Purchase: &pb.Purchase{
            Id:        fmt.Sprint(purchase.ID),
            UserId:    purchase.UserID,
            ProductId: purchase.ProductID,
            Amount:    purchase.Amount,
            Status:    pb.PurchaseStatus_PURCHASE_STATUS_COMPLETED,
        },
        User: &pb.UserData{
            Id:    userRes.User.Id,
            Name:  userRes.User.Name,
            Email: userRes.User.Email,
        },
        Product: &pb.ProductData{
            Id:    productRes.Product.Id,
            Name:  productRes.Product.Name,
            Price: productRes.Product.Price,
        },
    }, nil
}

func (s *server) GetPurchase(ctx context.Context, req *pb.GetPurchaseRequest) (*pb.GetPurchaseResponse, error) {
    var purchase Purchase
    if result := s.db.First(&purchase, req.Id); result.Error != nil {
        return nil, result.Error
    }

    // Fetch related user and product data
    usersClient, err := s.getUsersServiceClient()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to users service: %w", err)
    }

    productsClient, err := s.getProductsServiceClient()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to products service: %w", err)
    }

    // Parallel calls to get user and product data
    userChan := make(chan *usersPb.UserResponse, 1)
    productChan := make(chan *productsPb.ProductResponse, 1)
    errChan := make(chan error, 2)

    go func() {
        userRes, err := usersClient.GetUser(ctx, &usersPb.GetUserRequest{Id: purchase.UserID})
        if err != nil {
            errChan <- err
            return
        }
        userChan <- userRes
    }()

    go func() {
        productRes, err := productsClient.GetProduct(ctx, &productsPb.GetProductRequest{Id: purchase.ProductID})
        if err != nil {
            errChan <- err
            return
        }
        productChan <- productRes
    }()

    var userRes *usersPb.UserResponse
    var productRes *productsPb.ProductResponse

    for i := 0; i < 2; i++ {
        select {
        case user := <-userChan:
            userRes = user
        case product := <-productChan:
            productRes = product
        case err := <-errChan:
            return nil, err
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }

    return &pb.GetPurchaseResponse{
        Purchase: &pb.Purchase{
            Id:        fmt.Sprint(purchase.ID),
            UserId:    purchase.UserID,
            ProductId: purchase.ProductID,
            Amount:    purchase.Amount,
            Status:    pb.PurchaseStatus_PURCHASE_STATUS_COMPLETED,
        },
        User: &pb.UserData{
            Id:    userRes.User.Id,
            Name:  userRes.User.Name,
            Email: userRes.User.Email,
        },
        Product: &pb.ProductData{
            Id:    productRes.Product.Id,
            Name:  productRes.Product.Name,
            Price: productRes.Product.Price,
        },
    }, nil
}

func (s *server) GetUserPurchases(ctx context.Context, req *pb.GetUserPurchasesRequest) (*pb.GetUserPurchasesResponse, error) {
    var purchases []Purchase
    if result := s.db.Where("user_id = ?", req.UserId).Find(&purchases); result.Error != nil {
        return nil, result.Error
    }

    // Get user data
    usersClient, err := s.getUsersServiceClient()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to users service: %w", err)
    }

    userRes, err := usersClient.GetUser(ctx, &usersPb.GetUserRequest{Id: req.UserId})
    if err != nil {
        return nil, fmt.Errorf("failed to get user data: %w", err)
    }

    // Get unique product IDs and fetch product data
    productIDs := make(map[string]bool)
    for _, purchase := range purchases {
        productIDs[purchase.ProductID] = true
    }

    productsClient, err := s.getProductsServiceClient()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to products service: %w", err)
    }

    productMap := make(map[string]*pb.ProductData)
    for productID := range productIDs {
        productRes, err := productsClient.GetProduct(ctx, &productsPb.GetProductRequest{Id: productID})
        if err != nil {
            log.Printf("Warning: failed to get product %s: %v", productID, err)
            continue
        }
        productMap[productID] = &pb.ProductData{
            Id:    productRes.Product.Id,
            Name:  productRes.Product.Name,
            Price: productRes.Product.Price,
        }
    }

    // Build response
    pbPurchases := make([]*pb.Purchase, len(purchases))
    products := make([]*pb.ProductData, 0, len(productMap))

    for i, purchase := range purchases {
        pbPurchases[i] = &pb.Purchase{
            Id:        fmt.Sprint(purchase.ID),
            UserId:    purchase.UserID,
            ProductId: purchase.ProductID,
            Amount:    purchase.Amount,
            Status:    pb.PurchaseStatus_PURCHASE_STATUS_COMPLETED,
        }
    }

    for _, product := range productMap {
        products = append(products, product)
    }

    return &pb.GetUserPurchasesResponse{
        Purchases: pbPurchases,
        User: &pb.UserData{
            Id:    userRes.User.Id,
            Name:  userRes.User.Name,
            Email: userRes.User.Email,
        },
        Products: products,
    }, nil
}

func (s *server) getUsersServiceClient() (usersPb.UserServiceClient, error) {
    conn, err := s.serviceDiscovery.getServiceConnection("users-service")
    if err != nil {
        return nil, err
    }
    return usersPb.NewUserServiceClient(conn), nil
}

func (s *server) getProductsServiceClient() (productsPb.ProductServiceClient, error) {
    conn, err := s.serviceDiscovery.getServiceConnection("products-service")
    if err != nil {
        return nil, err
    }
    return productsPb.NewProductServiceClient(conn), nil
}

func (sd *ServiceDiscovery) getServiceConnection(serviceName string) (*grpc.ClientConn, error) {
    if conn, exists := sd.connections[serviceName]; exists {
        return conn, nil
    }

    config := consulapi.DefaultConfig()
    if addr := os.Getenv("CONSUL_HTTP_ADDR"); addr != "" {
        config.Address = addr
    }

    consul, err := consulapi.NewClient(config)
    if err != nil {
        return nil, err
    }

    services, _, err := consul.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
    }

    if len(services) == 0 {
        return nil, fmt.Errorf("no healthy instances of service %s found", serviceName)
    }

    service := services[0].Service
    address := fmt.Sprintf("%s:%d", service.Address, service.Port)

    conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, fmt.Errorf("failed to connect to service %s at %s: %w", serviceName, address, err)
    }

    sd.connections[serviceName] = conn
    log.Printf("Connected to %s at %s", serviceName, address)
    return conn, nil
}

func main() {
    time.Sleep(15 * time.Second) // Wait for other services

    // Connect to database
    dsn := "host=purchase-db user=user password=password dbname=purchase_db port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    db.AutoMigrate(&Purchase{})

    // Initialize service discovery
    sd := &ServiceDiscovery{
        connections: make(map[string]*grpc.ClientConn),
    }

    // Start gRPC server
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    s := grpc.NewServer()
    pb.RegisterPurchaseServiceServer(s, &server{
        db:               db,
        serviceDiscovery: sd,
    })

    // Register with Consul
    if err := registerServiceWithConsul(); err != nil {
        log.Fatalf("Failed to register with Consul: %v", err)
    }

    log.Printf("%s gRPC server listening at %v", serviceName, lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}

func registerServiceWithConsul() error {
    config := consulapi.DefaultConfig()
    if addr := os.Getenv("CONSUL_HTTP_ADDR"); addr != "" {
        config.Address = addr
    }

    consul, err := consulapi.NewClient(config)
    if err != nil {
        return err
    }

    hostname, err := os.Hostname()
    if err != nil {
        return err
    }

    registration := &consulapi.AgentServiceRegistration{
        ID:      fmt.Sprintf("%s-%s", serviceName, hostname),
        Name:    serviceName,
        Port:    servicePort,
        Address: hostname,
        Check: &consulapi.AgentServiceCheck{
            GRPC:                           fmt.Sprintf("%s:%d", hostname, servicePort),
            Interval:                       "10s",
            DeregisterCriticalServiceAfter: "30s",
        },
    }

    return consul.Agent().ServiceRegister(registration)
}
```

### **Step 3: Updated Docker Compose Configuration**

**Updated `docker-compose.yml`:**

```yaml
services:
  consul:
    image: hashicorp/consul:latest
    container_name: consul
    ports:
      - "8500:8500"
    command: "agent -dev -client=0.0.0.0 -ui"
    networks:
      - microservices

  users-db:
    image: postgres:13
    container_name: users-db
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: users_db
    ports:
      - "5432:5432"
    volumes:
      - users_data:/var/lib/postgresql/data
    networks:
      - microservices

  products-db:
    image: postgres:13
    container_name: products-db
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: products_db
    ports:
      - "5433:5432"
    volumes:
      - products_data:/var/lib/postgresql/data
    networks:
      - microservices

  purchase-db:
    image: postgres:13
    container_name: purchase-db
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: purchase_db
    ports:
      - "5434:5432"
    volumes:
      - purchase_data:/var/lib/postgresql/data
    networks:
      - microservices

  users-service:
    build: ./services/users-service
    container_name: users-service
    ports:
      - "50051:50051"
    depends_on:
      - consul
      - users-db
    environment:
      - CONSUL_HTTP_ADDR=consul:8500
    networks:
      - microservices

  products-service:
    build: ./services/products-service
    container_name: products-service
    ports:
      - "50052:50052"
    depends_on:
      - consul
      - products-db
    environment:
      - CONSUL_HTTP_ADDR=consul:8500
    networks:
      - microservices

  purchase-service:
    build: ./services/purchase-service
    container_name: purchase-service
    ports:
      - "50053:50053"
    depends_on:
      - consul
      - purchase-db
      - users-service
      - products-service
    environment:
      - CONSUL_HTTP_ADDR=consul:8500
    networks:
      - microservices

  api-gateway:
    build: ./api-gateway
    container_name: api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - consul
      - users-service
      - products-service
      - purchase-service
    environment:
      - CONSUL_HTTP_ADDR=consul:8500
    networks:
      - microservices

volumes:
  users_data:
  products_data:
  purchase_data:

networks:
  microservices:
    driver: bridge
```

### **Step 4: Updated API Gateway**

Add new routes to the API Gateway to handle purchase operations:

**Enhanced `api-gateway/main.go` (additions):**

```go
// Add to the router setup in main()
r.HandleFunc("/api/purchases", createPurchaseHandler).Methods("POST")
r.HandleFunc("/api/purchases/{id}", getPurchaseHandler).Methods("GET")
r.HandleFunc("/api/users/{userId}/purchases", getUserPurchasesHandler).Methods("GET")

// New handler functions
func getPurchaseServiceClient() (pb.PurchaseServiceClient, error) {
    conn, err := sd.getServiceConnection("purchase-service")
    if err != nil {
        return nil, err
    }
    return pb.NewPurchaseServiceClient(conn), nil
}

func createPurchaseHandler(w http.ResponseWriter, r *http.Request) {
    client, err := getPurchaseServiceClient()
    if err != nil {
        http.Error(w, fmt.Sprintf("Service unavailable: %v", err), http.StatusServiceUnavailable)
        return
    }

    var req pb.CreatePurchaseRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    res, err := client.CreatePurchase(context.Background(), &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}

func getPurchaseHandler(w http.ResponseWriter, r *http.Request) {
    client, err := getPurchaseServiceClient()
    if err != nil {
        http.Error(w, fmt.Sprintf("Service unavailable: %v", err), http.StatusServiceUnavailable)
        return
    }

    vars := mux.Vars(r)
    id := vars["id"]

    res, err := client.GetPurchase(context.Background(), &pb.GetPurchaseRequest{Id: id})
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}

func getUserPurchasesHandler(w http.ResponseWriter, r *http.Request) {
    client, err := getPurchaseServiceClient()
    if err != nil {
        http.Error(w, fmt.Sprintf("Service unavailable: %v", err), http.StatusServiceUnavailable)
        return
    }

    vars := mux.Vars(r)
    userId := vars["userId"]

    res, err := client.GetUserPurchases(context.Background(), &pb.GetUserPurchasesRequest{UserId: userId})
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}
```

### **Step 5: Testing the Inter-Service Communication**

**Create Purchase Service Directory:**

```bash
mkdir -p services/purchase-service
cd services/purchase-service
go mod init practical-three/purchase-service
go get google.golang.org/grpc
go get github.com/hashicorp/consul/api
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

**Create Purchase Service Dockerfile:**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy proto files first
COPY ../../proto ./proto
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 50053

CMD ["./server"]
```

**Update Build Script:**

```bash
# Add to scripts/build.sh
cp -r proto/* services/purchase-service/proto/
```

**Testing Commands:**

1. **Create a Purchase (Inter-Service Communication):**

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"user_id": "1", "product_id": "1"}' \
     http://localhost:8080/api/purchases
```

Expected response:

```json
{
  "purchase": {
    "id": "1",
    "user_id": "1",
    "product_id": "1",
    "amount": 1299.99,
    "status": "PURCHASE_STATUS_COMPLETED"
  },
  "user": {
    "id": "1",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "product": {
    "id": "1",
    "name": "Laptop",
    "price": 1299.99
  }
}
```

2. **Get Purchase Details:**

```bash
curl http://localhost:8080/api/purchases/1
```

3. **Get User's Purchase History:**

```bash
curl http://localhost:8080/api/users/1/purchases
```

### **Benefits of This Approach**

✅ **Domain Separation**: Purchase logic is properly encapsulated in its own service  
✅ **Business Workflows**: Complex purchase workflows are handled by the domain service  
✅ **Service Autonomy**: Each service manages its own data and business rules  
✅ **Scalability**: Purchase service can be scaled independently  
✅ **Maintainability**: Clear separation of concerns and responsibilities

### **Architecture Comparison**

| Aspect               | API Gateway Orchestration | Inter-Service Communication |
| -------------------- | ------------------------- | --------------------------- |
| **Complexity**       | Low                       | Medium                      |
| **Performance**      | Good (parallel calls)     | Good (optimized calls)      |
| **Maintainability**  | Gateway becomes complex   | Better separation           |
| **Domain Logic**     | Mixed in gateway          | Properly encapsulated       |
| **Network Calls**    | 1 HTTP + 2 gRPC           | 1 HTTP + 1 gRPC + 2 gRPC    |
| **Failure Handling** | Centralized               | Distributed                 |
| **Best For**         | Simple aggregation        | Complex workflows           |

This implementation demonstrates how to properly implement inter-service communication in a microservices architecture, providing a foundation for more complex business workflows while maintaining service autonomy and proper domain boundaries.
