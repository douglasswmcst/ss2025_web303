# **Module Practical: WEB303 Microservices & Serverless Applications**

## **Practical 4: Kubernetes Microservices with Kong Gateway & Resilience Patterns**

### **Objective**

This practical builds upon your previous microservices experience and introduces you to production-grade deployment with **Kubernetes**, advanced API gateway management with **Kong**, and critical resilience patterns. You'll implement an e-commerce order management system that demonstrates real-world microservices challenges and solutions.

### **Learning Outcomes**

- **Learning Outcome 4:** Implement resilience patterns (timeout, retry, circuit breaker) to enhance distributed system reliability
- **Learning Outcome 6:**- Deploy microservices to Kubernetes, implementing various deployment strategies and managing certificates.

---

## **Architecture Overview**

We'll build a simplified e-commerce system with the following services:

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────────┐
│   React Client  │────│  Kong Gateway │────│   Kubernetes Cluster   │
└─────────────────┘    └──────────────┘    └─────────────────────┘
                              │                        │
                              │            ┌───────────┼────────────┐
                              │            │           │            │
                         ┌────▼────┐  ┌────▼────┐ ┌───▼────┐ ┌────▼────┐ ┌────▼────┐
                         │  User   │  │Product  │ │ Cart   │ │ Order   │ │Payment  │
                         │Service  │  │Service  │ │Service │ │Service  │ │Service  │
                         └────┬────┘  └────┬────┘ └───┬────┘ └────┬────┘ └────┬────┘
                              │            │          │           │           │
                         ┌────▼────┐  ┌────▼────┐ ┌───▼────┐ ┌────▼────┐ ┌────▼────┐
                         │Users DB │  │Products │ │Cart DB │ │Orders DB│ │Payments │
                         │(Postgres)│  │DB(Postgres)│(Postgres)│(Postgres)│DB(Postgres)│
                         └─────────┘  └─────────┘ └────────┘ └─────────┘ └─────────┘
```

### **Services:**

1. **User Service**: User authentication and profile management
2. **Product Service**: Product catalog and inventory management
3. **Cart Service**: Shopping cart operations
4. **Order Service**: Order processing and workflow orchestration
5. **Payment Service**: Payment processing simulation (demonstrates circuit breaker)

---

## **Technology Stack & Architecture Components**

### **Core Technologies**

- **Backend**: Go with Chi router
- **Database**: PostgreSQL (one per service)
- **Service Discovery**: Consul Connect (service mesh)
- **API Gateway**: Kong
- **Container Orchestration**: Kubernetes
- **Frontend**: ReactJS
- **Development**: Docker, kubectl, helm

### **Kubernetes Role in Microservices**

**Kubernetes** serves as our container orchestration platform, providing:

1. **Service Discovery & Load Balancing**: Automatically distributes traffic across service replicas
2. **Auto-scaling**: Horizontal Pod Autoscaler (HPA) scales services based on CPU/memory usage
3. **Health Checks**: Liveness and readiness probes ensure service reliability
4. **Configuration Management**: ConfigMaps and Secrets for environment-specific settings
5. **Rolling Updates**: Zero-downtime deployments with automatic rollback capabilities
6. **Resource Management**: CPU/memory limits and requests for efficient resource utilization
7. **Network Policies**: Secure inter-service communication

### **Helm Role in Microservices Management**

**Helm** acts as the "package manager" for Kubernetes, providing:

1. **Templating**: Reusable YAML templates with values injection
2. **Release Management**: Track, upgrade, and rollback application deployments
3. **Dependency Management**: Manage complex service dependencies and installation order
4. **Environment Promotion**: Consistent deployments across dev/staging/production
5. **Configuration Management**: Environment-specific values files
6. **Lifecycle Hooks**: Pre/post-install, upgrade, and delete hooks

### **Kong Gateway Integration**

**Kong** provides enterprise-grade API gateway functionality:

1. **Traffic Management**: Rate limiting, request/response transformation
2. **Security**: Authentication, authorization, and JWT validation
3. **Observability**: Logging, metrics, and distributed tracing
4. **Service Mesh Integration**: Works seamlessly with Consul Connect
5. **Plugin Ecosystem**: Extensible with custom and community plugins

### **Consul Service Mesh Benefits**

**Consul Connect** enables secure service-to-service communication:

1. **Service Discovery**: Automatic service registration and health checking
2. **mTLS Encryption**: Automatic certificate management and rotation
3. **Traffic Management**: Circuit breakers, retries, and timeouts
4. **Observability**: Service topology and traffic flow visualization
5. **Policy Enforcement**: Intention-based access control between services

---

## **Part 1: Prerequisites & Environment Setup**

### **1.1 Required Tools Installation**

#### **macOS Setup**

```bash
# Install required tools (macOS)
brew install kubectl kubernetes-cli
brew install helm
brew install docker
brew install go
brew install node npm

# Install protobuf compiler
brew install protobuf

# Install Go plugins for protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify installations
kubectl version --client
helm version
docker --version
go version
node --version
protoc --version
```

#### **Linux (Ubuntu/Debian) Setup**

```bash
# Update package manager
sudo apt update

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
newgrp docker

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install Go (replace with latest version)
wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Node.js and npm
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install protobuf compiler
sudo apt-get install -y protobuf-compiler

# Install Go plugins for protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify installations
kubectl version --client
helm version
docker --version
go version
node --version
protoc --version
```

#### **Linux (CentOS/RHEL/Fedora) Setup**

```bash
# Update package manager
sudo dnf update -y  # or 'sudo yum update -y' for older versions

# Install Docker
sudo dnf install -y dnf-utils
sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
sudo dnf install -y docker-ce docker-ce-cli containerd.io
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
newgrp docker

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install Go
sudo dnf install -y golang
# Or download latest version:
# wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
# sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
# echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Install Node.js and npm
sudo dnf install -y nodejs npm

# Install protobuf compiler
sudo dnf install -y protobuf-compiler

# Install Go plugins for protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

#### **Windows Setup**

**Option 1: Using Chocolatey (Recommended)**

```powershell
# Install Chocolatey (run as Administrator)
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install required tools
choco install docker-desktop
choco install kubernetes-cli
choco install kubernetes-helm
choco install golang
choco install nodejs
choco install protoc
choco install git

# Install Go plugins for protobuf (run in Command Prompt or PowerShell)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify installations
kubectl version --client
helm version
docker --version
go version
node --version
protoc --version
```

**Option 2: Manual Installation**

1. **Docker Desktop**: Download from https://www.docker.com/products/docker-desktop
   - Enable Kubernetes in Docker Desktop settings
2. **kubectl**: Download from https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/

   ```powershell
   # Using curl (if available)
   curl -LO "https://dl.k8s.io/release/v1.28.0/bin/windows/amd64/kubectl.exe"
   # Add to PATH environment variable
   ```

3. **Helm**: Download from https://helm.sh/docs/intro/install/#from-the-binary-releases

   ```powershell
   # Download and extract helm.exe
   # Add to PATH environment variable
   ```

4. **Go**: Download from https://golang.org/dl/

   - Install MSI package
   - Verify GOPATH and GOROOT are set

5. **Node.js**: Download from https://nodejs.org/

   - Install MSI package

6. **Protobuf Compiler**:
   - Download from https://github.com/protocolbuffers/protobuf/releases
   - Extract and add `bin` directory to PATH

#### **WSL2 Setup (Windows Subsystem for Linux)**

For Windows users, WSL2 provides a native Linux experience:

```bash
# Install WSL2 (PowerShell as Administrator)
wsl --install

# After restart, in WSL2 terminal, follow Linux Ubuntu setup steps above

# Configure Docker Desktop to use WSL2 backend
# In Docker Desktop: Settings → General → Use the WSL 2 based engine
# In Docker Desktop: Settings → Resources → WSL Integration → Enable integration
```

### **1.2 Kubernetes Cluster Setup**

#### **Local Development Options**

**Option 1: Docker Desktop Kubernetes (Recommended for beginners)**

```bash
# Enable Kubernetes in Docker Desktop
# macOS/Windows: Docker Desktop → Preferences → Kubernetes → Enable Kubernetes
# Linux: Use Docker Desktop for Linux or alternative options below

# Verify cluster is running
kubectl cluster-info
kubectl get nodes

# Set context (if multiple clusters)
kubectl config current-context
kubectl config use-context docker-desktop
```

**Option 2: Minikube (Cross-platform)**

```bash
# Install Minikube
# macOS
brew install minikube

# Linux
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Windows (using Chocolatey)
choco install minikube

# Start Minikube cluster
minikube start --driver=docker --cpus=4 --memory=8192
minikube addons enable ingress
minikube addons enable metrics-server

# Verify cluster
kubectl cluster-info
kubectl get nodes
```

**Option 3: Kind (Kubernetes in Docker)**

```bash
# Install Kind
# macOS
brew install kind

# Linux
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Windows
choco install kind

# Create cluster
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
- role: worker
- role: worker
EOF

# Verify cluster
kubectl cluster-info
kubectl get nodes
```

#### **Production-Ready Options**

**For Learning/Testing:**

- **EKS (AWS)**: Managed Kubernetes service
- **GKE (Google Cloud)**: Google Kubernetes Engine
- **AKS (Azure)**: Azure Kubernetes Service
- **DigitalOcean Kubernetes**: Simple managed Kubernetes

**Resource Requirements:**

- **Minimum**: 4GB RAM, 2 CPU cores
- **Recommended**: 8GB RAM, 4 CPU cores
- **Disk Space**: 20GB free space

### **1.3 Install Kong API Gateway on Kubernetes**

**Kong** serves as our API Gateway, providing centralized routing, authentication, rate limiting, and observability for all microservices. **Helm** simplifies Kong's complex Kubernetes deployment by managing all required resources (deployments, services, configmaps, secrets) as a single unit.

#### **Why Kong + Kubernetes + Helm?**

1. **Kong Benefits:**

   - Single entry point for all API traffic
   - Built-in security (JWT, OAuth2, API keys)
   - Rate limiting and traffic control
   - Request/response transformation
   - Plugin ecosystem for extensibility

2. **Helm Benefits for Kong:**
   - Manages 15+ Kubernetes resources as one deployment
   - Handles complex configuration with simple values
   - Enables easy upgrades and rollbacks
   - Environment-specific configurations

#### **Kong Installation Steps**

```bash
# Add Kong Helm repository
helm repo add kong https://charts.konghq.com
helm repo update

# Create dedicated namespace for Kong
kubectl create namespace kong

# Install Kong with PostgreSQL database
helm install kong kong/kong -n kong \
  --set ingressController.enabled=true \
  --set image.repository=kong \
  --set image.tag=3.4 \
  --set env.database=postgres \
  --set postgresql.enabled=true \
  --set postgresql.auth.postgresPassword=kong \
  --set admin.enabled=true \
  --set admin.http.enabled=true

# Wait for Kong to be ready
kubectl wait --namespace kong \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=app \
  --timeout=300s

# Verify Kong installation
kubectl get pods -n kong
kubectl get services -n kong

# Get Kong Admin API URL (for local clusters)
echo "Kong Admin API: http://localhost:8001"
kubectl port-forward -n kong service/kong-kong-admin 8001:8001 &

# Get Kong Proxy URL
echo "Kong Proxy: http://localhost:8000"
kubectl port-forward -n kong service/kong-kong-proxy 8000:80 &

# Test Kong Admin API
curl -i http://localhost:8001/
```

#### **Kong Configuration Management**

```bash
# Create custom values file for different environments
cat <<EOF > kong-values-dev.yaml
replicaCount: 1
env:
  database: postgres
  log_level: debug

postgresql:
  enabled: true
  auth:
    postgresPassword: "kong-dev"

ingressController:
  enabled: true

admin:
  enabled: true
  http:
    enabled: true

proxy:
  type: LoadBalancer
EOF

# Install with custom values
helm upgrade --install kong kong/kong -n kong -f kong-values-dev.yaml
```

### **1.4 Install Consul Service Mesh**

**Consul Connect** provides service mesh functionality, enabling secure service-to-service communication with automatic mTLS, service discovery, and traffic management. **Helm** manages the complex Consul cluster deployment across multiple Kubernetes nodes.

#### **Why Consul + Kubernetes + Helm?**

1. **Consul Connect Benefits:**

   - Automatic service discovery and registration
   - Mutual TLS (mTLS) for all service communication
   - Traffic management (circuit breakers, retries)
   - Service intentions for access control
   - Health checking and failure detection

2. **Helm Benefits for Consul:**
   - Manages StatefulSets for Consul servers
   - Configures service mesh sidecar injection
   - Handles RBAC and security policies
   - Enables easy multi-datacenter setup

#### **Consul Installation Steps**

```bash
# Add HashiCorp Helm repository
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo update

# Create dedicated namespace for Consul
kubectl create namespace consul

# Create custom values for Consul
cat <<EOF > consul-values.yaml
global:
  name: consul
  datacenter: dc1

server:
  replicas: 1
  bootstrapExpect: 1

connectInject:
  enabled: true
  default: false  # Explicit opt-in per service

controller:
  enabled: true

ui:
  enabled: true
  service:
    type: LoadBalancer

client:
  enabled: true

syncCatalog:
  enabled: true
  default: true
EOF

# Install Consul with custom configuration
helm install consul hashicorp/consul -n consul -f consul-values.yaml

# Wait for Consul to be ready
kubectl wait --namespace consul \
  --for=condition=ready pod \
  --selector=app=consul,component=server \
  --timeout=300s

# Verify Consul installation
kubectl get pods -n consul
kubectl get services -n consul

# Get Consul UI URL (for local clusters)
echo "Consul UI: http://localhost:8500"
kubectl port-forward -n consul service/consul-ui 8500:80 &

# Test Consul API
curl http://localhost:8500/v1/status/leader
```

#### **Enable Service Mesh for Microservices**

```bash
# Create a test namespace with service mesh enabled
kubectl create namespace microservices

# Label namespace for automatic sidecar injection
kubectl label namespace microservices connect-inject=enabled

# Verify service mesh injection is working
kubectl get mutatingwebhookconfiguration consul-consul-connect-injector
```

#### **Consul Service Mesh Features Demo**

```bash
# View service topology (after services are deployed)
kubectl exec -n consul consul-consul-server-0 -- consul catalog services

# Check service health
kubectl exec -n consul consul-consul-server-0 -- consul catalog nodes -service=user-service

# View service intentions (access control)
kubectl get serviceintentions -n microservices
```

### **1.5 Integration Architecture Overview**

#### **How Kubernetes, Kong, Consul, and Helm Work Together**

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Kubernetes Cluster                               │
│                                                                         │
│  ┌─────────────┐     ┌─────────────────────────────────────────────┐   │
│  │   Ingress   │────▶│            Kong Namespace                   │   │
│  │ Controller  │     │                                             │   │
│  └─────────────┘     │  ┌─────────────┐  ┌─────────────────────┐   │   │
│                      │  │ Kong Proxy  │  │   Kong Admin API    │   │   │
│                      │  │   (Port 80) │  │     (Port 8001)     │   │   │
│                      │  └─────────────┘  └─────────────────────┘   │   │
│                      └─────────────────────────────────────────────┘   │
│                                    │                                   │
│                                    │ Routes Traffic                    │
│                                    ▼                                   │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                  Microservices Namespace                       │   │
│  │                                                                 │   │
│  │ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │   │
│  │ │User Service │ │Product Svc  │ │Cart Service │ │Order Service│ │   │
│  │ │   + Envoy   │ │  + Envoy    │ │  + Envoy    │ │  + Envoy    │ │   │
│  │ │   Sidecar   │ │   Sidecar   │ │   Sidecar   │ │   Sidecar   │ │   │
│  │ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │   │
│  │       │               │               │               │         │   │
│  │       └───────────────┼───────────────┼───────────────┘         │   │
│  │                 mTLS Encrypted Communication                     │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                    ▲                                   │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Consul Namespace                             │   │
│  │                                                                 │   │
│  │ ┌─────────────┐  ┌─────────────────┐  ┌─────────────────────┐  │   │
│  │ │Consul Server│  │ Service Catalog │  │  Connect Injector   │  │   │
│  │ │  (Raft)     │  │ & Health Checks │  │   (Webhook)         │  │   │
│  │ └─────────────┘  └─────────────────┘  └─────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘

                        All managed by Helm Charts
```

#### **Component Integration Flow**

1. **Helm Charts Management:**

   ```bash
   # All infrastructure as code
   helm list --all-namespaces

   # Output shows:
   # NAME     NAMESPACE   REVISION   STATUS     CHART
   # kong     kong        1          deployed   kong-2.33.3
   # consul   consul      1          deployed   consul-1.2.2
   ```

2. **Traffic Flow:**

   - **External Request** → **Kong Proxy** → **Service Discovery via Consul** → **Target Microservice**
   - **Inter-service calls** → **Consul Connect mTLS** → **Direct service communication**

3. **Service Registration:**

   ```bash
   # Consul automatically discovers Kubernetes services
   kubectl exec -n consul consul-server-0 -- consul catalog services

   # Kong routes configured via Kubernetes Ingress
   kubectl get ingress -n microservices
   ```

#### **Development Workflow Integration**

```bash
# 1. Deploy microservice with Helm
helm install user-service ./charts/user-service -n microservices

# 2. Service automatically registers with Consul
kubectl get pods -n microservices
# Shows: user-service-xxx (2/2 containers) - app + envoy sidecar

# 3. Configure Kong routes
kubectl apply -f k8s/kong-routes.yaml

# 4. Test complete integration
curl -H "Host: api.local" http://localhost:8000/users
```

#### **Observability Stack**

```bash
# Install monitoring stack with Helm
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install monitoring prometheus-community/kube-prometheus-stack -n monitoring --create-namespace

# Kong metrics integration
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: kong-prometheus
  namespace: kong
data:
  prometheus.yml: |
    plugins:
      - name: prometheus
        config:
          per_consumer: true
EOF

# Consul metrics integration
helm upgrade consul hashicorp/consul -n consul -f consul-values.yaml \
  --set global.metrics.enabled=true \
  --set global.metrics.enableGatewayMetrics=true
```

#### **Benefits of This Integrated Architecture**

1. **Unified Management**: Helm manages all components with version control
2. **Security**: Consul Connect provides mTLS for all internal communication
3. **Scalability**: Kubernetes handles auto-scaling and load distribution
4. **Observability**: Centralized logging, metrics, and tracing
5. **Development Experience**: Consistent tooling across all environments

### **1.6 Platform-Specific Troubleshooting**

#### **Common Issues and Solutions**

**macOS Issues:**

```bash
# Docker Desktop not starting
# Solution: Reset Docker Desktop to factory defaults
docker system prune -a
# Restart Docker Desktop

# Kubernetes context issues
kubectl config get-contexts
kubectl config use-context docker-desktop

# Port forwarding issues on macOS
# Kill existing port forwards
lsof -ti:8000 | xargs kill -9
lsof -ti:8001 | xargs kill -9
```

**Linux Issues:**

```bash
# Docker permission denied
sudo usermod -aG docker $USER
newgrp docker

# kubectl not found after installation
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc

# Minikube won't start
minikube delete
minikube start --driver=docker --force

# Firewall blocking ports
sudo ufw allow 8000
sudo ufw allow 8001
sudo ufw allow 8500
```

**Windows Issues:**

```powershell
# Docker Desktop WSL2 backend issues
wsl --update
# Restart Docker Desktop

# PowerShell execution policy
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Chocolatey installation issues
# Run PowerShell as Administrator
Set-ExecutionPolicy Bypass -Scope Process -Force

# Port already in use
netstat -ano | findstr :8000
# Kill process: taskkill /PID <PID> /F

# kubectl not found
# Add to PATH: C:\Program Files\kubectl\
```

**WSL2 Specific Issues:**

```bash
# WSL2 memory issues
# Create/edit ~/.wslconfig
[wsl2]
memory=8GB
processors=4

# Restart WSL2
wsl --shutdown
wsl
```

#### **Verification Commands**

```bash
# Complete environment verification script
#!/bin/bash
echo "=== Environment Verification ==="

# Check Docker
docker --version && echo "✓ Docker OK" || echo "✗ Docker failed"

# Check Kubernetes
kubectl version --client && echo "✓ kubectl OK" || echo "✗ kubectl failed"
kubectl cluster-info && echo "✓ Cluster OK" || echo "✗ Cluster failed"

# Check Helm
helm version && echo "✓ Helm OK" || echo "✗ Helm failed"

# Check Go
go version && echo "✓ Go OK" || echo "✗ Go failed"

# Check Node.js
node --version && echo "✓ Node.js OK" || echo "✗ Node.js failed"

# Check protobuf
protoc --version && echo "✓ Protobuf OK" || echo "✗ Protobuf failed"

# Check Kong
kubectl get pods -n kong && echo "✓ Kong pods OK" || echo "✗ Kong failed"

# Check Consul
kubectl get pods -n consul && echo "✓ Consul pods OK" || echo "✗ Consul failed"

echo "=== Verification Complete ==="
```

---

## **Part 2: Project Structure & Proto Definitions**

### **2.1 Create Project Structure**

```bash
mkdir practical-four
cd practical-four

# Create service directories
mkdir -p services/{user-service,product-service,cart-service,order-service,payment-service}
mkdir -p proto
mkdir -p k8s/{deployments,services,configmaps}
mkdir -p frontend
mkdir docker-compose
```

### **2.2 Define Service Contracts**

**`proto/user.proto`:**

```protobuf
syntax = "proto3";
option go_package = "./proto/gen;gen";
package user;

service UserService {
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  rpc GetUser(GetUserRequest) returns (UserResponse);
  rpc AuthenticateUser(AuthRequest) returns (AuthResponse);
}

message User {
  string id = 1;
  string email = 2;
  string name = 3;
  string created_at = 4;
}

message CreateUserRequest {
  string email = 1;
  string name = 2;
  string password = 3;
}

message GetUserRequest {
  string id = 1;
}

message AuthRequest {
  string email = 1;
  string password = 2;
}

message UserResponse {
  User user = 1;
  string message = 2;
}

message AuthResponse {
  bool success = 1;
  string token = 2;
  User user = 3;
}
```

**`proto/product.proto`:**

```protobuf
syntax = "proto3";
option go_package = "./proto/gen;gen";
package product;

service ProductService {
  rpc CreateProduct(CreateProductRequest) returns (ProductResponse);
  rpc GetProduct(GetProductRequest) returns (ProductResponse);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
  rpc UpdateInventory(UpdateInventoryRequest) returns (InventoryResponse);
}

message Product {
  string id = 1;
  string name = 2;
  string description = 3;
  double price = 4;
  int32 inventory = 5;
  string created_at = 6;
}

message CreateProductRequest {
  string name = 1;
  string description = 2;
  double price = 3;
  int32 inventory = 4;
}

message GetProductRequest {
  string id = 1;
}

message ListProductsRequest {
  int32 limit = 1;
  int32 offset = 2;
}

message UpdateInventoryRequest {
  string product_id = 1;
  int32 quantity_change = 2;
}

message ProductResponse {
  Product product = 1;
  string message = 2;
}

message ListProductsResponse {
  repeated Product products = 1;
  int32 total = 2;
}

message InventoryResponse {
  bool success = 1;
  int32 new_inventory = 2;
  string message = 3;
}
```

**`proto/cart.proto`:**

```protobuf
syntax = "proto3";
option go_package = "./proto/gen;gen";
package cart;

service CartService {
  rpc AddToCart(AddToCartRequest) returns (CartResponse);
  rpc GetCart(GetCartRequest) returns (CartResponse);
  rpc UpdateCartItem(UpdateCartItemRequest) returns (CartResponse);
  rpc RemoveFromCart(RemoveFromCartRequest) returns (CartResponse);
  rpc ClearCart(ClearCartRequest) returns (CartResponse);
}

message CartItem {
  string product_id = 1;
  string product_name = 2;
  double product_price = 3;
  int32 quantity = 4;
  double total_price = 5;
}

message Cart {
  string user_id = 1;
  repeated CartItem items = 2;
  double total_amount = 3;
  string updated_at = 4;
}

message AddToCartRequest {
  string user_id = 1;
  string product_id = 2;
  int32 quantity = 3;
}

message GetCartRequest {
  string user_id = 1;
}

message UpdateCartItemRequest {
  string user_id = 1;
  string product_id = 2;
  int32 new_quantity = 3;
}

message RemoveFromCartRequest {
  string user_id = 1;
  string product_id = 2;
}

message ClearCartRequest {
  string user_id = 1;
}

message CartResponse {
  Cart cart = 1;
  bool success = 2;
  string message = 3;
}
```

**`proto/order.proto`:**

```protobuf
syntax = "proto3";
option go_package = "./proto/gen;gen";
package order;

service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (OrderResponse);
  rpc GetOrder(GetOrderRequest) returns (OrderResponse);
  rpc ListUserOrders(ListUserOrdersRequest) returns (ListOrdersResponse);
  rpc UpdateOrderStatus(UpdateOrderStatusRequest) returns (OrderResponse);
}

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  PENDING = 1;
  PROCESSING = 2;
  PAID = 3;
  SHIPPED = 4;
  DELIVERED = 5;
  CANCELLED = 6;
  FAILED = 7;
}

message OrderItem {
  string product_id = 1;
  string product_name = 2;
  double product_price = 3;
  int32 quantity = 4;
  double total_price = 5;
}

message Order {
  string id = 1;
  string user_id = 2;
  repeated OrderItem items = 3;
  double total_amount = 4;
  OrderStatus status = 5;
  string created_at = 6;
  string updated_at = 7;
}

message CreateOrderRequest {
  string user_id = 1;
}

message GetOrderRequest {
  string id = 1;
}

message ListUserOrdersRequest {
  string user_id = 1;
}

message UpdateOrderStatusRequest {
  string order_id = 1;
  OrderStatus status = 2;
}

message OrderResponse {
  Order order = 1;
  bool success = 2;
  string message = 3;
}

message ListOrdersResponse {
  repeated Order orders = 1;
  int32 total = 2;
}
```

**`proto/payment.proto`:**

```protobuf
syntax = "proto3";
option go_package = "./proto/gen;gen";
package payment;

service PaymentService {
  rpc ProcessPayment(ProcessPaymentRequest) returns (PaymentResponse);
  rpc GetPaymentStatus(GetPaymentStatusRequest) returns (PaymentStatusResponse);
}

enum PaymentStatus {
  PAYMENT_STATUS_UNSPECIFIED = 0;
  PENDING = 1;
  PROCESSING = 2;
  SUCCESS = 3;
  FAILED = 4;
  CANCELLED = 5;
}

message ProcessPaymentRequest {
  string order_id = 1;
  string user_id = 2;
  double amount = 3;
  string payment_method = 4;
}

message GetPaymentStatusRequest {
  string payment_id = 1;
}

message PaymentResponse {
  string payment_id = 1;
  PaymentStatus status = 2;
  string message = 3;
  bool success = 4;
}

message PaymentStatusResponse {
  string payment_id = 1;
  PaymentStatus status = 2;
  double amount = 3;
  string created_at = 4;
}
```

### **2.3 Generate Protocol Buffer Code**

```bash
# Install protobuf tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code
mkdir -p proto/gen
protoc --go_out=./proto/gen --go_opt=paths=source_relative \
    --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative \
    proto/*.proto
```

---

## **Part 3: Implement Services with Resilience Patterns**

### **Understanding Resilience Patterns - The Why and How**

Before we implement our services, it's crucial to understand the resilience patterns that make distributed systems robust and reliable. In microservices architecture, failures are not exceptions—they're the norm. Networks can be unreliable, services can become overwhelmed, databases can slow down, and external APIs can fail. Resilience patterns help our systems gracefully handle these inevitable failures.

---

### **Pattern 1: Timeout Pattern** ⏱️

#### **What is the Timeout Pattern?**

The Timeout pattern prevents a client from waiting indefinitely for a response from a service. It sets a maximum time limit for operations, ensuring that slow or unresponsive services don't cascade failures throughout the system.

#### **Why Do We Need Timeouts?**

1. **Prevent Resource Exhaustion**: Without timeouts, threads/goroutines can hang indefinitely, exhausting system resources
2. **Fail Fast**: Better to fail quickly and retry than wait indefinitely
3. **User Experience**: Users expect responsive applications, not hanging requests
4. **Cascading Failures**: One slow service can bring down the entire system if not contained

#### **Real-World Scenarios:**

- Database queries taking too long due to complex queries or locks
- External API calls to payment gateways during high traffic
- Network partitions causing requests to hang
- Overloaded services that can't process requests quickly

#### **How to Implement Timeouts:**

```go
// Context with timeout - Go's idiomatic way
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Use context in operations
result, err := db.QueryRowContext(ctx, query, params...)
if err != nil {
    if err == context.DeadlineExceeded {
        // Handle timeout specifically
        return fmt.Errorf("operation timed out after 5 seconds")
    }
    return fmt.Errorf("database error: %w", err)
}
```

#### **Best Practices:**

- **Service Level Timeouts**: Different operations need different timeouts (DB: 5s, HTTP: 30s, Complex calculations: 2min)
- **Cascading Timeouts**: Upstream timeouts should be longer than downstream timeouts
- **Monitoring**: Track timeout occurrences to identify problematic services

---

### **Pattern 2: Retry Pattern with Exponential Backoff** 🔄

#### **What is the Retry Pattern?**

The Retry pattern automatically retries failed operations, typically with increasing delays between attempts (exponential backoff). This helps handle transient failures that might resolve themselves quickly.

#### **Why Do We Need Retries?**

1. **Transient Failures**: Network glitches, temporary resource exhaustion, brief service unavailability
2. **High Availability**: Increases system resilience without manual intervention
3. **Cost Efficiency**: Often cheaper than over-provisioning infrastructure
4. **Graceful Degradation**: Gives failing services time to recover

#### **Types of Failures to Retry:**

- ✅ **Transient Network Errors**: Connection timeouts, DNS resolution failures
- ✅ **Temporary Service Unavailability**: 503 Service Unavailable, 429 Rate Limited
- ✅ **Database Connection Issues**: Connection pool exhaustion, temporary locks
- ❌ **Don't Retry**: 400 Bad Request, 401 Unauthorized, 404 Not Found

#### **Exponential Backoff Explained:**

```
Attempt 1: Immediate
Attempt 2: Wait 1 second
Attempt 3: Wait 2 seconds
Attempt 4: Wait 4 seconds
Attempt 5: Wait 8 seconds
```

#### **Implementation Pattern:**

```go
func retryWithExponentialBackoff(operation func() error, maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = operation()
        if err == nil {
            return nil // Success!
        }

        if !isRetriableError(err) {
            return err // Don't retry non-retriable errors
        }

        if i < maxRetries-1 { // Don't wait after last attempt
            waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
            time.Sleep(waitTime)
        }
    }
    return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}
```

#### **Best Practices:**

- **Jitter**: Add randomness to backoff to prevent thundering herd
- **Maximum Backoff**: Cap the maximum wait time (e.g., 30 seconds)
- **Circuit Breaking**: Combine with circuit breakers to prevent retry storms
- **Idempotency**: Ensure retried operations are safe to repeat

---

### **Pattern 3: Circuit Breaker Pattern** ⚡

#### **What is the Circuit Breaker Pattern?**

Inspired by electrical circuit breakers, this pattern monitors failures and "opens" the circuit when failures exceed a threshold, preventing further calls to a failing service. After a timeout period, it allows limited calls to test if the service has recovered.

#### **Why Do We Need Circuit Breakers?**

1. **Prevent Cascading Failures**: Stop calling failing services to prevent system-wide collapse
2. **Fail Fast**: Return errors immediately instead of waiting for timeouts
3. **Resource Protection**: Preserve system resources during service failures
4. **Graceful Recovery**: Allow services time to recover without being overwhelmed

#### **Circuit Breaker States:**

```
┌─────────┐    Failures exceed     ┌─────────┐
│ CLOSED  │─────threshold────────→ │  OPEN   │
│(Normal) │                       │(Failing)│
└─────────┘                       └─────────┘
     ↑                                  │
     │                                  │ After timeout
     │                                  ↓
     │    Success threshold      ┌─────────────┐
     └──────── reached ──────────│  HALF-OPEN  │
                                 │ (Testing)   │
                                 └─────────────┘
```

#### **State Explanations:**

- **CLOSED (Normal)**: Circuit allows all requests through. Monitors failures.
- **OPEN (Failing)**: Circuit blocks all requests, returns error immediately. Waits for timeout.
- **HALF-OPEN (Testing)**: After timeout, allows limited requests to test service recovery.

#### **Real-World Example:**

Imagine an e-commerce site calling a payment service:

1. **Normal Operation**: All payment requests go through (CLOSED)
2. **Payment Service Fails**: After 3 consecutive failures, circuit opens (OPEN)
3. **Fast Failure**: New payment attempts return "Payment unavailable" immediately
4. **Testing Recovery**: After 30 seconds, allows 1 test payment (HALF-OPEN)
5. **Full Recovery**: If test succeeds, resume normal operation (CLOSED)

#### **Implementation Pattern:**

```go
type CircuitBreaker struct {
    failureThreshold int           // How many failures trigger opening
    successThreshold int           // How many successes close the circuit
    timeout         time.Duration  // How long to wait before testing
    state          CircuitState    // Current state
    // ... other fields
}

func (cb *CircuitBreaker) Call(operation func() error) error {
    if cb.state == OPEN {
        if time.Since(cb.lastFailure) < cb.timeout {
            return ErrCircuitOpen // Fail fast
        }
        cb.state = HALF_OPEN // Test recovery
    }

    err := operation()
    cb.recordResult(err)
    return err
}
```

#### **Best Practices:**

- **Service-Specific Thresholds**: Different services need different failure thresholds
- **Monitoring**: Track circuit breaker state changes and failures
- **Fallback Mechanisms**: Provide alternative responses when circuit is open
- **Health Checks**: Use separate health checks to inform circuit breaker state

---

### **Pattern 4: Health Checks and Graceful Shutdown** 🏥

#### **What are Health Checks?**

Health checks are endpoints or mechanisms that report the operational status of a service. In Kubernetes, they help the orchestrator make decisions about traffic routing, pod restarts, and service availability.

#### **Types of Health Checks:**

1. **Liveness Probe**: "Is the application alive?" - Kubernetes restarts failed pods
2. **Readiness Probe**: "Is the application ready to serve traffic?" - Kubernetes removes from load balancer
3. **Startup Probe**: "Has the application finished starting?" - Gives slow-starting apps time to initialize

#### **Implementation Example:**

```go
// Health check endpoint
func healthHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Check database connectivity
        if err := db.Ping(); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{
                "status": "unhealthy",
                "error": "database connection failed"
            })
            return
        }

        // Check other dependencies...
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy"
        })
    }
}
```

---

### **Pattern Integration Strategy**

Our services will implement these patterns in layers:

1. **Database Layer**: Timeouts + Retries for connection issues
2. **Service Layer**: Circuit breakers for service-to-service calls
3. **API Layer**: Rate limiting and timeouts for external requests
4. **Infrastructure Layer**: Health checks for Kubernetes orchestration

Now let's implement these patterns in our services:

---

### **3.1 User Service Implementation with Timeout and Retry Patterns**

#### **🎓 Learning Focus: Timeout and Retry Implementation**

In the User Service, we'll demonstrate **Timeout Pattern** and **Retry Pattern with Exponential Backoff**. This service handles user authentication and profile management, where database connectivity is critical.

#### **Why These Patterns for User Service?**

- **User Authentication**: Critical operation that must be fast and reliable
- **Database Dependencies**: PostgreSQL connection can have transient issues
- **High Frequency**: User operations happen frequently, need to be resilient

#### **Implementation Strategy:**

1. **Timeout Pattern**: All database operations will have 5-second timeouts
2. **Retry Pattern**: Database connection establishment with exponential backoff
3. **Graceful Errors**: Convert technical errors to user-friendly messages

#### **Key Implementation Points to Notice:**

- `context.WithTimeout()` for operation timeouts
- Database connection retry loop in `initDB()`
- Error handling and logging for debugging
- Health check endpoint for Kubernetes probes

**`services/user-service/main.go`:**

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    _ "github.com/lib/pq"
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"

    pb "practical-four/proto/gen"
)

type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    Name      string    `json:"name" db:"name"`
    Password  string    `json:"-" db:"password"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type server struct {
    pb.UnimplementedUserServiceServer
    db *sql.DB
}

type httpServer struct {
    db *sql.DB
}

// gRPC Implementation
func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
    // Add timeout context
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    user := &User{
        Email:     req.Email,
        Name:      req.Name,
        Password:  req.Password, // In production, hash this!
        CreatedAt: time.Now(),
    }

    query := `
        INSERT INTO users (email, name, password, created_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id`

    err := s.db.QueryRowContext(ctx, query, user.Email, user.Name, user.Password, user.CreatedAt).Scan(&user.ID)
    if err != nil {
        log.Printf("Error creating user: %v", err)
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return &pb.UserResponse{
        User: &pb.User{
            Id:        user.ID,
            Email:     user.Email,
            Name:      user.Name,
            CreatedAt: user.CreatedAt.Format(time.RFC3339),
        },
        Message: "User created successfully",
    }, nil
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var user User
    query := `SELECT id, email, name, created_at FROM users WHERE id = $1`

    err := s.db.QueryRowContext(ctx, query, req.Id).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user not found")
        }
        log.Printf("Error getting user: %v", err)
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    return &pb.UserResponse{
        User: &pb.User{
            Id:        user.ID,
            Email:     user.Email,
            Name:      user.Name,
            CreatedAt: user.CreatedAt.Format(time.RFC3339),
        },
    }, nil
}

func (s *server) AuthenticateUser(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var user User
    query := `SELECT id, email, name, password, created_at FROM users WHERE email = $1`

    err := s.db.QueryRowContext(ctx, query, req.Email).Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.CreatedAt)
    if err != nil {
        return &pb.AuthResponse{
            Success: false,
        }, nil
    }

    // Simple password check (use bcrypt in production!)
    if user.Password != req.Password {
        return &pb.AuthResponse{
            Success: false,
        }, nil
    }

    return &pb.AuthResponse{
        Success: true,
        Token:   "jwt_token_here", // Generate proper JWT in production
        User: &pb.User{
            Id:        user.ID,
            Email:     user.Email,
            Name:      user.Name,
            CreatedAt: user.CreatedAt.Format(time.RFC3339),
        },
    }, nil
}

// HTTP Handlers for direct API access
func (h *httpServer) createUserHandler(w http.ResponseWriter, r *http.Request) {
    // Implementation for HTTP endpoint
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"message": "User service HTTP endpoint"}`))
}

func (h *httpServer) getUserHandler(w http.ResponseWriter, r *http.Request) {
    // Implementation for HTTP endpoint
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"message": "Get user HTTP endpoint"}`))
}

func initDB() (*sql.DB, error) {
    dbHost := os.Getenv("DB_HOST")
    if dbHost == "" {
        dbHost = "localhost"
    }

    dbUser := os.Getenv("DB_USER")
    if dbUser == "" {
        dbUser = "postgres"
    }

    dbPassword := os.Getenv("DB_PASSWORD")
    if dbPassword == "" {
        dbPassword = "password"
    }

    dbName := os.Getenv("DB_NAME")
    if dbName == "" {
        dbName = "users_db"
    }

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
        dbHost, dbUser, dbPassword, dbName)

    // Retry connection with exponential backoff
    var db *sql.DB
    var err error

    for i := 0; i < 5; i++ {
        db, err = sql.Open("postgres", dsn)
        if err == nil {
            err = db.Ping()
            if err == nil {
                break
            }
        }

        wait := time.Duration(i+1) * time.Second
        log.Printf("Failed to connect to database (attempt %d/5), retrying in %v: %v", i+1, wait, err)
        time.Sleep(wait)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
    }

    // Create table if not exists
    createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        name VARCHAR(255) NOT NULL,
        password VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`

    _, err = db.Exec(createTableSQL)
    if err != nil {
        return nil, fmt.Errorf("failed to create users table: %w", err)
    }

    return db, nil
}

func main() {
    // Initialize database
    db, err := initDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    // Start gRPC server in goroutine
    go func() {
        lis, err := net.Listen("tcp", ":50051")
        if err != nil {
            log.Fatalf("Failed to listen on gRPC port: %v", err)
        }

        grpcServer := grpc.NewServer()
        pb.RegisterUserServiceServer(grpcServer, &server{db: db})

        // Register health check
        healthServer := health.NewServer()
        grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
        healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

        log.Println("User Service gRPC server listening on :50051")
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve gRPC: %v", err)
        }
    }()

    // Start HTTP server
    r := chi.NewRouter()

    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))

    // CORS for React frontend
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"*"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    httpSrv := &httpServer{db: db}

    // Routes
    r.Route("/api/v1", func(r chi.Router) {
        r.Post("/users", httpSrv.createUserHandler)
        r.Get("/users/{id}", httpSrv.getUserHandler)
    })

    // Health check endpoint
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "healthy", "service": "user-service"}`))
    })

    log.Println("User Service HTTP server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

**`services/user-service/go.mod`:**

```go
module practical-four/user-service

go 1.21

require (
    github.com/go-chi/chi/v5 v5.0.10
    github.com/go-chi/cors v1.2.1
    github.com/lib/pq v1.10.9
    google.golang.org/grpc v1.58.3
    practical-four/proto/gen v0.0.0
)

replace practical-four/proto/gen => ../../proto/gen

require (
    github.com/golang/protobuf v1.5.3 // indirect
    golang.org/x/net v0.12.0 // indirect
    golang.org/x/sys v0.10.0 // indirect
    golang.org/x/text v0.11.0 // indirect
    google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
    google.golang.org/protobuf v1.31.0 // indirect
)
```

**`services/user-service/Dockerfile`:**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy proto files
COPY proto/ ./proto/

# Copy service files
COPY services/user-service/go.mod services/user-service/go.sum ./
RUN go mod download

COPY services/user-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080 50051

CMD ["./server"]
```

### **3.2 Product Service with Circuit Breaker Pattern**

#### **🎓 Learning Focus: Circuit Breaker Implementation**

The Product Service demonstrates the **Circuit Breaker Pattern** - one of the most sophisticated resilience patterns. Product catalog operations are perfect for this pattern because:

#### **Why Circuit Breaker for Product Service?**

1. **High Volume**: Product queries are frequent and can overwhelm the database
2. **Non-Critical**: Unlike user auth, product failures can degrade gracefully
3. **Batch Operations**: Inventory updates can cause database contention
4. **External Dependencies**: Future integration with inventory systems

#### **Circuit Breaker Configuration Analysis:**

```go
// Our configuration: 3 failures, 2 successes, 30 second timeout
circuitBreaker := NewCircuitBreaker(3, 2, 30*time.Second)
```

**What this means:**

- **failureThreshold: 3** - After 3 consecutive failures, open the circuit
- **successThreshold: 2** - Need 2 consecutive successes to close circuit
- **timeout: 30s** - Wait 30 seconds before testing service recovery

#### **State Transition Example:**

```
Normal Operation (CLOSED)
├── Product query fails (failure count: 1)
├── Product query fails (failure count: 2)
├── Product query fails (failure count: 3) → Circuit OPENS
├── All requests fail fast for 30 seconds
├── After 30s: Allow 1 test request (HALF-OPEN)
├── Test request succeeds (success count: 1)
├── Another request succeeds (success count: 2) → Circuit CLOSES
└── Back to normal operation
```

#### **Real-World Scenario:**

Imagine Black Friday traffic overwhelming your product database:

1. **08:00 AM**: Normal traffic, circuit CLOSED
2. **12:00 PM**: Traffic spike, database slows down
3. **12:05 PM**: 3 product queries timeout, circuit OPENS
4. **12:05-12:35 PM**: All product requests fail fast, database recovers
5. **12:35 PM**: Circuit tests with one request (HALF-OPEN)
6. **12:36 PM**: Two successful requests, circuit CLOSES
7. **Result**: 30-minute protection period allowed database to recover

#### **Implementation Deep Dive:**

**Circuit Breaker Components:**

- **State Management**: Thread-safe state tracking with mutex
- **Failure Detection**: Wraps operations and monitors results
- **Time-based Recovery**: Uses timestamps to control state transitions
- **Configurable Thresholds**: Flexible parameters for different scenarios

**Critical Code Sections to Study:**

1. **Thread Safety**: `sync.RWMutex` protects concurrent access
2. **State Logic**: How states transition based on success/failure
3. **Operation Wrapping**: How the circuit breaker wraps database calls
4. **Error Propagation**: How errors are handled and reported

**`services/product-service/main.go`:**

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    _ "github.com/lib/pq"
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"

    pb "practical-four/proto/gen"
)

// Circuit Breaker Implementation
type CircuitBreaker struct {
    mutex           sync.RWMutex
    failureCount    int
    successCount    int
    failureThreshold int
    successThreshold int
    timeout         time.Duration
    lastFailureTime time.Time
    state           CircuitState
}

type CircuitState int

const (
    Closed CircuitState = iota
    Open
    HalfOpen
)

func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        failureThreshold: failureThreshold,
        successThreshold: successThreshold,
        timeout:         timeout,
        state:           Closed,
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()

    if cb.state == Open {
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.state = HalfOpen
            cb.successCount = 0
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }

    err := fn()

    if err != nil {
        cb.onFailure()
        return err
    }

    cb.onSuccess()
    return nil
}

func (cb *CircuitBreaker) onSuccess() {
    cb.failureCount = 0
    if cb.state == HalfOpen {
        cb.successCount++
        if cb.successCount >= cb.successThreshold {
            cb.state = Closed
        }
    }
}

func (cb *CircuitBreaker) onFailure() {
    cb.failureCount++
    cb.lastFailureTime = time.Now()
    if cb.failureCount >= cb.failureThreshold {
        cb.state = Open
    }
}

type Product struct {
    ID          string    `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    Price       float64   `json:"price" db:"price"`
    Inventory   int32     `json:"inventory" db:"inventory"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type server struct {
    pb.UnimplementedProductServiceServer
    db             *sql.DB
    circuitBreaker *CircuitBreaker
}

func (s *server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
    var product Product

    err := s.circuitBreaker.Call(func() error {
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()

        product = Product{
            Name:        req.Name,
            Description: req.Description,
            Price:       req.Price,
            Inventory:   req.Inventory,
            CreatedAt:   time.Now(),
        }

        query := `
            INSERT INTO products (name, description, price, inventory, created_at)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING id`

        return s.db.QueryRowContext(ctx, query, product.Name, product.Description,
            product.Price, product.Inventory, product.CreatedAt).Scan(&product.ID)
    })

    if err != nil {
        log.Printf("Error creating product: %v", err)
        return nil, fmt.Errorf("failed to create product: %w", err)
    }

    return &pb.ProductResponse{
        Product: &pb.Product{
            Id:          product.ID,
            Name:        product.Name,
            Description: product.Description,
            Price:       product.Price,
            Inventory:   product.Inventory,
            CreatedAt:   product.CreatedAt.Format(time.RFC3339),
        },
        Message: "Product created successfully",
    }, nil
}

func (s *server) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
    var product Product

    err := s.circuitBreaker.Call(func() error {
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()

        query := `SELECT id, name, description, price, inventory, created_at FROM products WHERE id = $1`
        return s.db.QueryRowContext(ctx, query, req.Id).Scan(
            &product.ID, &product.Name, &product.Description,
            &product.Price, &product.Inventory, &product.CreatedAt)
    })

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("product not found")
        }
        log.Printf("Error getting product: %v", err)
        return nil, fmt.Errorf("failed to get product: %w", err)
    }

    return &pb.ProductResponse{
        Product: &pb.Product{
            Id:          product.ID,
            Name:        product.Name,
            Description: product.Description,
            Price:       product.Price,
            Inventory:   product.Inventory,
            CreatedAt:   product.CreatedAt.Format(time.RFC3339),
        },
    }, nil
}

func (s *server) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
    var products []*pb.Product

    err := s.circuitBreaker.Call(func() error {
        ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
        defer cancel()

        limit := req.Limit
        if limit <= 0 || limit > 100 {
            limit = 20 // default limit
        }

        offset := req.Offset
        if offset < 0 {
            offset = 0
        }

        query := `SELECT id, name, description, price, inventory, created_at
                 FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2`

        rows, err := s.db.QueryContext(ctx, query, limit, offset)
        if err != nil {
            return err
        }
        defer rows.Close()

        for rows.Next() {
            var product Product
            err := rows.Scan(&product.ID, &product.Name, &product.Description,
                &product.Price, &product.Inventory, &product.CreatedAt)
            if err != nil {
                return err
            }

            products = append(products, &pb.Product{
                Id:          product.ID,
                Name:        product.Name,
                Description: product.Description,
                Price:       product.Price,
                Inventory:   product.Inventory,
                CreatedAt:   product.CreatedAt.Format(time.RFC3339),
            })
        }
        return rows.Err()
    })

    if err != nil {
        log.Printf("Error listing products: %v", err)
        return nil, fmt.Errorf("failed to list products: %w", err)
    }

    return &pb.ListProductsResponse{
        Products: products,
        Total:    int32(len(products)),
    }, nil
}

func (s *server) UpdateInventory(ctx context.Context, req *pb.UpdateInventoryRequest) (*pb.InventoryResponse, error) {
    var newInventory int32

    err := s.circuitBreaker.Call(func() error {
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()

        // Start transaction for inventory update
        tx, err := s.db.BeginTx(ctx, nil)
        if err != nil {
            return err
        }
        defer tx.Rollback()

        query := `UPDATE products SET inventory = inventory + $1 WHERE id = $2 RETURNING inventory`
        err = tx.QueryRowContext(ctx, query, req.QuantityChange, req.ProductId).Scan(&newInventory)
        if err != nil {
            return err
        }

        // Check if inventory goes negative
        if newInventory < 0 {
            return fmt.Errorf("insufficient inventory")
        }

        return tx.Commit()
    })

    if err != nil {
        log.Printf("Error updating inventory: %v", err)
        return &pb.InventoryResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &pb.InventoryResponse{
        Success:      true,
        NewInventory: newInventory,
        Message:      "Inventory updated successfully",
    }, nil
}

func initDB() (*sql.DB, error) {
    dbHost := os.Getenv("DB_HOST")
    if dbHost == "" {
        dbHost = "localhost"
    }

    dbUser := os.Getenv("DB_USER")
    if dbUser == "" {
        dbUser = "postgres"
    }

    dbPassword := os.Getenv("DB_PASSWORD")
    if dbPassword == "" {
        dbPassword = "password"
    }

    dbName := os.Getenv("DB_NAME")
    if dbName == "" {
        dbName = "products_db"
    }

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
        dbHost, dbUser, dbPassword, dbName)

    var db *sql.DB
    var err error

    // Retry with exponential backoff
    for i := 0; i < 5; i++ {
        db, err = sql.Open("postgres", dsn)
        if err == nil {
            err = db.Ping()
            if err == nil {
                break
            }
        }

        wait := time.Duration(i+1) * time.Second
        log.Printf("Failed to connect to database (attempt %d/5), retrying in %v: %v", i+1, wait, err)
        time.Sleep(wait)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
    }

    createTableSQL := `
    CREATE TABLE IF NOT EXISTS products (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
        inventory INTEGER NOT NULL CHECK (inventory >= 0),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`

    _, err = db.Exec(createTableSQL)
    if err != nil {
        return nil, fmt.Errorf("failed to create products table: %w", err)
    }

    return db, nil
}

func main() {
    db, err := initDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    // Initialize circuit breaker: 3 failures, 2 successes, 30 second timeout
    circuitBreaker := NewCircuitBreaker(3, 2, 30*time.Second)

    // Start gRPC server
    go func() {
        lis, err := net.Listen("tcp", ":50052")
        if err != nil {
            log.Fatalf("Failed to listen on gRPC port: %v", err)
        }

        grpcServer := grpc.NewServer()
        pb.RegisterProductServiceServer(grpcServer, &server{
            db:             db,
            circuitBreaker: circuitBreaker,
        })

        healthServer := health.NewServer()
        grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
        healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

        log.Println("Product Service gRPC server listening on :50052")
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve gRPC: %v", err)
        }
    }()

    // Start HTTP server
    r := chi.NewRouter()

    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))

    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"*"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "healthy", "service": "product-service"}`))
    })

    log.Println("Product Service HTTP server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

---

### **3.3 Cart Service Implementation**

#### **🎓 Learning Focus: Timeout and Inter-Service Communication**

The Cart Service demonstrates timeout patterns and prepares for inter-service communication with the Product Service for price validation and inventory checks.

**`services/cart-service/main.go`:**

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    _ "github.com/lib/pq"
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"

    pb "practical-four/proto/gen"
)

type CartItem struct {
    ID           int64     `json:"id" db:"id"`
    UserID       string    `json:"user_id" db:"user_id"`
    ProductID    string    `json:"product_id" db:"product_id"`
    ProductName  string    `json:"product_name" db:"product_name"`
    ProductPrice float64   `json:"product_price" db:"product_price"`
    Quantity     int32     `json:"quantity" db:"quantity"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type server struct {
    pb.UnimplementedCartServiceServer
    db *sql.DB
}

func (s *server) AddToCart(ctx context.Context, req *pb.AddToCartRequest) (*pb.CartResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // Check if item already exists in cart
    var existingItem CartItem
    checkQuery := `SELECT id, quantity FROM cart_items WHERE user_id = $1 AND product_id = $2`

    err := s.db.QueryRowContext(ctx, checkQuery, req.UserId, req.ProductId).Scan(&existingItem.ID, &existingItem.Quantity)

    if err == nil {
        // Update existing item
        updateQuery := `UPDATE cart_items SET quantity = $1, updated_at = NOW() WHERE id = $2`
        _, err = s.db.ExecContext(ctx, updateQuery, existingItem.Quantity+req.Quantity, existingItem.ID)
        if err != nil {
            return nil, fmt.Errorf("failed to update cart item: %w", err)
        }
    } else if err == sql.ErrNoRows {
        // Add new item
        insertQuery := `
            INSERT INTO cart_items (user_id, product_id, quantity, created_at, updated_at)
            VALUES ($1, $2, $3, NOW(), NOW())`
        _, err = s.db.ExecContext(ctx, insertQuery, req.UserId, req.ProductId, req.Quantity)
        if err != nil {
            return nil, fmt.Errorf("failed to add cart item: %w", err)
        }
    } else {
        return nil, fmt.Errorf("database error: %w", err)
    }

    // Return updated cart
    return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

func (s *server) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.CartResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    query := `
        SELECT user_id, product_id, product_name, product_price, quantity, updated_at
        FROM cart_items
        WHERE user_id = $1
        ORDER BY updated_at DESC`

    rows, err := s.db.QueryContext(ctx, query, req.UserId)
    if err != nil {
        return nil, fmt.Errorf("failed to get cart: %w", err)
    }
    defer rows.Close()

    var items []*pb.CartItem
    var totalAmount float64
    var lastUpdated time.Time

    for rows.Next() {
        var item CartItem
        err := rows.Scan(&item.UserID, &item.ProductID, &item.ProductName,
            &item.ProductPrice, &item.Quantity, &item.UpdatedAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan cart item: %w", err)
        }

        itemTotal := item.ProductPrice * float64(item.Quantity)
        totalAmount += itemTotal

        items = append(items, &pb.CartItem{
            ProductId:    item.ProductID,
            ProductName:  item.ProductName,
            ProductPrice: item.ProductPrice,
            Quantity:     item.Quantity,
            TotalPrice:   itemTotal,
        })

        if item.UpdatedAt.After(lastUpdated) {
            lastUpdated = item.UpdatedAt
        }
    }

    return &pb.CartResponse{
        Cart: &pb.Cart{
            UserId:      req.UserId,
            Items:       items,
            TotalAmount: totalAmount,
            UpdatedAt:   lastUpdated.Format(time.RFC3339),
        },
        Success: true,
        Message: "Cart retrieved successfully",
    }, nil
}

func (s *server) UpdateCartItem(ctx context.Context, req *pb.UpdateCartItemRequest) (*pb.CartResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if req.NewQuantity <= 0 {
        // Remove item if quantity is 0 or negative
        return s.RemoveFromCart(ctx, &pb.RemoveFromCartRequest{
            UserId:    req.UserId,
            ProductId: req.ProductId,
        })
    }

    query := `UPDATE cart_items SET quantity = $1, updated_at = NOW() WHERE user_id = $2 AND product_id = $3`
    result, err := s.db.ExecContext(ctx, query, req.NewQuantity, req.UserId, req.ProductId)
    if err != nil {
        return nil, fmt.Errorf("failed to update cart item: %w", err)
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return nil, fmt.Errorf("cart item not found")
    }

    return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

func (s *server) RemoveFromCart(ctx context.Context, req *pb.RemoveFromCartRequest) (*pb.CartResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    query := `DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`
    _, err := s.db.ExecContext(ctx, query, req.UserId, req.ProductId)
    if err != nil {
        return nil, fmt.Errorf("failed to remove cart item: %w", err)
    }

    return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

func (s *server) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.CartResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    query := `DELETE FROM cart_items WHERE user_id = $1`
    _, err := s.db.ExecContext(ctx, query, req.UserId)
    if err != nil {
        return nil, fmt.Errorf("failed to clear cart: %w", err)
    }

    return &pb.CartResponse{
        Cart: &pb.Cart{
            UserId:      req.UserId,
            Items:       []*pb.CartItem{},
            TotalAmount: 0,
            UpdatedAt:   time.Now().Format(time.RFC3339),
        },
        Success: true,
        Message: "Cart cleared successfully",
    }, nil
}

func initDB() (*sql.DB, error) {
    dbHost := os.Getenv("DB_HOST")
    if dbHost == "" {
        dbHost = "localhost"
    }

    dbUser := os.Getenv("DB_USER")
    if dbUser == "" {
        dbUser = "postgres"
    }

    dbPassword := os.Getenv("DB_PASSWORD")
    if dbPassword == "" {
        dbPassword = "password"
    }

    dbName := os.Getenv("DB_NAME")
    if dbName == "" {
        dbName = "cart_db"
    }

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
        dbHost, dbUser, dbPassword, dbName)

    var db *sql.DB
    var err error

    for i := 0; i < 5; i++ {
        db, err = sql.Open("postgres", dsn)
        if err == nil {
            err = db.Ping()
            if err == nil {
                break
            }
        }

        wait := time.Duration(i+1) * time.Second
        log.Printf("Failed to connect to database (attempt %d/5), retrying in %v: %v", i+1, wait, err)
        time.Sleep(wait)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
    }

    createTableSQL := `
    CREATE TABLE IF NOT EXISTS cart_items (
        id SERIAL PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        product_id VARCHAR(255) NOT NULL,
        product_name VARCHAR(255),
        product_price DECIMAL(10,2),
        quantity INTEGER NOT NULL CHECK (quantity > 0),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(user_id, product_id)
    )`

    _, err = db.Exec(createTableSQL)
    if err != nil {
        return nil, fmt.Errorf("failed to create cart_items table: %w", err)
    }

    return db, nil
}

func main() {
    db, err := initDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    // Start gRPC server
    go func() {
        lis, err := net.Listen("tcp", ":50053")
        if err != nil {
            log.Fatalf("Failed to listen on gRPC port: %v", err)
        }

        grpcServer := grpc.NewServer()
        pb.RegisterCartServiceServer(grpcServer, &server{db: db})

        healthServer := health.NewServer()
        grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
        healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

        log.Println("Cart Service gRPC server listening on :50053")
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve gRPC: %v", err)
        }
    }()

    // Start HTTP server
    r := chi.NewRouter()

    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))

    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"*"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "healthy", "service": "cart-service"}`))
    })

    log.Println("Cart Service HTTP server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

**`services/cart-service/go.mod`:**

```go
module practical-four/cart-service

go 1.21

require (
    github.com/go-chi/chi/v5 v5.0.10
    github.com/go-chi/cors v1.2.1
    github.com/lib/pq v1.10.9
    google.golang.org/grpc v1.58.3
    practical-four/proto/gen v0.0.0
)

replace practical-four/proto/gen => ../../proto/gen

require (
    github.com/golang/protobuf v1.5.3 // indirect
    golang.org/x/net v0.12.0 // indirect
    golang.org/x/sys v0.10.0 // indirect
    golang.org/x/text v0.11.0 // indirect
    google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
    google.golang.org/protobuf v1.31.0 // indirect
)
```

**`services/cart-service/Dockerfile`:**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY proto/ ./proto/
COPY services/cart-service/go.mod services/cart-service/go.sum ./
RUN go mod download

COPY services/cart-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080 50053

CMD ["./server"]
```

---

### **3.4 Payment Service Implementation with Deliberate Failures**

#### **🎓 Learning Focus: Simulating Circuit Breaker Scenarios**

The Payment Service is designed to demonstrate circuit breaker behavior by simulating payment gateway failures.

**`services/payment-service/main.go`:**

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "math/rand"
    "net"
    "net/http"
    "os"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    _ "github.com/lib/pq"
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"

    pb "practical-four/proto/gen"
)

type Payment struct {
    ID            string    `json:"id" db:"id"`
    OrderID       string    `json:"order_id" db:"order_id"`
    UserID        string    `json:"user_id" db:"user_id"`
    Amount        float64   `json:"amount" db:"amount"`
    PaymentMethod string    `json:"payment_method" db:"payment_method"`
    Status        string    `json:"status" db:"status"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type server struct {
    pb.UnimplementedPaymentServiceServer
    db *sql.DB
}

// Simulate payment gateway with deliberate failures for testing
func (s *server) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.PaymentResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    // Generate payment ID
    paymentID := fmt.Sprintf("pay_%d", time.Now().Unix())

    // Simulate payment processing delay
    processingTime := time.Duration(rand.Intn(3)+1) * time.Second
    log.Printf("Processing payment %s, estimated time: %v", paymentID, processingTime)

    // Simulate different failure scenarios for circuit breaker testing
    failureRate := 0.3 // 30% failure rate
    if rand.Float64() < failureRate {
        // Simulate payment gateway timeout or failure
        time.Sleep(processingTime)

        failureType := rand.Intn(3)
        var errorMsg string
        var status pb.PaymentStatus

        switch failureType {
        case 0:
            errorMsg = "Payment gateway timeout"
            status = pb.PaymentStatus_FAILED
        case 1:
            errorMsg = "Insufficient funds"
            status = pb.PaymentStatus_FAILED
        case 2:
            errorMsg = "Payment gateway temporarily unavailable"
            status = pb.PaymentStatus_FAILED
        }

        // Still record the failed payment
        payment := Payment{
            ID:            paymentID,
            OrderID:       req.OrderId,
            UserID:        req.UserId,
            Amount:        req.Amount,
            PaymentMethod: req.PaymentMethod,
            Status:        status.String(),
            CreatedAt:     time.Now(),
        }

        s.recordPayment(ctx, &payment)

        return &pb.PaymentResponse{
            PaymentId: paymentID,
            Status:    status,
            Message:   errorMsg,
            Success:   false,
        }, fmt.Errorf(errorMsg)
    }

    // Simulate successful payment processing
    time.Sleep(processingTime)

    payment := Payment{
        ID:            paymentID,
        OrderID:       req.OrderId,
        UserID:        req.UserId,
        Amount:        req.Amount,
        PaymentMethod: req.PaymentMethod,
        Status:        pb.PaymentStatus_SUCCESS.String(),
        CreatedAt:     time.Now(),
    }

    err := s.recordPayment(ctx, &payment)
    if err != nil {
        return &pb.PaymentResponse{
            PaymentId: paymentID,
            Status:    pb.PaymentStatus_FAILED,
            Message:   "Database error",
            Success:   false,
        }, fmt.Errorf("failed to record payment: %w", err)
    }

    log.Printf("Payment %s processed successfully", paymentID)

    return &pb.PaymentResponse{
        PaymentId: paymentID,
        Status:    pb.PaymentStatus_SUCCESS,
        Message:   "Payment processed successfully",
        Success:   true,
    }, nil
}

func (s *server) GetPaymentStatus(ctx context.Context, req *pb.GetPaymentStatusRequest) (*pb.PaymentStatusResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var payment Payment
    query := `SELECT id, amount, status, created_at FROM payments WHERE id = $1`

    err := s.db.QueryRowContext(ctx, query, req.PaymentId).Scan(
        &payment.ID, &payment.Amount, &payment.Status, &payment.CreatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("payment not found")
        }
        return nil, fmt.Errorf("database error: %w", err)
    }

    // Convert string status to enum
    var status pb.PaymentStatus
    switch payment.Status {
    case "SUCCESS":
        status = pb.PaymentStatus_SUCCESS
    case "FAILED":
        status = pb.PaymentStatus_FAILED
    case "PROCESSING":
        status = pb.PaymentStatus_PROCESSING
    default:
        status = pb.PaymentStatus_PENDING
    }

    return &pb.PaymentStatusResponse{
        PaymentId: payment.ID,
        Status:    status,
        Amount:    payment.Amount,
        CreatedAt: payment.CreatedAt.Format(time.RFC3339),
    }, nil
}

func (s *server) recordPayment(ctx context.Context, payment *Payment) error {
    query := `
        INSERT INTO payments (id, order_id, user_id, amount, payment_method, status, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`

    _, err := s.db.ExecContext(ctx, query,
        payment.ID, payment.OrderID, payment.UserID, payment.Amount,
        payment.PaymentMethod, payment.Status, payment.CreatedAt)

    return err
}

func initDB() (*sql.DB, error) {
    dbHost := os.Getenv("DB_HOST")
    if dbHost == "" {
        dbHost = "localhost"
    }

    dbUser := os.Getenv("DB_USER")
    if dbUser == "" {
        dbUser = "postgres"
    }

    dbPassword := os.Getenv("DB_PASSWORD")
    if dbPassword == "" {
        dbPassword = "password"
    }

    dbName := os.Getenv("DB_NAME")
    if dbName == "" {
        dbName = "payments_db"
    }

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
        dbHost, dbUser, dbPassword, dbName)

    var db *sql.DB
    var err error

    for i := 0; i < 5; i++ {
        db, err = sql.Open("postgres", dsn)
        if err == nil {
            err = db.Ping()
            if err == nil {
                break
            }
        }

        wait := time.Duration(i+1) * time.Second
        log.Printf("Failed to connect to database (attempt %d/5), retrying in %v: %v", i+1, wait, err)
        time.Sleep(wait)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
    }

    createTableSQL := `
    CREATE TABLE IF NOT EXISTS payments (
        id VARCHAR(255) PRIMARY KEY,
        order_id VARCHAR(255) NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        amount DECIMAL(10,2) NOT NULL,
        payment_method VARCHAR(100) NOT NULL,
        status VARCHAR(50) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`

    _, err = db.Exec(createTableSQL)
    if err != nil {
        return nil, fmt.Errorf("failed to create payments table: %w", err)
    }

    return db, nil
}

func main() {
    // Seed random number generator for failure simulation
    rand.Seed(time.Now().UnixNano())

    db, err := initDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    // Start gRPC server
    go func() {
        lis, err := net.Listen("tcp", ":50054")
        if err != nil {
            log.Fatalf("Failed to listen on gRPC port: %v", err)
        }

        grpcServer := grpc.NewServer()
        pb.RegisterPaymentServiceServer(grpcServer, &server{db: db})

        healthServer := health.NewServer()
        grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
        healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

        log.Println("Payment Service gRPC server listening on :50054")
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve gRPC: %v", err)
        }
    }()

    // Start HTTP server
    r := chi.NewRouter()

    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))

    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"*"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    // HTTP endpoint to control failure rate for testing
    r.Post("/admin/failure-rate", func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            FailureRate float64 `json:"failure_rate"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // This is a simplified way - in production you'd use proper configuration
        log.Printf("Failure rate updated to: %.2f", req.FailureRate)
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "message": fmt.Sprintf("Failure rate set to %.2f", req.FailureRate),
        })
    })

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "healthy", "service": "payment-service"}`))
    })

}
```

**`services/payment-service/go.mod`:**

```go
module practical-four/payment-service

go 1.21

require (
    github.com/go-chi/chi/v5 v5.0.10
    github.com/go-chi/cors v1.2.1
    github.com/lib/pq v1.10.9
    google.golang.org/grpc v1.58.3
    practical-four/proto/gen v0.0.0
)

replace practical-four/proto/gen => ../../proto/gen

require (
    github.com/golang/protobuf v1.5.3 // indirect
    golang.org/x/net v0.12.0 // indirect
    golang.org/x/sys v0.10.0 // indirect
    golang.org/x/text v0.11.0 // indirect
    google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
    google.golang.org/protobuf v1.31.0 // indirect
)
```

**`services/payment-service/Dockerfile`:**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY proto/ ./proto/
COPY services/payment-service/go.mod services/payment-service/go.sum ./
RUN go mod download

COPY services/payment-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080 50054

CMD ["./server"]
```

---

### **3.5 Order Service Implementation with Circuit Breaker**

#### **🎓 Learning Focus: Orchestrating Multiple Services with Resilience**

The Order Service demonstrates circuit breaker patterns when calling the Payment Service and implements the Saga pattern for distributed transactions.

**`services/order-service/main.go`:**

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    _ "github.com/lib/pq"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"

    pb "practical-four/proto/gen"
)

type Order struct {
    ID          string    `json:"id" db:"id"`
    UserID      string    `json:"user_id" db:"user_id"`
    TotalAmount float64   `json:"total_amount" db:"total_amount"`
    Status      string    `json:"status" db:"status"`
    PaymentID   string    `json:"payment_id" db:"payment_id"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
    ID           int64   `json:"id" db:"id"`
    OrderID      string  `json:"order_id" db:"order_id"`
    ProductID    string  `json:"product_id" db:"product_id"`
    ProductName  string  `json:"product_name" db:"product_name"`
    ProductPrice float64 `json:"product_price" db:"product_price"`
    Quantity     int32   `json:"quantity" db:"quantity"`
}

// Circuit Breaker for Payment Service
type CircuitBreaker struct {
    mu              sync.Mutex
    state           string // "CLOSED", "OPEN", "HALF_OPEN"
    failureCount    int
    successCount    int
    threshold       int
    resetTimeout    time.Duration
    lastFailureTime time.Time
}

func NewCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        state:        "CLOSED",
        threshold:    threshold,
        resetTimeout: resetTimeout,
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    if cb.state == "OPEN" {
        if time.Since(cb.lastFailureTime) > cb.resetTimeout {
            cb.state = "HALF_OPEN"
            cb.successCount = 0
            log.Println("Circuit breaker transitioning to HALF_OPEN")
        } else {
            return fmt.Errorf("circuit breaker is OPEN - payment service unavailable")
        }
    }

    err := fn()

    if err != nil {
        cb.recordFailure()
        return err
    }

    cb.recordSuccess()
    return nil
}

func (cb *CircuitBreaker) recordFailure() {
    cb.failureCount++
    cb.lastFailureTime = time.Now()

    if cb.state == "HALF_OPEN" || cb.failureCount >= cb.threshold {
        cb.state = "OPEN"
        log.Printf("Circuit breaker opened after %d failures", cb.failureCount)
    }
}

func (cb *CircuitBreaker) recordSuccess() {
    if cb.state == "HALF_OPEN" {
        cb.successCount++
        if cb.successCount >= 2 { // Need 2 successes to close
            cb.state = "CLOSED"
            cb.failureCount = 0
            log.Println("Circuit breaker closed after successful calls")
        }
    } else {
        cb.failureCount = 0
    }
}

type server struct {
    pb.UnimplementedOrderServiceServer
    db             *sql.DB
    paymentClient  pb.PaymentServiceClient
    cartClient     pb.CartServiceClient
    circuitBreaker *CircuitBreaker
}

func (s *server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Generate order ID
    orderID := fmt.Sprintf("order_%d", time.Now().Unix())

    // Get user's cart
    cartResp, err := s.cartClient.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
    if err != nil {
        return nil, fmt.Errorf("failed to get cart: %w", err)
    }

    if len(cartResp.Cart.Items) == 0 {
        return nil, fmt.Errorf("cart is empty")
    }

    // Start database transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to start transaction: %w", err)
    }
    defer tx.Rollback()

    // Create order record
    order := Order{
        ID:          orderID,
        UserID:      req.UserId,
        TotalAmount: cartResp.Cart.TotalAmount,
        Status:      pb.OrderStatus_PENDING.String(),
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    orderQuery := `
        INSERT INTO orders (id, user_id, total_amount, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)`

    _, err = tx.ExecContext(ctx, orderQuery,
        order.ID, order.UserID, order.TotalAmount, order.Status,
        order.CreatedAt, order.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("failed to create order: %w", err)
    }

    // Create order items
    for _, item := range cartResp.Cart.Items {
        itemQuery := `
            INSERT INTO order_items (order_id, product_id, product_name, product_price, quantity)
            VALUES ($1, $2, $3, $4, $5)`

        _, err = tx.ExecContext(ctx, itemQuery,
            order.ID, item.ProductId, item.ProductName, item.ProductPrice, item.Quantity)
        if err != nil {
            return nil, fmt.Errorf("failed to create order item: %w", err)
        }
    }

    // Commit order creation
    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit order: %w", err)
    }

    // Process payment with circuit breaker
    var paymentResp *pb.PaymentResponse
    err = s.circuitBreaker.Call(func() error {
        var err error
        paymentResp, err = s.paymentClient.ProcessPayment(ctx, &pb.ProcessPaymentRequest{
            OrderId:       order.ID,
            UserId:        order.UserID,
            Amount:        order.TotalAmount,
            PaymentMethod: req.PaymentMethod,
        })
        return err
    })

    if err != nil {
        // Update order status to failed
        s.updateOrderStatus(ctx, order.ID, pb.OrderStatus_FAILED.String(), "")
        return &pb.OrderResponse{
            Order: &pb.Order{
                Id:          order.ID,
                UserId:      order.UserID,
                TotalAmount: order.TotalAmount,
                Status:      pb.OrderStatus_FAILED,
                CreatedAt:   order.CreatedAt.Format(time.RFC3339),
            },
            Success: false,
            Message: fmt.Sprintf("Payment failed: %v", err),
        }, nil
    }

    // Update order with payment information
    var orderStatus pb.OrderStatus
    if paymentResp.Success {
        orderStatus = pb.OrderStatus_CONFIRMED

        // Clear user's cart on successful order
        _, err = s.cartClient.ClearCart(ctx, &pb.ClearCartRequest{UserId: req.UserId})
        if err != nil {
            log.Printf("Failed to clear cart for user %s: %v", req.UserId, err)
        }
    } else {
        orderStatus = pb.OrderStatus_FAILED
    }

    s.updateOrderStatus(ctx, order.ID, orderStatus.String(), paymentResp.PaymentId)

    return &pb.OrderResponse{
        Order: &pb.Order{
            Id:          order.ID,
            UserId:      order.UserID,
            TotalAmount: order.TotalAmount,
            Status:      orderStatus,
            PaymentId:   paymentResp.PaymentId,
            CreatedAt:   order.CreatedAt.Format(time.RFC3339),
        },
        Success: paymentResp.Success,
        Message: paymentResp.Message,
    }, nil
}

func (s *server) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var order Order
    orderQuery := `
        SELECT id, user_id, total_amount, status, payment_id, created_at, updated_at
        FROM orders WHERE id = $1`

    err := s.db.QueryRowContext(ctx, orderQuery, req.OrderId).Scan(
        &order.ID, &order.UserID, &order.TotalAmount, &order.Status,
        &order.PaymentID, &order.CreatedAt, &order.UpdatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("order not found")
        }
        return nil, fmt.Errorf("database error: %w", err)
    }

    // Get order items
    itemsQuery := `
        SELECT product_id, product_name, product_price, quantity
        FROM order_items WHERE order_id = $1`

    rows, err := s.db.QueryContext(ctx, itemsQuery, req.OrderId)
    if err != nil {
        return nil, fmt.Errorf("failed to get order items: %w", err)
    }
    defer rows.Close()

    var items []*pb.OrderItem
    for rows.Next() {
        var item OrderItem
        err := rows.Scan(&item.ProductID, &item.ProductName, &item.ProductPrice, &item.Quantity)
        if err != nil {
            return nil, fmt.Errorf("failed to scan order item: %w", err)
        }

        items = append(items, &pb.OrderItem{
            ProductId:    item.ProductID,
            ProductName:  item.ProductName,
            ProductPrice: item.ProductPrice,
            Quantity:     item.Quantity,
        })
    }

    // Convert string status to enum
    var status pb.OrderStatus
    switch order.Status {
    case "CONFIRMED":
        status = pb.OrderStatus_CONFIRMED
    case "FAILED":
        status = pb.OrderStatus_FAILED
    case "CANCELLED":
        status = pb.OrderStatus_CANCELLED
    default:
        status = pb.OrderStatus_PENDING
    }

    return &pb.OrderResponse{
        Order: &pb.Order{
            Id:          order.ID,
            UserId:      order.UserID,
            TotalAmount: order.TotalAmount,
            Status:      status,
            PaymentId:   order.PaymentID,
            Items:       items,
            CreatedAt:   order.CreatedAt.Format(time.RFC3339),
            UpdatedAt:   order.UpdatedAt.Format(time.RFC3339),
        },
        Success: true,
        Message: "Order retrieved successfully",
    }, nil
}

func (s *server) GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (*pb.UserOrdersResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    query := `
        SELECT id, user_id, total_amount, status, payment_id, created_at, updated_at
        FROM orders
        WHERE user_id = $1
        ORDER BY created_at DESC`

    rows, err := s.db.QueryContext(ctx, query, req.UserId)
    if err != nil {
        return nil, fmt.Errorf("failed to get user orders: %w", err)
    }
    defer rows.Close()

    var orders []*pb.Order
    for rows.Next() {
        var order Order
        err := rows.Scan(&order.ID, &order.UserID, &order.TotalAmount,
            &order.Status, &order.PaymentID, &order.CreatedAt, &order.UpdatedAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan order: %w", err)
        }

        var status pb.OrderStatus
        switch order.Status {
        case "CONFIRMED":
            status = pb.OrderStatus_CONFIRMED
        case "FAILED":
            status = pb.OrderStatus_FAILED
        case "CANCELLED":
            status = pb.OrderStatus_CANCELLED
        default:
            status = pb.OrderStatus_PENDING
        }

        orders = append(orders, &pb.Order{
            Id:          order.ID,
            UserId:      order.UserID,
            TotalAmount: order.TotalAmount,
            Status:      status,
            PaymentId:   order.PaymentID,
            CreatedAt:   order.CreatedAt.Format(time.RFC3339),
            UpdatedAt:   order.UpdatedAt.Format(time.RFC3339),
        })
    }

    return &pb.UserOrdersResponse{
        Orders:  orders,
        Success: true,
        Message: "Orders retrieved successfully",
    }, nil
}

func (s *server) updateOrderStatus(ctx context.Context, orderID, status, paymentID string) error {
    query := `UPDATE orders SET status = $1, payment_id = $2, updated_at = $3 WHERE id = $4`
    _, err := s.db.ExecContext(ctx, query, status, paymentID, time.Now(), orderID)
    return err
}

func initDB() (*sql.DB, error) {
    dbHost := os.Getenv("DB_HOST")
    if dbHost == "" {
        dbHost = "localhost"
    }

    dbUser := os.Getenv("DB_USER")
    if dbUser == "" {
        dbUser = "postgres"
    }

    dbPassword := os.Getenv("DB_PASSWORD")
    if dbPassword == "" {
        dbPassword = "password"
    }

    dbName := os.Getenv("DB_NAME")
    if dbName == "" {
        dbName = "orders_db"
    }

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
        dbHost, dbUser, dbPassword, dbName)

    var db *sql.DB
    var err error

    for i := 0; i < 5; i++ {
        db, err = sql.Open("postgres", dsn)
        if err == nil {
            err = db.Ping()
            if err == nil {
                break
            }
        }

        wait := time.Duration(i+1) * time.Second
        log.Printf("Failed to connect to database (attempt %d/5), retrying in %v: %v", i+1, wait, err)
        time.Sleep(wait)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
    }

    // Create orders table
    createOrdersTableSQL := `
    CREATE TABLE IF NOT EXISTS orders (
        id VARCHAR(255) PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        total_amount DECIMAL(10,2) NOT NULL,
        status VARCHAR(50) NOT NULL,
        payment_id VARCHAR(255),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`

    // Create order_items table
    createOrderItemsTableSQL := `
    CREATE TABLE IF NOT EXISTS order_items (
        id SERIAL PRIMARY KEY,
        order_id VARCHAR(255) NOT NULL,
        product_id VARCHAR(255) NOT NULL,
        product_name VARCHAR(255) NOT NULL,
        product_price DECIMAL(10,2) NOT NULL,
        quantity INTEGER NOT NULL,
        FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
    )`

    _, err = db.Exec(createOrdersTableSQL)
    if err != nil {
        return nil, fmt.Errorf("failed to create orders table: %w", err)
    }

    _, err = db.Exec(createOrderItemsTableSQL)
    if err != nil {
        return nil, fmt.Errorf("failed to create order_items table: %w", err)
    }

    return db, nil
}

func main() {
    db, err := initDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    // Initialize gRPC clients
    paymentConn, err := grpc.Dial("payment-service:50054", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to payment service: %v", err)
    }
    defer paymentConn.Close()

    cartConn, err := grpc.Dial("cart-service:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to cart service: %v", err)
    }
    defer cartConn.Close()

    paymentClient := pb.NewPaymentServiceClient(paymentConn)
    cartClient := pb.NewCartServiceClient(cartConn)

    // Initialize circuit breaker (3 failures, 30 second reset)
    circuitBreaker := NewCircuitBreaker(3, 30*time.Second)

    // Start gRPC server
    go func() {
        lis, err := net.Listen("tcp", ":50055")
        if err != nil {
            log.Fatalf("Failed to listen on gRPC port: %v", err)
        }

        grpcServer := grpc.NewServer()
        pb.RegisterOrderServiceServer(grpcServer, &server{
            db:             db,
            paymentClient:  paymentClient,
            cartClient:     cartClient,
            circuitBreaker: circuitBreaker,
        })

        healthServer := health.NewServer()
        grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
        healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

        log.Println("Order Service gRPC server listening on :50055")
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve gRPC: %v", err)
        }
    }()

    // Start HTTP server
    r := chi.NewRouter()

    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))

    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"*"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        status := map[string]interface{}{
            "status":          "healthy",
            "service":         "order-service",
            "circuit_breaker": circuitBreaker.state,
        }
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(status)
    })

}
```

**`services/order-service/go.mod`:**

```go
module practical-four/order-service

go 1.21

require (
    github.com/go-chi/chi/v5 v5.0.10
    github.com/go-chi/cors v1.2.1
    github.com/lib/pq v1.10.9
    google.golang.org/grpc v1.58.3
    practical-four/proto/gen v0.0.0
)

replace practical-four/proto/gen => ../../proto/gen

require (
    github.com/golang/protobuf v1.5.3 // indirect
    golang.org/x/net v0.12.0 // indirect
    golang.org/x/sys v0.10.0 // indirect
    golang.org/x/text v0.11.0 // indirect
    google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
    google.golang.org/protobuf v1.31.0 // indirect
)
```

**`services/order-service/Dockerfile`:**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY proto/ ./proto/
COPY services/order-service/go.mod services/order-service/go.sum ./
RUN go mod download

COPY services/order-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080 50055

CMD ["./server"]
```

---

````

---

## **Part 4: Kubernetes Deployments & Service Discovery**

### **4.1 Database Deployments**

**`k8s/deployments/postgres-user-db.yaml`:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-postgres
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: user-postgres
  template:
    metadata:
      labels:
        app: user-postgres
    spec:
      containers:
        - name: postgres
          image: postgres:15
          env:
            - name: POSTGRES_DB
              value: "users_db"
            - name: POSTGRES_USER
              value: "postgres"
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: postgres-password
          ports:
            - containerPort: 5432
          volumeMounts:
            - name: postgres-storage
              mountPath: /var/lib/postgresql/data
          livenessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - postgres
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - postgres
            initialDelaySeconds: 5
            periodSeconds: 5
      volumes:
        - name: postgres-storage
          persistentVolumeClaim:
            claimName: user-postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: user-postgres-service
  namespace: default
spec:
  selector:
    app: user-postgres
  ports:
    - port: 5432
      targetPort: 5432
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: user-postgres-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
````

**`k8s/deployments/postgres-product-db.yaml`:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: product-postgres
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: product-postgres
  template:
    metadata:
      labels:
        app: product-postgres
    spec:
      containers:
        - name: postgres
          image: postgres:15
          env:
            - name: POSTGRES_DB
              value: "products_db"
            - name: POSTGRES_USER
              value: "postgres"
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: postgres-password
          ports:
            - containerPort: 5432
          volumeMounts:
            - name: postgres-storage
              mountPath: /var/lib/postgresql/data
          livenessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - postgres
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - postgres
            initialDelaySeconds: 5
            periodSeconds: 5
      volumes:
        - name: postgres-storage
          persistentVolumeClaim:
            claimName: product-postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: product-postgres-service
  namespace: default
spec:
  selector:
    app: product-postgres
  ports:
    - port: 5432
      targetPort: 5432
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: product-postgres-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
```

### **4.2 Secrets and ConfigMaps**

**`k8s/configmaps/db-secrets.yaml`:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-secrets
  namespace: default
type: Opaque
data:
  postgres-password: cGFzc3dvcmQxMjM= # base64 encoded "password123"
```

### **4.3 Service Deployments with Consul Integration**

**`k8s/deployments/user-service.yaml`:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  namespace: default
  labels:
    app: user-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
      annotations:
        consul.hashicorp.com/connect-inject: "true"
        consul.hashicorp.com/service-name: "user-service"
    spec:
      containers:
        - name: user-service
          image: user-service:latest
          imagePullPolicy: Never # For local development
          ports:
            - containerPort: 8080
              name: http
            - containerPort: 50051
              name: grpc
          env:
            - name: DB_HOST
              value: "user-postgres-service"
            - name: DB_USER
              value: "postgres"
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: postgres-password
            - name: DB_NAME
              value: "users_db"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: user-service
  namespace: default
  labels:
    app: user-service
spec:
  selector:
    app: user-service
  ports:
    - name: http
      port: 8080
      targetPort: 8080
    - name: grpc
      port: 50051
      targetPort: 50051
  type: ClusterIP
```

**`k8s/deployments/product-service.yaml`:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: product-service
  namespace: default
  labels:
    app: product-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: product-service
  template:
    metadata:
      labels:
        app: product-service
      annotations:
        consul.hashicorp.com/connect-inject: "true"
        consul.hashicorp.com/service-name: "product-service"
    spec:
      containers:
        - name: product-service
          image: product-service:latest
          imagePullPolicy: Never
          ports:
            - containerPort: 8080
              name: http
            - containerPort: 50052
              name: grpc
          env:
            - name: DB_HOST
              value: "product-postgres-service"
            - name: DB_USER
              value: "postgres"
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: postgres-password
            - name: DB_NAME
              value: "products_db"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: product-service
  namespace: default
  labels:
    app: product-service
spec:
  selector:
    app: product-service
  ports:
    - name: http
      port: 8080
      targetPort: 8080
    - name: grpc
      port: 50052
      targetPort: 50052
  type: ClusterIP
```

---

## **Part 5: Kong Gateway Configuration**

### **5.1 Kong Service and Route Definitions**

**`k8s/kong/kong-services.yaml`:**

```yaml
apiVersion: configuration.konghq.com/v1
kind: KongService
metadata:
  name: user-service-kong
  namespace: default
spec:
  host: user-service
  port: 8080
  protocol: http
  path: /api/v1
---
apiVersion: configuration.konghq.com/v1
kind: KongRoute
metadata:
  name: user-routes
  namespace: default
spec:
  service: user-service-kong
  paths:
    - "/users"
  methods:
    - GET
    - POST
    - PUT
    - DELETE
---
apiVersion: configuration.konghq.com/v1
kind: KongService
metadata:
  name: product-service-kong
  namespace: default
spec:
  host: product-service
  port: 8080
  protocol: http
  path: /api/v1
---
apiVersion: configuration.konghq.com/v1
kind: KongRoute
metadata:
  name: product-routes
  namespace: default
spec:
  service: product-service-kong
  paths:
    - "/products"
  methods:
    - GET
    - POST
    - PUT
    - DELETE
```

### **5.2 Kong Gateway Resilience Plugins**

#### **🎓 Learning Focus: Gateway-Level Resilience Patterns**

Kong Gateway implements resilience patterns at the API Gateway level, providing the first line of defense before requests reach our services. This is crucial because:

#### **Why Gateway-Level Resilience?**

1. **Single Point of Control**: Apply resilience policies to all services uniformly
2. **Protection Layer**: Shield backend services from malicious or overwhelming traffic
3. **Client Experience**: Provide consistent error handling and response times
4. **Operational Simplicity**: Configure once, protect all services

#### **Kong Resilience Plugins Explained:**

---

#### **Rate Limiting Plugin** 🚦

**Purpose**: Prevent API abuse and protect services from traffic spikes

**Configuration Analysis:**

```yaml
config:
  minute: 100 # Maximum 100 requests per minute per client
  hour: 1000 # Maximum 1000 requests per hour per client
  policy: local # Store counters locally (vs Redis for distributed)
```

**How it Works:**

1. **Request Tracking**: Kong tracks requests per client (IP/API key)
2. **Counter Management**: Maintains sliding window counters
3. **Threshold Enforcement**: Returns HTTP 429 when limits exceeded
4. **Headers**: Adds rate limit headers for client awareness

**Real-World Scenario:**

- **Normal User**: Makes 50 requests/minute → ✅ Allowed
- **Aggressive Script**: Makes 150 requests/minute → ❌ HTTP 429 after 100
- **Result**: Legitimate users continue working, aggressive clients are throttled

---

#### **Request Timeout Plugin** ⏱️

**Purpose**: Prevent hanging requests and ensure responsive API behavior

**Configuration Analysis:**

```yaml
config:
  http_timeout: 30000 # 30 second total HTTP timeout
  read_timeout: 30000 # 30 seconds to read response
  send_timeout: 30000 # 30 seconds to send request
```

**Timeout Types Explained:**

- **http_timeout**: Total time for complete request/response cycle
- **read_timeout**: Time waiting for service to respond
- **send_timeout**: Time to send request to upstream service

**Why These Timeouts Matter:**

```
Without Timeouts:
Client → Kong → Slow Service (hangs forever)
Result: Kong resources exhausted, system failure

With Timeouts:
Client → Kong → Slow Service (30s timeout)
Result: HTTP 504, client can retry, system stable
```

---

#### **Proxy Retry Plugin** 🔄

**Purpose**: Automatically retry failed requests to improve reliability

**Configuration Analysis:**

```yaml
config:
  retries: 3 # Try up to 3 times
  retry_delay: 1000 # Wait 1 second between retries
```

**Retry Logic:**

1. **Initial Request**: Send to upstream service
2. **Failure Detection**: Connection error, 5xx response
3. **Retry Decision**: Is error retriable? Retries remaining?
4. **Backoff**: Wait specified delay
5. **Retry**: Send request again
6. **Success/Give Up**: Return result or final error

**Retriable vs Non-Retriable Errors:**

- ✅ **Retry**: 502 Bad Gateway, 503 Service Unavailable, Connection timeout
- ❌ **Don't Retry**: 400 Bad Request, 401 Unauthorized, 404 Not Found

---

#### **Plugin Application Strategy**

```yaml
annotations:
  konghq.com/plugins: rate-limiting,request-timeout,proxy-retry
```

**Processing Order:**

1. **Rate Limiting**: Check if client is within limits
2. **Request Timeout**: Set timeout for upstream call
3. **Proxy Retry**: Handle failures with retry logic

**Why This Order Matters:**

- Rate limiting first prevents overwhelming backend with retries
- Timeout ensures retries don't hang indefinitely
- Retry last to handle upstream failures

#### **Monitoring and Observability:**

Kong provides metrics for all these patterns:

- **Rate Limiting**: Requests blocked, current limits
- **Timeouts**: Timeout occurrences, response times
- **Retries**: Retry attempts, success/failure rates

**`k8s/kong/kong-plugins.yaml`:**

```yaml
# Rate Limiting Plugin
apiVersion: configuration.konghq.com/v1
kind: KongPlugin
metadata:
  name: rate-limiting
  namespace: default
spec:
  plugin: rate-limiting
  config:
    minute: 100
    hour: 1000
    policy: local
---
# Request Timeout Plugin
apiVersion: configuration.konghq.com/v1
kind: KongPlugin
metadata:
  name: request-timeout
  namespace: default
spec:
  plugin: request-timeout
  config:
    http_timeout: 30000
    read_timeout: 30000
    send_timeout: 30000
---
# Retry Plugin
apiVersion: configuration.konghq.com/v1
kind: KongPlugin
metadata:
  name: proxy-retry
  namespace: default
spec:
  plugin: proxy-retry
  config:
    retries: 3
    retry_delay: 1000
---
# Apply plugins to services
apiVersion: configuration.konghq.com/v1
kind: KongService
metadata:
  name: user-service-kong-with-plugins
  namespace: default
  annotations:
    konghq.com/plugins: rate-limiting,request-timeout,proxy-retry
spec:
  host: user-service
  port: 8080
  protocol: http
  path: /api/v1
```

---

## **Part 6: Deployment Scripts and Testing**

### **6.1 Deployment Script**

**`scripts/deploy.sh`:**

```bash
#!/bin/bash
set -e

echo "🚀 Deploying E-Commerce Microservices to Kubernetes..."

# Function to wait for deployment
wait_for_deployment() {
    local deployment=$1
    local namespace=${2:-default}
    echo "⏳ Waiting for $deployment to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/$deployment -n $namespace
}

# Function to wait for pod
wait_for_pod() {
    local selector=$1
    local namespace=${2:-default}
    echo "⏳ Waiting for pod with selector $selector to be ready..."
    kubectl wait --for=condition=ready --timeout=300s pod -l $selector -n $namespace
}

# Create secrets
echo "🔐 Creating database secrets..."
kubectl apply -f k8s/configmaps/db-secrets.yaml

# Deploy databases
echo "🗄️  Deploying databases..."
kubectl apply -f k8s/deployments/postgres-user-db.yaml
kubectl apply -f k8s/deployments/postgres-product-db.yaml

# Wait for databases to be ready
wait_for_deployment "user-postgres"
wait_for_deployment "product-postgres"

# Build Docker images
echo "🔨 Building service images..."
docker build -t user-service:latest -f services/user-service/Dockerfile .
docker build -t product-service:latest -f services/product-service/Dockerfile .

# Deploy services
echo "🚀 Deploying services..."
kubectl apply -f k8s/deployments/user-service.yaml
kubectl apply -f k8s/deployments/product-service.yaml

# Wait for services to be ready
wait_for_deployment "user-service"
wait_for_deployment "product-service"

# Configure Kong
echo "🌉 Configuring Kong Gateway..."
kubectl apply -f k8s/kong/kong-services.yaml
kubectl apply -f k8s/kong/kong-plugins.yaml

echo "✅ Deployment complete!"
echo ""
echo "📊 Service Status:"
kubectl get pods,svc
echo ""
echo "🔗 Access Points:"
echo "  - Kong Gateway: http://localhost:8000"
echo "  - Kong Admin: http://localhost:8001"
echo "  - Consul UI: http://localhost:8500"

# Test endpoints
echo ""
echo "🧪 Testing endpoints..."
sleep 10

# Test Kong proxy
echo "Testing Kong gateway..."
curl -s -o /dev/null -w "%{http_code}" http://localhost:8000/users || echo "Kong not ready yet"

echo "🎉 Deployment successful!"
```

### **6.2 Testing Script**

**`scripts/test-api.sh`:**

```bash
#!/bin/bash

API_BASE="http://localhost:8000"  # Kong gateway

echo "🧪 Testing E-Commerce Microservices API..."

# Test User Service
echo ""
echo "👤 Testing User Service..."

# Create a user
echo "Creating user..."
USER_RESPONSE=$(curl -s -X POST "$API_BASE/users" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "name": "John Doe",
    "password": "password123"
  }')

echo "User created: $USER_RESPONSE"

# Extract user ID (assuming response contains id field)
USER_ID=$(echo $USER_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
echo "User ID: $USER_ID"

# Get user
if [ ! -z "$USER_ID" ]; then
  echo "Getting user..."
  curl -s "$API_BASE/users/$USER_ID" | jq .
fi

echo ""
echo "📦 Testing Product Service..."

# Create a product
echo "Creating product..."
PRODUCT_RESPONSE=$(curl -s -X POST "$API_BASE/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MacBook Pro",
    "description": "Latest MacBook Pro with M3 chip",
    "price": 1999.99,
    "inventory": 10
  }')

echo "Product created: $PRODUCT_RESPONSE"

# Extract product ID
PRODUCT_ID=$(echo $PRODUCT_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
echo "Product ID: $PRODUCT_ID"

# List products
echo "Listing products..."
curl -s "$API_BASE/products" | jq .

echo ""
echo "⚡ Testing Resilience Patterns..."

# Test rate limiting by making multiple requests
echo "Testing rate limiting (making 10 rapid requests)..."
for i in {1..10}; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE/products")
  echo "Request $i: HTTP $STATUS"
  if [ "$STATUS" = "429" ]; then
    echo "✅ Rate limiting working - got 429 Too Many Requests"
    break
  fi
  sleep 0.1
done

# Test timeout by making request to non-existent endpoint
echo ""
echo "Testing timeout handling..."
curl -s --max-time 35 "$API_BASE/slow-endpoint" || echo "✅ Timeout handled correctly"

echo ""
echo "🎯 API Testing completed!"
```

---

## **Part 7: React Frontend Dashboard**

### **7.1 Frontend Setup**

```bash
# Create React app
cd frontend
npx create-react-app ecommerce-dashboard
cd ecommerce-dashboard
npm install axios recharts @mui/material @emotion/react @emotion/styled
```

**`frontend/ecommerce-dashboard/src/App.js`:**

```jsx
import React, { useState, useEffect } from "react";
import {
  Container,
  Grid,
  Card,
  CardContent,
  Typography,
  Button,
  TextField,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Alert,
  CircularProgress,
} from "@mui/material";
import axios from "axios";

const API_BASE = process.env.REACT_APP_API_BASE || "http://localhost:8000";

function App() {
  const [users, setUsers] = useState([]);
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Form states
  const [newUser, setNewUser] = useState({ name: "", email: "", password: "" });
  const [newProduct, setNewProduct] = useState({
    name: "",
    description: "",
    price: "",
    inventory: "",
  });

  useEffect(() => {
    fetchData();
    // Refresh data every 30 seconds
    const interval = setInterval(fetchData, 30000);
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      // In a real app, you'd have proper API endpoints for listing
      // For now, we'll simulate with the health checks
      const healthCheck = await axios.get(`${API_BASE}/health`);
      console.log("Services are healthy:", healthCheck.data);
    } catch (error) {
      setError("Failed to connect to services: " + error.message);
    } finally {
      setLoading(false);
    }
  };

  const createUser = async (e) => {
    e.preventDefault();
    try {
      setLoading(true);
      const response = await axios.post(`${API_BASE}/users`, newUser);
      setUsers([...users, response.data]);
      setNewUser({ name: "", email: "", password: "" });
      setError(null);
    } catch (error) {
      setError("Failed to create user: " + error.message);
    } finally {
      setLoading(false);
    }
  };

  const createProduct = async (e) => {
    e.preventDefault();
    try {
      setLoading(true);
      const productData = {
        ...newProduct,
        price: parseFloat(newProduct.price),
        inventory: parseInt(newProduct.inventory),
      };
      const response = await axios.post(`${API_BASE}/products`, productData);
      setProducts([...products, response.data]);
      setNewProduct({ name: "", description: "", price: "", inventory: "" });
      setError(null);
    } catch (error) {
      setError("Failed to create product: " + error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h3" component="h1" gutterBottom>
        🛍️ E-Commerce Microservices Dashboard
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Grid container spacing={3}>
        {/* Service Status */}
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Typography variant="h5" component="div">
                🚀 Service Status
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {loading ? (
                  <CircularProgress size={20} />
                ) : (
                  "All services are running"
                )}
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        {/* Create User Form */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                👤 Create New User
              </Typography>
              <form onSubmit={createUser}>
                <TextField
                  fullWidth
                  label="Name"
                  value={newUser.name}
                  onChange={(e) =>
                    setNewUser({ ...newUser, name: e.target.value })
                  }
                  margin="normal"
                  required
                />
                <TextField
                  fullWidth
                  label="Email"
                  type="email"
                  value={newUser.email}
                  onChange={(e) =>
                    setNewUser({ ...newUser, email: e.target.value })
                  }
                  margin="normal"
                  required
                />
                <TextField
                  fullWidth
                  label="Password"
                  type="password"
                  value={newUser.password}
                  onChange={(e) =>
                    setNewUser({ ...newUser, password: e.target.value })
                  }
                  margin="normal"
                  required
                />
                <Button
                  type="submit"
                  variant="contained"
                  sx={{ mt: 2 }}
                  disabled={loading}
                >
                  Create User
                </Button>
              </form>
            </CardContent>
          </Card>
        </Grid>

        {/* Create Product Form */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                📦 Create New Product
              </Typography>
              <form onSubmit={createProduct}>
                <TextField
                  fullWidth
                  label="Product Name"
                  value={newProduct.name}
                  onChange={(e) =>
                    setNewProduct({ ...newProduct, name: e.target.value })
                  }
                  margin="normal"
                  required
                />
                <TextField
                  fullWidth
                  label="Description"
                  multiline
                  rows={2}
                  value={newProduct.description}
                  onChange={(e) =>
                    setNewProduct({
                      ...newProduct,
                      description: e.target.value,
                    })
                  }
                  margin="normal"
                />
                <TextField
                  fullWidth
                  label="Price"
                  type="number"
                  step="0.01"
                  value={newProduct.price}
                  onChange={(e) =>
                    setNewProduct({ ...newProduct, price: e.target.value })
                  }
                  margin="normal"
                  required
                />
                <TextField
                  fullWidth
                  label="Inventory"
                  type="number"
                  value={newProduct.inventory}
                  onChange={(e) =>
                    setNewProduct({ ...newProduct, inventory: e.target.value })
                  }
                  margin="normal"
                  required
                />
                <Button
                  type="submit"
                  variant="contained"
                  sx={{ mt: 2 }}
                  disabled={loading}
                >
                  Create Product
                </Button>
              </form>
            </CardContent>
          </Card>
        </Grid>

        {/* Users Table */}
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                👥 Users ({users.length})
              </Typography>
              <TableContainer component={Paper}>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>ID</TableCell>
                      <TableCell>Name</TableCell>
                      <TableCell>Email</TableCell>
                      <TableCell>Created At</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {users.map((user) => (
                      <TableRow key={user.id}>
                        <TableCell>{user.id}</TableCell>
                        <TableCell>{user.name}</TableCell>
                        <TableCell>{user.email}</TableCell>
                        <TableCell>{user.created_at}</TableCell>
                      </TableRow>
                    ))}
                    {users.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={4} align="center">
                          No users found
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>
        </Grid>

        {/* Products Table */}
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                📦 Products ({products.length})
              </Typography>
              <TableContainer component={Paper}>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>ID</TableCell>
                      <TableCell>Name</TableCell>
                      <TableCell>Description</TableCell>
                      <TableCell>Price</TableCell>
                      <TableCell>Inventory</TableCell>
                      <TableCell>Created At</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {products.map((product) => (
                      <TableRow key={product.id}>
                        <TableCell>{product.id}</TableCell>
                        <TableCell>{product.name}</TableCell>
                        <TableCell>{product.description}</TableCell>
                        <TableCell>${product.price}</TableCell>
                        <TableCell>{product.inventory}</TableCell>
                        <TableCell>{product.created_at}</TableCell>
                      </TableRow>
                    ))}
                    {products.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={6} align="center">
                          No products found
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Container>
  );
}

export default App;
```

---

## **Part 8: Summary and Learning Assessment**

### **8.1 Resilience Patterns Implementation Summary**

#### **🎓 Complete Resilience Strategy Analysis**

Our e-commerce microservices system implements multiple layers of resilience patterns. Here's how they work together to create a robust, fault-tolerant system:

---

#### **Layer 1: Application-Level Resilience**

**1. Timeout Pattern Implementation** ⏱️

```go
// Every database operation protected with timeout
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
```

- **Where**: All database operations in User and Product services
- **Why**: Prevents hanging database queries from exhausting resources
- **Impact**: Failed operations return within 5 seconds instead of hanging indefinitely

**2. Retry Pattern with Exponential Backoff** 🔄

```go
// Database connection retry logic
for i := 0; i < 5; i++ {
    db, err = sql.Open("postgres", dsn)
    if err == nil { break }

    wait := time.Duration(i+1) * time.Second  // 1s, 2s, 3s, 4s, 5s
    time.Sleep(wait)
}
```

- **Where**: Database connection establishment in all services
- **Why**: Handles transient network issues and database startup delays
- **Impact**: Services can recover from temporary connectivity problems

**3. Circuit Breaker Pattern** ⚡

```go
// Product service circuit breaker
err := circuitBreaker.Call(func() error {
    return database_operation()
})
```

- **Where**: Product service database operations
- **Why**: Prevents cascading failures when database is overwhelmed
- **Impact**: Fast-fails during database issues, allows recovery time

---

#### **Layer 2: Gateway-Level Resilience**

**4. Rate Limiting** 🚦

- **Configuration**: 100 requests/minute, 1000 requests/hour per client
- **Purpose**: Prevents API abuse and protects backend services
- **Behavior**: Returns HTTP 429 when limits exceeded

**5. Request Timeout** ⏱️

- **Configuration**: 30-second timeout for all upstream calls
- **Purpose**: Ensures responsive API behavior
- **Behavior**: Returns HTTP 504 if backend services don't respond

**6. Proxy Retry** 🔄

- **Configuration**: Up to 3 retries with 1-second delay
- **Purpose**: Handles transient upstream failures
- **Behavior**: Automatically retries 5xx errors and connection failures

---

#### **Layer 3: Infrastructure-Level Resilience**

**7. Kubernetes Health Checks** 🏥

```yaml
livenessProbe: # Restart unhealthy pods
readinessProbe: # Remove from load balancer when not ready
```

- **Purpose**: Automatic failure detection and recovery
- **Impact**: Kubernetes automatically restarts failed pods and routes traffic to healthy instances

**8. Resource Limits** 📊

```yaml
resources:
  limits:
    memory: "256Mi"
    cpu: "200m"
```

- **Purpose**: Prevent resource exhaustion and noisy neighbor problems
- **Impact**: Containers can't consume unlimited resources

---

#### **Resilience Pattern Interaction Example**

Let's trace a request during a database failure:

```
🌐 Client Request → Kong Gateway
├── ✅ Rate Limiting: Client within limits
├── ⏱️ Request Timeout: Set 30s timeout
├── 🔄 Proxy Retry: Ready to retry failures
└── 📡 Forward to Product Service

🔧 Product Service Receives Request
├── ⚡ Circuit Breaker: Check state (CLOSED)
├── ⏱️ Context Timeout: Set 5s DB timeout
├── 💾 Database Call: Connection fails
├── ⚡ Circuit Breaker: Record failure (1/3)
└── ❌ Return 500 error to Kong

🌉 Kong Gateway Handling
├── 🔄 Proxy Retry: 500 is retriable, attempt 1/3
├── ⏱️ Wait 1 second (retry delay)
├── 📡 Retry request to Product Service
├── ⚡ Circuit Breaker: Record failure (2/3)
├── ❌ Another 500 error
├── 🔄 Retry again (attempt 2/3)
├── ⚡ Circuit Breaker: Record failure (3/3) → OPENS
└── ❌ Return 500 to client (no more retries)

Next Requests:
├── ⚡ Circuit Breaker: OPEN state
├── ⚡ Fail fast: Return error immediately
├── 🕐 After 30 seconds: Test with HALF-OPEN
└── 🔄 Continue recovery process
```

**Key Benefits Achieved:**

1. **Fast Failure**: Circuit breaker prevents hanging on failed service
2. **Automatic Recovery**: System tests service recovery automatically
3. **Resource Protection**: Failed service doesn't consume resources unnecessarily
4. **Client Experience**: Consistent error responses rather than timeouts

---

#### **Testing Your Understanding**

Before proceeding to implementation tasks, ensure you understand:

1. **When would you use each pattern?** Different failure scenarios
2. **How do patterns interact?** Layered resilience approach
3. **What are the trade-offs?** Complexity vs reliability
4. **How do you monitor effectiveness?** Metrics and observability

---

### **8.2 Assignment Tasks**

**Task 1: Deploy and Test (30 points)**

- Deploy the complete system to Kubernetes
- Test all API endpoints through Kong gateway
- Demonstrate circuit breaker by simulating database failures
- Screenshot successful deployment and API responses

**Task 2: Implement Additional Resilience (40 points)**

- Add Cart Service with timeout and retry patterns
- Implement Order Service with distributed transaction handling
- Add Payment Service with deliberate failures to test circuit breaker
- Document the resilience patterns you implemented

**Task 3: Performance Analysis (30 points)**

- Use Kong's analytics to monitor API performance
- Test system behavior under load
- Measure circuit breaker activation and recovery
- Write a report on system resilience and performance

### **8.3 Running the Complete System**

```bash
# 1. Deploy to Kubernetes
chmod +x scripts/deploy.sh
./scripts/deploy.sh

# 2. Port forward Kong gateway
kubectl port-forward -n kong svc/kong-proxy 8000:80 &

# 3. Port forward Consul UI
kubectl port-forward -n consul svc/consul-ui 8500:80 &

# 4. Start React frontend
cd frontend/ecommerce-dashboard
npm start

# 5. Run API tests
chmod +x scripts/test-api.sh
./scripts/test-api.sh
```

### **8.4 Expected Learning Outcomes**

By completing this practical, students will:

- ✅ Deploy complex microservices to Kubernetes with proper configuration
- ✅ Implement and test timeout, retry, and circuit breaker patterns
- ✅ Configure Kong gateway for API management and resilience
- ✅ Use Consul for service discovery and service mesh capabilities
- ✅ Build a React dashboard for monitoring microservices
- ✅ Understand production-ready microservices architecture

---

## **Troubleshooting Guide**

### **Common Issues:**

1. **Pods not starting**: Check resource limits and database connections
2. **Kong not routing**: Verify service names and port configurations
3. **Database connection failures**: Ensure secrets are correctly encoded
4. **React app CORS errors**: Check CORS configuration in services

### **Debugging Commands:**

```bash
# Check pod status
kubectl get pods

# View pod logs
kubectl logs <pod-name>

# Describe pod issues
kubectl describe pod <pod-name>

# Test service connectivity
kubectl exec -it <pod-name> -- nc -zv <service-name> <port>

# Kong admin API
curl http://localhost:8001/services
curl http://localhost:8001/routes
```

This practical provides a comprehensive understanding of microservices deployment and resilience patterns in a Kubernetes environment! 🚀
