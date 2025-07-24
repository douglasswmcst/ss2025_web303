# **Module Practical: WEB303 Microservices & Serverless Applications**

## **Practical 1: From Foundational Setup to Inter-Service Communication**

This comprehensive practical guides you through setting up your environment and then building your first multi-container microservice application. Part 1 covers the foundational tools, and Part 2 uses those tools to create two services that communicate using gRPC and run together with Docker Compose.

Reference to module descriptor practical 1: Set up a development environment for microservices using Go, gRPC, and Docker.

This practical supports the following module learning outcomes:

  * **Learning Outcome 2:** Design and implement microservices using gRPC and Protocol Buffers for efficient inter-service communication.
  * **Learning Outcome 6:** Deploy microservices to Kubernetes, implementing various deployment strategies and managing certificates. (Note: Docker Compose is a foundational step towards understanding container orchestration like Kubernetes).
  * **Learning Outcome 1 (supported):** Explain the fundamental concepts of microservices and serverless architectures, including their benefits, trade-offs, and appropriate use cases.

-----

## **Part 1: Foundational Development Environment Setup (Duration: 1 Hour)**

**Objective:** This part ensures your machine is ready for microservice development by installing and configuring Go, Protocol Buffers, and Docker.

### **1.1: Installing and Configuring Go**

Go is the programming language we will use for our microservices.

1.  **Download Go:** Navigate to the official Go downloads page ([https://go.dev/dl/](https://go.dev/dl/)) and download the installer for your operating system.
2.  **Install Go:** Run the installer. It will typically install to `C:\Go` on Windows or `/usr/local/go` on macOS/Linux.
3.  **Configure Environment Variables:** The installer usually handles this. To verify, open a **new** terminal and run `go version`. You should see the installed version number.
4.  **Verify Installation:** In your terminal, run the following commands:
    ```bash
    go version
    go env
    ```
    **Expected Outcome:** You will see the Go version and a list of Go's environment variables.

### **1.2: Installing Protocol Buffers & gRPC Tools**

Protocol Buffers (Protobuf) is a language-agnostic data serialization format we use to define our service contracts.

1.  **Install Protobuf Compiler (`protoc`):** Go to the Protobuf GitHub Releases page ([https://github.com/protocolbuffers/protobuf/releases](https://github.com/protocolbuffers/protobuf/releases)), download the `protoc-*-*.zip` for your OS, and place the `protoc` binary in a directory that is part of your system's PATH.
2.  **Install Go Plugins:** These tools generate Go code from `.proto` files. Run the following in your terminal:
    ```bash
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
    ```
3.  **Update PATH:** Make sure your Go binaries path is in your system's PATH. Add this line to your shell profile (`.bash_profile`, `.zshrc`, etc.) if it isn't already there:
    ```bash
    export PATH="$PATH:$(go env GOPATH)/bin"
    ```

### **1.3: Installing and Verifying Docker**

Docker allows us to package our services into portable containers.

1.  **Install Docker Desktop:** Download and install Docker Desktop for your machine from the official website ([https://www.docker.com/products/docker-desktop/](https://www.docker.com/products/docker-desktop/)).
2.  **Start Docker Desktop:** Launch the application. A whale icon in your menu bar or system tray will indicate it's running.
3.  **Verify Docker Installation:** Open a terminal and run the `hello-world` container:
    ```bash
    docker run hello-world
    ```
    **Expected Outcome:** 🐳 Docker will download and run the image, printing a "Hello from Docker\!" message. This confirms your Docker setup is working correctly.

-----

## **Part 2: Building and Orchestrating Communicating Microservices (Duration: 1 Hour)**

**Objective:** In this part, you will build and run a multi-container application. You will create a `greeter-service` that communicates with a `time-service` using gRPC, all managed by Docker Compose.

### **2.1: Project Setup & Service Contracts**

1.  **Create Project Structure:** Open your terminal and create the following directory structure.

    ```bash
    mkdir practical-one
    cd practical-one
    mkdir -p proto/gen
    mkdir greeter-service
    mkdir time-service
    ```

2.  **Define the `Time` Service (`proto/time.proto`):** This service returns the current time.

    ```protobuf
    syntax = "proto3";

    option go_package = "practical-one/proto/gen;gen";

    package time;

    service TimeService {
      rpc GetTime(TimeRequest) returns (TimeResponse);
    }

    message TimeRequest {}

    message TimeResponse {
      string current_time = 1;
    }
    ```

3.  **Define the `Greeter` Service (`proto/greeter.proto`):** This service provides a greeting.

    ```protobuf
    syntax = "proto3";

    option go_package = "practical-one/proto/gen;gen";

    package greeter;

    service GreeterService {
      rpc SayHello(HelloRequest) returns (HelloResponse);
    }

    message HelloRequest {
      string name = 1;
    }

    message HelloResponse {
      string message = 1;
    }
    ```

4.  **Generate Go Code:** Run this `protoc` command from the root `practical-one` directory.

    ```bash
    protoc --go_out=./proto/gen --go_opt=paths=source_relative \
        --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative \
        proto/*.proto
    ```

    **Expected Outcome:** 🔎 Your `proto/gen` directory will now contain four new `.go` files.

### **2.2: Implementing the Microservices in Go**

1.  **Implement the `time-service`:** This is our first gRPC server.

      * Navigate to the `time-service` folder: `cd time-service`
      * Initialize the Go module: `go mod init practical-one/time-service`
      * Get dependencies: `go get google.golang.org/grpc`
      * Create `main.go` with the following code:
        ```go
        package main

        import (
        	"context"
        	"log"
        	"net"
        	"time"
        	pb "practical-one/proto/gen"
        	"google.golang.org/grpc"
        )

        type server struct {
        	pb.UnimplementedTimeServiceServer
        }

        func (s *server) GetTime(ctx context.Context, in *pb.TimeRequest) (*pb.TimeResponse, error) {
        	log.Printf("Received request for time")
        	currentTime := time.Now().Format(time.RFC3339)
        	return &pb.TimeResponse{CurrentTime: currentTime}, nil
        }

        func main() {
        	lis, err := net.Listen("tcp", ":50052")
        	if err != nil {
        		log.Fatalf("failed to listen: %v", err)
        	}
        	s := grpc.NewServer()
        	pb.RegisterTimeServiceServer(s, &server{})
        	log.Printf("Time service listening at %v", lis.Addr())
        	if err := s.Serve(lis); err != nil {
        		log.Fatalf("failed to serve: %v", err)
        	}
        }
        ```
      * Return to the root directory: `cd ..`

2.  **Implement the `greeter-service`:** This service is both a gRPC server and a client.

      * Navigate to the `greeter-service` folder: `cd greeter-service`
      * Initialize the Go module: `go mod init practical-one/greeter-service`
      * Get dependencies: `go get google.golang.org/grpc`
      * Create `main.go` with the following code:
        ```go
        package main

        import (
        	"context"
        	"fmt"
        	"log"
        	"net"
        	pb "practical-one/proto/gen"
        	"google.golang.org/grpc"
        	"google.golang.org/grpc/credentials/insecure"
        )

        type server struct {
        	pb.UnimplementedGreeterServiceServer
        	timeClient pb.TimeServiceClient // Client to call the time-service
        }

        func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
        	log.Printf("Received SayHello request for: %v", in.GetName())
        	timeReq := &pb.TimeRequest{}
        	timeRes, err := s.timeClient.GetTime(ctx, timeReq)
        	if err != nil {
        		log.Printf("Failed to call time-service: %v", err)
        		return nil, err
        	}
        	message := fmt.Sprintf("Hello %s! The current time is %s", in.GetName(), timeRes.GetCurrentTime())
        	return &pb.HelloResponse{Message: message}, nil
        }

        func main() {
        	// Address 'time-service:50052' matches the service name in docker-compose.yml
        	conn, err := grpc.Dial("time-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
        	if err != nil {
        		log.Fatalf("did not connect to time-service: %v", err)
        	}
        	defer conn.Close()
        	timeClient := pb.NewTimeServiceClient(conn)

        	lis, err := net.Listen("tcp", ":50051")
        	if err != nil {
        		log.Fatalf("failed to listen: %v", err)
        	}
        	s := grpc.NewServer()
        	pb.RegisterGreeterServiceServer(s, &server{timeClient: timeClient})
        	log.Printf("Greeter service listening at %v", lis.Addr())
        	if err := s.Serve(lis); err != nil {
        		log.Fatalf("failed to serve: %v", err)
        	}
        }
        ```
      * Return to the root directory: `cd ..`

### **2.3: Containerization and Orchestration**

1.  **Create Dockerfile for `time-service` (`time-service/Dockerfile`):**

    ```dockerfile
    FROM golang:1.22-alpine AS builder
    WORKDIR /app
    COPY go.mod ./
    COPY go.sum ./
    RUN go mod download
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./time-service/main.go

    FROM alpine:latest
    WORKDIR /app
    COPY --from=builder /app/server .
    EXPOSE 50052
    CMD ["/app/server"]
    ```

2.  **Create Dockerfile for `greeter-service` (`greeter-service/Dockerfile`):**

    ```dockerfile
    FROM golang:1.22-alpine AS builder
    WORKDIR /app
    COPY go.mod ./
    COPY go.sum ./
    RUN go mod download
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./greeter-service/main.go

    FROM alpine:latest
    WORKDIR /app
    COPY --from=builder /app/server .
    EXPOSE 50051
    CMD ["/app/server"]
    ```

3.  **Create the `docker-compose.yml` File:** In the root `practical-one` directory, create this file to manage our services.

    ```yaml
    version: '3.8'

    services:
      time-service:
        build:
          context: .
          dockerfile: time-service/Dockerfile
        hostname: time-service

      greeter-service:
        build:
          context: .
          dockerfile: greeter-service/Dockerfile
        hostname: greeter-service
        ports:
          - "50051:50051"
        depends_on:
          - time-service
    ```

### **2.4: Run and Verify\! 🚀**

1.  **Run Docker Compose:** From the root of your `practical-one` directory, run:

    ```bash
    docker-compose up --build
    ```

    **Expected Outcome:** Docker will build and start both containers. You will see logs from both services in your terminal.

2.  **Test the Endpoint:** To test the flow, we'll use `grpcurl`.

      * **Install `grpcurl`:** If you don't have it, you can install it via Homebrew (`brew install grpcurl`) or other package managers.
      * **Open a NEW terminal.**
      * **Make the gRPC call:**
        ```bash
        grpcurl -plaintext \
            -import-path ./proto -proto greeter.proto \
            -d '{"name": "WEB303 Student"}' \
            0.0.0.0:50051 greeter.GreeterService/SayHello
        ```

**Final Expected Outcome:** ✅ You will receive a JSON response in your terminal:

```json
{
  "message": "Hello WEB303 Student! The current time is 2025-07-24T09:45:00Z"
}
```

-----

### **Conclusion and Submission**

Congratulations\! You have set up your development environment, built two distinct microservices, containerized them, and orchestrated their execution and communication with Docker Compose.

**Submission:** Please show your tutor the final output from the `grpcurl` command and the corresponding logs from `docker-compose`. Be prepared to explain how the `greeter-service` was able to find and call the `time-service` within the Docker network.