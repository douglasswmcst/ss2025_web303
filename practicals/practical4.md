# **Module Practical: WEB303 Microservices & Serverless Applications**

## **Practical 4: Kubernetes Microservices with Kong Gateway & Resilience Patterns**

## **Part 1**

### **Objective**

This practical builds upon your previous microservices experience and introduces you to production-grade deployment with **Kubernetes**, advanced API gateway management with **Kong**. You'll implement an cafe order management system that demonstrates real-world microservices challenges and solutions.

In the second part of the practical, you will then focus on implmenting the 3 resilience patterns, **timeout**, **retry**, and **circuit breaker** to enhance the reliability of your distributed system.

I have removed the need for gRPC & a stateful database to focus on the learning objectives of this practical to focus on kubernetes first and then resilience patterns.

### **Learning Outcomes**

Part 1

- \*\*Learning You should have now completed a multi-service application using Go, React, Kong, Consul, and Kubernetes!

## **⚠️ IMPORTANT: Exercise to Complete**

### **Order Submission Issue**

**Notice:** The order submission functionality in this application is intentionally broken. As part of your practical exercise, you need to identify and fix the issues preventing successful order placement.

### **Your Task:**

1. **Deploy and Test the Application:**

   - Follow all the steps above to deploy the complete microservices application
   - Access the React frontend through the Kong gateway URL
   - Try to place an order by adding items to cart and clicking "Place Order"
   - **You should observe that the order submission fails**

2. **Debug and Identify Issues:**

   Use the following debugging techniques to identify the problems:

   ```bash
   # Check if all pods are running
   kubectl get pods -n student-cafe

   # Check service endpoints
   kubectl get endpoints -n student-cafe

   # Check ingress configuration
   kubectl describe ingress cafe-ingress -n student-cafe

   # Monitor pod logs during order submission
   kubectl logs -f deployment/order-deployment -n student-cafe
   kubectl logs -f deployment/food-catalog-deployment -n student-cafe
   kubectl logs -f deployment/kong-kong -n student-cafe

   # Test API endpoints directly
   curl -X POST http://$(minikube ip):PORT/api/orders/orders \
     -H "Content-Type: application/json" \
     -d '{"item_ids": ["1", "2"]}'
   ```

## Troubleshooting Common Issues

### Issue 1: Go Version Compatibility

**Problem:** Docker build fails with error: `go: go.mod requires go >= 1.25.0`

**Solution:** Update the Go version in both Dockerfiles:

```dockerfile
# Change this:
FROM golang:1.21-alpine AS builder
# To this:
FROM golang:1.23-alpine AS builder
```

Also update your `go.mod` files to use a compatible version:

```go
module food-catalog-service

go 1.23  // Change from go 1.25.0
```

Run `go mod tidy` in each service directory after making these changes.

### Issue 2: Minikube Docker Environment

**Problem:** Docker images not found when deploying to Kubernetes

**Solution:** Make sure to configure your Docker client to use Minikube's Docker daemon:

```bash
eval $(minikube -p minikube docker-env)
```

Run this command before building your Docker images.

### Issue 3: Consul Connection Issues

**Problem:** If Services fail to register with Consul or can't find other services

**Solution:** The Go services needs to be updated to handle Consul gracefully:

- Consul client address is set to `consul-server:8500` for Kubernetes
- Service registration is non-blocking (runs in goroutine)
- Errors are logged but don't crash the service

### Issue 4: Kong Ingress Not Working

**Problem:** If can't access the application through Kong proxy

**Solution:**

1. Check if Kong is running: `kubectl get pods -n student-cafe | grep kong`
2. Verify the ingress was created: `kubectl get ingress -n student-cafe`
3. Get the Kong service URL: `minikube service -n student-cafe kong-kong-proxy --url`
4. Check Kong logs: `kubectl logs deployment/kong-kong -n student-cafe`

### Issue 5: Pods Stuck in Pending or ContainerCreating

**Problem:** Kubernetes pods don't start properly

**Solution:**

1. Check pod status: `kubectl describe pod <pod-name> -n student-cafe`
2. Ensure images are built with correct tags
3. Verify minikube has enough resources: `minikube start --cpus 2 --memory 4096`

### Useful Commands for Debugging

```bash
# Check all resources in namespace
kubectl get all -n student-cafe

# Check pod logs
kubectl logs <pod-name> -n student-cafe

# Check service endpoints
kubectl get endpoints -n student-cafe

# Check ingress status
kubectl describe ingress cafe-ingress -n student-cafe

# Get minikube IP
minikube ip

# Test API endpoints directly
curl http://$(minikube ip):32147/api/catalog/items
```

## **Submission Instructions**

### **What to Submit**

1. **Complete Project Structure**: Submit your entire practical 4 `student-cafe/` project directory containing:

   - `food-catalog-service/` folder with `main.go`, `Dockerfile`, `go.mod`, and `go.sum`
   - `order-service/` folder with `main.go`, `Dockerfile`, `go.mod`, and `go.sum`
   - `cafe-ui/` folder with complete React application and `Dockerfile`
   - `app-deployment.yaml` - Kubernetes deployment manifests
   - `kong-ingress.yaml` - Kong ingress configuration

2. **Documentation**: Create a `README.md` file in your project root that includes:

   - **Screenshots**: Include screenshots showing:
     - Your React frontend displaying the food menu & successful order placement
     - Kubernetes pods running (`kubectl get pods -n student-cafe`)
     - Kong services status
     - Output of `kubectl get services -n student-cafe` command

Part 2

- **Learning Outcome 4:** Implement resilience patterns (timeout, retry, circuit breaker) to enhance distributed system reliability

---

---

## Go Microservices Walkthrough: Student Cafe App

This guide will walk you through building and deploying a simple food ordering application. Students can view a list of food items and place an order.

### 1\. Architecture Overview

We will build a system with the following components:

- **Frontend (React.js):** A single-page application (SPA) that provides the user interface for students.
- **`food-catalog-service` (Go & Chi):** A microservice responsible for providing a list of available food items.
- **`order-service` (Go & Chi):** A microservice for creating and managing food orders.
- **Service Discovery (Consul):** Allows our microservices to find and communicate with each other without hardcoding IP addresses. Each service will register itself with Consul.
- **API Gateway (Kong):** A single entry point for all external traffic. The React frontend will communicate only with Kong, which will then intelligently route requests to the appropriate backend microservice.
- **Containerization & Orchestration (Docker & Kubernetes):** We will containerize each component (frontend, services) using Docker and deploy them on a local Kubernetes cluster (Minikube).

Here is the user flow:

1.  The student's browser loads the React application.
2.  The React app makes API calls to the Kong API Gateway.
3.  Kong routes traffic to the correct microservice (`/api/catalog` -\> `food-catalog-service`, `/api/orders` -\> `order-service`).
4.  The `order-service` needs data from the `food-catalog-service`, so it queries Consul to discover its location and then makes a direct internal request.
5.  All components run as pods within a Kubernetes cluster.

### 2\. Prerequisites

Ensure you have the following tools installed on your machine:

- **Go:** (version 1.18+)
- **Node.js & npm:** (for the React frontend)
- **Docker:** (for containerizing our apps)
- **Minikube:** (for a local Kubernetes cluster)
- **kubectl:** (the Kubernetes command-line tool)
- **Helm:** (the package manager for Kubernetes)

---

### \#\# 1. Go Language 🚀

#### **Windows**

- **Recommended:** Open PowerShell as Administrator and use Chocolatey:
  ```powershell
  choco install golang -y
  ```
- **Alternative:** Download the official MSI installer from the [Go website](https://go.dev/dl/) and follow the on-screen instructions.

#### **macOS**

- **Recommended:** Open your terminal and use Homebrew:
  ```bash
  brew install go
  ```

#### **Linux (Debian/Ubuntu)**

- Open your terminal and use the `apt` package manager:
  ```bash
  sudo apt update
  sudo apt install golang-go -y
  ```

**Verify Installation:**
Open a new terminal and run: `go version`

---

### \#\# 2. Node.js & npm 📦

#### **Windows**

- **Recommended:** Use Chocolatey in an Administrator PowerShell:
  ```powershell
  choco install nodejs-lts -y
  ```
- **Alternative:** Download the LTS (Long-Term Support) installer from the [Node.js website](https://nodejs.org/).

#### **macOS**

- **Recommended:** Use Homebrew in your terminal:
  ```bash
  brew install node
  ```

#### **Linux (Debian/Ubuntu)**

- Use `apt` in your terminal:
  ```bash
  sudo apt update
  sudo apt install nodejs npm -y
  ```

**Verify Installation:**
Run `node -v` and `npm -v` in a new terminal.

---

### \#\# 3. Docker 🐳

#### **Windows**

1.  First, ensure you have **WSL 2** (Windows Subsystem for Linux) installed. If not, open PowerShell as Administrator and run:
    ```powershell
    wsl --install
    ```
2.  Download and install **Docker Desktop for Windows** from the [Docker website](https://www.docker.com/products/docker-desktop/). It will guide you through the setup process.

#### **macOS**

1.  Download and install **Docker Desktop for Mac** from the [Docker website](https://www.docker.com/products/docker-desktop/).
2.  Drag the Docker icon to your Applications folder and run it.

#### **Linux (Debian/Ubuntu)**

1.  Uninstall old versions:
    ```bash
    sudo apt-get remove docker docker-engine docker.io containerd runc
    ```
2.  Set up Docker's official repository:
    ```bash
    sudo apt-get update
    sudo apt-get install ca-certificates curl gnupg
    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
      sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    ```
3.  Install Docker Engine:
    ```bash
    sudo apt-get update
    sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y
    ```
4.  (Optional) Add your user to the `docker` group to run commands without `sudo`:
    ```bash
    sudo usermod -aG docker $USER
    # You'll need to log out and log back in for this to take effect.
    ```

**Verify Installation:**
Run `docker --version` in your terminal.

---

### \#\# 4. kubectl (Kubernetes CLI) ⛵

#### **Windows**

- **Recommended:** Use Chocolatey in an Administrator PowerShell:
  ```powershell
  choco install kubernetes-cli
  ```

#### **macOS**

- **Recommended:** Use Homebrew in your terminal:
  ```bash
  brew install kubectl
  ```

#### **Linux (Debian/Ubuntu)**

- Download the binary using `curl`:
  ```bash
  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
  ```

**Verify Installation:**
Run `kubectl version --client` in your terminal.

---

### \#\# 5. Minikube (Local Kubernetes) ⚙️

#### **Windows**

- **Recommended:** Use Chocolatey in an Administrator PowerShell:
  ```powershell
  choco install minikube
  ```

#### **macOS**

- **Recommended:** Use Homebrew in your terminal:
  ```bash
  brew install minikube
  ```

#### **Linux (Debian/Ubuntu)**

- Download and install the binary:
  ```bash
  curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
  sudo install minikube /usr/local/bin/
  ```

**Verify Installation:**
Run `minikube version` in your terminal.

---

### \#\# 6. Helm (Kubernetes Package Manager) 🧭

#### **Windows**

- **Recommended:** Use Chocolatey in an Administrator PowerShell:
  ```powershell
  choco install kubernetes-helm
  ```

#### **macOS**

- **Recommended:** Use Homebrew in your terminal:
  ```bash
  brew install helm
  ```

#### **Linux (Debian/Ubuntu)**

- Use the official installer script:
  ```bash
  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
  ```

**Verify Installation:**
Run `helm version` in your terminal.

Start your local Kubernetes cluster:

```bash
minikube start --cpus 4 --memory 4096
# Point your local docker client to minikube's docker daemon
eval $(minikube -p minikube docker-env)
```

By using `eval $(minikube docker-env)`, you can build Docker images locally that are immediately available to your Minikube cluster without needing to push them to a remote registry.

### 3\. Project Structure

Create a root directory for your project. Inside, we'll have folders for each service and the frontend.

```
student-cafe/
├── food-catalog-service/
│   ├── main.go
│   └── Dockerfile
├── order-service/
│   ├── main.go
│   └── Dockerfile
└── cafe-ui/
    ├── (React app files)
    └── Dockerfile
```

### 4\. Microservice 1: `food-catalog-service`

This service simply returns a static list of food items.

**`food-catalog-service/main.go`**

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	consulapi "github.com/hashicorp/consul/api"
)

type FoodItem struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var foodItems = []FoodItem{
	{ID: "1", Name: "Coffee", Price: 2.50},
	{ID: "2", Name: "Sandwich", Price: 5.00},
	{ID: "3", Name: "Muffin", Price: 3.25},
}

// Service registration with Consul
func registerServiceWithConsul() {
	config := consulapi.DefaultConfig()
	// In Kubernetes, Consul service is available at consul-server
	config.Address = "consul-server:8500"

	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Printf("Warning: Could not create Consul client: %v", err)
		return
	}

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = "food-catalog-service"
	registration.Name = "food-catalog-service"
	registration.Port = 8080
	// Get hostname to use as address
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Could not get hostname: %v", err)
	}
	registration.Address = hostname

	// Add a health check
	registration.Check = &consulapi.AgentServiceCheck{
		HTTP:     fmt.Sprintf("http://%s:%d/health", hostname, 8080),
		Interval: "10s",
		Timeout:  "1s",
	}

	if err := consul.Agent().ServiceRegister(registration); err != nil {
		log.Printf("Warning: Failed to register service with Consul: %v", err)
		return
	}
	log.Println("Successfully registered service with Consul")
}

func main() {
	// Try to register with Consul, but don't fail if it's not available
	go registerServiceWithConsul()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(foodItems)
	})

	log.Println("Food Catalog Service starting on port 8080...")
	http.ListenAndServe(":8080", r)
}
```

**`food-catalog-service/Dockerfile`**

```dockerfile
# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
# Download all dependencies.
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o /food-catalog-service .

# Stage 2: Create a minimal final image
FROM alpine:latest

# We need consul binary for the health check
RUN apk --no-cache add curl

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /food-catalog-service /food-catalog-service

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["/food-catalog-service"]
```

Before building, initialize Go modules:

```bash
cd food-catalog-service
go mod init food-catalog-service
go get github.com/go-chi/chi/v5
go get github.com/hashicorp/consul/api
go mod tidy
cd ..
```

### 5\. Microservice 2: `order-service`

This service handles order creation. For simplicity, it stores orders in memory.

**`order-service/main.go`**

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-uuid"
)

type Order struct {
	ID      string   `json:"id"`
	ItemIDs []string `json:"item_ids"`
	Status  string   `json:"status"`
}

var orders = make(map[string]Order)

// Service registration with Consul
func registerServiceWithConsul() {
    config := consulapi.DefaultConfig()
	// In Kubernetes, Consul service is available at consul-server
	config.Address = "consul-server:8500"

	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Printf("Warning: Could not create Consul client: %v", err)
		return
	}

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = "order-service"
	registration.Name = "order-service"
	registration.Port = 8081
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Could not get hostname: %v", err)
	}
	registration.Address = hostname

	registration.Check = &consulapi.AgentServiceCheck{
		HTTP:     fmt.Sprintf("http://%s:%d/health", hostname, 8081),
		Interval: "10s",
		Timeout:  "1s",
	}

	if err := consul.Agent().ServiceRegister(registration); err != nil {
		log.Printf("Warning: Failed to register service with Consul: %v", err)
		return
	}
	log.Println("Successfully registered service with Consul")
}

// Discover other services using Consul
func findService(serviceName string) (string, error) {
    config := consulapi.DefaultConfig()
	// In Kubernetes, Consul service is available at consul-server
	config.Address = "consul-server:8500"

	consul, err := consulapi.NewClient(config)
	if err != nil {
		return "", err
	}

	services, _, err := consul.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", err
	}

	if len(services) == 0 {
		return "", fmt.Errorf("could not find any healthy instance of %s", serviceName)
	}

	// In a real app, you'd implement load balancing here.
	// For now, we just take the first healthy instance.
	addr := services[0].Service.Address
	port := services[0].Service.Port
	return fmt.Sprintf("http://%s:%d", addr, port), nil
}


func main() {
	// Try to register with Consul, but don't fail if it's not available
	go registerServiceWithConsul()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
		var newOrder Order
		if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

        // Example of inter-service communication
        // Here you would call the food-catalog-service to validate ItemIDs
        catalogAddr, err := findService("food-catalog-service")
        if err != nil {
            http.Error(w, "Food catalog service not available", http.StatusInternalServerError)
            log.Printf("Error finding catalog service: %v", err)
            return
        }
        log.Printf("Found food-catalog-service at: %s. Would validate items here.", catalogAddr)


		orderID, _ := uuid.GenerateUUID()
		newOrder.ID = orderID
		newOrder.Status = "received"
		orders[orderID] = newOrder

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newOrder)
	})

	log.Println("Order Service starting on port 8081...")
	http.ListenAndServe(":8081", r)
}

```

**`order-service/Dockerfile`**
(This Dockerfile is almost identical to the one for the `food-catalog-service`, just change the final binary name).

```dockerfile
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
```

Initialize Go modules for this service too:

```bash
cd order-service
go mod init order-service
go get github.com/go-chi/chi/v5
go get github.com/hashicorp/consul/api
go get github.com/hashicorp/go-uuid
go mod tidy
cd ..
```

### 6\. Frontend: `cafe-ui`

We'll use `create-react-app` for a quick setup.

```bash
npx create-react-app cafe-ui
cd cafe-ui
```

Replace `src/App.js` with the following:
**`cafe-ui/src/App.js`**

```javascript
import React, { useState, useEffect } from "react";
import "./App.css";

function App() {
  const [items, setItems] = useState([]);
  const [cart, setCart] = useState([]);
  const [message, setMessage] = useState("");

  useEffect(() => {
    // We fetch from the API Gateway's route, not the service directly
    fetch("/api/catalog/items")
      .then((res) => res.json())
      .then((data) => setItems(data))
      .catch((err) => console.error("Error fetching items:", err));
  }, []);

  const addToCart = (item) => {
    setCart((prevCart) => [...prevCart, item]);
  };

  const placeOrder = () => {
    if (cart.length === 0) {
      setMessage("Your cart is empty!");
      return;
    }

    const order = {
      item_ids: cart.map((item) => item.id),
    };

    fetch("/api/orders/orders", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(order),
    })
      .then((res) => res.json())
      .then((data) => {
        setMessage(`Order ${data.id} placed successfully!`);
        setCart([]); // Clear cart
      })
      .catch((err) => {
        setMessage("Failed to place order.");
        console.error("Error placing order:", err);
      });
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>Student Cafe</h1>
      </header>
      <main className="container">
        <div className="menu">
          <h2>Menu</h2>
          <ul>
            {items.map((item) => (
              <li key={item.id}>
                <span>
                  {item.name} - ${item.price.toFixed(2)}
                </span>
                <button onClick={() => addToCart(item)}>Add to Cart</button>
              </li>
            ))}
          </ul>
        </div>
        <div className="cart">
          <h2>Your Cart</h2>
          <ul>
            {cart.map((item, index) => (
              <li key={index}>{item.name}</li>
            ))}
          </ul>
          <button onClick={placeOrder} className="order-btn">
            Place Order
          </button>
          {message && <p className="message">{message}</p>}
        </div>
      </main>
    </div>
  );
}

export default App;
```

Add some basic styling in `src/App.css`.

**`cafe-ui/Dockerfile`**

```dockerfile
# Stage 1: Build the React app
FROM node:18-alpine as build
WORKDIR /app
COPY package.json ./
COPY package-lock.json ./
RUN npm install
COPY . ./
RUN npm run build

# Stage 2: Serve with Nginx
FROM nginx:stable-alpine
COPY --from=build /app/build /usr/share/nginx/html
# Nginx's default port is 80
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### 7\. Infrastructure on Kubernetes

Now we deploy the infrastructure services (Consul, Kong) using Helm.

**Create a namespace:**

```bash
kubectl create namespace student-cafe
```

**Deploy Consul:**

```bash
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install consul hashicorp/consul --set global.name=consul --namespace student-cafe --set server.replicas=1 --set server.bootstrapExpect=1
```

**Deploy Kong:**

```bash
helm repo add kong https://charts.konghq.com
helm repo update
helm install kong kong/kong --namespace student-cafe
```

### 8\. Build & Deploy Our Application

First, build the Docker images for our three applications. Remember to run `eval $(minikube -p minikube docker-env)` first.

```bash
# In the root project directory
docker build -t food-catalog-service:v1 ./food-catalog-service
docker build -t order-service:v1 ./order-service
docker build -t cafe-ui:v1 ./cafe-ui
```

Now, create the Kubernetes deployment and service manifests. Create a file named `app-deployment.yaml`:

**`app-deployment.yaml`**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: food-catalog-deployment
  namespace: student-cafe
spec:
  replicas: 1
  selector:
    matchLabels:
      app: food-catalog-service
  template:
    metadata:
      labels:
        app: food-catalog-service
    spec:
      containers:
        - name: food-catalog-service
          image: food-catalog-service:v1
          imagePullPolicy: IfNotPresent # Important for local minikube images
          ports:
            - containerPort: 8080
          env:
            - name: CONSUL_HTTP_ADDR
              value: "consul-server:8500"
---
apiVersion: v1
kind: Service
metadata:
  name: food-catalog-service
  namespace: student-cafe
spec:
  selector:
    app: food-catalog-service
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-deployment
  namespace: student-cafe
spec:
  replicas: 1
  selector:
    matchLabels:
      app: order-service
  template:
    metadata:
      labels:
        app: order-service
    spec:
      containers:
        - name: order-service
          image: order-service:v1
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8081
          env:
            - name: CONSUL_HTTP_ADDR
              value: "consul-server:8500"
---
apiVersion: v1
kind: Service
metadata:
  name: order-service
  namespace: student-cafe
spec:
  selector:
    app: order-service
  ports:
    - protocol: TCP
      port: 8081
      targetPort: 8081
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cafe-ui-deployment
  namespace: student-cafe
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cafe-ui
  template:
    metadata:
      labels:
        app: cafe-ui
    spec:
      containers:
        - name: cafe-ui
          image: cafe-ui:v1
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: cafe-ui-service
  namespace: student-cafe
spec:
  selector:
    app: cafe-ui
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
```

Apply the manifests:

```bash
kubectl apply -f app-deployment.yaml
```

### 9\. Configure Kong API Gateway

Finally, we need to tell Kong how to route traffic. We do this by creating an `Ingress` resource. Create a file named `kong-ingress.yaml`.

**`kong-ingress.yaml`**

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cafe-ingress
  namespace: student-cafe
  annotations:
    konghq.com/strip-path: "true"
spec:
  ingressClassName: kong
  rules:
    - http:
        paths:
          - path: /api/catalog
            pathType: Prefix
            backend:
              service:
                name: food-catalog-service
                port:
                  number: 8080
          - path: /api/orders
            pathType: Prefix
            backend:
              service:
                name: order-service
                port:
                  number: 8081
          - path: /
            pathType: Prefix
            backend:
              service:
                name: cafe-ui-service
                port:
                  number: 80
```

Apply the Ingress configuration:

```bash
kubectl apply -f kong-ingress.yaml
```

### 10\. Accessing the Application

1.  Get the external IP address for Kong from Minikube:

    ```bash
    minikube service -n student-cafe kong-kong-proxy --url
    ```

    This will output a URL like `http://192.168.49.2:31234`.

2.  Open this URL in your web browser. You should see the React frontend.

3.  You can view the food menu, add items to your cart, and place an order. Watch the logs of the pods to see the requests being processed:

    ```bash
    # Get pod names
    kubectl get pods -n student-cafe

    # Tail logs for a specific pod
    kubectl logs -f <pod-name-here> -n student-cafe
    ```

You should have now completed a multi-service application using Go, React, Kong, Consul, and Kubernetes\!
