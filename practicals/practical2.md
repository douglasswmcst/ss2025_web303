# **Module Practical: WEB303 Microservices & Serverless Applications**

## **Practical 2: API Gateway with Service Discovery**

You'll create a small ecosystem with two independent microservices—a `Users Service` and a `Products Service`. These services won't know about each other directly. Instead, they'll announce their presence to a central "phone book" called **Consul** (our service registry) 📖.

To manage all incoming traffic, you'll build a smart "receptionist"—an **API Gateway** 👨‍💼—that consults this phone book to dynamically route requests to the correct, healthy service.

By the end, you'll have a hands-on understanding of how this pattern allows you to update, restart, or scale individual services without reconfiguring the other parts of your system. It’s the foundation for building applications that are easy to maintain and resilient to failure.

This practical supports the following module learning outcomes:

- **Learning Outcome 2:** Design and implement microservices using gRPC and Protocol Buffers for efficient inter-service communication.
- **Learning Outcome 8:** Implement observability solutions for microservices and serverless applications, including distributed tracing, metrics, and logging.

---

## **Submission Instructions**

- You must maintain your own public GitHub repository for all practical work and report submissions related to this module.
- All code, configuration files, and documentation for each practical should be committed to your repository.
- For each practical, include a clear README describing your approach, steps taken, and any challenges encountered.
- When submitting your work, provide the link to your public GitHub repository to your tutor as instructed.
- Ensure your repository is well-organized and up-to-date before submission.

---

### The Architecture

Our goal is a scalable system where services can be added, removed, or restarted without reconfiguring the other parts.

- **API Gateway:** The single entry point for all external requests. It's a "smart" reverse proxy.
- **Service Discovery (Consul):** A central registry, or "phone book" 📖, that keeps track of all running services and their health.
- **Microservices:** Two independent services, `users-service` and `products-service`, that register themselves with Consul on startup.

## Prerequisites and Installation

Before starting this practical, you'll need to install several tools and libraries. Follow the installation instructions for your operating system:

### 1. Go Programming Language (Version 1.18+)

**Windows:**

1. Download Go from the official website: https://golang.org/dl/
2. Run the installer (.msi file) and follow the installation wizard
3. Add Go to your PATH (the installer usually does this automatically)
4. Open Command Prompt or PowerShell and verify: `go version`

**macOS:**

```bash
# Using Homebrew (recommended)
brew install go

# Or download from https://golang.org/dl/ and run the .pkg installer
```

**Linux (Ubuntu/Debian):**

```bash
# Remove any existing Go installation
sudo rm -rf /usr/local/go

# Download and install Go (replace with latest version)
wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add to PATH in ~/.bashrc or ~/.zshrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

**Verify Go Installation:**

```bash
go version
# Should output: go version go1.x.x [os/arch]

# Check Go environment
go env GOPATH
go env GOROOT
```

### 2. Docker and Docker Compose

**Windows:**

1. Download Docker Desktop from: https://www.docker.com/products/docker-desktop
2. Run the installer and follow the setup wizard
3. Start Docker Desktop after installation
4. Enable WSL 2 integration if prompted

**macOS:**

```bash
# Using Homebrew (recommended)
brew install --cask docker

# Or download Docker Desktop from https://www.docker.com/products/docker-desktop
```

**Linux (Ubuntu/Debian):**

```bash
# Update package index
sudo apt-get update

# Install required packages
sudo apt-get install apt-transport-https ca-certificates curl gnupg lsb-release

# Add Docker's official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Add Docker repository
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Add your user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

**Verify Docker Installation:**

```bash
docker --version
# Should output: Docker version x.x.x

docker run hello-world
# Should download and run a test container successfully
```

### 3. Required Go Libraries

The following Go libraries will be automatically downloaded when you run `go get` commands in each service:

**Chi Router (HTTP Router):**

- **Package:** `github.com/go-chi/chi/v5`
- **Purpose:** Lightweight HTTP router for building REST APIs
- **Features:** Middleware support, URL parameters, route groups

**Consul API Client:**

- **Package:** `github.com/hashicorp/consul/api`
- **Purpose:** Go client library for HashiCorp Consul
- **Features:** Service registration, health checks, service discovery

### 4. Development Tools (Optional but Recommended)

**cURL (for testing APIs):**

```bash
# Windows (using chocolatey)
choco install curl

# macOS (usually pre-installed, or via Homebrew)
brew install curl

# Linux (usually pre-installed, or)
sudo apt-get install curl
```

**Postman (Alternative to cURL):**

- Download from: https://www.postman.com/downloads/
- Useful for testing and debugging APIs with a graphical interface

**Go Language Extensions (for your IDE):**

- **VS Code:** Install the "Go" extension by Google
- **JetBrains GoLand:** Full-featured Go IDE
- **Vim/Neovim:** Install vim-go plugin

### 5. Directory Structure Setup

Create your workspace directory:

```bash
# Create main project directory
mkdir go-microservices-demo
cd go-microservices-demo

# Verify you're in the right directory
pwd
# Should show: /path/to/go-microservices-demo
```

### 6. Pre-flight Checks

Before proceeding, verify all tools are properly installed:

```bash
# Check Go
go version

# Check Docker
docker --version
docker ps

# Check if Docker daemon is running
docker info

# Test network connectivity (for downloading dependencies)
curl -I https://proxy.golang.org
```

**Troubleshooting Common Issues:**

1. **Go command not found:** Ensure Go is in your PATH environment variable
2. **Docker daemon not running:** Start Docker Desktop or run `sudo systemctl start docker` on Linux
3. **Permission denied for Docker:** Add your user to the docker group or use sudo
4. **Network issues:** Check firewall settings and proxy configuration
5. **Port conflicts:** Ensure ports 8080, 8081, 8082, and 8500 are available

### Next Steps

Once all prerequisites are installed and verified, you can proceed to Part 1 to set up the project structure.

---

### Part 1: Project Structure

First, let's create a clean directory structure for our project. This helps keep everything organized.

```bash
mkdir go-microservices-demo
cd go-microservices-demo

# Create a directory for our independent services
mkdir services
mkdir services/users-service
mkdir services/products-service

# Create a directory for the gateway
mkdir api-gateway
```

Your final project structure will look like this:

```bash
go-microservices-demo/
├── api-gateway/
│   └── main.go
└── services/
    ├── products-service/
    │   └── main.go
    └── users-service/
        └── main.go
```

### Part 2: Setting Up the Service Registry (Consul)

We'll use Consul as our service registry. The easiest way to run it for development is with Docker. The `-dev` flag starts Consul in a single-node development mode, which is perfect for our needs.

1.  **Run the Consul Docker Container:**
    Open your terminal and execute this command.

    ```bash
    docker run -d -p 8500:8500 --name=consul hashicorp/consul agent -dev -ui -client=0.0.0.0
    ```

    - `-p 8500:8500`: Maps the port for Consul's HTTP API and Web UI.
    - `-dev`: Runs Consul in a simplified development mode.
    - `-ui`: Enables the built-in web user interface.
    - `-client=0.0.0.0`: Binds the client interface to all network interfaces, allowing our Go apps to connect to it from our host machine.

2.  **Verify Consul is Running:**
    Check that the container is up and running.

    ```bash
    docker ps
    ```

    You should see an entry for the `consul` container. You can now access the **Consul Web UI** by navigating to `http://localhost:8500` in your web browser. It will show one healthy node (Consul itself) and zero registered services for now.

---

## Part 3: Building the `users-service`

This is our first microservice. It will expose a simple HTTP endpoint and, most importantly, register itself with Consul when it starts.

1.  **Initialize the Module:**
    Navigate to the service's directory and create a Go module.

    ```bash
    cd services/users-service
    go mod init users-service
    go get github.com/go-chi/chi/v5
    go get github.com/hashicorp/consul/api
    ```

2.  **Create `main.go`:**
    This file contains the full logic for the service. We've added detailed comments to explain each part.

    ```go
    // services/users-service/main.go
    package main

    import (
    	"fmt"
    	"log"
    	"net/http"
    	"os"

    	"github.com/go-chi/chi/v5"
    	consulapi "github.com/hashicorp/consul/api"
    )

    const serviceName = "users-service"
    const servicePort = 8081

    func main() {
    	// 1. Register this service with Consul.
    	// This is a crucial step for service discovery.
    	if err := registerServiceWithConsul(); err != nil {
    		log.Fatalf("Failed to register service with Consul: %v", err)
    	}

    	// 2. Set up the HTTP router and define endpoints.
    	r := chi.NewRouter()
    	r.Get("/health", healthCheckHandler) // Consul uses this to check service health.
    	r.Get("/users/{id}", getUserHandler)

    	log.Printf("'%s' starting on port %d...", serviceName, servicePort)

    	// 3. Start the HTTP server.
    	if err := http.ListenAndServe(fmt.Sprintf(":%d", servicePort), r); err != nil {
    		log.Fatalf("Failed to start server for service '%s': %v", serviceName, err)
    	}
    }

    // getUserHandler provides a simple response for a user endpoint.
    func getUserHandler(w http.ResponseWriter, r *http.Request) {
    	userID := chi.URLParam(r, "id")
    	fmt.Fprintf(w, "Response from '%s': Details for user %s\n", serviceName, userID)
    }

    // healthCheckHandler is essential for Consul to monitor the service's status.
    func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    	w.WriteHeader(http.StatusOK)
    	fmt.Fprintln(w, "Service is healthy")
    }

    // registerServiceWithConsul handles the logic of registering with the Consul agent.
    func registerServiceWithConsul() error {
    	config := consulapi.DefaultConfig()
    	consul, err := consulapi.NewClient(config)
    	if err != nil {
    		return err
    	}

    	// Get the hostname of the machine.
    	hostname, err := os.Hostname()
    	if err != nil {
    		return err
    	}

    	// Define the service registration details.
    	registration := &consulapi.AgentServiceRegistration{
    		ID:      fmt.Sprintf("%s-%s", serviceName, hostname), // A unique ID for this instance
    		Name:    serviceName,                                // The logical name of the service
    		Port:    servicePort,
    		Address: hostname,                                 // The address where this service is running
    		Check: &consulapi.AgentServiceCheck{
    			// Consul will hit this endpoint to check the service's health.
    			HTTP:     fmt.Sprintf("http://%s:%d/health", hostname, servicePort),
    			Interval: "10s", // Check every 10 seconds.
    			Timeout:  "1s",
    		},
    	}

    	// Perform the registration.
    	if err := consul.Agent().ServiceRegister(registration); err != nil {
    		return err
    	}

    	log.Printf("Successfully registered '%s' with Consul", serviceName)
    	return nil
    }
    ```

---

## Part 4: Building the `products-service`

To demonstrate the power of this pattern, we'll create a second, nearly identical service. This reinforces that services are independent and follow a repeatable registration pattern.

1.  **Initialize the Module:**

    ```bash
    # Make sure you are in the root directory 'go-microservices-demo'
    cd services/products-service
    go mod init products-service
    go get github.com/go-chi/chi/v5
    go get github.com/hashicorp/consul/api
    ```

2.  **Create `main.go`:**
    Copy the code from `users-service/main.go` and make these small changes:

    - Change `serviceName` to `"products-service"`.
    - Change `servicePort` to `8082` (it must run on a different port).
    - Change the response message in `getProductHandler` (rename `getUserHandler`).

    Here's the final code for `products-service/main.go`:

    ```go
    // services/products-service/main.go
    package main

    // ... (same imports as users-service)
    import (
        "fmt"
        "log"
        "net/http"
        "os"

        "github.com/go-chi/chi/v5"
        consulapi "github.com/hashicorp/consul/api"
    )

    const serviceName = "products-service"
    const servicePort = 8082 // <-- Different Port

    func main() {
        if err := registerServiceWithConsul(); err != nil {
            log.Fatalf("Failed to register service with Consul: %v", err)
        }

        r := chi.NewRouter()
        r.Get("/health", healthCheckHandler)
        r.Get("/products/{id}", getProductHandler) // <-- Different Endpoint

        log.Printf("'%s' starting on port %d...", serviceName, servicePort)
        if err := http.ListenAndServe(fmt.Sprintf(":%d", servicePort), r); err != nil {
            log.Fatalf("Failed to start server for service '%s': %v", serviceName, err)
        }
    }

    // getProductHandler provides a simple response for a product endpoint.
    func getProductHandler(w http.ResponseWriter, r *http.Request) {
        productID := chi.URLParam(r, "id")
        // <-- Different Response Message
        fmt.Fprintf(w, "Response from '%s': Details for product %s\n", serviceName, productID)
    }

    // healthCheckHandler and registerServiceWithConsul functions are IDENTICAL
    // to the ones in users-service, just with the new consts.
    func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    	w.WriteHeader(http.StatusOK)
    	fmt.Fprintln(w, "Service is healthy")
    }

    func registerServiceWithConsul() error {
        // ... (This function's code is exactly the same)
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
    		Check: &consulapi.AgentServiceCheck{
    			HTTP:     fmt.Sprintf("http://%s:%d/health", hostname, servicePort),
    			Interval: "10s",
    			Timeout:  "1s",
    		},
    	}

    	if err := consul.Agent().ServiceRegister(registration); err != nil {
    		return err
    	}

    	log.Printf("Successfully registered '%s' with Consul", serviceName)
    	return nil
    }
    ```

---

## Part 5: Building the API Gateway

The gateway is the public entry point. Its job is to receive a request, determine the target service from the URL, look up that service's address in Consul, and forward the request.

1.  **Initialize the Module:**

    ```bash
    # Make sure you are in the root directory 'go-microservices-demo'
    cd api-gateway
    go mod init api-gateway
    go get github.com/hashicorp/consul/api
    ```

2.  **Create `main.go`:**
    The core logic here is a dynamic reverse proxy.

    ```go
    // api-gateway/main.go
    package main

    import (
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httputil"
    	"net/url"
    	"strings"

    	consulapi "github.com/hashicorp/consul/api"
    )

    const gatewayPort = 8080

    func main() {
    	// The handler function is responsible for all routing logic.
    	http.HandleFunc("/", routeRequest)

    	log.Printf("API Gateway starting on port %d...", gatewayPort)
    	if err := http.ListenAndServe(fmt.Sprintf(":%d", gatewayPort), nil); err != nil {
    		log.Fatalf("Failed to start API Gateway: %v", err)
    	}
    }

    // routeRequest determines the target service from the request path.
    func routeRequest(w http.ResponseWriter, r *http.Request) {
    	log.Printf("Gateway received request for: %s", r.URL.Path)

    	// A simple routing rule: /api/users/* goes to "users-service".
    	// And /api/products/* goes to "products-service".
    	pathSegments := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
    	if len(pathSegments) < 3 || pathSegments[0] != "api" {
    		http.Error(w, "Invalid request path", http.StatusBadRequest)
    		return
    	}
    	serviceName := pathSegments[1] + "-service" // e.g., "users" -> "users-service"

    	// Step 1: Discover the service's location using Consul.
    	serviceURL, err := discoverService(serviceName)
    	if err != nil {
    		log.Printf("Error discovering service '%s': %v", serviceName, err)
    		http.Error(w, err.Error(), http.StatusServiceUnavailable)
    		return
    	}
    	log.Printf("Discovered '%s' at %s", serviceName, serviceURL)

    	// Step 2: Create a reverse proxy to forward the request.
    	proxy := httputil.NewSingleHostReverseProxy(serviceURL)

    	// Step 3: Rewrite the request URL to be sent to the downstream service.
    	// We strip `/api/<service-name>` from the original path.
    	// e.g., /api/users/123 becomes /users/123
    	r.URL.Path = "/" + strings.Join(pathSegments[1:], "/")
    	log.Printf("Forwarding request to: %s%s", serviceURL, r.URL.Path)

    	// Step 4: Forward the request.
    	proxy.ServeHTTP(w, r)
    }

    // discoverService queries Consul to find a healthy instance of a service.
    func discoverService(name string) (*url.URL, error) {
    	config := consulapi.DefaultConfig()
    	consul, err := consulapi.NewClient(config)
    	if err != nil {
    		return nil, err
    	}

    	// Query Consul's health endpoint for healthy service instances.
    	services, _, err := consul.Health().Service(name, "", true, nil)
    	if err != nil {
    		return nil, fmt.Errorf("could not query Consul for service '%s': %w", name, err)
    	}

    	if len(services) == 0 {
    		return nil, fmt.Errorf("no healthy instances of service '%s' found in Consul", name)
    	}

    	// For simplicity, we use the first available healthy instance.
    	// In a production system, you might use a load-balancing strategy here.
    	service := services[0].Service
    	serviceAddress := fmt.Sprintf("http://%s:%d", service.Address, service.Port)

    	return url.Parse(serviceAddress)
    }
    ```

---

## Part 6: Running and Testing the Full System

You will need **four separate terminal windows** for this part: one for Consul, one for each service, and one for the gateway.

1.  **Terminal 1: Consul**
    Make sure your Docker container is running from Part 2. If not, start it.

2.  **Terminal 2: `users-service`**
    Navigate to the `users-service` directory and run it.

    ```bash
    cd go-microservices-demo/services/users-service
    go run .
    # Expected Output:
    # Successfully registered 'users-service' with Consul
    # 'users-service' starting on port 8081...
    ```

    **Verification:** Refresh the Consul UI at `http://localhost:8500`. You should now see **`users-service`** listed with a green, passing health check. ✅

3.  **Terminal 3: `products-service`**
    Navigate to the `products-service` directory and run it.

    ```bash
    cd go-microservices-demo/services/products-service
    go run .
    # Expected Output:
    # Successfully registered 'products-service' with Consul
    # 'products-service' starting on port 8082...
    ```

    **Verification:** Refresh the Consul UI again. You should now see both services listed and healthy.

4.  **Terminal 4: `api-gateway`**
    Finally, start the gateway.

    ```bash
    cd go-microservices-demo/api-gateway
    go run .
    # Expected Output:
    # API Gateway starting on port 8080...
    ```

5.  **Test the End-to-End Flow:**
    Now, make requests **only to the gateway's port (8080)**.

    ```bash
    # Test the users service
    curl http://localhost:8080/api/users/123

    # Expected Response:
    # Response from 'users-service': Details for user 123

    # Actual Response 
    # no healthy instances of service 'users-service' found in Consul

    # Test the products service
    curl http://localhost:8080/api/products/abc

    # Expected Response:
    # Response from 'products-service': Details for product abc

    # Actual Response:
    # Response from 'products-service': Details for product abc
    ```

   The current issue right now, is that the services have registered to a Consul Instance in a docker environment while the web services are running on the host machine. This means that the gateway cannot reach the services using the addresses registered in Consul.

   ### Resolution

   #### Option 1

   1. Install Consul Locally 
   2. Navigate to the following URL https://developer.hashicorp.com/consul/install
   3. After successful installation run the following command to initialise your Consul Instance on host machine on a new terminal window. ***TERMINATE YOUR CURRENT CONSUL INSTANCE ON DOCKER.***
```
    consul agent -dev
```
4. Next, navigate to http://localhost:8500/ui to verify that your consul instance is up and running.

#### Option 2

1. Containerise all 3 services; api-gateway,user-service,product-service
2. Create a Dockerfile for each of your go services in their respective folders
3. Create a Docker Compose file at the root of your project to define and run all services

---

## Part 7: Demonstrating Resilience (Bonus)

Let's see the magic of service discovery in action.

1.  **Stop a Service:** Go to the terminal running `users-service` and stop it with `Ctrl + C`.

2.  **Check Consul:** Refresh the Consul UI. Within about 10 seconds, the health check for `users-service` will fail, and its status will turn to a critical red. ❌

3.  **Try to Access the Service:**

    ```bash
    curl http://localhost:8080/api/users/123
    # Expected Response:
    # no healthy instances of service 'users-service' found in Consul
    ```

    The gateway correctly reports that the service is unavailable because Consul told it so. The `products-service` will continue to work perfectly.

4.  **Restart the Service:** Go back to the `users-service` terminal and restart it with `go run .`.

5.  **Check Consul and Re-Test:** The service will re-register itself, and its health check will turn green again in the UI. Now, run the `curl` command again:

    ```bash
    curl http://localhost:8080/api/users/123
    # Expected Response:
    # Response from 'users-service': Details for user 123
    ```

    It works again\! You didn't have to touch the gateway or any other part of the system. This demonstrates the power of a decoupled, resilient architecture.
