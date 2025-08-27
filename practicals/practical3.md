### **Module Practical: WEB303 Microservices & Serverless Applications**

## **Practical 3: Full-Stack Microservices with gRPC, Databases, and Service Discovery**

### **Objective**

In this practical, you will build a complete microservices ecosystem from the ground up. This session will guide you through creating two independent services that communicate via **gRPC**, manage their own data with a **PostgreSQL** database and the **GORM** ORM, and register themselves with **Consul** for service discovery. An **API Gateway** will serve as the single entry point, translating external **HTTP** requests into internal gRPC calls. A new endpoint will be created to demonstrate how the API Gateway can query and aggregate data from multiple services for a single client request.

This exercise provides a comprehensive, hands-on understanding of how to build a scalable, decoupled, and resilient microservice architecture. 🚀

### **Learning Outcomes Supported**

  * **Learning Outcome 2:** Design and implement microservices using gRPC and Protocol Buffers for efficient inter-service communication.
  * **Learning Outcome 4:** Implement data persistence and state management strategies in microservices and serverless applications.
  * **Learning Outcome 8:** Implement observability solutions for microservices and serverless applications, including distributed tracing, metrics, and logging.

-----

### **The Architecture**

This practical will guide you through building the following system:

  * **API Gateway**: The public-facing entry point that receives all incoming HTTP requests. It acts as a smart router, converting these requests into gRPC messages to be sent to the appropriate internal service. It will also feature an endpoint to aggregate data from both services.
  * **Service Discovery (Consul)**: A central registry, or "phone book" 📖, that keeps track of all running services and their locations. This allows our services to find each other without being tightly coupled.
  * **Microservices (`users-service` & `products-service`)**: Two independent services, each with its own dedicated PostgreSQL database. They will expose a gRPC server for internal communication and register themselves with Consul upon startup.
  * **Databases**: Each microservice will have its own PostgreSQL database, managed by Docker, ensuring complete data isolation.

-----
### **Submission Instructions & Requirements**

1. You are to FIX the ```api-gateway/main.go``` and ensure that the http service is properly engaging the Consul Service to discover the relevant services with respect to the endpoint.

2. Present situation the api-gateway is directly calling the user/product services directly via defining it's ports. You should realise that the composite endpoint for aggregating the user and product services are not done properly. You are to FIX this as well. Both services are currently not communicating with each other.

3. Upload your files to a seperate repository and submit your work to your submission repository as usual.

4. include the screenshots of your sample requests either via cUrl/Postman.
-----
### **Part 1: Prerequisites and Project Setup**

Before we start, ensure you have **Go (1.18+)** and **Docker** installed on your system.

1.  **Install gRPC and Protobuf Tools:**
    First, you'll need the Protocol Buffers compiler (`protoc`) and the Go plugins for gRPC.

    ```bash
    # Install the Go plugins for protobuf and gRPC
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

    # Ensure your Go bin directory is in your system's PATH
    # Add this to your ~/.bashrc, ~/.zshrc, or equivalent shell profile:
    export PATH="$PATH:$(go env GOPATH)/bin"
    ```

2.  **Create the Project Structure:**
    Let's set up a clean directory structure for our project.

    ```bash
    mkdir practical-three
    cd practical-three
    mkdir -p proto/gen
    mkdir api-gateway
    mkdir services
    mkdir services/users-service
    mkdir services/products-service
    ```

3.  **Define the Service Contracts (`.proto` files):**
    We'll use Protocol Buffers to define the structure of our gRPC services. Create the following files inside the `proto` directory.

    **`proto/users.proto`:**

    ```protobuf
    syntax = "proto3";

    option go_package = "practical-three/proto/gen;gen";

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

    option go_package = "practical-three/proto/gen;gen";

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

4.  **Generate the Go gRPC Code:**
    Now, from the root of your `practical-three` directory, run the `protoc` command to generate the necessary Go files.

    ```bash
    protoc --go_out=./proto/gen --go_opt=paths=source_relative \
        --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative \
        proto/*.proto
    ```

    This will populate your `proto/gen` directory with the generated Go code.

-----

### **Part 2: Orchestration with Docker Compose**

We will use a `docker-compose.yml` file to define and manage our entire application stack, including our services, databases, and Consul. Create this file in the root of your project.

**`docker-compose.yml`:**

```yaml
services:
  consul:
    image: hashicorp/consul:latest
    container_name: consul
    ports:
      - "8500:8500"
    command: "agent -dev -client=0.0.0.0 -ui"

  users-db:
    image: postgres:13
    container_name: users-db
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: users_db
    ports:
      - "5432:5432" # Exposing for local inspection
    volumes:
      - users_data:/var/lib/postgresql/data

  products-db:
    image: postgres:13
    container_name: products-db
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: products_db
    ports:
      - "5433:5432" # Mapping to a different host port
    volumes:
      - products_data:/var/lib/postgresql/data

  users-service:
    build: ./services/users-service
    container_name: users-service
    ports:
      - "50051:50051"
    depends_on:
      - consul
      - users-db

  products-service:
    build: ./services/products-service
    container_name: products-service
    ports:
      - "50052:50052"
    depends_on:
      - consul
      - products-db

  api-gateway:
    build: ./api-gateway
    container_name: api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - consul
      - users-service
      - products-service

volumes:
  users_data:
  products_data:
```

-----

### **Part 3: Implementing the `users-service`**

Now let's build our first microservice. This service will be a gRPC server that connects to its own PostgreSQL database.

1.  **Initialize Go Module and Dependencies:**
    Navigate to the `users-service` directory and run the following commands:

    ```bash
    cd services/users-service
    go mod init practical-three/users-service
    go get google.golang.org/grpc
    go get github.com/hashicorp/consul/api
    go get gorm.io/gorm
    go get gorm.io/driver/postgres
    ```

2.  **Create the `main.go` file:**
    This file will contain the logic for connecting to the database, starting the gRPC server, and registering with Consul.

    ```go
    // services/users-service/main.go
    package main

    import (
        "context"
        "fmt"
        "log"
        "net"
        "os"

        "google.golang.org/grpc"
        "gorm.io/driver/postgres"
        "gorm.io/gorm"

        pb "practical-three/proto/gen"
        consulapi "github.com/hashicorp/consul/api"
    )

    const serviceName = "users-service"
    const servicePort = 50051

    // GORM model for our User
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
        // 1. Connect to the database
        dsn := "host=users-db user=user password=password dbname=users_db port=5432 sslmode=disable"
        db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        db.AutoMigrate(&User{})

        // 2. Start the gRPC server
        lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
        if err != nil {
            log.Fatalf("Failed to listen: %v", err)
        }
        s := grpc.NewServer()
        pb.RegisterUserServiceServer(s, &server{db: db})

        // 3. Register with Consul
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
        }

        return consul.Agent().ServiceRegister(registration)
    }
    ```

3.  **Create a `Dockerfile` for the service:**
    Place this in the `services/users-service` directory.

    ```dockerfile
    # services/users-service/Dockerfile
    FROM golang:1.24.2-alpine AS builder
    WORKDIR /app
    COPY go.mod ./
    COPY go.sum ./
    RUN go mod download
    COPY . .
    # Note the path to the main.go file
    RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./main.go

    FROM alpine:latest
    WORKDIR /app
    COPY --from=builder /app/server .
    EXPOSE 50051
    CMD ["/app/server"]
    ```

-----

### **Part 4: Implementing the `products-service`**

Next, we'll create the `products-service`. It follows the same pattern as the `users-service`, but manages products instead.

1.  **Initialize Go Module and Dependencies:**
    Navigate to the `products-service` directory.

    ```bash
    cd ../products-service
    go mod init practical-three/products-service
    go get google.golang.org/grpc
    go get github.com/hashicorp/consul/api
    go get gorm.io/gorm
    go get gorm.io/driver/postgres
    ```

2.  **Create the `main.go` file:**
    This file will handle the product logic.

    ```go
    // services/products-service/main.go
    package main

    import (
        "context"
        "fmt"
        "log"
        "net"
        "os"

        "google.golang.org/grpc"
        "gorm.io/driver/postgres"
        "gorm.io/gorm"

        pb "practical-three/proto/gen"
        consulapi "github.com/hashicorp/consul/api"
    )

    const serviceName = "products-service"
    const servicePort = 50052

    // GORM model for our Product
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
        // 1. Connect to the database
        dsn := "host=products-db user=user password=password dbname=products_db port=5432 sslmode=disable"
        db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        db.AutoMigrate(&Product{})

        // 2. Start the gRPC server
        lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
        if err != nil {
            log.Fatalf("Failed to listen: %v", err)
        }
        s := grpc.NewServer()
        pb.RegisterProductServiceServer(s, &server{db: db})

        // 3. Register with Consul
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
        }

        return consul.Agent().ServiceRegister(registration)
    }
    ```

3.  **Create a `Dockerfile` for the service:**
    Place this in the `services/products-service` directory.

    ```dockerfile
    # services/products-service/Dockerfile
    FROM golang:1.24.2-alpine AS builder
    WORKDIR /app
    COPY go.mod ./
    COPY go.sum ./
    RUN go mod download
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./main.go

    FROM alpine:latest
    WORKDIR /app
    COPY --from=builder /app/server .
    EXPOSE 50052
    CMD ["/app/server"]
    ```

-----

### **Part 5: Implementing the API Gateway**

The API Gateway will be an HTTP server that acts as a client to our gRPC services.

1.  **Initialize Go Module and Dependencies:**
    Navigate to the `api-gateway` directory.

    ```bash
    cd ../../api-gateway
    go mod init practical-three/api-gateway
    go get google.golang.org/grpc
    go get github.com/gorilla/mux
    ```

2.  **Create the `main.go` file:**
    This gateway will discover services using Consul and translate HTTP requests into gRPC calls. It will also have an endpoint to aggregate data from both services.

    ```go
    // api-gateway/main.go
    package main

    import (
    	"context"
    	"encoding/json"
    	"log"
    	"net/http"
    	"sync"

    	"github.com/gorilla/mux"
    	"google.golang.org/grpc"
    	"google.golang.org/grpc/credentials/insecure"

    	pb "practical-three/proto/gen"
    )

    var usersClient pb.UserServiceClient
    var productsClient pb.ProductServiceClient

    // A struct to hold the aggregated data
    type UserPurchaseData struct {
    	User    *pb.User    `json:"user"`
    	Product *pb.Product `json:"product"`
    }

    func main() {
    	// Connect to the users-service
    	userConn, err := grpc.Dial("users-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    	if err != nil {
    		log.Fatalf("Did not connect to users-service: %v", err)
    	}
    	defer userConn.Close()
    	usersClient = pb.NewUserServiceClient(userConn)

    	// Connect to the products-service
    	productConn, err := grpc.Dial("products-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
    	if err != nil {
    		log.Fatalf("Did not connect to products-service: %v", err)
    	}
    	defer productConn.Close()
    	productsClient = pb.NewProductServiceClient(productConn)

    	r := mux.NewRouter()
    	// User routes
    	r.HandleFunc("/api/users", createUserHandler).Methods("POST")
    	r.HandleFunc("/api/users/{id}", getUserHandler).Methods("GET")
    	// Product routes
    	r.HandleFunc("/api/products", createProductHandler).Methods("POST")
    	r.HandleFunc("/api/products/{id}", getProductHandler).Methods("GET")

    	// The new endpoint to get combined data
    	r.HandleFunc("/api/purchases/user/{userId}/product/{productId}", getPurchaseDataHandler).Methods("GET")

    	log.Println("API Gateway listening on port 8080...")
    	http.ListenAndServe(":8080", r)
    }

    // User Handlers
    func createUserHandler(w http.ResponseWriter, r *http.Request) {
    	var req pb.CreateUserRequest
    	json.NewDecoder(r.Body).Decode(&req)
    	res, err := usersClient.CreateUser(context.Background(), &req)
    	if err != nil {
    		http.Error(w, err.Error(), http.StatusInternalServerError)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	json.NewEncoder(w).Encode(res.User)
    }

    func getUserHandler(w http.ResponseWriter, r *http.Request) {
    	vars := mux.Vars(r)
    	id := vars["id"]
    	res, err := usersClient.GetUser(context.Background(), &pb.GetUserRequest{Id: id})
    	if err != nil {
    		http.Error(w, err.Error(), http.StatusNotFound)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	json.NewEncoder(w).Encode(res.User)
    }

    // Product Handlers
    func createProductHandler(w http.ResponseWriter, r *http.Request) {
    	var req pb.CreateProductRequest
    	json.NewDecoder(r.Body).Decode(&req)
    	res, err := productsClient.CreateProduct(context.Background(), &req)
    	if err != nil {
    		http.Error(w, err.Error(), http.StatusInternalServerError)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	json.NewEncoder(w).Encode(res.Product)
    }

    func getProductHandler(w http.ResponseWriter, r *http.Request) {
    	vars := mux.Vars(r)
    	id := vars["id"]
    	res, err := productsClient.GetProduct(context.Background(), &pb.GetProductRequest{Id: id})
    	if err != nil {
    		http.Error(w, err.Error(), http.StatusNotFound)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	json.NewEncoder(w).Encode(res.Product)
    }

    // New handler for combined data
    func getPurchaseDataHandler(w http.ResponseWriter, r *http.Request) {
    	vars := mux.Vars(r)
    	userId := vars["userId"]
    	productId := vars["productId"]

    	var wg sync.WaitGroup
    	var user *pb.User
    	var product *pb.Product
    	var userErr, productErr error

    	wg.Add(2)

    	go func() {
    		defer wg.Done()
    		res, err := usersClient.GetUser(context.Background(), &pb.GetUserRequest{Id: userId})
    		if err != nil {
    			userErr = err
    			return
    		}
    		user = res.User
    	}()

    	go func() {
    		defer wg.Done()
    		res, err := productsClient.GetProduct(context.Background(), &pb.GetProductRequest{Id: productId})
    		if err != nil {
    			productErr = err
    			return
    		}
    		product = res.Product
    	}()

    	wg.Wait()

    	if userErr != nil || productErr != nil {
    		http.Error(w, "Could not retrieve all data", http.StatusNotFound)
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

3.  **Create a `Dockerfile` for the gateway:**
    Place this in the `api-gateway` directory.

    ```dockerfile
    # api-gateway/Dockerfile
    FROM golang:1.24.2-alpine AS builder
    WORKDIR /app
    COPY go.mod ./
    COPY go.sum ./
    RUN go mod download
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./main.go

    FROM alpine:latest
    WORKDIR /app
    COPY --from=builder /app/server .
    EXPOSE 8080
    CMD ["/app/server"]
    ```

-----

### **Part 6: Running and Testing the System ✅**

You're now ready to run the entire system\!

1.  **Build and Start the Containers:**
    From the root of your `practical-three` directory, run:

    ```bash
    docker-compose up --build
    ```

    This command will build the Docker images for each of your services and start all the containers. You should see logs from all services, indicating that they are running.

2.  **Verify with Consul:**
    Open your web browser and navigate to `http://localhost:8500`. You should see the Consul UI, with both `users-service` and `products-service` registered and healthy.

3.  **Test the API Gateway:**
    Use `curl` or an API client like Postman to send requests to your API Gateway.

    **Create a new user:**

    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{"name": "Jane Doe", "email": "jane.doe@example.com"}' \
         http://localhost:8080/api/users
    ```

    You should get a JSON response with the newly created user's details.

    **Retrieve the user:**

    ```bash
    curl http://localhost:8080/api/users/1
    ```

    This should return the details of the user you just created.

    **Create a new product:**

    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{"name": "Laptop", "price": 1200.50}' \
         http://localhost:8080/api/products
    ```

    This will return the details of the newly created product.

    **Retrieve the product:**

    ```bash
    curl http://localhost:8080/api/products/1
    ```

    This will return the product details.

    **Retrieve the combined purchase data:**

    ```bash
    curl http://localhost:8080/api/purchases/user/1/product/1
    ```

    This will return a JSON object containing both the user and product details.
